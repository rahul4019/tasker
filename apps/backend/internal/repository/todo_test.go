package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rahul4019/tasker/internal/model/todo"
	"github.com/rahul4019/tasker/internal/repository"
	testing_pkg "github.com/rahul4019/tasker/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoRepository_CreateTodo(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	t.Run("create todo successfully", func(t *testing.T) {
		userID := uuid.New().String()
		dueDate := time.Now().Add(24 * time.Hour)
		payload := &todo.CreateTodoPayload{
			Title:       "Test Todo",
			Description: testing_pkg.Ptr("Test todo description"),
			Priority:    testing_pkg.Ptr(todo.PriorityHigh),
			DueDate:     &dueDate,
		}

		result, err := todoRepo.CreateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, payload.Title, result.Title)
		assert.Equal(t, payload.Description, result.Description)
		assert.Equal(t, *payload.Priority, result.Priority)
		assert.Equal(t, payload.DueDate.Unix(), result.DueDate.Unix())
		assert.Equal(t, todo.StatusDraft, result.Status)
		assert.Nil(t, result.CompletedAt)
		testing_pkg.AssertTimestampsValid(t, result)
	})
}
