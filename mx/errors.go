package mx

import "fmt"

// CSTAError описывает ошибку CSTA.
type CSTAError struct {
	Message string `xml:",any" json:"message"`
}

// Error возвращает текстовое описание ошибки.
func (e *CSTAError) Error() string { return e.Message }

// TimeoutError описывает ошибку ожидания ответа от сервера MX.
type TimeoutError struct{}

// Error возвращает текст описания ошибки.
func (TimeoutError) Error() string { return "MX response timeout" }

// Timeout возвращает true.
func (TimeoutError) Timeout() bool { return true }

// Temporary возвращает false.
func (TimeoutError) Temporary() bool { return false }

// ErrTimeout возвращается когда ответ от сервера на команду не получен за время
// ReadTimeout.
var ErrTimeout error = new(TimeoutError)

// ErrLogout описывает сообщение о закрытии соединения сервером.
type ErrLogout struct {
	Mode string `xml:"mode,attr" json:"mode,omitempty"`
}

// Error возвращает текстовое описание ошибки принудительного закрытия
// соединения.
func (e *ErrLogout) Error() string { return fmt.Sprintf("logout: %s", e.Mode) }

// LoginError описывает ошибку авторизации.
type LoginError struct {
	Code       uint8  `xml:"code,attr" json:"code,omitempty"`
	APIVersion uint16 `xml:"apiversion,attr" json:"api,omitempty"`
	Message    string `xml:"" json:"message"`
}

func (e *LoginError) Error() string { return e.Message }
