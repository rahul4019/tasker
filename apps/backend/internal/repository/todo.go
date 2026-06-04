package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rahul4019/tasker/internal/model/todo"
	"github.com/rahul4019/tasker/internal/server"
)

type TodoRepository struct {
	server *server.Server
}

func NewTodoRepository(server *server.Server) *TodoRepository {
	return &TodoRepository{server: server}
}

func (r *TodoRepository) CreateTodo(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	stmt := `
		INSERT INTO
			todos(
				user_id,
				title,
				description,
				priority,
				due_date,
				parent_todo_id,
				category_id,
				metadata
			)
		VALUES
			(
				@user_id,
				@title,
				@description,
				@priority,
				@due_date,
				@parent_todo_id,
				@category_id,
				@metadata
			)
		RETURNING
		*
	`
	priority := todo.PriorityMedium
	if payload.Priority != nil {
		priority = *payload.Priority
	}

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":        userID,
		"title":          payload.Title,
		"description":    payload.Description,
		"priority":       priority,
		"due_date":       payload.DueDate,
		"parent_todo_id": payload.ParentTodoID,
		"category_id":    payload.CategoryID,
		"metadata":       payload.Metadata,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute create todo query for user_id=%s title=%s: %w", userID, payload.Title, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		return nil, fmt.Errorf("failed to execute create todo query for user_id=%s title=%s: %w", userID, payload.Title, err)
	}
	return &todoItem, nil
}
