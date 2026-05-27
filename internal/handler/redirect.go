package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/rs/zerolog/log"
)

// Redirect defines the interface for redirect HTTP request handling.
type Redirect interface {
	Redirect(c *gin.Context)
}

type redirectHandler struct {
	shortenService service.ShortenService
}

// NewRedirect creates and returns a new Redirect handler instance.
func NewRedirect(shortenSvc service.ShortenService) Redirect {
	return &redirectHandler{shortenService: shortenSvc}
}

// Redirect handles HTTP GET /v1/links/redirect/:code.
//
// @Summary      Redirect to original URL
// @Description  Looks up the shortened code in storage and issues an HTTP 302 redirect to the original URL.
// @Tags         links
// @Produce      json
// @Param        code  path      string  true  "Shortened code"
// @Success      302
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /v1/links/redirect/{code} [get]
func (h *redirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.shortenService.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrCodeNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "code not found"})
			return
		}
		log.Error().
			Err(err).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("redirect url failed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if c.Query("json") == "true" {
		c.JSON(http.StatusOK, gin.H{"redirect_url": url})
		return
	}

	c.Redirect(http.StatusFound, url)
}
