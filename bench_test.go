package texheaders

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkDecodeFixture(b *testing.B) {
	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		b.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	b.SetBytes(int64(len(raw)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err = Read(bytes.NewReader(raw)); err != nil {
			b.Fatalf("Read(fixture) error: %v", err)
		}
	}
}

func BenchmarkEncodeDecodedFixture(b *testing.B) {
	raw, err := os.ReadFile("testdata/texHeaders.bin")
	if err != nil {
		b.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	f, err := Read(bytes.NewReader(raw))
	if err != nil {
		b.Fatalf("Read(fixture) error: %v", err)
	}

	b.SetBytes(int64(len(raw)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var out bytes.Buffer
		if err = Write(&out, f); err != nil {
			b.Fatalf("Write(decoded fixture) error: %v", err)
		}
	}
}

func BenchmarkBuildFromAppendedFiles(b *testing.B) {
	fixture, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		b.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	baseDir, err := filepath.Abs("testdata")
	if err != nil {
		b.Fatalf("filepath.Abs(testdata) error: %v", err)
	}

	inputs := make([]string, 0, len(fixture.Textures))
	for _, tex := range fixture.Textures {
		inputs = append(inputs, filepath.Join(baseDir, filepath.FromSlash(strings.ReplaceAll(tex.PAAFile, "\\", "/"))))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder(BuildOptions{
			BaseDir:        baseDir,
			LowercasePaths: true,
			BackslashPaths: true,
		})

		for _, in := range inputs {
			if err = builder.Append(in); err != nil {
				b.Fatalf("Append(%q) error: %v", in, err)
			}
		}

		if _, err = builder.Build(); err != nil {
			b.Fatalf("Build() error: %v", err)
		}
	}
}
