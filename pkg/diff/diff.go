// Package diff handles parsing and working with git diff data.
package diff

import (
	"bufio"
	"fmt"
	"strings"
)

// ChangeType represents the type of a file change.
type ChangeType int

const (
	Addition ChangeType = iota
	Modification
	Deletion
)

func (ct ChangeType) String() string {
	switch ct {
	case Addition:
		return "added"
	case Modification:
		return "modified"
	case Deletion:
		return "deleted"
	default:
		return "unknown"
	}
}

// Line represents a single line in a diff.
type Line struct {
	LineNumber int       // Line number in the new file (0 for deleted file lines)
	OldNumber  int       // Line number in the old file (0 for added file lines)
	Content    string    // The line content (with +/- prefix removed)
	Prefix     string    // The diff prefix: " ", "+", "-", "@"
	HunkHeader string    // The hunk header if this is a hunk line
	IsHunk     bool      // Whether this line is a hunk header (@@ ...)
	IsHeader   bool      // Whether this line is a diff header (diff --git ...)
	IsNewFile  bool      // Whether this line is --- a/file
	IsOldFile  bool      // Whether this line is +++ b/file
}

// FileDiff represents the diff for a single file.
type FileDiff struct {
	OldPath    string       // Original file path (before rename)
	NewPath    string       // New file path (after rename)
	OldMode    string       // Old file mode (e.g., 100644)
	NewMode    string       // New file mode
	ChangeType ChangeType   // Type of change
	Added      int          // Number of lines added
	Deleted    int          // Number of lines deleted
	Lines      []Line       // All lines in the diff
	Index      int          // Line index in the full diff
	IsBinary   bool         // Whether the file is binary
	IsRename   bool         // Whether this is a rename
	RenameTo   string       // New path for renames
	RenameFrom string       // Old path for renames
}

// Diff represents the full repository diff.
type Diff struct {
	Files    []*FileDiff
	TotalAdd int
	TotalDel int
	Raw      string
}

// ParseDiff parses a unified diff string into a Diff structure.
func ParseDiff(raw string) (*Diff, error) {
	d := &Diff{Raw: raw}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var currentFile *FileDiff
	var currentLines []Line

	flushFile := func() {
		if currentFile != nil {
			currentFile.Lines = currentLines
			d.Files = append(d.Files, currentFile)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		l := parseLine(line, len(currentLines))

		if l.IsHeader {
			// New file diff block starts
			flushFile()
			currentFile = &FileDiff{Index: len(d.Files)}
			// Parse path from diff --git a/path b/path
			parts := strings.SplitN(l.Content, " ", 4)
			if len(parts) >= 4 {
				currentFile.OldPath = parseGitPath(parts[2])
				currentFile.NewPath = parseGitPath(parts[3])
			}
			currentLines = nil
		} else if l.IsHunk || l.IsNewFile || l.IsOldFile {
			// Hunk header, --- file, +++ file — all part of current file
			if currentFile == nil {
				// Orphan line before any file header
				currentFile = &FileDiff{Index: len(d.Files)}
				currentLines = nil
			}
			currentLines = append(currentLines, l)
		} else if l.Prefix == "+" || l.Prefix == "-" || l.Prefix == " " {
			// Content line
			if currentFile != nil {
				currentLines = append(currentLines, l)
			}
		}
		// Blank lines and other lines are ignored
	}
	flushFile()

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("scan diff: %w", scanErr)
	}

	return d, nil
}

// parseGitPath extracts the path from "a/path" or "b/path" format.
func parseGitPath(s string) string {
	if strings.HasPrefix(s, "a/") || strings.HasPrefix(s, "b/") {
		return s[2:]
	}
	return s
}

