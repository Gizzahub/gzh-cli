package filesystem

import (
	"context"
	"io"
	"io/fs"
	"time"
)

// FileInfo represents information about a file or directory.
type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	ModTime time.Time   `json:"modTime"`
	IsDir   bool        `json:"isDir"`
	Path    string      `json:"path"`
}

// FileSystem defines the interface for file system operations.
type FileSystem interface {
	// File operations
	ReadFile(ctx context.Context, filename string) ([]byte, error)
	WriteFile(ctx context.Context, filename string, data []byte, perm fs.FileMode) error
	AppendFile(ctx context.Context, filename string, data []byte) error

	// File existence and properties
	Exists(ctx context.Context, path string) bool
	IsFile(ctx context.Context, path string) bool
	IsDir(ctx context.Context, path string) bool
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
	GetFileSize(ctx context.Context, path string) (int64, error)

	// Directory operations
	MkdirAll(ctx context.Context, path string, perm fs.FileMode) error
	RemoveAll(ctx context.Context, path string) error
	ListDir(ctx context.Context, path string) ([]FileInfo, error)
	WalkDir(ctx context.Context, root string, fn func(path string, info FileInfo, err error) error) error

	// File manipulation
	CopyFile(ctx context.Context, src, dst string) error
	MoveFile(ctx context.Context, src, dst string) error
	CreateSymlink(ctx context.Context, oldname, newname string) error
	ReadSymlink(ctx context.Context, name string) (string, error)

	// File streaming
	OpenFile(ctx context.Context, name string, flag int, perm fs.FileMode) (File, error)
	CreateFile(ctx context.Context, name string) (File, error)

	// Path operations
	Abs(path string) (string, error)
	Join(paths ...string) string
	Dir(path string) string
	Base(path string) string
	Ext(path string) string
	Clean(path string) string

	// Temporary files and directories
	TempDir(ctx context.Context, dir, pattern string) (string, error)
	TempFile(ctx context.Context, dir, pattern string) (File, error)
}

// File defines the interface for file operations.
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer

	// File-specific operations
	Name() string
	Stat() (FileInfo, error)
	Sync() error
	Truncate(size int64) error

	// Read operations
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)

	// Write operations
	Write(b []byte) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	WriteString(s string) (n int, err error)
}

// WatchService defines the interface for file system watching.
type WatchService interface {
	// Start watching paths for changes
	Watch(ctx context.Context, paths []string) error

	// Stop watching
	Stop() error

	// Get events channel
	Events() <-chan WatchEvent

	// Get errors channel
	Errors() <-chan error

	// Add path to watch
	AddPath(ctx context.Context, path string) error

	// Remove path from watch
	RemovePath(ctx context.Context, path string) error
}

// WatchEvent represents a file system change event.
type WatchEvent struct {
	Path      string    `json:"path"`
	Operation string    `json:"operation"` // create, write, remove, rename, chmod
	IsDir     bool      `json:"isDir"`
	Time      time.Time `json:"time"`
}

// PermissionManager defines the interface for managing file permissions.
type PermissionManager interface {
	// Get file permissions
	GetPermissions(ctx context.Context, path string) (fs.FileMode, error)

	// Set file permissions
	SetPermissions(ctx context.Context, path string, mode fs.FileMode) error

	// Check if path is readable
	IsReadable(ctx context.Context, path string) bool

	// Check if path is writable
	IsWritable(ctx context.Context, path string) bool

	// Check if path is executable
	IsExecutable(ctx context.Context, path string) bool

	// Get file owner information
	GetOwner(ctx context.Context, path string) (string, error)

	// Change file owner
	ChangeOwner(ctx context.Context, path, owner string) error
}

// ArchiveService defines the interface for archive operations.
type ArchiveService interface {
	// Create archive from directory
	CreateArchive(ctx context.Context, sourcePath, archivePath string, format ArchiveFormat) error

	// Extract archive to directory
	ExtractArchive(ctx context.Context, archivePath, destPath string) error

	// List archive contents
	ListArchive(ctx context.Context, archivePath string) ([]FileInfo, error)

	// Get supported formats
	GetSupportedFormats() []ArchiveFormat
}

// ArchiveFormat represents supported archive formats.
type ArchiveFormat string

// Archive format constants define the supported archive types.
const (
	ArchiveFormatTar    ArchiveFormat = "tar"
	ArchiveFormatTarGz  ArchiveFormat = "tar.gz"
	ArchiveFormatTarBz2 ArchiveFormat = "tar.bz2"
	ArchiveFormatZip    ArchiveFormat = "zip"
)

// BackupService defines the interface for file backup operations.
type BackupService interface {
	// Create backup of file or directory
	CreateBackup(ctx context.Context, sourcePath, backupPath string) error

	// Restore from backup
	RestoreBackup(ctx context.Context, backupPath, destPath string) error

	// List available backups
	ListBackups(ctx context.Context, path string) ([]BackupInfo, error)

	// Delete backup
	DeleteBackup(ctx context.Context, backupPath string) error

	// Verify backup integrity
	VerifyBackup(ctx context.Context, backupPath string) error
}

// BackupInfo represents information about a backup.
type BackupInfo struct {
	Path         string    `json:"path"`
	OriginalPath string    `json:"originalPath"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"createdAt"`
	Checksum     string    `json:"checksum"`
}

// SearchService defines the interface for file search operations.
type SearchService interface {
	// Find files by name pattern
	FindFiles(ctx context.Context, root, pattern string) ([]string, error)

	// Find files by content
	FindInFiles(ctx context.Context, root, pattern string) ([]SearchResult, error)

	// Find directories
	FindDirectories(ctx context.Context, root, pattern string) ([]string, error)

	// Search with advanced filters
	SearchWithFilters(ctx context.Context, root string, filters SearchFilters) ([]SearchResult, error)
}

// SearchResult represents a search result.
type SearchResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Match   string `json:"match"`
	Context string `json:"context"`
}

// SearchFilters represents search filter options.
type SearchFilters struct {
	NamePattern    string    `json:"namePattern"`
	ContentPattern string    `json:"contentPattern"`
	Extensions     []string  `json:"extensions"`
	MinSize        int64     `json:"minSize"`
	MaxSize        int64     `json:"maxSize"`
	ModifiedAfter  time.Time `json:"modifiedAfter"`
	ModifiedBefore time.Time `json:"modifiedBefore"`
	IncludeDirs    bool      `json:"includeDirs"`
	FollowSymlinks bool      `json:"followSymlinks"`
	MaxDepth       int       `json:"maxDepth"`
}

// Service provides a unified interface for all file system operations.
type Service interface {
	FileSystem
	WatchService
	PermissionManager
	ArchiveService
	BackupService
	SearchService
}
