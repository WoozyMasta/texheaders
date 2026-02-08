package texheaders

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/woozymasta/paa"
)

// BuildOptions controls builder behavior.
type BuildOptions struct {
	// SuffixOverrides maps normalized path to forced suffix type value.
	SuffixOverrides map[string]uint32 `json:"suffix_overrides,omitempty" yaml:"suffix_overrides,omitempty"`
	// BaseDir is used for relative paths stored in PAAFile.
	// If empty, absolute input paths are made relative to current working dir when possible.
	BaseDir string `json:"base_dir,omitempty" yaml:"base_dir,omitempty"`
	// SkipInvalid keeps building when one input fails.
	SkipInvalid bool `json:"skip_invalid,omitempty" yaml:"skip_invalid,omitempty"`
	// LowercasePaths stores entry paths in lowercase.
	LowercasePaths bool `json:"lowercase_paths,omitempty" yaml:"lowercase_paths,omitempty"`
	// BackslashPaths stores entry paths with backslash separators.
	BackslashPaths bool `json:"backslash_paths,omitempty" yaml:"backslash_paths,omitempty"`
}

// BuildIssue reports one skipped input in lenient mode.
type BuildIssue struct {
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Builder builds texheaders file from source texture files.
type Builder struct {
	opts   BuildOptions
	inputs []string
	issues []BuildIssue
}

// NewBuilder creates a new builder with options.
func NewBuilder(opts BuildOptions) *Builder {
	if !opts.LowercasePaths {
		opts.LowercasePaths = true
	}

	if !opts.BackslashPaths {
		opts.BackslashPaths = true
	}

	return &Builder{
		opts:   opts,
		inputs: make([]string, 0, 16),
		issues: make([]BuildIssue, 0),
	}
}

// Append registers one source texture path for build.
func (b *Builder) Append(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("empty input path")
	}

	b.inputs = append(b.inputs, path)
	return nil
}

// AppendMany registers multiple source texture paths for build.
func (b *Builder) AppendMany(paths ...string) error {
	for _, path := range paths {
		if err := b.Append(path); err != nil {
			return err
		}
	}

	return nil
}

// Inputs returns a copy of currently appended paths.
func (b *Builder) Inputs() []string {
	out := make([]string, len(b.inputs))
	copy(out, b.inputs)
	return out
}

// Issues returns skipped input issues collected during Build with SkipInvalid=true.
func (b *Builder) Issues() []BuildIssue {
	out := make([]BuildIssue, len(b.issues))
	copy(out, b.issues)
	return out
}

// Build compiles appended source files into texheaders model.
func (b *Builder) Build() (*File, error) {
	sort.Strings(b.inputs)
	b.issues = b.issues[:0]

	file := &File{
		Magic:    FileMagic,
		Version:  SupportedVersion,
		Textures: make([]TextureEntry, 0, len(b.inputs)),
	}

	for _, in := range b.inputs {
		entry, err := b.buildEntry(in)
		if err != nil {
			if b.opts.SkipInvalid {
				b.issues = append(b.issues, BuildIssue{
					Path:  in,
					Error: err.Error(),
				})
				continue
			}

			return nil, fmt.Errorf("build %q: %w", in, err)
		}

		file.Textures = append(file.Textures, entry)
	}

	return file, nil
}

// Write builds and writes texheaders model to stream.
func (b *Builder) Write(w io.Writer) error {
	f, err := b.Build()
	if err != nil {
		return err
	}

	if err = Write(w, f); err != nil {
		return err
	}

	return nil
}

// WriteFile builds and writes texheaders model to file.
func (b *Builder) WriteFile(path string) error {
	f, err := b.Build()
	if err != nil {
		return err
	}

	if err = WriteFile(path, f); err != nil {
		return err
	}

	return nil
}

// buildEntry builds one texture entry from one source file.
func (b *Builder) buildEntry(path string) (TextureEntry, error) {
	var entry TextureEntry

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".paa":
	case ".pac":
		return entry, fmt.Errorf("%w: %s", ErrPACUnsupported, path)
	default:
		return entry, fmt.Errorf("%w: %s", ErrUnsupportedInputFormat, path)
	}

	fh, err := os.Open(path)
	if err != nil {
		return entry, fmt.Errorf("open source: %w", err)
	}

	defer func() {
		_ = fh.Close()
	}()

	info, err := fh.Stat()
	if err != nil {
		return entry, fmt.Errorf("stat source: %w", err)
	}

	meta, err := paa.DecodeMetadata(fh)
	if err != nil {
		return entry, fmt.Errorf("scan paa metadata: %w", err)
	}

	paxFormat, err := paxTypeToU8(meta.Type)
	if err != nil {
		return entry, err
	}

	rel := b.normalizePath(path)
	entry.ColorPaletteCount = 1
	entry.PalettePtr = 0
	entry.ClampFlags = 0
	entry.TransparentColor = 0xFFFFFFFF
	entry.LittleEndian = true
	entry.IsPAA = strings.EqualFold(ext, ".paa")
	entry.PAAFile = rel
	entry.PaxFormat = uint32(meta.Type)
	entry.PaxSuffixType = b.resolveSuffixType(rel)
	entry.PaxFileSize, err = int64ToU32Strict(info.Size())
	if err != nil {
		return entry, err
	}

	assignColorTags(&entry, meta.Taggs)
	assignFlagTag(&entry, meta.Taggs)
	if err = assignMipmaps(&entry, meta.MipHeaders, paxFormat); err != nil {
		return entry, err
	}

	return entry, nil
}

