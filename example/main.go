package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/po3rin/esreindexer"
	"github.com/po3rin/esreindexer/store"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conf := elasticsearch.Config{}
	es, err := elasticsearch.NewClient(conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m := esreindexer.NewReindexManager(
		esreindexer.NewESClient(es), store.NewMemoryStore(),
	)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return m.Monitor(ctx)
	})

	time.Sleep(2 * time.Second)
	id, err := m.PublishReindexTask(ctx, "example-v1", "example-v2")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("published id: %+v\n", id)

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		cancel()
	case <-ctx.Done():
	}

	if err := eg.Wait(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			os.Exit(0)
		}
		fmt.Println(err)
		os.Exit(1)
	}
}
