package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

type VersionResp struct {
	UpToDateVersion string `json:"upToDateVersion"`
	GameTypes       string `json:"gameTypes"`
}

type ChampionDataRespDataItem struct {
	WinRate   float32 `json:"winRate"`
	Ratio     []int   `json:"ratio"`
	Frequency float32 `json:"frequency"`
}

type ChampionDataResp struct {
	WinRate     float32                             `json:"winRate"`
	Rank        int                                 `json:"rank"`
	BanRate     float32                             `json:"banRate"`
	Stats       map[string]float32                  `json:"stats"`
	NumGames    int                                 `json:"numGames"`
	Runes       map[string]ChampionDataRespDataItem `json:"runes"`
	Skills      map[string]ChampionDataRespDataItem `json:"skills"`
	Summoners   map[string]ChampionDataRespDataItem `json:"summoners"`
	Duration    map[string]ChampionDataRespDataItem `json:"duration"`
	NumBans     int                                 `json:"numBans"`
	Adjustments string                              `json:"adjustments"`
	Frequency   float32                             `json:"frequency"`
	NumWins     int                                 `json:"numWins"`
	Items       struct {
		Counter  map[string]ChampionDataRespDataItem   `json:"counter"`
		Order    []map[string]ChampionDataRespDataItem `json:"order"`
		Build    map[string]ChampionDataRespDataItem   `json:"build"`
		Starting map[string]ChampionDataRespDataItem   `json:"starting"`
	} `json:"items"`
}

const MurderBridgeBUrl = `https://d23wati96d2ixg.cloudfront.net`

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

func getChampionData(alias string, version string) (*ChampionDataResp, error) {
	url := MurderBridgeBUrl + `/save/` + version + `/ARAM/` + alias + `.json`
	body, err := MakeRequest(url)
	if err != nil {
		return nil, err
	}

	var data ChampionDataResp
	_ = json.Unmarshal(body, &data)
	return &data, nil
}

func ImportMB(championAliasList map[string]string) {
	ver, _ := getLatestVersion()
	fmt.Println(ver)

	wg := new(sync.WaitGroup)
	cnt := 0
	ch := make(chan ChampionDataResp, len(championAliasList))
	for _, alias := range championAliasList {
		cnt += 1
		wg.Add(1)
		go func(_alias string, _ver string, _cnt int) {
			d, _ := getChampionData(_alias, _ver)
			ch <- *d
			wg.Done()
		}(alias, ver, cnt)
	}
	wg.Wait()

	fmt.Printf(`done, %v`, ch)
}
