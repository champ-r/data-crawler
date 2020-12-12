package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type VersionResp struct {
	UpToDateVersion string `json:"upToDateVersion"`
	GameTypes       string `json:"gameTypes"`
}

type StatItem struct {
	WinRate   float64 `json:"winRate"`
	Ratio     []int   `json:"ratio"`
	Frequency float64 `json:"frequency"`
}

type ChampionDataResp struct {
	WinRate     float64             `json:"winRate"`
	Rank        int                 `json:"rank"`
	BanRate     float64             `json:"banRate"`
	Stats       map[string]float64  `json:"stats"`
	NumGames    int                 `json:"numGames"`
	Runes       map[string]StatItem `json:"runes"`
	Skills      map[string]StatItem `json:"skills"`
	Summoners   map[string]StatItem `json:"summoners"`
	Duration    map[string]StatItem `json:"duration"`
	NumBans     int                 `json:"numBans"`
	Adjustments string              `json:"adjustments"`
	Frequency   float64             `json:"frequency"`
	NumWins     int                 `json:"numWins"`
	Items       struct {
		Counter  map[string]StatItem   `json:"counter"`
		Order    []map[string]StatItem `json:"order"`
		Build    map[string]StatItem   `json:"build"`
		Starting map[string]StatItem   `json:"starting"`
	} `json:"items"`
}

type ScoreItem struct {
	RawItem string  `json:"RawItem"`
	Score   float64 `json:"score"`
}

type SubPerkItem struct {
	Rune0 RespRuneItem `json:"perk0"`
	Rune1 RespRuneItem `json:"perk1"`
	Score float64      `json:"score"`
	Style int          `json:"style"`
}

type OptimalSubPerk struct {
	SubRunes []int   `json:"subRunes"`
	SubStyle int     `json:"subStyle"`
	SubScore float64 `json:"subScore"`
}

type PerkStyleItem struct {
	Style     int     `json:"style"`
	Score     float64 `json:"mainScore"`
	Runes     []int   `json:"runes"`
	SubStyle  int     `json:"subStyle"`
	SubScore  float64 `json:"subScore"`
	SubRunes  []int   `json:"subRunes"`
	Fragments []int   `json:"fragments"`
}

const MurderBridge = `murderbridge`
const MurderBridgeBUrl = `https://d23wati96d2ixg.cloudfront.net`
const e = 2.71828
const generalMean = 2.5
const generalRatio = float64(50)
const spread = 100 - generalRatio

var items *map[string]BuildItem
var runeLoopUp map[int]*RespRuneItem
var allRunes *[]RuneSlot

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

func getItemList(data map[string]StatItem, limit int) []ScoreItem {
	keyScoreMap := []ScoreItem{}
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

	l := len(keyScoreMap)
	if l == 0 {
		return keyScoreMap
	}
	if limit > l {
		return keyScoreMap[0:l]
	}
	return keyScoreMap[0:limit]
}

func makeBlocks(data ChampionDataResp) []ItemBuildBlockItem {
	starting := getItemList(data.Items.Starting, 3)
	builds := getItemList(data.Items.Build, 13)

	var startingItems []string
	var buildItems []string
	var bootIds []string

	for _, v := range starting {
		var itemSet [][2]int
		_ = json.Unmarshal([]byte(v.RawItem), &itemSet)
		for _, j := range itemSet {
			startingItems = NoRepeatPush(strconv.Itoa(j[0]), startingItems)
		}
	}
	// wards
	for _, id := range WardItems {
		startingItems = NoRepeatPush(id, startingItems)
	}
	// trinkets
	for _, id := range TrinketItems {
		startingItems = NoRepeatPush(id, startingItems)
	}

	for _, v := range builds {
		if IsBoot(v.RawItem, *items) {
			bootIds = append(bootIds, v.RawItem)
			continue
		}

		buildItems = NoRepeatPush(v.RawItem, buildItems)
	}

	startingBlocks := MakeBuildBlock(startingItems, `Starter Items`)
	buildBlocks := MakeBuildBlock(buildItems, `Recommended Builds`)
	bootBlocks := MakeBuildBlock(bootIds, `Boots`)
	consumableItems := MakeBuildBlock(ConsumableItems, `Consumable Items`)

	items := []ItemBuildBlockItem{
		startingBlocks,
		buildBlocks,
		bootBlocks,
		consumableItems,
	}
	return items
}

