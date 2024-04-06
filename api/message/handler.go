package message

import (
	"log"
	"minevillages/dm/api/db"
	"net/http"

	"github.com/gorilla/websocket"
)

var Clients = make(map[*websocket.Conn]bool) // 연결된 클라이언트들을 저장하기 위한 맵
var Broadcast = make(chan db.Message)        // 메시지를 브로드캐스트하는 채널

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handler() {
	for {
		// 브로드캐스트 채널로부터 메시지를 받음
		msg := <-Broadcast
		msg.Insert()
		// 연결된 모든 클라이언트로 메시지를 보냄
		for client := range Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(Clients, client)
			}
		}
	}
}
