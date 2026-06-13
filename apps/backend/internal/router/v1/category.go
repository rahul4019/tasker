package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/rahul4019/tasker/internal/handler"
	"github.com/rahul4019/tasker/internal/middleware"
)

func registerCategoriesRoutes(r *echo.Group, h *handler.CategoryHandler, auth *middleware.AuthMiddleware) {
	// Categories operations
	categories := r.Group("/categories")
	categories.Use(auth.RequireAuth)

	// Category collection operations
	categories.POST("", h.CreateCategory)
	categories.GET("", h.GetCategories)

	// Individual comment operations
	dynamicCategory := categories.Group("/:id")
	dynamicCategory.PATCH("", h.UpdateCategory)
	dynamicCategory.DELETE("", h.DeleteCategory)
}
