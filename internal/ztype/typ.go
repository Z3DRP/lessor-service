package ztype

import (
	"errors"
	"mime/multipart"
)

type JsonResponse map[string]any

type FileUploadDto struct {
	File    multipart.File
	FileKey string
	Header  *multipart.FileHeader
}

func (f FileUploadDto) Validate() error {
	if f.File == nil {
		return errors.New("error file is required")
	}

	if f.FileKey == "" {
		return errors.New("error file key is empty")
	}

	if f.Header == nil {
		return errors.New("error file header is required")
	}

	return nil
}
