package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ChampionListItem struct {
	Alias     string   `json:"alias"`
	Positions []string `json:"positions"`
}

type OverviewData struct {
	Version      string
	ChampionList []ChampionListItem
	Unavailable  []string
}

type ChampionDataItem struct {
	Index    int      `json:"index"`
	Version  string   `json:"version"`
	Alias    string   `json:"alias"`
	Position string   `json:"position"`
	Skills   []string `json:"skills"`
}

func genOverview() (*OverviewData, int) {
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
	d := OverviewData{
		Version: verArr[len(verArr)-1],
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
			c := ChampionListItem{Alias: alias}
			c.Positions = positions
			d.ChampionList = append(d.ChampionList, c)
			count += len(positions)
		} else {
			d.Unavailable = append(d.Unavailable, alias)
		}
	})

	return &d, count
}

func genPositionData(alias string, position string) (*ChampionDataItem, error) {
	url := "https://www.op.gg/champion/" + alias + "/statistics/" + position
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	d := ChampionDataItem{
		Alias:    alias,
		Position: position,
	}
	// skills
	doc.Find(`.champion-overview__table--summonerspell > tbody:last-child .champion-stats__list .champion-stats__list__item span`).Each(func(i int, selection *goquery.Selection) {
		s := selection.Text()
		d.Skills = append(d.Skills, s)
	})

	return &d, nil
}

func worker(alias string, position string, index int) *ChampionDataItem {
	time.Sleep(time.Second * 1)

	fmt.Printf("⌛️️ No.%d, %s @ %s\n", index, alias, position)

	d, _ := genPositionData(alias, position)
	result := ChampionDataItem{
		Alias:    alias,
		Position: position,
		Index:    index,
	}

	if d != nil {
		result.Skills = d.Skills
	}

	fmt.Printf("🌟 %d, %+v\n", index, d)
	return &result
}

func importTask() {
	start := time.Now()
	fmt.Println("🤖 Start...")
	d, count := genOverview()
	fmt.Printf("🤪 Got champions & positions, count: %d \n", count)

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan ChampionDataItem, count)

	//outLoop:
	for _, cur := range d.ChampionList {
		for _, p := range cur.Positions {
			cnt += 1

			if cnt%7 == 0 {
				time.Sleep(time.Second * 5)
			}

			wg.Add(1)
			go func(_alias string, _p string, _cnt int) {
				ch <- *worker(_alias, _p, _cnt)
				wg.Done()
			}(cur.Alias, p, cnt)
		}
	}

	wg.Wait()
	close(ch)

	failed := 0
	for champion := range ch {
		flag := "🎉"
		done := champion.Skills != nil
		champion.Version = d.Version

		if done {
			file, _ := json.MarshalIndent(champion, "", " ")
			fileName := champion.Alias + "-" + champion.Position + ".json"
			wErr := ioutil.WriteFile(fileName, file, 0644)

			if wErr != nil {
				log.Fatal(wErr)
			}
		} else {
			flag = "❌"
			failed += 1
		}

		fmt.Printf("%s %+v\n", flag, champion)
	}

	duration := time.Since(start)
	fmt.Printf("🟢 All finished, success: %d, failed: %d, took %s \n", count-failed, failed, duration)
}

func main() {
	importTask()
}
