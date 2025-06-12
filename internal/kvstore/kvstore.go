package kvstore

import (
	"strings"
	"sync"
)

type KVStore struct {
	data  map[string]string
	wal   *WAL
	mutex sync.RWMutex
}

func NewKVStore(walPath string) (*KVStore, error) {
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, err
	}

	store := &KVStore{
		data:  make(map[string]string),
		wal:   wal,
		mutex: sync.RWMutex{},
	}

	entries, err := wal.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		store.data[entry.Key] = entry.Value
	}

	return store, nil
}

func (kvs *KVStore) Set(key, value string) error {

	if err := kvs.wal.Append(key, value); err != nil {
		return err
	}

	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()
	kvs.data[key] = value
	return nil
}

func (kvs *KVStore) Get(key string) (string, bool) {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()
	val, ok := kvs.data[key]
	return val, ok
}

func (kvs *KVStore) GetPrefix(prefix string) map[string]string {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()

	result := make(map[string]string)
	for k, v := range kvs.data {
		if strings.HasPrefix(k, prefix) {
			result[k] = v
		}
	}
	return result
}
