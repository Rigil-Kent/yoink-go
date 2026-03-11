package comic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanup(t *testing.T) {
	t.Run("keeps cover image 001 and removes others", func(t *testing.T) {
		tmpDir := t.TempDir()
		title := "TestComic"
		comicDir := filepath.Join(tmpDir, title)
		os.MkdirAll(comicDir, os.ModePerm)

		files := map[string]bool{
			"TestComic 001.jpg": true,  // should be kept
			"TestComic 002.jpg": false, // should be removed
			"TestComic 003.jpg": false, // should be removed
		}

		for name := range files {
			os.WriteFile(filepath.Join(comicDir, name), []byte("fake"), 0644)
		}

		c := &Comic{
			Title:       title,
			LibraryPath: tmpDir,
		}

		err := c.Cleanup()
		if err != nil {
			t.Fatalf("Cleanup() unexpected error: %v", err)
		}

		for name, shouldExist := range files {
			path := filepath.Join(comicDir, name)
			_, err := os.Stat(path)
			exists := !os.IsNotExist(err)

			if shouldExist && !exists {
				t.Errorf("expected %s to be kept, but it was removed", name)
			}
			if !shouldExist && exists {
				t.Errorf("expected %s to be removed, but it still exists", name)
			}
		}
	})

	t.Run("keeps non-image files", func(t *testing.T) {
		tmpDir := t.TempDir()
		title := "TestComic"
		comicDir := filepath.Join(tmpDir, title)
		os.MkdirAll(comicDir, os.ModePerm)

		os.WriteFile(filepath.Join(comicDir, "TestComic.cbz"), []byte("archive"), 0644)
		os.WriteFile(filepath.Join(comicDir, "metadata.json"), []byte("data"), 0644)

		c := &Comic{
			Title:       title,
			LibraryPath: tmpDir,
		}

		err := c.Cleanup()
		if err != nil {
			t.Fatalf("Cleanup() unexpected error: %v", err)
		}

		for _, name := range []string{"TestComic.cbz", "metadata.json"} {
			path := filepath.Join(comicDir, name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("expected non-image file %s to be kept", name)
			}
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

		err := c.Cleanup()
		if err != nil {
			t.Fatalf("Cleanup() unexpected error for empty dir: %v", err)
		}
	})
}
