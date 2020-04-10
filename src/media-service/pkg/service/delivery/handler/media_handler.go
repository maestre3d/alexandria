package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/application"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"go.uber.org/multierr"
	"net/http"
	"strings"
)

type MediaHandler struct {
	logger       util.ILogger
	mediaUseCase *application.MediaUseCase
}

func NewMediaHandler(logger util.ILogger, mediaUseCase *application.MediaUseCase) *MediaHandler {
	logger.Print("media handler created", "service.delivery.handler")
	return &MediaHandler{logger, mediaUseCase}
}

func (m *MediaHandler) Create(c *gin.Context) {
	params := &application.MediaParams{
		Title:       c.PostForm("title"),
		DisplayName: c.PostForm("display_name"),
		Description: c.PostForm("description"),
		UserID:      c.PostForm("user_id"),
		AuthorID:    c.PostForm("author_id"),
		PublishDate: c.PostForm("publish_date"),
		MediaType:   c.PostForm("media_type"),
	}
	err := m.mediaUseCase.Create(params)
	if err != nil {
		if errors.Is(err, global.EntityExists) {
			c.JSON(http.StatusConflict, &gin.H{
				"code":    http.StatusConflict,
				"message": err.Error(),
			})
			return
		} else if errors.Is(err, global.InvalidID) || errors.Is(err, global.RequiredField) || errors.Is(err, global.InvalidFieldFormat) ||
			errors.Is(err, global.InvalidFieldRange) {
			// Business exception
			errs := multierr.Errors(err)
			for _, err = range errs {
				errString := strings.Split(err.Error(), ":")
				if len(errString) > 1 {
					c.JSON(http.StatusBadRequest, &gin.H{
						"code":    http.StatusBadRequest,
						"message": errString[1],
					})
					return
				}
			}

			// Use case / Infrastructure exception
			c.JSON(http.StatusBadRequest, &gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, &gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	// Return created media
	media, err := m.mediaUseCase.GetByTitle(params.Title)
	if err != nil {
		if errors.Is(err, global.EmptyQuery) {
			c.JSON(http.StatusBadRequest, &gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
			return
		} else if errors.Is(err, global.EntityNotFound) {
			c.JSON(http.StatusNotFound, &gin.H{
				"code":    http.StatusNotFound,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, &gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"code":  http.StatusOK,
		"media": media,
	})
}

func (m *MediaHandler) Get(c *gin.Context) {
	id := c.Param("media_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, &gin.H{
			"code":    http.StatusBadRequest,
			"message": global.InvalidID.Error(),
		})
		return
	}

	media, err := m.mediaUseCase.GetByID(id)
	if err != nil {
		if errors.Is(err, global.EntityNotFound) {
			c.JSON(http.StatusNotFound, &gin.H{
				"code":    http.StatusNotFound,
				"message": err.Error(),
			})
			return
		} else if errors.Is(err, global.InvalidID) {
			c.JSON(http.StatusBadRequest, &gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
			return
		}

		// Generic error
		c.JSON(http.StatusInternalServerError, &gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"code":  http.StatusOK,
		"media": media,
	})
}

func (m *MediaHandler) List(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")

	params := util.NewPaginationParams(page, limit)

	medias, err := m.mediaUseCase.GetAll(params)
	if err != nil {
		if errors.Is(err, global.EntitiesNotFound) {
			c.JSON(http.StatusNotFound, &gin.H{
				"code":    http.StatusNotFound,
				"message": err.Error(),
			})
			return
		}

		// Generic error
		c.JSON(http.StatusInternalServerError, &gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"code":  http.StatusOK,
		"media": medias,
	})
}

func (m *MediaHandler) UpdateOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from media handler update",
	})
}

func (m *MediaHandler) DeleteOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from media handler delete " + c.Param("book_id"),
	})
}
