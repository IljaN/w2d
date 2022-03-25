package wikipedia

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	tests := map[string]struct {
		in  string
		exp string
	}{
		"title_simple": {in: "<h1 id=\"firstHeading\">The Title</h1>", exp: "# The Title\n\n"},
		"title_no_id":  {in: "<h1>The Title</h1>", exp: ""},
		"titles_multi": {in: "<h1 id=\"someHeading\">Other Title</h1><h1 id=\"firstHeading\">The Title</h1>", exp: "# The Title\n\n"},
		"title_nested": {in: "<div><h1 id=\"someHeading\">Other Title</h1><span><h1 id=\"firstHeading\">The Title</h1></span></div>", exp: "# The Title\n\n"},

		"begin_at_output":             {in: "<p>ignored</p><div class=\"mw-parser-output\"><<p>expected</p>/div>", exp: "expected\n\n"},
		"subheading":                  {in: "<div class=\"mw-parser-output\"><h2>Subheading</h2><p>paragraph</p></div>", exp: "## Subheading\n\nparagraph\n\n"},
		"no_empty_subheading_end":     {in: "<div class=\"mw-parser-output\"><h2>Subheading1</h2><p>paragraph</p><h2>Subheading2</h2><h2>Subheading3</h2></div>", exp: "## Subheading1\n\nparagraph\n\n"},
		"no_empty_subheading_between": {in: "<div class=\"mw-parser-output\"><h2>Subheading1</h2><p>paragraph1</p><h2>Subheading2</h2><h2>Subheading3</h2><p>paragraph3</p></div>", exp: "## Subheading1\n\nparagraph1\n\n## Subheading3\n\nparagraph3\n\n"},

		"edit_box_removed": {in: "<div class=\"mw-parser-output\"><span class=\"mw-editsection\">editbox</span><<p>p1</p>/div>", exp: "p1\n\n"},
		"link_transformer": {in: "<div class=\"mw-parser-output\"><p>paragraph <a href=\"https://example.com\">link</a> end</div>", exp: "paragraph link end\n\n"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := NewArticleParser()
			act, err := p.Parse(io.NopCloser(strings.NewReader(tc.in)))

			assert.NoError(t, err)
			assert.Equal(t, tc.exp, act)
		})
	}
}
