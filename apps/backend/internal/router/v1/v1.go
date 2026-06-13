package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/rahul4019/tasker/internal/handler"
	"github.com/rahul4019/tasker/internal/middleware"
)

func RegisterTodoRoutes(router *echo.Group, handlers *handler.Handlers, middlewares *middleware.Middlewares) {
	// Register todo routes
	registerTodoRoutes(router, handlers.Todo, handlers.Comment, middlewares.Auth)

	// Register Category routes
	registerCategoriesRoutes(router, handlers.Category, middlewares.Auth)

	// Register Comment routes
	registerCommentRoutes(router, handlers.Comment, middlewares.Auth)
}
