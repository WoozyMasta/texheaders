// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/texheaders

package texheaders

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

// encoder is a reusable little-endian writer with shared scratch buffer.
type encoder struct {
	w   io.Writer
	tmp [8]byte
}

// WriteFile encodes texHeaders.bin into file path.
func WriteFile(path string, f *File) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}

	defer func() {
		_ = out.Close()
	}()

	if err = Write(out, f); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}

	return nil
}

// Write encodes texHeaders.bin into stream.
func Write(w io.Writer, f *File) error {
	if f == nil {
		return errors.New("file is nil")
	}

	e := encoder{w: w}

	magic := f.Magic
	if magic == "" {
		magic = FileMagic
	}

	if len(magic) != 4 {
		return fmt.Errorf("%w: got %q", ErrInvalidMagic, magic)
	}

	if _, err := e.w.Write([]byte(magic)); err != nil {
		return fmt.Errorf("write magic: %w", err)
	}

	version := f.Version
	if version == 0 {
		version = SupportedVersion
	}

	if version != SupportedVersion {
		return fmt.Errorf("%w: got %d", ErrUnsupportedVersion, version)
	}

	if err := e.writeU32(version); err != nil {
		return fmt.Errorf("write version: %w", err)
	}

	if err := e.writeU32FromInt(len(f.Textures)); err != nil {
		return fmt.Errorf("write texture count: %w", err)
	}

	for i := range f.Textures {
		if err := e.writeTextureEntry(&f.Textures[i]); err != nil {
			return fmt.Errorf("write texture entry %d: %w", i, err)
		}
	}

	return nil
}

// writeTextureEntry encodes one texture entry.
func (e *encoder) writeTextureEntry(entry *TextureEntry) error {
	if err := e.writeU32(entry.ColorPaletteCount); err != nil {
		return fmt.Errorf("write color palette count: %w", err)
	}

	if err := e.writeU32(entry.PalettePtr); err != nil {
		return fmt.Errorf("write palette ptr: %w", err)
	}

	for i := range entry.AverageColorF {
		if err := e.writeF32(entry.AverageColorF[i]); err != nil {
			return fmt.Errorf("write average float color[%d]: %w", i, err)
		}
	}

	if _, err := e.w.Write(entry.AverageColor[:]); err != nil {
		return fmt.Errorf("write average color bytes: %w", err)
	}

	if _, err := e.w.Write(entry.MaxColor[:]); err != nil {
		return fmt.Errorf("write max color bytes: %w", err)
	}

	if err := e.writeU32(entry.ClampFlags); err != nil {
		return fmt.Errorf("write clamp flags: %w", err)
	}

	if err := e.writeU32(entry.TransparentColor); err != nil {
		return fmt.Errorf("write transparent color: %w", err)
	}

	if err := e.writeBool8(entry.HasMaxCtagg); err != nil {
		return fmt.Errorf("write has_max_ctagg: %w", err)
	}

	if err := e.writeBool8(entry.IsAlpha); err != nil {
		return fmt.Errorf("write is_alpha: %w", err)
	}

	if err := e.writeBool8(entry.IsTransparent); err != nil {
		return fmt.Errorf("write is_transparent: %w", err)
	}

	if err := e.writeBool8(entry.IsAlphaNonOpaque); err != nil {
		return fmt.Errorf("write is_alpha_non_opaque: %w", err)
	}

	if err := e.writeU32(entry.MipMapCount); err != nil {
		return fmt.Errorf("write mip count: %w", err)
	}

	if err := e.writeU32(entry.PaxFormat); err != nil {
		return fmt.Errorf("write pax format: %w", err)
	}

	if err := e.writeBool8(entry.LittleEndian); err != nil {
		return fmt.Errorf("write little_endian: %w", err)
	}

	if err := e.writeBool8(entry.IsPAA); err != nil {
		return fmt.Errorf("write is_paa: %w", err)
	}

	if err := e.writeASCIIZ(entry.PAAFile); err != nil {
		return fmt.Errorf("write paa path: %w", err)
	}

	if err := e.writeU32(entry.PaxSuffixType); err != nil {
		return fmt.Errorf("write pax suffix type: %w", err)
	}

	if err := e.writeU32(entry.MipMapCountCopy); err != nil {
		return fmt.Errorf("write mip count copy: %w", err)
	}

	for i := range entry.MipMaps {
		if err := e.writeMipMap(&entry.MipMaps[i]); err != nil {
			return fmt.Errorf("write mipmap %d: %w", i, err)
		}
	}

	if err := e.writeU32(entry.PaxFileSize); err != nil {
		return fmt.Errorf("write pax file size: %w", err)
	}

	return nil
}

// writeMipMap encodes one mip descriptor.
func (e *encoder) writeMipMap(m *MipMap) error {
	if err := e.writeU16(m.Width); err != nil {
		return fmt.Errorf("write width: %w", err)
	}

	if err := e.writeU16(m.Height); err != nil {
		return fmt.Errorf("write height: %w", err)
	}

	if err := e.writeU16(m.AlwaysZero); err != nil {
		return fmt.Errorf("write always zero: %w", err)
	}

	if err := e.writeU8(m.PaxFormat); err != nil {
		return fmt.Errorf("write mip pax format: %w", err)
	}

	if err := e.writeU8(m.AlwaysThree); err != nil {
		return fmt.Errorf("write always three: %w", err)
	}

	if err := e.writeU32(m.DataOffset); err != nil {
		return fmt.Errorf("write data offset: %w", err)
	}

	return nil
}

// writeASCIIZ writes zero-terminated string.
func (e *encoder) writeASCIIZ(s string) error {
	if _, err := e.w.Write([]byte(s)); err != nil {
		return err
	}

	if err := e.writeU8(0); err != nil {
		return err
	}

	return nil
}

func (e *encoder) writeU8(v uint8) error {
	e.tmp[0] = v
	_, err := e.w.Write(e.tmp[:1])
	return err
}

func (e *encoder) writeBool8(v bool) error {
	if v {
		return e.writeU8(1)
	}

	return e.writeU8(0)
}

func (e *encoder) writeU16(v uint16) error {
	binary.LittleEndian.PutUint16(e.tmp[:2], v)
	_, err := e.w.Write(e.tmp[:2])
	return err
}

func (e *encoder) writeU32(v uint32) error {
	binary.LittleEndian.PutUint32(e.tmp[:4], v)
	_, err := e.w.Write(e.tmp[:4])
	return err
}

func (e *encoder) writeF32(v float32) error {
	return e.writeU32(math.Float32bits(v))
}

// writeU32FromInt writes a uint32 field from int with strict bounds check.
func (e *encoder) writeU32FromInt(v int) error {
	u32, err := intToU32Strict(v)
	if err != nil {
		return fmt.Errorf("%w: %d", ErrTooManyTextures, v)
	}

	return e.writeU32(u32)
}
