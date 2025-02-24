package api

import "fmt"

type ErrInvalidBucketDir struct {
	InvalidDir string
}

func (e ErrInvalidBucketDir) Error() string {
	return fmt.Sprintf("invalid bucket directory %v does not exist", e.InvalidDir)
}

type ErrFileRead struct {
	Err error
}

func (e ErrFileRead) Error() string {
	return fmt.Sprintf("failed to read file %v", e.Err.Error())
}

type ErrMaxSize struct {
	Err error
}

func (e ErrMaxSize) Error() string {
	return fmt.Sprintf("image to large, max size %v", maxSize)
}

type ErrFileObjUpload struct {
	Err error
}

func (e ErrFileObjUpload) Error() string {
	return fmt.Sprintf("file upload failed %v", e.Err)
}

type ErrFileObjRead struct {
	Err error
}

func (e ErrFileObjRead) Error() string {
	return fmt.Sprintf("file read failed %v", e.Err)
}
