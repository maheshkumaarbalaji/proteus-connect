package test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the working of the IsAbsolute() function of FileSystem.
func Test_FileSystem_IsAbsolute(t *testing.T) {
	root := t.TempDir()
	AbsFolderExists := filepath.Join(root, "abs-exists")
	AbsFolderNoExists := filepath.Join(root, "abs-noexists")
	err := CreateDirectories(t, root, []string{"abs-exists"})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating test folders: %s"), err.Error())
		return
	}

	fs := new(internal.FileSystem)
	testCases := []struct {
		Name   string
		IpPath string
		ExpOp  bool
	}{
		{"A valid absolute path that exists in the local file system", AbsFolderExists, true},
		{"A valid absolute path that does not exist in the local file system", AbsFolderNoExists, true},
		{"A relative path that exists in the file system", "./rel-exists", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			isAbsolute := fs.IsAbsolute(testCase.IpPath)
			if isAbsolute == testCase.ExpOp {
				tt.Logf("The result returned by isAbsolute() for path [%s] matches the expected output", testCase.IpPath)
			} else {
				tt.Errorf(internal.TextColor.Red("The result returned by isAbsolute() for path [%s] does not match the expected output"), testCase.IpPath)
			}
		})
	}
}

// Test case to validate the working of the IsDirectory() method of FileSystem.
func Test_FileSystem_IsDirectory(t *testing.T) {
	root := t.TempDir()
	ExistsFolder := filepath.Join(root, "folderone")
	NotExistsFolder := filepath.Join(root, "foldertwo")
	err := CreateDirectories(t, root, []string{"folderone"})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating test folders: %s"), err.Error())
		return
	}

	ExistsTextFile := filepath.Join(root, "exists.txt")
	err = CreateFiles(t, root, map[string][]byte{
		"exists.txt": []byte("Hello, this is a sample text file"),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating test file: %s"), err.Error())
		return
	}

	NotExistsTextFile := filepath.Join(root, "not-exists.txt")
	fs := new(internal.FileSystem)
	testCases := []struct {
		Name   string
		IpPath string
		ExpOp  bool
	}{
		{"A folder that exists in the file system", ExistsFolder, true},
		{"A folder that does not exist in the file system", NotExistsFolder, false},
		{"A file that exists in the file system", ExistsTextFile, false},
		{"A file that does not exist in the file system", NotExistsTextFile, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			isDir := fs.IsDirectory(testCase.IpPath)
			if isDir == testCase.ExpOp {
				tt.Logf("The result returned by isDirectory() for path [%s] matches the expected output", testCase.IpPath)
			} else {
				tt.Errorf(internal.TextColor.Red("The result returned by isDirectory() for path [%s] does not match the expected output"), testCase.IpPath)
			}
		})
	}
}

