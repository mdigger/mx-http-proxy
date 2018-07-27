package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/mdigger/log.v4"
	"gopkg.in/mdigger/mx.v2"
)

var (
	appName = "MX-HTTP-Proxy"
	version = "dev"
	commit  = ""   // версия git
	date    = ""   // дата сборки
	agent   string // используется как строка с именем сервиса

	// MXHost содержит адрес сервера MX.
	MXHost = "631hc.connector73.net"
)

func init() {
	// избавляемся от префикса `v` в версии
	version = strings.TrimPrefix(version, "v")
	// выводим информацию о текущей версии
	var verInfoFields = []log.Field{
		log.Field{Name: "name", Value: appName},
		log.Field{Name: "version", Value: version},
	}
	// инициализируем строку с агентом
	agent = fmt.Sprintf("%s/%s", appName, version)
	// если удалось разобрать дату, то добавляем ее в лог
	if date, err := time.Parse(time.RFC3339, date); err == nil {
		verInfoFields = append(verInfoFields,
			log.Field{Name: "built", Value: date.Format("2006-01-02")})
	}
	// добавляем идентификатор коммита, если он задан
	if commit != "" {
		verInfoFields = append(verInfoFields,
			log.Field{Name: "commit", Value: commit})
		agent += fmt.Sprintf(" (%s)", commit)
	}
	log.Info("service", verInfoFields)
}

func main() {
	// разбираем параметры сервиса
	var host = "localhost:8000"
	flag.StringVar(&host, "host", host, "http server `host`")
	flag.StringVar(&MXHost, "mx", MXHost, "mx server `host`")
	flag.Var(log.Flag(), "log", "log `level`")
	flag.Parse()

	// разбираем адрес HTTP-сервера
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		if err, ok := err.(*net.AddrError); ok && err.Err == "missing port in address" {
			hostname = err.Addr
		} else {
			log.Error("http host parse error", err)
			os.Exit(2)
		}
	}
	// формируем адрес для обращения к серверу
	var serverURL = &url.URL{Scheme: "http", Host: host, Path: "/"}
	// вычисляем, требуется ли получение сертификата
	var ssl = (port == "443" || port == "") &&
		hostname != "" &&
		hostname != "localhost" &&
		net.ParseIP(hostname) == nil &&
		strings.Trim(hostname, "[]") != "::1"
	if ssl {
		serverURL.Scheme = "https"
	}
	if hostname == "" {
		serverURL.Host = "localhost"
		if (ssl && port != "443") || (!ssl && port != "80") {
			serverURL.Host += ":" + port
		}
	}

	// настраиваем вывод лога MX
	mx.Logger = log.StdLog(log.TRACE, "mx")
}
