package search

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var jar, _ = cookiejar.New(nil)
var defaultClient = &http.Client{
	Jar: jar,
}

func GetStockValue(stockCode string) (stockValue string) {
	res, err := defaultClient.Get("http://google.com/search?q=BVMF:" + stockCode)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div").Filter(".BNeawe .iBp4i").Each(func(i int, s *goquery.Selection) {
		stockValue = strings.Split(s.Text(), " ")[0]
	})

	return stockValue
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("You must provide at least one argument.")
	}
	fmt.Println(GetStockValue(os.Args[1]))
}
