package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/service"
)

// Shorten defines the interface for shorten URL HTTP request handling.
type Shorten interface {
	ShortenURL(c *gin.Context)
}

type shortenHandler struct {
	shortenService service.ShortenService
}

// NewShorten creates and returns a new Shorten handler instance.
func NewShorten(shortenSvc service.ShortenService) Shorten {
	return &shortenHandler{
		shortenService: shortenSvc,
	}
}

type shortenRequest struct {
	URL string `json:"url" binding:"required,url"`
	Exp int    `json:"exp" binding:"required,min=1"`
}

// ShortenURL handles HTTP POST /v1/links/shorten.
func (s *shortenHandler) ShortenURL(c *gin.Context) {
	var req shortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code, err := s.shortenService.ShortenURL(
		c.Request.Context(),
		req.URL,
		time.Duration(req.Exp)*time.Second,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": code})
}