// resolveSuffixType resolves suffix type with optional per-path override.
func (b *Builder) resolveSuffixType(rel string) uint32 {
	key := rel
	if b.opts.LowercasePaths {
		key = strings.ToLower(key)
	}

	if b.opts.SuffixOverrides != nil {
		if v, ok := b.opts.SuffixOverrides[key]; ok {
			return v
		}
	}

	v, _ := GuessSuffixTypeFromPath(rel)
	return v
}

// normalizePath returns path stored into PAAFile field.
func (b *Builder) normalizePath(in string) string {
	cleanIn := filepath.Clean(in)
	baseDir := strings.TrimSpace(b.opts.BaseDir)

	rel := cleanIn
	if baseDir != "" {
		if r, err := filepath.Rel(baseDir, cleanIn); err == nil {
			rel = r
		}
	} else if filepath.IsAbs(cleanIn) {
		if cwd, err := os.Getwd(); err == nil {
			if r, relErr := filepath.Rel(cwd, cleanIn); relErr == nil {
				rel = r
			}
		}
	}

	if b.opts.BackslashPaths {
		rel = strings.ReplaceAll(rel, "/", "\\")
	}

	rel = strings.TrimPrefix(rel, ".\\")
	if b.opts.LowercasePaths {
		rel = strings.ToLower(rel)
	}

	return rel
}

// assignColorTags maps CGVA/CXAM tags into entry color fields.
func assignColorTags(entry *TextureEntry, tags map[string][]byte) {
	avg, ok := tags["CGVA"]
	if ok && len(avg) >= 4 {
		copy(entry.AverageColor[:], avg[:4])
	}

	maxTag, ok := tags["CXAM"]
	if ok && len(maxTag) >= 4 {
		copy(entry.MaxColor[:], maxTag[:4])
		entry.HasMaxCtagg = true
	} else {
		entry.MaxColor = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
		entry.HasMaxCtagg = false
	}

	// BI stores byte color as B,G,R,A while float tuple is exposed as R,G,B,A.
	entry.AverageColorF[0] = float32(entry.AverageColor[2]) / 255.0
	entry.AverageColorF[1] = float32(entry.AverageColor[1]) / 255.0
	entry.AverageColorF[2] = float32(entry.AverageColor[0]) / 255.0
	entry.AverageColorF[3] = float32(entry.AverageColor[3]) / 255.0
}

// assignFlagTag maps GALF flags into alpha booleans.
func assignFlagTag(entry *TextureEntry, tags map[string][]byte) {
	flagsRaw, ok := tags["GALF"]
	if !ok || len(flagsRaw) < 4 {
		entry.IsAlpha = false
		entry.IsTransparent = false
		entry.IsAlphaNonOpaque = false
		return
	}

	flags := binary.LittleEndian.Uint32(flagsRaw[:4])
	entry.IsAlpha = (flags & 1) != 0
	entry.IsTransparent = (flags & 2) != 0
	entry.IsAlphaNonOpaque = entry.IsAlpha && entry.AverageColor[3] < 0x80
}

// assignMipmaps maps scanned mip headers into texheaders mip descriptors.
func assignMipmaps(entry *TextureEntry, mips []paa.MipHeader, paxFormat uint8) error {
	entry.MipMaps = make([]MipMap, 0, len(mips))

	for _, mip := range mips {
		entry.MipMaps = append(entry.MipMaps, MipMap{
			Width:       mip.Width,
			Height:      mip.Height,
			AlwaysZero:  0,
			PaxFormat:   paxFormat,
			AlwaysThree: 3,
			DataOffset:  mip.Offset,
		})
	}

	mipCount, err := intToU32Strict(len(entry.MipMaps))
	if err != nil {
		return err
	}

	entry.MipMapCount = mipCount
	entry.MipMapCountCopy = entry.MipMapCount

	return nil
}

// intToU32Strict safely converts int to uint32 without unsafe cast.
func intToU32Strict(v int) (uint32, error) {
	if v < 0 || uint64(v) > math.MaxUint32 {
		return 0, fmt.Errorf("value out of uint32 range: %d", v)
	}

	buf := [4]byte{
		byte(v),
		byte(v >> 8),
		byte(v >> 16),
		byte(v >> 24),
	}

	return binary.LittleEndian.Uint32(buf[:]), nil
}

// int64ToU32Strict safely converts int64 to uint32 without unsafe cast.
func int64ToU32Strict(v int64) (uint32, error) {
	if v < 0 || uint64(v) > math.MaxUint32 {
		return 0, fmt.Errorf("value out of uint32 range: %d", v)
	}

	buf := [4]byte{
		byte(v),
		byte(v >> 8),
		byte(v >> 16),
		byte(v >> 24),
	}

	return binary.LittleEndian.Uint32(buf[:]), nil
}

// paxTypeToU8 maps known paa pax types to uint8 texheaders format field.
func paxTypeToU8(t paa.PaxType) (uint8, error) {
	switch t {
	case paa.PaxGRAYA:
		return 1, nil
	case paa.PaxARGBA5:
		return 3, nil
	case paa.PaxARGB4:
		return 4, nil
	case paa.PaxARGB8:
		return 5, nil
	case paa.PaxDXT1:
		return 6, nil
	case paa.PaxDXT2:
		return 7, nil
	case paa.PaxDXT3:
		return 8, nil
	case paa.PaxDXT4:
		return 9, nil
	case paa.PaxDXT5:
		return 10, nil
	default:
		return 0, fmt.Errorf("unsupported pax format: %d", t)
	}
}
