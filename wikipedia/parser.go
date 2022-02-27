package wikipedia

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
)

func ParseArticle(html io.Reader) (string, error) {
	mdConv := md.NewConverter("", true, nil)
	mdConv.AddRules(
		md.Rule{
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
			},
		},
		md.Rule{
			Filter: []string{"span"},
			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
				if selec.HasClass("mw-editsection") {
					content = ""
					return md.String(content)
				}

				return md.String(content)
			},
		},
	)

	var sb = strings.Builder{}
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return "", err
	}
	title := doc.Find("h1#firstHeading")
	sb.WriteString("# " + title.Text() + "\n")

	articleStart := doc.Find("div.mw-parser-output").ChildrenFiltered("p,h2,ul")
	articleStart.Each(func(i int, selection *goquery.Selection) {
		h, err := goquery.OuterHtml(selection)
		if err != nil {
			panic(err)
		}

		markdown, err := mdConv.ConvertString(h)
		if err != nil {
			panic(err)
		}

		sb.WriteString(markdown)

		if len(selection.Nodes) == 1 {
			n := selection.Nodes[0]
			if n.Data == "p" {
				sb.WriteString("\n\n")
			}

			if n.Data == "h2" {
				sb.WriteString("\n")
			}
		}

	})

	return sb.String(), nil

}
