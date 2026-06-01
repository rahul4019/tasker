package todo

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ----------------------------------------------------------

type CreateTodoPayload struct {
	Title        string     `json:"title" validate:"required,min=1,max=255"`
	Description  *string    `json:"description" validate:"omitempty,max=1000"`
	Priority     *Priority  `json:"priority" validate:"omitempty,oneof= low medium high"`
	DueDate      *time.Time `json:"dueDate"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata  `json:"metadata"`
}

func (p *CreateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ----------------------------------------------------------

type UpdateTodoPayload struct {
	ID           uuid.UUID  `json:"id" validate:"required,uuid"`
	Title        *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description  *string    `json:"description" validate:"omitempty,max=1000"`
	Status       *Status    `json:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority     *Priority  `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate      *time.Time `json:"dueDate"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata  `json:"metadata"`
}

func (p *UpdateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
