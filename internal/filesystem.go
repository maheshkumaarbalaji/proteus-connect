package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Size in bytes for each chunk of data being read from a file.
	CHUNK_SIZE = 1024
)

// Structure to represent a file in the local file system.
type File struct {
	// Base name of the file.
	Name string
	// Complete Path of the file in the local file system.
	Path string
	// Stats interface associated with the given file. If the value is nil, it implies the path points to a file that does not exist.
	stats os.FileInfo
}

// Contents reads and returns complete file content as byte slice with chunked processing.
// Use for static file serving, template loading, file uploads, and binary content handling.
// Opens file, reads in configurable chunks, assembles complete content, closes handle automatically.
// Parameters: No parameters (uses file.Path).
// Returns: []byte (complete file content), error (FileSystemError if reading fails).
func (file *File) Contents() ([]byte, error) {
	CompleteFilePath := file.Path
	fileContents := make([]byte, 0)
	fileHandler, err := os.Open(CompleteFilePath)
	if err != nil {
		fsfErr := new(FileSystemError)
		fsfErr.TargetPath = CompleteFilePath
		fsfErr.Message = fmt.Sprintf("Error occurred while reading file contents: %s", err.Error())
		return nil, fsfErr
	}
	defer fileHandler.Close()
	reader := bufio.NewReader(fileHandler)
	for {
		chunk := make([]byte, CHUNK_SIZE)
		bytesRead, err := reader.Read(chunk)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		if bytesRead < CHUNK_SIZE {
			newChunk := chunk[0:bytesRead]
			fileContents = append(fileContents, newChunk...)
		} else {
			fileContents = append(fileContents, chunk...)
		}
	}

	return fileContents, nil
}

// Extension extracts file extension in normalized lowercase format without leading period.
// Use for MIME type detection, file validation, routing decisions, and security checks.
// Removes period, converts to lowercase, trims whitespace for consistent comparison.
// Parameters: No parameters (uses file.Path).
// Returns: string (normalized file extension like "pdf", "html", empty string if no extension).
func (file *File) Extension() string {
	CompleteFilePath := file.Path
	fileExtension := filepath.Ext(CompleteFilePath)
	fileExtension = strings.TrimPrefix(fileExtension, ".")
	fileExtension = strings.TrimSpace(fileExtension)
	fileExtension = strings.ToLower(fileExtension)
	return fileExtension
}

// MediaType determines MIME type for file based on extension for proper HTTP Content-Type headers.
// Use for static file serving, file uploads, API responses, and browser rendering control.
// Looks up extension in internal registry, supports web/document/multimedia types, fallback to default.
// Parameters: No parameters (uses file extension).
// Returns: string (MIME type like "text/html", "image/png", "application/pdf").
//
// Note: MIME type detection is based solely on file extension.
// For security-sensitive applications, consider additional content
// validation beyond extension-based detection.
func (file *File) MediaType() string {
	fileExtension := file.Extension()
	contentType, exists := AllowedContentTypes[fileExtension]
	if exists {
		return contentType
	} else {
		defaultContentType := GetServerDefaults("content_type").(string)
		return strings.TrimSpace(defaultContentType)
	}
}

// Size returns file size in bytes from filesystem metadata without reading content.
// Use for Content-Length headers, upload validation, bandwidth estimation, and quota management.
// Retrieved from os.FileInfo statistics, efficient for large files, accurate for all file types.
// Parameters: No parameters (uses file.stats).
// Returns: int64 (file size in bytes, 0 if file doesn't exist or stats unavailable).
func (file *File) Size() int64 {
	if file.stats == nil {
		return 0
	} else {
		return file.stats.Size()
	}
}

// LastModified returns file modification timestamp from filesystem metadata for caching support.
// Use for Last-Modified headers, conditional GET requests, cache invalidation, and file monitoring.
// Retrieved from os.FileInfo, returns zero time if file doesn't exist, suitable for HTTP headers.
// Parameters: No parameters (uses file.stats).
// Returns: time.Time (file modification timestamp, zero time if file doesn't exist).
//
// Note: Modification time precision and behavior may vary between
// different filesystems and operating systems.
func (file *File) LastModified() time.Time {
	if file.stats == nil {
		return time.Time{}
	} else {
		return file.stats.ModTime()
	}
}

