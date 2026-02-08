package texheaders

// FileMagic is the required 4-byte file signature.
const FileMagic = "0DHT"

// SupportedVersion is the only currently supported file version.
const SupportedVersion uint32 = 1

// File represents texHeaders.bin content.
type File struct {
	// Magic is expected to be "0DHT".
	Magic string `json:"magic,omitempty" yaml:"magic,omitempty"`
	// Textures holds all texture entries in file order.
	Textures []TextureEntry `json:"textures,omitempty" yaml:"textures,omitempty"`
	// Version is expected to be 1.
	Version uint32 `json:"version,omitempty" yaml:"version,omitempty"`
}

// TextureEntry describes one texture metadata entry.
type TextureEntry struct {
	// PAAFile is a path relative to texHeaders.bin location.
	PAAFile string `json:"paa_file,omitempty" yaml:"paa_file,omitempty"`
	// MipMaps contains mip descriptors.
	MipMaps []MipMap `json:"mipmaps,omitempty" yaml:"mipmaps,omitempty"`

	// ColorPaletteCount is usually 1.
	ColorPaletteCount uint32 `json:"color_palette_count,omitempty" yaml:"color_palette_count,omitempty"`
	// PalettePtr is usually 0.
	PalettePtr uint32 `json:"palette_ptr,omitempty" yaml:"palette_ptr,omitempty"`

	// AverageColorF stores average color as float32 tuple.
	AverageColorF [4]float32 `json:"average_color_f,omitempty" yaml:"average_color_f,omitempty"`
	// AverageColor stores average color as byte tuple.
	AverageColor [4]byte `json:"average_color,omitempty" yaml:"average_color,omitempty"`
	// MaxColor stores max color as byte tuple.
	MaxColor [4]byte `json:"max_color,omitempty" yaml:"max_color,omitempty"`

	// ClampFlags is usually 0.
	ClampFlags uint32 `json:"clamp_flags,omitempty" yaml:"clamp_flags,omitempty"`
	// TransparentColor is usually 0xFFFFFFFF.
	TransparentColor uint32 `json:"transparent_color,omitempty" yaml:"transparent_color,omitempty"`

	// HasMaxCtagg means MaxColor was set by source paa.
	HasMaxCtagg bool `json:"has_max_ctagg,omitempty" yaml:"has_max_ctagg,omitempty"`
	// IsAlpha means FLAGTAG = 1 basic transparency.
	IsAlpha bool `json:"is_alpha,omitempty" yaml:"is_alpha,omitempty"`
	// IsTransparent means FLAGTAG = 2 non-interpolated alpha.
	IsTransparent bool `json:"is_transparent,omitempty" yaml:"is_transparent,omitempty"`
	// IsAlphaNonOpaque means IsAlpha and average alpha < 0x80.
	IsAlphaNonOpaque bool `json:"is_alpha_non_opaque,omitempty" yaml:"is_alpha_non_opaque,omitempty"`

	// MipMapCount is usually equal to MipMapCountCopy.
	MipMapCount uint32 `json:"mipmap_count,omitempty" yaml:"mipmap_count,omitempty"`
	// PaxFormat describes texture storage format.
	PaxFormat uint32 `json:"pax_format,omitempty" yaml:"pax_format,omitempty"`
	// LittleEndian is expected to be true.
	LittleEndian bool `json:"little_endian,omitempty" yaml:"little_endian,omitempty"`
	// IsPAA tells whether source file is .paa.
	IsPAA bool `json:"is_paa,omitempty" yaml:"is_paa,omitempty"`
	// PaxSuffixType is texture suffix class identifier.
	PaxSuffixType uint32 `json:"pax_suffix_type,omitempty" yaml:"pax_suffix_type,omitempty"`

	// MipMapCountCopy is usually equal to MipMapCount.
	MipMapCountCopy uint32 `json:"mipmap_count_copy,omitempty" yaml:"mipmap_count_copy,omitempty"`
	// PaxFileSize stores source pax file size in bytes.
	PaxFileSize uint32 `json:"pax_file_size,omitempty" yaml:"pax_file_size,omitempty"`
}

// MipMap describes one mipmap descriptor.
type MipMap struct {
	Width  uint16 `json:"width,omitempty" yaml:"width,omitempty"`
	Height uint16 `json:"height,omitempty" yaml:"height,omitempty"`
	// AlwaysZero is expected to be 0 in known files.
	AlwaysZero uint16 `json:"always_zero,omitempty" yaml:"always_zero,omitempty"`
	// PaxFormat usually matches entry PaxFormat.
	PaxFormat uint8 `json:"pax_format,omitempty" yaml:"pax_format,omitempty"`
	// AlwaysThree is expected to be 3 in known files.
	AlwaysThree uint8 `json:"always_three,omitempty" yaml:"always_three,omitempty"`
	// DataOffset points to mip payload inside source pax.
	DataOffset uint32 `json:"data_offset,omitempty" yaml:"data_offset,omitempty"`
}
