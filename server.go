package main

import (
	"fmt"
	"net/http"
	"time"
)

func NewHandler(stationId int) {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, stationId)
	})
	server := http.Server{
		Addr:    fmt.Sprintf(":%v", 9000+stationId), // :{port}
		Handler: router,
	}
	go server.ListenAndServe()
	go Start(stationId, fmt.Sprintf("%v", stationId)+"/songs")
}

func serve(w http.ResponseWriter, r *http.Request, stationId int) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "audio/mpeg")

	for {
		select {
		case <-r.Context().Done():
			break
		default:
			t := time.Now()
			fmt.Println(t.Format("20060102150405"))
			w.Write(c[stationId].Value())
			flusher.Flush()
			time.Sleep(time.Second)
		}
	}
}
