package yaml

import (
	"testing"
)

func TestYamlImageUpdate(t *testing.T) {
	err := ImageUpdate("./test.yaml", "test-fantasy", "test.com/fantasy:v0.0.3")
	if err != nil {
		t.Fatalf("yaml analysis is fail%v", err)
	}
}
