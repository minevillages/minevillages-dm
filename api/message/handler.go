package message

import (
	"fmt"
	"log"
	"minevillages/dm/api/db"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// ChatRooms는 각 채팅방을 관리하는 맵입니다.
type ChatRooms map[string]*ChatRoom

// ChatRoom은 채팅방을 나타내는 구조체입니다.
type ChatRoom struct {
	Name            string
	Users           map[string]*websocket.Conn // 사용자 목록
	Join            chan *websocket.Conn       // 사용자 참가 채널
	Leave           chan string                // 사용자 퇴장 채널
	Broadcast       chan db.Message            // 메시지 브로드캐스트 채널
	OfflineMessages map[string][]db.Message    // 오프라인 메시지 저장 맵
	mu              sync.Mutex                 // 동기화를 위한 뮤텍스
}

// Run은 채팅방을 실행하는 메서드입니다.
func (cr *ChatRoom) Run(sender string, receiver string) {
	// 오프라인 메시지 불러오기
	cr.loadOfflineMessages()

	for {
		select {
		case conn := <-cr.Join:
			// 새로운 사용자가 채팅방에 참가함을 알림
			cr.mu.Lock()

			if cr.Users[sender] == nil {
				cr.Users[sender] = conn
				fmt.Println(sender)
			} else {
				cr.Users[receiver] = conn
				fmt.Println(receiver)
			}
			cr.mu.Unlock()

			// 오프라인 메시지 확인 후 전달
			if offlineMsgs, ok := cr.OfflineMessages[sender]; ok {
				for _, msg := range offlineMsgs {
					cr.mu.Lock()
					if _, ok := cr.Users[msg.Receiver]; ok {
						err := cr.Users[msg.Receiver].WriteJSON(msg)
						if err != nil {
							log.Printf("Error sending offline message: %v", err)
							delete(cr.Users, msg.Receiver)
						}
					}
					cr.mu.Unlock()
				}
				delete(cr.OfflineMessages, sender) // 전달한 메시지 삭제
			}
			go cr.handleMessages(conn, sender)
		case sender := <-cr.Leave:
			// 사용자가 채팅방을 나감을 알림
			cr.mu.Lock()
			delete(cr.Users, sender)
			cr.mu.Unlock()
		case msg := <-cr.Broadcast:
			cr.mu.Lock()
			// receiver에게 메시지 전송
			if receiverConn, ok := cr.Users[receiver]; ok {
				err := receiverConn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending message to receiver: %v", err)
					// 사용자가 오프라인인 경우 오프라인 메시지로 저장
					cr.saveOfflineMessage(msg)
				}
			} else {
				// 사용자가 오프라인인 경우 오프라인 메시지로 저장
				cr.saveOfflineMessage(msg)
			}

			// sender에게 메시지 전송
			if senderConn, ok := cr.Users[sender]; ok {
				err := senderConn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending message to sender: %v", err)
					// 사용자가 오프라인인 경우 오프라인 메시지로 저장
					cr.saveOfflineMessage(msg)
				}
			} else {
				// 사용자가 오프라인인 경우 오프라인 메시지로 저장
				cr.saveOfflineMessage(msg)
			}

			cr.mu.Unlock()
		}
	}
}

// handleMessages는 사용자의 메시지를 처리하는 메서드입니다.
func (cr *ChatRoom) handleMessages(conn *websocket.Conn, sender string) {
	defer func() {
		// 사용자가 채팅방을 나감을 알림
		cr.Leave <- sender
		conn.Close()
	}()

	for {
		var msg db.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}
		// 메시지를 채팅방에 브로드캐스트
		cr.Broadcast <- msg
	}
}

// saveOfflineMessage는 오프라인 메시지를 저장하는 메서드입니다.
func (cr *ChatRoom) saveOfflineMessage(msg db.Message) {
	if err := msg.Insert(); err != nil {
		log.Printf("Error saving offline message: %v", err)
	}
}

// loadOfflineMessages는 오프라인 메시지를 로드하는 메서드입니다.
func (cr *ChatRoom) loadOfflineMessages() {
	var messages []db.Message
	message := &db.Message{}
	messages, err := message.Find()
	if err != nil {
		log.Printf("Error loading offline messages: %v", err)
		return
	}

	for _, msg := range messages {
		cr.OfflineMessages[msg.Receiver] = append(cr.OfflineMessages[msg.Receiver], msg)
	}
}

// Upgrader는 웹소켓 연결 업그레이드를 위한 구조체입니다.
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
