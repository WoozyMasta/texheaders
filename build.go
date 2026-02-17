// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/texheaders

package texheaders

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/woozymasta/paa"
)

// WorkersAuto enables automatic worker selection for BuildOptions.Workers.
const WorkersAuto = -1

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
	// Workers controls parallelism in Build.
	//  - Workers <= 1 disables parallel build (default, no worker overhead).
	//  - Workers == WorkersAuto selects workers automatically from host CPU count.
	//  - Workers > 1 enables parallel entry build with that worker count.
	Workers int `json:"workers,omitempty" yaml:"workers,omitempty"`
}

// BuildIssue reports one skipped input in lenient mode.
type BuildIssue struct {
	// Path is the path of the skipped input.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Error is the error message of the skipped input.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Builder builds texheaders file from source texture files.
type Builder struct {
	inputs       []string     // inputs is the list of source texture paths.
	issues       []BuildIssue // issues is the list of skipped inputs.
	opts         BuildOptions // opts is the builder options.
	inputsSorted bool         // inputsSorted tracks whether inputs are already sorted lexicographically.
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
		// Empty or append-increasing input is already sorted.
		inputsSorted: true,
		issues:       make([]BuildIssue, 0),
	}
}

// Append registers one source texture path for build.
func (b *Builder) Append(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrEmptyInputPath
	}

	if b.inputsSorted && len(b.inputs) > 0 && b.inputs[len(b.inputs)-1] > path {
		b.inputsSorted = false
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
	if !b.inputsSorted {
		sort.Strings(b.inputs)
		b.inputsSorted = true
	}

	b.issues = b.issues[:0]

	file := &File{
		Magic:    FileMagic,
		Version:  SupportedVersion,
		Textures: make([]TextureEntry, 0, len(b.inputs)),
	}

	if len(b.inputs) == 0 {
		return file, nil
	}

	workers := resolveBuildWorkers(b.opts.Workers, len(b.inputs))

	// Handle serial build.
	if workers <= 1 {
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
	if workers > len(b.inputs) {
		workers = len(b.inputs)
	}

	// Initialize result arrays.
	entries := make([]TextureEntry, len(b.inputs))
	errs := make([]error, len(b.inputs))
	jobs := make(chan int, len(b.inputs))
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for i := range jobs {
				entry, err := b.buildEntry(b.inputs[i])
				if err != nil {
					errs[i] = err
					continue
				}

				entries[i] = entry
			}
		}()
	}

	// Dispatch jobs to workers.
	for i := range b.inputs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	// Collect results from workers.
	for i, in := range b.inputs {
		if errs[i] == nil {
			file.Textures = append(file.Textures, entries[i])
			continue
		}

		if b.opts.SkipInvalid {
			b.issues = append(b.issues, BuildIssue{
				Path:  in,
				Error: errs[i].Error(),
			})
			continue
		}

		return nil, fmt.Errorf("build %q: %w", in, errs[i])
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

	meta, err := paa.DecodeMetadataHeaders(fh)
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

	assignColorHeaders(&entry, meta)
	assignFlagHeaders(&entry, meta)
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

// assignColorHeaders maps PAA header color metadata into entry color fields.
func assignColorHeaders(entry *TextureEntry, meta *paa.MetadataHeaders) {
	if meta.HasAverageColor {
		entry.AverageColor = meta.AverageColor
	}

	if meta.HasMaxColor {
		entry.MaxColor = meta.MaxColor
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

// assignFlagHeaders maps GALF metadata flags into alpha booleans.
func assignFlagHeaders(entry *TextureEntry, meta *paa.MetadataHeaders) {
	if !meta.HasGALF {
		entry.IsAlpha = false
		entry.IsTransparent = false
		entry.IsAlphaNonOpaque = false
		return
	}

	flags := meta.GALF
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

// resolveBuildWorkers resolves requested worker setting to an effective count.
func resolveBuildWorkers(requested, fileCount int) int {
	if fileCount <= 1 {
		return 1
	}

	switch {
	case requested == WorkersAuto:
		return autoBuildWorkers(fileCount)
	case requested <= 1:
		return 1
	default:
		return min(requested, fileCount)
	}
}

// autoBuildWorkers chooses worker count from CPU parallelism and file count.
func autoBuildWorkers(fileCount int) int {
	workers := runtime.GOMAXPROCS(0) / 4
	workers = max(workers, 2)
	workers = min(workers, fileCount)

	return floorPow2(workers)
}

// floorPow2 returns the largest power of two not greater than v.
func floorPow2(v int) int {
	if v <= 1 {
		return 1
	}

	p := 1
	for p<<1 <= v {
		p <<= 1
	}

	return p
}

// intToU32Strict safely converts int to uint32 without unsafe cast.
func intToU32Strict(v int) (uint32, error) {
	if v < 0 || uint64(v) > math.MaxUint32 {
		return 0, fmt.Errorf("value out of uint32 range: %d", v)
	}

	return uint32(v), nil
}

// int64ToU32Strict safely converts int64 to uint32 without unsafe cast.
func int64ToU32Strict(v int64) (uint32, error) {
	if v < 0 || uint64(v) > math.MaxUint32 {
		return 0, fmt.Errorf("value out of uint32 range: %d", v)
	}

	return uint32(v), nil
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
