package lolalytics

type ITierList struct {
	Cid map[string]interface{} `json:"cid"`
}

type IChampionData struct {
	Header struct {
		N           int    `json:"n"`
		DefaultLane string `json:"defaultLane"`
		Lane        string `json:"lane"`
		Counters    struct {
			Strong []int `json:"strong"`
			Weak   []int `json:"weak"`
		} `json:"counters"`
		Wr        float64 `json:"wr"`
		Pr        float64 `json:"pr"`
		Br        float64 `json:"br"`
		Rank      int     `json:"rank"`
		RankTotal int     `json:"rankTotal"`
		Tier      string  `json:"tier"`
		TopWin    float64 `json:"topWin"`
		TopElo    string  `json:"topElo"`
		Damage    struct {
			Physical float64 `json:"physical"`
			Magic    float64 `json:"magic"`
			True     float64 `json:"true"`
		} `json:"damage"`
	} `json:"header"`
	Summary struct {
		Skillpriority struct {
			Win struct {
				ID string  `json:"id"`
				N  int     `json:"n"`
				Wr float64 `json:"wr"`
			} `json:"win"`
			Pick struct {
				ID string  `json:"id"`
				N  int     `json:"n"`
				Wr float64 `json:"wr"`
			} `json:"pick"`
		} `json:"skillpriority"`
		Skillorder struct {
			Win struct {
				ID int64   `json:"id"`
				N  int     `json:"n"`
				Wr float64 `json:"wr"`
			} `json:"win"`
			Pick struct {
				ID int64   `json:"id"`
				N  int     `json:"n"`
				Wr float64 `json:"wr"`
			} `json:"pick"`
		} `json:"skillorder"`
		Sum struct {
			Pick struct {
				ID string  `json:"id"`
				N  int     `json:"n"`
				Wr float64 `json:"wr"`
			} `json:"pick"`
			Win struct {
				ID string  `json:"id"`
				N  int     `json:"n"`
				Wr float64 `json:"wr"`
			} `json:"win"`
		} `json:"sum"`
		Sums  []int `json:"sums"`
		Runes struct {
			Pick struct {
				Wr   float64 `json:"wr"`
				N    int     `json:"n"`
				Page struct {
					Pri int `json:"pri"`
					Sec int `json:"sec"`
				} `json:"page"`
				Set struct {
					Pri []int `json:"pri"`
					Sec []int `json:"sec"`
					Mod []int `json:"mod"`
				} `json:"set"`
			} `json:"pick"`
			Win struct {
				Wr   float64 `json:"wr"`
				N    int     `json:"n"`
				Page struct {
					Pri int `json:"pri"`
					Sec int `json:"sec"`
				} `json:"page"`
				Set struct {
					Pri []int `json:"pri"`
					Sec []int `json:"sec"`
					Mod []int `json:"mod"`
				} `json:"set"`
			} `json:"win"`
		} `json:"runes"`
		Items struct {
			Win struct {
				Start struct {
					N   int     `json:"n"`
					Wr  float64 `json:"wr"`
					Set []int   `json:"set"`
				} `json:"start"`
				Core struct {
					Set []int   `json:"set"`
					Wr  float64 `json:"wr"`
					N   int     `json:"n"`
				} `json:"core"`
				Item4 []struct {
					ID int     `json:"id"`
					N  int     `json:"n"`
					Wr float64 `json:"wr"`
				} `json:"item4"`
				Item5 []struct {
					ID int     `json:"id"`
					N  int     `json:"n"`
					Wr float64 `json:"wr"`
				} `json:"item5"`
				Item6 []struct {
					ID int     `json:"id"`
					N  int     `json:"n"`
					Wr float64 `json:"wr"`
				} `json:"item6"`
			} `json:"win"`
			Pick struct {
				Start struct {
					N   int     `json:"n"`
					Wr  float64 `json:"wr"`
					Set []int   `json:"set"`
				} `json:"start"`
				Core struct {
					Set []int   `json:"set"`
					Wr  float64 `json:"wr"`
					N   int     `json:"n"`
				} `json:"core"`
				Item4 []struct {
					ID int     `json:"id"`
					N  int     `json:"n"`
					Wr float64 `json:"wr"`
				} `json:"item4"`
				Item5 []struct {
					ID int     `json:"id"`
					N  int     `json:"n"`
					Wr float64 `json:"wr"`
				} `json:"item5"`
				Item6 []struct {
					ID int     `json:"id"`
					N  int     `json:"n"`
					Wr float64 `json:"wr"`
				} `json:"item6"`
			} `json:"pick"`
		} `json:"items"`
	} `json:"summary"`
	Graph struct {
		Dates []string `json:"dates"`
		Wr    struct {
			All         []float64 `json:"all"`
			DiamondPlus []float64 `json:"diamond_plus"`
			Platinum    []float64 `json:"platinum"`
			Gold        []float64 `json:"gold"`
			Silver      []float64 `json:"silver"`
			Bronze      []float64 `json:"bronze"`
			Iron        []float64 `json:"iron"`
		} `json:"wr"`
		Wrs struct {
			All         []float64 `json:"all"`
			DiamondPlus []float64 `json:"diamond_plus"`
			Platinum    []float64 `json:"platinum"`
			Gold        []float64 `json:"gold"`
			Silver      []float64 `json:"silver"`
			Bronze      []float64 `json:"bronze"`
			Iron        []float64 `json:"iron"`
		} `json:"wrs"`
		Pr struct {
			All         []float64 `json:"all"`
			DiamondPlus []float64 `json:"diamond_plus"`
			Platinum    []int     `json:"platinum"`
			Gold        []float64 `json:"gold"`
			Silver      []float64 `json:"silver"`
			Bronze      []float64 `json:"bronze"`
			Iron        []float64 `json:"iron"`
		} `json:"pr"`
		N struct {
			All         []int `json:"all"`
			DiamondPlus []int `json:"diamond_plus"`
			Platinum    []int `json:"platinum"`
			Gold        []int `json:"gold"`
			Silver      []int `json:"silver"`
			Bronze      []int `json:"bronze"`
			Iron        []int `json:"iron"`
		} `json:"n"`
		Br struct {
			All         []float64 `json:"all"`
			DiamondPlus []float64 `json:"diamond_plus"`
			Platinum    []float64 `json:"platinum"`
			Gold        []float64 `json:"gold"`
			Silver      []float64 `json:"silver"`
			Bronze      []float64 `json:"bronze"`
			Iron        []float64 `json:"iron"`
		} `json:"br"`
	} `json:"graph"`
	Nav struct {
		Lanes struct {
			Top     float64 `json:"top"`
			Jungle  float64 `json:"jungle"`
			Middle  float64 `json:"middle"`
			Bottom  float64 `json:"bottom"`
			Support float64 `json:"support"`
		} `json:"lanes"`
	} `json:"nav"`
	Analysed   int             `json:"analysed"`
	AvgWinRate float64         `json:"avgWinRate"`
	Top        [][]interface{} `json:"top"`
	Depth      []interface{}   `json:"depth"`
	N          int             `json:"n"`
	Skills     struct {
		SkillEarly  [][][]int       `json:"skillEarly"`
		Skill6Pick  int             `json:"skill6Pick"`
		Skill10Pick int             `json:"skill10Pick"`
		SkillOrder  [][]interface{} `json:"skillOrder"`
	} `json:"skills"`
	Time map[string]int `json:"time"`
	TimeWin map[string]int `json:"timeWin"`
	TopStats struct {
		Toppick  int     `json:"toppick"`
		Toprank  int     `json:"toprank"`
		Topcount int     `json:"topcount"`
		Topwin   float64 `json:"topwin"`
		Topelo   string  `json:"topelo"`
	} `json:"topStats"`
	Stats      [][]interface{} `json:"stats"`
	StatsCount int             `json:"statsCount"`
	Runes      struct {
		Stats map[string][]float64 `json:"stats"`
	} `json:"runes"`
	Objective map[string][]float64 `json:"objective"`
	Spell    [][]float64         `json:"spell"`
	Spells   [][]interface{} `json:"spells"`
	ItemSets struct {
		ItemBootSet1 map[string][]int `json:"itemBootSet1"`
		ItemBootSet2 map[string][]int `json:"itemBootSet2"`
		ItemBootSet3 map[string][]int `json:"itemBootSet3"`
	} `json:"itemSets"`
	StartItem    [][]float64         `json:"startItem"`
	StartSet     [][]interface{} `json:"startSet"`
	EarlyItem    [][]float64         `json:"earlyItem"`
	Boots        [][]float64         `json:"boots"`
	MythicItem   [][]float64         `json:"mythicItem"`
	PopularItem  [][]float64         `json:"popularItem"`
	WinningItem  [][]float64         `json:"winningItem"`
	Item         [][]float64         `json:"item"`
	Item1        [][]float64         `json:"item1"`
	Item2        [][]float64         `json:"item2"`
	Item3        [][]float64         `json:"item3"`
	Item4        [][]float64         `json:"item4"`
	Item5        [][]float64         `json:"item5"`
	EnemyTop     [][]float64     `json:"enemy_top"`
	EnemyJungle  [][]float64     `json:"enemy_jungle"`
	EnemyMiddle  [][]float64     `json:"enemy_middle"`
	EnemyBottom  [][]float64     `json:"enemy_bottom"`
	EnemySupport [][]float64     `json:"enemy_support"`
	Response     struct {
		Platform string `json:"platform"`
		Version  int    `json:"version"`
		EndPoint string `json:"endPoint"`
		Valid    bool   `json:"valid"`
		Duration string `json:"duration"`
	} `json:"response"`
}
