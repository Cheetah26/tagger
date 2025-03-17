package tagger

import (
	"errors"
	"fmt"
)

var ErrTagNotExist = errors.New("Tag does not exist")
var ErrFileNotExist = errors.New("File does not exist")

type DatabaseError struct {
	Err error
}

func NewDatabaseError(orig error) error {
	if orig == nil {
		return nil
	}

	return &DatabaseError{
		Err: orig,
	}
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("Database Error: %v", e.Err)
}
