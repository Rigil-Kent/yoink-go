package comic

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ArchiveError struct {
	Message string
	Code    int
}

func (a ArchiveError) Error() string {
	return a.Message
}

// Archive creates a zip archive of the comic files.
//
// It takes no parameters.
// Returns an error if the operation fails.
func (c *Comic) Archive() error {

	outputPath := filepath.Join(c.LibraryPath, c.Title, c.Title+".cbz")
	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return ArchiveError{
			Message: "error creating directory",
			Code:    1,
		}
	}
	zipFile, err := os.Create(outputPath)

	if err != nil {
		return err
	}
	defer zipFile.Close()

	zwriter := zip.NewWriter(zipFile)
	defer zwriter.Close()

	sourcePath := filepath.Join(c.LibraryPath, c.Title)

	err = filepath.Walk(
		filepath.Dir(sourcePath),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return ArchiveError{
					Message: "error walking archive",
					Code:    1,
				}
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
				return nil
			}

			relPath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return ArchiveError{
					Message: "error walking archive",
					Code:    1,
				}
			}

			file, err := os.Open(path)
			if err != nil {
				return ArchiveError{
					Message: "error walking archive",
					Code:    1,
				}
			}
			defer file.Close()

			zipEntry, err := zwriter.Create(relPath)
			if err != nil {
				return ArchiveError{
					Message: "error walking archive",
					Code:    1,
				}
			}

			_, err = io.Copy(zipEntry, file)
			if err != nil {
				return ArchiveError{
					Message: "error walking archive",
					Code:    1,
				}
			}

			return nil
		},
	)

	if err != nil {
		return ArchiveError{
			Message: "error writing files to archive",
			Code:    1,
		}
	}

	log.Printf("Created archive\n: %s", outputPath)
	return nil
}
