package bitrue

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"log"
)

//wss://ws.bitrue.com/kline-api/ws

func SubDepthWs(symbol, address string) chan *DepthWs {
	conn, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Println("websocket err:", err, symbol)
	}
	subMsg := `{"event":"sub","params":{"cb_id":"` + symbol + `","channel":"market_` + symbol + `_depth_step0"}}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(subMsg))
	if err != nil {
		log.Println(err)
	}

	_, message, err := conn.ReadMessage()
	//log.Println(msgType)

	unzipmsg, _ := ParseGzip(message)
	log.Printf(string(unzipmsg))
	ok := jsoniter.Get(unzipmsg, "status").ToString()
	ch := make(chan *DepthWs, 10)

	if ok == "ok" {
		go func() {
			for {
				_, message, err = conn.ReadMessage()
				unzipmsg, err := ParseGzip(message)
				if err != nil {
					log.Println(string(message), err)
					continue
				}

				ms := TimestampNowMs()
				//log.Printf(string(unzipmsg))
				cmd := string(unzipmsg[2:6])
				if cmd == "ping" {
					pong := fmt.Sprintf("{\"pong\":%d}", ms)
					conn.WriteMessage(websocket.TextMessage, []byte(pong))
					continue
				}
				//log.Println(string(unzipmsg))
				depthWs := &DepthWs{}
				err = json.Unmarshal(unzipmsg, depthWs)
				//data, _ := json.Marshal(depthWs)
				//log.Println(string(data))

				if err != nil {
					log.Println(err)
				}
				ch <- depthWs
			}
		}()
	}
	return ch
}
