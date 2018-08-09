package main

import (
	"expvar"
	"fmt"
	"strings"
)

// Statistic описывает данные для мониторинга статистики сервиса.
type Statistic struct {
	MXHost   *expvar.String
	Connects *expvar.Int
	Commands *expvar.Map
	Events   *expvar.Map
	Hosts    *expvar.String
	TLS      *expvar.String
	SSE      *expvar.Int
}

func (s *Statistic) String() string {
	var buf strings.Builder
	buf.WriteString(`{"mx":{`)
	fmt.Fprintf(&buf, "\"host\":%s,", s.MXHost.String())
	fmt.Fprintf(&buf, "\"connects\":%s,", s.Connects.String())
	fmt.Fprintf(&buf, "\"commands\":%s,", s.Commands.String())
	fmt.Fprintf(&buf, "\"events\":%s", s.Events.String())
	buf.WriteString(`},"http":{`)
	fmt.Fprintf(&buf, "\"hosts\":%s,", s.Hosts.String())
	fmt.Fprintf(&buf, "\"tls\":%s,", s.TLS.String())
	fmt.Fprintf(&buf, "\"sse\":%s", s.SSE.String())
	buf.WriteString(`},"info":{`)
	var first = true
	for _, info := range logInfo {
		if !first {
			buf.WriteRune(',')
		} else {
			first = false
		}
		fmt.Fprintf(&buf, "%q:%q", info.Name, info.Value)
	}
	buf.WriteString(`}}`)
	return buf.String()
}

var staistic = &Statistic{
	MXHost:   new(expvar.String),
	Connects: new(expvar.Int),
	Commands: new(expvar.Map),
	Events:   new(expvar.Map),
	Hosts:    new(expvar.String),
	TLS:      new(expvar.String),
	SSE:      new(expvar.Int),
}

func init() {
	expvar.Publish("mx_proxy", staistic)
}