// Structure to connect to the local file system and access files/folders.
type FileSystem struct{}

// CleanPath normalizes filesystem paths by removing redundant elements and standardizing separators.
// Use for security (prevents path traversal), consistency, and cross-platform compatibility.
// Removes redundant separators, resolves . and .. references, trims whitespace automatically.
// Parameter: Path (raw filesystem path that may contain redundant elements).
// Returns: string (cleaned and normalized path suitable for filesystem operations).
func (fs *FileSystem) CleanPath(Path string) string {
	Path = strings.TrimSpace(Path)
	Path = filepath.Clean(Path)
	return Path
}

// GetFile creates File instance with metadata for specified path with validation and normalization.
// Use for static file serving, file access validation, and metadata extraction operations.
// Verifies existence, ensures regular file (not directory), extracts metadata, cleans path.
// Parameter: CompleteFilePath (absolute or relative path to target file).
// Returns: *File (instance with metadata and methods), error (FileSystemError if access fails).
func (fs *FileSystem) GetFile(CompleteFilePath string) (*File, error) {
	CompleteFilePath = fs.CleanPath(CompleteFilePath)
	fileStat, err := os.Stat(CompleteFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fsfErr := new(FileSystemError)
			fsfErr.TargetPath = CompleteFilePath
			fsfErr.Message = "GetFile :: File or Directory referenced by the given path does not exist in the file system"
			return nil, fsfErr
		}
		fsfErr := new(FileSystemError)
		fsfErr.TargetPath = CompleteFilePath
		fsfErr.Message = fmt.Sprintf("GetFile: Error occurred while fetching file stats: %s", err.Error())
		return nil, fsfErr
	}
	fileMode := fileStat.Mode()
	if fileMode.IsRegular() {
		file := new(File)
		file.Path = CompleteFilePath
		file.Name = filepath.Base(file.Path)
		file.stats = fileStat
		return file, nil
	} else {
		fsfErr := new(FileSystemError)
		fsfErr.TargetPath = CompleteFilePath
		fsfErr.Message = "Given path does not point to a file"
		return nil, fsfErr
	}
}

// IsAbsolute determines whether provided path is absolute filesystem path for security validation.
// Use for path validation, security checks, configuration verification, and preventing traversal attacks.
// Checks Unix ("/path") and Windows ("C:\path", "\\server\share") absolute path formats.
// Parameter: CompleteFilePath (path string to validate, cleaned automatically).
// Returns: bool (true if path is absolute, false if relative).
func (fs *FileSystem) IsAbsolute(CompleteFilePath string) bool {
	CompleteFilePath = fs.CleanPath(CompleteFilePath)
	return filepath.IsAbs(CompleteFilePath)
}

// IsDirectory determines whether specified path points to accessible directory with safe error handling.
// Use for static file path validation, upload directory verification, and configuration checks.
// Cleans path, retrieves filesystem stats, returns false for non-existent/inaccessible paths.
// Parameter: CompletePath (path to validate, cleaned automatically).
// Returns: bool (true if path is accessible directory, false otherwise).
func (fs *FileSystem) IsDirectory(CompletePath string) bool {
	CompletePath = fs.CleanPath(CompletePath)
	stats, err := os.Stat(CompletePath)
	if err != nil {
		return false
	}
	mode := stats.Mode()
	if mode.IsDir() {
		return true
	} else {
		return false
	}
}

// Exists checks whether file or directory exists at specified path with reliable validation.
// Use for file upload conflict detection, configuration presence checking, and asset validation.
// Cleans path, uses lightweight os.Stat() operation, works with all filesystem entity types.
// Parameter: CompletePath (path to check, cleaned automatically).
// Returns: bool (true if path exists and accessible, false otherwise).
func (fs *FileSystem) Exists(CompletePath string) bool {
	CompletePath = fs.CleanPath(CompletePath)
	_, err := os.Stat(CompletePath)
	return err == nil
}
