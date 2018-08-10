package app

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// DefaultPort используется как порт по умолчанию при разборе функцией Port.
var DefaultPort = "8000"

// Port возвращает нормализованное имя порта и хоста (опционально), на котором
// будет отвечать веб сервер.
func Port(port string) (string, error) {
	if port == "" {
		return ":" + DefaultPort, nil // порт по умолчанию
	}
	if p, err := strconv.ParseInt(port, 10, 16); err == nil && p > 0 {
		return ":" + port, nil // указан только порт
	}
	if _, _, err := net.SplitHostPort(port); err != nil {
		if err, ok := err.(*net.AddrError); ok && err.Err == "missing port in address" {
			return net.JoinHostPort(strings.Trim(err.Addr, "[]"), DefaultPort), nil
		}
		return "", err
	}
	return port, nil // возвращаем как есть, ибо все хорошо
}

// Настройки, используемые для получения и хранения сертификатов Let's Encrypt.
var (
	LetsEncryptEmail = "dmitrys@xyzrd.com"
	LetsEncryptCache = "letsEncrypt.cache"
)

// LetsEncrypt возвращает инициализированный конфиг поддержки TLS с
// использованием сертификатов Let's Encrypt. Автоматически запускает веб
// сервер на 80 порту, который поддерживает проверку валидности домена и
// перенаправляет все запросы на 443 порт.
func LetsEncrypt(hosts ...string) *tls.Config {
	var manager = autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
		Email:      LetsEncryptEmail,
		Cache:      autocert.DirCache(LetsEncryptCache),
	}
	// поддержка получения сертификата Let's Encrypt и редирект на HTTPS
	go http.ListenAndServe(":http", manager.HTTPHandler(nil))
	return &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: manager.GetCertificate,
	}
}

// LoadCertificates загружает пары сертификатов из указанного каталога.
// Сертификаты должны быть в формате PEM и иметь одинаковые имена для приватного
// ключа (*.key) и публичного сертификата (*.crt). Если ни одного сертификата
// не загружено, то возвращает nil.
func LoadCertificates(dir string) (*tls.Config, error) {
	// проверяем, что есть локальные сертификаты и загружаем их
	keys, err := filepath.Glob(filepath.Join(dir, "*.key"))
	if err != nil {
		return nil, err // ошибка может быть только в шаблоне
	}
	if len(keys) == 0 {
		return nil, nil // ничего не найдено
	}
	var certificates = make([]tls.Certificate, 0, len(keys))
	// перебираем все найденные файлы
	for _, keyfile := range keys {
		// загружаем пару файлов с сертификатами
		cert, err := tls.LoadX509KeyPair(keyfile[:len(keyfile)-3]+"crt", keyfile)
		if err != nil {
			if _, ok := err.(*os.PathError); ok {
				continue // игнорируем ошибки с файлами
			}
			return nil, err
		}
		certificates = append(certificates, cert)
	}
	if len(certificates) == 0 {
		return nil, nil // ни одного сертификата не загружено
	}
	var cfg = &tls.Config{Certificates: certificates}
	cfg.BuildNameToCertificate() // инициализируем список хостов сертификатов
	return cfg, nil
}
