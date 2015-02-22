package cache

import "sync"

type InMemory struct {
	hash map[string][]byte
	mu   sync.RWMutex
}

func NewInMemory() *InMemory {
	return &InMemory{hash: make(map[string][]byte)}
}

func (im *InMemory) Get(key string) ([]byte, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if v, found := im.hash[key]; found {
		return v, nil
	}

	return []byte{}, ErrNotFound
}

func (im *InMemory) Set(key string, value []byte) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, found := im.hash[key]; found {
		return ErrAlreadyExists
	}

	im.hash[key] = value
	return nil
}

func (im *InMemory) Remove(key string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, found := im.hash[key]; found {
		delete(im.hash, key)
		return nil
	}

	return ErrNotFound
}

func (im *InMemory) Update(key string, value []byte) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, found := im.hash[key]; found {
		im.hash[key] = value
		return nil
	}

	return ErrNotFound
}

func (im *InMemory) Keys() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	cumul := []string{}
	for k, _ := range im.hash {
		cumul = append(cumul, k)
	}
	return cumul
}
