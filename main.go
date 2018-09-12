//go:generate go run -tags=dev assets_generate.go

package main

import (
	"context"
	"expvar"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/mdigger/app-info"
	"github.com/mdigger/mx-http-proxy/mx"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
)

var (
	appName = "MX-HTTP-Proxy"
	version string // версия приложения
	commit  string // идентификатор коммита git
	date    string // дата сборки

	mxhost = app.Env("MX", "") // адрес сервера MX
)

func main() {
	// разбираем параметры сервиса
	flag.StringVar(&mxhost, "mx", mxhost, "mx server `host`")
	var httphost = flag.String("port", app.Env("PORT", ":8000"),
		"http server `port`")
	flag.Parse()

	// выводим в лог информацию о версии сервиса
	app.Parse(appName, version, commit, date)
	log.Info("service", app.LogInfo())

	var mxlogger = log.New("mx")
	mx.Logger = mxlogger.StdLog(log.TRACE) // настраиваем вывод лога MX
	// проверяем доступность сервера MX
	if _, err := mx.Connect(mxhost, nil); err != nil {
		mxlogger.Error("mx server unavailable", "host", mxhost, err)
		os.Exit(2)
	}

	mxlogger.Info("using mx server", "host", mxhost)
	mhost.Set(mxhost)
	var conns = new(Conns) // инициализируем список подключений к MX
	defer conns.Close()    // закрываем все соединения по окончании

	// разбираем имя хоста и порт, на котором будет слушать веб-сервер
	port, err := app.Port(*httphost)
	if err != nil {
		log.Error("http host parse error", err)
		os.Exit(2)
	}
	// инициализируем обработку HTTP запросов
	var httplogger = log.New("http")
	var mux = &rest.ServeMux{
		Headers: map[string]string{
			"Server":                      app.Agent,
			"Access-Control-Allow-Origin": "*",
		},
		Logger: httplogger,
	}
	// обработчики команд
	mux.Handle("POST", "/login", conns.Login)
	mux.Handle("POST", "/logout", conns.Logout)
	mux.Handle("POST", "/:cmd", conns.Commands)
	mux.Handle("GET", "/events", conns.Events)
	// добавляем обработку отдачи документации и дополнительных статических
	// файлов через веб-сервер
	mux.Handle("GET", "/*file", rest.HTTPFiles(assets, "index.html"))
	// проверяем, что вывод осуществляется из "живого" каталога, а не из
	// внедренных в исполняемый файл данных
	if assets, ok := assets.(http.Dir); ok {
		httplogger.Warn("live http assets", "folder", assets)
	}
	// добавляем версию документации
	if file, err := assets.Open("openapi.yaml"); err == nil {
		if data, err := ioutil.ReadAll(file); err == nil {
			var v = regexp.MustCompile(`version:\s(.+)`).FindSubmatch(data)
			if len(v) == 2 {
				var ver = string(v[1])
				httplogger.Info("api docs", "version", ver)
				mux.Headers["X-API-Version"] = ver // добавляем версию API
			}
		}
		file.Close()
	} else {
		// документация недоступна ни в виде отдельного каталога, ни в виде
		// встроенных в исполняемый файл данных
		httplogger.Warn("no http documentation")
	}

	// добавляем в статистику
	mux.Handle("GET", "/debug/vars", rest.HTTPHandler(expvar.Handler()))

	// инициализируем и запускаем сервер HTTP
	var server = http.Server{
		Addr:              port,
		Handler:           mux,
		IdleTimeout:       10 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          httplogger.StdLog(log.ERROR),
	}

	// отслеживаем сигнал о прерывании и останавливаем по нему сервер
	go func() {
		var sigint = make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			httplogger.Error("server shutdown", err)
		}
	}()
	// добавляем в статистику и выводим в лог информацию о запущенном сервере
	if server.TLSConfig != nil {
		// добавляем заголовок с обязательством использования защищенного
		// соединения в ближайший час
		mux.Headers["Strict-Transport-Security"] = "max-age=3600"
	}
	httplogger.Info("server",
		"listen", server.Addr,
		"tls", server.TLSConfig != nil,
	)

	if err = server.ListenAndServe(); err != http.ErrServerClosed {
		httplogger.Error("server", err)
	} else {
		httplogger.Info("server stopped")
	}
}
