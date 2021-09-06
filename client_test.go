package esreindexer_test

import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/po3rin/esreindexer"
)

func TestIndexSetting(t *testing.T) {
	tests := []struct {
		name             string
		index            string
		numberOfReplicas int
		refreshInterval  int
	}{
		{
			name:             "update-revert",
			index:            "test-client",
			numberOfReplicas: 3,
			refreshInterval:  -1,
		},
	}
	conf := elasticsearch.Config{
		Addresses: []string{url},
	}
	es, err := elasticsearch.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	client := esreindexer.NewESClient(es)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := client.UpdateIndexSetting(
				ctx, tt.index, tt.numberOfReplicas, tt.refreshInterval,
			)
			if err != nil {
				t.Fatalf("failed to update: %+v", err)
			}

			gotNR, gotRI, err := client.GetIndexSetting(ctx, tt.index)
			if gotNR != tt.numberOfReplicas || gotRI != tt.refreshInterval {
				t.Errorf("numberOfReplicas %v:%v, refreshInterval %v:%v", tt.numberOfReplicas, gotNR, tt.refreshInterval, gotRI)
			}
		})
	}
}
