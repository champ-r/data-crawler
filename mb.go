package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"
)

type VersionResp struct {
	UpToDateVersion string `json:"upToDateVersion"`
	GameTypes       string `json:"gameTypes"`
}

type ChampionDataRespDataItem struct {
	WinRate   float64 `json:"winRate"`
	Ratio     []int   `json:"ratio"`
	Frequency float64 `json:"frequency"`
}

type ChampionDataResp struct {
	WinRate     float64                             `json:"winRate"`
	Rank        int                                 `json:"rank"`
	BanRate     float64                             `json:"banRate"`
	Stats       map[string]float64                  `json:"stats"`
	NumGames    int                                 `json:"numGames"`
	Runes       map[string]ChampionDataRespDataItem `json:"runes"`
	Skills      map[string]ChampionDataRespDataItem `json:"skills"`
	Summoners   map[string]ChampionDataRespDataItem `json:"summoners"`
	Duration    map[string]ChampionDataRespDataItem `json:"duration"`
	NumBans     int                                 `json:"numBans"`
	Adjustments string                              `json:"adjustments"`
	Frequency   float64                             `json:"frequency"`
	NumWins     int                                 `json:"numWins"`
	Items       struct {
		Counter  map[string]ChampionDataRespDataItem   `json:"counter"`
		Order    []map[string]ChampionDataRespDataItem `json:"order"`
		Build    map[string]ChampionDataRespDataItem   `json:"build"`
		Starting map[string]ChampionDataRespDataItem   `json:"starting"`
	} `json:"items"`
}

type ScoreItem struct {
	RawItem string   `json:"RawItem"`
	Items   []string `json:"items"`
	Score   float64  `json:"score"`
}

const MurderBridgeBUrl = `https://d23wati96d2ixg.cloudfront.net`
const e = 2.71828
const generalMean = 2.5
const generalRatio = 50
const spread = 100 - generalRatio

var items *map[string]BuildItem

func getLatestVersion() (string, error) {
	url := MurderBridgeBUrl + `/save/general.json`
	body, err := MakeRequest(url)
	if err != nil {
		return "", err
	}

	var verResp VersionResp
	_ = json.Unmarshal(body, &verResp)
	return verResp.UpToDateVersion, nil
}

func scorer(winRate float64, frequency float64) float64 {
	if frequency == 0 {
		return 0
	}

	score := 1 / (1 + math.Pow(e, (spread/30)*(generalMean-frequency)))
	if frequency < 0.25 {
		score *= math.Pow(frequency, 2)
	}

	if frequency > generalMean {
		return math.Pow(frequency, 1/spread) * math.Pow(winRate, math.Pow(spread, 0.1)) * score
	}
	return winRate * score
}

func getItem(data map[string]ChampionDataRespDataItem, limit int) []ScoreItem {
	var keyScoreMap []ScoreItem
	for k, v := range data {
		item := ScoreItem{
			Score:   scorer(v.WinRate, v.Frequency),
			RawItem: k,
		}
		keyScoreMap = append(keyScoreMap, item)
	}

	sort.Slice(keyScoreMap, func(i, j int) bool {
		return keyScoreMap[i].Score > keyScoreMap[j].Score
	})

	return keyScoreMap[0:limit]
}

func makeItems(data ChampionDataResp) ([]ScoreItem, []ScoreItem) {
	starting := getItem(data.Items.Starting, 3)
	builds := getItem(data.Items.Build, 13)
	var bootIds []string

	for i, v := range starting {
		var itemSet [][2]int
		_ = json.Unmarshal([]byte(v.RawItem), &itemSet)
		for _, j := range itemSet {
			starting[i].Items = append(starting[i].Items, strconv.Itoa(j[0]))
		}
	}
	for i, v := range builds {
		if IsBoot(v.RawItem, *items) {
			bootIds = append(bootIds, v.RawItem)
			continue
		}

		builds[i].Items = append(builds[i].Items, v.RawItem)
	}

	return starting, builds
}

func getChampionData(alias string, version string) (*ChampionDataResp, error) {
	url := MurderBridgeBUrl + `/save/` + version + `/ARAM/` + alias + `.json`
	body, err := MakeRequest(url)
	if err != nil {
		return nil, err
	}

	var data ChampionDataResp
	_ = json.Unmarshal(body, &data)

	makeItems(data)
	fmt.Println(alias)

	return &data, nil
}

func ImportMB(championAliasList map[string]string) {
	ver, _ := getLatestVersion()
	items, _ = GetItemList(ver)

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan ChampionDataResp, len(championAliasList))
	for _, alias := range championAliasList {
		//if cnt > 3 {
		//	break
		//}

		if cnt%7 == 0 {
			time.Sleep(time.Second * 5)
		}

		cnt += 1
		wg.Add(1)
		go func(_alias string, _ver string, _cnt int) {
			d, err := getChampionData(_alias, _ver)
			if d != nil {
				ch <- *d
			} else {
				fmt.Println(_alias, err)
			}
			wg.Done()
		}(alias, ver, cnt)
	}
	wg.Wait()

	fmt.Printf(`done, %v`, ch)
}
