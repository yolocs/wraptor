package huggingface

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	downloader := New("microsoft/DialoGPT-medium")
	defer func() {
		if err := downloader.Cleanup(); err != nil {
			t.Logf("Failed to cleanup: %v", err)
		}
	}()

	if downloader.repoID != "microsoft/DialoGPT-medium" {
		t.Errorf("Expected repoID to be 'microsoft/DialoGPT-medium', got %s", downloader.repoID)
	}

	if downloader.cacheDir != "./.wraptor/caches" {
		t.Errorf("Expected default cacheDir to be './.wraptor/caches', got %s", downloader.cacheDir)
	}

	if downloader.maxConcurrency != 5 {
		t.Errorf("Expected default maxConcurrency to be 5, got %d", downloader.maxConcurrency)
	}

	if downloader.revision != "main" {
		t.Errorf("Expected default revision to be 'main', got %s", downloader.revision)
	}
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	downloader := New("test/repo",
		WithAuth("test-token"),
		WithCacheDir("/tmp/test"),
		WithMaxConcurrency(10),
		WithRevision("v1.0"))
	defer func() {
		if err := downloader.Cleanup(); err != nil {
			t.Logf("Failed to cleanup: %v", err)
		}
	}()

	if downloader.authToken != "test-token" {
		t.Errorf("Expected authToken to be 'test-token', got %s", downloader.authToken)
	}

	if downloader.cacheDir != "/tmp/test" {
		t.Errorf("Expected cacheDir to be '/tmp/test', got %s", downloader.cacheDir)
	}

	if downloader.maxConcurrency != 10 {
		t.Errorf("Expected maxConcurrency to be 10, got %d", downloader.maxConcurrency)
	}

	if downloader.revision != "v1.0" {
		t.Errorf("Expected revision to be 'v1.0', got %s", downloader.revision)
	}
}

func TestCreateCacheDir(t *testing.T) {
	downloader := New("test/repo")
	defer func() {
		if err := downloader.Cleanup(); err != nil {
			t.Logf("Failed to cleanup: %v", err)
		}
	}()

	cacheDir, err := downloader.createCacheDir()
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	if cacheDir == "" {
		t.Error("Cache directory path should not be empty")
	}

	// Test that same repo/revision creates same directory
	cacheDir2, err := downloader.createCacheDir()
	if err != nil {
		t.Fatalf("Failed to create cache directory second time: %v", err)
	}

	if cacheDir != cacheDir2 {
		t.Errorf("Same repo should create same cache directory, got %s and %s", cacheDir, cacheDir2)
	}
}

// Note: This test would require network access to HuggingFace, so it's commented out
// Uncomment and run manually to test actual downloading
// func TestLoadSmallModel(t *testing.T) {
// 	// Use a very small model for testing
// 	downloader := New("hf-internal-testing/tiny-random-bert")

// 	ctx := context.Background()
// 	readers, err := downloader.Load(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to load model: %v", err)
// 	}

// 	if len(readers) == 0 {
// 		t.Error("Expected at least one file, got none")
// 	}

// 	// Clean up
// 	for _, reader := range readers {
// 		fmt.Println("Closing reader for file:", reader.Name())
// 		reader.Close()
// 	}
// }

func TestCleanup(t *testing.T) {
	downloader := New("test/repo")

	// Create cache directory
	cacheDir, err := downloader.createCacheDir()
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatalf("Cache directory should exist but doesn't: %s", cacheDir)
	}

	// Test cleanup
	err = downloader.Cleanup()
	if err != nil {
		t.Fatalf("Failed to cleanup cache directory: %v", err)
	}

	// Verify directory is removed
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Errorf("Cache directory should be removed but still exists: %s", cacheDir)
	}
}
