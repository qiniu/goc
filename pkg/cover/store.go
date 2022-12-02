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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")

const (
	timeFormat      = "2006-01-02T15:04:05.000Z07:00"
	timeoutDuration = time.Second * 20
)

type serviceAddress struct {
	Address string
	Active  time.Time
}

// isTimeout if timeout should be evicted
func (s *serviceAddress) isTimeout() bool {
	return time.Since(s.Active) > timeoutDuration
}

// Store persistents the registered service information
type Store interface {
	// Add adds the given service to store
	Add(s ServiceUnderTest) error

	// Get returns the registered service information with the given service's name
	Get(name string) []string

	// Get returns all the registered service information as a map
	GetAll() map[string][]string

	// Init cleanup all the registered service information
	Init() error

	// Set stores the services information into internal state
	Set(services map[string][]*serviceAddress) error

	// Remove the service from the store by address
	Remove(addr string) error

	//GetRaw return all registered information
	GetRaw() map[string][]*serviceAddress

	// Evict delete all addresses which is timeout.
	// return true if any address removed.
	Evict() (bool, error)
}

// fileStore holds the registered services into memory and persistent to a local file
type fileStore struct {
	mu             sync.RWMutex
	persistentFile string

	memoryStore Store
}

// NewFileStore creates a store using local file
func NewFileStore(persistenceFile string) (store Store, err error) {
	path, err := filepath.Abs(persistenceFile)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, err
	}
	l := &fileStore{
		persistentFile: path,
		memoryStore:    NewMemoryStore(),
	}

	if err := l.load(); err != nil {
		log.Fatalf("load failed, file: %s, err: %v", l.persistentFile, err)
	}

	return l, nil
}

// Add adds the given service to file Store
func (l *fileStore) Add(s ServiceUnderTest) error {
	if err := l.memoryStore.Add(s); err != nil {
		return err
	}

	// persistent to local store
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.save()
}

// Get returns the registered service information with the given name
func (l *fileStore) Get(name string) []string {
	return l.memoryStore.Get(name)
}

// Get returns all the registered service information
func (l *fileStore) GetAll() map[string][]string {
	return l.memoryStore.GetAll()
}

// Remove the service from the memory store and the file store
func (l *fileStore) Remove(addr string) error {
	err := l.memoryStore.Remove(addr)
	if err != nil {
		return err
	}

	// persistent to local store
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.save()
}

// GetRaw return all registered information
func (l *fileStore) GetRaw() map[string][]*serviceAddress {
	return l.memoryStore.GetRaw()
}

// Evict delete all addresses which is timeout.
// return true if any address removed.
func (l *fileStore) Evict() (bool, error) {
	removed, err := l.memoryStore.Evict()
	if err != nil {
		return removed, err
	}
	if !removed {
		return false, nil
	}

	// persistent to local store
	l.mu.Lock()
	defer l.mu.Unlock()
	return true, l.save()
}

