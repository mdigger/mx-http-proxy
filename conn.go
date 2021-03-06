package main

import (
	"github.com/Connector73/mx-http-proxy/mx"
	"github.com/mdigger/log"
	"github.com/mdigger/sse"
)

// Conn описывает соединение с сервером MX.
type Conn struct {
	*mx.Conn             // соединение с сервером MX
	login    mx.Login    // информация для авторизации
	sse      *sse.Server // брокер для отсылки событий
}

// Connect подключается к серверу MX и авторизует пользователя.
func Connect(host string, login *mx.Login) (*Conn, error) {
	conn, err := mx.Connect(host, login) // устанавливаем соединение и авторизуемся
	if err != nil {
		return nil, err
	}
	var mxconn = &Conn{
		Conn:  conn,
		login: *login,          // копируем данные о логине
		sse:   new(sse.Server), // инициализируем брокера для отправики событий
	}
	go mxconn.reading() // запускаем обработчик входящих событий от сервера MX
	return mxconn, nil
}

// Close закрывает соедиение с сервером MX.
func (c *Conn) Close() error {
	c.sse.Event("", "close", nil) // отправляем уведомление о закрытии соединения
	return c.Conn.Close()         // закрываем соединение
}

// reading ожидает события от сервера MX и передает их в виде Server-Sent Events
func (c *Conn) reading() {
	// получаем событие
	for event := range c.Events() {
		var name = event.Name
		// выбираем формат описания события в зависимости от его имени
		var obj interface{}
		switch name {
		case "CSTAErrorCode":
			obj = new(mx.CSTAError)
		case "Logout":
			obj = new(mx.ErrLogout)
		case "presence":
			obj = new(StatusMessageEvent)
		case "message":
			obj = &ServerMessageEvent{New: true}
		case "messageHist":
			obj = &ServerMessageEvent{New: false}
		// TODO: messageHistChunks
		// case "messageHistChunks":
		// 	obj = new(ServerMessageHistoryEvent)
		case "DivertedEvent":
			obj = new(DivertedEvent)
		case "DeliveredEvent":
			obj = new(DeliveredEvent)
		case "EstablishedEvent":
			obj = new(EstablishedEvent)
		case "HeldEvent":
			obj = new(HeldEvent)
		case "RecordingStateEvent":
			obj = new(RecordingStateEvent)
		case "ServiceInitiatedEvent":
			obj = new(ServiceInitiatedEvent)
		case "ConnectionClearedEvent":
			obj = new(ConnectionClearedEvent)
		case "OriginatedEvent":
			obj = new(OriginatedEvent)
		case "NetworkReachedEvent":
			obj = new(NetworkReachedEvent)
		case "FailedEvent":
			obj = new(FailedEvent)
		case "RetrievedEvent":
			obj = new(RetrievedEvent)
		case "TransferedEvent", "TransferredEvent":
			obj = new(TransferredEvent)
			name = "TransferredEvent" // исправляем ошибку в написании
		case "callParkInfo":
			obj = new(ParkedEvent)
		case "callloginfo":
			obj = new(CallLoginfo)
		case "ablist": // игнорируем
			continue
		default:
			log.Warn("unsupported event", "name", name)
			continue
		}

		events.Add(name, 1)
		// разбираем XML с описанием события
		if err := event.Decode(obj); err != nil {
			log.Error("decode event error", err)
			continue
		}
		// отправляем информацию о событии в соответствующий обработчик
		c.sse.Event("", name, obj)
		log.Debug("sse",
			"user", c.login.UserName,
			"event", name)
	}
	c.sse.Close() // закрываем по окончании
}
