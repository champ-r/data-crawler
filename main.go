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
	"strconv"
	"strings"
	"sync"
	"time"
)

type ChampionListItem struct {
	Id        string   `json:"id"`
	Alias     string   `json:"alias"`
	Name      string   `json:"name"`
	Positions []string `json:"positions"`
}

type OverviewData struct {
	Version      string             `json:"version"`
	ChampionList []ChampionListItem `json:"championlist"`
	Unavailable  []string           `json:"unavailable"`
}

type BlockItem struct {
	Id    string `json:"id"`
	Count int    `json:"count"`
}

type ItemBuildBlockItem struct {
	Type  string      `json:"type"`
	Items []BlockItem `json:"items"`
}

type ItemBuild struct {
	Title               string               `json:"title"`
	AssociatedMaps      []int                `json:"associatedMaps"`
	AssociatedChampions []int                `json:"associatedChampions"`
	Blocks              []ItemBuildBlockItem `json:"blocks"`
}

type RuneItem struct {
	Name            string `json:"name"`
	PickCount       string `json:"pickCount"`
	WinRate         string `json:"winRate"`
	PrimaryStyleId  int    `json:"primaryStyleId"`
	SubStyleId      int    `json:"subStyleId"`
	SelectedPerkIds []int  `json:"selectedPerkIds"`
}

type ChampionDataItem struct {
	Index      int         `json:"index"`
	Id         string      `json:"id"`
	Version    string      `json:"version"`
	Timestamp  int64       `json:"timestamp"`
	Alias      string      `json:"alias"`
	Name       string      `json:"name"`
	Position   string      `json:"position"`
	Skills     []string    `json:"skills"`
	Spells     []string    `json:"spells"`
	ItemBuilds []ItemBuild `json:"itemBuilds"`
	Runes      []RuneItem  `json:"runes"`
}

type ChampionItem struct {
	Version string `json:"version"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Blurb   string `json:"blurb"`
	Info    struct {
		Attack     int `json:"attack"`
		Defense    int `json:"defense"`
		Magic      int `json:"magic"`
		Difficulty int `json:"difficulty"`
	} `json:"info"`
	Image struct {
		Full   string `json:"full"`
		Sprite string `json:"sprite"`
		Group  string `json:"group"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		W      int    `json:"w"`
		H      int    `json:"h"`
	} `json:"image"`
	Tags    []string `json:"tags"`
	Partype string   `json:"partype"`
	Stats   struct {
		Hp                   int     `json:"hp"`
		Hpperlevel           int     `json:"hpperlevel"`
		Mp                   int     `json:"mp"`
		Mpperlevel           int     `json:"mpperlevel"`
		Movespeed            int     `json:"movespeed"`
		Armor                int     `json:"armor"`
		Armorperlevel        int     `json:"armorperlevel"`
		Spellblock           int     `json:"spellblock"`
		Spellblockperlevel   int     `json:"spellblockperlevel"`
		Attackrange          int     `json:"attackrange"`
		Hpregen              int     `json:"hpregen"`
		Hpregenperlevel      int     `json:"hpregenperlevel"`
		Mpregen              int     `json:"mpregen"`
		Mpregenperlevel      int     `json:"mpregenperlevel"`
		Crit                 int     `json:"crit"`
		Critperlevel         int     `json:"critperlevel"`
		Attackdamage         int     `json:"attackdamage"`
		Attackdamageperlevel int     `json:"attackdamageperlevel"`
		Attackspeedperlevel  float32 `json:"attackspeedperlevel"`
		Attackspeed          float32 `json:"attackspeed"`
	} `json:"stats"`
}

type ChampionListResp struct {
	Type    string                  `json:"type"`
	Format  string                  `json:"format"`
	Version string                  `json:"version"`
	Data    map[string]ChampionItem `json:"data"`
}

const DataDragonUrl = "https://ddragon.leagueoflegends.com"

func MatchSpellName(src string) string {
	if len(src) == 0 {
		return ""
	}

	r := regexp.MustCompile("Summoner(.*)\\.png")
	result := r.FindStringSubmatch(src)
	s := strings.ToLower(result[len(result)-1])
	return s
}

func MatchId(src string) string {
	if len(src) == 0 {
		return ""
	}

	r := regexp.MustCompile("\\/(\\d+)\\.png")
	result := r.FindStringSubmatch(src)
	s := strings.ToLower(result[len(result)-1])
	return s
}

