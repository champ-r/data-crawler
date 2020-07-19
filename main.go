package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func generateOverview() {
	res, err := http.Get("https://www.op.gg/champion/statistics")
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
	doc.Find(`.champion-index__champion-list .champion-index__champion-item`).Each(func(i int, s *goquery.Selection) {
		alias := s.Find(".champion-index__champion-item__name").Text()
		var positions []string
		s.Find(".champion-index__champion-item__position > span").Each(func(i int, selection *goquery.Selection) {
			position := selection.Text()
			positions = append(positions, position)
		})
		fmt.Println(i, alias, positions)
	})
}

func importTask() {
	fmt.Println("start...")
	generateOverview()
}

func main() {
	importTask()
}
