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

package cover

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/cover"
	"k8s.io/test-infra/gopherage/pkg/cov"
)

// LogFile a file to save log.
const LogFile = "goc.log"

type server struct {
	PersistenceFile string
	Store           Store
}

// NewFileBasedServer new a file based server with persistenceFile
func NewFileBasedServer(persistenceFile string) (*server, error) {
	store, err := NewFileStore(persistenceFile)
	if err != nil {
		return nil, err
	}
	return &server{
		PersistenceFile: persistenceFile,
		Store:           store,
	}, nil
}

// NewMemoryBasedServer new a memory based server without persistenceFile
func NewMemoryBasedServer() *server {
	return &server{
		Store: NewMemoryStore(),
	}
}

// Run starts coverage host center
func (s *server) Run(port string) {
	f, err := os.Create(LogFile)
	if err != nil {
		log.Fatalf("failed to create log file %s, err: %v", LogFile, err)
	}

	// both log to stdout and file by default
	mw := io.MultiWriter(f, os.Stdout)
	r := s.Route(mw)
	log.Fatal(r.Run(port))
}

// Router init goc server engine
func (s *server) Route(w io.Writer) *gin.Engine {
	if w != nil {
		gin.DefaultWriter = w
	}
	r := gin.Default()
	// api to show the registered services
	r.StaticFile("static", "./"+s.PersistenceFile)

	v1 := r.Group("/v1")
	{
		v1.POST("/cover/register", s.registerService)
		v1.GET("/cover/profile", s.profile)
		v1.POST("/cover/profile", s.profile)
		v1.POST("/cover/clear", s.clear)
		v1.POST("/cover/init", s.initSystem)
		v1.GET("/cover/list", s.listServices)
		v1.POST("/cover/remove", s.removeServices)

		v1.POST("/cover/gitinfo", s.gitInfo)
		v1.GET("/cover/gitinfo", s.gitInfo)
	}

	return r
}

// ServiceUnderTest is a entry under being tested
type ServiceUnderTest struct {
	Name    string `form:"name" json:"name" binding:"required"`
	Address string `form:"address" json:"address" binding:"required"`
}

// ProfileParam is param of profile API
type ProfileParam struct {
	Force             bool     `form:"force" json:"force"`
	Service           []string `form:"service" json:"service"`
	Address           []string `form:"address" json:"address"`
	CoverFilePatterns []string `form:"coverfile" json:"coverfile"`
}

type GitInfo struct {
	CommitID string
	Branch   string
}

//listServices list all the registered services
func (s *server) listServices(c *gin.Context) {
	services := s.Store.GetAll()
	c.JSON(http.StatusOK, services)
}

func (s *server) registerService(c *gin.Context) {
	var service ServiceUnderTest
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
		log.Printf("the registered host %s of service %s is different with the real one %s, here we choose the real one", service.Name, host, realIP)
		service.Address = fmt.Sprintf("http://%s:%s", realIP, port)
	}

	address := s.Store.Get(service.Name)
	if !contains(address, service.Address) {
		if err := s.Store.Add(service); err != nil && err != ErrServiceAlreadyRegistered {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
	return
}

func RunCommandStr( cmdDir , cmdStr string) (string, error) {
	cmds := strings.Split(strings.TrimSpace(cmdStr), " ")
	var cmd *exec.Cmd
	if ( cmds[0] == "sh" || cmds[0] == "bash" ) &&  cmds[1] == "-c" { // 兼容管道命令
		cmd = exec.Command(cmds[0], cmds[1], strings.Join( cmds[2:]," ")  )
	}else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	}
	cmd.Dir = cmdDir // 指定工作目录为git仓库目录
	//cmd.Stderr = os.Stderr
	msg, err := cmd.CombinedOutput() // 混合输出stdout+stderr
	_ = cmd.Run()
	// 报错时 exit status 1
	return string(msg), err
}

func GetCurrentGitInfo( path  string) ( *GitInfo , error ) {
	cmdStr1 := "sh -c git rev-parse HEAD"
	cmdStr2 := "sh -c git symbolic-ref --short -q HEAD"
	if msg1 , err := RunCommandStr(path,cmdStr1) ; err != nil {
		return nil ,err
	}else  if msg2 , err := RunCommandStr(path,cmdStr2) ; err != nil {
		return nil ,err
	}else {
		return &GitInfo{msg1,msg2} ,nil
	}
}

func (s *server) gitInfo(c *gin.Context) {
	var body ProfileParam
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}
	svrsUnderTest := s.Store.GetAll()
	filterAddrList, err := filterAddrs(body.Service, body.Address, true, svrsUnderTest)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}
	for _, addr := range filterAddrList {
		if info , err := GetCurrentGitInfo(".") ; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(),"addr":addr})
			return
		}else {
			if _, err := fmt.Fprintf(c.Writer, "%s", info); err != nil {
				return
			}
		}
	}
}