// Test case to validate the working of GetFile() method of FileSystem.
func Test_FileSystem_GetFile(t *testing.T) {
	root := t.TempDir()
	Folder := filepath.Join(root, "folder-tc")
	ExistsFile := filepath.Join(root, "exists.txt")
	NotExistsFile := filepath.Join(root, "not-exists.txt")
	err := CreateDirectories(t, root, []string{"folder-tc"})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating test folder: %s"), err.Error())
		return
	}
	err = CreateFiles(t, root, map[string][]byte{
		"exists.txt": []byte("This is a sample text!"),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating test file: %s"), err.Error())
		return
	}
	fs := new(internal.FileSystem)
	testCases := []struct {
		Name         string
		IpPath       string
		ExpSize      int64
		ExpExtension string
		ExpMediaType string
		ExpError     string
	}{
		{"A valid text file available in the file system", ExistsFile, 22, "txt", "text/plain", ""},
		{"A valid text file not available in the file system", NotExistsFile, 0, "txt", "text/plain", "FileSystemError"},
		{"A path pointing to a folder in the file system", Folder, 0, "", "text/plain", "FileSystemError"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			file, err := fs.GetFile(testCase.IpPath)
			if err != nil {
				if strings.EqualFold(testCase.ExpError, "FileSystemError") {
					fsErr, ok := err.(*internal.FileSystemError)
					if !ok {
						tt.Errorf(internal.TextColor.Red("Expected a FileSystemError, but received something else: %#v"), err)
					} else {
						tt.Logf("Expected a FileSystemError, and received an error of same type - %#v", *fsErr)
					}
				} else {
					tt.Errorf(internal.TextColor.Red("An error was not expected, but yet received one - %s"), err.Error())
				}
				return
			}

			if file.Size() == testCase.ExpSize {
				tt.Logf("The expected file size [%d] matches the computed file size [%d]", testCase.ExpSize, file.Size())
			} else {
				tt.Errorf(internal.TextColor.Red("The expected file size [%d] does not match the computed file size [%d]"), testCase.ExpSize, file.Size())
			}

			if strings.EqualFold(file.MediaType(), testCase.ExpMediaType) {
				tt.Logf("The expected media type [%s] matches the fetched file's media type [%s]", testCase.ExpMediaType, file.MediaType())
			} else {
				tt.Errorf(internal.TextColor.Red("The expected media type [%s] does not match the fetched file's media type [%s]"), testCase.ExpMediaType, file.MediaType())
			}

			if strings.EqualFold(file.Extension(), testCase.ExpExtension) {
				tt.Logf("The expected file extension [%s] matches the fetched file's extension [%s]", testCase.ExpExtension, file.Extension())
			} else {
				tt.Errorf(internal.TextColor.Red("The expected file extension [%s] does not match the fetched file's extension [%s]"), testCase.ExpExtension, file.Extension())
			}
		})
	}
}

// Test case to validate the working of method that fetches the contents of a given file from the underlying file system.
func Test_FileSystem_FetchContents(t *testing.T) {
	root := t.TempDir()
	FileContentsToBeWritten := "This is a sample text\nThis is another line\nAnd another line"
	err := CreateFiles(t, root, map[string][]byte{
		"file-to-be-read.txt": []byte(FileContentsToBeWritten),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating test file: %s"), err.Error())
		return
	}
	FileToBeRead := filepath.Join(root, "file-to-be-read.txt")
	fs := new(internal.FileSystem)
	file, err := fs.GetFile(FileToBeRead)
	if err != nil {
		t.Errorf(internal.TextColor.Red("Error occurred while getting file properties, instead of successful parse opertion: %s"), err.Error())
		return
	}
	fileContentBytes, err := file.Contents()
	if err != nil {
		t.Errorf(internal.TextColor.Red("Error occurred while reading the file contents, instead of successful read opertion: %s"), err.Error())
		return
	}

	fileContents := string(fileContentBytes)
	if strings.EqualFold(FileContentsToBeWritten, fileContents) {
		t.Logf("The contents of file [%s] read from the file system matches the expected content", FileToBeRead)
	} else {
		t.Errorf(internal.TextColor.Red("The contents of file [%s] read from the file system does not match the expected content"), FileToBeRead)
	}
}

// Test case to validate the Exists() function of FileSystem.
func Test_FileSystem_Exists(t *testing.T) {
	root := t.TempDir()
	err := CreateDirectories(t, root, []string{"static"})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating temporary folders for testing: %s"), err.Error())
		return
	}
	err = CreateFiles(t, root, map[string][]byte{
		"home.html": []byte("<p>Hello, World!</p>"),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error occurred while creating temporary files for testing: %s"), err.Error())
		return
	}
	fs := new(internal.FileSystem)
	testCases := []struct {
		Name    string
		IpPath  string
		OpValue bool
	}{
		{"A folder that exists in the file system", filepath.Join(root, "static"), true},
		{"A folder that does not exist in the file system", filepath.Join(root, "public"), false},
		{"A file that exists in the file system", filepath.Join(root, "home.html"), true},
		{"A file that does not exist in the file system", filepath.Join(root, "index.html"), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			isExists := fs.Exists(testCase.IpPath)
			if isExists == testCase.OpValue {
				tt.Logf("The expected value [%t] matches the returned value [%t].", testCase.OpValue, isExists)
			} else {
				tt.Errorf(internal.TextColor.Red("The expected value [%t] does not match the returned value [%t]."), testCase.OpValue, isExists)
			}
		})
	}
}

