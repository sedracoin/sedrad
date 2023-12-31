// +build gofuzz

package muhash

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestFuzzCorpuses(t *testing.T) {
	t.Parallel()
	err := filepath.WalkDir("corpus", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		corpus, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		Fuzz(corpus)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}
