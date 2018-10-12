// fcenter project fcenter.go
package main

import (
	"flag"
	//"fmt"
	log "logger"
	parsers "parser"
	"strings"

	errs "errorshandler"
	"priceloader"

	goquery "github.com/PuerkitoBio/goquery"
	/*




		"os"
		"os/signal"

		"regexp"
		"strconv"
		"webreader"


	*/)

var (
	logMode      string = "info"
	city         string = ""
	HTTP_HEADERS map[string]string
)

type ParserActions struct {
	mainParser *parsers.ParserObject
}

func (pCcustomAct ParserActions) ParseCategories(html string) {
	log.Info("PARSE_CATEGORIES")
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	errs.ErrorHandle(err)

	catalog := dom.Find("#bottomCatalog").First()
	columns := catalog.Find(".category-data")
	for i := range columns.Nodes {
		subCategoriesNodes := columns.Eq(i)
		categoryName := strings.TrimSpace(subCategoriesNodes.Find(".category-name").Text())
		log.Info("LEVEL0:", categoryName)
		priceloader.PriceList.SetCurrentCategory(categoryName, "", 0)

		anchors := subCategoriesNodes.Find("a")
		anchors.Each(func(i int, s *goquery.Selection) {
			subCategoryName := s.Text()
			link, _ := s.Attr("href")
			log.Info("LEVEL1:", subCategoryName, link)
			priceloader.PriceList.SetCurrentCategory(subCategoryName, link, 1)
		})

	}
}

func (pCcustomAct ParserActions) ParseItems() {
	log.Info("Items")
}

func (pCcustomAct ParserActions) ParserInit(parser *parsers.ParserObject) {
	pCcustomAct.mainParser = parser
	HTTP_HEADERS = map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "ru,en-US;q=0.7,en;q=0.3",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}
	parser.Options.AddHeaders(HTTP_HEADERS)
}

func (pCcustomAct ParserActions) ParserRun() {

}

func init() {
	flag.StringVar(&logMode, "lm", logMode, "режим логгирования")
	flag.StringVar(&city, "city", logMode, "город для которого разбирается прайс")

	logMode = "debug"
}

func main() {
	flag.Parse()
	log.SetLogLevel(logMode)
	log.Info("LOGLEVEL", logMode)
	log.Info("START")
	custom := &parsers.ParserOptions{
		Name:           "fcenter",
		URL:            "http://fcenter.ru",
		Loaders:        5,
		LoaderCapacity: 5,
	}
	methods := ParserActions{}
	pParser := parsers.ParserObject{
		CustomParserOptions: custom,
		CustomParserActions: methods,
	}
	pParser.Run()
	/*

				initParser()
				parser := parsers.GetParser()
				result := webreader.DoRequest(URL, parser.Options)
				if logMode == "debug" {
					fileHandler, err := os.OpenFile("/home/robot/test.html", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
					errs.ErrorHandle(err)
					defer fileHandler.Close()
					fileHandler.Truncate(0)
					fileHandler.WriteString(result)
					logger.Debug(len(result))
				}
		=======
			logger.SetLogLevel(logMode)
			logger.Info("LOGLEVEL", logMode)
			logger.Info("START")
			initParser()
			parser := parsers.GetParser()
			result, err := webreader.DoRequest(URL, parser.Options)
			errs.ErrorHandle(err)
			logger.CheckHtml(URL, result, "debug")
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(result))
			errs.ErrorHandle(err)
		>>>>>>> 81811e20fc2e73420d5a99250c48b03d16a66b7a

				dom, err := goquery.NewDocumentFromReader(strings.NewReader(result))
				errs.ErrorHandle(err)

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

	*/
	log.Info("DONE")
}

/*
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
			task := priceloader.LoadTask{Pointer: subCat, Message: "TASK"}
			out <- task
		}
	}
	endTask := priceloader.LoadTask{Pointer: nil, Message: parsers.ENDMESSAGE}
	out <- endTask
}

func getItemHtml(itemLoadTask priceloader.LoadTask) {
	parser := parsers.GetParser()
	logger.Info("Загрузка позиций из", itemLoadTask.Pointer.Name)
	var pageUrl string = itemLoadTask.Pointer.URL
	var toContinue bool = true
	for toContinue {
		logger.Info("URL:", pageUrl)
		result, err := webreader.DoRequest(pageUrl, parser.Options)
		errs.ErrorHandle(err)
		logger.CheckHtml(pageUrl, result, "debug")
		dom, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		errs.ErrorHandle(err)
		itemCells := dom.Find(".pic-table-item")
		re := regexp.MustCompile("[^\\d]")
		for i := range itemCells.Nodes {
			itemCell := itemCells.Eq(i)
			var code string = re.ReplaceAllString(strings.TrimSpace(itemCell.Find(".goods-number").First().Text()), "")
			var name string = strings.TrimSpace(itemCell.Find(".goods-name").Find("a").First().Text())
			logger.Info(code, name)
			var priceStr string = re.ReplaceAllString(strings.TrimSpace(itemCell.Find("div.do-price").First().Text()), "")
			price, err := strconv.ParseInt(priceStr, 10, 64)
			errs.ErrorHandle(err)
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
*/
