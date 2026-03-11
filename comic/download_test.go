package comic

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestComicDownloadError(t *testing.T) {
	err := ComicDownloadError{Message: "download failed", Code: 1}
	if err.Error() != "download failed" {
		t.Errorf("Error() = %q, want %q", err.Error(), "download failed")
	}
}

func TestHandleRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("User-Agent") == "" {
				t.Error("expected User-Agent header to be set")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("image data"))
		}))
		defer server.Close()

		resp, err := handleRequest(server.URL)
		if err != nil {
			t.Fatalf("handleRequest() unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("handleRequest() status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := handleRequest(server.URL)
		if err == nil {
			t.Error("handleRequest() expected error for 404 response, got nil")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := handleRequest("http://invalid.localhost:0/bad")
		if err == nil {
			t.Error("handleRequest() expected error for invalid URL, got nil")
		}
	})
}

func TestDownloadFile(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fake image content"))
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		c := &Comic{
			Title:       "TestComic",
			LibraryPath: tmpDir,
		}

		err := downloadFile(server.URL+"/image.jpg", 1, c)
		if err != nil {
			t.Fatalf("downloadFile() unexpected error: %v", err)
		}

		expectedPath := filepath.Join(tmpDir, "TestComic", "TestComic 001.jpg")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", expectedPath)
		}
	})

	t.Run("formats page number with leading zeros", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fake image content"))
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		c := &Comic{
			Title:       "TestComic",
			LibraryPath: tmpDir,
		}

		err := downloadFile(server.URL+"/image.jpg", 42, c)
		if err != nil {
			t.Fatalf("downloadFile() unexpected error: %v", err)
		}

		expectedPath := filepath.Join(tmpDir, "TestComic", "TestComic 042.jpg")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", expectedPath)
		}
	})

	t.Run("server error returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		c := &Comic{
			Title:       "TestComic",
			LibraryPath: tmpDir,
		}

		err := downloadFile(server.URL+"/image.jpg", 1, c)
		if err == nil {
			t.Error("downloadFile() expected error for server error, got nil")
		}
	})

	t.Run("empty response body returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// write nothing
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		c := &Comic{
			Title:       "TestComic",
			LibraryPath: tmpDir,
		}

		err := downloadFile(server.URL+"/image.jpg", 1, c)
		if err == nil {
			t.Error("downloadFile() expected error for empty body, got nil")
		}
	})
}
