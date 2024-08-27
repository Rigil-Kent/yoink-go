package comic

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
)

// func _downloadFile(wg *sync.WaitGroup, url string, page int, c *Comic) error {
// 	defer wg.Done()

// 	pageNumber := fmt.Sprintf("%03d", page)
// 	formattedImagePath := fmt.Sprintf("%s %s.jpg", c.Title, pageNumber)
// 	imageFilepath, _ := filepath.Abs(filepath.Join(c.LibraryPath, c.Title, formattedImagePath))

// 	if err := os.MkdirAll(
// 		filepath.Dir(imageFilepath),
// 		os.ModePerm,
// 	); err != nil {
// 		return ComicDownloadError{
// 			Message: "error creating directory",
// 			Code:    1,
// 		}
// 	}

// 	// get image data
// 	res, err := handleRequest(url)
// 	if err != nil {
// 		return ComicDownloadError{
// 			Message: "invalid request",
// 			Code:    1,
// 		}
// 	}
// 	defer res.Body.Close()

// 	var fileChannel = make(chan *os.File)

// 	go func() error {
// 		imageFile, err := os.Create(imageFilepath)
// 		if err != nil {
// 			return ComicDownloadError{
// 				Message: "error creating image file",
// 				Code:    1,
// 			}
// 		}
// 		defer imageFile.Close()

// 		fileChannel <- imageFile

// 		return nil
// 	}()

// 	println("Downloading", imageFilepath)

// 	go func(
// 		fc chan *os.File,
// 		res *http.Response,
// 	) error {
// 		buffer := make([]byte, 64*1024)

// 		defer close(fileChannel)

// 		// write image data
// 		_, err := io.CopyBuffer(<-fc, res.Body, buffer)
// 		if err != nil {
// 			return ComicDownloadError{
// 				Message: "Unable to save file contents",
// 				Code:    1,
// 			}
// 		}

// 		return nil
// 	}(fileChannel, res)

// 	return nil
// }

func downloadFile(url string, page int, c *Comic) error {
	pageNumber := fmt.Sprintf("%03d", page)
	formattedImagePath := fmt.Sprintf("%s %s.jpg", c.Title, pageNumber)
	imageFilepath, _ := filepath.Abs(filepath.Join(c.LibraryPath, c.Title, formattedImagePath))

	if err := os.MkdirAll(
		filepath.Dir(imageFilepath),
		os.ModePerm,
	); err != nil {
		return ComicDownloadError{
			Message: "error creating directory",
			Code:    1,
		}
	}

	res, err := handleRequest(url)
	if err != nil {
		return ComicDownloadError{
			Message: "invalid request",
			Code:    1,
		}
	}
	defer res.Body.Close()

	imageFile, err := os.Create(imageFilepath)
	if err != nil {
		return ComicDownloadError{
			Message: "error creating image file",
			Code:    1,
		}
	}
	defer imageFile.Close()

	written, err := io.Copy(imageFile, res.Body)
	if err != nil {
		return ComicDownloadError{
			Message: "Unable to save file contents",
			Code:    1,
		}
	}

	if written == 0 {
		return ComicDownloadError{
			Message: "Unable to save file contents",
			Code:    1,
		}
	}

	return nil
}

func handleRequest(url string) (*http.Response, error) {
	// adjust timeout and keep-alive to avoid connection timeout
	transport := &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 32,
	}

	// add cloudflare bypass
	cfTransport := cloudflarebp.AddCloudFlareByPass(transport)

	// prevents cloudflarebp from occasionally returning the wrong type
	if converted, ok := cfTransport.(*http.Transport); ok {
		transport = converted
	}

	client := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}

	// mimic generic browser
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ComicDownloadError{
			Message: "invalid request",
			Code:    1,
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return nil, ComicDownloadError{
			Message: "invalid request",
			Code:    1,
		}
	}

	if res.StatusCode != http.StatusOK {
		return nil, ComicDownloadError{
			Message: "bad response",
			Code:    1,
		}
	}

	return res, nil
}

func (c *Comic) Download(concurrency int) []error {
	// var wg sync.WaitGroup
	// wg.Add(len(c.Filelist))

	// for i, link := range c.Filelist {
	// 	go downloadFile(link, i+1, c)
	// }

	// wg.Wait()
	// return nil
	jobs := make(chan Download)
	results := make(chan error)

	for worker := 1; worker <= concurrency; worker++ {
		go workerPool(jobs, results)
	}

	for i, url := range c.Filelist {
		jobs <- Download{
			URL:   url,
			Page:  i + 1,
			Comic: c,
		}
	}

	var errors []error
	for i := 0; i < len(c.Filelist); i++ {
		err := <-results
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

type Download struct {
	URL   string
	Page  int
	Comic *Comic
}

func workerPool(jobs <-chan Download, results chan<- error) {
	for job := range jobs {
		results <- downloadFile(job.URL, job.Page, job.Comic)
	}
}

func DownloadComicImages(c *Comic, concurrency int) []error {
	jobs := make(chan Download)
	results := make(chan error)

	for worker := 1; worker <= concurrency; worker++ {
		go workerPool(jobs, results)
	}

	for i, url := range c.Filelist {
		jobs <- Download{
			URL:   url,
			Page:  i + 1,
			Comic: c,
		}
	}

	var errors []error
	for i := 0; i < len(c.Filelist); i++ {
		err := <-results
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
