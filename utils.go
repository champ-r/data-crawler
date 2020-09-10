package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"io/ioutil"
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
	for _, v := range arr {
		if v == el {
			return arr
		}
	}

	return append(arr, el)
}

func MakeRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	body, _ := ioutil.ReadAll(res.Body)
	return body, nil
}

func GetChampionList() (*ChampionListResp, string, error) {
	body, err := MakeRequest(DataDragonUrl + "/api/versions.json")
	if err != nil {
		return nil, "", err
	}

	var versionArr []string
	_ = json.Unmarshal(body, &versionArr)
	version := versionArr[0]

	cBody, cErr := MakeRequest(DataDragonUrl + "/cdn/" + version + "/data/en_US/champion.json")
	if cErr != nil {
		return nil, "", errors.New(`data dragon: request champion list failed`)
	}

	var resp ChampionListResp
	_ = json.Unmarshal(cBody, &resp)

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
	body, err := MakeRequest(url)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	reader := ioutil.NopCloser(buf)
	return goquery.NewDocumentFromReader(reader)
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
