package texheaders

import (
	"errors"
	"testing"
)

func TestValidateFile_OK(t *testing.T) {
	t.Parallel()

	f, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	if err = ValidateFile(f); err != nil {
		t.Fatalf("ValidateFile(valid fixture) error: %v", err)
	}
}

func TestValidateFile_Nil(t *testing.T) {
	t.Parallel()

	err := ValidateFile(nil)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ValidateFile(nil) error = %v, want %v", err, ErrValidation)
	}
}

func TestValidateFile_InvalidHeader(t *testing.T) {
	t.Parallel()

	f, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	f.Magic = "XXXX"
	f.Version = 99
	err = ValidateFile(f)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ValidateFile(invalid header) error = %v, want %v", err, ErrValidation)
	}
}

func TestValidateEntry_MipCountMismatch(t *testing.T) {
	t.Parallel()

	f, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	e := f.Textures[0]
	e.MipMapCount++
	err = ValidateEntry(&e, 0)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ValidateEntry(mip count mismatch) error = %v, want %v", err, ErrValidation)
	}
}

func TestValidateEntry_InvalidMipConstants(t *testing.T) {
	t.Parallel()

	f, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	e := f.Textures[0]
	e.MipMaps[0].AlwaysThree = 2
	e.MipMaps[0].AlwaysZero = 1
	err = ValidateEntry(&e, 0)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ValidateEntry(invalid mip constants) error = %v, want %v", err, ErrValidation)
	}
}
