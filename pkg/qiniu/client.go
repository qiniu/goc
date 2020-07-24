/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package qiniu

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/client"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/sirupsen/logrus"
)

// Config store the credentials to connect with qiniu cloud
type Config struct {
	Bucket    string `json:"bucket"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`

	// domain used to download files from qiniu cloud
	Domain string `json:"domain"`
}

type Client interface {
	QiniuObjectHandle(key string) ObjectHandle
	ReadObject(key string) ([]byte, error)
	ListAll(ctx context.Context, prefix string, delimiter string) ([]string, error)
	GetAccessURL(key string, timeout time.Duration) string
	GetArtifactDetails(key string) (*LogHistoryTemplate, error)
	ListSubDirs(prefix string) ([]string, error)
}

// QnClient for the operation with qiniu cloud
type QnClient struct {
	cfg           *Config
	BucketManager *storage.BucketManager
}

// NewClient creates a new QnClient to work with qiniu cloud
func NewClient(cfg *Config) *QnClient {
	return &QnClient{
		cfg:           cfg,
		BucketManager: storage.NewBucketManager(qbox.NewMac(cfg.AccessKey, cfg.SecretKey), nil),
	}
}

// QiniuObjectHandle construct a object hanle to access file in qiniu
func (q *QnClient) QiniuObjectHandle(key string) ObjectHandle {
	return &QnObjectHandle{
		key:    key,
		cfg:    q.cfg,
		bm:     q.BucketManager,
		mac:    qbox.NewMac(q.cfg.AccessKey, q.cfg.SecretKey),
		client: &client.Client{Client: http.DefaultClient},
	}
}

// ReadObject to read all the content of key
func (q *QnClient) ReadObject(key string) ([]byte, error) {
	objectHandle := q.QiniuObjectHandle(key)
	reader, err := objectHandle.NewReader(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting qiniu artifact reader: %v", err)
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

// ListAll to list all the files with contains the expected prefix
func (q *QnClient) ListAll(ctx context.Context, prefix string, delimiter string) ([]string, error) {
	var files []string
	artifacts, err := q.listEntries(prefix, delimiter)
	if err != nil {
		return files, err
	}

	for _, item := range artifacts {
		files = append(files, item.Key)
	}

	return files, nil
}

// ListAll to list all the entries with contains the expected prefix
func (q *QnClient) listEntries(prefix string, delimiter string) ([]storage.ListItem, error) {
	var marker string
	var artifacts []storage.ListItem

	wait := []time.Duration{16, 32, 64, 128, 256, 256, 512, 512}
	for i := 0; ; {
		entries, _, nextMarker, hashNext, err := q.BucketManager.ListFiles(q.cfg.Bucket, prefix, delimiter, marker, 500)
		if err != nil {
			logrus.WithField("prefix", prefix).WithError(err).Error("Error accessing QINIU artifact.")
			if i >= len(wait) {
				return artifacts, fmt.Errorf("timed out: error accessing QINIU artifact: %v", err)
			}
			time.Sleep((wait[i] + time.Duration(rand.Intn(10))) * time.Millisecond)
			i++
			continue
		}
		artifacts = append(artifacts, entries...)

		if hashNext {
			marker = nextMarker
		} else {
			break
		}
	}

	return artifacts, nil
}

// GetAccessURL return a url which can access artifact directly in qiniu
func (q *QnClient) GetAccessURL(key string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout).Unix()
	return storage.MakePrivateURL(qbox.NewMac(q.cfg.AccessKey, q.cfg.SecretKey), q.cfg.Domain, key, deadline)
}

type LogHistoryTemplate struct {
	BucketName string
	KeyPath    string
	Items      []logHistoryItem
}

type logHistoryItem struct {
	Name string
	Size string
	Time string
	Url  string
}

// Artifacts lists all artifacts available for the given job source
func (q *QnClient) GetArtifactDetails(key string) (*LogHistoryTemplate, error) {
	tmpl := new(LogHistoryTemplate)
	item := logHistoryItem{}
	listStart := time.Now()
	artifacts, err := q.listEntries(key, "")
	if err != nil {
		return tmpl, err
	}

	for _, entry := range artifacts {
		item.Name = splitKey(entry.Key, key)
		item.Size = size(entry.Fsize)
		item.Time = timeConv(entry.PutTime)
		item.Url = q.GetAccessURL(entry.Key, time.Duration(time.Second*60*60))
		tmpl.Items = append(tmpl.Items, item)
	}

	logrus.WithField("duration", time.Since(listStart).String()).Infof("Listed %d artifacts.", len(tmpl.Items))
	return tmpl, nil
}

func splitKey(item, key string) string {
	return strings.TrimPrefix(item, key)
}

func size(fsize int64) string {
	return strings.Join([]string{strconv.FormatInt(fsize, 10), "bytes"}, " ")
}

func timeConv(ptime int64) string {
	s := strconv.FormatInt(ptime, 10)[0:10]
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logrus.Errorf("time string parse int error : %v", err)
		return ""
	}
	tm := time.Unix(t, 0)
	return tm.Format("2006-01-02 03:04:05 PM")
}

func (q *QnClient) ListSubDirs(prefix string) ([]string, error) {
	var dirs []string
	var marker string

	wait := []time.Duration{16, 32, 64, 128, 256, 256, 512, 512}
	for i := 0; ; {
		// use rsf list v2 interface to get the sub folder based on the delimiter
		entries, err := q.BucketManager.ListBucketContext(context.Background(), q.cfg.Bucket, prefix, "/", marker)
		if err != nil {
			logrus.WithField("prefix", prefix).WithError(err).Error("Error accessing QINIU artifact.")
			if i >= len(wait) {
				return dirs, fmt.Errorf("timed out: error accessing QINIU artifact: %v", err)
			}
			time.Sleep((wait[i] + time.Duration(rand.Intn(10))) * time.Millisecond)
			i++
			continue
		}

		for entry := range entries {
			if entry.Dir != "" {
				// entry.Dir should be like "logs/kodo-periodics-integration-test/1181915661132107776/"
				// the sub folder is 1181915661132107776, also known as prowjob buildid.
				buildId := getBuildId(entry.Dir)
				if buildId != "" {
					dirs = append(dirs, buildId)
				} else {
					logrus.Warnf("invalid dir format: %v", entry.Dir)
				}
			}

			marker = entry.Marker
		}

		if marker != "" {
			i = 0
		} else {
			break
		}
	}

	return dirs, nil
}

var nonPRLogsBuildIdSubffixRe = regexp.MustCompile("([0-9]+)/$")

// extract the build number from dir path
// expect the dir as the following formats:
// 1. logs/kodo-periodics-integration-test/1181915661132107776/
func getBuildId(dir string) string {
	matches := nonPRLogsBuildIdSubffixRe.FindStringSubmatch(dir)
	if len(matches) == 2 {
		return matches[1]
	}

	return ""
}
