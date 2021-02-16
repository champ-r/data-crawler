package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
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
	ChampionList []ChampionListItem `json:"championList"`
	Unavailable  []string           `json:"unavailable"`
}

const OPGG = `op.gg`
const OPGGUrl = `https://www.op.gg/champion`

func genOverview(allChampions map[string]ChampionItem, aliasList map[string]string) (*OverviewData, int) {
	doc, err := ParseHTML(OPGGUrl + `/statistics`)
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

func genPositionData(alias string, position string, id int, version string) (*ChampionDataItem, error) {
	pos := position
	if position == `middle` {
		pos = `mid`
	} else if position == `bottom` {
		pos = `bot`
	}
	url := OPGGUrl + "/" + alias + "/statistics/" + pos

	doc, err := ParseHTML(url)
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
		Title:               "[OP.GG] " + alias + " @ " + position + ` ` + version,
		AssociatedMaps:      []int{11, 12},
		AssociatedChampions: []int{id},
		Map:                 "any",
		Mode:                "any",
		PreferredItemSlots:  []string{},
		Sortrank:            1,
		StartedFrom:         "blank",
		Type:                "custom",
	}

	// item builds
	doc.Find(`.champion-overview__table:nth-child(2) .champion-overview__row--first`).Each(func(blockIdx int, selection *goquery.Selection) {
		blockType := strings.TrimSpace(selection.Find(`th.champion-overview__sub-header`).Text())
		isRecommendedBuild := strings.Contains(strings.ToLower(blockType), `recommended builds`)

		if isRecommendedBuild {
			var firstBlock ItemBuildBlockItem

			pickCnt := strings.ReplaceAll(selection.Find(`td.champion-overview__stats--pick.champion-overview__border > span`).Text(), `,`, ``)
			winRate := selection.Find(`td.champion-overview__stats--win.champion-overview__border > strong`).Text()
			firstBlock.Type = `Recommended Builds: Pick ` + pickCnt + `, Win Rate` + winRate
			selection.Find("li.champion-stats__list__item img").Each(func(i int, img *goquery.Selection) {
				src, _ := img.Attr("src")
				id := MatchId(src)
				firstBlock.Items = append(firstBlock.Items, BlockItem{
					Id: id,
					Count: 1,
				})
			})

			build.Blocks = append(build.Blocks, firstBlock)

			selection.NextUntil(`tr.champion-overview__row--first`).Each(func(trIdx int, tr *goquery.Selection) {
				pickCnt := strings.ReplaceAll(tr.Find(`td.champion-overview__stats--pick.champion-overview__border > span`).Text(), `,`, ``)
				winRate := tr.Find(`td.champion-overview__stats--win.champion-overview__border > strong`).Text()

				var block ItemBuildBlockItem
				block.Type = `Recommended Builds: Pick ` + pickCnt + `, Win Rate` + winRate

				tr.Find("li.champion-stats__list__item img").Each(func(i int, img *goquery.Selection) {
					src, _ := img.Attr("src")
					id := MatchId(src)
					block.Items = append(block.Items, BlockItem{
						Id: id,
						Count: 1,
					})
				})

				build.Blocks = append(build.Blocks, block)
			})

			return
		}

		var block ItemBuildBlockItem
		block.Type = blockType

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

		// starter items
		if blockIdx == 0 {
			// wards
			for _, id := range WardItems {
				itemIds = NoRepeatPush(id, itemIds)
			}

			// trinkets
			for _, id := range TrinketItems {
				itemIds = NoRepeatPush(id, itemIds)
			}
		}

		for _, val := range itemIds {
			item := BlockItem{
				Id:    val,
				Count: 1,
			}
			block.Items = append(block.Items, item)
		}
		build.Blocks = append(build.Blocks, block)
	})

	// consumables
	b := ItemBuildBlockItem{
		Type: "Consumables",
	}
	for _, id := range ConsumableItems {
		item := BlockItem{
			Id:    id,
			Count: 1,
		}
		b.Items = append(b.Items, item)
	}
	build.Blocks = append(build.Blocks, b)

	d.ItemBuilds = append(d.ItemBuilds, build)

	// runes
	doc.Find(`[class*=ChampionKeystoneRune] tr`).Each(func(_ int, tr *goquery.Selection) {
		var runeItem RuneItem
		runeItem.Alias = alias
		runeItem.Position = position

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

		pickCount := tr.Find(`.champion-overview__stats--pick .pick-ratio__text`).Next().Next().Text()
		runeItem.PickCount, _ = strconv.Atoi(strings.ReplaceAll(pickCount, `,`, ``))
		runeItem.WinRate = tr.Find(`.champion-overview__stats--pick .win-ratio__text`).Next().Text()

		runeItem.Name = "[OP.GG] " + alias + "@" + position + " - " + runeItem.WinRate + ", " + fmt.Sprint(runeItem.PickCount)

		d.Runes = append(d.Runes, runeItem)
	})

	sort.Slice(d.Runes, func(i, j int) bool {
		return d.Runes[i].PickCount > d.Runes[j].PickCount
	})

	return &d, nil
}

