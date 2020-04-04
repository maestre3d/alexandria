package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/application"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"net/http"
)

type BookHandler struct {
	logger      util.ILogger
	bookUseCase *application.BookUseCase
}

func NewBookHandler(logger util.ILogger, bookUseCase *application.BookUseCase) *BookHandler {
	logger.Print("Create HTTP book handler", "presentation.delivery.http.handler")
	return &BookHandler{logger, bookUseCase}
}

func (b *BookHandler) Create(c *gin.Context) {
	err := b.bookUseCase.Create(c.PostForm("title"), c.PostForm("published_at"), c.PostForm("uploaded_by"), c.PostForm("author"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, &gin.Error{
			Err:  err,
			Type: http.StatusInternalServerError,
			Meta: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"message": "book created",
	})
}

func (b *BookHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from book handler getOne",
	})
}

func (b *BookHandler) GetAll(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")

	params := global.NewPaginationParams(page, limit)

	books, err := b.bookUseCase.GetAll(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &gin.Error{
			Err:  err,
			Type: http.StatusInternalServerError,
			Meta: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, books)
}

func (b *BookHandler) UpdateOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from book handler update",
	})
}

func (b *BookHandler) DeleteOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from book handler delete " + c.Param("id"),
	})
}
