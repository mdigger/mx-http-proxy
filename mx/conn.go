package mx

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ConnectionTimeout задает максимальное время ожидания установки соединения
	// с сервером.
	ConnectionTimeout = time.Second * 20
	// ReadTimeout задает время по умолчанию дл ожидания ответа от сервера.
	ReadTimeout = time.Second * 7
	// KeepAliveDuration задает интервал для отправки keep-alive сообщений в
	// случае простоя.
	KeepAliveDuration = time.Minute
	// Logger используется для вывода информации о командах и событиях MX
	Logger = log.New(ioutil.Discard, "", 0)
)

// Conn описывает соединение с сервером MX.
type Conn struct {
	Info                   // информация об авторизации и сервере MX
	conn      net.Conn     // сокетное соединение с сервером MX
	mu        sync.RWMutex // блокировщик изменения таймера
	keepAlive *time.Timer  // таймер для отсылки keep-alive сообщений
	counter   uint32       // счетчик отосланных команд
	waitResp  sync.Map     // список каналов для обработки ответов
	waitEvent sync.Map     // список именованных событий
	events    responseChan // канал с событиями
	logPrefix string       // префикс для вывода в лог
	closer    func(error)  // функция, вызываемая при закрытии соединения
}

// Connect устанавливает соединение с сервером MX.
func Connect(host string, login Login) (*Conn, error) {
	// добавляем порт по умолчанию, если он не задан в адресе сервера
	if _, _, err := net.SplitHostPort(host); err != nil {
		var err, ok = err.(*net.AddrError)
		if ok && err.Err == "missing port in address" {
			host = net.JoinHostPort(host, "7778")
		}
	}
	// устанавливаем защищенное соединение с сервером MX
	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout:   ConnectionTimeout,
			KeepAlive: KeepAliveDuration,
		},
		"tcp", host,
		&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	// инициализируем описание соединения
	var mx = &Conn{
		conn:      conn,
		logPrefix: fmt.Sprintf("%s: ", login.UserName),
	}
	// авторизуем соединение
	if err := login.authorize(mx); err != nil {
		mx.Close()
		return nil, err
	}

	// запускаем отправку keepAlive сообщений
	mx.keepAlive = time.AfterFunc(KeepAliveDuration, mx.sendKeepAlive)
	// запускаем процесс чтения ответов от сервера
	go func(mx *Conn) {
		var (
			resp *Response // команда или событие
			err  error
		)
		for {
			resp, err = mx.read() // читаем ответ на команду или событие
			if err != nil {
				break
			}
			if resp.ID < 9999 {
				// проверяем, что ответом на эту команду мы интересуемся
				if respChan, ok := mx.waitResp.Load(resp.ID); ok {
					respChan.(responseChan) <- resp
				}
				continue // дальше только обработка событий
			}
			// отправляем информацию о событии в обработчик
			mx.mu.RLock()
			if mx.events != nil {
				mx.events <- resp
			}
			mx.mu.RUnlock()
			// проверяем, что мы мониторим данное событие
			if ch, ok := mx.waitEvent.Load(resp.Name); ok {
				ch.(responseChan) <- resp // отправляем информацию
			}
			// проверяем на принудительное завершение сессии
			if resp.Name == "Logout" {
				var logout = new(ErrLogout)
				if err = resp.Decode(logout); err == nil {
					err = logout
				}
				break // закрываем соединение
			}
		}
		// соединение закрыто
		mx.mu.Lock()
		mx.keepAlive.Stop() // останавливаем отправку keepAlive сообщений
		if mx.events != nil {
			close(mx.events)
		}
		if mx.closer != nil {
			mx.closer(err) // выполнить функцию при закрытии соединения
		}
		mx.mu.Unlock()
	}(mx)

	return mx, nil
}

// buffers используется как пул буферов для формирования новых команд,
// отправляемых на сервер.
var buffers = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}

// send отправляет команду на сервер. Возвращает идентификатор отправленной
// команды.
func (c *Conn) send(cmd interface{}) (uint16, error) {
	// преобразуем данные команды к формату XML
	var xmlData []byte
	switch data := cmd.(type) {
	case string:
		xmlData = []byte(data)
	case []byte:
		xmlData = data
	default:
		var err error
		if xmlData, err = xml.Marshal(cmd); err != nil {
			return 0, err
		}
	}
	// увеличиваем счетчик отправленных команд (не больше 4-х цифр)
	var counter = atomic.AddUint32(&c.counter, 1)
	// 9999 зарезервирован для событий в ответах сервера
	if counter > 9998 {
		counter = 1
		atomic.StoreUint32(&c.counter, 1)
	}
	// формируем бинарное представление команды и отправляем ее
	var buf = buffers.Get().(*bytes.Buffer)
	buf.Reset()             // сбрасываем полученный из пула буфер
	buf.Write([]byte{0, 0}) // первые два байта сообщения нули
	// записываем длину сообщения
	binary.Write(buf, binary.BigEndian, uint16(len(xmlData)+8))
	fmt.Fprintf(buf, "%04d", counter) // идентификатор команды
	buf.Write(xmlData)                // содержимое команды
	_, err := buf.WriteTo(c.conn)     // отсылаем команду
	buffers.Put(buf)                  // освобождаем буфер
	if err != nil {
		return 0, err
	}
	// выводим команду в лог
	Logger.Printf("%s<- %04d %s", c.logPrefix, counter, xmlData)
	// откладываем посылку keepAlive
	c.mu.Lock()
	if c.keepAlive != nil {
		c.keepAlive.Reset(KeepAliveDuration)
	}
	c.mu.Unlock()
	return uint16(counter), nil
}

