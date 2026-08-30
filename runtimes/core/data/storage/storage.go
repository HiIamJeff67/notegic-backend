package storage

import (
	"context"
	"io"
	"time"
)

type PutOptions struct {
	ContentType string
	Metadata    map[string]string
}

type GetOptions struct {
}

type PresignOptions struct {
	Expires     time.Duration
	ContentType string
}

type Object struct {
	Key            string
	Data           []byte
	Size           int64
	ContentType    string
	ParseMediaType string
	LastModified   time.Time
	ETag           string
}

type StorageInterface interface {
	ListKeys() []string
	GetKey(ownerIndicator string, objectIndicator string, salt string) string
	NewObject(key string, reader io.Reader, size int64) (*Object, error)
	PutObjectByKey(ctx context.Context, key string, object *Object) error
	GetObjectByKey(ctx context.Context, key string, option *GetOptions) (io.ReadCloser, *Object, error)
	DeleteObjectByKey(ctx context.Context, key string) error
	PresignPutObjectByKey(ctx context.Context, key string, option *PresignOptions) (string, error)
	PresignGetObjectByKey(ctx context.Context, key string, option *PresignOptions) (string, error)
}
