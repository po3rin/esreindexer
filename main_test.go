package esreindexer_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/ory/dockertest"
	"github.com/po3rin/eskeeper"
)

var url string

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run(
		"docker.elastic.co/elasticsearch/elasticsearch",
		"7.14.0",
		[]string{
			"ES_JAVA_OPTS=-Xms512m -Xmx512m",
			"discovery.type=single-node",
			"node.name=es01",
		},
	)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	port := resource.GetPort("9200/tcp")
	url = fmt.Sprintf("http://localhost:%s", port)

	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		res, err := es.Info()
		if err != nil {
			log.Println("waiting to be ready...")
			return err
		}
		defer res.Body.Close()
		return nil
	}); err != nil {
		log.Fatalf("could not retry to connect : %s\n", err)
	}

	k, err := eskeeper.New(
		[]string{url},
		eskeeper.SkipPreCheck(true),
	)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	r, err := os.Open("testdata/test.eskeeper.yml")
	if err != nil {
		log.Fatalf("read eskeeper file: %+v", err)
	}

	ctx := context.Background()
	err = k.Sync(ctx, r)
	if err != nil {
		log.Fatalf("sync index with eskeeper: %+v", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
