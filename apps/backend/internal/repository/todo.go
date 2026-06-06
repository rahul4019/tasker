package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rahul4019/tasker/internal/model"
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

func (r *TodoRepository) CheckTodoExists(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error) {
	stmt := `
		SELECT 
			*
		FROM
			todos
		WHERE
			id=@id
			AND user_id=@user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      todoID,
		"user_id": userID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check if todo exist for todo_id=%s user_id=%s: %w", todoID.String(), userID, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:todos for todo_id=%s user_id=%s: %w", todoID.String(), userID, err)
	}

	return &todoItem, nil
}

func (r *TodoRepository) GetTodos(ctx context.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
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
`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	conditions := []string{"t.user_id = @user_id"}

	if query.Status != nil {
		conditions = append(conditions, "t.status = @status")
		args["status"] = *query.Status
	}

	if query.Priority != nil {
		conditions = append(conditions, "t.priority = @priority")
		args["priority"] = *query.Priority
	}

	if query.CategoryID != nil {
		conditions = append(conditions, "t.category_id = @category_id")
		args["category_id"] = *query.CategoryID
	}

	if query.ParentTodoID != nil {
		conditions = append(conditions, "t.parent_todo_id = @parent_todo_id")
		args["parent_todo_id"] = *query.ParentTodoID
	} else {
		// By default, only show root todos (no parent)
		conditions = append(conditions, "t.parent_todo_id IS NULL")
	}

	if query.DueFrom != nil {
		conditions = append(conditions, "t.due_date >= @due_from")
		args["due_from"] = *query.DueFrom
	}

	if query.DueTo != nil {
		conditions = append(conditions, "t.due_date <= @due_to")
		args["due_to"] = *query.DueTo
	}

	if query.Overdue != nil {
		conditions = append(conditions, "t.due_date <= NOW() AND t.status != 'completed'")
	}

	if query.Completed != nil {
		if *query.Completed {
			conditions = append(conditions, "t.status = 'completed'")
		} else {
			conditions = append(conditions, "t.status != 'completed'")
		}
	}

	if query.Search != nil {
		conditions = append(conditions, "(t.title ILIKE @search OR t.description ILIKE @search)")
		args["search"] = "%" + *query.Search + "%"
	}

	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	countStmt := "SELECT COUNT(*) FROM todos t"
	if len(conditions) > 0 {
		countStmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count for todos user_id=%s: %w", userID, err)
	}

	stmt += " GROUP BY t.id, c.id"

	if query.Sort != nil {
		stmt += " ORDER BY t." + *query.Sort
		if query.Order != nil && *query.Order == "desc" {
			stmt += " DESC"
		} else {
			stmt += " ASC"
		}
	} else {
		stmt += " ORDER BY t.created_at DESC"
	}

	stmt += " LIMIT @limit OFFSET @offset"
	args["limit"] = *query.Limit
	args["offset"] = (*query.Page - 1) * (*query.Limit)

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todos query for user_id=%s: %w", userID, err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[todo.PopulatedTodo]{
				Data:       []todo.PopulatedTodo{},
				Page:       *query.Page,
				Limit:      *query.Limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:todos user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[todo.PopulatedTodo]{
		Data:       todos,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      total,
		TotalPages: (total + *query.Limit - 1) / *query.Limit,
	}, nil
}
