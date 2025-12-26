package entry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPathEntry_ValidAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	entry, err := NewPathEntry(tempFile, tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.AbsPath != tempFile {
		t.Errorf("expected AbsPath %s, go %s", tempFile, entry.AbsPath)
	}
	if entry.FType != FileTypeFile {
		t.Errorf("expected %s, got %s", FileTypeFile, entry.FType)
	}
}

func TestNewPathEntry_RelativePathError(t *testing.T) {
	_, err := NewPathEntry("relative/path", "/absolute/base")
	if err != ErrNotAbsolute {
		t.Errorf("expected %v, got %v", ErrNotAbsolute, err)
	}

	_, err = NewPathEntry("/absolute/path", "relative/base")
	if err != ErrNotAbsolute {
		t.Errorf("expected %v, got %v", ErrNotAbsolute, err)
	}
}

func TestNewPathEntry_Directory(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	entry, err := NewPathEntry(subDir, tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.FType != FileTypeDir {
		t.Errorf("expected %v, got %v", FileTypeDir, entry.FType)
	}
}

func TestNewPathEntry_Distance(t *testing.T) {
	tempDir := t.TempDir()
	dirA := filepath.Join(tempDir, "a")
	dirB := filepath.Join(tempDir, "a", "b")
	dirC := filepath.Join(tempDir, "a", "b", "c")
	sibling := filepath.Join(tempDir, "sibling")
	fileInA := filepath.Join(tempDir, "a", "file.txt")
	fileInRoot := filepath.Join(tempDir, "root.txt")

	for _, dir := range []string{dirA, dirB, dirC, sibling} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range []string{fileInA, fileInRoot} {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		entry    string
		base     string
		expected int
	}{
		{"same directory", tempDir, tempDir, 0},
		{"one level down", dirA, tempDir, 1},
		{"two levels down", dirB, tempDir, 2},
		{"three levels down", dirC, tempDir, 3},
		{"sibling directory", sibling, dirA, 2},
		{"parent directory", tempDir, dirA, 1},
		{"file in same dir", fileInRoot, tempDir, 0},
		{"file in one level down", fileInA, tempDir, 1},
		{"file from sibling dir", fileInA, sibling, 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry, err := NewPathEntry(test.entry, test.base)
			if err != nil {
				t.Fatalf("unexpcted error: %v", err)
			}
			if entry.Distance != test.expected {
				t.Errorf("expected distance %d, got %d", test.expected, entry.Distance)
			}
		})
	}
}

func TestDistanceBetween(t *testing.T) {
	tests := []struct {
		name     string
		pathA    string
		pathB    string
		expected int
	}{
		{"same path", "/home/user", "/home/user", 0},
		{"one level down", "/home/user", "/home/user/a", 1},
		{"two levels down", "/home/user", "/home/user/a/b", 2},
		{"one level up", "/home/user/a", "/home/user", 1},
		{"sibling", "/home/user/a", "/home/user/b", 2},
		{"cousin", "/home/user/a/x", "/home/user/b/y", 4},
		{"deep to shallow", "/home/user/a/b/c", "/home/user", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &PathEntry{AbsPath: test.pathA}
			b := &PathEntry{AbsPath: test.pathB}

			dist, err := DistanceBetween(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dist != test.expected {
				t.Errorf("expected distance %d, got %d", test.expected, dist)
			}
		})
	}
}

func TestGetFileType_NonexistancePath(t *testing.T) {
	_, err := getFileType("/nonexistent/path/12345")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}
