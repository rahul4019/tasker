package repository

import "github.com/rahul4019/tasker/internal/server"

type Repositories struct {
	Todo    *TodoRepository
	Comment *CommentRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		Todo:    NewTodoRepository(s),
		Comment: NewCommentRepository(s),
	}
}
