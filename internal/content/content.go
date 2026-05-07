package content

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

const MaxMarkdownBytes = 500 * 1024

var SlugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,119}$`)

type FrontMatter struct {
	Title       string   `json:"title" yaml:"title"`
	Date        string   `json:"date" yaml:"date"`
	Draft       bool     `json:"draft" yaml:"draft"`
	Tags        []string `json:"tags" yaml:"tags"`
	Categories  []string `json:"categories" yaml:"categories"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Image       string   `json:"image,omitempty" yaml:"image,omitempty"`
	Math        bool     `json:"math,omitempty" yaml:"math,omitempty"`
}

type PostDraft struct {
	Slug        string      `json:"slug"`
	FrontMatter FrontMatter `json:"frontMatter"`
	Body        string      `json:"body"`
	Assets      []Asset     `json:"assets"`
	Large       bool        `json:"large,omitempty"`
}

type Asset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func ValidateSlug(slug string) error {
	if !SlugPattern.MatchString(slug) {
		return errors.New("slug must contain only letters, numbers, dash and underscore")
	}
	return nil
}

func (s *Service) Compose(draft PostDraft) ([]byte, error) {
	if err := ValidateSlug(draft.Slug); err != nil {
		return nil, err
	}
	if strings.TrimSpace(draft.FrontMatter.Title) == "" {
		return nil, errors.New("title is required")
	}
	if draft.FrontMatter.Date == "" {
		draft.FrontMatter.Date = time.Now().Format(time.RFC3339)
	}
	fm, err := yaml.Marshal(draft.FrontMatter)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(fm)
	out.WriteString("---\n\n")
	out.WriteString(draft.Body)
	if !strings.HasSuffix(draft.Body, "\n") {
		out.WriteString("\n")
	}
	return out.Bytes(), nil
}

func (s *Service) Parse(raw []byte) (FrontMatter, string, error) {
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return FrontMatter{Title: "Untitled"}, text, nil
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return FrontMatter{}, "", fmt.Errorf("front matter terminator not found")
	}
	end += 4
	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(text[4:end]), &fm); err != nil {
		return FrontMatter{}, "", err
	}
	body := text[end+4:]
	for strings.HasPrefix(body, "\n") {
		body = strings.TrimPrefix(body, "\n")
	}
	return fm, body, nil
}

func (s *Service) IsLargeMarkdown(body string) bool {
	return len(body) > MaxMarkdownBytes
}

func (s *Service) Diff(before, after []byte) string {
	dmp := diffmatchpatch.New()
	return dmp.DiffToDelta(dmp.DiffMain(string(before), string(after), false))
}

func ExtractMarkdownImages(markdown string) []string {
	re := regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	matches := re.FindAllStringSubmatch(markdown, -1)
	out := []string{}
	seen := map[string]bool{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value := strings.TrimSpace(match[1])
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "/") {
			continue
		}
		if !seen[value] {
			out = append(out, value)
			seen[value] = true
		}
	}
	return out
}
