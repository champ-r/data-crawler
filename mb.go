package main

import (
	"encoding/json"
	"fmt"
)

type VersionResponse struct {
	UpToDateVersion string `json:"upToDateVersion"`
	GameTypes       string `json:"gameTypes"`
}

const MurderBridgeBUrl = `https://d23wati96d2ixg.cloudfront.net`

func getLatestVersion() (string, error) {
	url := MurderBridgeBUrl + `/save/general.json`
	body, err := MakeRequest(url)
	if err != nil {
		return "", err
	}

	var verResp VersionResponse
	_ = json.Unmarshal(body, &verResp)
	return verResp.UpToDateVersion, nil
}

func ImportMB() {
	ver, _ := getLatestVersion()
	fmt.Println(ver)
}
