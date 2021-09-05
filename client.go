package esreindexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
)

type ESClient struct {
	es *elasticsearch.Client
}

func NewESClient(es *elasticsearch.Client) *ESClient {
	return &ESClient{es: es}
}

type ReindexRes struct {
	Task string `json:"task"`
}

type GetTaskRes struct {
	Completed bool `json:"completed"`
}

func (c *ESClient) Reindex(ctx context.Context, src string, dest string) (string, error) {
	ri := c.es.Reindex
	body := strings.NewReader(
		fmt.Sprintf(`
{
  "source": {
    "index": "%s"
  },
  "dest": {
    "index": "%s"
  }
}`,
			src, dest,
		),
	)

	res, err := ri(
		body,
		ri.WithContext(ctx),
		ri.WithSlices("auto"),
		ri.WithWaitForCompletion(false),
	)
	if err != nil {
		return "", fmt.Errorf("reindex: %w", err)
	}
	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("failed to reindex [index=%v, statusCode=%v, res=%v]", src, res.StatusCode, string(body))
	}

	var reindexRes ReindexRes
	err = json.NewDecoder(res.Body).Decode(&reindexRes)
	if err != nil {
		return "", err

	}
	return reindexRes.Task, nil
}

func (c *ESClient) UpdateIndexSetting(ctx context.Context, index string, numberOfReplicas int, refreshInterval int) error {
	body := strings.NewReader(
		fmt.Sprintf(
			`{"index": {"number_of_replicas": %d, "refresh_interval" :%d}}`,
			numberOfReplicas, refreshInterval,
		),
	)

	ps := c.es.Indices.PutSettings

	res, err := ps(
		body,
		ps.WithIndex(index),
		ps.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to reindex [index=%v, statusCode=%v, res=%v]", index, res.StatusCode, string(body))
	}

	return nil
}

func (c *ESClient) GetIndexSetting(ctx context.Context, index string) (numberOfReplicas int, refreshInterval int, err error) {
	gs := c.es.Indices.GetSettings

	res, err := gs(
		gs.WithIndex(index),
		gs.WithContext(ctx),
	)
	if err != nil {
		return 0, 0, err
	}
	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return 0, 0, err
		}
		return 0, 0, fmt.Errorf("failed to reindex [index=%v, statusCode=%v, res=%v]", index, res.StatusCode, string(body))
	}

	return 0, 0, nil
}

func (c *ESClient) CompletedTask(ctx context.Context, taskID string) (bool, error) {
	tasks := c.es.Tasks.Get

	res, err := tasks(
		taskID,
		tasks.WithContext(ctx),
	)
	if err != nil {
		return false, err
	}
	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("failed to get task [taskID=%v, statusCode=%v, res=%v]", taskID, res.StatusCode, string(body))
	}

	var getTaskRes GetTaskRes
	err = json.NewDecoder(res.Body).Decode(&getTaskRes)
	if err != nil {
		return false, err

	}
	return getTaskRes.Completed, nil
}
