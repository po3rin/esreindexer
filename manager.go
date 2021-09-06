package esreindexer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/po3rin/esreindexer/entity"
	"golang.org/x/sync/errgroup"
)

type Store interface {
	PutTaskInfo(index string, taskID string, NumberOfReplicas int, RefreshInterval int) error
	TaskInfo(taskID string) (numberOfReplicas int, refreshInterval int, err error)
	AllTaskd() map[string]entity.Task
}

type ReindexManager struct {
	client *ESClient
	store  Store
}

func NewReindexManager(client *ESClient, store Store) *ReindexManager {
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

	err = m.store.PutTaskInfo(dest, taskID, numberOfReplicas, refreshInterval)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

func (m *ReindexManager) API(ctx context.Context) error {
	return errors.New("no impliments")
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
				ids := m.store.AllTaskd()
				if len(ids) == 0 {
					fmt.Println("no tasks")
				}

				for id, task := range ids {
					completed, err := m.client.CompletedTask(ctx, id)
					if err != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							return err
						}
						fmt.Println(err)
						continue
					}
					if completed {
						fmt.Printf("%v is completed!\n", id)
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
							fmt.Println(err)
							continue
						}
						continue
					}
					fmt.Printf("%v is running!\n", id)
				}
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
