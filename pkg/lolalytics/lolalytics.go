package lolalytics

import (
	"data-crawler/pkg/common"
	"encoding/json"
	"fmt"
	"regexp"
)

var cidReg = regexp.MustCompile("&cid=\\d+?&")
var laneReg = regexp.MustCompile("&lane=[a-zA-Z]+?&")
var epReg = regexp.MustCompile("ep=.*?region=all")
var patchReg = regexp.MustCompile("&patch=((\\d+\\.)+\\d?)&")

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

func Import(championAliasList map[string]common.ChampionItem, timestamp int64) string {
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

	fmt.Println(cIds, sourceVersion)
	return fmt.Sprintf("🟢 [lolalytics.com] Finished. Took some time.")
}