func generateOptimalSubPerks(runes map[string]StatItem) []SubPerkItem {
	var optimalSubPerks []SubPerkItem

	for _, i := range *allRunes {
		var targetRunes []*RespRuneItem
		for _, r := range runeLoopUp {
			if r.Style == i.Id && r.Slot != 0 {
				targetRunes = append(targetRunes, r)
			}
		}

		var bestScore float64
		var bestRunes SubPerkItem

		for _, r1 := range targetRunes {
			for _, r2 := range targetRunes {
				if r1.Slot != r2.Slot {
					score := scorer(runes[strconv.Itoa(r1.Id)].WinRate, runes[strconv.Itoa(r1.Id)].Frequency) + scorer(runes[strconv.Itoa(r2.Id)].WinRate, runes[strconv.Itoa(r2.Id)].Frequency)

					if score > bestScore {
						bestScore = score
						bestRunes = SubPerkItem{
							Rune0: *r1,
							Rune1: *r2,
							Score: score,
							Style: i.Id,
						}
					}
				}
			}
		}

		optimalSubPerks = append(optimalSubPerks, bestRunes)
	}

	sort.Slice(optimalSubPerks, func(i, j int) bool {
		return optimalSubPerks[i].Score > optimalSubPerks[j].Score
	})

	return optimalSubPerks
}

