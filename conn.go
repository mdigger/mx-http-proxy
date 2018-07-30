package main

import (
	"gopkg.in/mdigger/mx.v2"
)

// Conn описывает соединение с сервером MX.
type Conn struct {
	*mx.Conn          // соединение с сервером MX
	login    mx.Login // информация для авторизации
}

// Connect подключается к серверу MX и авторизует пользователя.
func Connect(host string, login mx.Login) (*Conn, error) {
	conn, err := mx.Connect(host, login) // устанавливаем соединение и авторизуемся
	if err != nil {
		return nil, err
	}
	var mxconn = &Conn{
		Conn:  conn,
		login: login,
	}
	// go mxconn.reading() // запускаем обработчик входящих событий от сервера MX
	return mxconn, nil
}

// Close закрывает соедиение с сервером MX.
func (c *Conn) Close() error {
	return c.Conn.Close() // закрываем соединение
}
