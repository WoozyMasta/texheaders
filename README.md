# texheaders

`texheaders` is a Go package for reading and writing
DayZ/Arma `texHeaders.bin` files.

It supports two main flows:

* decode/encode `texHeaders.bin` from/to streams or files;
* build `texHeaders.bin` from a list of source texture files (`.paa`).

## Install

```bash
go get github.com/woozymasta/texheaders
```

## Usage

### Decode

```go
f, err := texheaders.ReadFile("testdata/texHeaders.bin")
if err != nil {
    return err
}

fmt.Println(f.Version, len(f.Textures))
```

### Encode

```go
if err := texheaders.WriteFile("out.bin", f); err != nil {
    return err
}
```

### Build From `.paa`

```go
baseDir := "P:/modsource"
b := texheaders.NewBuilder(texheaders.BuildOptions{
    BaseDir:        baseDir,
    LowercasePaths: true,
    BackslashPaths: true,
})

if err := b.AppendMany(
    "P:/modsource/data/test_co.paa",
    "P:/modsource/data/test_nohq.paa",
); err != nil {
    return err
}

if err := b.WriteFile("P:/modsource/texHeaders.bin"); err != nil {
    return err
}
```

### Build With Skip Invalid Inputs

```go
b := texheaders.NewBuilder(texheaders.BuildOptions{SkipInvalid: true})
_ = b.Append("ok_co.paa")
_ = b.Append("not_texture.txt")

f, err := b.Build()
if err != nil {
    return err
}

for _, issue := range b.Issues() {
    fmt.Println(issue.Path, issue.Error)
}

_ = f
```

## Path Normalization

Builder stores `TextureEntry.PAAFile` as normalized relative path:

* relative to `BuildOptions.BaseDir` when possible;
* lowercase by default;
* backslash separators by default.

## Build Parallelism

`BuildOptions.Workers` controls build parallelism:

* `0` or `1`: serial build (default, no worker overhead);
* `>1`: explicit worker count;
* `texheaders.WorkersAuto` (`-1`): auto mode based on `GOMAXPROCS/4`,
  rounded down to nearest power of two and capped by input file count.

## Known Unsupported

* `.pac` source input is currently not supported (`ErrPACUnsupported`).

## Compatibility

Current target is structural compatibility with official output.
Exact byte parity is best-effort and depends on
source metadata/tooling differences.
