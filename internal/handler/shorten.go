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

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
	Exp int    `json:"exp" binding:"required,min=1"`
}

// ShortenURL handles HTTP POST /v1/links/shorten.
//
// @Summary      Shorten a URL
// @Description  Generates a 7-character alphanumeric code for the given URL and stores it in Redis with the specified TTL
// @Tags         links
// @Accept       json
// @Produce      json
// @Param        request  body      ShortenRequest    true  "Shorten URL request"
// @Success      201      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /v1/links/shorten [post]
func (s *shortenHandler) ShortenURL(c *gin.Context) {
	var req ShortenRequest

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

	c.JSON(http.StatusCreated, gin.H{
		"code":    code,
		"message": "Shorten URL generated successfully!",
	})
}
