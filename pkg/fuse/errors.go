package fuse

import (
	"errors"
	"fmt"
)

var (
	ErrMountFailed   = errors.New("mount failed")
	ErrUnmountFailed = errors.New("unmount failed")
)

type TFSError struct {
	info string
	err  error
}

func NewTFSError(additionalInfo string, fuseError error) TFSError {
	return TFSError{
		info: additionalInfo,
		err:  fuseError,
	}
}

func (e TFSError) Error() string {
	return fmt.Sprintf("TFS error (%s): %v", e.info, e.err)
}
