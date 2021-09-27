package esreindexer

import (
	"context"
	"errors"
	"time"

	"github.com/po3rin/esreindexer/entity"
	"github.com/po3rin/esreindexer/logger"
	"golang.org/x/sync/errgroup"
)

type Store interface {
	PutTaskInfo(index string, taskID string, NumberOfReplicas int, RefreshInterval int) error
	TaskInfo(taskID string) (numberOfReplicas int, refreshInterval int, err error)
	DeleteTask(taskID string) error
	DoneTask(taskID string) error
	AllTask() map[string]entity.Task
}

type AfterCompletionPlugin interface {
	Run(ctx context.Context, taskID string) error
}

type ReindexManager struct {
	client                *ESClient
	store                 Store
	afterCompletionPlugin AfterCompletionPlugin
}

func NewReindexManager(client *ESClient, store Store) *ReindexManager {
	return &ReindexManager{
		client: client,
		store:  store,
	}
}

func (m *ReindexManager) NotifyCompletionPlugin(p AfterCompletionPlugin) {
	m.afterCompletionPlugin = p
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

	err = m.store.PutTaskInfo(dest, taskID, numberOfReplicas, refreshInterval)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

func (m *ReindexManager) Monitor(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				ids := m.store.AllTask()
				if len(ids) == 0 {
					continue
				}

				now := time.Now()

				for id, task := range ids {

					// ignore done status task.
					// delete task if task is expired.
					if task.Status == entity.Done {
						if now.After(task.ExpireDate) {
							err := m.store.DeleteTask(id)
							if err != nil {
								logger.L.Errorf("delete completed task %v: %v", id, err)
							}
							logger.L.Infof("completed task %v is expired. So, task info is deleted in store", id)
						}
						continue
					}

					// check if the task is finished
					completed, err := m.client.CompletedTask(ctx, id)
					if err != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							return err
						}
						logger.L.Warn(err)
						continue
					}

					// if task is completed, update task status and run completion plugin func.
					if completed {
						logger.L.Infof("task %v is completed!", id)
						err := m.client.UpdateIndexSetting(
							ctx,
							task.Index,
							task.NumberOfReplicas,
							task.RefreshInterval,
						)
						if err != nil {
							if errors.Is(err, context.DeadlineExceeded) {
								return err
							}
							logger.L.Warnf("update index setting: %v", err)
							continue
						}
						err = m.store.DoneTask(id)
						if err != nil {
							logger.L.Errorf("update task status to done %v: %v", id, err)
						}

						if m.afterCompletionPlugin == nil {
							continue
						}
						err = m.afterCompletionPlugin.Run(ctx, id)
						if err != nil {
							logger.L.Errorf("run conpletion plugin: %v", err)
						}

						continue
					}
					logger.L.Infof("task %v is running!", id)
				}
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
