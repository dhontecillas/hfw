package config

import (
    "testing"
)

func TestJSONConfLoader(t *testing.T) {
    source := ```
{
    "authors": {
        "main": {
            "name": " 
        }, 
        "reviewers": [
            {
                
            }
        ]
    },
    "title": "Book Title",
    "year": 2021
}
```

    type Book struct {
        Title string `json:"title"`
        Year  int    `json:"year"`
    }

    type Author struct {
        Name string `json:"name"`
        Surname string `json:"surname"`
    }
    

    l, err := newJSONConfLoader(source)
    if err != nil {
        t.Errorf("cannot load content: %s", err.Error())
        return
    }

    var b Book
    l.Load(
}
