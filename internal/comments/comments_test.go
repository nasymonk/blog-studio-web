package comments

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadTwikooLikeJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	if err := os.WriteFile(path, []byte(`{"comment":[{"_id":"1","url":"/post/a/","nick":"X","comment":"Hi","created":1710000000000}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	service := NewService(path, "/comment/admin/")
	recent, err := service.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 1 || recent[0].Nick != "X" {
		t.Fatalf("unexpected recent: %#v", recent)
	}
	summary, err := service.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 1 || summary[0].Count != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func TestReadTwikooLokiJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	data := `{
		"collections": [
			{"name":"comment","data":[{"_id":"c1","url":"/post/a/","nick":"X","comment":"Hi","created":1710000000000}]},
			{"name":"counter","data":[{"url":"/post/a/","time":3}]}
		]
	}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	service := NewService(path, "/comment/admin/")
	recent, err := service.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 1 || recent[0].Comment != "Hi" {
		t.Fatalf("unexpected recent: %#v", recent)
	}
	summary, err := service.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 1 || summary[0].Count != 3 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	diagnosis := service.Diagnosis()
	if diagnosis.Status != "ok" || diagnosis.CommentCount != 1 || diagnosis.CounterCount != 1 {
		t.Fatalf("unexpected diagnosis: %#v", diagnosis)
	}
}

func TestEmptyTwikooLokiDiagnosis(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	data := `{"collections":[{"name":"comment","data":[]},{"name":"counter","data":[]}]}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	diagnosis := NewService(path, "/comment/admin/").Diagnosis()
	if diagnosis.Status != "empty" {
		t.Fatalf("unexpected diagnosis: %#v", diagnosis)
	}
}