func worker(champ ChampionListItem, position string, index int, version string) *ChampionDataItem {
	time.Sleep(time.Second * 1)

	alias := champ.Alias
	fmt.Printf("⌛ [OP.GG]️️ No.%d, %s @ %s\n", index, alias, position)

	id, _ := strconv.Atoi(champ.Id)
	d, _ := genPositionData(alias, position, id, version)
	if d != nil {
		d.Index = index
		d.Id = champ.Id
		d.Name = champ.Name
	}

	fmt.Printf("🌟 [OP.GG] No.%d, %s @ %s\n", index, alias, position)
	return d
}

func ImportOPGG(allChampions map[string]ChampionItem, aliasList map[string]string, officialVer string, timestamp int64, debug bool) string {
	start := time.Now()
	fmt.Println("🤖 [OP.GG] Start...")

	d, count := genOverview(allChampions, aliasList)
	fmt.Printf("🤪 [OP.GG] Got champions & positions, count: %d \n", count)

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan ChampionDataItem, count)

	listLoop:
	for _, cur := range d.ChampionList {
		for _, p := range cur.Positions {
			cnt += 1

			if cnt%7 == 0 {
				if debug {
					wg.Done()
					break listLoop
				}
				time.Sleep(time.Second * 5)
			}

			wg.Add(1)
			go func(_cur ChampionListItem, _p string, _cnt int, _ver string) {
				ch <- *worker(_cur, _p, _cnt, _ver)
				wg.Done()
			}(cur, p, cnt, officialVer)
		}
	}

	wg.Wait()
	close(ch)

	outputPath := filepath.Join(".", "output", OPGG)
	_ = os.MkdirAll(outputPath, os.ModePerm)

	failed := 0
	r := make(map[string][]ChampionDataItem)

	for champion := range ch {
		if champion.Skills != nil {
			champion.Timestamp = timestamp
			champion.Version = d.Version
			champion.OfficialVersion = officialVer
			r[champion.Alias] = append(r[champion.Alias], champion)
		}
	}

	for k, v := range r {
		fileName := outputPath + "/" + k + ".json"
		_ = SaveJSON(fileName, v)
	}

	_ = SaveJSON("output/index.json", allChampions)

	pkg, _ := GenPkgInfo("tpl/package.json", PkgInfo{
		Timestamp:       timestamp,
		SourceVersion:   d.Version,
		OfficialVersion: officialVer,
		PkgName:         OPGG,
	})
	_ = ioutil.WriteFile("output/"+OPGG+"/package.json", []byte(pkg), 0644)

	duration := time.Since(start)
	return fmt.Sprintf("🟢 [OP.GG] All finished, success: %d, failed: %d, took %s", cnt-failed, failed, duration)
}
