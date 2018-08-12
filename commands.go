package main

import (
	"encoding/base64"
	"encoding/xml"

	"github.com/mdigger/mx-http-proxy/mx"
)

// Содержит описание поддерживаемых команд и ответов на них.
//
// Для тех ответов, которые ненужно разбирать, используется простая "заглушка"
// в виде NamedResponse. Это позволяет дожидаться ответа от сервера MX и
// отслеживать ответы с ошибками.
type (
	// NamedResponse описывает пустой ответ, заставляющий ожидать
	// ответа на команду от сервера MX.
	NamedResponse struct {
		// XMLName xml.Name `json:"-"` // в принципе, нам это имя ничего не дает
	}

	// MonitorStartRequest command starts monitoring of CSTA events
	// (MonitorStartResponse).
	MonitorStartRequest struct {
		XMLName       xml.Name `xml:"MonitorStart" json:"-"`
		Object        string   `xml:"monitorObject>deviceObject,omitempty" json:"device"`
		UseConfEvents bool     `xml:"confEvents,omitempty" json:"conf,omitempty"`
	}
	// MonitorStartResponse описывает информацию, возвращаемую при старте
	// монитора.
	MonitorStartResponse struct {
		XMLName xml.Name `xml:"MonitorStartResponse" json:"-"`
		ID      int      `xml:"monitorCrossRefID" json:"monitor"`
		Voice   bool     `xml:"actualMonitorMediaClass>voice" json:"voice,omitempty"`
	}
	// MonitorStopRequest command stops monitoring of CSTA events
	// (empty MonitorStopResponse).
	MonitorStopRequest struct {
		XMLName xml.Name `xml:"MonitorStop" json:"-"`
		ID      int      `xml:"monitorCrossRefID" json:"monitor"`
	}
	// MonitorStartAbRequest запускает монитор отслеживания изменений в контактах
	MonitorStartAbRequest struct {
		XMLName xml.Name `xml:"MonitorStartAb" json:"-"`
	}
	// MonitorStopAbRequest останавливает монитор отслеживания изменений в контактах
	MonitorStopAbRequest struct {
		XMLName xml.Name `xml:"MonitorStopAb" json:"-"`
	}

	// MakeCall command allows setting up a call between a calling device and a
	// called device.
	MakeCall struct {
		XMLName       xml.Name      `xml:"MakeCall" json:"-"`
		CallingDevice CallingDevice `xml:"callingDevice" json:"device"` // специальная обработка
		Called        string        `xml:"calledDirectoryNumber" json:"to"`
		Group         string        `xml:"group,omitempty" json:"group,omitempty"`
		CallerID      string        `xml:"callerID,omitempty" json:"callerId,omitempty"`
		CallerName    string        `xml:"callerName,omitempty" json:"callerName,omitempty"`
	}
	// Call описывает информацию о звонке.
	Call struct {
		CallID   int    `xml:"callID" json:"id"`
		DeviceID string `xml:"deviceID" json:"device"`
		// GlobalCallID mx.JID `xml:"globalCallID,omitempty" json:"global,string,omitempty"`
	}
	// MakeCallResponse описывает информацию о звонке, возвращаемую сервером
	// на команду MakeCall.
	MakeCallResponse struct {
		Call   Call   `xml:"callingDevice" json:"call"`
		Device string `xml:"calledDevice,omitempty" json:"calledDevice,omitempty"`
	}
	// ClearConnection command releases all devices from an existing call.
	ClearConnection struct {
		XMLName xml.Name `xml:"ClearConnection" json:"-"`
		Call    Call     `xml:"connectionToBeCleared" json:"call"`
	}
	// AnswerCall command answers the inbound call.
	AnswerCall struct {
		XMLName xml.Name `xml:"AnswerCall" json:"-"`
		Call    Call     `xml:"callToBeAnswered" json:"call"`
	}
	// HoldCall command places a connected connection on hold at the same device.
	HoldCall struct {
		XMLName xml.Name `xml:"HoldCall" json:"-"`
		Call    Call     `xml:"callToBeHeld" json:"call"`
	}
	// ParkCall command moves a specified call at a device to a specified
	// (parked-to) destination.
	ParkCall struct {
		XMLName xml.Name `xml:"ParkCall" json:"-"`
		Call    Call     `xml:"parking" json:"call"`
	}
	// RetrieveCall command connects a specified held/parked connection.
	RetrieveCall struct {
		XMLName xml.Name `xml:"RetrieveCall" json:"-"`
		Call    Call     `xml:"callToBeRetrieved" json:"call"`
	}
	// SingleStepTransferCall command transfers an existing connection at a
	// device to another device.
	SingleStepTransferCall struct {
		XMLName xml.Name `xml:"SingleStepTransferCall" json:"-"`
		Call    Call     `xml:"activeCall" json:"call"`
		To      string   `xml:"transferredTo" json:"to"`
	}
	// SingleStepTransferCallResponse описывает информацию о перенаправленном
	// зноке, возвращаемую в ответ на команду SingleStepTransferCall.
	SingleStepTransferCallResponse struct {
		TransferredCall Call `xml:"transferredCall" json:"call"`
	}
	// DeflectCall command allows a call to be diverted to one or more
	// destinations.
	DeflectCall struct {
		XMLName        xml.Name `xml:"DeflectCall" json:"-"`
		Call           Call     `xml:"callToBeDiverted" json:"call"`
		NewDestination string   `xml:"newDestination" json:"to"`
	}
	// TransferCall объединяет звонки.
	TransferCall struct {
		XMLName    xml.Name `xml:"TransferCall" json:"-"`
		HeldCall   Call     `xml:"heldCall" json:"heldCall"`
		ActiveCall Call     `xml:"activeCall" json:"activeCall"`
	}
	// GetCallLog запрашивает получение списка звонков.
	GetCallLog struct {
		XMLName   xml.Name `xml:"iq" json:"-"`
		Type      string   `xml:"type,attr" json:"-"`
		ID        string   `xml:"id,attr" json:"-"`
		Timestamp int64    `xml:"timestamp,attr" json:"timestamp"`
	}

	// DeviceName описывает имя и тип устройства
	DeviceName struct {
		Type string `xml:"type,attr,omitempty" json:"type,omitempty"`
		ID   string `xml:",chardata" json:"name"`
	}
	// AssignDevice описывает имя назначаемого устройства.
	AssignDevice struct {
		XMLName xml.Name    `xml:"AssignDevice" json:"-"`
		Device  *DeviceName `xml:"deviceID,omitempty" json:"device,omitempty"`
	}
	// SetCallMode изменяет режим звонка.
	SetCallMode struct {
		XMLName   xml.Name `xml:"iq" json:"-"`
		Type      string   `xml:"type,attr" json:"-"`
		ID        string   `xml:"id,attr" json:"-"`
		Mode      string   `xml:"mode,attr" json:"mode,omitempty"`
		RingDelay uint32   `xml:"ringdelay,attr,omitempty" json:"ringDelay,omitempty"`
		VMDelay   uint32   `xml:"vmdelay,attr,omitempty" json:"vmDelay,omitempty"`
		From      string   `xml:"address,omitempty" json:"device,omitempty"`
	}
	// StartRecording запускает запись звонка на сервере
	StartRecording struct {
		XMLName xml.Name `xml:"StartRecording" json:"-"`
		Call    Call     `xml:"Call" json:"call"`
		GroupID string   `xml:"groupID,omitempty" json:"group,omitempty"`
	}
	// StopRecording останавливает запись звонка на сервере
	StopRecording struct {
		XMLName xml.Name `xml:"StopRecording" json:"-"`
		Call    Call     `xml:"Call" json:"call"`
		GroupID string   `xml:"groupID,omitempty" json:"group,omitempty"`
	}

	// MailGetListIncoming запрашивает получение списка голосовых сообщений.
	MailGetListIncoming struct {
		XMLName   xml.Name `xml:"MailGetListIncoming" json:"-"`
		UserID    mx.JID   `xml:"userId,omitempty" json:"user,string"`
		MediaType string   `xml:"mediaType,omitempty" json:"mediaType,omitempty"`
	}
	// VoiceMail описывает информацию о голосовом сообщении.
	VoiceMail struct {
		ID         int    `xml:"mailId" json:"id"`
		From       string `xml:"from,attr" json:"from"`
		FromName   string `xml:"fromName,attr" json:"fromName,omitempty"`
		CallerName string `xml:"callerName,attr" json:"callerName,omitempty"`
		To         string `xml:"to,attr" json:"to"`
		OwnerType  string `xml:"ownerType,attr" json:"ownerType"`
		MediaType  string `xml:"mediaType" json:"mediaType,omitempty"`
		Received   int64  `xml:"received" json:"received"`
		Duration   uint16 `xml:"duration" json:"duration,omitempty"`
		Read       bool   `xml:"read" json:"read,omitempty"`
		Note       string `xml:"note" json:"note,omitempty"`
	}
	// MailGetListIncomingResponse описывает возвращаемый список голосовых
	// сообщений.
	MailGetListIncomingResponse struct {
		Total uint16       `xml:"rowCount,attr" json:"total,omitempty"`
		Mails []*VoiceMail `xml:"mail" json:"mails,omitempty"`
	}
	// MailDeleteIncoming удаляет голосовое сообщение.
	MailDeleteIncoming struct {
		XMLName   xml.Name `xml:"MailDeleteIncoming" json:"-"`
		UserID    mx.JID   `xml:"userId,omitempty" json:"user,string"`
		MailID    int      `xml:"mailId" json:"id"`
		MediaType string   `xml:"mediaType,omitempty" json:"mediaType,omitempty"`
	}
	// MailSetStatus изменяет статус прочтения голосового сообщения.
	MailSetStatus struct {
		XMLName   xml.Name `xml:"MailSetStatus" json:"-"`
		UserID    mx.JID   `xml:"userId,omitempty" json:"user,string,omitempty"`
		MailID    int      `xml:"mailId" json:"id"`
		MediaType string   `xml:"mediaType,omitempty" json:"mediaType,omitempty"`
		Read      bool     `xml:"read" json:"read,omitempty"`
	}
	// UpdateVMNote позволяет изменить текст заметки голосового сообщения.
	UpdateVMNote struct {
		XMLName   xml.Name `xml:"UpdateVmNote" json:"-"`
		MailID    int      `xml:"mailId" json:"id"`
		MediaType string   `xml:"mediaType,omitempty" json:"mediaType,omitempty"`
		Note      string   `xml:"note" json:"note,omitempty"`
	}
	// MailReceiveIncoming запрашивает получение файла с голосовым сообщением.
	MailReceiveIncoming struct {
		XMLName   xml.Name `xml:"MailReceiveIncoming" json:"-"`
		MailID    int      `xml:"faxSessionID" json:"id"`
		MediaType string   `xml:"mediaType,omitempty" json:"mediaType,omitempty"`
	}
	// VMChunk описывает кусочек файла голосовой почты
	VMChunk struct {
		ID           int          `xml:"mailId,attr"`
		Number       int          `xml:"chunkNumber,attr"`
		Total        int          `xml:"totalChunks,attr"`
		ChunkSize    int          `xml:"chunkSize,attr"`
		Format       string       `xml:"fileFormat"`
		Name         string       `xml:"documentName"`
		MediaContent xml.CharData `xml:"mediaContent"`
	}
	// MailCancelReceive отменяет запрос получения файла с голосовой почтой
	MailCancelReceive struct {
		XMLName   xml.Name `xml:"MailCancelReceive" json:"-"`
		MailID    int      `xml:"mailId" json:"id"`
		MediaType string   `xml:"mediaType,omitempty" json:"mediaType,omitempty"`
	}

	// GetAddressBook ...
	GetAddressBook struct {
		XMLName xml.Name `xml:"iq" json:"-"`
		Type    string   `xml:"type,attr" json:"-"`
		ID      string   `xml:"id,attr" json:"-"`
		Index   uint     `xml:"index,attr" json:"-"`
		Sort    string   `xml:"sortmode,attr,omitempty" json:"sort,omitempty"`
	}
	// Contact описывает контактную информацю пользователя из адресной книги
	Contact struct {
		JID        mx.JID `xml:"jid,attr" json:"id,string"`
		Ext        string `xml:"businessPhone" json:"device"`
		FirstName  string `xml:"firstName" json:"firstName"`
		LastName   string `xml:"lastName" json:"lastName"`
		HomePhone  string `xml:"homePhone" json:"homePhone,omitempty"`
		CellPhone  string `xml:"cellPhone" json:"cellPhone,omitempty"`
		Email      string `xml:"email" json:"email,omitempty"`
		HomeSystem mx.JID `xml:"homeSystem" json:"homeSystem,string,omitempty"`
		DID        string `xml:"did" json:"did,omitempty"`
		ExchangeID string `xml:"exchangeId" json:"exchangeId,omitempty"`
	}
	// ABList описывает событие, в котором возвращается порция контаков из
	// серверной адресной книги.
	ABList struct {
		Size     uint       `xml:"size,attr" json:"size"`
		Index    uint       `xml:"index,attr" json:"index,omitempty"`
		Contacts []*Contact `xml:"abentry" json:"contacts,omitempty"`
	}
	// GetServiceList возвращает список с информацией о настройках сервисов.
	GetServiceList struct {
		XMLName xml.Name `xml:"GetServiceList" json:"-"`
	}
	// ServiceInfo описывает информацию о сервисе MX.
	ServiceInfo struct {
		ID         mx.JID `xml:"serviceId" json:"id,string"`
		Name       string `xml:"serviceName" json:"name"`
		Type       string `xml:"serviceType" json:"type"`
		Ext        string `xml:"extension" json:"device"`
		HomeSystem mx.JID `xml:"homeSystem" json:"homeSystem,string,omitempty"`
	}
	// Services содержит список конфигураций настроенных сервисов.
	Services struct {
		Services []*ServiceInfo `xml:"Service" json:"services"`
	}

	// SetStatus command sets specified presence.
	SetStatus struct {
		XMLName xml.Name `xml:"presence" json:"-"`
		Status  string   `xml:"status,attr" json:"presence"`
	}
	// SnapshotDevice command provides information about calls associated with a
	// specified device.
	SnapshotDevice struct {
		XMLName xml.Name `xml:"SnapshotDevice" json:"-"`
		Object  string   `xml:"snapshotObject" json:"device"`
	}

	// GetAgentState command provides information about specified Agent roles.
	GetAgentState struct {
		XMLName xml.Name `xml:"GetAgentState" json:"-"`
		Device  string   `xml:"device" json:"device"`
	}
	// GetAgentStateResponse описывает ответ с соотоянием агента. Возвращает
	// в ответ на запрос GetAgentState.
	GetAgentStateResponse struct {
		LoggedOnState bool `xml:"loggedOnState" json:"loggedOnState"`
		ReadyState    bool `xml:"readyState" json:"readyState"`
	}
	// SetAgentState command sets a logged state for specified Agent roles.
	SetAgentState struct {
		XMLName        xml.Name `xml:"SetAgentState" json:"-"`
		Device         string   `xml:"device" json:"device"`
		State          string   `xml:"requestedAgentState" json:"state"`
		AgentID        mx.JID   `xml:"agentID" json:"agent,string"`
		Password       string   `xml:"password" json:"password"`
		PhysicalDevice string   `xml:"physicalDevice" json:"physicalDevice"`
	}
)

// CallingDevice описывает идентификатор устройства, который специальным
// образом формирует XML, в зависимости от своего значения.
type CallingDevice string

// MarshalXML задает специальное правило формирования информации о вызывающем
// устройстве в формате XML.
func (s CallingDevice) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var device = &struct {
		TypeOfNumber string `xml:"typeOfNumber,attr,omitempty"`
		Device       string `xml:",chardata"`
	}{Device: string(s)}
	if s == "" {
		device.TypeOfNumber = "deviceID"
	}
	return e.EncodeElement(device, start)
}

// Decode разбирает содержимое и декодирует его из base64
func (ch *VMChunk) Decode() ([]byte, error) {
	var enc = base64.StdEncoding
	var data = make([]byte, enc.DecodedLen(len(ch.MediaContent)))
	n, err := enc.Decode(data, ch.MediaContent)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}
