package main

import (
	"encoding/xml"

	"github.com/mdigger/mx-http-proxy/mx"
)

type (
	// StatusMessageEvent (presence) описывает событие об изменении статуса
	// пользователя. Если поле From не задано (0), то значит, что это статус
	// текущего пользователя.
	StatusMessageEvent struct {
		XMLName  xml.Name `xml:"presence" json:"-"`
		From     mx.JID   `xml:"from,attr" json:"user,string,omitempty"`
		Status   string   `xml:"mxStatus,attr" json:"status,omitempty"`
		Presence string   `xml:"status,attr" json:"presence,omitempty"`
		Note     string   `xml:"presenceNote" json:"note,omitempty"`
	}
	// ServerMessageEvent (messageHist и message) описывает формат сообщения.
	//
	// FIX: в новом формате текст содержится внутри вложенного элемента text
	// для всех типов сообщений. В старом формате текст сообщения может быть
	// вложен непосредственно текстом в корневой элемент.
	ServerMessageEvent struct {
		XMLName      xml.Name `json:"-"` // FIX: т.к. поддерживает два типа
		ID           int      `xml:"msgId,attr" json:"id"`
		PersistentID int      `xml:"persistId,attr" json:"gid"`
		New          bool     `xml:"-" json:"new,omitempty"` // флаг, что сообщение не из истории
		From         mx.JID   `xml:"from,attr" json:"from,string"`
		FromName     string   `xml:"name,attr" json:"fromName,omitempty"`
		To           mx.JID   `xml:"toRecipId,attr" json:"to,string"`
		ToName       string   `xml:"toRecipName,attr" json:"toName,omitempty"`
		Group        mx.JID   `xml:"groupId,attr" json:"group,string,omitempty"`
		DID          string   `xml:"did" json:"did,omitempty"`
		RequestID    int      `xml:"reqId,attr" json:"req,omitempty"`
		Delivered    bool     `xml:"delivered,attr" json:"delivered"`
		Seen         bool     `xml:"seen,attr" json:"seen"`
		RecipType    string   `xml:"recipType,attr" json:"recipType,omitempty"` // [User, Server, Group]
		Type         string   `xml:"packetType,attr" json:"type,omitempty"`     // [Text, Binary, Conf]
		Timestamp    int64    `xml:"timestamp" json:"timestamp,omitempty"`
		Finished     bool     `xml:"finished,attr" json:"finished"`                   // только для конференций
		ContentState string   `xml:"contentState,attr" json:"contentState,omitempty"` // только для бинарных сообщений
		ContentSize  int      `xml:"contentSize,attr" json:"contentSize,omitempty"`   // только для бинарных сообщений
		Message      string   `xml:"text,chardata" json:"text"`                       // в новом формате это "text", в старом - ",chardata"
	}
	// ServerMessageHistoryEvent (messageHistChunks).
	// TODO: посмотреть с примерами как нужно правильно разбирать этот тип
	ServerMessageHistoryEvent struct {
		XMLName     xml.Name              `xml:"messageHistChunks" json:"-"`
		ChunkSize   int                   `xml:"chunkSize" json:"chunkSize"`
		Index       int                   `xml:"fromIndex" json:"index"`
		Total       int                   `xml:"totalCount" json:"total"`
		RecipType   string                `xml:"recipType,attr" json:"recipType,omitempty"` // [User, Server, Group]
		RecipTypeID mx.JID                `xml:"recipId,attr" json:"recip,string"`
		Messages    []*ServerMessageEvent `xml:"message" json:"messages,omitempty"`
	}
	// CallWithGlobal описывает информацию о звонке.
	CallWithGlobal struct {
		Call
		GlobalCallID mx.JID `xml:"globalCallID,omitempty" json:"global,string,omitempty"`
	}
	// DivertedEvent indicates that a call has been diverted from a
	// device and that the call is no longer present at the device.
	DivertedEvent struct {
		XMLName           xml.Name `xml:"DivertedEvent" json:"-"`
		MonitorCrossRefID int      `xml:"monitorCrossRefID" json:"monitor"`
		Call              Call     `xml:"connection" json:"call"`
		DivertingDevice   string   `xml:"divertingDevice>deviceIdentifier" json:"diverting"`
		NewDestination    string   `xml:"newDestination>deviceIdentifier" json:"to"`
		Cause             string   `xml:"cause" json:"cause"`
		CmdsAllowed       int      `xml:"cmdsAllowed" json:"allowed,omitempty"`
		CallTypeFlags     int      `xml:"callTypeFlags" json:"flags,omitempty"`
	}
	// Cad описывает формат данных с дополнительной информацией о звонке.
	Cad struct {
		Name  string `xml:"name,attr" json:"name"`
		Type  string `xml:"type,attr" json:"type"`
		Value string `xml:",chardata" json:"value"`
	}
	// DeliveredEvent indicates that a call is being presented
	// to a device in either the Ringing or Entering distribution
	// modes of the alerting state.
	DeliveredEvent struct {
		XMLName               xml.Name       `xml:"DeliveredEvent" json:"-"`
		MonitorCrossRefID     int            `xml:"monitorCrossRefID" json:"monitor"`
		Call                  CallWithGlobal `xml:"connection" json:"call"`
		AlertingDevice        string         `xml:"alertingDevice>deviceIdentifier" json:"alerting"`
		AlertingDisplayName   string         `xml:"alertingDisplayName" json:"alertingName"`
		NetworkCallingDevice  string         `xml:"networkCallingDevice>deviceIdentifier" json:"networkCalling"`
		CallingDevice         string         `xml:"callingDevice>deviceIdentifier" json:"calling"`
		CallingDisplayName    string         `xml:"callingDisplayName" json:"callingName"`
		CalledDevice          string         `xml:"calledDevice>deviceIdentifier" json:"called"`
		LastRedirectionDevice string         `xml:"lastRedirectionDevice>deviceIdentifier" json:"lastRedirection,omitempty"`
		LocalConnectionInfo   string         `xml:"localConnectionInfo" json:"localConnection"`
		Cause                 string         `xml:"cause" json:"cause"`
		CmdsAllowed           int            `xml:"cmdsAllowed" json:"cmdsAllowed,omitempty"`
		CallTypeFlags         int            `xml:"callTypeFlags" json:"callTypeFlags,omitempty"`
		Cads                  []*Cad         `xml:"cad,omitempty" json:"cads,omitempty"`
	}
	// EstablishedEvent indicates that a call has been answered
	// at a device or that a call has been connected to a device.
	EstablishedEvent struct {
		XMLName               xml.Name       `xml:"EstablishedEvent" json:"-"`
		MonitorCrossRefID     int            `xml:"monitorCrossRefID" json:"monitor"`
		Call                  CallWithGlobal `xml:"establishedConnection" json:"call"`
		AnsweringDevice       string         `xml:"answeringDevice>deviceIdentifier" json:"answering"`
		AnsweringDisplayName  string         `xml:"answeringDisplayName" json:"answeringName"`
		CallingDevice         string         `xml:"callingDevice>deviceIdentifier" json:"calling"`
		CallingDisplayName    string         `xml:"callingDisplayName" json:"callingName"`
		CalledDevice          string         `xml:"calledDevice>deviceIdentifier" json:"called"`
		LastRedirectionDevice string         `xml:"lastRedirectionDevice>deviceIdentifier" json:"lastRedirection,omitempty"`
		Cause                 string         `xml:"cause" json:"cause"`
		CmdsAllowed           int            `xml:"cmdsAllowed" json:"allowed,omitempty"`
		CallTypeFlags         int            `xml:"callTypeFlags" json:"flags,omitempty"`
		Cads                  []*Cad         `xml:"cad,omitempty" json:"cads,omitempty"`
	}
	// HeldEvent indicates that a call has been placed on hold.
	HeldEvent struct {
		XMLName             xml.Name `xml:"HeldEvent" json:"-"`
		MonitorCrossRefID   int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                Call     `xml:"heldConnection" json:"call"`
		HoldingDevice       string   `xml:"holdingDevice>deviceIdentifier" json:"holding"`
		LocalConnectionInfo string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause               string   `xml:"cause" json:"cause"`
		CmdsAllowed         int      `xml:"cmdsAllowed" json:"allowed,omitempty"`
		CallTypeFlags       int      `xml:"callTypeFlags" json:"flags,omitempty"`
	}
	// RecordingStateEvent описывает событие о записи звонка.
	RecordingStateEvent struct {
		XMLName           xml.Name `xml:"RecordingStateEvent" json:"-"`
		MonitorCrossRefID int      `xml:"monitorCrossRefID" json:"monitor"`
		Call              Call     `xml:"connection" json:"call"`
		Available         bool     `xml:"RecIsAvailable" json:"available"`
		Active            bool     `xml:"RecIsActive" json:"active"`
	}
	// ServiceInitiatedEvent indicates that a telephony service has been
	// initiated at a monitored device.
	ServiceInitiatedEvent struct {
		XMLName             xml.Name `xml:"ServiceInitiatedEvent" json:"-"`
		MonitorCrossRefID   int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                Call     `xml:"initiatedConnection" json:"call"`
		InitiatingDevice    string   `xml:"initiatingDevice>deviceIdentifier" json:"initiating"`
		LocalConnectionInfo string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause               string   `xml:"cause" json:"cause"`
	}
	// ConnectionClearedEvent indicates that a call has been cleared and no
	// longer exists.
	ConnectionClearedEvent struct {
		XMLName           xml.Name `xml:"ConnectionClearedEvent" json:"-"`
		MonitorCrossRefID int      `xml:"monitorCrossRefID" json:"monitor"`
		Call              Call     `xml:"droppedConnection" json:"call"`
		ReleasingDevice   string   `xml:"releasingDevice>deviceIdentifier" json:"releasing"`
		Cause             string   `xml:"cause" json:"cause"`
	}
	// OriginatedEvent indicates that a call is being attempted
	// from a device.
	OriginatedEvent struct {
		XMLName             xml.Name `xml:"OriginatedEvent" json:"-"`
		MonitorCrossRefID   int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                Call     `xml:"originatedConnection" json:"call"`
		CallingDevice       string   `xml:"callingDevice>deviceIdentifier" json:"calling"`
		CalledDevice        string   `xml:"calledDevice>deviceIdentifier" json:"called"`
		LocalConnectionInfo string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause               string   `xml:"cause" json:"cause"`
		CmdsAllowed         int      `xml:"cmdsAllowed" json:"allowed,omitempty"`
		CallTypeFlags       int      `xml:"callTypeFlags" json:"flags,omitempty"`
	}
	// NetworkReachedEvent indicates that a call has cut
	// through the switching sub-domain boundary to another network.
	NetworkReachedEvent struct {
		XMLName              xml.Name `xml:"NetworkReachedEvent" json:"-"`
		MonitorCrossRefID    int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                 Call     `xml:"outboundConnection" json:"call"`
		NetworkInterfaceUsed string   `xml:"networkInterfaceUsed>deviceIdentifier" json:"network"`
		CallingDevice        string   `xml:"callingDevice>deviceIdentifier" json:"calling"`
		CalledDevice         string   `xml:"calledDevice>deviceIdentifier" json:"called"`
		LocalConnectionInfo  string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause                string   `xml:"cause" json:"cause"`
	}
	// FailedEvent indicates that a call cannot be completed or a
	// connection has entered the Failed state.
	FailedEvent struct {
		XMLName             xml.Name `xml:"FailedEvent" json:"-"`
		MonitorCrossRefID   int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                Call     `xml:"failedConnection" json:"call"`
		CallingDevice       string   `xml:"callingDevice>deviceIdentifier" json:"calling"`
		CalledDevice        string   `xml:"calledDevice>deviceIdentifier" json:"called"`
		LocalConnectionInfo string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause               string   `xml:"cause" json:"cause"`
	}
	// RetrievedEvent indicates that a previously held call has been
	// retrieved.
	RetrievedEvent struct {
		XMLName             xml.Name `xml:"RetrievedEvent" json:"-"`
		MonitorCrossRefID   int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                Call     `xml:"retrievedConnection" json:"call"`
		RetrievingDevice    string   `xml:"retrievingDevice>deviceIdentifier" json:"retrieving"`
		LocalConnectionInfo string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause               string   `xml:"cause" json:"cause"`
		CmdsAllowed         int      `xml:"cmdsAllowed" json:"allowed,omitempty"`
		CallTypeFlags       int      `xml:"callTypeFlags" json:"flags,omitempty"`
	}
	// TransferredEvent indicates that an existing call has been
	// transferred to another device and the transferring device has
	// been dropped from the call.
	TransferredEvent struct {
		XMLName             xml.Name `xml:"TransferedEvent" json:"-"` // да, в xml название с ошибкой
		MonitorCrossRefID   int      `xml:"monitorCrossRefID" json:"monitor"`
		Call                Call     `xml:"primaryOldCall" json:"call"`
		TransferringDevice  string   `xml:"transferringDevice>deviceIdentifier" json:"transferring"`
		TransferredToDevice string   `xml:"transferredToDevice>deviceIdentifier" json:"to"`
		LocalConnectionInfo string   `xml:"localConnectionInfo" json:"localConnection"`
		Cause               string   `xml:"cause" json:"cause"`
	}
	// ParkedEvent описывает событие о парковке звонка.
	ParkedEvent struct {
		XMLName           xml.Name `xml:"callParkInfo" json:"-"`
		MonitorCrossRefID int      `xml:"monitorCrossRefID" json:"monitor"`
		ParkID            int      `xml:"parkID" json:"park"`
	}

	// CallInfo описывает информацию о записи в логе звонков.
	CallInfo struct {
		RecordID              int    `xml:"record_id" json:"id"`
		Missed                bool   `xml:"missed,attr" json:"missed"` // всегда отдавать
		Direction             string `xml:"direction,attr" json:"direction"`
		GCID                  string `xml:"gcid" json:"gcid"`
		ConnectTimestamp      int64  `xml:"connectTimestamp" json:"connectTimestamp,omitempty"`
		DisconnectTimestamp   int64  `xml:"disconnectTimestamp" json:"disconnectTimestamp,omitempty"`
		CallingPartyNo        string `xml:"callingPartyNo" json:"callingPartyNo"`
		OriginalCalledPartyNo string `xml:"originalCalledPartyNo" json:"originalCalledPartyNo"`
		FirstName             string `xml:"firstName" json:"firstName,omitempty"`
		LastName              string `xml:"lastName" json:"lastName,omitempty"`
		Extension             string `xml:"extension" json:"device,omitempty"`
		ServiceName           string `xml:"serviceName" json:"serviceName,omitempty"`
		ServiceExtension      string `xml:"serviceExtension" json:"serviceExtension,omitempty"`
		CallType              int    `xml:"callType" json:"callType,omitempty"`
		LegType               int    `xml:"legType" json:"legType,omitempty"`
		SelfLegType           int    `xml:"selfLegType" json:"selfLegType,omitempty"`
		MonitorType           int    `xml:"monitorType" json:"monitorType,omitempty"`
	}
	// CallLoginfo описывает событие со списком звонков.
	CallLoginfo struct {
		XMLName  xml.Name    `xml:"callloginfo" json:"-"`
		LogItems []*CallInfo `xml:"callinfo"  json:"callLog,omitempty"`
	}
)
