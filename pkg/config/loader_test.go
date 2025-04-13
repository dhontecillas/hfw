package config

import (
	"os"
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

func TestConfigMerge(t *testing.T) {
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
	if err := os.Setenv("HFWTEST_AUTHORS_MAIN_NAME", "Pepe"); err != nil {
		t.Errorf("cannot set env var: %s", err.Error())
		return
	}

	envM := newMapConfFromEnv("HFWTEST", "_")
	type Person struct {
		Name    string `json:"name"`
		Surname string `json:"surname"`
	}
	type Authors struct {
		MainAuthor Person   `json:"main"`
		Reviewers  []Person `json:"reviewers"`
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

	mc.Merge(envM)

	var b Book
	err = mc.Parse(&b)
	if err != nil {
		t.Errorf("cannot load config: %s", err.Error())
		return
	}

	if b.Authors.MainAuthor.Name != "Pepe" {
		t.Errorf("expected to have main author overriden, got: %s -> %#v\n%#v\n",
			b.Authors.MainAuthor.Name, mc.mi, b)
		return
	}
	if len(b.Authors.Reviewers) < 1 {
		t.Errorf("expected to parse reviewers: %#v", b)
		return
	}

	if b.Authors.Reviewers[0].Name != "Skapania" {
		t.Errorf("config not loaded as expected: %s\n %#v\n",
			b.Authors.Reviewers[0].Name, b)
		return
	}
}
