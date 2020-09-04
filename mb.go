package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type VersionResponse struct {
	UpToDateVersion string `json:"upToDateVersion"`
	GameTypes       string `json:"gameTypes"`
}

const MurderBridgeBUrl = `https://d23wati96d2ixg.cloudfront.net`

func getLatestVersion() (string, error) {
	url := MurderBridgeBUrl + `/save/general.json`
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", errors.New(res.Status)
	}

	body, _ := ioutil.ReadAll(res.Body)
	var verResp VersionResponse
	_ = json.Unmarshal(body, &verResp)
	return verResp.UpToDateVersion, nil
}

func Import() {
	ver, _ := getLatestVersion()
	fmt.Println(ver)
}
