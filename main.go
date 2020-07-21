package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
	"time"
)

type championListItem struct {
	alias     string
	positions []string
}

type data struct {
	version      string
	championList []championListItem
	unavailable  []string
}

type championDataItem struct {
	alias  string
	skills []string
}

func genOverview() (*data, int) {
	res, err := http.Get("https://www.op.gg/champion/statistics")
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("[op.gg]: request overview error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	verInfo := doc.Find(".champion-index__version").Text()
	verArr := strings.Split(strings.Trim(verInfo, " \n"), ` : `)
	d := data{
		version: verArr[len(verArr)-1],
	}

	count := 0
	doc.Find(`.champion-index__champion-list .champion-index__champion-item`).Each(func(i int, s *goquery.Selection) {
		alias := s.Find(".champion-index__champion-item__name").Text()
		var positions []string
		s.Find(".champion-index__champion-item__position > span").Each(func(i int, selection *goquery.Selection) {
			position := selection.Text()
			positions = append(positions, position)
		})
		if len(positions) > 0 {
			c := championListItem{alias: alias}
			c.positions = positions
			d.championList = append(d.championList, c)
			count += len(positions)
		} else {
			d.unavailable = append(d.unavailable, alias)
		}
	})

	return &d, count
}

func genPositionData(alias string, position string) *championDataItem {
	url := "https://www.op.gg/champion/" + alias + "/statistics/" + position
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("[op.gg]: %s @ %s error: %s", alias, position, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	d := championDataItem{
		alias: alias,
	}
	// skills
	doc.Find(`.champion-overview__table--summonerspell > tbody:last-child .champion-stats__list .champion-stats__list__item span`).Each(func(i int, selection *goquery.Selection) {
		s := selection.Text()
		d.skills = append(d.skills, s)
	})

	fmt.Printf("%s %s: done \n", alias, position)
	return &d
}

func importTask() {
	fmt.Println("start...")
	d, count := genOverview()
	fmt.Printf("got champions & positions, count: %d \n", count)

	chanArr := make(chan *championDataItem, count)
	cnt := 0
	for _, cur := range d.championList {
		for _, p := range cur.positions {
			cnt += 1

			go func(_alias string, _p string, _c int) {
				time.Sleep(time.Millisecond * 2000)
				chanArr <- genPositionData(_alias, _p)
			}(cur.alias, p, cnt)
		}
	}

	for el := range chanArr {
		fmt.Println(el)
	}
}

func main() {
	importTask()
}
