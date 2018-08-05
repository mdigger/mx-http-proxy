package mx

import "encoding/xml"

// Response описывает ответ или событие, принимаемые от сервера MX.
type Response struct {
	ID   uint16 `json:"id"`   // идентификатор ответа
	Name string `json:"name"` // название события
	data []byte // не разобранное содержимое ответа
}

// String возвращает название события.
func (r *Response) String() string { return r.Name }

// Decode декодирует сообщение в указанный объект.
func (r *Response) Decode(v interface{}) error {
	return xml.Unmarshal(r.data, v)
}

// responseChan описывает канал для получения ответов.
type responseChan = chan *Response
