package mx

import (
	"encoding/xml"
	"log"
	"testing"
	"time"
)

func TestEvents(t *testing.T) {
	conn, err := Connect(mxhost, login)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		for event := range conn.Events() {
			log.Println("EVENT", event)
		}
	}()

	if err = conn.Send(&struct {
		XMLName xml.Name `xml:"MonitorStart"`
		Ext     string   `xml:"monitorObject>deviceObject"`
	}{
		Ext: conn.Ext,
	}, nil); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 10)
}
