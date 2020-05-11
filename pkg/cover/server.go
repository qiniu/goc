/*
 Copyright 2020 Qiniu Cloud (七牛云)

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

package cover

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/gopherage/pkg/cov"
)

// LocalStore implements the IPersistence interface
var LocalStore Store

// Client implements the Action interface
var Client Action

// LogFile a file to save log.
const LogFile = "goc.log"

// StartServer starts coverage host center
func StartServer(port string) {
	LocalStore = NewStore()
	Client = NewWorker()

	f, err := os.Create(LogFile)
	if err != nil {
		log.Fatalf("failed to create log file %s, err: %v", LogFile, err)
	}

	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	r := gin.Default()

	// api to show the registerd services
	r.StaticFile(PersistenceFile, "./"+PersistenceFile)

	v1 := r.Group("/v1")
	{
		v1.POST("/cover/register", registerService)
		v1.GET("/cover/profile", profile)
		v1.POST("/cover/clear", clear)
		v1.POST("/cover/init", initSystem)
	}

	log.Fatal(r.Run(port))
}

// Service is a entry under being tested
type Service struct {
	Name    string `form:"name" json:"name" binding:"required"`
	Address string `form:"address" json:"address" binding:"required"`
}

func registerService(c *gin.Context) {
	var service Service
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := url.Parse(service.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	realIP := c.ClientIP()
	if host != realIP {
		log.Printf("the registed host %s of service %s is different with the real one %s, here we choose the real one", service.Name, host, realIP)
		service.Address = fmt.Sprintf("http://%s:%s", realIP, port)
	}
	if err := LocalStore.Add(service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": service.Name, "address": service.Address})
}

func profile(c *gin.Context) {
	svrsUnderTest := LocalStore.GetAll()
	var mergedProfiles = make([][]*cover.Profile, len(svrsUnderTest))
	for _, addrs := range svrsUnderTest {
		for _, addr := range addrs {
			pp, err := Client.Profile(addr)
			if err != nil {
				c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
				return
			}
			profile, err := convertProfile(pp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			mergedProfiles = append(mergedProfiles, profile)
		}
	}

	if len(mergedProfiles) == 0 {
		c.JSON(http.StatusOK, "no profiles")
		return
	}

	merged, err := cov.MergeMultipleProfiles(mergedProfiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := cov.DumpProfile(merged, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func clear(c *gin.Context) {
	if err := LocalStore.Init(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := Client.Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, "TO BE IMPLEMENTED")
}

func initSystem(c *gin.Context) {
	if err := LocalStore.Init(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: 需要让目标被测服务，重新注册吗？如果需要的话，还需要提供接口？会不会考虑的过于复杂了？
	c.JSON(http.StatusOK, "")
}

func convertProfile(p []byte) ([]*cover.Profile, error) {
	// Annoyingly, ParseProfiles only accepts a filename, so we have to write the bytes to disk
	// so it can read them back.
	// We could probably also just give it /dev/stdin, but that'll break on Windows.
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file, err: %v", err)
	}
	defer tf.Close()
	defer os.Remove(tf.Name())
	if _, err := io.Copy(tf, bytes.NewReader(p)); err != nil {
		return nil, fmt.Errorf("failed to copy data to temp file, err: %v", err)
	}

	return cover.ParseProfiles(tf.Name())
}
