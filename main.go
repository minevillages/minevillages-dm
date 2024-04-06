package main

import (
	"log"
	"minevillages/dm/api"
	baseapi "minevillages/dm/api/base"
	"minevillages/dm/api/db"
	"net/http"
	"runtime"
	// "github.com/quic-go/quic-go/http3"
)

func main() {
	// 프로세스의 모든 CPU 코어를 활용하도록 설정합니다.
	runtime.GOMAXPROCS(runtime.NumCPU())

	handler := api.HostRouteHandler{}
	handler.RegisterHost("127.0.0.1:7001", baseapi.HTTPHandler{})

	// MongoDB 를 연결합니다.
	mongo := &db.Mongo{
		Uri: "mongodb://127.0.0.1:27017",
	}
	mongo.Connection()

	if err := http.ListenAndServe(
		":7001",
		// "ssl/certificate.crt",
		// "ssl/private.key",
		&handler,
	); err != nil {
		log.Fatalln("HTTP 관련 수신 소켓을 초기화하는 과정에서 예외가 발생하였습니다.", err.Error())
	}

}
