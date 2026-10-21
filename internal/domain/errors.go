package domain

import "errors"

// Risk errors
var (
	ErrRiskNotFound       = errors.New("risk not found")
	ErrLinkedRiskSameType = errors.New("linked risk must be opposite type")
	ErrRiskAlreadyClosed  = errors.New("risk is already closed")
)

// Person/Auth errors
var (
	ErrPersonNotFound     = errors.New("person not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
)

// Comment errors
var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrNotCommentAuthor = errors.New("not the comment author")
)

// Photo errors
var (
	ErrPhotoNotFound     = errors.New("photo not found")
	ErrPhotoUploadFailed = errors.New("photo upload failed")
)

// Validation errors
var (
	ErrValidationFailed = errors.New("validation failed")
)
