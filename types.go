package main

type BlockItem struct {
	Id    string `json:"id"`
	Count int    `json:"count"`
}

type ItemBuildBlockItem struct {
	Type  string      `json:"type"`
	Items []BlockItem `json:"items"`
}

type ItemBuild struct {
	Title               string               `json:"title"`
	AssociatedMaps      []int                `json:"associatedMaps"`
	AssociatedChampions []int                `json:"associatedChampions"`
	Blocks              []ItemBuildBlockItem `json:"blocks"`
}

type RuneItem struct {
	Name            string `json:"name"`
	PickCount       string `json:"pickCount"`
	WinRate         string `json:"winRate"`
	PrimaryStyleId  int    `json:"primaryStyleId"`
	SubStyleId      int    `json:"subStyleId"`
	SelectedPerkIds []int  `json:"selectedPerkIds"`
}

type ChampionDataItem struct {
	Index      int         `json:"index"`
	Id         string      `json:"id"`
	Version    string      `json:"version"`
	Timestamp  int64       `json:"timestamp"`
	Alias      string      `json:"alias"`
	Name       string      `json:"name"`
	Position   string      `json:"position"`
	Skills     []string    `json:"skills"`
	Spells     []string    `json:"spells"`
	ItemBuilds []ItemBuild `json:"itemBuilds"`
	Runes      []RuneItem  `json:"runes"`
}

type ChampionItem struct {
	Version string `json:"version"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Blurb   string `json:"blurb"`
	Info    struct {
		Attack     int `json:"attack"`
		Defense    int `json:"defense"`
		Magic      int `json:"magic"`
		Difficulty int `json:"difficulty"`
	} `json:"info"`
	Image struct {
		Full   string `json:"full"`
		Sprite string `json:"sprite"`
		Group  string `json:"group"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		W      int    `json:"w"`
		H      int    `json:"h"`
	} `json:"image"`
	Tags    []string `json:"tags"`
	Partype string   `json:"partype"`
	Stats   struct {
		Hp                   int     `json:"hp"`
		Hpperlevel           int     `json:"hpperlevel"`
		Mp                   int     `json:"mp"`
		Mpperlevel           int     `json:"mpperlevel"`
		Movespeed            int     `json:"movespeed"`
		Armor                int     `json:"armor"`
		Armorperlevel        int     `json:"armorperlevel"`
		Spellblock           int     `json:"spellblock"`
		Spellblockperlevel   int     `json:"spellblockperlevel"`
		Attackrange          int     `json:"attackrange"`
		Hpregen              int     `json:"hpregen"`
		Hpregenperlevel      int     `json:"hpregenperlevel"`
		Mpregen              int     `json:"mpregen"`
		Mpregenperlevel      int     `json:"mpregenperlevel"`
		Crit                 int     `json:"crit"`
		Critperlevel         int     `json:"critperlevel"`
		Attackdamage         int     `json:"attackdamage"`
		Attackdamageperlevel int     `json:"attackdamageperlevel"`
		Attackspeedperlevel  float32 `json:"attackspeedperlevel"`
		Attackspeed          float32 `json:"attackspeed"`
	} `json:"stats"`
}

type ChampionListResp struct {
	Type    string                  `json:"type"`
	Format  string                  `json:"format"`
	Version string                  `json:"version"`
	Data    map[string]ChampionItem `json:"data"`
}

type PkgInfo struct {
	Timestamp int64 `json:"timestamp"`
}
