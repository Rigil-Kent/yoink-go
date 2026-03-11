package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"yoink/comic"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

type Options struct {
	Verbose     bool
	LibraryPath string
}

var cli = &cobra.Command{
	Use:   "yoink",
	Short: "yoink",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		library, ok := os.LookupEnv("YOINK_LIBRARY")
		if !ok {
			userHome, _ := os.UserHomeDir()
			library = filepath.Join(userHome, ".yoink")
		}

		options := Options{
			Verbose:     false,
			LibraryPath: library,
		}

		var markupChannel = make(chan *goquery.Document)
		var imageChannel = make(chan []string)

		comic := comic.NewComic(args[0], options.LibraryPath, imageChannel, markupChannel)

		fmt.Println(comic.Title)

		err := comic.Download(len(comic.Filelist))
		for e := range err {
			fmt.Println(e)
		}

		comic.Archive()
		comic.Cleanup()
	},
	Version: "1.2.1",
}

func Execute() error {
	if err := cli.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return nil
}
