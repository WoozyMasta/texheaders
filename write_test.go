package texheaders

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestReadWriteRead_RoundTripModel(t *testing.T) {
	t.Parallel()

	in, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	var out bytes.Buffer
	if err = Write(&out, in); err != nil {
		t.Fatalf("Write(roundtrip) error: %v", err)
	}

	got, err := Read(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("Read(roundtrip bytes) error: %v", err)
	}

	if !reflect.DeepEqual(in, got) {
		t.Fatalf("roundtrip model mismatch")
	}
}

func TestReadWrite_BytesEqualFixture(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	in, err := Read(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Read(fixture bytes) error: %v", err)
	}

	var out bytes.Buffer
	if err = Write(&out, in); err != nil {
		t.Fatalf("Write(bytes parity) error: %v", err)
	}

	if !bytes.Equal(raw, out.Bytes()) {
		t.Fatalf("encoded bytes differ from fixture: got=%d want=%d", out.Len(), len(raw))
	}
}
