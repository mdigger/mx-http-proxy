package main

import "expvar"

var (
	events     expvar.Map
	commands   expvar.Map
	connects   expvar.Int
	ssecounter expvar.Int
	mhost      expvar.String
)

func init() {
	var m = expvar.NewMap("mx")
	m.Set("host", &mhost)
	m.Set("connects", &connects)
	m.Set("sse", &ssecounter)
	commands.Init()
	m.Set("commands", &commands)
	events.Init()
	m.Set("events", &events)
}
