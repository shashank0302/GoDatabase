package storage

import "errors"

var (
	// ErrInvalidStorageType is returned when an invalid storage type is specified
	ErrInvalidStorageType = errors.New("invalid storage type")
	
	// ErrKeyNotFound is returned when a key is not found
	ErrKeyNotFound = errors.New("key not found")
	
	// ErrKeyExists is returned when a key already exists
	ErrKeyExists = errors.New("key already exists")
	
	// ErrInvalidDatabase is returned when the database file is invalid
	ErrInvalidDatabase = errors.New("invalid database file")
	
	// ErrUnsupportedVersion is returned when the database version is not supported
	ErrUnsupportedVersion = errors.New("unsupported database version")
) 