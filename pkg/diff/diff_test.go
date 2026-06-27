// Package diff_test tests diff parsing and analysis.
package diff_test

import (
	"testing"

	"diffscope/pkg/diff"
)

const sampleDiff = `diff --git a/pkg/main.go b/pkg/main.go
index abc123..def456 100644
--- a/pkg/main.go
+++ b/pkg/main.go
@@ -10,3 +10,5 @@ package main
 	"fmt"
 	"os"
 )
+
+func NewApp() *App { return &App{} }
diff --git a/pkg/handler.go b/pkg/handler.go
index 789abc..012def 100644
--- a/pkg/handler.go
+++ b/pkg/handler.go
@@ -5,4 +5,3 @@ package pkg
 func Handle() {
 	fmt.Println("hello")
-	fmt.Println("old")
+	fmt.Println("new")
 }
diff --git a/pkg/newfile.go b/pkg/newfile.go
new file mode 100644
index 000000..123456
--- /dev/null
+++ b/pkg/newfile.go
@@ -0,0 +1,3 @@
+package pkg
+
+func NewThing() {}
`

func TestParseDiff(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	if len(d.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(d.Files))
	}
}

func TestDiffAnalyze(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()

	if d.TotalAdd == 0 {
		t.Error("expected some additions")
	}
	if d.TotalDel == 0 {
		t.Error("expected some deletions")
	}
}

func TestGetAddedLines(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()
	added := d.GetAddedLines()
	if len(added) == 0 {
		t.Error("expected added lines")
	}
	for _, l := range added {
		if l.Prefix != "+" {
			t.Errorf("expected + prefix, got %q", l.Prefix)
		}
	}
}

func TestGetDeletedLines(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()
	deleted := d.GetDeletedLines()
	if len(deleted) == 0 {
		t.Error("expected deleted lines")
	}
	for _, l := range deleted {
		if l.Prefix != "-" {
			t.Errorf("expected - prefix, got %q", l.Prefix)
		}
	}
}

func TestGetModifiedPaths(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()
	paths := d.GetModifiedPaths()
	if len(paths) == 0 {
		t.Error("expected modified paths")
	}
}

func TestParseEmptyDiff(t *testing.T) {
	d, err := diff.ParseDiff("")
	if err != nil {
		t.Fatalf("ParseDiff on empty failed: %v", err)
	}
	if len(d.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(d.Files))
	}
}

func TestParseSingleFileDiff(t *testing.T) {
	single := `diff --git a/file.go b/file.go
index 123..456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
+
+func Hello() string { return "hi" }
 `
	d, err := diff.ParseDiff(single)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	if len(d.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(d.Files))
	}
	d.Analyze()
	if d.TotalAdd != 2 {
		t.Errorf("expected 2 additions, got %d", d.TotalAdd)
	}
}

func TestGetChangedLines(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()
	changed := d.GetChangedLines()
	if changed <= 0 {
		t.Errorf("expected positive changed lines, got %d", changed)
	}
}

func TestParseNewFile(t *testing.T) {
	newFile := `diff --git a/pkg/new.go b/pkg/new.go
new file mode 100644
index 000000..abcdef
--- /dev/null
+++ b/pkg/new.go
@@ -0,0 +1,2 @@
+package pkg
+func New() {}
`
	d, err := diff.ParseDiff(newFile)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}
	d.Analyze()
	if d.TotalAdd != 2 {
		t.Errorf("expected 2 additions for new file, got %d", d.TotalAdd)
	}
}

func TestParseDeletedFile(t *testing.T) {
	deleted := `diff --git a/pkg/old.go b/pkg/old.go
deleted file mode 100644
index abcdef..000000
--- a/pkg/old.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package pkg
-func Old() {}
`
	d, err := diff.ParseDiff(deleted)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}
	d.Analyze()
	if d.TotalDel != 2 {
		t.Errorf("expected 2 deletions, got %d", d.TotalDel)
	}
}

func TestFileChangeTypes(t *testing.T) {
	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()

	// File 0: main.go — modified
	if d.Files[0].ChangeType != diff.Modification {
		t.Errorf("Files[0] changeType = %v, want modification", d.Files[0].ChangeType)
	}
	// File 1: handler.go — modified
	if d.Files[1].ChangeType != diff.Modification {
		t.Errorf("Files[1] changeType = %v, want modification", d.Files[1].ChangeType)
	}
	// File 2: newfile.go — added
	if d.Files[2].ChangeType != diff.Addition {
		t.Errorf("Files[2] changeType = %v, want addition", d.Files[2].ChangeType)
	}
}
