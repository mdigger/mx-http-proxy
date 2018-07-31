package main

import (
	"crypto/rand"
	"encoding/base64"
	"sync"

	"github.com/mdigger/log"
)

// Conns описывает список подключений к серверам MX.
type Conns struct {
	list sync.Map // список подключений
}

// Close закрывает все соединеия из списка.
func (l *Conns) Close() {
	l.list.Range(func(k, v interface{}) bool {
		l.list.Delete(k)
		v.(*Conn).Close()
		return true
	})
}

// Delete удаляет соединение из списка не закрывая его.
func (l *Conns) Delete(token string) {
	log.Debug("delete connection", "token", token)
	l.list.Delete(token)
}

const tokenSize = 12 // задает размер токена

// Store добавляет новое соединение в список и возвращает ассоциированный с ним
// токен.
func (l *Conns) Store(conn *Conn) string {
	// генерируем случайный токен
	var b = make([]byte, tokenSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	var token = base64.RawURLEncoding.EncodeToString(b)
	log.Debug("store connection", "login", conn.login.UserName, "token", token)
	l.list.Store(token, conn) // сохраняем соединение в списке
	// при закрытии соединения автоматически удалить из списка
	conn.SetCloser(func(err error) {
		l.Delete(token)
	})
	return token
}
