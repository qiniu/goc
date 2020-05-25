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
	"log"
	"os"
	"strings"
	"sync"
)

// Store persistents the registered service information
type Store interface {
	// Add adds the given service to store
	Add(s Service) error

	// Get returns the registered service informations with the given service's name
	Get(name string) []string

	// Get returns all the registered service informations as a map
	GetAll() map[string][]string

	// Init cleanup all the registered service informations
	Init() error
}

// PersistenceFile is the file to save services address information
const PersistenceFile = "_svrs_address.txt"

// localStore holds the registered services into memory and persistent to a local file
type localStore struct {
	mu             sync.RWMutex
	servicesMap    map[string][]string
	persistentFile string
}

// Add adds the given service to localStore
func (l *localStore) Add(s Service) error {
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

	// persistent to local sotre
	return l.appendToFile(s)
}

// Get returns the registered service informations with the given name
func (l *localStore) Get(name string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.servicesMap[name]
}

// Get returns all the registered service informations
func (l *localStore) GetAll() map[string][]string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.servicesMap
}

// Init cleanup all the registered service informations
// and the local persistent file
func (l *localStore) Init() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := os.Remove(l.persistentFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s, err: %v", l.persistentFile, err)
	}

	l.servicesMap = make(map[string][]string, 0)
	return nil
}

// load all registered servcie from file to memory
func (l *localStore) load() (map[string][]string, error) {
	var svrsMap = make(map[string][]string, 0)

	f, err := os.Open(l.persistentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return svrsMap, nil
		}
		return svrsMap, fmt.Errorf("failed to open file, path: %s, err: %v", l.persistentFile, err)
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
		return svrsMap, fmt.Errorf("read file failed, file: %s, err: %v", l.persistentFile, err)
	}

	return svrsMap, nil
}

func (l *localStore) appendToFile(s Service) error {
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

// NewStore creates a store using local file
func NewStore() Store {
	l := &localStore{
		persistentFile: PersistenceFile,
		servicesMap:    make(map[string][]string, 0),
	}

	services, err := l.load()
	if err != nil {
		log.Fatalf("load failed, file: %s, err: %v", l.persistentFile, err)
	}
	l.servicesMap = services
	return l
}
