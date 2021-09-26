package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReindexCtl interface {
	PublishReindexTask(ctx context.Context, src, dest string) (string, error)
}

type Handler struct {
	reindexCtl ReindexCtl
}

func NewHandler(r ReindexCtl) *Handler {
	return &Handler{
		reindexCtl: r,
	}
}

func (r *Handler) Reindex(c *gin.Context) {
	var json ReindexReq
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := r.reindexCtl.PublishReindexTask(c, json.Source.Index, json.Dest.Index)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, Err{Msg: err.Error()})
	}

	fmt.Printf("published id: %+v\n", id)
	c.JSON(http.StatusOK, OK{ID: id})
}