// Send формирует и отправляет команду на сервер. Ожидает ответ, если параметр
// resp не nil.
func (c *Conn) Send(cmd interface{}, resp interface{}) error {
	var id, err = c.send(cmd) // отправлем команду
	if err != nil {
		return err
	}
	if resp == nil {
		return nil // если ответа не ожидается, то на этом все
	}
	// ожидаем либо ответа на нашу команду, либо истечения времени
	var (
		event  *Response                    // ожидаемый ответ
		respCh = make(responseChan, 1)      // канал с ответом
		timer  = time.NewTimer(ReadTimeout) // таймер ожидания
	)
	// сохраняем канал для отдачи ответа в ассоциации с идентификатором
	// отосланной команды
	c.waitResp.Store(id, respCh)
	// ожидаем ответа или истечения времени ожидания
	select {
	case event = <-respCh: // получен ответ от сервера
		// отдельно обрабатываем ответы с описанием ошибки
		if event.Name == "CSTAErrorCode" {
			cstaError := new(CSTAError)
			if err = event.Decode(cstaError); err == nil {
				err = cstaError // подменяем ошибку
			}
			event = nil // сбрасываем данные
		}
	case <-timer.C: // превышено время ожидания ответа от сервера
		err = ErrTimeout
	}
	c.waitResp.Delete(id) // удаляем из списка ожидания
	close(respCh)         // закрываем канал
	timer.Stop()          // останавливаем таймер по окончании
	if err != nil {
		return err
	}
	return event.Decode(resp)
}

// reBinaryReplace используется для замены вывода в лог бинарных данных
var reBinaryReplace = regexp.MustCompile("(?s)<mediaContent>.+</mediaContent>")

// read читает ответ на команду или событие, возвращаемое сервером MX.
func (c *Conn) read() (*Response, error) {
	var header = make([]byte, 8) // буфер для разбора заголовка ответа
	// читаем заголовок сообщения
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, err
	}
	// разбираем номер команды ответа (для событий - 9999)
	id, err := strconv.ParseUint(string(header[4:]), 10, 16)
	if err != nil {
		return nil, err
	}
	// вычисляем длину ответа
	var length = binary.BigEndian.Uint16(header[2:4]) - 8
	// читаем данные с ответом
	var data = make([]byte, length)
	if _, err = io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}
	// разбираем xml ответа
	var xmlDecoder = xml.NewDecoder(bytes.NewReader(data))
readToken:
	var offset = xmlDecoder.InputOffset() // начало элемента xml
	token, err := xmlDecoder.Token()      // разбираем токен xml
	if err != nil {
		return nil, err
	}
	// пропускаем все до корневого элемента XML
	startToken, ok := token.(xml.StartElement)
	if !ok {
		goto readToken // игнорируем все до корневого элемента XML.
	}
	// формируем ответ
	var resp = &Response{
		ID:   uint16(id),            // идентификатор команды
		Name: startToken.Name.Local, // название элемента
		data: data[offset:],         // неразобранные данные с ответом,
	}

	// подменяем бинарные данные перед выводом в лог
	var logData = reBinaryReplace.ReplaceAll(resp.data,
		[]byte("<mediaContent>[bin data]</mediaContent>"))
	var logID string
	if id < 9999 {
		logID = fmt.Sprintf("-> %04d ", id)
	}
	// выводим в лог
	Logger.Print(c.logPrefix, logID, string(logData))

	return resp, nil
}

// Close закрывает соединение с сервером.
func (c *Conn) Close() error {
	return c.conn.Close() // закрываем соединение
}

// SetCloser задает функцию, которая автоматически выполнится при закрытии
// соединения с сервером MX. В качестве параметра функции будет передана
// ошибка соединения. В случае планового закрытия соединения ошибка будет
// пустой.
func (c *Conn) SetCloser(f func(error)) {
	c.mu.Lock()
	c.closer = f
	c.mu.Unlock()
}

// sendKeepAlive отсылает на сервер keep-alive сообщение для поддержки активного
// соединения и взводит таймер для отправки следующего.
func (c *Conn) sendKeepAlive() {
	// отправляем keepAlive сообщение на сервер
	// чтобы не создавать его каждый раз, а оно не изменяется, оно создано
	// заранее и приведено в бинарный вид команды
	_, err := c.conn.Write([]byte{0x00, 0x00, 0x00, 0x15, 0x30, 0x30, 0x30,
		0x30, 0x3c, 0x6b, 0x65, 0x65, 0x70, 0x61, 0x6c, 0x69, 0x76, 0x65, 0x20,
		0x2f, 0x3e})
	if err == nil {
		// взводим таймер отправки следующего keepAlive сообщения
		c.mu.Lock()
		c.keepAlive.Reset(KeepAliveDuration)
		c.mu.Unlock()
	}
}
