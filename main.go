//go:generate go run -tags=dev assets_generate.go

package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mdigger/mx-http-proxy/mx"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
	"golang.org/x/crypto/acme/autocert"
)

var (
	mxhost = env("MX", "") // адрес сервера MX
)

func env(name, def string) string {
	if s, ok := os.LookupEnv(name); ok {
		return s
	}
	return def
}

func main() {
	// разбираем параметры сервиса
	flag.StringVar(&mxhost, "mx", mxhost, "mx server `host`")
	var httphost = flag.String("port", env("PORT", ":8000"), "http server `port`")
	var letsencrypt = flag.String("letsencrypt", "", "domain `host` name")
	flag.Parse()
	log.Info("service", logInfo) // выводим в лог информацию о версии сервиса

	var mxlogger = log.New("mx")
	mx.Logger = mxlogger.StdLog(log.TRACE) // настраиваем вывод лога MX
	// проверяем доступность сервера MX
	if _, err := mx.Connect(mxhost, nil); err != nil {
		mxlogger.Error("connection", "addr", mxhost, err)
		os.Exit(2)
	}
	mxlogger.Info("server", "addr", mxhost)
	var conns = new(Conns) // инициализируем список подключений к MX
	defer conns.Close()    // закрываем все соединения по окончании

	// разбираем имя хоста
	if port, err := strconv.ParseInt(*httphost, 10, 16); err == nil && port > 0 {
		*httphost = fmt.Sprintf(":%s", *httphost) // указан только порт
	} else if _, _, err := net.SplitHostPort(*httphost); err != nil {
		if err, ok := err.(*net.AddrError); ok && err.Err == "missing port in address" {
			*httphost = net.JoinHostPort(strings.Trim(err.Addr, "[]"), "80")
		} else {
			log.Error("http host parse error", err)
			os.Exit(2)
		}
	}
	// инициализируем обработку HTTP запросов
	var httplogger = log.New("http")
	var mux = &rest.ServeMux{
		Headers: map[string]string{
			"Server":                      appAgent,
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

	// инициализируем и запускаем сервер HTTP
	var server = http.Server{
		Addr:              *httphost,
		Handler:           mux,
		IdleTimeout:       10 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          httplogger.StdLog(log.ERROR),
	}
	// настраиваем автоматическое получение сертификата
	if *letsencrypt != "" {
		if *letsencrypt == "localhost" || net.ParseIP(*letsencrypt) != nil {
			httplogger.Error("let's encrypt host",
				fmt.Errorf("bad host name: %s", *letsencrypt))
			os.Exit(2)
		}
		// добавляем заголовок с обязательством использования защищенного
		// соединения в ближайший час
		mux.Headers["Strict-Transport-Security"] = "max-age=3600"
		// настраиваем поддержку TLS для сервера
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			// NextProtos: []string{http2.NextProtoTLS, "http/1.1"},
		}
		// добавляем сертификат, если он уже загружен
		var tlsCertificates []tls.Certificate // загруженные и разобранные сертификаты
		if tlsCertificates != nil {
			server.TLSConfig.Certificates = tlsCertificates
			server.TLSConfig.BuildNameToCertificate()
			var hosts = make([]string, 0, len(server.TLSConfig.NameToCertificate))
			for name := range server.TLSConfig.NameToCertificate {
				hosts = append(hosts, name)
			}
			httplogger.Info("server with tls certificate",
				"listen", server.Addr,
				"tls", true,
				"hosts", hosts,
			)
		} else {
			var manager = autocert.Manager{
				Prompt: autocert.AcceptTOS,
				HostPolicy: func(_ context.Context, host string) error {
					if host != *letsencrypt {
						httplogger.Warn("ignore address", "host", host)
						return errors.New("acme/autocert: host not configured")
					}
					return nil
				},
				Email: "dmitrys@xyzrd.com",
				Cache: autocert.DirCache("letsEncrypt.cache"),
			}
			// добавляем получение и обновление сертификатов
			server.TLSConfig.GetCertificate = manager.GetCertificate
			server.Addr = ":https" // подменяем порт на 443
			// поддержка получения сертификата Let's Encrypt и редирект на HTTPS
			go http.ListenAndServe(":http", manager.HTTPHandler(nil))
			httplogger.Info("server with let'encrypt autocert",
				"listen", []string{":80", ":443"},
				"tls", true,
				"host", *letsencrypt,
			)
		}
	} else {
		httplogger.Info("server",
			"listen", server.Addr,
			"tls", false,
		)
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
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		httplogger.Error("server", err)
	} else {
		httplogger.Info("server stopped")
	}

	if isDocker() {
		info()
	}
}

// печатает информацию о содержимом контейнера
// используется для отладки
func info() {
	fmt.Println("----------------------------------------------")

	if val, err := os.Getwd(); err == nil {
		fmt.Printf("pwd: %s\n", val)
	}
	if val, err := os.Hostname(); err == nil {
		fmt.Printf("host: %s\n", val)
	}
	if val, err := user.Current(); err == nil {
		fmt.Printf("user: %v\n", val)
	}

	fmt.Println("environment:")
	for _, env := range os.Environ() {
		fmt.Printf("- %s\n", env)
	}

	readFile("/etc/hosts")
	readFile("/etc/passwd")

	fmt.Println("files:")
	if err := filepath.Walk("/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("error: %v\n", err)
				return nil
			}
			if info.IsDir() {
				if name := info.Name(); name == "proc" || name == "sys" {
					return filepath.SkipDir
				}
			}
			fmt.Printf("- %s\n", path)
			return nil
		}); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("----------------------------------------------")
}

func readFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Printf("%s:\n", name)
	var r = bufio.NewReader(file)
	for {
		str, err := r.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Print("- ", str)
	}
	return nil
}
