package comic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseBatcaveBizImageLinks(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expectCount int
		expectErr   bool
		expectURLs  []string
	}{
		{
			name: "extracts images from __DATA__",
			html: `<html><body><script>
				var __DATA__ = {"images":["https://cdn.batcave.biz/img/001.jpg","https://cdn.batcave.biz/img/002.jpg"]};
			</script></body></html>`,
			expectCount: 2,
			expectErr:   false,
			expectURLs:  []string{"https://cdn.batcave.biz/img/001.jpg", "https://cdn.batcave.biz/img/002.jpg"},
		},
		{
			name: "unescapes forward slashes in URLs",
			html: `<html><body><script>
				var __DATA__ = {"images":["https:\/\/cdn.batcave.biz\/img\/001.jpg"]};
			</script></body></html>`,
			expectCount: 1,
			expectErr:   false,
			expectURLs:  []string{"https://cdn.batcave.biz/img/001.jpg"},
		},
		{
			name: "extracts images with spaces around colon and bracket",
			html: `<html><body><script>
				var __DATA__ = {"images" : [ "https://cdn.batcave.biz/img/001.jpg" ]};
			</script></body></html>`,
			expectCount: 1,
			expectErr:   false,
			expectURLs:  []string{"https://cdn.batcave.biz/img/001.jpg"},
		},
		{
			name: "no __DATA__ script",
			html: `<html><body><script>
				var foo = "bar";
			</script></body></html>`,
			expectCount: 0,
			expectErr:   true,
		},
		{
			name: "__DATA__ present but no images key",
			html: `<html><body><script>
				var __DATA__ = {"title":"Nightwing"};
			</script></body></html>`,
			expectCount: 0,
			expectErr:   true,
		},
		{
			name:        "no script tags",
			html:        `<html><body><p>nothing here</p></body></html>`,
			expectCount: 0,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			ch := make(chan []string, 1)

			links, err := ParseBatcaveBizImageLinks(doc, ch)

			if tt.expectErr && err == nil {
				t.Error("ParseBatcaveBizImageLinks() expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("ParseBatcaveBizImageLinks() unexpected error: %v", err)
			}
			if len(links) != tt.expectCount {
				t.Errorf("ParseBatcaveBizImageLinks() returned %d links, want %d", len(links), tt.expectCount)
			}
			for i, expected := range tt.expectURLs {
				if i >= len(links) {
					t.Errorf("missing link at index %d: want %q", i, expected)
					continue
				}
				if links[i] != expected {
					t.Errorf("links[%d] = %q, want %q", i, links[i], expected)
				}
			}

			channelLinks := <-ch
			if len(channelLinks) != tt.expectCount {
				t.Errorf("channel received %d links, want %d", len(channelLinks), tt.expectCount)
			}
		})
	}
}

func TestImageParseError(t *testing.T) {
	err := ImageParseError{Message: "test error", Code: 1}
	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}

func TestParseImageLinks(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expectCount int
		expectErr   bool
	}{
		{
			name: "extracts blogspot images",
			html: `<html><body>
				<img src="https://bp.blogspot.com/page-001.jpg" />
				<img src="https://bp.blogspot.com/page-002.jpg" />
			</body></html>`,
			expectCount: 2,
			expectErr:   false,
		},
		{
			name: "extracts blogger googleusercontent images",
			html: `<html><body>
				<img src="https://blogger.googleusercontent.com/page-001.jpg" />
			</body></html>`,
			expectCount: 1,
			expectErr:   false,
		},
		{
			name: "extracts covers images",
			html: `<html><body>
				<img src="https://example.com/covers/cover-001.jpg" />
			</body></html>`,
			expectCount: 1,
			expectErr:   false,
		},
		{
			name: "excludes logo images",
			html: `<html><body>
				<img src="https://bp.blogspot.com/logo-site.jpg" />
				<img src="https://bp.blogspot.com/page-001.jpg" />
			</body></html>`,
			expectCount: 1,
			expectErr:   false,
		},
		{
			name: "excludes non-matching images",
			html: `<html><body>
				<img src="https://other-site.com/image.jpg" />
				<img src="https://cdn.example.com/banner.png" />
			</body></html>`,
			expectCount: 0,
			expectErr:   true,
		},
		{
			name:        "no images at all",
			html:        `<html><body><p>No images here</p></body></html>`,
			expectCount: 0,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			ch := make(chan []string, 1)

			links, err := ParseImageLinks(doc, ch)

			if tt.expectErr && err == nil {
				t.Error("ParseImageLinks() expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("ParseImageLinks() unexpected error: %v", err)
			}
			if len(links) != tt.expectCount {
				t.Errorf("ParseImageLinks() returned %d links, want %d", len(links), tt.expectCount)
			}

			// Verify the channel also received the links
			channelLinks := <-ch
			if len(channelLinks) != tt.expectCount {
				t.Errorf("channel received %d links, want %d", len(channelLinks), tt.expectCount)
			}
		})
	}
}
