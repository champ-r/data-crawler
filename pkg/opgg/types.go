package opgg

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
