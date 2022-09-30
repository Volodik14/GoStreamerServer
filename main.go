package main

import (
	//"bytes"
	//"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"os"

	//"os/signal"
	//"github.com/bobertlo/go-mpg123/mpg123"
	//"github.com/gordonklaus/portaudio"
	socketio "github.com/googollee/go-socket.io"
	"github.com/labstack/echo"
)

var gotApiRequest = false
var apiRequest = ""
var trackDidChange = false
var stationsDidChange = false
var changedStationId = 0

type Station struct {
	StationId            string `json:"stationId"`
	StationURL           string `json:"stationURL"`
	StationImageURL      string `json:"stationImageURL"`
	StationDescription   string `json:"stationDescription"`
	StationTitle         string `json:"stationTitle"`
	CurrentTrackName     string `json:"currentTrackName"`
	CurrentTrackImageURL string `json:"currentTrackImageURL"`
}

type SocketAnswer struct {
	Event    string    `json:"event"`
	Stations []Station `json:"stations"`
}

var stations []Station

var addr = flag.String("addr", "localhost:9090", "http service address")
var store = flag.String("store", "./store", "path to mp3 storage")

func main() {
	loadData()

	go checkDidChange()

	startServers()

	http.HandleFunc("/getCurrentStations", func(w http.ResponseWriter, r *http.Request) {
		answer := SocketAnswer{Event: "stationsDidChange", Stations: stations[:]}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			jsonResp, err := json.Marshal(answer.Stations)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write(jsonResp)
			//json.NewEncoder(w).Encode(station)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}
	})
	// TODO: Получение из БД.
	http.HandleFunc("/setCurrentStations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			err := json.NewDecoder(r.Body).Decode(&stations)
			if err != nil {
				fmt.Println(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeData()
			startServers()
		} else {
			http.Error(w, "Invalid request method.", 405)
		}
	})
	trackDidChange = true

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			query := r.URL.Query()
			request := query.Get("request")
			w.WriteHeader(200)
			if request == "" {
				http.Error(w, "Empty request", 405)
			} else {
				w.Write([]byte("OK"))
				apiRequest = request
				gotApiRequest = true
			}

		} else {
			http.Error(w, "Invalid request method.", 405)
		}
	})

	go http.ListenAndServe(":9090", nil)
	startSocketServer()
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func checkDidChange() {
	for {
		if gotApiRequest {
			if apiRequest == "updateStations" {
				stationsDidChange = true
			}
			if apiRequest == "trackDidChange" {
				trackDidChange = true
			}
			gotApiRequest = false
		}
	}
}

func startServers() {
	for i := 0; i < len(stations); i++ {
		newChunk := chunk{}
		newChunk.buffer = make([]byte, 40000)
		newChunk.done = make(chan struct{})
		if i > len(c)-1 {
			c = append(c, newChunk)
			go NewHandler(i)
		} else {
			c[i] = newChunk
		}
	}
}

func loadData() {
	jsonFile, err := os.Open("stations.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully opened stations.json")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &stations)
}

func writeData() {
	stationsJson, _ := json.Marshal(stations)
	ioutil.WriteFile("stations.json", stationsJson, 0644)
}

func startSocketServer() {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("Connected")
		log.Println("connected:", s.ID())
		go socketHandler(s)
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		log.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	e := echo.New()
	e.HideBanner = true

	e.Static("/", "../asset")
	e.Any("/socket.io/", func(context echo.Context) error {
		server.ServeHTTP(context.Response(), context.Request())
		return nil
	})
	e.Logger.Fatal(e.Start(":9091"))
}

func socketHandler(s socketio.Conn) {
	for {
		if trackDidChange {
			trackDidChange = false
			var data = stations[changedStationId]
			s.Emit("CurrentTrackDidChange", data)
		}
		if stationsDidChange {
			loadData()
			startServers()
			stationsDidChange = false
			s.Emit("StationsDidChange", stations)
		}
	}
}
