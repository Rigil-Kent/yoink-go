package comic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func newDocFromHTML(html string) *goquery.Document {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	return doc
}

func TestExtractTitleFromMarkup(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		url      string
		expected string
	}{
		{
			name:     "standard title with year",
			html:     `<html><head><title>Ultraman X Avengers 001 (2024)</title></head></html>`,
			expected: "Ultraman X Avengers 001",
		},
		{
			name:     "title with year and extra text",
			html:     `<html><head><title>Batman 042 (2023 Digital)</title></head></html>`,
			expected: "Batman 042",
		},
		{
			name:     "title with colon removed",
			html:     `<html><head><title>Spider-Man: No Way Home 001 (2022)</title></head></html>`,
			expected: "Spider-Man No Way Home 001",
		},
		{
			name:     "no title tag",
			html:     `<html><head></head></html>`,
			expected: "Untitled",
		},
		{
			name:     "title without year pattern",
			html:     `<html><head><title>Some Random Page</title></head></html>`,
			expected: "Untitled",
		},
		{
			name:     "empty title",
			html:     `<html><head><title></title></head></html>`,
			expected: "Untitled",
		},
		{
			name:     "title starts with # falls back to h1",
			html:     `<html><head><title>#018 (2026)</title></head><body><h1>Absolute Batman #018 (2026)</h1></body></html>`,
			expected: "Absolute Batman #018",
		},
		{
			name:     "title starts with # but h1 also starts with #, falls back to slug",
			html:     `<html><head><title>#018 (2026)</title></head><body><h1>#018 (2026)</h1></body></html>`,
			url:      "https://readallcomics.com/absolute-batman-018-2026/",
			expected: "Absolute Batman 018",
		},
		{
			name:     "title starts with # falls back to slug when no h1",
			html:     `<html><head><title>#018 (2026)</title></head></html>`,
			url:      "https://readallcomics.com/absolute-batman-018-2026/",
			expected: "Absolute Batman 018",
		},
		{
			name:     "title starts with # no h1 no url",
			html:     `<html><head><title>#018 (2026)</title></head></html>`,
			expected: "#018",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := newDocFromHTML(tt.html)
			c := Comic{Markup: doc, URL: tt.url}
			result := extractTitleFromMarkup(c)
			if result != tt.expected {
				t.Errorf("extractTitleFromMarkup() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTitleFromSlug(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "standard comic URL",
			url:      "https://readallcomics.com/absolute-batman-018-2026/",
			expected: "Absolute Batman 018",
		},
		{
			name:     "no trailing slash",
			url:      "https://readallcomics.com/absolute-batman-018-2026",
			expected: "Absolute Batman 018",
		},
		{
			name:     "no year in slug",
			url:      "https://readallcomics.com/absolute-batman-018/",
			expected: "Absolute Batman 018",
		},
		{
			name:     "single word slug",
			url:      "https://readallcomics.com/batman/",
			expected: "Batman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := titleFromSlug(tt.url)
			if result != tt.expected {
				t.Errorf("titleFromSlug() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCover(t *testing.T) {
	tests := []struct {
		name        string
		filelist    []string
		wantSuffix  string
		expectErr   bool
	}{
		{
			name:       "finds cover ending in 001.jpg",
			filelist:   []string{"https://example.com/image-002.jpg", "https://example.com/image-001.jpg", "https://example.com/image-003.jpg"},
			wantSuffix: "image-001.jpg",
		},
		{
			name:       "finds cover ending in 000.jpg",
			filelist:   []string{"https://example.com/image-000.jpg", "https://example.com/image-001.jpg"},
			wantSuffix: "image-000.jpg",
		},
		{
			name:      "returns error when no cover found",
			filelist:  []string{"https://example.com/image-002.jpg", "https://example.com/image-003.jpg"},
			expectErr: true,
		},
		{
			name:      "returns error for empty filelist",
			filelist:  []string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comic{Filelist: tt.filelist}
			cover, err := c.Cover()
			if tt.expectErr && err == nil {
				t.Error("Cover() expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Cover() unexpected error: %v", err)
			}
			if tt.wantSuffix != "" && !strings.HasSuffix(cover, tt.wantSuffix) {
				t.Errorf("Cover() = %q, want path ending in %q", cover, tt.wantSuffix)
			}
		})
	}
}
