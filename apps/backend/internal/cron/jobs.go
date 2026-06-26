package cron

import (
	"context"

	"github.com/rahul4019/tasker/internal/lib/job"
)

type DueDateRemindersJob struct{}

func (j *DueDateRemindersJob) Name() string {
	return "due-date-reminders"
}

func (j *DueDateRemindersJob) Description() string {
	return "Enqueue email reminders for todos due soon"
}

func (j *DueDateRemindersJob) Run(ctx context.Context, jobCtx *JobContext) error {

	todos, err := jobCtx.Repositories.Todo.GetTodosDueInHours(
		ctx,
		jobCtx.Config.Cron.ReminderHours,
		jobCtx.Config.Cron.BatchSize,
	)
	if err != nil {
		return err
	}

	jobCtx.Server.Logger.Info().
		Int("todo_count", len(todos)).
		Int("hours", jobCtx.Config.Cron.ReminderHours).
		Msg("Found todos due soon")

	userTodos := make(map[string][]string)
	enueuedCount := 0

	for _, todo := range todos {
		if len(userTodos[todo.UserID]) < jobCtx.Config.Cron.MaxTodosPerUserNotification {
			userTodos[todo.UserID] = append(userTodos[todo.UserID], todo.Title)
		}

		reminderTask := &job.ReminderEmailTask{
			UserID:    todo.UserID,
			TodoID:    todo.ID,
			TodoTitle: todo.Title,
			DueDate:   *todo.DueDate,
			TaskType:  "due_date_reminder",
		}

		err := job.EnqueueReminderEmail(jobCtx.JobClient, reminderTask)
		if err != nil {
			jobCtx.Server.Logger.Error().
				Err(err).
				Str("todo_id", todo.ID.String()).
				Str("user_id", todo.UserID).
				Msg("Failed to enqueue reminder email")
			continue
		}

		enueuedCount++
		jobCtx.Server.Logger.Info().
			Str("todo_id", todo.ID.String()).
			Str("todo_title", todo.Title).
			Str("user_id", todo.UserID).
			Msg("Enqueued reminder for todo")
	}

	jobCtx.Server.Logger.Info().
		Int("enqueued_count", enueuedCount).
		Int("total_todos", len(todos)).
		Msg("Due date reminder emails enqueued")

	for userID, titles := range userTodos {
		jobCtx.Server.Logger.Info().
			Str("user_id", userID).
			Int("reminder_count", len(titles)).
			Msg("User reminders enqueued")
	}

	return nil

}
