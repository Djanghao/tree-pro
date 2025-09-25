package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Options controls how the filesystem is traversed.
type Options struct {
	MaxFiles int
	MaxLevel int
}

// Directory represents a directory and its contents used for rendering.
type Directory struct {
	Name               string
	Path               string
	Level              int
	Subdirs            []*Directory
	Files              []FileEntry
	HiddenFiles        int
	ImmediateDirCount  int
	ImmediateFileCount int
	TotalDirs          int
	TotalFiles         int
	Signature          string
	Err                error
}

// FileEntry captures the metadata required to render a file node.
type FileEntry struct {
	Name string
}

// Walk builds a Directory tree starting at the provided path according to the
// supplied options. Returns an error if the root path is inaccessible.
func Walk(path string, opts Options) (*Directory, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", path)
	}

	clean := filepath.Clean(path)
	root := walkDir(clean, info.Name(), 0, opts)
	if root.Err != nil {
		return nil, root.Err
	}
	return root, nil
}

func walkDir(path, name string, level int, opts Options) *Directory {
	node := &Directory{
		Name:  name,
		Path:  path,
		Level: level,
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		node.Err = err
		node.Signature = signatureForError(path, err)
		return node
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	maxFiles := opts.MaxFiles
	if maxFiles <= 0 {
		maxFiles = math.MaxInt
	}

	fileExtCounts := map[string]int{}
	hiddenFiles := 0
	files := make([]FileEntry, 0, len(entries))
	subdirs := make([]*Directory, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			joined := filepath.Join(path, entry.Name())
			if opts.MaxLevel != 0 && level >= opts.MaxLevel {
				subdir := &Directory{
					Name:      entry.Name(),
					Path:      joined,
					Level:     level + 1,
					Signature: signatureForLeaf(joined),
				}
				subdirs = append(subdirs, subdir)
				continue
			}

			child := walkDir(joined, entry.Name(), level+1, opts)
			subdirs = append(subdirs, child)
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == "" {
			ext = "<noext>"
		}
		fileExtCounts[ext]++

		if len(files) < maxFiles {
			files = append(files, FileEntry{Name: filename})
		} else {
			hiddenFiles++
		}
	}

	node.Subdirs = subdirs
	node.Files = files
	node.HiddenFiles = hiddenFiles
	node.ImmediateDirCount = len(subdirs)
	node.ImmediateFileCount = len(files) + hiddenFiles

	totalDirs := len(subdirs)
	totalFiles := node.ImmediateFileCount
	for _, child := range subdirs {
		totalDirs += child.TotalDirs
		totalFiles += child.TotalFiles
	}
	node.TotalDirs = totalDirs
	node.TotalFiles = totalFiles

	node.Signature = signatureForDirectory(fileExtCounts, subdirs)

	return node
}

func signatureForDirectory(fileExtCounts map[string]int, subdirs []*Directory) string {
	hasher := fnv.New64a()
	hasher.Write([]byte("files:"))

	exts := make([]string, 0, len(fileExtCounts))
	for ext := range fileExtCounts {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	for _, ext := range exts {
		hasher.Write([]byte(ext))
		hasher.Write([]byte{0})
		hasher.Write(intToBytes(fileExtCounts[ext]))
	}

	hasher.Write([]byte("dirs:"))
	childSigs := make([]string, 0, len(subdirs))
	for _, child := range subdirs {
		childSigs = append(childSigs, child.Signature)
	}
	sort.Strings(childSigs)
	for _, sig := range childSigs {
		hasher.Write([]byte(sig))
		hasher.Write([]byte{0})
	}

	return fmt.Sprintf("d:%x", hasher.Sum64())
}

func signatureForLeaf(path string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte("leaf:"))
	hasher.Write([]byte(path))
	return fmt.Sprintf("l:%x", hasher.Sum64())
}

func signatureForError(path string, err error) string {
	return fmt.Sprintf("err:%s:%T", path, err)
}

func intToBytes(v int) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b
}

// IsPermissionError reports whether the directory encountered a permission error.
func (d *Directory) IsPermissionError() bool {
	if d == nil || d.Err == nil {
		return false
	}
	return errors.Is(d.Err, fs.ErrPermission)
}
