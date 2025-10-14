package test

import (
	"testing"
	"time"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to cover File.Size and File.LastModified when file stats are unavailable.
func Test_File_SizeAndLastModified_NoStats(t *testing.T) {
	file := &internal.File{
		Name: "ghost",
		Path: "/tmp/ghost",
	}

	if size := file.Size(); size != 0 {
		t.Errorf(internal.TextColor.Red("Expected size 0 for file without stats, got %d"), size)
	}

	if modified := file.LastModified(); !modified.Equal(time.Time{}) {
		t.Errorf(internal.TextColor.Red("Expected zero time for file without stats, got %v"), modified)
	}
}

// Test case to cover error branch when file contents cannot be read.
func Test_File_Contents_Error(t *testing.T) {
	file := &internal.File{
		Path: "missing-file.txt",
	}

	if _, err := file.Contents(); err == nil {
		t.Fatal(internal.TextColor.Red("Expected error when reading missing file"))
	}
}

// Test case to cover GetFile path that resolves to a directory.
func Test_FileSystem_GetFile_DirectoryPath(t *testing.T) {
	fs := new(internal.FileSystem)
	root := t.TempDir()

	if _, err := fs.GetFile(root); err == nil {
		t.Fatal(internal.TextColor.Red("Expected error when requesting directory path"))
	}
}
