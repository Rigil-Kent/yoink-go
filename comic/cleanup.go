package comic

import (
	"os"
	"path/filepath"
	"strings"
)

// Cleanup removes unnecessary files from the comic directory.
//
// It walks through the directory and deletes files with .jpg, .jpeg, or .png extensions that do not start with "001".
// No parameters.
// Returns an error if the operation fails.
func (c *Comic) Cleanup() error {
	filepath.Walk(
		filepath.Join(c.LibraryPath, c.Title),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			for _, ext := range []string{".jpg", ".jpeg", ".png"} {
				edited := strings.Replace(info.Name(), c.Title, "", 1)
				edited = strings.Trim(edited, " ")

				if !strings.HasPrefix(strings.ToLower(edited), "001") && strings.HasSuffix(info.Name(), ext) {
					return os.Remove(path)
				}
			}
			return nil
		})
	return nil
}
