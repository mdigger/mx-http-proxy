package mx

import (
	"time"
)

// Events возвращает канал для уведомления о событиях от сервера MX.
func (c *Conn) Events() <-chan *Response {
	if c.events == nil {
		c.mu.Lock()
		c.events = make(responseChan, 1)
		c.mu.Unlock()
	}
	return c.events
}

// WaitEvent ожидает событие с заданным именем и возвращает его.
func (c *Conn) WaitEvent(name string, resp interface{}) error {
	var (
		event  *Response                    // ожидаемый ответ
		respCh = make(responseChan, 1)      // канал с ответом
		timer  = time.NewTimer(ReadTimeout) // таймер ожидания
		err    error
	)
	c.waitEvent.Store(name, respCh)
	select {
	case event = <-respCh: // получен ответ от сервера
	case <-timer.C: // превышено время ожидания ответа от сервера
		err = ErrTimeout
	}
	c.waitEvent.Delete(name)
	close(respCh)
	timer.Stop() // останавливаем таймер по окончании
	if err != nil {
		return err
	}
	return event.Decode(resp)
}
