// services/entitymanager/manager.go
package entitymanager

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"

	"multiverse-core/internal/entity"
	"multiverse-core/internal/eventbus"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Manager struct {
	minio *minio.Client // ← теперь это *github.com/minio/minio-go/v7.Client
}

// NewManager creates a new EntityManager with MinIO client.
func NewManager() (*Manager, error) {
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "minio:9000"
	}

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &Manager{minio: minioClient}, nil
}

// getBucketForEntity determines the MinIO bucket for an entity.
func (m *Manager) getBucketForEntity(ent *entity.Entity, evt *eventbus.Event) string {
	return "entities-" + evt.WorldID
	//if worldID, exists := evt.WorldID; exists && worldID != "" {
	//	return "entities-" + worldID
	//}
	//return "entities-global"
}

// ensureBucket creates a bucket if it doesn't exist.
func (m *Manager) ensureBucket(ctx context.Context, bucket string) error {
	exists, err := m.minio.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return m.minio.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}

// loadEntityFromMinIO loads an entity from MinIO (first in world bucket, then global).
func (m *Manager) loadEntityFromMinIO(ctx context.Context, entityID, worldID string) (*entity.Entity, error) {
	// Try world-specific bucket
	bucket := "entities-" + worldID
	obj, err := m.minio.GetObject(ctx, bucket, entityID+".json", minio.GetObjectOptions{})
	if err == nil {
		defer obj.Close()
		var ent entity.Entity
		if err := json.NewDecoder(obj).Decode(&ent); err != nil {
			return nil, err
		}
		return &ent, nil
	}

	// Try global bucket
	bucket = "entities-global"
	obj, err = m.minio.GetObject(ctx, bucket, entityID+".json", minio.GetObjectOptions{})
	if err == nil {
		defer obj.Close()
		var ent entity.Entity
		if err := json.NewDecoder(obj).Decode(&ent); err != nil {
			return nil, err
		}
		return &ent, nil
	}

	return nil, err // not found
}

// saveSnapshotToMinIO saves an entity to its appropriate bucket.
func (m *Manager) saveSnapshotToMinIO(ctx context.Context, ent *entity.Entity, evt *eventbus.Event) error {
	bucket := m.getBucketForEntity(ent, evt)
	if err := m.ensureBucket(ctx, bucket); err != nil {
		return err
	}

	data, err := json.Marshal(ent)

	if err != nil {
		return err
	}

	_, err = m.minio.PutObject(ctx, bucket, ent.EntityID+".json",
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/json; charset=utf-8"})
	return err
}

// HandleEvent processes an event from any topic.
func (m *Manager) HandleEvent(ev eventbus.Event) {
	ctx := context.Background()

	// 1. Process entity_snapshots (for travel events, full update)
	if snapshotsRaw, exists := ev.Payload["entity_snapshots"]; exists {
		if snapshots, ok := snapshotsRaw.([]interface{}); ok {
			for _, snapRaw := range snapshots {
				if snapMap, ok := snapRaw.(map[string]interface{}); ok {
					// Convert map to Entity
					data, _ := json.Marshal(snapMap)
					var ent entity.Entity
					json.Unmarshal(data, &ent)

					if err := m.saveSnapshotToMinIO(ctx, &ent, &ev); err != nil {
						log.Printf("Failed to save snapshot for %s: %v", ent.EntityID, err)
					} else {
						log.Printf("Saved entity %s to bucket %s", ent.EntityID, m.getBucketForEntity(&ent, &ev))
					}
				}
			}
		}
	}

	// 2. Process state_changes (for regular updates)
	if changesRaw, exists := ev.Payload["state_changes"]; exists {
		if changes, ok := changesRaw.([]interface{}); ok {
			for _, changeRaw := range changes {
				if changeMap, ok := changeRaw.(map[string]interface{}); ok {
					entityID, ok := changeMap["entity_id"].(string)
					if !ok {
						continue
					}

					// Load existing entity (from world or global)
					ent, err := m.loadEntityFromMinIO(ctx, entityID, ev.WorldID)
					if err != nil {
						// Create new entity if not found
						ent = entity.NewEntity(entityID, "unknown", nil)
					}

					// Apply operations
					if opsRaw, ok := changeMap["operations"].([]interface{}); ok {
						for _, opRaw := range opsRaw {
							if opMap, ok := opRaw.(map[string]interface{}); ok {
								opType, _ := opMap["op"].(string)
								path, _ := opMap["path"].(string)

								switch OperationType(opType) {
								case OpSet:
									if value, exists := opMap["value"]; exists {
										ent.Set(path, value)
									}
								case OpAddToSlice:
									if value, exists := opMap["value"].(string); exists {
										ent.AddToStringSlice(path, value)
									}
								case OpRemoveFromSlice:
									if value, exists := opMap["value"].(string); exists {
										ent.RemoveFromStringSlice(path, value)
									}
								case OpRemove:
									ent.Remove(path)
								}
							}
						}
					}

					// Add to history and save
					ent.AddHistoryEntry(ev.EventID, ev.Timestamp)
					if err := m.saveSnapshotToMinIO(ctx, ent, &ev); err != nil {
						log.Printf("Failed to update entity %s: %v", entityID, err)
					} else {
						log.Printf("Updated entity %s", entityID)
					}
				}
			}
		}
	}

	// 3. Process entity.created events (for new entities)
	if ev.EventType == "entity.created" {
		// Extract entity data from event payload
		entityID, idExists := ev.Payload["entity_id"].(string)
		entityType, typeExists := ev.Payload["entity_type"].(string)
		payloadRaw, payloadExists := ev.Payload["payload"]

		if idExists && typeExists && payloadExists {
			// Convert payload to map[string]interface{}
			var payload map[string]interface{}
			if payloadMap, ok := payloadRaw.(map[string]interface{}); ok {
				payload = payloadMap
			} else {
				// Try to marshal and unmarshal if it's a different type
				data, err := json.Marshal(payloadRaw)
				if err != nil {
					log.Printf("Failed to marshal payload for entity %s: %v", entityID, err)
					return
				}
				if err := json.Unmarshal(data, &payload); err != nil {
					log.Printf("Failed to unmarshal payload for entity %s: %v", entityID, err)
					return
				}
			}

			// Create new entity
			ent := entity.NewEntity(entityID, entityType, payload)

			// Add history entry
			ent.AddHistoryEntry(ev.EventID, ev.Timestamp)

			// Save to MinIO
			if err := m.saveSnapshotToMinIO(ctx, ent, &ev); err != nil {
				log.Printf("Failed to save created entity %s: %v", entityID, err)
			} else {
				log.Printf("Saved newly created entity %s to bucket %s", entityID, m.getBucketForEntity(ent, &ev))
			}
		} else {
			log.Printf("Incomplete entity.created event payload for entity_id=%v, entity_type=%v, payload=%v", 
				entityID, entityType, payloadExists)
		}
	}

	log.Println(ev)

}
