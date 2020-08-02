package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ChampionListItem struct {
	Id        string
	Alias     string
	Name      string
	Positions []string
}

type OverviewData struct {
	Version      string
	ChampionList []ChampionListItem
	Unavailable  []string
}

type ChampionDataItem struct {
	Index    int
	Id       string
	Version  string
	Alias    string
	Name     string
	Position string
	Skills   []string
	Spells   []string
}

type ChampionItem struct {
	Version string
	Id      string
	Key     string
	Name    string
	Title   string
	Blurb   string
	Info    struct {
		Attack     int
		Defense    int
		Magic      int
		Difficulty int
	}
	Image struct {
		Full   string
		Sprite string
		Group  string
		X      int
		Y      int
		W      int
		H      int
	}
	Tags    []string
	Partype string
	Stats   struct {
		Hp                   int
		Hpperlevel           int
		Mp                   int
		Mpperlevel           int
		Movespeed            int
		Armor                int
		Armorperlevel        int
		Spellblock           int
		Spellblockperlevel   int
		Attackrange          int
		Hpregen              int
		Hpregenperlevel      int
		Mpregen              int
		Mpregenperlevel      int
		Crit                 int
		Critperlevel         int
		Attackdamage         int
		Attackdamageperlevel int
		Attackspeedperlevel  float32
		Attackspeed          float32
	}
}

type ChampionListResp struct {
	Type    string
	Format  string
	Version string
	Data    map[string]ChampionItem
}

const DataDragonUrl = "https://ddragon.leagueoflegends.com"

func getSpellName(src string) string {
	if len(src) == 0 {
		return ""
	}

	r := regexp.MustCompile("Summoner(.*)\\.png")
	result := r.FindStringSubmatch(src)
	s := strings.ToLower(result[len(result)-1])
	return s
}

func getChampionList() (*ChampionListResp, string) {
	res, err := http.Get(DataDragonUrl + "/api/versions.json")
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatal("Request lol version failed.")
	}

	body, _ := ioutil.ReadAll(res.Body)
	var versionArr []string
	_ = json.Unmarshal(body, &versionArr)
	version := versionArr[0]

	res, err = http.Get(DataDragonUrl + "/cdn/" + version + "/data/en_US/champion.json")
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatal("Request lol version failed.")
	}

	body, _ = ioutil.ReadAll(res.Body)
	var resp ChampionListResp
	_ = json.Unmarshal(body, &resp)

	fmt.Printf("🤖 Got official champion list, total %d \n", len(resp.Data))
	return &resp, version
}

func genOverview(allChampions map[string]ChampionItem, aliasList map[string]string) (*OverviewData, int) {
	res, err := http.Get("https://www.op.gg/champion/statistics")
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("[OP.GG]: request overview error: %d %s", res.StatusCode, res.Status)
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
		name := s.Find(".champion-index__champion-item__name").Text()
		alias := aliasList[name]
		var positions []string
		s.Find(".champion-index__champion-item__position > span").Each(func(i int, selection *goquery.Selection) {
			position := strings.ToLower(selection.Text())
			positions = append(positions, position)
		})
		if len(positions) > 0 {
			c := ChampionListItem{Alias: alias, Name: name, Id: allChampions[alias].Key}
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

	doc.Find(`.champion-overview__table--summonerspell > tbody`).First().Find(`img`).Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		s := getSpellName(src)
		if len(s) > 0 {
			d.Spells = append(d.Spells, s)
		}
	})

	return &d, nil
}

func worker(champ ChampionListItem, position string, index int) *ChampionDataItem {
	time.Sleep(time.Second * 1)

	alias := champ.Alias
	fmt.Printf("⌛️️ No.%d, %s @ %s\n", index, alias, position)

	d, _ := genPositionData(alias, position)
	if d != nil {
		d.Index = index
		d.Id = champ.Id
		d.Name = champ.Name
	}

	fmt.Printf("🌟 No.%d, %s @ %s\n", index, alias, position)
	return d
}

func importTask(allChampions map[string]ChampionItem, aliasList map[string]string) {
	start := time.Now()
	fmt.Println("🤖 Start...")
	d, count := genOverview(allChampions, aliasList)
	fmt.Printf("🤪 Got champions & positions, count: %d \n", count)

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan ChampionDataItem, count)

//output:
	for _, cur := range d.ChampionList {
		for _, p := range cur.Positions {
			cnt += 1

			//if cnt > 3 {
			//	break output
			//}

			if cnt%7 == 0 {
				time.Sleep(time.Second * 5)
			}

			wg.Add(1)
			go func(_cur ChampionListItem, _p string, _cnt int) {
				ch <- *worker(_cur, _p, _cnt)
				wg.Done()
			}(cur, p, cnt)
		}
	}

	wg.Wait()
	close(ch)

	outputPath := filepath.Join(".", "output", "op.gg")
	_ = os.MkdirAll(outputPath, os.ModePerm)

	failed := 0
	r := make(map[string][]ChampionDataItem)

	for champion := range ch {
		if champion.Skills != nil {
			champion.Version = d.Version
			r[champion.Alias] = append(r[champion.Alias], champion)
		}
	}

	for k, v := range r {
		file, _ := json.MarshalIndent(v, "", "  ")
		fileName := outputPath + "/" + k + ".json"
		wErr := ioutil.WriteFile(fileName, file, 0644)

		if wErr != nil {
			log.Fatal(wErr)
		}
	}

	file, _ := json.MarshalIndent(allChampions, "", "  ")
	fileName := "output/index.json"
	_ = ioutil.WriteFile(fileName, file, 0644)

	duration := time.Since(start)
	fmt.Printf("🟢 All finished, success: %d, failed: %d, took %s \n", count-failed, failed, duration)
}

func main() {
	allChampionData, _ := getChampionList()

	var championAliasList = make(map[string]string)
	for k, v := range allChampionData.Data {
		championAliasList[v.Name] = k
	}

	importTask(allChampionData.Data, championAliasList)
}
