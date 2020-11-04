package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
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
	go func() {
		ch <- ImportOPGG(allChampionData.Data, championAliasList, officialVer, timestamp)
	}()
	go func() {
		ch <- ImportMB(allChampionData.Data, timestamp)
	}()

	x := <-ch
	y := <-ch
	fmt.Println(x)
	fmt.Println(y)
}
