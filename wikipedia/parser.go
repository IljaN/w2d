package wikipedia

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
)

func NewArticleParser() *articleParser {
	return &articleParser{
		md: md.NewConverter("", true, nil).
			AddRules(linkRemover, editBoxRemover, newLineFixer).
			ClearAfter(),
	}
}

func (p *articleParser) Parse(html io.Reader) (string, error) {
	var sb = strings.Builder{}
	var doc *goquery.Document
	var err error
	sb.Grow(256000)

	doc, err = goquery.NewDocumentFromReader(html)
	if err != nil {
		return "", err
	}
	title := doc.Find("h1#firstHeading")
	sb.WriteString("# " + title.Text() + "\n\n")

	articleStart := doc.Find("div.mw-parser-output").ChildrenFiltered("p,h2,ul")
	articleStart.Each(func(i int, selection *goquery.Selection) {
		h, err := goquery.OuterHtml(selection)
		if err != nil {
			panic(err)
		}

		markdown, err := p.md.ConvertString(h)
		if err != nil {
			panic(err)
		}

		sb.WriteString(markdown)

	})

	return sb.String(), nil

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

type articleParser struct {
	md *md.Converter
}