// Init cleanup all the registered service information
// and the local persistent file
func (l *fileStore) Init() error {
	if err := l.memoryStore.Init(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if err := os.Remove(l.persistentFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s, err: %v", l.persistentFile, err)
	}

	return nil
}

// load all registered service from file to memory
func (l *fileStore) load() error {
	var svrsMap = make(map[string][]*serviceAddress, 0)

	f, err := os.Open(l.persistentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open file, path: %s, err: %v", l.persistentFile, err)
	}
	defer f.Close()

	ns := bufio.NewScanner(f)
	for ns.Scan() {
		line := ns.Text()
		ss := strings.FieldsFunc(line, split)

		// TODO: use regex
		if len(ss) == 3 {
			name := ss[0]
			addr := ss[1]
			active, er := time.ParseInLocation(timeFormat, ss[2], time.Local)
			if er != nil {
				continue
			}
			if urls, ok := svrsMap[name]; ok {
				urls = append(urls, &serviceAddress{Address: addr, Active: active})
				svrsMap[name] = urls
			} else {
				svrsMap[name] = []*serviceAddress{{Address: addr, Active: active}}
			}
		}
	}

	if err := ns.Err(); err != nil {
		return fmt.Errorf("read file failed, file: %s, err: %v", l.persistentFile, err)
	}

	// set information to memory
	return l.memoryStore.Set(svrsMap)
}

// save all registered service to file
func (l *fileStore) save() error {
	f, err := os.OpenFile(l.persistentFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	s := ""
	for name, addrs := range l.memoryStore.GetRaw() {
		for _, addr := range addrs {
			s += fmt.Sprintf("%s&%s&%s\n", name, addr.Address, addr.Active.In(time.Local).Format(timeFormat))
		}
	}

	_, err = f.WriteString(s)
	if err != nil {
		return err
	}

	return f.Sync()
}

func (l *fileStore) Set(services map[string][]*serviceAddress) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// no error will return from memorystore.set
	err := l.memoryStore.Set(services)
	if err != nil {
		return err
	}

	return l.save()
}

func split(r rune) bool {
	return r == '&'
}

// memoryStore holds the registered services only into memory
type memoryStore struct {
	mu          sync.RWMutex
	servicesMap map[string][]*serviceAddress
}

// NewMemoryStore creates a memory store
func NewMemoryStore() Store {
	return &memoryStore{
		servicesMap: make(map[string][]*serviceAddress, 0),
	}
}

// Add adds the given service to MemoryStore
func (l *memoryStore) Add(s ServiceUnderTest) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// load to memory
	if addrs, ok := l.servicesMap[s.Name]; ok {
		for _, addr := range addrs {
			if addr.Address == s.Address {
				addr.Active = time.Now()
				return ErrServiceAlreadyRegistered
			}
		}
		addrs = append(addrs, &serviceAddress{Address: s.Address, Active: time.Now()})
		l.servicesMap[s.Name] = addrs
	} else {
		l.servicesMap[s.Name] = []*serviceAddress{{Address: s.Address, Active: time.Now()}}
	}

	return nil
}

// Get returns the registered service information with the given name
func (l *memoryStore) Get(name string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	v := l.servicesMap[name]
	if v == nil {
		return nil
	}
	addrs := make([]string, 0, len(v))
	for _, addr := range v {
		addrs = append(addrs, addr.Address)
	}
	return addrs
}

// Get returns all the registered service information
func (l *memoryStore) GetAll() map[string][]string {
	res := make(map[string][]string)
	l.mu.RLock()
	defer l.mu.RUnlock()
	for k, v := range l.servicesMap {
		addrs := make([]string, 0, len(v))
		for _, addr := range v {
			addrs = append(addrs, addr.Address)
		}
		res[k] = addrs
	}
	return res
}

// Init cleanup all the registered service information
// and the local persistent file
func (l *memoryStore) Init() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.servicesMap = make(map[string][]*serviceAddress, 0)
	return nil
}

func (l *memoryStore) Set(services map[string][]*serviceAddress) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.servicesMap = services

	return nil
}

// Remove one service from the memory store
// if service is not fount, return "no service found" error
func (l *memoryStore) Remove(removeAddr string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	flag := false
	for name, addrs := range l.servicesMap {
		newAddrs := make([]*serviceAddress, 0, len(addrs))
		for _, addr := range addrs {
			if removeAddr != addr.Address {
				newAddrs = append(newAddrs, addr)
			} else {
				flag = true
			}
		}
		// if no services left, remove by name
		if len(newAddrs) == 0 {
			delete(l.servicesMap, name)
		} else {
			l.servicesMap[name] = newAddrs
		}
	}

	if !flag {
		return fmt.Errorf("no service found")
	}

	return nil
}

// GetRaw return all registered information
func (l *memoryStore) GetRaw() map[string][]*serviceAddress {
	return l.servicesMap
}

// Evict delete all addresses which is timeout.
// return true if any address removed.
func (l *memoryStore) Evict() (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	flag := false
	for name, addrs := range l.servicesMap {
		newAddrs := make([]*serviceAddress, 0, len(addrs))
		for _, addr := range addrs {
			if !addr.isTimeout() {
				newAddrs = append(newAddrs, addr)
			} else {
				flag = true
			}
		}
		// if no services left, remove by name
		if len(newAddrs) == 0 {
			delete(l.servicesMap, name)
		} else {
			l.servicesMap[name] = newAddrs
		}
	}

	return flag, nil
}
