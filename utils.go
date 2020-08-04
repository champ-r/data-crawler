package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

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

func GetChampionList() (*ChampionListResp, string, error) {
	res, err := http.Get(DataDragonUrl + "/api/versions.json")
	if err != nil {
		return nil, "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, "", errors.New(`data dragon: request version failed`)
	}

	body, _ := ioutil.ReadAll(res.Body)
	var versionArr []string
	_ = json.Unmarshal(body, &versionArr)
	version := versionArr[0]

	cRes, cErr := http.Get(DataDragonUrl + "/cdn/" + version + "/data/en_US/champion.json")
	if cErr != nil {
		return nil, "", errors.New(`data dragon: request champion list failed`)
	}

	defer cRes.Body.Close()
	if cRes.StatusCode != 200 {
		log.Fatal("Request lol version failed.")
	}

	body, _ = ioutil.ReadAll(cRes.Body)
	var resp ChampionListResp
	_ = json.Unmarshal(body, &resp)

	fmt.Printf("🤖 Got official champion list, total %d \n", len(resp.Data))
	return &resp, version, nil
}

func SaveJSON(fileName string, data interface{}) error {
	file, _ := json.MarshalIndent(data, "", "  ")
	wErr := ioutil.WriteFile(fileName, file, 0644)

	if wErr != nil {
		return wErr
	}

	return nil
}

func ParseHTML(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	return goquery.NewDocumentFromReader(res.Body)
}

func GenPkgInfo(tplPath string, vars interface{}) (string, error) {
	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		return "", err
	}

	var tplBytes bytes.Buffer
	err = tpl.Execute(&tplBytes, vars)
	if err != nil {
		return "", err
	}

	return tplBytes.String(), nil
}
