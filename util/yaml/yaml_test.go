package yaml

import (
	"testing"
)

func TestYamlImageUpdate(t *testing.T) {
	err := ImageUpdate("./test.yaml", "test-fantasy-cms", "test.com/fantasy-cms:v0.0.2")
	if err != nil {
		t.Fatalf("yaml analysis is fail%v", err)
	}
}
