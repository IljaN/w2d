package wikipedia

import (
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const baseDataPath = "../test/data/"

func TestArticlesSmoke(t *testing.T) {
	tests := map[string]struct {
		in  string
		exp string
	}{
		"de_warentrenner": {in: "articles/warentrenner.html", exp: "articles/warentrenner.md"},
		"de_ukraine":      {in: "articles/ukraine.html", exp: "articles/ukraine.md"},
		"en_hearth":       {in: "articles/hearth.html", exp: "articles/hearth.md"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			inp := open(path.Join(baseDataPath, tc.in), t)
			exp := string(readFile(path.Join(baseDataPath, tc.exp), t))

			defer inp.Close()

			parser := NewArticleParser()
			act, err := parser.Parse(inp)

			assert.NoError(t, err)
			assert.Equal(t, exp, act)
		})
	}

}

func readFile(path string, t *testing.T) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed read full file %s for testing %s", path, err)
	}

	return b
}

func open(path string, t *testing.T) io.ReadCloser {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file %s for testing %s", path, err)
	}

	return f
}
