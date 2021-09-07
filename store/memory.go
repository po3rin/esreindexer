package store

import (
	"fmt"
	"sync"

	"github.com/po3rin/esreindexer/entity"
)

type MemoryStore struct {
	Store map[string]entity.Task
	mu    sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Store: make(map[string]entity.Task, 0),
	}
}

func (m *MemoryStore) PutTaskInfo(index string, taskID string, numberOfReplicas int, refreshInterval int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := entity.Task{
		Index:            index,
		NumberOfReplicas: numberOfReplicas,
		RefreshInterval:  refreshInterval,
	}

	m.Store[taskID] = t
	return nil
}

func (m *MemoryStore) TaskInfo(taskID string) (numberOfReplicas int, refreshInterval int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.Store[taskID]
	if !ok {
		return 0, 0, fmt.Errorf("taskID %v is not found", taskID)
	}
	return info.NumberOfReplicas, info.RefreshInterval, nil
}

func (m *MemoryStore) AllTaskd() map[string]entity.Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Store
}
