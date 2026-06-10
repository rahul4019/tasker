package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rahul4019/tasker/internal/model/comment"
	"github.com/rahul4019/tasker/internal/server"
)

type CommentRepository struct {
	server *server.Server
}

func NewCommentRepository(server *server.Server) *CommentRepository {
	return &CommentRepository{server: server}
}

func (r *CommentRepository) AddComment(ctx context.Context, userID string, todoID uuid.UUID,
	payload *comment.AddCommentPayload,
) (*comment.Comment, error) {
	stmt := `
		INSERT INTO
			todo_comments (
				todo_id,
				user_id,
				content
			)
		VALUES
			(
				@todo_id,
				@user_id,
				@content
			)
		RETURNING
		*
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
		"user_id": userID,
		"content": payload.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute add comment query for todo_id=%s user_id=%s: %w", todoID.String(), userID, err)
	}

	commentItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[comment.Comment])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:todo_comments for todo_id=%s user_id=%s: %w", todoID.String(), userID, err)
	}

	return &commentItem, nil
}
