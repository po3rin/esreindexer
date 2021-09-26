package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	server http.Server
}

func router(ctl ReindexCtl) *gin.Engine {
	h := NewHandler(ctl)
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	rg := r.Group("api/v1")
	{
		rg.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok!")
		})
		rg.POST("/reindex", h.Reindex)
	}
	return r
}

func New(port string, ctl ReindexCtl) *Server {
	return &Server{
		server: http.Server{
			Addr:    port,
			Handler: router(ctl),
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return s.server.ListenAndServe()
	})

	<-ctx.Done()
	sCtx, sCancel := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer sCancel()
	if err := s.server.Shutdown(sCtx); err != nil {
		return err
	}

	return eg.Wait()
}