// profile API examples:
// POST /v1/cover/profile
// { "force": "true", "service":["a","b"], "address":["c","d"],"coverfile":["e","f"] }
func (s *server) profile(c *gin.Context) {
	var body ProfileParam
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}

	allInfos := s.Store.GetAll()
	filterAddrList, err := filterAddrs(body.Service, body.Address, body.Force, allInfos)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}

	var mergedProfiles = make([][]*cover.Profile, 0)
	for _, addr := range filterAddrList {
		pp, err := NewWorker(addr).Profile(ProfileParam{})
		if err != nil {
			if body.Force {
				log.Warnf("get profile from [%s] failed, error: %s", addr, err.Error())
				continue
			}

			c.JSON(http.StatusExpectationFailed, gin.H{"error": fmt.Sprintf("failed to get profile from %s, error %s", addr, err.Error())})
			return
		}

		profile, err := convertProfile(pp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		mergedProfiles = append(mergedProfiles, profile)
	}

	if len(mergedProfiles) == 0 {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": "no profiles"})
		return
	}

	merged, err := cov.MergeMultipleProfiles(mergedProfiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(body.CoverFilePatterns) > 0 {
		merged, err = filterProfile(body.CoverFilePatterns, merged)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to filter profile based on the patterns: %v, error: %v", body.CoverFilePatterns, err)})
			return
		}
	}

	if err := cov.DumpProfile(merged, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// filterProfile filters profiles of the packages matching the coverFile pattern
func filterProfile(coverFile []string, profiles []*cover.Profile) ([]*cover.Profile, error) {
	var out = make([]*cover.Profile, 0)
	for _, profile := range profiles {
		for _, pattern := range coverFile {
			matched, err := regexp.MatchString(pattern, profile.FileName)
			if err != nil {
				return nil, fmt.Errorf("filterProfile failed with pattern %s for profile %s, err: %v", pattern, profile.FileName, err)
			}
			if matched {
				out = append(out, profile)
				break // no need to check again for the file
			}
		}
	}

	return out, nil
}

func (s *server) clear(c *gin.Context) {
	var body ProfileParam
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}
	svrsUnderTest := s.Store.GetAll()
	filterAddrList, err := filterAddrs(body.Service, body.Address, true, svrsUnderTest)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}
	for _, addr := range filterAddrList {
		pp, err := NewWorker(addr).Clear(ProfileParam{})
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
			return
		}
		fmt.Fprintf(c.Writer, "Register service %s coverage counter %s", addr, string(pp))
	}

}

func (s *server) initSystem(c *gin.Context) {
	if err := s.Store.Init(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (s *server) removeServices(c *gin.Context) {
	var body ProfileParam
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}
	svrsUnderTest := s.Store.GetAll()
	filterAddrList, err := filterAddrs(body.Service, body.Address, true, svrsUnderTest)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
		return
	}
	for _, addr := range filterAddrList {
		err := s.Store.Remove(addr)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"error": err.Error()})
			return
		}
		fmt.Fprintf(c.Writer, "Register service %s removed from the center.", addr)
	}
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

func contains(arr []string, str string) bool {
	for _, element := range arr {
		if str == element {
			return true
		}
	}
	return false
}

// filterAddrs filter address list by given service and address list
func filterAddrs(serviceList, addressList []string, force bool, allInfos map[string][]string) (filterAddrList []string, err error) {
	addressAll := []string{}
	for _, addr := range allInfos {
		addressAll = append(addressAll, addr...)
	}

	if len(serviceList) != 0 && len(addressList) != 0 {
		return nil, fmt.Errorf("use 'service' flag and 'address' flag at the same time may cause ambiguity, please use them separately")
	}

	// Add matched services to map
	for _, name := range serviceList {
		if addr, ok := allInfos[name]; ok {
			filterAddrList = append(filterAddrList, addr...)
			continue // jump to match the next service
		}
		if !force {
			return nil, fmt.Errorf("service [%s] not found", name)
		}
		log.Warnf("service [%s] not found", name)
	}

	// Add matched addresses to map
	for _, addr := range addressList {
		if contains(addressAll, addr) {
			filterAddrList = append(filterAddrList, addr)
			continue
		}
		if !force {
			return nil, fmt.Errorf("address [%s] not found", addr)
		}
		log.Warnf("address [%s] not found", addr)
	}

	if len(addressList) == 0 && len(serviceList) == 0 {
		filterAddrList = addressAll
	}

	// Return all services when all param is nil
	return filterAddrList, nil
}
