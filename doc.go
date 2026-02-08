/*
Package texheaders reads and writes DayZ/Arma texHeaders.bin files.

The format stores texture metadata index entries (path, color tags, mip
descriptors, pax format, and suffix type). The package provides stream/file
APIs for decode/encode and a builder API for creating texHeaders models from
source .paa files.

Basic read:

	f, err := texheaders.ReadFile("texHeaders.bin")
	if err != nil {
		return err
	}

Basic write:

	err := texheaders.WriteFile("out.bin", f)
	if err != nil {
		return err
	}

Build from textures:

	b := texheaders.NewBuilder(texheaders.BuildOptions{BaseDir: "P:/mod"})
	_ = b.AppendMany(
		"P:/mod/data/test_co.paa",
		"P:/mod/data/test_nohq.paa",
	)
	err = b.WriteFile("P:/mod/texHeaders.bin")
	if err != nil {
		return err
	}
*/
package texheaders
