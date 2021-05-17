package lolalytics

import (
	"data-crawler/pkg/common"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var cidReg = regexp.MustCompile("&cid=\\d+?&")
var laneReg = regexp.MustCompile("&lane=[a-zA-Z]+?&")
var epReg = regexp.MustCompile("ep=.*?region=all")
var patchReg = regexp.MustCompile("&patch=((\\d+\\.)+\\d+?)&")

const ApiUrl = "https://apix1.op.lol"

func makeQuery(query string) func(string, string) string {
	oldQ := query
	return func(cid string, lane string) string {
		q := cidReg.ReplaceAllString(oldQ, "&cid="+cid+"&")
		q = laneReg.ReplaceAllString(q, "&lane="+lane+"&")
		return q
	}
}

func getSourceVersion(q string) string {
	m := patchReg.FindAllStringSubmatch(q, 1)
	return m[0][1]
}

func getTierList(q string) (ITierList, error) {
	var data ITierList

	// list sort by name
	body, err := common.MakeRequest(ApiUrl + "/tierlist/7/?" + q)
	if err != nil {
		return data, err
	}

	_ = json.Unmarshal(body, &data)
	return data, nil
}

func getChampionById(id string, championAliasList map[string]common.ChampionItem) common.ChampionItem {
	var ret common.ChampionItem
	for _, champ := range championAliasList {
		if id != champ.Key {
			continue
		}

		ret = champ
		break
	}

	return ret
}

func makeBlock(title string, set []int) common.ItemBuildBlockItem {
	blockItem := common.ItemBuildBlockItem{
		Type: title,
	}

	for _, itemId := range set {
		item := common.BlockItem{
			Id:    strconv.Itoa(itemId),
			Count: 1,
		}
		blockItem.Items = append(blockItem.Items, item)
	}

	return blockItem
}

func extractItemIds(items []IItemN) []int {
	var ids []int
	for _, i := range items {
		ids = append(ids, i.ID)
	}

	return ids
}

func makeBuildBlocksFromSet(data IItems) []common.ItemBuildBlockItem {
	var blocks []common.ItemBuildBlockItem
	startingTitle := "Starting items, win rate " + fmt.Sprintf("%.2f%", data.Start.Wr)
	startingBlock := makeBlock(startingTitle, data.Start.Set)
	blocks = append(blocks, startingBlock)

	coreTitle := "Core items, win rate " + fmt.Sprintf("%.2f%", data.Core.Wr)
	coreBlock := makeBlock(coreTitle, data.Core.Set)
	blocks = append(blocks, coreBlock)

	item4Ids := extractItemIds(data.Item4)
	item4Block := makeBlock("Item 4", item4Ids)
	blocks = append(blocks, item4Block)

	item5Ids := extractItemIds(data.Item5)
	item5Block := makeBlock("Item 5", item5Ids)
	blocks = append(blocks, item5Block)

	item6Ids := extractItemIds(data.Item6)
	item6Block := makeBlock("Item 6", item6Ids)
	blocks = append(blocks, item6Block)

	return blocks
}

func makeBuild(champion common.ChampionItem, query string, sourceVersion string, timestamp int64, cnt int) (*common.ChampionDataItem, error) {
	body, err := common.MakeRequest(ApiUrl + "/mega?" + query)

	if err != nil {
		fmt.Println("[lolalytics] Fetch champion data failed.", champion.Id)
		return nil, err
	}

	var resp IChampionData
	_ = json.Unmarshal(body, &resp)
	ID, _ := strconv.Atoi(champion.Key)

	ret := &common.ChampionDataItem{
		Position:  resp.Header.Lane,
		Index:     cnt,
		Id:        champion.Key,
		Version:   sourceVersion,
		Timestamp: timestamp,
		Alias:     champion.Id,
		Name:      champion.Name,
	}

	highestWinBuild := common.ItemBuild{
		Title:               "[lolalytics](Gold+) Highest win: " + champion.Name + "@" + resp.Header.Lane + " " + sourceVersion,
		AssociatedMaps:      []int{11, 12},
		AssociatedChampions: []int{ID},
		Map:                 "any",
		Mode:                "any",
		PreferredItemSlots:  []string{},
		Sortrank:            1,
		StartedFrom:         "blank",
		Type:                "custom",
		Blocks:              makeBuildBlocksFromSet(resp.Summary.Items.Win),
	}
	ret.ItemBuilds = append(ret.ItemBuilds, highestWinBuild)

	mostCommonBuild := common.ItemBuild{
		Title:               "[lolalytics](Gold+) Most common: " + champion.Name + "@" + resp.Header.Lane + " " + sourceVersion,
		AssociatedMaps:      []int{11, 12},
		AssociatedChampions: []int{ID},
		Map:                 "any",
		Mode:                "any",
		PreferredItemSlots:  []string{},
		Sortrank:            1,
		StartedFrom:         "blank",
		Type:                "custom",
		Blocks:              makeBuildBlocksFromSet(resp.Summary.Items.Pick),
	}
	ret.ItemBuilds = append(ret.ItemBuilds, mostCommonBuild)

	fmt.Printf("[lolalytics] Fetched: %s@%s \n", champion.Name, resp.Header.Lane)
	return ret, nil
}

func Import(championAliasList map[string]common.ChampionItem, timestamp int64, debug bool) string {
	start := time.Now()
	fmt.Println("🌉 [lolalytics]: Start...")

	// get initial patch version/ep etc.
	body, err := common.MakeRequest("https://lolalytics.com/lol/rengar/build/")
	if err != nil {
		return err.Error()
	}

	html := string(body)
	eps := epReg.FindAllStringSubmatch(html, -1) // "ep=champion&p=d&v=9&patch=11.9&cid=107&lane=default&tier=platinum_plus&queue=420&region=all"
	epQuery := eps[0][0]
	sourceVersion := getSourceVersion(epQuery)
	queryMaker := makeQuery(epQuery)

	q := queryMaker("103", "middle")
	tierList, err := getTierList(q)
	if err != nil {
		return err.Error()
	}

	cIds := make([]string, 0, len(tierList.Cid))
	for key := range tierList.Cid {
		cIds = append(cIds, key)
	}

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan common.ChampionDataItem, len(cIds)*3)

	for _, cid := range cIds {
		if debug && cnt == 7 {
			break
		}

		if cnt > 0 && cnt%7 == 0 {
			fmt.Println(`🌉 Take a break...`)
			time.Sleep(time.Second * 5)
		}

		cnt += 1
		wg.Add(1)

		champion := getChampionById(cid, championAliasList)
		query := queryMaker(cid, "default")

		go func() {
			ret, err := makeBuild(champion, query, sourceVersion, timestamp, cnt)
			if err == nil {
				ch <- *ret
			}

			wg.Done()
		}()
	}

	wg.Wait()
	close(ch)

	for item := range ch {
		fmt.Println(item.ItemBuilds)
	}

	duration := time.Since(start)

	return fmt.Sprintf("🟢 [lolalytics.com] Finished, took: %s.", duration)
}
