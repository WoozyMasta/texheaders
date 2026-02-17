// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/texheaders

package texheaders

import "errors"

var (
	// ErrInvalidMagic means the file signature is not "0DHT".
	ErrInvalidMagic = errors.New("invalid texheaders magic")
	// ErrUnsupportedVersion means version is not supported by decoder.
	ErrUnsupportedVersion = errors.New("unsupported texheaders version")
	// ErrInvalidASCIIZ means string payload is missing zero terminator.
	ErrInvalidASCIIZ = errors.New("invalid ASCIIZ payload")
	// ErrTooManyTextures means texture count does not fit uint32 file field.
	ErrTooManyTextures = errors.New("too many texture entries")
	// ErrUnsupportedInputFormat means source texture extension is not supported.
	ErrUnsupportedInputFormat = errors.New("unsupported input texture format")
	// ErrPACUnsupported means .pac source support is not implemented yet.
	ErrPACUnsupported = errors.New(".pac source is not supported")
	// ErrValidation means semantic model validation failed.
	ErrValidation = errors.New("texheaders validation failed")
)
