package web

import (
	"testing"
	"time"

	"github.com/fatalbanana/filetundra/internal/env"
)

func TestRunWebserver(t *testing.T) {
	env.Env.HTTPAddress = "127.0.0.1"
	go RunWebserver()
	for Server == nil {
		time.Sleep(10 * time.Microsecond)
	}
	Server.Close()
}
