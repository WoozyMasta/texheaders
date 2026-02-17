// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/texheaders

package texheaders

import "strings"

// Known pax suffix kinds from available format docs.
const (
	SuffixDiffuseSRGB           uint32 = 0
	SuffixDiffuseLinear         uint32 = 1
	SuffixDetailLinear          uint32 = 2
	SuffixNormalMap             uint32 = 3
	SuffixIrradianceMap         uint32 = 4
	SuffixRandom05To1           uint32 = 5
	SuffixTreeCrownCalc         uint32 = 6
	SuffixMacroObjectSRGB       uint32 = 7
	SuffixAmbientShadow         uint32 = 8
	SuffixSpecularAmount        uint32 = 9
	SuffixDitherTexture         uint32 = 10
	SuffixDetailSpecularAmount  uint32 = 11
	SuffixMultiShaderMask       uint32 = 12
	SuffixThermalImageTextureCA uint32 = 13
)

// suffixGuessRule describes one suffix inference rule.
type suffixGuessRule struct {
	token string
	value uint32
}

// Ordered longest-first where overlap exists.
var suffixGuessRules = []suffixGuessRule{
	{token: "_nohq_alpha", value: SuffixDiffuseSRGB},
	{token: "_dtsmdi", value: SuffixDetailSpecularAmount},
	{token: "_ti_ca", value: SuffixThermalImageTextureCA},
	{token: "_smdi", value: SuffixSpecularAmount},
	{token: "_detail", value: SuffixDetailLinear},
	{token: "_normalmap", value: SuffixNormalMap},
	{token: "_nohq", value: SuffixNormalMap},
	{token: "_novhq", value: SuffixNormalMap},
	{token: "_nofhq", value: SuffixNormalMap},
	{token: "_nofex", value: SuffixNormalMap},
	{token: "_noex", value: SuffixNormalMap},
	{token: "_nsex", value: SuffixNormalMap},
	{token: "_nshq", value: SuffixNormalMap},
	{token: "_nopx", value: SuffixNormalMap},
	{token: "_non", value: SuffixNormalMap},
	{token: "_nof", value: SuffixNormalMap},
	{token: "_nse", value: SuffixNormalMap},
	{token: "_ns", value: SuffixNormalMap},
	{token: "_no", value: SuffixNormalMap},
	{token: "_mask", value: SuffixMultiShaderMask},
	{token: "_sky", value: SuffixDiffuseLinear},
	{token: "_lco", value: SuffixDiffuseLinear},
	{token: "_dxt5", value: SuffixDiffuseLinear},
	{token: "_mco", value: SuffixDetailLinear},
	{token: "_cdt", value: SuffixDetailLinear},
	{token: "_dt", value: SuffixDetailLinear},
	{token: "_mc", value: SuffixMacroObjectSRGB},
	{token: "_as", value: SuffixAmbientShadow},
	{token: "_sm", value: SuffixSpecularAmount},
	{token: "_ca", value: SuffixDiffuseSRGB},
	{token: "_co", value: SuffixDiffuseSRGB},
}

// GuessSuffixTypeFromPath tries to infer pax suffix type from texture file path.
//
// This is heuristic mapping based on known DayZ/Arma naming conventions.
// Unknown patterns fall back to diffuse_srgb (0) and return ok=false.
func GuessSuffixTypeFromPath(path string) (value uint32, ok bool) {
	s := strings.ToLower(path)
	dot := strings.LastIndexByte(s, '.')
	if dot > 0 {
		s = s[:dot]
	}

	for _, rule := range suffixGuessRules {
		if containsTokenBoundary(s, rule.token) {
			return rule.value, true
		}
	}

	return SuffixDiffuseSRGB, false
}

// containsTokenBoundary checks token match with a separator/end right after token.
func containsTokenBoundary(s, token string) bool {
	from := 0
	for {
		idx := strings.Index(s[from:], token)
		if idx < 0 {
			return false
		}

		pos := from + idx
		next := pos + len(token)
		if next >= len(s) {
			return true
		}

		ch := s[next]
		if ch == '_' || ch == '-' || ch == '.' {
			return true
		}

		from = pos + 1
	}
}
