package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

func main() {
	wsu := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsu.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		for {
			t, r, err := conn.NextReader()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			if t != websocket.TextMessage {
				continue
			}
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := io.Copy(w, r); err != nil {
				log.Fatal(err)
			}
			if err := w.Close(); err != nil {
				log.Fatal(err)
			}

		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<!DOCTYPE html>
<h1 id="ohhai">oh hai</h1>
<script>
var ws = new WebSocket('ws://localhost:8080/ws')
// a dumb controller model
var c = {
  left: 0.0,
  right: 0.0 
}
var ohhai = document.getElementById('ohhai')
ws.onopen = function() {
  ws.onmessage = function(e) {
    try {
      m = JSON.parse(e.data)
      ohhai.style.position = 'absolute'
      ohhai.style.left = (100 * m.left) + (-25 * m.right) + 'px'
    } catch (err) {}
  }
  document.addEventListener('keyup', function(e) {
    switch(e.keyCode) {
    case 65:
      if (c.left == 0) {
        return
      }
      c.left = 0
      break
    case 68:
      if (c.right == 0) {
        return
      }
      c.right = 0
      break
    default:
      return
    }
    ws.send(JSON.stringify(c))
  })
  document.addEventListener('keydown', function(e) {
    switch(e.keyCode) {
    case 65:
      if (c.left == 1) {
        return
      }
      c.left = 1
      break
    case 68:
      if (c.right == 1) {
        return
      }
      c.right = 1
      break
    default:
      return
    }
    ws.send(JSON.stringify(c))
  })
  ws.send(JSON.stringify(c))
}
</script>
`)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
