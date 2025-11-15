package gameservice

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"multiverse-core/internal/entity"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client *minio.Client
}

func NewMinioClient() (*MinioClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "minio:9000"
	}

	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{client: minioClient}, nil
}

func (mc *MinioClient) LoadEntity(ctx context.Context, entityID, worldID string) (*entity.Entity, error) {
	// Try world-specific bucket
	bucket := "entities-" + worldID
	obj, err := mc.client.GetObject(ctx, bucket, entityID+".json", minio.GetObjectOptions{})
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
	obj, err = mc.client.GetObject(ctx, bucket, entityID+".json", minio.GetObjectOptions{})
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

func (mc *MinioClient) SaveEntity(ctx context.Context, ent *entity.Entity, worldID string) error {
	bucket := "entities-" + worldID
	
	// Ensure bucket exists
	exists, err := mc.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := mc.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	// Save entity
	data, err := json.Marshal(ent)
	if err != nil {
		return err
	}

	_, err = mc.client.PutObject(ctx, bucket, ent.EntityID+".json", 
		bytes.NewReader(data), int64(len(data)), 
		minio.PutObjectOptions{ContentType: "application/json"})
	return err
}