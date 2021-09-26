package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/po3rin/esreindexer"
	"github.com/po3rin/esreindexer/agent/server"
	"github.com/po3rin/esreindexer/store"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
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

	srv := server.New(":8888", m)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return srv.Run(ctx)
	})
	eg.Go(func() error {
		return m.Monitor(ctx)
	})

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		cancel()
	case <-ctx.Done():
	}

	if err := eg.Wait(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Println(err)
		os.Exit(1)
	}
}
