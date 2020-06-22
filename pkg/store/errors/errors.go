package errors

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"

	log "github.com/sirupsen/logrus"
)

var errorMessages = map[StorageErrorType]string{
	NotFoundError:       "Entity not found",
	AccessDeniedError:   "No access to entity",
	AlreadyExistsError:  "Entity already exists",
	ForeignKeyViolation: "Foreign key violation",
	InternalError:       "Internal error",
}

type StorageErrorType string

const (
	NotFoundError       StorageErrorType = "Not Found Error"
	ForeignKeyViolation StorageErrorType = "Foreign Key Violation"
	InternalError       StorageErrorType = "Internal Error"
	AlreadyExistsError  StorageErrorType = "Already exists error"
	// AccessDeniedError is returned when the user is trying to access an entity which it doesn't control
	AccessDeniedError StorageErrorType = "Access Denied Error"
)

// ErrorResponse is a generic response type for errors in HTTP requests
type StorageError struct {
	Type         StorageErrorType
	Message      string
	WrappedError error
}

func (storageError *StorageError) Error() string {
	return fmt.Sprintf("%s: %s", storageError.Message, storageError.WrappedError.Error())
}

// newStorageError takes an SQL error and returns a corresponding storage error
func NewStorageError(errorType StorageErrorType, err error) error {
	storageError := NewStorageErrorFromError(err)

	if storageError, ok := storageError.(*StorageError); ok {
		storageError.Type = errorType
		storageError.Message = errorMessages[errorType]
	}

	return storageError
}

// NewStorageErrorFromError is a helper function to easily handle a error received
// during interacting with an sql engine. It will try to find the correct StorageErrorType
// and populate a StorageError accordingly.
func NewStorageErrorFromError(err error) error {
	var storageError StorageError

	if err == nil {
		// Err is nil, returning err directly.
		return err
	}

	var errorType StorageErrorType

	switch err := err.(type) {
	case pq.Error:
		switch err.Code {
		case "23503":
			errorType = ForeignKeyViolation
		case "23505":
			errorType = AlreadyExistsError
		default:
			log.Warnf("Could not find PostgreSQL error number '%s': %s", err.Code, err.Code.Name())
			errorType = InternalError
		}
	case sqlite3.Error:
		switch err.Code {
		case sqlite3.ErrConstraint:
			errorType = ForeignKeyViolation
		default:
			log.Warnf("Could not find SQLite error number '%s', '%s'", err.Code, err)
			errorType = InternalError
		}
	default:
		switch err {
		case sql.ErrTxDone:
			errorType = InternalError
		case sql.ErrNoRows:
			errorType = NotFoundError
		case sql.ErrConnDone:
			errorType = InternalError
		default:
			errorType = InternalError
		}
	}

	storageError = StorageError{
		Type:         errorType,
		Message:      errorMessages[errorType],
		WrappedError: err,
	}

	return &storageError
}
