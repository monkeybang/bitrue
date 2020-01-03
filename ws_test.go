package bitrue

import (
	"testing"
	"time"
)

func TestStartWs(t *testing.T) {
	StartWs("")
	for {
		time.Sleep(time.Second)
	}
}

func TestGetSigned(t *testing.T) {
}
