package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
)

/* ============================== Interface & Constructor ============================== */

type InMemoryObject struct {
	Data           []byte
	ContentType    string
	ParseMediaType string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ETag           string
}

type inMemoryStorage struct {
	storageMutex sync.RWMutex
	data         map[string]*InMemoryObject
}

func NewInMemoryStorage() StorageInterface {
	return &inMemoryStorage{
		data: map[string]*InMemoryObject{},
	}
}

/* ============================== Helper Functions ============================== */

func (s *inMemoryStorage) ListKeys() []string {
	s.storageMutex.RLock()
	defer s.storageMutex.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	return keys
}

func (s *inMemoryStorage) GetKey(ownerIndicator string, objectIndicator string, salt string) string {
	origin := "In-Memory-Key<" + ownerIndicator + "|" + objectIndicator + "|" + salt + ">"
	hash := sha256.Sum256([]byte(origin))
	return hex.EncodeToString(hash[:])
}

func (s *inMemoryStorage) GenerateETag(data []byte) string {
	return "In-Memory-ETag<" + string(int32(len(data))) + ">" + time.Now().String()
}

func (s *inMemoryStorage) NewObject(key string, reader io.Reader, size int64) (*Object, error) {
	if size > sconstants.MaxInMemoryStorageFileSize.ToInt64() {
		return nil, fmt.Errorf("object size %d exceeds limit %d", size, sconstants.MaxInMemoryStorageFileSize.ToInt64())
	}

	limitReader := io.LimitReader(reader, sconstants.MaxInMemoryStorageFileSize.ToInt64()+1)
	b, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("read object bytes: %w", err)
	}

	actualSize := int64(len(b))
	if actualSize > sconstants.MaxInMemoryStorageFileSize.ToInt64() {
		return nil, fmt.Errorf("object size %d exceeds limit %d", actualSize, sconstants.MaxInMemoryStorageFileSize.ToInt64())
	}

	contentTypes := strings.Split(http.DetectContentType(b), "; ")
	if len(contentTypes) == 0 {
		return nil, fmt.Errorf("detect object content type")
	}

	eTag := s.GenerateETag(b)
	now := time.Now()

	object := Object{
		Key:            key,
		Data:           b,
		Size:           actualSize,
		ContentType:    contentTypes[0],
		ParseMediaType: "charset=utf-8",
		LastModified:   now,
		ETag:           eTag,
	}
	if len(contentTypes) > 1 {
		object.ParseMediaType = contentTypes[1]
	}

	return &object, nil
}

func (s *inMemoryStorage) PutObjectByKey(ctx context.Context, key string, object *Object) error {
	s.storageMutex.Lock()
	s.data[key] = &InMemoryObject{
		Data:        object.Data,
		ContentType: object.ContentType,
		CreatedAt:   object.LastModified,
		UpdatedAt:   object.LastModified,
		ETag:        object.ETag,
	}
	s.storageMutex.Unlock()

	return nil
}

func (s *inMemoryStorage) GetObjectByKey(ctx context.Context, key string, option *GetOptions) (io.ReadCloser, *Object, error) {
	s.storageMutex.RLock()
	object, ok := s.data[key]
	s.storageMutex.RUnlock()
	if !ok {
		return nil, nil, fmt.Errorf("object %q not found", key)
	}

	rc := io.NopCloser(bytes.NewReader(object.Data))
	metadata := &Object{
		Key:          key,
		Data:         object.Data,
		Size:         int64(len(object.Data)),
		ContentType:  object.ContentType,
		LastModified: object.UpdatedAt,
		ETag:         object.ETag,
	}

	return rc, metadata, nil
}

func (s *inMemoryStorage) DeleteObjectByKey(ctx context.Context, key string) error {
	s.storageMutex.Lock()
	defer s.storageMutex.Unlock()
	if _, ok := s.data[key]; !ok {
		return fmt.Errorf("object %q not found", key)
	}
	delete(s.data, key)
	return nil
}

// [not implemented] For Testing：return fake URL
func (s *inMemoryStorage) PresignPutObjectByKey(ctx context.Context, key string, option *PresignOptions) (string, error) {
	return "http://localhost:" + "/" + cgateway.APIDevelopmentBaseURL + "/" + "storage/mock://put/" + key, nil
}

// For Testing：return localhost URL, give the frontend ability to visit
func (s *inMemoryStorage) PresignGetObjectByKey(ctx context.Context, key string, option *PresignOptions) (string, error) {
	return "http://localhost:" + "/" + cgateway.APIDevelopmentBaseURL + "/" + "storage/mock/files/" + key, nil
}
