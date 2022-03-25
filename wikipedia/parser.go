package wikipedia

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"regexp"
	"strings"
)

func NewArticleParser() *ArticleParser {
	return &ArticleParser{
		md: md.NewConverter("", true, nil).
			AddRules(linkRemover, editBoxRemover, newLineFixer).
			ClearAfter().
			After(afterHook),
	}
}

func (p *ArticleParser) Parse(html io.ReadCloser) (string, error) {
	var sb = strings.Builder{}
	var doc *goquery.Document
	var err error
	sb.Grow(256000)

	defer html.Close()

	doc, err = goquery.NewDocumentFromReader(html)
	if err != nil {
		return "", err
	}
	title := doc.Find("h1#firstHeading")
	sb.WriteString("# " + title.Text() + "\n\n")

	articleStart := doc.Find("div.mw-parser-output").ChildrenFiltered("h2,p,ul")
	articleStart.EachWithBreak(func(i int, selection *goquery.Selection) bool {

		if isEmptyHeading(i, articleStart.Nodes) {
			return true
		}

		var h = ""
		h, err = goquery.OuterHtml(selection)
		if err != nil {
			return false
		}

		var markdown = ""
		markdown, err = p.md.ConvertString(h)
		if err != nil {
			return false
		}

		sb.WriteString(markdown)
		return true
	})

	return sb.String(), err

}

// isEmptyHeading returns true if the current and next node in nodes relative to curIdx is a heading
func isEmptyHeading(curIdx int, nodes []*html.Node) bool {
	isLast, next := curIdx == len(nodes)-1, curIdx+1

	if !isLast && nodes[curIdx].Data == "h2" && nodes[next].Data == "h2" {
		return true
	}

	if isLast && nodes[curIdx].Data == "h2" {
		return true
	}

	return false
}

// Reduce many newline characters `\n` to at most 2 new line characters.
var multipleNewLinesRegex = regexp.MustCompile(`[\n]{2,}`)

// afterHook Remove superfluous whitespace
func afterHook(markdown string) string {
	markdown = strings.TrimLeft(markdown, "\n")
	return multipleNewLinesRegex.ReplaceAllString(markdown, "\n\n")
}

var (
	newLineFixer = md.Rule{
		Filter: []string{"h2", "p"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			if len(selec.Nodes) == 1 {
				content = strings.ReplaceAll(content, "\n", "")
				if selec.Nodes[0].Data == "p" {
					content = content + "\n\n"
					return md.String(content)
				}

				if selec.Nodes[0].Data == "h2" {
					content = "## " + content + "\n\n"
					return md.String(content)
				}
			}
			return nil
		}}

	linkRemover = md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			if sel, ok := selec.Attr("href"); ok && strings.HasPrefix(sel, "#") {
				content = ""
				return md.String(content)
			}

			if selec.HasClass("mw-editsection-visualeditor") {
				content = ""
				return md.String(content)
			}

			content = selec.Text()
			return md.String(content)
		}}

	editBoxRemover = md.Rule{
		Filter: []string{"span"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			if selec.HasClass("mw-editsection") {
				content = ""
				return md.String(content)
			}

			return md.String(content)
		}}
)

type ArticleParser struct {
	md *md.Converter
}
