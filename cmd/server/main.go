package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/supershabam/gamepad"
	"io"
	"io/ioutil"
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

func eventChan(in <-chan []byte) <-chan gamepad.Event {
	out := make(chan gamepad.Event)
	go func() {
		defer close(out)
		for b := range in {
			log.Printf("parsing: %s", b)
			m := map[string]interface{}{}
			err := json.Unmarshal(b, &m)
			if err != nil {
				log.Print(err)
				return
			}
			if t, ok := m["type"].(string); ok {
				switch t {
				case "up":
					if v, ok := m["value"].(bool); ok {
						out <- gamepad.UpDPadEvent{v}
					}
				}
			}
		}
	}()
	return out
}

func main() {
	flag.Parse()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(*assets))))
	http.HandleFunc("/controller.ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		bc := make(chan []byte)
		gp := gamepad.NewGamepad(eventChan(bc))

		done := make(chan struct{})
		btnc := make(chan gamepad.Event, 1)
		gp.Notify(btnc, gamepad.DPadUp)
		go func() {
			for {
				t, r, err := conn.NextReader()
				if err != nil {
					if err == io.EOF {
						close(done)
						return
					}
					log.Fatal(err)
				}
				if t == websocket.TextMessage {
					b, err := ioutil.ReadAll(r)
					if err != nil {
						log.Print(err)
						continue
					}
					bc <- b
				}
			}
		}()

		for {
			select {
			case <-done:
				return
			case b := <-btnc:
				if b.Bool() == false {
					log.Printf("let off up!")
					close(done)
				}
			case <-time.Tick(time.Second):
				log.Printf("%+v", gp.State())
			}
		}
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
