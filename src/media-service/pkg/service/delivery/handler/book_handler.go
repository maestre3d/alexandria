package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/application"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"go.uber.org/multierr"
	"net/http"
)

type BookHandler struct {
	logger      util.ILogger
	bookUseCase *application.BookUseCase
}

func NewBookHandler(logger util.ILogger, bookUseCase *application.BookUseCase) *BookHandler {
	logger.Print("Create HTTP media handler", "service.delivery.http.handler")
	return &BookHandler{logger, bookUseCase}
}

func (b *BookHandler) Create(c *gin.Context) {
	err := b.bookUseCase.Create(c.PostForm("title"), c.PostForm("published_at"), c.PostForm("uploaded_by"), c.PostForm("author"))
	// Read all errors
	errors := multierr.Errors(err)
	if len(errors) > 0 {
		for i, err := range errors {
			// If client wants to create an already existing entity, go 409
			if err == global.EntityExists {
				c.JSON(http.StatusConflict, &gin.Error{
					Err:  err,
					Type: http.StatusConflict,
					Meta: err.Error(),
				})
				return
			} else if err == global.EntityDomainError {
				// If client sent something wrong, go 400
				// A DOMAIN ENTITY ERROR MUST BE FOLLOWED BY THE CUSTOM ERROR
				if errors[i+1] != nil {
					c.JSON(http.StatusBadRequest, &gin.Error{
						Err:  errors[i+1],
						Type: http.StatusBadRequest,
						Meta: errors[i+1].Error(),
					})
					return
				}

				c.JSON(http.StatusBadRequest, &gin.Error{
					Err:  err,
					Type: http.StatusBadRequest,
					Meta: err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, &gin.Error{
				Err:  err,
				Type: http.StatusInternalServerError,
				Meta: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, &gin.H{
		"message": "media created",
	})
}

func (b *BookHandler) Get(c *gin.Context) {
	id := c.Param("book_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, &gin.Error{
			Err:  global.InvalidID,
			Type: http.StatusBadRequest,
			Meta: global.InvalidID.Error(),
		})
		return
	}

	book, err := b.bookUseCase.GetByID(id)
	errors := multierr.Errors(err)
	if len(errors) > 0 {
		for _, err = range errors {
			if err == global.EntityNotFound {
				c.JSON(http.StatusNotFound, &gin.Error{
					Err:  err,
					Type: http.StatusNotFound,
					Meta: err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, &gin.Error{
				Err:  err,
				Type: http.StatusInternalServerError,
				Meta: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, book)
}

func (b *BookHandler) List(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")

	params := global.NewPaginationParams(page, limit)

	books, err := b.bookUseCase.GetAll(params)
	errors := multierr.Errors(err)
	if len(errors) > 0 {
		for _, err = range errors {
			if err == global.EntitiesNotFound {
				c.JSON(http.StatusNotFound, &gin.Error{
					Err:  err,
					Type: http.StatusNotFound,
					Meta: err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, &gin.Error{
				Err:  err,
				Type: http.StatusInternalServerError,
				Meta: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, books)
}

func (b *BookHandler) UpdateOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from media handler update",
	})
}

func (b *BookHandler) DeleteOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from media handler delete " + c.Param("book_id"),
	})
}
