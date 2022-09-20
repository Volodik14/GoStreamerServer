package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Менять при разном количестве станций!!
var c [2]chunk

//var c chunk

func Start(stationId int, storePath string) {
	files, err := ioutil.ReadDir(storePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })
		for _, file := range files {
			fName := file.Name()
			stations[stationId].CurrentTrackName = fName
			changedStationId = stationId
			fmt.Println(fName)
			if filepath.Ext(fName) != ".mp3" {
				continue
			}
			f, err := os.Open(path.Join(storePath, fName))
			defer f.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

			go c[stationId].Load(f)
			<-c[stationId].done
			apiRequest = "trackDidChange"
			gotApiRequest = true
		}
	}
}
