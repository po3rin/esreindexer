package esreindexer

import (
	"context"
)

type Client interface {
	Reindex(ctx context.Context, src string, dest string) (string, error)
	GetIndexSetting(ctx context.Context, index string) (numberOfReplicas int, refreshInterval int, err error)
	UpdateIndexSetting(ctx context.Context, index string, numberOfReplicas int, refreshInterval int) error
}

type Store interface {
	PutTaskInfo(id string, NumberOfReplicas int, RefreshInterval int) error
	TaskInfo(id string) (numberOfReplicas int, refreshInterval int, err error)
}

type ReindexManager struct {
	client Client
	store  Store
}

func NewReindexManager(client Client, store Store) *ReindexManager {
	return &ReindexManager{
		client: client,
		store:  store,
	}
}

func (m *ReindexManager) PublishReindexTask(ctx context.Context, src, dest string) (string, error) {
	numberOfReplicas, refreshInterval, err := m.client.GetIndexSetting(ctx, dest)
	if err != nil {
		return "", err
	}

	err = m.client.UpdateIndexSetting(ctx, dest, 0, -1)
	if err != nil {
		return "", err
	}

	taskID, err := m.client.Reindex(ctx, src, dest)
	if err != nil {
		return "", err
	}

	err = m.store.PutTaskInfo(taskID, numberOfReplicas, refreshInterval)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

func (m *ReindexManager) Run(ctx context.Context) error {
	return nil
}
