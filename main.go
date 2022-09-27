package main

import (
	//"bytes"
	//"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	//"net"
	"net/http"
	"os"

	//"os/signal"
	//"github.com/bobertlo/go-mpg123/mpg123"
	//"github.com/gordonklaus/portaudio"
	socketio "github.com/googollee/go-socket.io"
	"github.com/labstack/echo"
	//
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

var addr = flag.String("addr", "localhost:8080", "http service address")
var store = flag.String("store", "./store", "path to mp3 storage")

func main() {
	loadData()

	go checkDidChange()

	for i := 0; i < 2; i++ {
		c[i].buffer = make([]byte, 40000)
		c[i].done = make(chan struct{})
	}

	//var handlers [2]*http.ServeMux
	log.SetFlags(0)
	//go Start(0, "0/songs")
	go NewHandler(0)
	fmt.Println(0)
	//go Start(1, "1/songs")
	go NewHandler(1)
	fmt.Println(1)
	// for i, station := range stations {
	// 	log.SetFlags(0)
	// 	go Start(i, station.StationId+"/songs")
	// 	handlers[i] = NewHandler(i)
	// 	fmt.Println(i)
	// }
	// for i, _ := range stations {
	// 	go log.Fatal(http.ListenAndServe("localhost:808"+stations[i].StationId, handlers[i]))
	// }

	// TODO: Получение из БД.
	go http.HandleFunc("/getCurrentStations", func(w http.ResponseWriter, r *http.Request) {
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

	go http.ListenAndServe(":8080", nil)

	// //TODO: Добавить API для админки.
	// listener, _ := net.Listen("tcp", "localhost:8081") // открываем слушающий сокет
	// for {
	// 	conn, err := listener.Accept() // принимаем TCP-соединение от клиента и создаем новый сокет
	// 	if err != nil {
	// 		continue
	// 	}
	// 	go handleClient(conn) // обрабатываем запросы клиента в отдельной го-рутине
	// }

	startSocketServer()
}

// func handleClient(conn net.Conn) {
// 	// station := Station{StationId: "1", StationURL: "2", StationImageURL: "3", StationDescription: "4", StationTitle: "Title", CurrentTrackName: "6", CurrentTrackImageURL: "12"}
// 	// var stations [1]Station
// 	// stations[0] = station
// 	answer := SocketAnswer{Event: "stationsDidChange", Stations: stations[:]}
// 	defer conn.Close() // закрываем сокет при выходе из функции
// 	for {
// 		if gotApiRequest {
// 			if trackDidChange {
// 				answer = SocketAnswer{Event: "stationsDidChange", Stations: stations[:]}
// 			}
// 			if apiRequest == "updateStations" {
// 				stationsDidChange = true
// 			}
// 			jsonResp, err := json.Marshal(answer.Stations)
// 			if err != nil {
// 				fmt.Println(err)
// 				//return
// 			}
// 			conn.Write(jsonResp) // пишем в сокет
// 			gotApiRequest = false
// 		}
// 	}
// }

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func checkDidChange() {
	for {
		if gotApiRequest {
			if apiRequest == "updateStations" {
				loadData()
				stationsDidChange = true
			}
			if apiRequest == "trackDidChange" {
				trackDidChange = true
			}
			gotApiRequest = false
		}
	}
}

func loadData() {
	jsonFile, err := os.Open("resources/stations.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully opened stations.json")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &stations)
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
	e.Logger.Fatal(e.Start(":8081"))
}

func socketHandler(s socketio.Conn) {
	for {
		if trackDidChange {
			trackDidChange = false
			var data = stations[changedStationId]
			s.Emit("CurrentTrackDidChange", data)
		}
		if stationsDidChange {
			stationsDidChange = false
			s.Emit("StationsDidChange", stations)
		}
	}
}