// Test case to validate File.Contents() with large files and edge cases.
func Test_File_Contents_EdgeCases(t *testing.T) {
	root := t.TempDir()
	fs := new(internal.FileSystem)

	// Create files with different content sizes and types
	smallContent := "Small content"
	largeContent := strings.Repeat("A", 5000) // Content larger than chunk size (1024)
	binaryContent := make([]byte, 2048)
	for i := range binaryContent {
		binaryContent[i] = byte(i % 256)
	}

	err := CreateFiles(t, root, map[string][]byte{
		"small.txt":  []byte(smallContent),
		"large.txt":  []byte(largeContent),
		"binary.bin": binaryContent,
		"empty.txt":  []byte(""),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error creating test files: %s"), err.Error())
		return
	}

	testCases := []struct {
		Name            string
		FileName        string
		ExpectedSize    int
		ExpectedContent []byte
	}{
		{"Small text file", "small.txt", len(smallContent), []byte(smallContent)},
		{"Large text file", "large.txt", len(largeContent), []byte(largeContent)},
		{"Binary file", "binary.bin", len(binaryContent), binaryContent},
		{"Empty file", "empty.txt", 0, []byte("")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			filePath := filepath.Join(root, testCase.FileName)
			file, err := fs.GetFile(filePath)
			if err != nil {
				tt.Errorf(internal.TextColor.Red("Error getting file: %s"), err.Error())
				return
			}

			contents, err := file.Contents()
			if err != nil {
				tt.Errorf(internal.TextColor.Red("Error reading file contents: %s"), err.Error())
				return
			}

			if len(contents) != testCase.ExpectedSize {
				tt.Errorf(internal.TextColor.Red("Expected content size %d, got %d"), testCase.ExpectedSize, len(contents))
			} else {
				tt.Logf("File contents size matches expected size: %d bytes", len(contents))
			}

			if !bytes.Equal(contents, testCase.ExpectedContent) {
				tt.Error(internal.TextColor.Red("File contents don't match expected content"))
			} else {
				tt.Log("File contents match expected content")
			}
		})
	}
}

// Test case to validate File.MediaType() with unknown and edge case extensions.
func Test_File_MediaType_EdgeCases(t *testing.T) {
	root := t.TempDir()
	fs := new(internal.FileSystem)

	err := CreateFiles(t, root, map[string][]byte{
		"test.unknown": []byte("content"),
		"test.XYZ":     []byte("content"),
		"test.":        []byte("content"),
		"test":         []byte("content"),
		"test.PDF":     []byte("content"), // uppercase extension
		"test.Html":    []byte("content"), // mixed case extension
		"test.mp4":     []byte("content"), // extension not in AllowedContentTypes
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error creating test files: %s"), err.Error())
		return
	}

	defaultContentType := internal.GetServerDefaults("content_type").(string)

	testCases := []struct {
		Name                string
		FileName            string
		ExpectedContentType string
	}{
		{"Unknown extension", "test.unknown", defaultContentType},
		{"Unknown uppercase extension", "test.XYZ", defaultContentType},
		{"Empty extension", "test.", defaultContentType},
		{"No extension", "test", defaultContentType},
		{"Uppercase PDF extension", "test.PDF", "application/pdf"}, // should normalize to lowercase
		{"Mixed case HTML extension", "test.Html", "text/html"},    // should normalize to lowercase
		{"Extension not in allowed list", "test.mp4", defaultContentType},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			filePath := filepath.Join(root, testCase.FileName)
			file, err := fs.GetFile(filePath)
			if err != nil {
				tt.Errorf(internal.TextColor.Red("Error getting file: %s"), err.Error())
				return
			}

			mediaType := file.MediaType()
			if strings.EqualFold(mediaType, testCase.ExpectedContentType) {
				tt.Logf("Media type for [%s] correctly determined as [%s]", testCase.FileName, mediaType)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected media type [%s] for file [%s], got [%s]"), testCase.ExpectedContentType, testCase.FileName, mediaType)
			}
		})
	}
}

