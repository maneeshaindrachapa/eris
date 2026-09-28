package domain

import (
	"errors"
	"fmt"
)

type ErrorKind string

const (
	ErrorKindNotFound      ErrorKind = "not_found"
	ErrorKindInvalidGrant  ErrorKind = "invalid_grant"
	ErrorKindInvalidClient ErrorKind = "invalid_client"
	ErrorKindInvalidScope  ErrorKind = "invalid_scope"
	ErrorKindPKCEMismatch  ErrorKind = "pkce_mismatch"
	ErrorKindValidation    ErrorKind = "validation"
	ErrorKindUnauthorized  ErrorKind = "unauthorized"
)

type DomainError interface {
	error
	Kind() ErrorKind
}

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidGrant  = errors.New("invalid grant")
	ErrInvalidClient = errors.New("invalid client")
	ErrInvalidScope  = errors.New("invalid scope")
	ErrPKCEMismatch  = errors.New("pkce mismatch")
	ErrValidation    = errors.New("validation failed")
	ErrUnauthorized  = errors.New("unauthorized")
)

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

func (e NotFoundError) Kind() ErrorKind {
	return ErrorKindNotFound
}

func (e NotFoundError) Unwrap() error {
	return ErrNotFound
}

func ErrorNotFound(format string, args ...any) error {
	return NotFoundError{Message: fmt.Sprintf(format, args...)}
}

type InvalidGrantError struct {
	Message string
}

func (e InvalidGrantError) Error() string {
	return e.Message
}

func (e InvalidGrantError) Kind() ErrorKind {
	return ErrorKindInvalidGrant
}

func (e InvalidGrantError) Unwrap() error {
	return ErrInvalidGrant
}

func ErrorInvalidGrant(format string, args ...any) error {
	return InvalidGrantError{Message: fmt.Sprintf(format, args...)}
}

type InvalidClientError struct {
	Message string
}

func (e InvalidClientError) Error() string {
	return e.Message
}

func (e InvalidClientError) Kind() ErrorKind {
	return ErrorKindInvalidClient
}

func (e InvalidClientError) Unwrap() error {
	return ErrInvalidClient
}

func ErrorInvalidClient(format string, args ...any) error {
	return InvalidClientError{Message: fmt.Sprintf(format, args...)}
}

type InvalidScopeError struct {
	Message string
}

func (e InvalidScopeError) Error() string {
	return e.Message
}

func (e InvalidScopeError) Kind() ErrorKind {
	return ErrorKindInvalidScope
}

func (e InvalidScopeError) Unwrap() error {
	return ErrInvalidScope
}

func ErrorInvalidScope(format string, args ...any) error {
	return InvalidScopeError{Message: fmt.Sprintf(format, args...)}
}

type PKCEMismatchError struct {
	Message string
}

func (e PKCEMismatchError) Error() string {
	return e.Message
}

func (e PKCEMismatchError) Kind() ErrorKind {
	return ErrorKindPKCEMismatch
}

func (e PKCEMismatchError) Unwrap() error {
	return ErrPKCEMismatch
}

func ErrorPKCEMismatch(format string, args ...any) error {
	return PKCEMismatchError{Message: fmt.Sprintf(format, args...)}
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func (e ValidationError) Kind() ErrorKind {
	return ErrorKindValidation
}

func (e ValidationError) Unwrap() error {
	return ErrValidation
}

func ErrorValidation(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}

type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

func (e UnauthorizedError) Kind() ErrorKind {
	return ErrorKindUnauthorized
}

func (e UnauthorizedError) Unwrap() error {
	return ErrUnauthorized
}

func ErrorUnauthorized(format string, args ...any) error {
	return UnauthorizedError{Message: fmt.Sprintf(format, args...)}
}
