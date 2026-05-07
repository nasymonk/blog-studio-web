package content

import "testing"

func TestParseComposeFrontMatter(t *testing.T) {
	service := NewService()
	raw, err := service.Compose(PostDraft{
		Slug: "hello-world",
		FrontMatter: FrontMatter{
			Title:      "Hello",
			Date:       "2026-05-01",
			Draft:      false,
			Tags:       []string{"go", "hugo"},
			Categories: []string{"demo"},
		},
		Body: "Body text.",
	})
	if err != nil {
		t.Fatal(err)
	}
	fm, body, err := service.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if fm.Title != "Hello" || len(fm.Tags) != 2 || body != "Body text.\n" {
		t.Fatalf("unexpected parse result: %#v body=%q", fm, body)
	}
}

func TestValidateSlug(t *testing.T) {
	for _, slug := range []string{"hello", "hello-world_2026"} {
		if err := ValidateSlug(slug); err != nil {
			t.Fatalf("valid slug failed: %s", slug)
		}
	}
	for _, slug := range []string{"../secret", "/root", "hello world", ""} {
		if err := ValidateSlug(slug); err == nil {
			t.Fatalf("invalid slug passed: %s", slug)
		}
	}
}

func TestExtractMarkdownImages(t *testing.T) {
	images := ExtractMarkdownImages("![a](cover.jpg)\n![b](https://example.com/x.png)\n![c](assets/a.png)")
	if len(images) != 2 || images[0] != "cover.jpg" || images[1] != "assets/a.png" {
		t.Fatalf("unexpected images: %#v", images)
	}
}
