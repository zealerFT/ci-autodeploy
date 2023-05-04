package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.String(http.StatusOK, ":)")
}

func Handle404(c *gin.Context) {
	c.String(http.StatusNotFound, "404 NotFound")
}
