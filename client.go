package esreindexer

import (
	"context"
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
	return "", nil
}
