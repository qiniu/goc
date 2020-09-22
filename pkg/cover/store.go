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
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Store persistents the registered service information
type Store interface {
	// Add adds the given service to store
	Add(s Service) error

	// Get returns the registered service information with the given service's name
	Get(name string) []string

	// Get returns all the registered service information as a map
	GetAll() map[string][]string

	// Init cleanup all the registered service information
	Init() error

	// Set stores the services information into internal state
	Set(services map[string][]string)
}

// PersistenceFile is the file to save services address information
const PersistenceFile = "_svrs_address.txt"

// fileStore holds the registered services into memory and persistent to a local file
type fileStore struct {
	mu             sync.RWMutex
	persistentFile string

	memoryStore Store
}

// NewFileStore creates a store using local file
func NewFileStore(persistenceFile string) Store {
	l := &fileStore{
		persistentFile: persistenceFile,
		memoryStore:    NewMemoryStore(),
	}

	if err := l.load(); err != nil {
		log.Fatalf("load failed, file: %s, err: %v", l.persistentFile, err)
	}

	return l
}

// Add adds the given service to file Store
func (l *fileStore) Add(s Service) error {
	l.memoryStore.Add(s)

	// persistent to local store
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.appendToFile(s)
}

// Get returns the registered service information with the given name
func (l *fileStore) Get(name string) []string {
	return l.memoryStore.Get(name)
}

// Get returns all the registered service information
func (l *fileStore) GetAll() map[string][]string {
	return l.memoryStore.GetAll()
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
	var svrsMap = make(map[string][]string, 0)

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
		if len(ss) == 2 {
			if urls, ok := svrsMap[ss[0]]; ok {
				urls = append(urls, ss[1])
				svrsMap[ss[0]] = urls
			} else {
				svrsMap[ss[0]] = []string{ss[1]}
			}
		}
	}

	if err := ns.Err(); err != nil {
		return fmt.Errorf("read file failed, file: %s, err: %v", l.persistentFile, err)
	}

	// set information to memory
	l.memoryStore.Set(svrsMap)
	return nil
}

func (l *fileStore) Set(services map[string][]string) {
	panic("TO BE IMPLEMENTED")
}

func (l *fileStore) appendToFile(s Service) error {
	f, err := os.OpenFile(l.persistentFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(format(s) + "\n")
	if err != nil {
		return err
	}

	f.Sync()
	return nil
}

func format(s Service) string {
	return fmt.Sprintf("%s&%s", s.Name, s.Address)
}

func split(r rune) bool {
	return r == '&'
}

// memoryStore holds the registered services only into memory
type memoryStore struct {
	mu          sync.RWMutex
	servicesMap map[string][]string
}

// NewMemoryStore creates a memory store
func NewMemoryStore() Store {
	return &memoryStore{
		servicesMap: make(map[string][]string, 0),
	}
}

// Add adds the given service to MemoryStore
func (l *memoryStore) Add(s Service) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// load to memory
	if addrs, ok := l.servicesMap[s.Name]; ok {
		for _, addr := range addrs {
			if addr == s.Address {
				log.Printf("service registered already, name: %s, address: %s", s.Name, s.Address)
				return nil
			}
		}
		addrs = append(addrs, s.Address)
		l.servicesMap[s.Name] = addrs
	} else {
		l.servicesMap[s.Name] = []string{s.Address}
	}

	return nil
}

// Get returns the registered service information with the given name
func (l *memoryStore) Get(name string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.servicesMap[name]
}

// Get returns all the registered service information
func (l *memoryStore) GetAll() map[string][]string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.servicesMap
}

// Init cleanup all the registered service information
// and the local persistent file
func (l *memoryStore) Init() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.servicesMap = make(map[string][]string, 0)
	return nil
}

func (l *memoryStore) Set(services map[string][]string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.servicesMap = services
}
