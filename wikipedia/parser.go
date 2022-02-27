package wikipedia

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func Parse(articleHtml string) (string, error) {
	resp, err := http.Get("https://en.wikipedia.org/wiki/Ukraine")

	if err != nil {
		return "", err
	}

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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	title := doc.Find("h1#firstHeading")
	fmt.Println("# " + title.Text() + "\n")

	articleStart := doc.Find("div.mw-parser-output").ChildrenFiltered("p,h2,ul")
	articleStart.Each(func(i int, selection *goquery.Selection) {
		html, err := goquery.OuterHtml(selection)
		if err != nil {
			panic(err)
		}

		markdown, err := mdConv.ConvertString(html)
		if err != nil {
			panic(err)
		}

		if len(selection.Nodes) == 1 {
			n := selection.Nodes[0]
			if n.Data == "p" {
				markdown = markdown + "\n\n"
			}

			if n.Data == "h2" {
				markdown = markdown + "\n"
			}
		}

		fmt.Print(markdown)

	})

	return "", nil

}
