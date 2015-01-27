package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	assets = flag.String("assets", "./assets", "assets directory")
)

type Bike struct {
	Color string  `json:"color"`
	Path  []int64 `json:"path"`
}

func genBike() Bike {
	pathc := rand.Intn(300)
	bike := Bike{Color: "green", Path: []int64{}}
	for count := 0; count < pathc; count++ {
		point := rand.Int63n(1000)
		bike.Path = append(bike.Path, point)
	}
	return bike
}

func main() {
	flag.Parse()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(*assets))))
	http.HandleFunc("/lightbike.ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		go func() {
			for _ = range time.Tick(100 * time.Millisecond) {
				bike := genBike()
				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Print(err)
					return
				}
				b, err := json.Marshal(bike)
				if err != nil {
					log.Print(err)
					return
				}
				if _, err := w.Write(b); err != nil {
					log.Print(err)
					return
				}
				if err := w.Close(); err != nil {
					log.Print(err)
					return
				}
			}
		}()
		for {
			_, _, err := conn.NextReader()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
		}
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
