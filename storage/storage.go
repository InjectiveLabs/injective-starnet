package storage

import (
	"encoding/json"
	"os"
)

type Storage interface {
	Get(key string) (Record, error)
	Set(key string, record Record) error
	GetAll() ([]Record, error)
	SetAll(records []Record) error
	DeleteAll() error
}

type Record struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	ID       string `json:"id"`
}

// FileStore is a simple file-based storage implementation.
const defaultPath = "./storage.json"

type FileStore struct {
	Path string
}

func NewFileStore(path string) *FileStore {
	if path == "" {
		path = defaultPath
	}

	return &FileStore{Path: path}
}

func (s *FileStore) GetAll() ([]Record, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}

	// If file is empty, return empty slice
	if len(data) == 0 {
		return []Record{}, nil
	}

	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *FileStore) SetAll(records []Record) error {
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}

	return os.WriteFile(s.Path, data, 0644)
}

func (s *FileStore) DeleteAll() error {
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(s.Path)
}

func (s *FileStore) Get(key string) (Record, error) {
	panic("not implemented")
}

func (s *FileStore) Set(key string, record Record) error {
	panic("not implemented")
}