func NoRepeatPush(el string, arr []string) []string {
	index := -1
	for idx, v := range arr {
		if v == el {
			index = idx
			break
		}
	}

	if index <= 0 {
		return append(arr, el)
	}
	return arr
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

func genPositionData(alias string, position string, id int) (*ChampionDataItem, error) {
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

	doc.Find(`.champion-overview__table--summonerspell > tbody:last-child .champion-stats__list .champion-stats__list__item span`).Each(func(_ int, selection *goquery.Selection) {
		s := selection.Text()
		d.Skills = append(d.Skills, s)
	})

	doc.Find(`.champion-overview__table--summonerspell > tbody`).First().Find(`img`).Each(func(_ int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		s := MatchSpellName(src)
		if len(s) > 0 {
			d.Spells = append(d.Spells, s)
		}
	})

	build := ItemBuild{
		Title:               "[OP.GG] " + alias + " " + position,
		AssociatedMaps:      []int{11, 12},
		AssociatedChampions: []int{id},
	}

	// item builds
	doc.Find(`.champion-overview__table:nth-child(2) .champion-overview__row--first`).Each(func(_ int, selection *goquery.Selection) {
		var block ItemBuildBlockItem
		block.Type = strings.TrimSpace(selection.Find(`th.champion-overview__sub-header`).Text())

		var itemIds []string
		selection.Find("li.champion-stats__list__item img").Each(func(i int, img *goquery.Selection) {
			src, _ := img.Attr("src")
			id := MatchId(src)
			itemIds = NoRepeatPush(id, itemIds)
		})
		selection.NextUntil(`tr.champion-overview__row--first`).Find("li.champion-stats__list__item img").Each(func(_ int, img *goquery.Selection) {
			src, _ := img.Attr("src")
			id := MatchId(src)
			itemIds = NoRepeatPush(id, itemIds)
		})

		for _, val := range itemIds {
			item := BlockItem{
				Id:    val,
				Count: 1,
			}
			block.Items = append(block.Items, item)
		}
		build.Blocks = append(build.Blocks, block)
	})

	d.ItemBuilds = append(d.ItemBuilds, build)

	// runes
	doc.Find(`[class*=ChampionKeystoneRune] tr`).Each(func(_ int, tr *goquery.Selection) {
		var runeItem RuneItem

		tr.Find(`.perk-page__item--active img`).Each(func(_ int, img *goquery.Selection) {
			src, _ := img.Attr(`src`)
			sId, _ := strconv.Atoi(MatchId(src))
			runeItem.SelectedPerkIds = append(runeItem.SelectedPerkIds, sId)
		})

		tr.Find(`.fragment__detail img.active`).Each(func(_ int, img *goquery.Selection) {
			src, _ := img.Attr(`src`)
			fId, _ := strconv.Atoi(MatchId(src))
			runeItem.SelectedPerkIds = append(runeItem.SelectedPerkIds, fId)
		})

		pIdSrc, _ := tr.Find(`.perk-page__item--mark img`).First().Attr(`src`)
		runeItem.PrimaryStyleId, _ = strconv.Atoi(MatchId(pIdSrc))

		sIdSrc, _ := tr.Find(`.perk-page__item--mark img`).Last().Attr(`src`)
		runeItem.SubStyleId, _ = strconv.Atoi(MatchId(sIdSrc))

		runeItem.PickCount = tr.Find(`.champion-overview__stats--pick .pick-ratio__text`).Next().Next().Text()
		runeItem.WinRate = tr.Find(`.champion-overview__stats--pick .win-ratio__text`).Next().Text()

		runeItem.Name = "[OP.GG] " + alias + "@" + position + " - " + runeItem.WinRate + ", " + runeItem.PickCount

		d.Runes = append(d.Runes, runeItem)
	})

	return &d, nil
}

func worker(champ ChampionListItem, position string, index int) *ChampionDataItem {
	time.Sleep(time.Second * 1)

	alias := champ.Alias
	fmt.Printf("⌛️️ No.%d, %s @ %s\n", index, alias, position)

	id, _ := strconv.Atoi(champ.Id)
	d, _ := genPositionData(alias, position, id)
	if d != nil {
		d.Index = index
		d.Id = champ.Id
		d.Name = champ.Name
	}

	fmt.Printf("🌟 No.%d, %s @ %s\n", index, alias, position)
	return d
}

func importTask(allChampions map[string]ChampionItem, aliasList map[string]string) {
	timestamp := time.Now().UTC().UnixNano() / int64(time.Millisecond)
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

			//if cnt > 8 {
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
			champion.Timestamp = timestamp
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
