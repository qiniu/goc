package clients

import "github.com/qiniu/goc/pkg/cover"

type MockPrComment struct {
}

func (s *MockPrComment) GetPrChangedFiles() (files []string, err error) {
	return nil, nil
}
func (s *MockPrComment) PostComment(content, commentPrefix string) error {
	return nil
}
func (s *MockPrComment) EraseHistoryComment(commentPrefix string) error {
	return nil
}

func (s *MockPrComment) CreateGithubComment(commentPrefix string, diffCovList cover.DeltaCovList) (err error) {
	return nil
}

func (s *MockPrComment) GetCommentFlag() string {
	return ""
}
