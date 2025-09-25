package internal

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/fatih/color"
)

// PrinterOptions controls how the tree is rendered.
type PrinterOptions struct {
	Writer   io.Writer
	MaxDirs  int
	UseColor bool
}

// PrintTree renders the directory tree rooted at dir using the provided label for the root.
func PrintTree(rootLabel string, dir *Directory, opts PrinterOptions) error {
	if dir == nil {
		return fmt.Errorf("nil directory")
	}

	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}

	originalNoColor := color.NoColor
	color.NoColor = !opts.UseColor
	defer func() {
		color.NoColor = originalNoColor
	}()

	palette := newPalette()
	fmt.Fprintln(writer, palette.dir.Sprintf("%s", rootLabel))

	printChildren(writer, dir, "", opts, palette)
	fmt.Fprintf(writer, "%s\n", palette.stats.Sprintf("[%d directories, %d files]", dir.TotalDirs+1, dir.TotalFiles))
	return nil
}

type palette struct {
	dir     *color.Color
	file    *color.Color
	summary *color.Color
	stats   *color.Color
	err     *color.Color
}

func newPalette() palette {
	return palette{
		dir:     color.New(color.FgBlue, color.Bold),
		file:    color.New(color.FgWhite),
		summary: color.New(color.Faint),
		stats:   color.New(color.FgGreen, color.Bold),
		err:     color.New(color.FgRed, color.Bold),
	}
}

type itemKind int

const (
	itemDir itemKind = iota
	itemCollapse
	itemFile
	itemFileSummary
)

type treeItem struct {
	kind          itemKind
	dir           *Directory
	file          FileEntry
	collapseCount int
}

func printChildren(writer io.Writer, dir *Directory, prefix string, opts PrinterOptions, palette palette) {
	items := buildItems(dir, opts)
	for idx, item := range items {
		isLast := idx == len(items)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		switch item.kind {
		case itemDir:
			child := item.dir
			label := child.Name
			if child.Err != nil {
				msg := errorMessage(child, palette)
				fmt.Fprintf(writer, "%s%s%s %s\n", prefix, connector, palette.dir.Sprintf("%s", label), msg)
			} else {
				fmt.Fprintf(writer, "%s%s%s/\n", prefix, connector, palette.dir.Sprintf("%s", label))
				nextPrefix := extendPrefix(prefix, isLast)
				printChildren(writer, child, nextPrefix, opts, palette)
			}
		case itemCollapse:
			fmt.Fprintf(writer, "%s%s%s\n", prefix, connector, palette.summary.Sprintf("... (%d identical dirs)", item.collapseCount))
		case itemFile:
			fmt.Fprintf(writer, "%s%s%s\n", prefix, connector, palette.file.Sprintf("%s", item.file.Name))
		case itemFileSummary:
			fmt.Fprintf(
				writer,
				"%s%s%s\n",
				prefix,
				connector,
				palette.summary.Sprintf(
					"... [%d directories, %d files, showing first %d]",
					dir.ImmediateDirCount,
					dir.ImmediateFileCount,
					dir.ImmediateFileCount-item.collapseCount,
				),
			)
		}
	}
}

func buildItems(dir *Directory, opts PrinterOptions) []treeItem {
	maxDirs := opts.MaxDirs
	if maxDirs <= 0 {
		maxDirs = math.MaxInt
	}

	groups := GroupIdentical(dir.Subdirs)
	items := make([]treeItem, 0, len(dir.Subdirs)+len(dir.Files)+1)

	for _, group := range groups {
		limit := len(group.Members)
		if limit > maxDirs {
			limit = maxDirs
		}
		for i := 0; i < limit; i++ {
			items = append(items, treeItem{kind: itemDir, dir: group.Members[i]})
		}
		if len(group.Members) > limit {
			items = append(items, treeItem{kind: itemCollapse, collapseCount: len(group.Members) - limit})
		}
	}

	for _, file := range dir.Files {
		items = append(items, treeItem{kind: itemFile, file: file})
	}

	if dir.HiddenFiles > 0 {
		items = append(items, treeItem{kind: itemFileSummary, collapseCount: dir.HiddenFiles})
	}

	return items
}

func extendPrefix(prefix string, isLast bool) string {
	if isLast {
		return prefix + "    "
	}
	return prefix + "│   "
}

func errorMessage(dir *Directory, palette palette) string {
	if dir.IsPermissionError() {
		return palette.summary.Sprintf("[Permission denied]")
	}
	trimmed := strings.TrimSpace(dir.Err.Error())
	if trimmed == "" {
		trimmed = "error"
	}
	return palette.err.Sprintf("[%s]", trimmed)
}
