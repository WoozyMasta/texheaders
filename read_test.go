package texheaders

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestReadFile_Fixture(t *testing.T) {
	t.Parallel()

	f, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	if f.Magic != FileMagic {
		t.Fatalf("magic = %q, want %q", f.Magic, FileMagic)
	}

	if f.Version != SupportedVersion {
		t.Fatalf("version = %d, want %d", f.Version, SupportedVersion)
	}

	if len(f.Textures) != 46 {
		t.Fatalf("textures = %d, want 46", len(f.Textures))
	}

	first := f.Textures[0]
	if first.PAAFile == "" {
		t.Fatalf("first texture path is empty")
	}

	if first.MipMapCount == 0 || len(first.MipMaps) == 0 {
		t.Fatalf("first texture mipmaps are empty")
	}
}

func TestRead_InvalidMagic(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	raw[0] = 'X'

	_, readErr := Read(bytes.NewReader(raw))
	if !errors.Is(readErr, ErrInvalidMagic) {
		t.Fatalf("Read(invalid magic) error = %v, want %v", readErr, ErrInvalidMagic)
	}
}

func TestRead_UnsupportedVersion(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	// version at offsets [4:8]
	raw[4] = 2
	raw[5] = 0
	raw[6] = 0
	raw[7] = 0

	_, readErr := Read(bytes.NewReader(raw))
	if !errors.Is(readErr, ErrUnsupportedVersion) {
		t.Fatalf("Read(unsupported version) error = %v, want %v", readErr, ErrUnsupportedVersion)
	}
}

func TestRead_Truncated(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	raw = raw[:100]

	_, readErr := Read(bytes.NewReader(raw))
	if readErr == nil {
		t.Fatalf("Read(truncated) error = nil, want non-nil")
	}
}
