package mx

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"time"
)

// JID задает формат уникального идентификатора MX.
type JID = uint64

// Info описывает информацию об авторизованном пользователе и сервере MX.
type Info struct {
	// информация об авторизованном пользователе
	UserID       JID    `xml:"userId,attr,omitempty" json:"user,string,omitempty"`
	Ext          string `xml:"ext,attr,omitempty" json:"device,omitempty"`
	SoftPhonePwd string `xml:"softPhonePwd,attr,omitempty" json:"softPhonePwd,omitempty"`
	// информация о сервере MX
	APIVersion    uint16 `xml:"apiversion,attr,omitempty" json:"api,omitempty"`
	MXID          string `xml:"sn,attr,omitempty" json:"mx,omitempty"`
	Capab         string `xml:"mxCapab,attr,omitempty" json:"capab,omitempty"`
	MaxFileSizeMb int    `xml:"maxMsgFileSizeMb,attr,omitempty" json:"maxMsgFileSizeMb,omitempty"`
}

// Login описывает параметры для авторизации.
type Login struct {
	UserName   string `xml:"userName" json:"login"`
	Password   string `xml:"-" json:"password"`
	Type       string `xml:"type,attr,omitempty" json:"type"`
	Platform   string `xml:"platform,attr,omitempty" json:"platform,omitempty"`
	Version    string `xml:"version,attr,omitempty" json:"version,omitempty"`
	ClientType string `xml:"clientType,attr,omitempty" json:"clientType,omitempty"` // "", "Mobile", "Desktop", "CRM"
	ServerType string `xml:"serverType,attr,omitempty" json:"serverType,omitempty"` // "", "SMS"
	LoginCapab string `xml:"loginCapab,attr,omitempty" json:"loginCapab,omitempty"` // "Audio|Video|Im|911Support|BinIm|WebChat"
	MediaCapab string `xml:"mediaCapab,attr,omitempty" json:"mediaCapab,omitempty"` // "Voicemail|Fax|CallRec"
	ABNotify   bool   `xml:"abNotify,attr,omitempty" json:"abNotify,omitempty"`
	Forced     bool   `xml:"forced,attr,omitempty" json:"forced,omitempty"`
	APIVersion int32  `xml:"apiVersion,attr,omitempty" json:"apiVersion,omitempty"`
}

// HashedPassword возвращает пароль в виде хеша.
func (l Login) HashedPassword() string {
	var passwd = l.Password // пароль пользователя для авторизации
	// эвристическим способом проверяем, что пароль похож на base64 от sha1.
	if len(passwd) > 4 && passwd[len(passwd)-1] == '\n' {
		data, err := base64.StdEncoding.DecodeString(passwd[:len(passwd)-1])
		if err == nil && len(data) == sha1.Size {
			return passwd // пароль уже представлен в виде хеша
		}
	}
	// вычисляем хеш от пароля
	var pwdHash = sha1.Sum([]byte(passwd))
	// возвращаем хешированный пароль
	return base64.StdEncoding.EncodeToString(pwdHash[:]) + "\n"
}

// authorize авторизует установленное соединение.
func (l Login) authorize(c *Conn) error {
	// формируем команду для авторизации пользователя
	var loginCommand = &struct {
		XMLName  xml.Name `xml:"loginRequest"`
		Login             // копируем все параметры логина
		Password string   `xml:"pwd"`
	}{
		Login:    l,
		Password: l.HashedPassword(),
	}
send:
	cmdID, err := c.send(loginCommand) // отправляем команду
	if err != nil {
		return err
	}
	// ожидаем ответа или истечения времени ожидания
	var (
		timer = time.NewTimer(ReadTimeout) // таймер ожидания
		resp  *Response
	)
	for {
		select {
		case <-timer.C: // превышено время ожидания ответа от сервера
			err = ErrTimeout
		default:
			resp, err = c.read() // читаем ответ
		}
		timer.Stop() // в любом случае останавливаем
		if err != nil {
			return err
		}
		// проверяем, что это ответ на наш запрос
		if resp.ID == cmdID {
			break
		}
		timer.Reset(ReadTimeout) // откладываем ожидание и читаем следующий ответ
	}
	// разбираем ответ авторизации
	switch resp.Name {
	case "loginResponce": // да, именно так, с ошибкой в названии команды
		// разбираем информацию об авторизации
		var info = new(Info)
		if err := resp.Decode(info); err != nil {
			return err
		}
		c.Info = *info // сохраняем в контексте соединения
		return nil
	case "loginFailed": // ошибка авторизации
		// разбираем ошибку
		var loginError = new(LoginError)
		if err := resp.Decode(loginError); err != nil {
			return err
		}
		// если ошибка связана с тем, что пароль передан в виде хеш,
		// то повторяем попытку авторизации с паролем в открытом виде
		if (loginError.Code == 2 || loginError.Code == 4) &&
			loginError.APIVersion > 2 &&
			(loginCommand.Password != l.Password) {
			loginCommand.Password = l.Password
			goto send // повторяем с открытым паролем
		}
		return loginError // возвращаем ошибку авторизации
	default: // неожиданный ответ
		return fmt.Errorf("unknown mx login response %q", resp.Name)
	}
}
