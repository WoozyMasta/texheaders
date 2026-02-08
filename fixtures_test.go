package texheaders

import (
	"bytes"
	"os"
	"testing"
)

func TestReadFile_FixtureTexHeadersHasEntries(t *testing.T) {
	t.Parallel()

	got, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	if len(got.Textures) == 0 {
		t.Fatalf("decoded fixture has no textures")
	}
}

func TestReadWrite_BytesEqualTexHeadersFixture(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	in, err := Read(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Read(testdata/texHeaders.bin) error: %v", err)
	}

	var out bytes.Buffer
	if err = Write(&out, in); err != nil {
		t.Fatalf("Write(testdata/texHeaders.bin) error: %v", err)
	}

	if !bytes.Equal(raw, out.Bytes()) {
		t.Fatalf("encoded bytes differ from testdata/texHeaders.bin: got=%d want=%d", out.Len(), len(raw))
	}
}
