package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	opggFlag := flag.Bool("opgg", false, "Fetch & generate data from op.gg")
	mbFlag := flag.Bool("mb", false, "Fetch & generate murderbridge.com")
	debugFlag := flag.Bool("debug", false, "only for debug")

	flag.Parse()
	fmt.Println(os.Args)

	timestamp := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	allChampionData, officialVer, err := GetChampionList()
	if err != nil {
		log.Fatal(err)
	}

	var championAliasList = make(map[string]string)
	for k, v := range allChampionData.Data {
		championAliasList[v.Name] = k
	}

	ch := make(chan string)
	var opgg, mb string

	if *opggFlag {
		fmt.Println("[CMD] Fetch data from op.gg")
		go func() {
			ch <- ImportOPGG(allChampionData.Data, championAliasList, officialVer, timestamp, *debugFlag)
		}()
	}

	if *mbFlag {
		fmt.Println("[CMD] Fetch data from murderbridge.com")
		go func() {
			ch <- ImportMB(allChampionData.Data, timestamp)
		}()
	}

	opgg = <- ch
	mb = <- ch
	if opgg != "" {
		fmt.Println(opgg)
	}
	if mb != "" {
		fmt.Println(mb)
	}
}
