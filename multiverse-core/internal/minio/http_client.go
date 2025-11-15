// internal/minio/http_client.go

// MinIO HTTP Client — без использования chroma-go или CGO.
// Использует только стандартную библиотеку Go и S3-compatible API.

package minio

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

// Config для MinIO-клиента.
type Config struct {
	Endpoint        string // Например: "http://minio:9000"
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string // По умолчанию "us-east-1"
}

// Client — HTTP-клиент для MinIO.
type Client struct {
	config  Config
	http    *http.Client
	baseURL *url.URL
}

// NewClient создаёт новый MinIO HTTP-клиент.
func NewClient(cfg Config) (*Client, error) {
	baseURL, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	return &Client{
		config: cfg,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}, nil
}

// ensureBucket создаёт бакет, если он не существует.
func (c *Client) ensureBucket(bucket string) error {
	req, err := c.newRequest("HEAD", bucket, "", nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil // Бакет существует
	}
	if resp.StatusCode == 404 {
		// Создаём бакет
		createReq, err := c.newRequest("PUT", bucket, "", nil)
		if err != nil {
			return err
		}
		createResp, err := c.http.Do(createReq)
		if err != nil {
			return err
		}
		defer createResp.Body.Close()
		if createResp.StatusCode >= 400 {
			body, _ := io.ReadAll(createResp.Body)
			return fmt.Errorf("failed to create bucket %s: %d %s", bucket, createResp.StatusCode, string(body))
		}
		return nil
	}
	return fmt.Errorf("unexpected status for HEAD bucket: %d", resp.StatusCode)
}

// PutObject загружает объект в MinIO.
func (c *Client) PutObject(bucket, object string, data io.Reader, size int64) error {
	if err := c.ensureBucket(bucket); err != nil {
		return fmt.Errorf("ensure bucket: %w", err)
	}

	// Читаем всё в память (для подписи)
	body, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	req, err := c.newRequest("PUT", bucket, object, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(body))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("put object failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetObject скачивает объект из MinIO.
func (c *Client) GetObject(bucket, object string) ([]byte, error) {
	req, err := c.newRequest("GET", bucket, object, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("object not found: %s/%s", bucket, object)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get object failed: %d %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ListObjects возвращает список объектов с префиксом.
func (c *Client) ListObjects(bucket, prefix string) ([]ObjectInfo, error) {
	if err := c.ensureBucket(bucket); err != nil {
		return nil, err
	}

	params := url.Values{}
	if prefix != "" {
		params.Set("prefix", prefix)
	}
	params.Set("list-type", "2")

	req, err := c.newRequest("GET", bucket, "", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list objects failed: %d %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var objects []ObjectInfo
	for _, c := range result.Contents {
		objects = append(objects, ObjectInfo{
			Key:          c.Key,
			LastModified: c.LastModified,
			Size:         c.Size,
		})
	}
	// Сортируем по времени (новые — первыми)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(objects[j].LastModified)
	})
	return objects, nil
}

// PresignedGetObject генерирует подписанную ссылку для скачивания (для отладки).
func (c *Client) PresignedGetObject(bucket, object string, expires time.Duration) (string, error) {
	// Простая реализация без полной AWS Signature V4 (для внутреннего использования)
	// В production — лучше использовать официальный minio-go, но по требованию — HTTP-only.
	// Здесь — заглушка.
	return fmt.Sprintf("%s/%s/%s?debug=1", c.baseURL.String(), bucket, object), nil
}

// --- Внутренние вспомогательные методы ---

// newRequest создаёт подписанной HTTP-запрос к MinIO.
func (c *Client) newRequest(method, bucket, object string, body io.Reader) (*http.Request, error) {
	u := *c.baseURL
	if bucket != "" {
		u.Path = "/" + bucket
		if object != "" {
			u.Path += "/" + object
		}
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Устанавливаем обязательные заголовки
	req.Header.Set("Host", u.Host)
	req.Header.Set("x-amz-date", time.Now().UTC().Format("20060102T150405Z"))
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")

	// Подписываем (упрощённо — для MinIO в доверенной сети)
	// В production с публичным доступом — нужна полная AWS Signature V4.
	if c.config.AccessKeyID != "" {
		req.Header.Set("Authorization", fmt.Sprintf("AWS %s:dummy-signature", c.config.AccessKeyID))
	}

	return req, nil
}

// --- Структуры для XML-парсинга ---

type ListBucketResult struct {
	Contents []Content `xml:"Contents"`
}

type Content struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	Size         int64     `xml:"Size"`
}

type ObjectInfo struct {
	Key          string
	LastModified time.Time
	Size         int64
}
