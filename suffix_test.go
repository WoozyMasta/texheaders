package texheaders

import "testing"

func TestGuessSuffixTypeFromPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantType uint32
		wantOK   bool
	}{
		{
			name:     "co diffuse",
			path:     "models/terminal/ibm_3178_co.paa",
			wantType: SuffixDiffuseSRGB,
			wantOK:   true,
		},
		{
			name:     "nohq normal",
			path:     "models/terminal/ibm_3178_nohq.paa",
			wantType: SuffixNormalMap,
			wantOK:   true,
		},
		{
			name:     "smdi specular",
			path:     "models/tickets/ticket_smdi.paa",
			wantType: SuffixSpecularAmount,
			wantOK:   true,
		},
		{
			name:     "dtsmdi detail specular",
			path:     "a/b/c/thing_dtsmdi.paa",
			wantType: SuffixDetailSpecularAmount,
			wantOK:   true,
		},
		{
			name:     "unknown fallback",
			path:     "a/b/c/plain_texture.paa",
			wantType: SuffixDiffuseSRGB,
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotType, gotOK := GuessSuffixTypeFromPath(tt.path)
			if gotType != tt.wantType || gotOK != tt.wantOK {
				t.Fatalf("GuessSuffixTypeFromPath(%q) = (%d,%v), want (%d,%v)", tt.path, gotType, gotOK, tt.wantType, tt.wantOK)
			}
		})
	}
}
