package jstore

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

type Store struct {
	Dir *os.File
}

func NewStore(dir string) (*Store, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	return &Store{Dir: f}, nil
}

// LoadLatest lists all the files in the store directory and returns the latest
func (s *Store) LoadLatest() (*FileData, error) {
	files, err := s.Dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files in store")
	}

	// sort by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// load the latest file
	return s.Load(files[len(files)-1].Name())
}

// Save file with timestamp in the name
func (s *Store) Save(data *FileData) error {
	ts := time.Now().Format("2006_01_02_15_04_05")
	f, err := os.Create(fmt.Sprintf("%s/%s.json", s.Dir.Name(), ts))
	if err != nil {
		return err
	}
	defer f.Close()

	end := json.NewEncoder(f)
	end.SetIndent("", "  ")
	return end.Encode(data)
}

func (s *Store) Load(name string) (*FileData, error) {
	b, err := os.ReadFile(s.Dir.Name() + "/" + name)
	if err != nil {
		return nil, err
	}

	var data FileData
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
