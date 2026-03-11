package comic

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveError(t *testing.T) {
	err := ArchiveError{Message: "archive failed", Code: 1}
	if err.Error() != "archive failed" {
		t.Errorf("Error() = %q, want %q", err.Error(), "archive failed")
	}
}

func TestArchive(t *testing.T) {
	t.Run("creates cbz with image files", func(t *testing.T) {
		tmpDir := t.TempDir()
		title := "TestComic"
		comicDir := filepath.Join(tmpDir, title)
		os.MkdirAll(comicDir, os.ModePerm)

		// Create fake image files
		for _, name := range []string{"TestComic 001.jpg", "TestComic 002.jpg", "TestComic 003.png"} {
			os.WriteFile(filepath.Join(comicDir, name), []byte("fake image"), 0644)
		}

		c := &Comic{
			Title:       title,
			LibraryPath: tmpDir,
		}

		err := c.Archive()
		if err != nil {
			t.Fatalf("Archive() unexpected error: %v", err)
		}

		archivePath := filepath.Join(comicDir, title+".cbz")
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Fatalf("expected archive %s to exist", archivePath)
		}

		// Verify the zip contains the image files
		reader, err := zip.OpenReader(archivePath)
		if err != nil {
			t.Fatalf("failed to open archive: %v", err)
		}
		defer reader.Close()

		if len(reader.File) != 3 {
			t.Errorf("archive contains %d files, want 3", len(reader.File))
		}
	})

	t.Run("excludes non-image files from archive", func(t *testing.T) {
		tmpDir := t.TempDir()
		title := "TestComic"
		comicDir := filepath.Join(tmpDir, title)
		os.MkdirAll(comicDir, os.ModePerm)

		// Create mixed files
		os.WriteFile(filepath.Join(comicDir, "page-001.jpg"), []byte("image"), 0644)
		os.WriteFile(filepath.Join(comicDir, "readme.txt"), []byte("text"), 0644)
		os.WriteFile(filepath.Join(comicDir, "data.json"), []byte("json"), 0644)

		c := &Comic{
			Title:       title,
			LibraryPath: tmpDir,
		}

		err := c.Archive()
		if err != nil {
			t.Fatalf("Archive() unexpected error: %v", err)
		}

		archivePath := filepath.Join(comicDir, title+".cbz")
		reader, err := zip.OpenReader(archivePath)
		if err != nil {
			t.Fatalf("failed to open archive: %v", err)
		}
		defer reader.Close()

		if len(reader.File) != 1 {
			t.Errorf("archive contains %d files, want 1 (only .jpg)", len(reader.File))
		}
	})

	t.Run("handles empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		title := "EmptyComic"
		comicDir := filepath.Join(tmpDir, title)
		os.MkdirAll(comicDir, os.ModePerm)

		c := &Comic{
			Title:       title,
			LibraryPath: tmpDir,
		}

		err := c.Archive()
		if err != nil {
			t.Fatalf("Archive() unexpected error: %v", err)
		}

		archivePath := filepath.Join(comicDir, title+".cbz")
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Fatalf("expected archive %s to exist even if empty", archivePath)
		}
	})
}
