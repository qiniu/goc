package clients

import (
	"github.com/qiniu/goc/pkg/cover"
)

type MockPrComment struct {
	GetPrChangedFilesRes   []string
	GetPrChangedFilesErr   error
	PostCommentErr         error
	EraseHistoryCommentErr error
	CreateGithubCommentErr error
	CommentFlag            string
}

func (s *MockPrComment) GetPrChangedFiles() (files []string, err error) {
	return s.GetPrChangedFilesRes, s.GetPrChangedFilesErr
}

func (s *MockPrComment) PostComment(content, commentPrefix string) error {
	return s.PostCommentErr
}

func (s *MockPrComment) EraseHistoryComment(commentPrefix string) error {
	return s.EraseHistoryCommentErr
}

func (s *MockPrComment) CreateGithubComment(commentPrefix string, diffCovList cover.DeltaCovList) (err error) {
	return s.CreateGithubCommentErr
}

func (s *MockPrComment) GetCommentFlag() string {
	return s.CommentFlag
}