// Test case to validate File.Extension() with various file name patterns.
func Test_File_Extension_EdgeCases(t *testing.T) {
	root := t.TempDir()
	fs := new(internal.FileSystem)

	err := CreateFiles(t, root, map[string][]byte{
		"test.txt":             []byte("content"),
		"test.TAR.GZ":          []byte("content"),
		"test.":                []byte("content"),
		"test":                 []byte("content"),
		".hidden":              []byte("content"),
		".hidden.txt":          []byte("content"),
		"multi.part.name.html": []byte("content"),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error creating test files: %s"), err.Error())
		return
	}

	testCases := []struct {
		Name              string
		FileName          string
		ExpectedExtension string
	}{
		{"Simple extension", "test.txt", "txt"},
		{"Uppercase extension", "test.TAR.GZ", "gz"}, // should return last extension in lowercase
		{"Empty extension", "test.", ""},
		{"No extension", "test", ""},
		{"Hidden file no extension", ".hidden", "hidden"},
		{"Hidden file with extension", ".hidden.txt", "txt"},
		{"Multiple dots in filename", "multi.part.name.html", "html"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			filePath := filepath.Join(root, testCase.FileName)
			file, err := fs.GetFile(filePath)
			if err != nil {
				tt.Errorf(internal.TextColor.Red("Error getting file: %s"), err.Error())
				return
			}

			extension := file.Extension()
			if strings.EqualFold(extension, testCase.ExpectedExtension) {
				tt.Logf("Extension for [%s] correctly determined as [%s]", testCase.FileName, extension)
			} else {
				tt.Errorf("Expected extension [%s] for file [%s], got [%s]", testCase.ExpectedExtension, testCase.FileName, extension)
			}
		})
	}
}

// Test case to validate FileSystem.CleanPath() with various path patterns.
func Test_FileSystem_CleanPath_EdgeCases(t *testing.T) {
	fs := new(internal.FileSystem)

	testCases := []struct {
		Name         string
		InputPath    string
		ExpectedPath string
	}{
		{"Already clean path", "/home/user/file.txt", "/home/user/file.txt"},
		{"Path with double slashes", "/home//user//file.txt", "/home/user/file.txt"},
		{"Path with dot segments", "/home/user/./file.txt", "/home/user/file.txt"},
		{"Path with dotdot segments", "/home/user/../other/file.txt", "/home/other/file.txt"},
		{"Path with trailing slash", "/home/user/", "/home/user"},
		{"Path with leading and trailing whitespace", "  /home/user/file.txt  ", "/home/user/file.txt"},
		{"Relative path", "user/file.txt", "user/file.txt"},
		{"Current directory", ".", "."},
		{"Parent directory", "..", ".."},
		{"Complex path with multiple issues", "  /home//user/../other/./file.txt/  ", "/home/other/file.txt"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			cleanedPath := fs.CleanPath(testCase.InputPath)
			if cleanedPath == testCase.ExpectedPath {
				tt.Logf("Path [%s] correctly cleaned to [%s]", testCase.InputPath, cleanedPath)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected cleaned path [%s] for input [%s], got [%s]"), testCase.ExpectedPath, testCase.InputPath, cleanedPath)
			}
		})
	}
}
