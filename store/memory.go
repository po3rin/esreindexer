package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/po3rin/esreindexer/entity"
)

type MemoryStore struct {
	Store          map[string]entity.Task
	expireDuration time.Duration
	mu             sync.Mutex
}

func NewMemoryStore(expireDuration time.Duration) *MemoryStore {
	return &MemoryStore{
		Store:          make(map[string]entity.Task, 0),
		expireDuration: expireDuration,
	}
}

func (m *MemoryStore) PutTaskInfo(index string, taskID string, numberOfReplicas int, refreshInterval int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := entity.Task{
		Index:            index,
		NumberOfReplicas: numberOfReplicas,
		RefreshInterval:  refreshInterval,
		Status:           entity.Running,
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

func (m *MemoryStore) DoneTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	expire := time.Now().Add(m.expireDuration)

	info, ok := m.Store[taskID]
	if !ok {
		return fmt.Errorf("taskID %v is not found", taskID)
	}

	info.Status = entity.Done
	info.ExpireDate = expire
	m.Store[taskID] = info

	return nil
}

func (m *MemoryStore) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Store, taskID)
	return nil
}

func (m *MemoryStore) AllTask() map[string]entity.Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Store
}
