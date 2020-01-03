package bitrue

import (
	"testing"
	"time"
)

func TestSubDepthWs(t *testing.T) {
	SubDepthWs("btrusdt", "wss://ws.bitrue.com/kline-api/ws")
	for {
		time.Sleep(time.Second)
	}
}
