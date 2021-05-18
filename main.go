package main

import (
	"data-crawler/pkg/common"
	la "data-crawler/pkg/lolalytics"
	mb "data-crawler/pkg/murderbridge"
	op "data-crawler/pkg/opgg"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	debugFlag := flag.Bool("debug", false, "only for debug")
	opggFlag := flag.Bool("opgg", false, "Fetch & generate data from op.gg")
	mbFlag := flag.Bool("mb", false, "Fetch & generate murderbridge.com")
	laFlag := flag.Bool("la", false, "Fetch & generate lolalytics.com")

	flag.Parse()
	fmt.Println(os.Args)

	timestamp := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	allChampionData, officialVer, err := common.GetChampionList()
	if err != nil {
		log.Fatal(err)
	}
	runeLoopUp, allRunes, err := common.GetRunesReforged(officialVer)
	if err != nil {
		log.Fatal(err)
	}

	var championAliasList = make(map[string]string)
	for k, v := range allChampionData.Data {
		championAliasList[v.Name] = k
	}

	ch := make(chan string)
	var opggRet, mbRet, opggAramRet, laRet string

	if *opggFlag {
		fmt.Println("[CMD] Fetch data from op.gg")
		go func() {
			ch <- op.Import(allChampionData.Data, championAliasList, officialVer, timestamp, *debugFlag)
		}()
		go func() {
			ch <- op.ImportAram(allChampionData.Data, championAliasList, officialVer, timestamp, *debugFlag)
		}()

		opggRet = <-ch
		opggAramRet = <-ch
		fmt.Println(opggRet)
		fmt.Println(opggAramRet)
	}

	if *mbFlag {
		fmt.Println("[CMD] Fetch data from murderbridge.com")
		go func() {
			ch <- mb.Import(allChampionData.Data, timestamp, runeLoopUp, allRunes, *debugFlag)
		}()

		mbRet = <-ch
		fmt.Println(mbRet)
	}

	if *laFlag {
		fmt.Println("[CMD] Fetch data from lolalytics.com")
		go func() {
			ch <- la.Import(allChampionData.Data, officialVer, timestamp, runeLoopUp, *debugFlag)
		}()

		laRet = <-ch
		fmt.Println(laRet)
	}
}
