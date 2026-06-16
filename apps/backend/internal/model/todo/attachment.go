package todo

import (
	"github.com/google/uuid"
	"github.com/rahul4019/tasker/internal/model"
)

type TodoAttachment struct {
	model.Base
	TodoID      uuid.UUID `json:"todoId" db:"todo_id"`
	Name        string    `json:"name" db:"name"`
	UploadedBy  string    `json:"uploadedBy" db:"uploaded_by"`
	DownloadKey uuid.UUID `json:"downloadKey" db:"download_key"`
	FileSize    uuid.UUID `json:"fileSize" db:"file_size"`
	MimeType    uuid.UUID `json:"mimeType" db:"mime_type"`
}
