package kvstore

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type WALEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type WAL struct {
	file  *os.File
	mutex sync.Mutex
}

func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file:  file,
		mutex: sync.Mutex{},
	}, nil
}

func (w *WAL) Append(key, value string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	entry := WALEntry{Key: key, Value: value}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := w.file.Write(append(data, '\n')); err != nil {
		return err
	}

	return w.file.Sync()
}

func (w *WAL) ReadAll() ([]WALEntry, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var entries []WALEntry
	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		var entry WALEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}
