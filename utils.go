package main

import (
	"encoding/json"
	"mime"
	"net"
	"net/http"
	"strings"

	"github.com/Connector73/mx-http-proxy/mx"
	"github.com/mdigger/log"
	"github.com/mdigger/rest"
)

// jsonBind разбирает запрос в формате JSON и сериализует его в указанный объект.
func jsonBind(r *http.Request, v interface{}) *rest.Error {
	mediatype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	charset, ok := params["charset"]
	if ok && strings.ToLower(charset) != "utf-8" {
		return rest.NewError(http.StatusUnsupportedMediaType,
			"Unsupported mediatype charset")
	}
	switch mediatype {
	case "":
		return nil
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(v); err != nil {
			log.Error("json params decode error", err)
			return rest.NewError(http.StatusBadRequest, "Decode params error")
		}
		return nil
	default:
		return rest.NewError(http.StatusUnsupportedMediaType,
			"Unsupported mediatype")
	}
}

// httpError подменяет ошибки, возвращаемые сервером MX на ошибки HTTP.
func httpError(err error) *rest.Error {
	// в зависимости от типа ошибки возвращаем разное описание
	switch err := err.(type) {
	case *mx.LoginError: // ошибка авторизации пользователя
		return rest.NewError(http.StatusForbidden, err.Error())
	case *mx.TimeoutError: // ошибка ожидания ответа
		return rest.NewError(http.StatusGatewayTimeout, err.Error())
	case *mx.CSTAError: // ошибка от сервера MX
		return rest.NewError(http.StatusBadRequest, err.Error())
	}
	// сервер MX не отвечает
	if err, ok := err.(net.Error); ok {
		if err.Timeout() {
			log.Warn("mx response timeout", err)
			return rest.NewError(http.StatusGatewayTimeout,
				"MX response timeout")
		}
		log.Error("mx not availible", err)
		return rest.NewError(http.StatusBadGateway, "MX not availible")
	}
	// ошибка не предусмотрена обработчиком
	log.Error("mx error", err)
	return rest.NewError(http.StatusInternalServerError, "MX error")
}
