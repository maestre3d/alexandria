package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/application"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"net/http"
)

type BookHandler struct {
	logger      util.ILogger
	router      *gin.RouterGroup
	bookUseCase *application.BookUseCase
}

func NewBookHandler(logger util.ILogger, router *gin.RouterGroup, bookUseCase *application.BookUseCase) error {
	bookHandler := &BookHandler{logger, router, bookUseCase}
	bookHandler.mapRoutes()

	logger.Print("Create HTTP book handler", "presentation.delivery.http.handler")
	return nil
}

func (b *BookHandler) mapRoutes() {
	bookRouter := b.router.Group("/book")

	bookRouter.POST("/", b.create)
	bookRouter.GET("/", b.getAll)
	bookRouter.GET("/:id", b.get)
	bookRouter.PATCH("/:id", b.updateOne)
	bookRouter.DELETE("/:id", b.deleteOne)
}

func (b *BookHandler) create(c *gin.Context) {
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

func (b *BookHandler) get(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from book handler getOne",
	})
}

func (b *BookHandler) getAll(c *gin.Context) {
	b.logger.Print("Received GET request", "presentation.delivery.http.handler.book")

	books, err := b.bookUseCase.GetAll(nil)
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

func (b *BookHandler) updateOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from book handler update",
	})
}

func (b *BookHandler) deleteOne(c *gin.Context) {
	c.JSON(http.StatusOK, &gin.H{
		"message": "Hello from book handler delete " + c.Param("id"),
	})
}
