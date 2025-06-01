package config

import (
	"encoding/json"
	"testing"
)

func TestInitSendGrid(t *testing.T) {
	example := json.RawMessage(`
{
	"key": "foo",
	"senderemail": "m.name@example.com",
	"sendername": "MyName"
}`)

	conf, err := configSendGrid(example)
	if err != nil {
		t.Errorf("Expected unexpected error: %s", err)
		return
	}
	if len(conf.Key) == 0 {
		t.Errorf("Expected conf.Key to be empty")
		return
	}
}
