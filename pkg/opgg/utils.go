package opgg

import (
	"data-crawler/pkg/common"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
)

func genOverview(allChampions map[string]common.ChampionItem, aliasList map[string]string, aram bool) (*OverviewData, int) {
	url := SourceUrl
	if aram {
		url = AramSourceUrl
	}
	doc, err := common.ParseHTML(url + `/statistics`)
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
		if aram {
			count += 1
		} else {
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
		}
	})

	return &d, count
}
