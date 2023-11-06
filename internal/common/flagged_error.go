package common

import "errors"

type Flag string

const (
	FlagNotFound        Flag = "not found"
	FlagAlreadyExists   Flag = "already exists"
	FlagInvalidArgument Flag = "invalid argument"
)

type Flagged interface {
	error
	Flag() Flag
}

func FlagError(err error, flag Flag) FlaggedError {
	return FlaggedError{error: err, flag: flag}
}

type FlaggedError struct {
	error
	flag Flag
}

func (e FlaggedError) Unwrap() error {
	return e.error
}

func (e FlaggedError) Flag() Flag {
	return e.flag
}

func IsFlaggedError(err error, flag Flag) bool {
	if err == nil {
		return false
	}

	var flagged Flagged
	if errors.As(err, &flagged) {
		return flagged.Flag() == flag
	}

	return false
}
