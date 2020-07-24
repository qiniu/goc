package clients

import (
	"context"
	"github.com/qiniu/goc/pkg/qiniu"
	"time"
)

type MockQnClient struct {
}

func (s *MockQnClient) QiniuObjectHandle(key string) qiniu.ObjectHandle {
	return nil
}

func (s *MockQnClient) ReadObject(key string) ([]byte, error) {
	return nil, nil
}

func (s *MockQnClient) ListAll(ctx context.Context, prefix string, delimiter string) ([]string, error) {
	return nil, nil
}

func (s *MockQnClient) GetAccessURL(key string, timeout time.Duration) string {
	return ""
}

func (s *MockQnClient) GetArtifactDetails(key string) (*qiniu.LogHistoryTemplate, error) {
	return nil, nil
}

func (s *MockQnClient) ListSubDirs(prefix string) ([]string, error) {
	return nil, nil
}
