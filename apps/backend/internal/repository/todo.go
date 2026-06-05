package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
		return nil, fmt.Errorf("failed to collect row from table:todos for user_id=%s title=%s: %w", userID, payload.Title, err)
	}
	return &todoItem, nil
}

func (r *TodoRepository) GetTodoByID(ctx context.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	stmt := `
		SELECT
			t.*,
			CASE
				WHEN c.id IS NOT NULL THEN to_jsonb(camel (c))
				ELSE NULL
			END AS category,
			COALESCE(
				jsonb_agg(
					to_jsonb(camel (child))
					ORDER BY
						child.sort_order ASC,
						child.created_at ASC,
				) FILTER (
					WHERE 
						child.id IS NOT NULL	 
				),
				'[]'::JSONB
			) AS children
		FROM
			todos t
			LEFT JOIN todo_categories c ON c.id=t.category_id
			AND c.user_id=@user_id
			LEFT JOIN todos child ON child.parent_todo_id=t.id
			AND child.user_id=@user_id
			LEFT JOIN todo_comments com ON com.todo_id=t.id
			AND com.user_id=@user_id
		WHERE
			t.id=@id
			AND t.user_id=@user_id
		GROUP BY
			t.id,
			c.id

`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      todoID,
		"user_id": userID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute get todo by id query for todo_id=%s user_id=%s: %w", todoID.String(), userID, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:todos for todo_id=%s user_id=%s: %w", todoID.String(), userID, err)
	}

	return &todoItem, nil
}