// parseLine converts a raw diff line into a Line struct.
func parseLine(raw string, lineCount int) Line {
	l := Line{LineNumber: lineCount}

	if strings.HasPrefix(raw, "@@") {
		l.Prefix = "@"
		l.Content = raw
		l.IsHunk = true
		l.HunkHeader = raw
		return l
	}

	if strings.HasPrefix(raw, "diff --git") {
		l.Prefix = "diff"
		l.Content = raw
		l.IsHeader = true
		return l
	}

	if strings.HasPrefix(raw, "--- ") {
		l.Content = strings.TrimPrefix(raw, "--- ")
		l.IsNewFile = true
		return l
	}

	if strings.HasPrefix(raw, "+++ ") {
		l.Content = strings.TrimPrefix(raw, "+++ ")
		l.IsOldFile = true
		return l
	}

	if raw == "" {
		l.Content = ""
		return l
	}

	if strings.HasPrefix(raw, "+") {
		l.Prefix = "+"
		l.Content = raw[1:]
		l.LineNumber = lineCount
		return l
	}

	if strings.HasPrefix(raw, "-") {
		l.Prefix = "-"
		l.Content = raw[1:]
		return l
	}

	l.Prefix = " "
	l.Content = raw
	return l
}

// Analyze computes summary statistics for the diff.
func (d *Diff) Analyze() {
	for _, fd := range d.Files {
		added, deleted := analyzeFile(fd)
		fd.Added = added
		fd.Deleted = deleted

		// Determine change type based on paths and content
		hasNewPath := fd.NewPath != ""
		hasOldPath := fd.OldPath != ""
		isNewFile := hasNewPath && !hasOldPath
		isDeletedFile := hasOldPath && !hasNewPath

		// Check for /dev/null indicators
		for _, l := range fd.Lines {
			if l.Content == "/dev/null" {
				// --- /dev/null means source is empty = new file being added
				if l.IsNewFile {
					isNewFile = true
				}
				// +++ /dev/null means dest is empty = file being deleted
				if l.IsOldFile {
					isDeletedFile = true
				}
			}
		}

		if isNewFile {
			fd.ChangeType = Addition
		} else if isDeletedFile {
			fd.ChangeType = Deletion
		} else if fd.Added > 0 || fd.Deleted > 0 {
			fd.ChangeType = Modification
		} else {
			fd.ChangeType = Modification
		}

		// Detect renames
		if fd.OldPath != "" && fd.NewPath != "" && fd.OldPath != fd.NewPath {
			if added == 0 {
				fd.ChangeType = Deletion
				fd.IsRename = true
			} else if fd.Deleted == 0 {
				fd.ChangeType = Addition
				fd.IsRename = true
			}
		}

		d.TotalAdd += added
		d.TotalDel += deleted
	}
}

// analyzeFile counts added/deleted lines for a single file diff.
func analyzeFile(fd *FileDiff) (added, deleted int) {
	for _, l := range fd.Lines {
		if l.Prefix == "+" && !l.IsHunk && !l.IsHeader && !l.IsNewFile && !l.IsOldFile {
			if l.Content != "/dev/null" {
				added++
			}
		} else if l.Prefix == "-" && !l.IsHunk && !l.IsHeader && !l.IsNewFile && !l.IsOldFile {
			if l.Content != "/dev/null" {
				deleted++
			}
		}
	}
	return added, deleted
}

// GetAddedLines returns all added lines across the diff.
func (d *Diff) GetAddedLines() []Line {
	var lines []Line
	for _, fd := range d.Files {
		for _, l := range fd.Lines {
			if l.Prefix == "+" && !l.IsHunk && !l.IsHeader && !l.IsNewFile && !l.IsOldFile {
				lines = append(lines, l)
			}
		}
	}
	return lines
}

// GetDeletedLines returns all deleted lines across the diff.
func (d *Diff) GetDeletedLines() []Line {
	var lines []Line
	for _, fd := range d.Files {
		for _, l := range fd.Lines {
			if l.Prefix == "-" && !l.IsHunk && !l.IsHeader && !l.IsNewFile && !l.IsOldFile {
				lines = append(lines, l)
			}
		}
	}
	return lines
}

// GetModifiedPaths returns all files that were modified.
func (d *Diff) GetModifiedPaths() []string {
	var paths []string
	for _, fd := range d.Files {
		if fd.ChangeType == Modification || fd.ChangeType == Addition {
			if fd.NewPath != "" {
				paths = append(paths, fd.NewPath)
			}
		}
		if fd.ChangeType == Deletion && fd.OldPath != "" {
			paths = append(paths, fd.OldPath)
		}
	}
	return paths
}

// GetChangedLines returns all changed lines (added + deleted).
func (d *Diff) GetChangedLines() int {
	return d.TotalAdd + d.TotalDel
}
