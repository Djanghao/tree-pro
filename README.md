# tree-pro

`tree-pro` prints a concise directory tree, folding identical subdirectories and limiting the number of files shown per level.

## Install

Option A — via `go install` (recommended):

```bash
go install github.com/Djanghao/tree-pro@v0.2.0
# or use the latest tag
go install github.com/Djanghao/tree-pro@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH` (or set `GOBIN`):

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Option B — build from source after cloning:

```bash
git clone https://github.com/Djanghao/tree-pro
cd tree-pro
# install to your Go bin (no sudo required)
go install .
# or build a local binary
go build -o tree-pro .
```

## Usage

```bash
tree-pro [path] [flags]
```
- `-f, --files` limit files per directory (default 5)
- `-d, --dirs` expand identical directories (default 1)
- `-L, --level` max depth (0 = unlimited)

Example:
```bash
tree-pro -f 2 -d 1 ~/datasets/train
```

## Sample

Command:

```bash
tree-pro -f 2 examples/sample
```

Output (colors stripped for README):

```
examples/sample/
├── pkg/
│   ├── serviceA/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── ... (1 identical dirs)
│   │   └── config.yaml
│   └── ... (2 identical dirs)
└── scripts/
    ├── build.sh
    ├── deploy.sh
    └── ... [0 directories, 3 files, showing first 2]
[12 directories, 12 files]
```
