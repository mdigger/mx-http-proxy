package main

import (
	"net/http"
	"strings"
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
