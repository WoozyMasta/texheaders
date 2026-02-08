package texheaders

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// decoder is a reusable little-endian reader with shared scratch buffer.
type decoder struct {
	r   io.Reader
	tmp [8]byte
}

// ReadFile decodes texHeaders.bin from file path.
func ReadFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}

	defer func() {
		_ = f.Close()
	}()

	file, err := Read(f)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	return file, nil
}

// Read decodes texHeaders.bin from stream.
func Read(r io.Reader) (*File, error) {
	d := decoder{r: r}

	if _, err := io.ReadFull(d.r, d.tmp[:4]); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}

	magic := string(d.tmp[:4])
	if magic != FileMagic {
		return nil, fmt.Errorf("%w: got %q", ErrInvalidMagic, magic)
	}

	version, err := d.readU32()
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}

	if version != SupportedVersion {
		return nil, fmt.Errorf("%w: got %d", ErrUnsupportedVersion, version)
	}

	textureCount, err := d.readU32()
	if err != nil {
		return nil, fmt.Errorf("read texture count: %w", err)
	}

	file := &File{
		Magic:    magic,
		Version:  version,
		Textures: make([]TextureEntry, 0, textureCount),
	}

	for i := uint32(0); i < textureCount; i++ {
		entry, entryErr := d.readTextureEntry()
		if entryErr != nil {
			return nil, fmt.Errorf("read texture entry %d: %w", i, entryErr)
		}

		file.Textures = append(file.Textures, entry)
	}

	return file, nil
}

// readTextureEntry decodes one texture entry block.
func (d *decoder) readTextureEntry() (TextureEntry, error) {
	var entry TextureEntry

	count, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read color palette count: %w", err)
	}

	entry.ColorPaletteCount = count

	palettePtr, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read palette ptr: %w", err)
	}

	entry.PalettePtr = palettePtr

	for i := range entry.AverageColorF {
		v, floatErr := d.readF32()
		if floatErr != nil {
			return entry, fmt.Errorf("read average float color[%d]: %w", i, floatErr)
		}

		entry.AverageColorF[i] = v
	}

	if _, err = io.ReadFull(d.r, entry.AverageColor[:]); err != nil {
		return entry, fmt.Errorf("read average color bytes: %w", err)
	}

	if _, err = io.ReadFull(d.r, entry.MaxColor[:]); err != nil {
		return entry, fmt.Errorf("read max color bytes: %w", err)
	}

	clampFlags, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read clamp flags: %w", err)
	}

	entry.ClampFlags = clampFlags

	transparentColor, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read transparent color: %w", err)
	}

	entry.TransparentColor = transparentColor

	if entry.HasMaxCtagg, err = d.readBool8(); err != nil {
		return entry, fmt.Errorf("read has_max_ctagg: %w", err)
	}

	if entry.IsAlpha, err = d.readBool8(); err != nil {
		return entry, fmt.Errorf("read is_alpha: %w", err)
	}

	if entry.IsTransparent, err = d.readBool8(); err != nil {
		return entry, fmt.Errorf("read is_transparent: %w", err)
	}

	if entry.IsAlphaNonOpaque, err = d.readBool8(); err != nil {
		return entry, fmt.Errorf("read is_alpha_non_opaque: %w", err)
	}

	mipCount, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read mip count: %w", err)
	}

	entry.MipMapCount = mipCount

	paxFormat, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read pax format: %w", err)
	}

	entry.PaxFormat = paxFormat

	if entry.LittleEndian, err = d.readBool8(); err != nil {
		return entry, fmt.Errorf("read little_endian: %w", err)
	}

	if entry.IsPAA, err = d.readBool8(); err != nil {
		return entry, fmt.Errorf("read is_paa: %w", err)
	}

	paaFile, err := d.readASCIIZ()
	if err != nil {
		return entry, fmt.Errorf("read paa path: %w", err)
	}

	entry.PAAFile = paaFile

	paxSuffixType, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read pax suffix type: %w", err)
	}

	entry.PaxSuffixType = paxSuffixType

	mipCountCopy, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read mip count copy: %w", err)
	}

	entry.MipMapCountCopy = mipCountCopy
	entry.MipMaps = make([]MipMap, 0, mipCountCopy)

	for i := uint32(0); i < mipCountCopy; i++ {
		m, mipErr := d.readMipMap()
		if mipErr != nil {
			return entry, fmt.Errorf("read mipmap %d: %w", i, mipErr)
		}

		entry.MipMaps = append(entry.MipMaps, m)
	}

	paxFileSize, err := d.readU32()
	if err != nil {
		return entry, fmt.Errorf("read pax file size: %w", err)
	}

	entry.PaxFileSize = paxFileSize

	return entry, nil
}

// readMipMap decodes one mip descriptor.
func (d *decoder) readMipMap() (MipMap, error) {
	var m MipMap

	width, err := d.readU16()
	if err != nil {
		return m, fmt.Errorf("read width: %w", err)
	}

	m.Width = width

	height, err := d.readU16()
	if err != nil {
		return m, fmt.Errorf("read height: %w", err)
	}

	m.Height = height

	alwaysZero, err := d.readU16()
	if err != nil {
		return m, fmt.Errorf("read always zero: %w", err)
	}

	m.AlwaysZero = alwaysZero

	paxFormat, err := d.readU8()
	if err != nil {
		return m, fmt.Errorf("read mip pax format: %w", err)
	}

	m.PaxFormat = paxFormat

	alwaysThree, err := d.readU8()
	if err != nil {
		return m, fmt.Errorf("read always three: %w", err)
	}

	m.AlwaysThree = alwaysThree

	dataOffset, err := d.readU32()
	if err != nil {
		return m, fmt.Errorf("read data offset: %w", err)
	}

	m.DataOffset = dataOffset

	return m, nil
}

// readASCIIZ reads zero-terminated UTF-8/byte string.
func (d *decoder) readASCIIZ() (string, error) {
	buf := make([]byte, 0, 64)
	var tmp [1]byte

	for {
		if _, err := io.ReadFull(d.r, tmp[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return "", ErrInvalidASCIIZ
			}

			return "", err
		}

		b := tmp[0]
		if b == 0 {
			return string(buf), nil
		}

		buf = append(buf, b)
	}
}

func (d *decoder) readU8() (uint8, error) {
	if _, err := io.ReadFull(d.r, d.tmp[:1]); err != nil {
		return 0, err
	}

	return d.tmp[0], nil
}

func (d *decoder) readBool8() (bool, error) {
	v, err := d.readU8()
	if err != nil {
		return false, err
	}

	return v != 0, nil
}

func (d *decoder) readU16() (uint16, error) {
	if _, err := io.ReadFull(d.r, d.tmp[:2]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(d.tmp[:2]), nil
}

func (d *decoder) readU32() (uint32, error) {
	if _, err := io.ReadFull(d.r, d.tmp[:4]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(d.tmp[:4]), nil
}

func (d *decoder) readF32() (float32, error) {
	u, err := d.readU32()
	if err != nil {
		return 0, err
	}

	return math.Float32frombits(u), nil
}
