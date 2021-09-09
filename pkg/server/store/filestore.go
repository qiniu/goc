package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type FileStore struct {
	storePath string
	mu        sync.Mutex
}

func NewFileStore(path string) (Store, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return &FileStore{
		storePath: path,
	}, nil
}

func (s *FileStore) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.storePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)

		keyCandidate := items[0]
		value := items[1]

		if key == keyCandidate {
			return value, nil
		}
	}

	if scanner.Err() != nil {
		return "", scanner.Err()
	}

	return "", fmt.Errorf("no key found")
}

func (s *FileStore) Set(key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.storePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var outputLines string
	scanner := bufio.NewScanner(f)
	isFound := false

	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)

		keyCandidate := items[0]

		if key == keyCandidate {
			line = key + " " + value
			isFound = true
		} else {
		}
		outputLines += line + "\n"
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	if !isFound {
		outputLines += key + " " + value + "\n"
	}

	if err := os.Truncate(s.storePath, 0); err != nil {
		return err
	}

	f.Seek(0, os.SEEK_SET)
	f.WriteString(outputLines)
	f.Sync()

	return nil
}

func (s *FileStore) Remove(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.storePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var outputLines string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)

		keyCandidate := items[0]

		if key == keyCandidate {
			// pass
		} else {
			outputLines += line + "\n"
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	if err := os.Truncate(s.storePath, 0); err != nil {
		return err
	}

	f.Seek(0, os.SEEK_SET)
	f.WriteString(outputLines)
	f.Sync()

	return nil
}

func (s *FileStore) Range(pattern string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.storePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	output := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)
		fmt.Println(items)
		keyCandidate := items[0]
		value := items[1]

		if strings.HasPrefix(keyCandidate, pattern) {
			output = append(output, value)
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("no key found")
	} else {
		return output, nil
	}
}

func (s *FileStore) RangeRemove(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.storePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var outputLines string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, " ", 2)

		keyCandidate := items[0]

		if strings.HasPrefix(keyCandidate, pattern) {
			// pass
		} else {
			outputLines += line + "\n"
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	if err := os.Truncate(s.storePath, 0); err != nil {
		return err
	}

	f.Seek(0, os.SEEK_SET)
	f.WriteString(outputLines)
	f.Sync()

	return nil
}
