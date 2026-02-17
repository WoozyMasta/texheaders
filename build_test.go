package texheaders

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuilder_BuildMatchesFixtureJSON(t *testing.T) {
	t.Parallel()

	wantFile, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	baseDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("filepath.Abs(testdata) error: %v", err)
	}

	b := NewBuilder(BuildOptions{
		BaseDir:        baseDir,
		LowercasePaths: true,
		BackslashPaths: true,
	})

	for _, tex := range wantFile.Textures {
		absPath := filepath.Join(baseDir, stringsFromBackslashes(tex.PAAFile))
		if err = b.Append(absPath); err != nil {
			t.Fatalf("Append(%q) error: %v", absPath, err)
		}
	}

	got, err := b.Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	wantMap := mapEntriesByPath(wantFile.Textures)
	gotMap := mapEntriesByPath(got.Textures)
	for path, wantEntry := range wantMap {
		gotEntry, ok := gotMap[path]
		if !ok {
			t.Fatalf("missing generated entry for %q", path)
		}

		if err = assertEntryEqual(path, wantEntry, gotEntry); err != nil {
			t.Fatalf("entry mismatch: %v", err)
		}
	}
}

func TestBuilder_SkipInvalid(t *testing.T) {
	t.Parallel()

	baseDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("filepath.Abs(testdata) error: %v", err)
	}

	b := NewBuilder(BuildOptions{
		BaseDir:        baseDir,
		SkipInvalid:    true,
		LowercasePaths: true,
		BackslashPaths: true,
	})

	if err = b.Append(filepath.Join(baseDir, "test_co.paa")); err != nil {
		t.Fatalf("Append(valid) error: %v", err)
	}

	invalidPath := filepath.Join(t.TempDir(), "not_a_texture.txt")
	if err = os.WriteFile(invalidPath, []byte("not a texture"), 0o600); err != nil {
		t.Fatalf("WriteFile(invalid fixture) error: %v", err)
	}

	if err = b.Append(invalidPath); err != nil {
		t.Fatalf("Append(invalid) error: %v", err)
	}

	got, err := b.Build()
	if err != nil {
		t.Fatalf("Build(skip invalid) error: %v", err)
	}

	if len(got.Textures) != 1 {
		t.Fatalf("textures = %d, want 1", len(got.Textures))
	}

	if len(b.Issues()) != 1 {
		t.Fatalf("issues = %d, want 1", len(b.Issues()))
	}
}

func TestBuilder_FailFastInvalid(t *testing.T) {
	t.Parallel()

	baseDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("filepath.Abs(testdata) error: %v", err)
	}

	b := NewBuilder(BuildOptions{
		BaseDir:        baseDir,
		LowercasePaths: true,
		BackslashPaths: true,
	})

	invalidPath := filepath.Join(t.TempDir(), "not_a_texture.txt")
	if err = os.WriteFile(invalidPath, []byte("not a texture"), 0o600); err != nil {
		t.Fatalf("WriteFile(invalid fixture) error: %v", err)
	}

	if err = b.Append(invalidPath); err != nil {
		t.Fatalf("Append(invalid) error: %v", err)
	}

	_, err = b.Build()
	if !errors.Is(err, ErrUnsupportedInputFormat) {
		t.Fatalf("Build(fail fast) error = %v, want %v", err, ErrUnsupportedInputFormat)
	}
}

func TestBuilder_AppendMany(t *testing.T) {
	t.Parallel()

	b := NewBuilder(BuildOptions{})
	if err := b.AppendMany("a.paa", "b.paa", "c.paa"); err != nil {
		t.Fatalf("AppendMany(valid) error: %v", err)
	}

	got := b.Inputs()
	if len(got) != 3 {
		t.Fatalf("inputs len = %d, want 3", len(got))
	}

	if got[0] != "a.paa" || got[1] != "b.paa" || got[2] != "c.paa" {
		t.Fatalf("inputs order mismatch: %#v", got)
	}
}

func TestBuilder_AppendManyFailFastOnInvalid(t *testing.T) {
	t.Parallel()

	b := NewBuilder(BuildOptions{})
	err := b.AppendMany("ok.paa", " ", "never-added.paa")
	if err == nil {
		t.Fatalf("AppendMany(invalid) error = nil, want error")
	}

	got := b.Inputs()
	if len(got) != 1 {
		t.Fatalf("inputs len = %d, want 1", len(got))
	}

	if got[0] != "ok.paa" {
		t.Fatalf("first input = %q, want %q", got[0], "ok.paa")
	}
}

