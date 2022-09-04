package main

import (
	"fmt"
	"net/http"
	"time"
)

func NewHandler(stationId int) {
	//router := http.NewServeMux()
	http.HandleFunc("/"+fmt.Sprintf("%v", stationId), func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, stationId)
	})
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
