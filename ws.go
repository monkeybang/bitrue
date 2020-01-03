package bitrue

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"net/url"
	"time"
)

func StartWs(symbol string) {

	//url := `wss://ws.bitrue.com/kline-api/ws`

	//origin := "http://127.0.0.1:8080/"
	u := url.URL{Scheme: "wss", Host: "ws.bitrue.com", Path: "/kline-api/ws"}
	log.Println(u)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("websocket err:", err, symbol)
	}

	subMsg := `{"event":"sub","params":{"channel":"market_btrusdt_kline_1min","cb_id":"btrusdt"}}`
	//err = sendWs([]byte(subMsg), ws)
	conn.WriteMessage(websocket.TextMessage, []byte(subMsg))
	if err != nil {
		log.Println("sub detail err:", err)
	}
	msgType, message, err := conn.ReadMessage()
	log.Println(msgType)

	unzipmsg, _ := ParseGzip(message)
	log.Printf(string(unzipmsg))
	ok := jsoniter.Get(unzipmsg, "status").ToString()
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
				log.Printf(string(unzipmsg))
				cmd := string(unzipmsg[2:6])
				if cmd == "ping" {
					pong := fmt.Sprintf("{\"pong\":%d}", ms)
					conn.WriteMessage(websocket.TextMessage, []byte(pong))
					continue
				}
				kline := &Kline{}
				err = json.Unmarshal(unzipmsg, kline)
				if err != nil {
					log.Println(err)
				}
				log.Printf("%# v", pretty.Formatter(kline))

			}
		}()
	}

}

func sendWs(message []byte, ws *websocket.Conn) error {
	return nil
}

func ParseGzip(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, data)
	r, err := gzip.NewReader(b)
	if err != nil {
		fmt.Println("[ParseGzip] NewReader error: , maybe data is ungzip: ", err, string(data))
		return nil, err
	} else {
		defer r.Close()
		undatas, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println("[ParseGzip]  ioutil.ReadAll error: :", err, string(data))
			return nil, err
		}
		return undatas, nil
	}
}

func TimestampNowMs() int64 {
	var timestamp int64
	timestamp = time.Now().UTC().UnixNano() / 1000000
	return timestamp
}
