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
	"regexp"
	"strconv"
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

	logger.SetLogLevel(logMode)
	logger.Info("LOGLEVEL", logMode)
}

func initParser() {
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
	flag.Parse()
	logger.Info("START")
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
	checkCategoriesStructure()
	logger.Info("DONE")
}

func checkCategoriesStructure() {
	pPriceList := priceloader.PriceList
	for _, category := range pPriceList.Categories {
		logger.Info(category.Name)
		for _, subCat := range category.Categories {
			logger.Info("  ", subCat.Name, subCat.URL)
			logger.Info("  ITEMS:", len(subCat.Items))
			for _, item := range subCat.Items {
				logger.Info("  ITEMS:", item.Code, item.Name)
			}
		}
	}
}

func errorHandle(e error) {
	if e != nil {
		panic(e)
	}
}

func loadItems() {
	//Подготовим каналы и балансировщик
	taskChan := make(chan priceloader.LoadTask)
	quitChan := make(chan bool)
	pBalancer := new(Balancer)
	pBalancer.init(taskChan)

	//Приготовимся перехватывать сигнал останова в канал keys
	keys := make(chan os.Signal, 1)
	signal.Notify(keys, os.Interrupt)

	//Запускаем балансировщик и генератор
	go pBalancer.balance(quitChan)
	go generator(taskChan)

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

func generator(out chan priceloader.LoadTask) {
	pPriceList := priceloader.PriceList
	for _, value := range pPriceList.Categories {
		logger.Info(value.Name)
		for _, subCat := range value.Categories {
			subCat.URL = URL + subCat.URL
			logger.Info("  ", subCat.Name, subCat.URL)
			task := priceloader.LoadTask{subCat, "TASK"}
			out <- task
		}
	}
	endTask := priceloader.LoadTask{nil, ENDMESSAGE}
	out <- endTask
}

func getItemHtml(itemLoadTask priceloader.LoadTask) {
	parser := parsers.GetParser()
	logger.Info("Загрузка позиций из", itemLoadTask.Pointer.Name)
	var pageUrl string = itemLoadTask.Pointer.URL
	var toContinue bool = true
	for toContinue {
		logger.Info("URL:", pageUrl)
		result := webreader.DoRequest(pageUrl, parser.Options)
		if logMode == "debug" {
			fileName := "/home/robot/" + pageUrl[strings.LastIndex(pageUrl, "/")+1:]
			fileHandler, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
			errorHandle(err)
			defer fileHandler.Close()
			fileHandler.Truncate(0)
			fileHandler.WriteString(result)
		}
		dom, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		errorHandle(err)
		itemCells := dom.Find(".pic-table-item")
		re := regexp.MustCompile("[^\\d]")
		for i := range itemCells.Nodes {
			itemCell := itemCells.Eq(i)
			var code string = re.ReplaceAllString(strings.TrimSpace(itemCell.Find(".goods-number").First().Text()), "")
			var name string = strings.TrimSpace(itemCell.Find(".goods-name").Find("a").First().Text())
			logger.Info(code, name)
			var priceStr string = re.ReplaceAllString(strings.TrimSpace(itemCell.Find("div.do-price").First().Text()), "")
			price, err := strconv.ParseInt(priceStr, 10, 64)
			errorHandle(err)
			logger.Info(price)
			pItem := &priceloader.Item{
				Name:     name,
				Code:     code,
				Store:    "Есть",
				PriceRur: price,
			}
			priceloader.PriceList.AddItem(itemLoadTask.Pointer, pItem)
		}

		nextPageAnchor := dom.Find(".pager").Find("a.nextLink")
		logger.Debug("NEXT_PAGE_CELLS:", len(nextPageAnchor.Nodes))
		if len(nextPageAnchor.Nodes) > 0 {
			nextPageUrl, _ := nextPageAnchor.First().Attr("href")
			pageUrl = URL + nextPageUrl
		} else {
			toContinue = false
		}
	}

}
