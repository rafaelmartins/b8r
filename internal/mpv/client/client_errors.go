package client

import (
	"errors"
	"strings"
)

var (
	ErrMpvSuccess             = errors.New("success")
	ErrMpvEventQueueFull      = errors.New("event queue full")
	ErrMpvNoMem               = errors.New("memory allocation failed")
	ErrMpvUninitialized       = errors.New("core not uninitialized")
	ErrMpvInvalidParameter    = errors.New("invalid parameter")
	ErrMpvOptionNotFound      = errors.New("option not found")
	ErrMpvOptionFormat        = errors.New("unsupported format for accessing option")
	ErrMpvOptionError         = errors.New("error setting option")
	ErrMpvPropertyNotFound    = errors.New("property not found")
	ErrMpvPropertyFormat      = errors.New("unsupported format for accessing property")
	ErrMpvPropertyUnavailable = errors.New("property unavailable")
	ErrMpvPropertyError       = errors.New("error accessing property")
	ErrMpvCommand             = errors.New("error running command")
	ErrMpvLoadingFailed       = errors.New("loading failed")
	ErrMpvAoInitFailed        = errors.New("audio output initialization failed")
	ErrMpvVoInitFailed        = errors.New("video output initialization failed")
	ErrMpvNothingToPlay       = errors.New("no audio or video data played")
	ErrMpvUnknownFormat       = errors.New("unrecognized file format")
	ErrMpvUnsupported         = errors.New("not supported")
	ErrMpvNotImplemented      = errors.New("operation not implemented")
	ErrMpvGeneric             = errors.New("something happened")

	errorReg = []error{
		ErrMpvSuccess,
		ErrMpvEventQueueFull,
		ErrMpvNoMem,
		ErrMpvUninitialized,
		ErrMpvInvalidParameter,
		ErrMpvOptionNotFound,
		ErrMpvOptionFormat,
		ErrMpvOptionError,
		ErrMpvPropertyNotFound,
		ErrMpvPropertyFormat,
		ErrMpvPropertyUnavailable,
		ErrMpvPropertyError,
		ErrMpvCommand,
		ErrMpvLoadingFailed,
		ErrMpvAoInitFailed,
		ErrMpvVoInitFailed,
		ErrMpvNothingToPlay,
		ErrMpvUnknownFormat,
		ErrMpvUnsupported,
		ErrMpvNotImplemented,
		ErrMpvGeneric,
	}
)

func matchError(e string) error {
	for _, err := range errorReg {
		if err.Error() == strings.TrimSpace(e) {
			return err
		}
	}
	return ErrMpvGeneric
}
