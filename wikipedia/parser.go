package wikipedia

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/anaskhan96/soup"
	"strings"
)

func Parse() (string, error) {
	resp, err := soup.Get("https://en.wikipedia.org/wiki/Foundations_of_Geopolitics")
	//resp, err := soup.Get("https://de.wikipedia.org/wiki/Grundlagen_der_Geopolitik")
	//resp, err := soup.Get("https://en.wikipedia.org/wiki/Ukraine")

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
	doc := soup.HTMLParse(resp)

	title := doc.Find("h1", "id", "firstHeading")
	fmt.Println("# " + title.FullText() + "\n")

	// Overview
	mwp := doc.FindStrict("div", "class", "mw-parser-output")
	mchd := mwp.Children()

	for k := range mchd {
		el := mchd[k]

		switch el.NodeValue {
		case "p", "h2", "ul":
			str, err := mdConv.ConvertString(el.HTML())
			if err != nil {
				panic(err)
			}

			fmt.Println(str + "\n")
		}
	}

	return "", nil

}
