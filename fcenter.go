// fcenter project fcenter.go
package main

import (
	"flag"
	"logger"
	parsers "parser"

	"fmt"
	"os"
	"strings"
	"webreader"

	//"github.com/opesun/goquery"
	goquery "github.com/PuerkitoBio/goquery"
)

const (
	supplierCode string = "mvideo"
	URL          string = "http://fcenter.ru"
)

var (
	logMode string = "info"
	city    string = ""
)

func init() {
	flag.StringVar(&logMode, "lm", logMode, "режим логгирования")
	flag.StringVar(&city, "city", logMode, "город для которого разбирается прайс")

	flag.Parse()
	logMode = "debug"
	logger.SetLogLevel(logMode)
}

func InitParser() {
	parser := parsers.GetParser()
	parser.Options.Url = URL
	parser.Options.AddHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	parser.Options.AddHeader("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")
	parser.Options.AddHeader("Cache-Control", "max-age=0")
	parser.Options.AddHeader("Connection", "keep-alive")
	parser.Options.AddHeader("Host", "fcenter.ru")
	parser.Options.AddHeader("Upgrade-Insecure-Requests", "1")
	parser.Options.AddHeader("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:61.0) Gecko/20100101 Firefox/61.0")

}

func main() {
	parser := parsers.GetParser()
	result := webreader.DoRequest(URL, parser.Options)
	fileHandler, err := os.OpenFile("/home/robot/test.html", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	errorHandle(err)
	defer fileHandler.Close()
	fileHandler.Truncate(0)
	fileHandler.WriteString(result)
	logger.Debug(len(result))

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	errorHandle(err)

	dom.Find("#bottomCatalog")
	catalog := dom.Find("#bottomCatalog").First()
	columns := catalog.Find(".category-data")
	/*
		columns.Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			band := s.Find(".category-name").Text()
			//title := s.Find("i").Text()
			fmt.Printf("Review %d: %s\n", i, band)
		})
	*/
	//.Find(".category-name").Text()

	for i := range columns.Nodes {
		subCategoriesNodes := columns.Eq(i)
		//subcategoriesNodes := goquery.NewDocumentFromNode(node)
		categoryName := subCategoriesNodes.Find(".category-name").Text()
		fmt.Println(categoryName)
		anchors := subCategoriesNodes.Find("a")
		anchors.Each(func(i int, s *goquery.Selection) {

			// For each item found, get the band and title
			band := s.Text()
			link, _ := s.Attr("href")
			//title := s.Find("i").Text()
			fmt.Println("Sub", i, band, link)
		})
		//logger.Debug(cell)
	}

	/*
		dom.Find("#main article .entry-title").Each(func(index int, item *goquery.Selection) {
			title := item.Text()
			linkTag := item.Find("a")
			link, _ := linkTag.Attr("href")
			fmt.Printf("Post #%d: %s - %s\n", index, title, link)
		})
	*/
}

func errorHandle(e error) {
	if e != nil {
		panic(e)
	}
}
