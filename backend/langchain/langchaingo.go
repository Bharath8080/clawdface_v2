package langchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"golang.org/x/net/html"
)

type FirecrawlResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Markdown string `json:"markdown"`
		Text     string `json:"text"`
		HTML     string `json:"html,omitempty"`
	} `json:"data"`
}

// LoadUniversal loads mixed inputs (file paths, URLs, raw text)
// Splits into 1000-char chunks with 200-char overlap
func LoadUniversal(inputs ...schema.Document) ([]schema.Document, error) {
	var all []schema.Document

	for _, input := range inputs {
		var docs []schema.Document
		var err error

		// Always keep filename, filetype, etc.
		meta := input.Metadata
		if meta == nil {
			meta = map[string]any{
				"source":   input.PageContent,
				"filename": "unknown",
				"filetype": "text/plain",
			}
		}

		docs = []schema.Document{input}

		if err != nil {
			return nil, fmt.Errorf("failed loading %q: %w", input.PageContent, err)
		}

		// Merge metadata into all loaded docs
		for i := range docs {
			if docs[i].Metadata == nil {
				docs[i].Metadata = make(map[string]any)
			}
			for k, v := range meta {
				docs[i].Metadata[k] = v
			}
		}

		// Split into chunks and add chunk index
		split, err := splitDocs(docs, 1000, 200)
		if err != nil {
			return nil, fmt.Errorf("failed splitting %q: %w", input.PageContent, err)
		}

		// Add chunk metadata (chunk_id, total chunks, etc.)
		for i := range split {
			if split[i].Metadata == nil {
				split[i].Metadata = make(map[string]any)
			}
			split[i].Metadata["chunk_id"] = i + 1
			split[i].Metadata["total_chunks"] = len(split)
		}

		all = append(all, split...)
	}

	return all, nil
}

// chunkSize=1000, overlap=200 (tweak as needed)
func splitDocs(docs []schema.Document, chunkSize, overlap int) ([]schema.Document, error) {
	sp := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(overlap),
	)

	var out []schema.Document
	for _, d := range docs {
		chunks, err := sp.SplitText(d.PageContent)
		if err != nil {
			return nil, err
		}
		for _, c := range chunks {
			out = append(out, schema.Document{
				PageContent: c,
				Metadata:    d.Metadata, // keep original metadata
			})
		}
	}
	return out, nil
}

func isURL(s string) bool {
	// First try parsing with scheme
	u, err := url.ParseRequestURI(s)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return u.Scheme == "http" || u.Scheme == "https"
	}

	// If no scheme, check if it looks like a domain
	// e.g. "example.com", "foo.org/path"
	domainPattern := `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(/.*)?$`
	match, _ := regexp.MatchString(domainPattern, s)
	return match
}

func LoadFromURL(uri string) ([]schema.Document, error) {
	endpoint := os.Getenv("FIRECRAWL_API_URL")
	if endpoint == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_URL not set")
	}

	payload := map[string]string{
		"url": uri,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("FIRECRAWL_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"firecrawl failed: status=%d body=%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var fcResp FirecrawlResponse
	if err := json.Unmarshal(respBody, &fcResp); err != nil {
		return nil, err
	}

	if !fcResp.Success {
		return nil, fmt.Errorf("firecrawl response unsuccessful")
	}

	content := fcResp.Data.Markdown
	if content == "" {
		content = fcResp.Data.Text
	}

	if content == "" {
		return nil, fmt.Errorf("firecrawl returned empty content")
	}

	return []schema.Document{
		{
			PageContent: content,
			Metadata: map[string]any{
				"source": uri,
			},
		},
	}, nil
}

func parseHTML(r io.Reader) []schema.Document {
	doc, err := html.Parse(r)
	if err != nil {
		return nil
	}
	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(text + " ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return []schema.Document{{PageContent: buf.String()}}
}
