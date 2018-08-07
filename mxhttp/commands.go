package main

import "encoding/xml"

// Содержит описание поддерживаемых команд и ответов на них.
//
// Для тех ответов, которые ненужно разбирать, используется простая "заглушка"
// в виде NamedResponse. Это позволяет дожидаться ответа от сервера MX и
// отслеживать ответы с ошибками.
type (
	// NamedResponse описывает пустой ответ, заставляющий ожидать
	// ответа на команду от сервера MX.
	NamedResponse struct {
		// XMLName xml.Name `json:"-"` // в принципе, нам это имя ничего не дает
	}

	// MonitorStartRequest command starts monitoring of CSTA events
	// (MonitorStartResponse).
	MonitorStartRequest struct {
		XMLName       xml.Name `xml:"MonitorStart" json:"-"`
		Object        string   `xml:"monitorObject>deviceObject" json:"device"`
		UseConfEvents bool     `xml:"confEvents,omitempty" json:"conf,omitempty"`
	}
	// MonitorStartResponse описывает информацию, возвращаемую при старте
	// монитора.
	MonitorStartResponse struct {
		XMLName xml.Name `xml:"MonitorStartResponse" json:"-"`
		ID      int      `xml:"monitorCrossRefID" json:"monitor"`
		Voice   bool     `xml:"actualMonitorMediaClass>voice" json:"voice,omitempty"`
	}
	// MonitorStopRequest command stops monitoring of CSTA events
	// (empty MonitorStopResponse).
	MonitorStopRequest struct {
		XMLName xml.Name `xml:"MonitorStop" json:"-"`
		ID      int      `xml:"monitorCrossRefID" json:"monitor"`
	}
	// MonitorStartAbRequest запускает монитор отслеживания изменений в контактах
	MonitorStartAbRequest struct {
		XMLName xml.Name `xml:"MonitorStartAb" json:"-"`
	}
	// MonitorStopAbRequest останавливает монитор отслеживания изменений в контактах
	MonitorStopAbRequest struct {
		XMLName xml.Name `xml:"MonitorStopAb" json:"-"`
	}
)
