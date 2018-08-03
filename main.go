package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
	"golang.org/x/crypto/acme/autocert"
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
	var conns = new(Conns) // инициализируем список подключений к MX
	defer conns.Close()    // закрываем все соединения по окончании

	// инициализируем обработку HTTP запросов
	var mux = &rest.ServeMux{
		Headers: map[string]string{"Server": agent},
		Logger:  log.New("http"),
	}
	// обработчики команд
	mux.Handle("POST", "/login", conns.Login)
	mux.Handle("POST", "/logout", conns.Logout)
	mux.Handle("POST", "/:cmd", conns.Commands)
	mux.Handle("GET", "/events", conns.Events)

	// инициализируем и запускаем сервер HTTP
	var server = http.Server{
		Addr:              host,
		Handler:           mux,
		IdleTimeout:       10 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          mux.Logger.StdLog(log.ERROR),
	}
	// добавляем автоматическую поддержку TLS сертификатов для сервиса
	if ssl {
		var manager = autocert.Manager{
			Prompt: autocert.AcceptTOS,
			HostPolicy: func(_ context.Context, host string) error {
				if host != hostname {
					mux.Logger.Error("unsupported https host", "host", host)
					return errors.New("acme/autocert: host not configured")
				}
				return nil
			},
			Email: "dmitrys@xyzrd.com",
			Cache: autocert.DirCache("letsEncrypt.cache"),
		}
		server.TLSConfig = &tls.Config{
			GetCertificate: manager.GetCertificate,
		}
		server.Addr = ":https"
		// поддержка получения сертификата Let's Encrypt и редирект на HTTPS
		go http.ListenAndServe(":http", manager.HTTPHandler(nil))
	}
	mux.Logger.Info("server",
		"listen", server.Addr, "tls", ssl, "url", serverURL.String())

	// отслеживаем сигнал о прерывании и останавливаем по нему сервер
	go func() {
		var sigint = make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			mux.Logger.Error("server shutdown", err)
		}
	}()

	// запускаем веб-сервер
	if ssl {
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		mux.Logger.Error("server", err)
	} else {
		mux.Logger.Info("server stopped")
	}
}
