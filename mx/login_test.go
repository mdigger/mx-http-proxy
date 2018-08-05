package mx

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

func init() {
	Logger = log.New(os.Stderr, "MX ", log.Ltime)
}

var (
	mxhost = "631hc.connector73.net"
	login  = Login{
		UserName: "dmtest2",
		Password: "dmtest2",
		Type:     "User",
	}
	slogin = Login{
		UserName: "cstatest",
		Password: "cstapass",
		Type:     "Server",
	}
)

func TestLogin(t *testing.T) {
	conn, err := Connect(mxhost, slogin)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	var enc = json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(conn.Info); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}
