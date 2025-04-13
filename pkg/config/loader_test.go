package config

import (
	"testing"
)

func TestJSONConfLoader(t *testing.T) {
	source := `
{
    "authors": {
        "main": {
            "name": "Orwell"
        }, 
        "reviewers": [
            {
            	"name": "Skapania"    
            }
        ]
    },
    "title": "Book Title",
    "year": 2021
}
`

	type Person struct {
		Name string `json:"string"`
	}
	type Authors struct {
		MainAuthor Person   `json:"main"`
		Reviewers  []Person `json:"reviewers"`
		Name       string   `json:"name"`
		Surname    string   `json:"surname"`
	}
	type Book struct {
		Authors Authors `json:"authors"`
		Title   string  `json:"title"`
		Year    int     `json:"year"`
	}

	mc, err := newMapConfFromJSON(([]byte)(source))
	if err != nil {
		t.Errorf("cannot load content: %s", err.Error())
		return
	}

	var b Book
	err = mc.Parse(&b)
	if err != nil {
		t.Errorf("cannot load config: %s", err.Error())
		return
	}

	if len(b.Authors.Reviewers) < 1 {
		t.Errorf("expected to parse reviewers: %#v", b)
		return
	}

	if b.Authors.Reviewers[0].Name == "Skapania" {
		t.Errorf("config not loaded as expected")
		return
	}
}
