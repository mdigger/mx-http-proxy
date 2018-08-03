package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
	"gopkg.in/mdigger/mx.v2"
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
		if err != nil {
			conn.sse.Data("error", err.Error(), "")
		}
		l.Delete(token)
	})
	return token
}

// Login обрабатывает авторизацю соединения.
func (l *Conns) Login(c *rest.Context) error {
	// разбираем параметры логина
	var login = new(mx.Login)
	if err := jsonBind(c.Request, login); err != nil {
		return err
	}
	c.AddLogField("user", login.UserName)
	// подключаемся к серверу и авторизуем пользователя
	conn, err := Connect(MXHost, *login)
	if err != nil {
		return httpError(err)
	}
	// сохраняем подключение в списке и отдаем токен в ответ
	var token = l.Store(conn)
	return c.Write(&struct {
		Token string `json:"token,omitempty"`
		mx.Info
	}{
		Token: token,
		Info:  conn.Info,
	})
}

// authorize авторизует запрос и возвращает информацию о соединении.
func (l *Conns) authorize(c *rest.Context) (*Conn, *rest.Error) {
	const authMethodName = "Bearer "
	// запрашивает токен авторизации из заголовка или параметра запроса
	var token = c.Header("Authorization")
	if strings.HasPrefix(token, authMethodName) {
		token = strings.TrimPrefix(token, authMethodName)
	} else {
		token = c.Query("access_token")
	}
	if token == "" {
		c.SetHeader("WWW-Authenticate", fmt.Sprintf("Token realm=%q", appName))
		return nil, rest.ErrUnauthorized
	}
	c.SetData("token", token) // сохраняем токен в контексте запроса
	// получаем соединение по токену
	mxconn, ok := l.list.Load(token)
	if !ok {
		return nil, rest.NewError(http.StatusForbidden, "Connection token not valid")
	}
	var conn = mxconn.(*Conn)
	c.AddLogField("user", conn.login.UserName)
	return conn, nil
}

// Logout завершает сессию и удаляет ее из списка.
func (l *Conns) Logout(c *rest.Context) error {
	// получаем соединение из списка по авторизационному токену
	conn, err := l.authorize(c)
	if err != nil {
		return err
	}
	conn.Send("<logout/>", nil) // деавторизуем пользователя
	// после закрытия соединения оно само удалится из списка активных
	conn.Close() // закрываем соединение
	return nil   // ответ не требуется
}

// Commands обрабатывает команды к серверу MX.
func (l *Conns) Commands(c *rest.Context) error {
	// авторизуемся и получаем соединение из списка
	conn, err := l.authorize(c)
	if err != nil {
		return err
	}
	// формируем команду для сервера MX и структуру для разбора ответа
	var cmd, resp interface{}
	switch methodName := c.Param("cmd"); methodName {
	case "monitorStart":
		cmd = new(MonitorStartRequest)
		resp = new(MonitorStartResponse)
	case "monitorStop":
		cmd = new(MonitorStopRequest)
		resp = new(NamedResponse)
	case "monitorStartAb":
		cmd = new(MonitorStartAbRequest)
		resp = new(NamedResponse)
	case "monitorStopAb":
		cmd = new(MonitorStopAbRequest)
		resp = new(NamedResponse)

	default:
		return c.Error(http.StatusNotFound,
			fmt.Sprintf("Unsupported command %q", methodName))
	}

	// разбираем параметры и формируем команду
	if err := jsonBind(c.Request, cmd); err != nil {
		return err
	}
	// отсылаем команду на сервер
	if err := conn.Send(cmd, resp); err != nil {
		return httpError(err)
	}

	// обрабатываем ответ от сервера MX
	switch obj := resp.(type) {
	case *NamedResponse:
		resp = nil // для пустого ответа ничего не возвращаем
		_ = obj
	}

	return c.Write(resp) // возвращаем ответ
}

// Events позволяет получать события от сервера MX.
func (l *Conns) Events(c *rest.Context) error {
	conn, err := l.authorize(c)
	if err != nil {
		return err
	}
	log.Debug("sse connect",
		"user", conn.login.UserName,
		"count", conn.sse.Connected()+1)
	conn.sse.ServeHTTP(c.Response, c.Request)
	log.Debug("sse disconnect",
		"user", conn.login.UserName,
		"count", conn.sse.Connected())
	return nil
}
