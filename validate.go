// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/texheaders

package texheaders

import (
	"errors"
	"fmt"
	"math"
)

// ValidateFile validates file-level and entry-level invariants.
func ValidateFile(f *File) error {
	if f == nil {
		return fmt.Errorf("%w: file is nil", ErrValidation)
	}

	var issues []error
	if f.Magic != "" && f.Magic != FileMagic {
		issues = append(issues, fmt.Errorf("%w: magic=%q want=%q", ErrValidation, f.Magic, FileMagic))
	}

	if f.Version != 0 && f.Version != SupportedVersion {
		issues = append(issues, fmt.Errorf("%w: version=%d want=%d", ErrValidation, f.Version, SupportedVersion))
	}

	if len(f.Textures) > math.MaxUint32 {
		issues = append(issues, fmt.Errorf("%w: texture count out of range: %d", ErrValidation, len(f.Textures)))
	}

	for i := range f.Textures {
		if err := ValidateEntry(&f.Textures[i], i); err != nil {
			issues = append(issues, err)
		}
	}

	if len(issues) == 0 {
		return nil
	}

	return errors.Join(issues...)
}

// ValidateEntry validates one texture entry invariants.
func ValidateEntry(entry *TextureEntry, entryIndex int) error {
	if entry == nil {
		return fmt.Errorf("%w: texture[%d] is nil", ErrValidation, entryIndex)
	}

	var issues []error
	prefix := fmt.Sprintf("texture[%d]", entryIndex)

	if entry.PAAFile == "" {
		issues = append(issues, fmt.Errorf("%w: %s.paa_file is empty", ErrValidation, prefix))
	}

	if entry.PaxFormat > math.MaxUint8 {
		issues = append(issues, fmt.Errorf("%w: %s.pax_format out of uint8 range: %d", ErrValidation, prefix, entry.PaxFormat))
	}

	mipLen, convErr := intToU32Strict(len(entry.MipMaps))
	if convErr != nil {
		issues = append(issues, fmt.Errorf("%w: %s.mipmaps length out of range: %d", ErrValidation, prefix, len(entry.MipMaps)))
		mipLen = 0
	}
	if entry.MipMapCount != mipLen {
		issues = append(issues, fmt.Errorf("%w: %s.mipmap_count=%d len(mipmaps)=%d", ErrValidation, prefix, entry.MipMapCount, mipLen))
	}

	if entry.MipMapCountCopy != mipLen {
		issues = append(issues, fmt.Errorf("%w: %s.mipmap_count_copy=%d len(mipmaps)=%d", ErrValidation, prefix, entry.MipMapCountCopy, mipLen))
	}

	if entry.MipMapCount != entry.MipMapCountCopy {
		issues = append(issues, fmt.Errorf("%w: %s.mipmap_count=%d != mipmap_count_copy=%d", ErrValidation, prefix, entry.MipMapCount, entry.MipMapCountCopy))
	}

	var prevOffset uint32
	for i := range entry.MipMaps {
		m := entry.MipMaps[i]
		mp := fmt.Sprintf("%s.mipmaps[%d]", prefix, i)

		if m.Width == 0 || m.Height == 0 {
			issues = append(issues, fmt.Errorf("%w: %s has zero dimension (%d x %d)", ErrValidation, mp, m.Width, m.Height))
		}

		if m.AlwaysZero != 0 {
			issues = append(issues, fmt.Errorf("%w: %s.always_zero=%d want=0", ErrValidation, mp, m.AlwaysZero))
		}

		if m.AlwaysThree != 3 {
			issues = append(issues, fmt.Errorf("%w: %s.always_three=%d want=3", ErrValidation, mp, m.AlwaysThree))
		}

		if entry.PaxFormat <= math.MaxUint8 && uint32(m.PaxFormat) != entry.PaxFormat {
			issues = append(issues, fmt.Errorf("%w: %s.pax_format=%d entry.pax_format=%d", ErrValidation, mp, m.PaxFormat, entry.PaxFormat))
		}

		if i > 0 && m.DataOffset < prevOffset {
			issues = append(issues, fmt.Errorf("%w: %s.data_offset=%d is less than previous=%d", ErrValidation, mp, m.DataOffset, prevOffset))
		}

		prevOffset = m.DataOffset
	}

	if len(issues) == 0 {
		return nil
	}

	return errors.Join(issues...)
}
