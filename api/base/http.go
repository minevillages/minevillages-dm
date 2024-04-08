package baseapi

import (
	"log"
	"minevillages/dm/api/db"
	"minevillages/dm/api/health"
	"minevillages/dm/api/message"
	"minevillages/dm/cache"
	"minevillages/dm/json"
	"minevillages/dm/util"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/gorilla/websocket"
)

// HTTPHandler는 HTTP 요청을 처리하는 핸들러입니다.
type HTTPHandler struct {
	http.Handler
}

// ChatRooms는 채팅방을 관리하는 맵입니다.
var ChatRooms map[string]*message.ChatRoom

// ServeHTTP는 HTTP 요청을 처리하는 메서드입니다.
func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	firstURI := util.FirstURI(r.URL.Path)
	if firstURI == "ws" {
		h.handleWebSocket(w, r)
	} else if firstURI == "api" {
		h.BackHandler(w, r)
	} else {
		h.FrontHandler(w, r)
	}
}

// BackHandler는 API 요청을 처리하는 메서드입니다.
func (HTTPHandler) BackHandler(w http.ResponseWriter, r *http.Request) {
	var data interface{}
	if r.URL.Path == "/api/health" {
		data = health.Health()
	}
	json.WriteWith(data, w)
}

// FrontHandler는 프론트엔드 리소스 요청을 처리하는 메서드입니다.
func (HTTPHandler) FrontHandler(w http.ResponseWriter, r *http.Request) {
	ext := filepath.Ext(r.URL.Path)
	res := cache.Resource{}

	if ext == "" {
		res.Path = util.JoinPath("../client", "dist/index.html")
		w.Header().Set("Cache-Control", "max-age=60")
	} else {
		res.Path = util.JoinPath("../client", filepath.Join("dist", r.URL.Path))
	}

	if ext == ".ttf" || ext == ".otf" || ext == ".woff" || ext == ".woff2" {
		res.IgnoreCompress = true
		w.Header().Set("Cache-Control", "max-age=31536000, public")
	}
	res.WriteWith(w, r)
}

// handleWebSocket는 웹소켓 요청을 처리하는 메서드입니다.
func (HTTPHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	sender := r.URL.Query().Get("sender")

	receiver := r.URL.Query().Get("receiver")
	array := []string{sender, receiver}
	sort.Strings(array)
	name := array[0] + "_" + array[1]

	if ChatRooms == nil {
		ChatRooms = make(map[string]*message.ChatRoom)
	}

	if _, exists := ChatRooms[name]; !exists {
		cr := &message.ChatRoom{
			Name:            name,
			Users:           make(map[string]*websocket.Conn),
			Join:            make(chan *websocket.Conn),
			Leave:           make(chan string),
			Broadcast:       make(chan db.Message),
			OfflineMessages: make(map[string][]db.Message),
		}
		ChatRooms[name] = cr

		go cr.Run(sender, receiver)
		cr.Join <- conn
	} else {
		cr := ChatRooms[name]
		cr.Join <- conn
	}

}