func generateOptimalPerks(runes map[string]StatItem) []PerkStyleItem {
	var bestScore float64
	var result []PerkStyleItem
	scoreMap := make(map[int]float64)

	var fragments []int
	for _, ids := range Fragments {
		sort.Slice(ids, func(i, j int) bool {
			iid := strconv.Itoa(ids[i])
			jid := strconv.Itoa(ids[j])
			iScore := scorer(runes[iid].WinRate, runes[iid].Frequency)
			jScore := scorer(runes[jid].WinRate, runes[jid].Frequency)

			scoreMap[ids[i]] = iScore
			scoreMap[ids[j]] = jScore

			return iScore > jScore
		})
		fragments = append(fragments, ids[0])
	}

	for _, primaryRuneSlot := range *allRunes {
		var totalScore float64
		var runeSet []int

		for sIdx, slot := range primaryRuneSlot.Slots {
			sort.Slice(slot.Runes, func(i, j int) bool {
				aId := strconv.Itoa(slot.Runes[i].Id)
				bId := strconv.Itoa(slot.Runes[j].Id)
				a := runes[aId]
				b := runes[bId]
				aScore := scorer(a.WinRate, a.Frequency)
				bScore := scorer(b.WinRate, b.Frequency)
				scoreMap[slot.Runes[i].Id] = aScore
				scoreMap[slot.Runes[j].Id] = bScore

				return aScore > bScore
			})
			runeSet = append(runeSet, slot.Runes[0].Id)

			rId := slot.Runes[0].Id
			if sIdx == 0 {
				totalScore += 3 * scoreMap[rId]
			} else {
				totalScore += scoreMap[rId]
			}
		}

		if totalScore > bestScore {
			bestScore = totalScore
		}

		primaryPerk := PerkStyleItem{
			Style: primaryRuneSlot.Id,
			Score: totalScore,
			Runes: runeSet,
		}

		subPerks := generateOptimalSubPerks(runes)

		var bestSubPerks []OptimalSubPerk
		for _, s := range subPerks {
			if s.Style == primaryRuneSlot.Id {
				continue
			}

			if len(bestSubPerks) < 2 {
				bestSubPerks = append(bestSubPerks, OptimalSubPerk{
					SubStyle: s.Style,
					SubScore: s.Score,
					SubRunes: []int{s.Rune0.Id, s.Rune1.Id},
				})
			}
		}

		for _, s := range bestSubPerks {
			result = append(result, PerkStyleItem{
				Style:     primaryPerk.Style,
				Score:     primaryPerk.Score,
				Runes:     primaryPerk.Runes,
				SubStyle:  s.SubStyle,
				SubScore:  s.SubScore,
				SubRunes:  s.SubRunes,
				Fragments: fragments,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result
}

func genChampionData(champion ChampionItem, version string, timestamp int64) (*ChampionDataItem, error) {
	url := MurderBridgeBUrl + `/save/` + version + `/ARAM/` + champion.Id + `.json`
	body, err := MakeRequest(url)
	if err != nil {
		return nil, err
	}

	result := ChampionDataItem{
		Id:        champion.Id,
		Version:   version,
		Alias:     champion.Id,
		Name:      champion.Name,
		Timestamp: timestamp,
	}
	var data ChampionDataResp
	_ = json.Unmarshal(body, &data)
	key, _ := strconv.Atoi(champion.Key)

	build := ItemBuild{
		Title:               `[MB] ` + champion.Id,
		AssociatedMaps:      []int{12},
		AssociatedChampions: []int{key},
		Map:                 "any",
		Mode:                "any",
		PreferredItemSlots:  []string{},
		Sortrank:            1,
		StartedFrom:         "blank",
		Type:                "custom",
		Blocks:              makeBlocks(data),
	}
	result.ItemBuilds = append(result.ItemBuilds, build)

	optimalRunes := generateOptimalPerks(data.Runes)
	for _, r := range optimalRunes {
		item := RuneItem{
			Alias:          champion.Id,
			Name:           `[MB] ` + champion.Name,
			Position:       ``,
			PrimaryStyleId: r.Style,
			SubStyleId:     r.SubStyle,
			Score:          r.Score + r.SubScore,
		}
		for _, i := range r.Runes {
			item.SelectedPerkIds = append(item.SelectedPerkIds, i)
		}
		for _, i := range r.SubRunes {
			item.SelectedPerkIds = append(item.SelectedPerkIds, i)
		}
		for _, i := range r.Fragments {
			item.SelectedPerkIds = append(item.SelectedPerkIds, i)
		}
		result.Runes = append(result.Runes, item)
	}

	fmt.Printf("🤪 [MB] %s: Fetched data. \n", result.Alias)
	return &result, nil
}

func ImportMB(championAliasList map[string]ChampionItem, timestamp int64) string {
	start := time.Now()
	fmt.Println("🌉 [MB]: Start...")

	ver, _ := getLatestVersion()
	items, _ = GetItemList(ver)
	runeLoopUp, allRunes, _ = GetRunesReforged(ver)

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan ChampionDataItem, len(championAliasList))
	for _, champion := range championAliasList {
		//if cnt > 3 {
		//	break
		//}
		if cnt > 0 && cnt%7 == 0 {
			fmt.Println(`🌉 Take a break...`)
			time.Sleep(time.Second * 5)
		}

		cnt += 1
		wg.Add(1)
		go func(_champion ChampionItem, _ver string, _cnt int, _timestamp int64) {
			d, err := genChampionData(_champion, _ver, timestamp)
			if d != nil {
				ch <- *d
			} else {
				fmt.Println(_champion.Id, err)
			}
			wg.Done()
		}(champion, ver, cnt, timestamp)
	}
	wg.Wait()
	close(ch)

	outputPath := filepath.Join(".", "output", MurderBridge)
	_ = os.MkdirAll(outputPath, os.ModePerm)

	for data := range ch {
		fileName := outputPath + "/" + data.Alias + ".json"
		content := []ChampionDataItem{data}
		_ = SaveJSON(fileName, content)
	}
	pkg, _ := GenPkgInfo("tpl/package.json", PkgInfo{
		Timestamp:       timestamp,
		SourceVersion:   ver,
		OfficialVersion: ver,
		PkgName:         MurderBridge,
	})
	_ = ioutil.WriteFile("output/"+MurderBridge+"/package.json", []byte(pkg), 0644)

	duration := time.Since(start)

	return fmt.Sprintf("🟢 [MB] Finished. Took %s.", duration)
}