func TestBuilder_ParallelParity(t *testing.T) {
	t.Parallel()

	wantFile, err := ReadFile("testdata/texHeaders.bin")
	if err != nil {
		t.Fatalf("ReadFile(testdata/texHeaders.bin) error: %v", err)
	}

	baseDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("filepath.Abs(testdata) error: %v", err)
	}

	serial := NewBuilder(BuildOptions{
		BaseDir:        baseDir,
		LowercasePaths: true,
		BackslashPaths: true,
		Workers:        1,
	})
	parallel := NewBuilder(BuildOptions{
		BaseDir:        baseDir,
		LowercasePaths: true,
		BackslashPaths: true,
		Workers:        4,
	})

	for _, tex := range wantFile.Textures {
		absPath := filepath.Join(baseDir, stringsFromBackslashes(tex.PAAFile))
		if err = serial.Append(absPath); err != nil {
			t.Fatalf("serial Append(%q) error: %v", absPath, err)
		}
		if err = parallel.Append(absPath); err != nil {
			t.Fatalf("parallel Append(%q) error: %v", absPath, err)
		}
	}

	serialOut, err := serial.Build()
	if err != nil {
		t.Fatalf("serial Build() error: %v", err)
	}
	parallelOut, err := parallel.Build()
	if err != nil {
		t.Fatalf("parallel Build() error: %v", err)
	}

	if len(serialOut.Textures) != len(parallelOut.Textures) {
		t.Fatalf("textures length mismatch: serial=%d parallel=%d", len(serialOut.Textures), len(parallelOut.Textures))
	}

	for i := range serialOut.Textures {
		if err = assertEntryEqual(serialOut.Textures[i].PAAFile, serialOut.Textures[i], parallelOut.Textures[i]); err != nil {
			t.Fatalf("parallel parity mismatch: %v", err)
		}
	}
}

func TestResolveBuildWorkers(t *testing.T) {
	oldProcs := runtime.GOMAXPROCS(0)
	runtime.GOMAXPROCS(20)
	defer runtime.GOMAXPROCS(oldProcs)

	tests := []struct {
		name      string
		requested int
		fileCount int
		want      int
	}{
		{name: "default serial", requested: 0, fileCount: 100, want: 1},
		{name: "explicit serial", requested: 1, fileCount: 100, want: 1},
		{name: "explicit cap by files", requested: 8, fileCount: 3, want: 3},
		{name: "auto large set", requested: WorkersAuto, fileCount: 100, want: 4}, // 20/4=5 -> floorPow2=4
		{name: "auto small set", requested: WorkersAuto, fileCount: 3, want: 2},
		{name: "single file always serial", requested: WorkersAuto, fileCount: 1, want: 1},
	}

	for _, tt := range tests {
		got := resolveBuildWorkers(tt.requested, tt.fileCount)
		if got != tt.want {
			t.Fatalf("%s: resolveBuildWorkers(%d, %d)=%d, want %d", tt.name, tt.requested, tt.fileCount, got, tt.want)
		}
	}
}

func mapEntriesByPath(in []TextureEntry) map[string]TextureEntry {
	out := make(map[string]TextureEntry, len(in))
	for _, e := range in {
		out[e.PAAFile] = e
	}

	return out
}

func stringsFromBackslashes(in string) string {
	return filepath.FromSlash(strings.ReplaceAll(in, "\\", "/"))
}

func assertEntryEqual(path string, want, got TextureEntry) error {
	if want.PAAFile != got.PAAFile ||
		want.ColorPaletteCount != got.ColorPaletteCount ||
		want.PalettePtr != got.PalettePtr ||
		want.ClampFlags != got.ClampFlags ||
		want.TransparentColor != got.TransparentColor ||
		want.HasMaxCtagg != got.HasMaxCtagg ||
		want.IsAlpha != got.IsAlpha ||
		want.IsTransparent != got.IsTransparent ||
		want.IsAlphaNonOpaque != got.IsAlphaNonOpaque ||
		want.MipMapCount != got.MipMapCount ||
		want.PaxFormat != got.PaxFormat ||
		want.LittleEndian != got.LittleEndian ||
		want.IsPAA != got.IsPAA ||
		want.PaxSuffixType != got.PaxSuffixType ||
		want.MipMapCountCopy != got.MipMapCountCopy ||
		want.PaxFileSize != got.PaxFileSize ||
		want.AverageColor != got.AverageColor ||
		want.MaxColor != got.MaxColor {
		return errors.New(path + ": non-float entry field mismatch")
	}

	for i := range want.AverageColorF {
		if !float32Near(want.AverageColorF[i], got.AverageColorF[i], 1e-6) {
			return errors.New(path + ": average_color_f mismatch")
		}
	}

	if len(want.MipMaps) != len(got.MipMaps) {
		return errors.New(path + ": mipmaps length mismatch")
	}

	for i := range want.MipMaps {
		if want.MipMaps[i] != got.MipMaps[i] {
			return errors.New(path + ": mipmap mismatch")
		}
	}

	return nil
}

func float32Near(a, b, eps float32) bool {
	return float32(math.Abs(float64(a-b))) <= eps
}
