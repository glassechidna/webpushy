package webpushy

type receiverHello struct {
	MessageType string `json:"messageType"`
	UseWebpush  bool   `json:"use_webpush"`
	UAID        string `json:"uaid,omitempty"`
}

type serviceHello struct {
	MessageType string `json:"messageType"`
	UseWebpush  bool   `json:"use_webpush"`
	UAID        string `json:"uaid,omitempty"`
	Status      int    `json:"status"`
}

type receiverRegister struct {
	MessageType string `json:"messageType"`
	ChannelID   string `json:"channelID"`
	Key         string `json:"key"`
}

type serviceRegister struct {
	MessageType  string `json:"messageType"`
	ChannelID    string `json:"channelID"`
	PushEndpoint string `json:"pushEndpoint"`
	Status       int    `json:"status"`
}

type serviceNotification struct {
	MessageType string            `json:"messageType"`
	ChannelID   string            `json:"channelID"`
	Version     string            `json:"version"`
	Data        string            `json:"data"`
	Headers     map[string]string `json:"headers"`
}

type receiverNotificationAck struct {
	MessageType string               `json:"messageType"`
	Updates     []notificationUpdate `json:"updates"`
}

type notificationUpdate struct {
	ChannelID string `json:"channelID"`
	Version   string `json:"version"`
	Code      int    `json:"code"`
}
