// fcenter project fcenter.go
package main

import (
	"flag"
	"logger"
	parsers "parser"

	"fmt"
	"os"
	"priceloader"
	"strings"
	"webreader"

	//"github.com/opesun/goquery"
	goquery "github.com/PuerkitoBio/goquery"
)

const (
	SUPPLIER_CODE string = "mvideo"
	URL           string = "http://fcenter.ru"
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

func initParser() {
	parser := parsers.GetParser()
	//options := parser.Options
	//options.Url = URL
	parser.Options.AddHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	parser.Options.AddHeader("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")
	parser.Options.AddHeader("Cache-Control", "max-age=0")
	parser.Options.AddHeader("Connection", "keep-alive")
	parser.Options.AddHeader("Host", "fcenter.ru")
	parser.Options.AddHeader("Upgrade-Insecure-Requests", "1")
	parser.Options.AddHeader("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:61.0) Gecko/20100101 Firefox/61.0")
	priceloader.PriceList.PriceList(SUPPLIER_CODE)

}

func main() {

	initParser()
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
	for i := range columns.Nodes {
		subCategoriesNodes := columns.Eq(i)
		categoryName := strings.TrimSpace(subCategoriesNodes.Find(".category-name").Text())
		fmt.Println("LEVEL0: ", categoryName)
		category := priceloader.PriceList.SetCurrentCategory(categoryName, 0)
		fmt.Println("CREATED0", *category)

		anchors := subCategoriesNodes.Find("a")
		anchors.Each(func(i int, s *goquery.Selection) {
			subCategoryName := s.Text()
			link, _ := s.Attr("href")
			fmt.Println("LEVEL1", subCategoryName, link)
			subCategory := priceloader.PriceList.SetCurrentCategory(subCategoryName, 1)
			fmt.Println("CREATED1", *subCategory)
			//fmt.Println(category)

		})

	}
	checkCategoriesStructure()
}

func checkCategoriesStructure() {
	pPriceList := priceloader.PriceList
	for name, value := range pPriceList.Categories {
		fmt.Println(name, len(value.Categories))
	}
}

func errorHandle(e error) {
	if e != nil {
		panic(e)
	}
}
