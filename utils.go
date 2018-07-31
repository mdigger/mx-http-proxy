package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
)

// getToken возвращает авторизационный токен из запроса
func getToken(r *http.Request) string {
	const authMethodName = "Bearer "
	// запрашивает токен авторизации из заголовка или параметра запроса
	var token = r.Header.Get("Authorization")
	if strings.HasPrefix(token, authMethodName) {
		return strings.TrimPrefix(token, authMethodName)
	}
	return r.URL.Query().Get("access_token")
}

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
