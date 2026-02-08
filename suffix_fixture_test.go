package texheaders

import (
	"testing"
)

func TestGuessSuffixTypeFromPath_TexHeadersFixture(t *testing.T) {
	t.Parallel()

	f, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	for _, tex := range f.Textures {
		tex := tex
		t.Run(tex.PAAFile, func(t *testing.T) {
			t.Parallel()

			got, ok := GuessSuffixTypeFromPath(tex.PAAFile)
			if got != tex.PaxSuffixType {
				t.Fatalf("GuessSuffixTypeFromPath(%q) = %d, want %d", tex.PAAFile, got, tex.PaxSuffixType)
			}

			if tex.PaxSuffixType != 0 && !ok {
				t.Fatalf("GuessSuffixTypeFromPath(%q) should be recognized for non-zero suffix", tex.PAAFile)
			}
		})
	}
}
