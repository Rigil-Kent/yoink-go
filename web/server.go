package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"yoink/comic"
)

//go:embed static
var staticFiles embed.FS

type JobStatus string

const (
	StatusPending  JobStatus = "pending"
	StatusRunning  JobStatus = "running"
	StatusComplete JobStatus = "complete"
	StatusError    JobStatus = "error"
)

type Job struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Status    JobStatus `json:"status"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ComicEntry struct {
	Title        string    `json:"title"`
	CoverURL     string    `json:"cover_url"`
	FileURL      string    `json:"file_url"`
	DownloadedAt time.Time `json:"downloaded_at"`
}

type Server struct {
	libraryPath string
	jobs        map[string]*Job
	mu          sync.RWMutex
}

func NewServer(libraryPath string) *Server {
	return &Server{
		libraryPath: libraryPath,
		jobs:        make(map[string]*Job),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Embedded static assets
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Library files: covers (inline) and cbz downloads (attachment)
	mux.Handle("/covers/", http.StripPrefix("/covers/", http.FileServer(http.Dir(s.libraryPath))))
	mux.Handle("/files/", http.StripPrefix("/files/", s.downloadHandler()))

	// API
	mux.HandleFunc("/api/download", s.handleDownload)
	mux.HandleFunc("/api/comics", s.handleComics)
	mux.HandleFunc("/api/jobs", s.handleJobs)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// SPA root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, _ := staticFiles.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	return mux
}

// downloadHandler wraps the library file server to force Content-Disposition: attachment.
func (s *Server) downloadHandler() http.Handler {
	fs := http.FileServer(http.Dir(s.libraryPath))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment")
		fs.ServeHTTP(w, r)
	})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	job := &Job{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		URL:       req.URL,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	go s.runJob(job)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (s *Server) runJob(job *Job) {
	s.mu.Lock()
	job.Status = StatusRunning
	s.mu.Unlock()

	markupCh := make(chan *goquery.Document)
	imageCh := make(chan []string)

	c := comic.NewComic(job.URL, s.libraryPath, imageCh, markupCh)

	s.mu.Lock()
	job.Title = c.Title
	s.mu.Unlock()

	errs := c.Download(len(c.Filelist))
	if len(errs) > 0 {
		s.mu.Lock()
		job.Status = StatusError
		job.Error = errs[0].Error()
		s.mu.Unlock()
		return
	}

	if err := c.Archive(); err != nil {
		s.mu.Lock()
		job.Status = StatusError
		job.Error = err.Error()
		s.mu.Unlock()
		return
	}

	c.Cleanup()

	s.mu.Lock()
	job.Status = StatusComplete
	s.mu.Unlock()
}

func (s *Server) handleComics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	entries := []ComicEntry{}

	dirs, err := os.ReadDir(s.libraryPath)
	if err != nil {
		json.NewEncoder(w).Encode(entries)
		return
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		title := dir.Name()
		dirPath := filepath.Join(s.libraryPath, title)

		var coverURL, fileURL string
		var downloadedAt time.Time

		files, _ := os.ReadDir(dirPath)
		for _, f := range files {
			name := f.Name()

			if strings.HasSuffix(name, ".cbz") {
				fileURL = "/files/" + url.PathEscape(title) + "/" + url.PathEscape(name)
				if info, err := f.Info(); err == nil {
					downloadedAt = info.ModTime()
				}
			}

			// Cover kept by Cleanup: "<Title> 001.jpg"
			stripped := strings.TrimSpace(strings.TrimPrefix(name, title))
			if strings.HasPrefix(strings.ToLower(stripped), "001") {
				coverURL = "/covers/" + url.PathEscape(title) + "/" + url.PathEscape(name)
			}
		}

		if fileURL != "" {
			entries = append(entries, ComicEntry{
				Title:        title,
				CoverURL:     coverURL,
				FileURL:      fileURL,
				DownloadedAt: downloadedAt,
			})
		}
	}

	// Default: newest first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DownloadedAt.After(entries[j].DownloadedAt)
	})

	json.NewEncoder(w).Encode(entries)
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	jobs := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func Listen(addr string, libraryPath string) error {
	srv := NewServer(libraryPath)
	fmt.Printf("Yoink web server listening on %s\n", addr)
	return http.ListenAndServe(addr, srv.Handler())
}
