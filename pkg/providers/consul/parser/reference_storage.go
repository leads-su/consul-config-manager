package parser

import (
	"fmt"
	"sync"
)

type ReferenceStorage struct {
	sync.RWMutex
	pathToKey map[string]string
	keyToPath map[string]string
}

// NewReferenceStorage creates new instance of reference storage
func NewReferenceStorage() *ReferenceStorage {
	return &ReferenceStorage{
		pathToKey: make(map[string]string),
		keyToPath: make(map[string]string),
	}
}

// Set add value to references storage
func (storage *ReferenceStorage) Set(path, key string) {
	storage.setPathToKeyReference(path, key)
	storage.setKeyToPathReference(key, path)
}

// Get retrieve value from references storage
func (storage *ReferenceStorage) Get(pathOrKey string) (string, error) {
	if storage.pathToKeyHas(pathOrKey) {
		return storage.pathToKey[pathOrKey], nil
	}
	if storage.keyToPathHas(pathOrKey) {
		return storage.keyToPath[pathOrKey], nil
	}
	return "", fmt.Errorf("unable to find reference by path or key")
}

// remove remove data from references storage
func (storage *ReferenceStorage) remove(path, key string) {
	storage.removePathToKeyReference(path)
	storage.removeKeyToPathReference(key)
}

// setPathToKeyReference set path to key reference
func (storage *ReferenceStorage) setPathToKeyReference(path, key string) {
	storage.Lock()
	defer storage.Unlock()
	storage.pathToKey[path] = key
}

// removePathToKeyReference remove path to key reference
func (storage *ReferenceStorage) removePathToKeyReference(path string) {
	storage.Lock()
	defer storage.Unlock()
	if storage.pathToKeyHas(path) {
		delete(storage.pathToKey, path)
	}
}

// pathToKeyHas check if path to key map has key
func (storage *ReferenceStorage) pathToKeyHas(path string) bool {
	if _, ok := storage.pathToKey[path]; ok {
		return true
	}
	return false
}

// pathToKeyGet get value from pathToKey map by path value
func (storage *ReferenceStorage) pathToKeyGet(path string) (string, error) {
	if storage.pathToKeyHas(path) {
		return storage.pathToKey[path], nil
	}
	return "", fmt.Errorf("requested path `%s` does not exists in pathToKey map", path)
}

// pathToKeyList list pathToKey map values
func (storage *ReferenceStorage) pathToKeyList() map[string]string {
	return storage.pathToKey
}

// setKeyToPathReference set key to path reference
func (storage *ReferenceStorage) setKeyToPathReference(key, path string) {
	storage.Lock()
	defer storage.Unlock()
	storage.keyToPath[key] = path
}

// removeKeyToPathReference remove key to path reference
func (storage *ReferenceStorage) removeKeyToPathReference(key string) {
	storage.Lock()
	defer storage.Unlock()
	if storage.keyToPathHas(key) {
		delete(storage.keyToPath, key)
	}
}

// keyToPathHas check if key to path map has key
func (storage *ReferenceStorage) keyToPathHas(key string) bool {
	if _, ok := storage.keyToPath[key]; ok {
		return true
	}
	return false
}

// keyToPathGet get value from keyToPath map by key value
func (storage *ReferenceStorage) keyToPathGet(key string) (string, error) {
	if storage.keyToPathHas(key) {
		return storage.keyToPath[key], nil
	}
	return "", fmt.Errorf("requested key `%s` does not exists in keyToPath map", key)
}

// keyToPathList list keyToPath map values
func (storage *ReferenceStorage) keyToPathList() map[string]string {
	return storage.keyToPath
}
