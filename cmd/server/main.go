package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/supershabam/gamepad"
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

type gamepadEventJson struct {
	Button  string `json:"button"`
	Pressed bool   `json:"pressed"`
}

func gamepadEvent(in []byte) (e gamepad.Event, err error) {
	gej := gamepadEventJson{}
	err = json.Unmarshal(in, &gej)
	if err != nil {
		return
	}
	e.Pressed = gej.Pressed
	switch gej.Button {
	case "up":
		e.Button = gamepad.Up
	case "down":
		e.Button = gamepad.Down
	case "left":
		e.Button = gamepad.Left
	case "right":
		e.Button = gamepad.Right
	default:
		err = fmt.Errorf("unknown button: %s", gej.Button)
		return
	}
	return e, nil
}

type Gamestate struct {
	Color string  `json:"color"`
	Paths []int64 `json:"paths"`
}

func game(p1 *gamepad.Gamepad) <-chan Gamestate {
	gs := Gamestate{
		Color: "green",
		Paths: []int64{34},
	}
	out := make(chan Gamestate)
	go func() {
		defer close(out)
		t := time.Tick(time.Millisecond * 16)
		for {
			select {
			case <-t:
				s := p1.State()
				if s.Right {
					gs.Paths = append(gs.Paths, gs.Paths[len(gs.Paths)-1]+1)
				}
			}
			out <- gs
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
	http.HandleFunc("/lightbike.ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		done := make(chan struct{})

		defer close(done)
		eventc := make(chan gamepad.Event)
		defer close(eventc)
		p1 := gamepad.NewGamepad(eventc)
		go func() {
			statec := game(p1)
			for {
				select {
				case <-done:
					return
				case s := <-statec:
					b, err := json.Marshal(s)
					if err != nil {
						log.Print(err)
						continue
					}
					err = conn.WriteMessage(websocket.TextMessage, b)
					if err != nil {
						log.Print(err)
						continue
					}
				}
			}
		}()
		for {
			t, b, err := conn.ReadMessage()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			if t == websocket.TextMessage {
				e, err := gamepadEvent(b)
				if err != nil {
					log.Print(err)
					continue
				}
				eventc <- e
			}
		}
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
