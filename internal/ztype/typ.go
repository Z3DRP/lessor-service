package ztype

import (
	"errors"
	"mime/multipart"
)

type JsonResponse map[string]any

type FileUploadDto struct {
	File   multipart.File
	Header *multipart.FileHeader
}

func (f FileUploadDto) Validate() error {
	if f.File == nil {
		return errors.New("error file is required")
	}

	if f.Header == nil {
		return errors.New("error file header is required")
	}

	return nil
}
