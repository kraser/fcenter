// fcenter project fcenter.go
package main

import (
	"flag"
	"logger"
	parsers "parser"

	//"fmt"
	"os"

	"os/signal"
	"priceloader"
	"strings"
	"webreader"

	goquery "github.com/PuerkitoBio/goquery"
)

const (
	SUPPLIER_CODE string = "mvideo"
	URL           string = "http://fcenter.ru"
	ENDMESSAGE    string = "ItemsLoaded"
	WORKERS       int    = 5
	WORKERSCAP    int    = 5
)

var (
	logMode string = "info"
	city    string = ""
)

func init() {
	flag.StringVar(&logMode, "lm", logMode, "режим логгирования")
	flag.StringVar(&city, "city", logMode, "город для которого разбирается прайс")

	logMode = "info"
	logger.SetLogLevel(logMode)
}

func initParser() {
	flag.Parse()
	parser := parsers.GetParser()
	parser.Options.Url = URL
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

	catalog := dom.Find("#bottomCatalog").First()
	columns := catalog.Find(".category-data")
	for i := range columns.Nodes {
		subCategoriesNodes := columns.Eq(i)
		categoryName := strings.TrimSpace(subCategoriesNodes.Find(".category-name").Text())
		logger.Info("LEVEL0:", categoryName)
		priceloader.PriceList.SetCurrentCategory(categoryName, "", 0)

		anchors := subCategoriesNodes.Find("a")
		anchors.Each(func(i int, s *goquery.Selection) {
			subCategoryName := s.Text()
			link, _ := s.Attr("href")
			logger.Info("LEVEL1:", subCategoryName, link)
			priceloader.PriceList.SetCurrentCategory(subCategoryName, link, 1)
		})

	}
	loadItems()
}

func errorHandle(e error) {
	if e != nil {
		panic(e)
	}
}

func loadItems() {
	//Подготовим каналы и балансировщик
	linksChan := make(chan string)
	quitChan := make(chan bool)
	pBalancer := new(Balancer)
	pBalancer.init(linksChan)

	//Приготовимся перехватывать сигнал останова в канал keys
	keys := make(chan os.Signal, 1)
	signal.Notify(keys, os.Interrupt)

	//Запускаем балансировщик и генератор
	go pBalancer.balance(quitChan)
	go generator(linksChan)

	logger.Info("Начинаем загрузку позиций")
	//Основной цикл программы:
	for {
		select {
		case <-keys: //пришла информация от нотификатора сигналов:
			logger.Info("CTRL-C: Ожидаю завершения активных загрузок")
			quitChan <- true //посылаем сигнал останова балансировщику

		case <-quitChan: //пришло подтверждение о завершении от балансировщика
			logger.Info("Загрузки завершены!")
			return
		}
	}
}

func generator(out chan string) {
	pPriceList := priceloader.PriceList
	for _, value := range pPriceList.Categories {
		for _, subCat := range value.Categories {
			logger.Info("  ", subCat.Name, subCat.URL)
			out <- URL + subCat.URL
		}
	}
	out <- ENDMESSAGE
}

func getItemHtml(pageUrl string) {
	parser := parsers.GetParser()
	result := webreader.DoRequest(pageUrl, parser.Options)
	fileName := "/home/robot/" + pageUrl[strings.LastIndex(pageUrl, "/")+1:]
	fileHandler, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	errorHandle(err)
	defer fileHandler.Close()
	fileHandler.Truncate(0)
	fileHandler.WriteString(result)
}
