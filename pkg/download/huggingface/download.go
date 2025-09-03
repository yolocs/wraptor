package huggingface

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gomlx/go-huggingface/hub"
	"github.com/yolocs/wraptor/pkg/file"
)

// Downloader handles downloading Hugging Face models to local temp directory
// and provides them as SizedReader for use with wrap.go
type Downloader struct {
	repoID         string
	cacheDir       string
	maxConcurrency int
	authToken      string
	revision       string
}

// Option is a function that configures a Downloader
type Option func(*Downloader)

// WithAuth sets the Hugging Face authentication token
func WithAuth(token string) Option {
	return func(d *Downloader) {
		d.authToken = token
	}
}

// WithCacheDir sets the cache directory for downloaded files
func WithCacheDir(dir string) Option {
	return func(d *Downloader) {
		d.cacheDir = dir
	}
}

// WithMaxConcurrency sets the maximum number of concurrent downloads
func WithMaxConcurrency(max int) Option {
	return func(d *Downloader) {
		d.maxConcurrency = max
	}
}

// WithRevision sets the model revision/branch to download
func WithRevision(revision string) Option {
	return func(d *Downloader) {
		d.revision = revision
	}
}

// New creates a new Downloader for the specified Hugging Face repository
func New(repoID string, opts ...Option) *Downloader {
	d := &Downloader{
		repoID:         repoID,
		cacheDir:       "./.wraptor/caches",
		maxConcurrency: 5,
		revision:       "main",
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// createCacheDir creates a unique cache subdirectory for this repo/revision
func (d *Downloader) createCacheDir() (string, error) {
	// Create a unique subdirectory based on repo ID and revision
	hash := md5.Sum([]byte(d.repoID + ":" + d.revision))
	uniqueDir := fmt.Sprintf("%x", hash)

	cacheDir := filepath.Join(d.cacheDir, uniqueDir)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	return cacheDir, nil
}

// Load downloads all files from the Hugging Face repository and returns them as file.Reader slice
func (d *Downloader) Load(ctx context.Context) ([]*file.Reader, error) {
	// Create cache directory
	cacheDir, err := d.createCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create repo with configuration
	repo := hub.New(d.repoID).
		WithCacheDir(cacheDir).
		WithRevision(d.revision)

	repo.MaxParallelDownload = d.maxConcurrency

	if d.authToken != "" {
		repo = repo.WithAuth(d.authToken)
	}

	// Get all file names from the repo
	var fileNames []string
	for fileName, err := range repo.IterFileNames() {
		if err != nil {
			return nil, fmt.Errorf("failed to iterate file names: %w", err)
		}
		fileNames = append(fileNames, fileName)
	}

	if len(fileNames) == 0 {
		return nil, fmt.Errorf("no files found in repository %s", d.repoID)
	}

	// Download all files
	downloadedPaths, err := repo.DownloadFiles(fileNames...)
	if err != nil {
		return nil, fmt.Errorf("failed to download files: %w", err)
	}

	// Convert downloaded files to file.Reader slice
	var readers []*file.Reader
	for i, filePath := range downloadedPaths {
		f, err := os.Open(filePath)
		if err != nil {
			// Clean up already opened files
			for _, reader := range readers {
				reader.Close()
			}
			return nil, fmt.Errorf("failed to open downloaded file %s: %w", filePath, err)
		}

		// Get file info for size
		fileInfo, err := f.Stat()
		if err != nil {
			f.Close()
			// Clean up already opened files
			for _, reader := range readers {
				reader.Close()
			}
			return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
		}

		reader := &file.Reader{
			ReadCloser: f,
			Name:       fileNames[i], // Use original filename, not path
			Size:       fileInfo.Size(),
		}

		readers = append(readers, reader)
	}

	return readers, nil
}

// Cleanup removes the cache directory for this downloader's repo/revision
func (d *Downloader) Cleanup() error {
	cacheDir, err := d.createCacheDir()
	if err != nil {
		return fmt.Errorf("failed to determine cache directory: %w", err)
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to remove cache directory %s: %w", cacheDir, err)
	}

	return nil
}
