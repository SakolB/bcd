package entry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type FileType string

const (
	FileTypeFile    FileType = "file"
	FileTypeDir     FileType = "dir"
	FileTypeSymlink FileType = "symlink"
)

type PathEntry struct {
	AbsPath  string
	Distance int
	FType    FileType
}

var ErrNotAbsolute = errors.New("path is not absolute")

func NewPathEntry(entryAbsPath string, baseDirAbsPath string) (*PathEntry, error) {
	if !filepath.IsAbs(entryAbsPath) || !filepath.IsAbs(baseDirAbsPath) {
		return nil, ErrNotAbsolute
	}

	filetype, err := getFileType(entryAbsPath)
	if err != nil {
		return nil, err
	}

	pathForDistance := entryAbsPath
	if filetype == FileTypeFile {
		pathForDistance = filepath.Dir(pathForDistance)
	}

	relPath, err := filepath.Rel(baseDirAbsPath, pathForDistance)
	if err != nil {
		return nil, err
	}

	var distance int
	if relPath == "." {
		distance = 0
	} else {
		distance = strings.Count(relPath, string(filepath.Separator)) + 1
	}

	return &PathEntry{
		AbsPath:  entryAbsPath,
		Distance: distance,
		FType:    filetype,
	}, nil
}

func DistanceBetween(a, b *PathEntry) (int, error) {
	pathForDistanceA := a.AbsPath
	pathForDistanceB := b.AbsPath
	if a.FType == FileTypeFile {
		pathForDistanceA = filepath.Dir(pathForDistanceA)
	}
	if b.FType == FileTypeFile {
		pathForDistanceB = filepath.Dir(pathForDistanceB)
	}
	relPath, err := filepath.Rel(pathForDistanceA, pathForDistanceB)
	if err != nil {
		return 0, err
	}
	var distance int
	if relPath == "." {
		distance = 0
	} else {
		distance = strings.Count(relPath, string(filepath.Separator)) + 1
	}
	return distance, nil
}

func getFileType(entryAbsPath string) (FileType, error) {
	fileinfo, err := os.Lstat(entryAbsPath)
	if err != nil {
		return FileTypeFile, err
	}
	if fileinfo.Mode()&os.ModeSymlink != 0 {
		return FileTypeSymlink, nil
	}
	if fileinfo.IsDir() {
		return FileTypeDir, nil
	}
	return FileTypeFile, nil
}
