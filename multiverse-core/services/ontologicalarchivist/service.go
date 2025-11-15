// Package ontologicalarchivist implements the OntologicalArchivist service.
package ontologicalarchivist

import (
	"context"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	KafkaBrokers   []string
}

// Service manages ontological schemas in MinIO.
type Service struct {
	minio *minio.Client
}

// NewService creates a new OntologicalArchivist service.
func NewService(cfg Config) *Service {

	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("Failed to connect to MinIO:", err)
	}

	// Create schemas bucket
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	minioClient.MakeBucket(ctx, "schemas", minio.MakeBucketOptions{})

	return &Service{minio: minioClient}
}

// SaveSchema saves a schema to MinIO.
func (s *Service) SaveSchema(ctx context.Context, schemaType, name, version string, schemaData []byte) error {
	key := schemaType + "/" + name + "/v" + version + ".json"
	_, err := s.minio.PutObject(ctx, "schemas", key,
		NewBytesReader(schemaData), int64(len(schemaData)),
		minio.PutObjectOptions{ContentType: "application/json; charset=utf-8"})
	return err
}

// GetSchema retrieves a schema from MinIO.
func (s *Service) GetSchema(ctx context.Context, schemaType, name, version string) ([]byte, error) {
	key := schemaType + "/" + name + "/v" + version + ".json"
	obj, err := s.minio.GetObject(ctx, "schemas", key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return ReadAll(obj)
}
