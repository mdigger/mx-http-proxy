package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"sync"

	"github.com/mdigger/mx-http-proxy/mx"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
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
	connects.Add(-1)
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
	connects.Add(1)           // увеличиваем счетчик подключений
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
	conn, err := Connect(mxhost, login)
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

func init() {
	// регистрируем mimetype для указанного расширения, чтобы IE корректно
	// мог его проигрывать, потому что стандартный тип для него - audio/x-wav.
	mime.AddExtensionType(".wav", "audio/wave")
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
	var methodName = c.Param("cmd")
	switch methodName {
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
	case "makeCall":
		cmd = new(MakeCall) // используется хак для формирования XML
		resp = new(MakeCallResponse)
	case "clearConnection":
		cmd = new(ClearConnection)
		resp = new(NamedResponse)
	case "answerCall":
		cmd = new(AnswerCall)
		resp = new(NamedResponse)
	case "holdCall":
		cmd = new(HoldCall)
		resp = new(NamedResponse)
	case "parkCall":
		cmd = new(ParkCall)
		resp = new(NamedResponse)
	case "retriveCall":
		cmd = new(RetrieveCall)
		resp = new(NamedResponse)
	case "singleStepTranferCall":
		cmd = new(SingleStepTransferCall)
		resp = new(SingleStepTransferCallResponse)
	case "deflectCall":
		cmd = new(DeflectCall)
		resp = new(NamedResponse)
	case "transferCall":
		cmd = new(TransferCall)
		resp = new(NamedResponse)
	case "getCallLog":
		cmd = &GetCallLog{Type: "get", ID: "calllog", Timestamp: -1}
	case "assignDevice":
		cmd = new(AssignDevice)
		resp = new(NamedResponse)
	case "setCallMode":
		cmd = &SetCallMode{Type: "set", ID: "mode", Mode: "local",
			RingDelay: 20, VMDelay: 10}
	case "startRecording":
		cmd = new(StartRecording)
		resp = new(NamedResponse)
	case "stopRecording":
		cmd = new(StopRecording)
		resp = new(NamedResponse)
	case "vmGetList":
		cmd = new(MailGetListIncoming)
		resp = new(MailGetListIncomingResponse)
	case "vmDelete":
		cmd = new(MailDeleteIncoming)
		resp = new(NamedResponse)
	case "vmSetStatus":
		cmd = new(MailSetStatus)
		resp = new(NamedResponse)
	case "vmUpdateNote":
		cmd = new(UpdateVMNote)
		resp = new(NamedResponse)
	case "vmReceive":
		cmd = new(MailReceiveIncoming)
		resp = new(VMChunk)
	case "getAddressBook":
		cmd = &GetAddressBook{Type: "get", ID: "addressbook"}
	case "getServiceList":
		cmd = new(GetServiceList)
		resp = new(Services)
	case "setStatus":
		cmd = new(SetStatus)
	case "snapshotDevice":
		cmd = new(SnapshotDevice)
		resp = new(NamedResponse)
	case "setAgentState":
		cmd = new(SetAgentState)
		resp = new(NamedResponse)
	case "getAgentState":
		cmd = new(GetAgentState)
		resp = new(GetAgentStateResponse)

	default:
		return c.Error(http.StatusNotFound,
			fmt.Sprintf("Unsupported command %q", methodName))
	}

	commands.Add(methodName, 1) // увеличиваем счетчик команд

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
	case *MailGetListIncomingResponse:
		if len(obj.Mails) > 0 {
			resp = obj.Mails // отдаем только список голосовой почты
		} else {
			resp = nil
		}
	case *Services:
		if len(obj.Services) > 0 {
			resp = obj.Services // отдаем только список сервисов
		} else {
			resp = nil
		}
	case *VMChunk:
		// формируем строку с описанием типа содержимого
		var mimetype = mime.TypeByExtension("." + obj.Format)
		if mimetype == "" {
			mimetype = "application/octet-stream"
		}
		c.SetHeader("Content-Type", mimetype)
		c.SetHeader("Content-Disposition",
			fmt.Sprintf("attachment; filename=%q", obj.Name))
		// разрешаем отдавать ответ кусочками
		c.AllowMultiple = true
		// формируем команду для получения следующей порции файла
		var next = &struct {
			*MailReceiveIncoming
			Next string `xml:"nextChunk"`
		}{
			MailReceiveIncoming: cmd.(*MailReceiveIncoming),
		}
		var done = c.Request.Context().Done() // для отслеживания закрытия
	nextChunk:
		data, err := obj.Decode() // декодируем полученные данные
		if err != nil {
			log.Error("vmChunk decode error", err)
			return c.Error(http.StatusInternalServerError,
				"vmChunk decode error")
		}
		select {
		default: // отдаем кусочек данных пользователю
			// отсылаем кусок данных в ответ
			if err = c.Write(data); err != nil {
				log.Error("response vmChunk error", err)
				return c.Error(http.StatusInternalServerError,
					"Response vmChunk error")
			}
		case <-done: // пользователь закрыл соединение
			// отменяем загрузку данных
			if err = conn.Send(&MailCancelReceive{
				MailID:    next.MailID,
				MediaType: next.MediaType,
			}, new(NamedResponse)); err != nil {
				return httpError(err)
			}
			if err := c.Request.Context().Err(); err != nil {
				log.Error("response error", err)
			}
			return nil
		}
		// проверяем, что это не последний блок
		if obj.Number < obj.Total {
			// отсылаем команду на сервер для получения следующего блока
			if err := conn.Send(next, obj); err != nil {
				return httpError(err)
			}
			goto nextChunk // повторяем разбор и отдачу данных
		}
		return nil // ответ уже отослан
	case nil:
		switch obj := cmd.(type) {
		case *GetAddressBook: // адресная книга
			var contacts []*Contact // адресная книга
			var ablist = new(ABList)
		getAddressBook:
			// ожидаем события ablist от сервера MX со списком контактов
			if err := conn.WaitEvent("ablist", ablist); err != nil {
				return httpError(err)
			}
			// инициализируем адресную книгу
			if contacts == nil {
				contacts = make([]*Contact, 0, ablist.Size)
			}
			// заполняем адресную книгу полученными контактами
			contacts = append(contacts, ablist.Contacts...)
			// проверяем, что получена вся адресная книга
			if (ablist.Index+1)*50 < ablist.Size {
				// увеличиваем номер для получения следующей порции контактов
				obj.Index = ablist.Index + 1
				// отправляем запрос на получение следующей порции
				if err := conn.Send(obj, nil); err != nil {
					return httpError(err)
				}
				goto getAddressBook // запрашиваем следующую порцию контактов
			}
			resp = contacts // возвращаем контакты
		}
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
	ssecounter.Add(1)
	conn.sse.ServeHTTP(c.Response, c.Request)
	ssecounter.Add(-1)
	log.Debug("sse disconnect",
		"user", conn.login.UserName,
		"count", conn.sse.Connected())
	return nil
}
