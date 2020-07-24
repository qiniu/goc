package clients

import (
	"context"
	"github.com/qiniu/goc/pkg/qiniu"
	"os"
	"time"
)

type MockQnClient struct {
	QiniuObjectHandleRes  qiniu.ObjectHandle
	ReadObjectRes         []byte
	ReadObjectErr         error
	ListAllRes            []string
	ListAllErr            error
	GetAccessURLRes       string
	GetArtifactDetailsRes *qiniu.LogHistoryTemplate
	GetArtifactDetailsErr error
	ListSubDirsRes        []string
	ListSubDirsErr        error
}

func (s *MockQnClient) QiniuObjectHandle(key string) qiniu.ObjectHandle {
	return s.QiniuObjectHandleRes
}

func (s *MockQnClient) ReadObject(key string) ([]byte, error) {
	return s.ReadObjectRes, s.ReadObjectErr
}

func (s *MockQnClient) ListAll(ctx context.Context, prefix string, delimiter string) ([]string, error) {
	return s.ListAllRes, s.ListAllErr
}

func (s *MockQnClient) GetAccessURL(key string, timeout time.Duration) string {
	return s.GetAccessURLRes
}

func (s *MockQnClient) GetArtifactDetails(key string) (*qiniu.LogHistoryTemplate, error) {
	return s.GetArtifactDetailsRes, s.GetArtifactDetailsErr
}

func (s *MockQnClient) ListSubDirs(prefix string) ([]string, error) {
	return s.ListSubDirsRes, s.ListSubDirsErr
}

type MockArtifacts struct {
	ProfilePathRes           string
	CreateChangedProfileRes  *os.File
	GetChangedProfileNameRes string
}

func (s *MockArtifacts) ProfilePath() string {
	return s.ProfilePathRes
}
func (s *MockArtifacts) CreateChangedProfile() *os.File {
	return s.CreateChangedProfileRes
}
func (s *MockArtifacts) GetChangedProfileName() string {
	return s.GetChangedProfileNameRes
}
