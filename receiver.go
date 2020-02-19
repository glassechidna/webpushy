package webpushy

import (
	"context"
	"encoding/base64"
	"github.com/gorilla/websocket"
	"net/http"
)

type ReceiverOptions struct {
	Id           ReceiverId
	PublicKey    string
	Deserializer func([]byte) (interface{}, error)
	Headers      http.Header
	ServiceUrl   string
}

const DefaultServiceUrl = "wss://push.services.mozilla.com/"
const DefaultChannelId = "6319157f-095a-084d-9820-6f3cae005518"

type ReceiverId struct {
	Id       string
	Endpoint string
}

type Receiver struct {
	socket *websocket.Conn
	opts   *ReceiverOptions
	ch     chan interface{}
}

func NewReceiver(opts *ReceiverOptions) (*Receiver, error) {
	h := opts.Headers
	if h == nil {
		h = http.Header{}
	}

	if len(h.Get("User-Agent")) == 0 {
		h.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:73.0) Gecko/20100101 Firefox/73.0 webpushy/1.0")
	}

	if len(opts.ServiceUrl) == 0 {
		opts.ServiceUrl = DefaultServiceUrl
	}

	sock, _, err := websocket.DefaultDialer.DialContext(context.Background(), opts.ServiceUrl, h)
	if err != nil {
		return nil, err
	}

	r := &Receiver{
		socket: sock,
		opts:   opts,
		ch:     make(chan interface{}),
	}

	err = r.hello()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Receiver) Receive() <-chan interface{} {
	return r.ch
}

func (r *Receiver) Run(ctx context.Context) error {
	defer r.socket.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// no-op
		}

		snotification := serviceNotification{}
		err := r.socket.ReadJSON(&snotification)
		if err != nil {
			return err
		}

		err = r.socket.WriteJSON(&receiverNotificationAck{
			MessageType: "ack",
			Updates: []notificationUpdate{
				{
					ChannelID: snotification.ChannelID,
					Version:   snotification.Version,
					Code:      100,
				},
			},
		})
		if err != nil {
			return err
		}

		raw, err := base64.RawStdEncoding.DecodeString(snotification.Data)
		if err != nil {
			return err
		}

		msg, err := r.opts.Deserializer(raw)
		if err != nil {
			return err
		}

		r.ch <- msg
	}
}

func (r *Receiver) hello() error {
	firstTime := len(r.opts.Id.Id) == 0

	err := r.socket.WriteJSON(&receiverHello{
		MessageType: "hello",
		UseWebpush:  true,
		UAID:        r.opts.Id.Id,
	})
	if err != nil {
		return err
	}

	shello := serviceHello{}
	err = r.socket.ReadJSON(&shello)
	if err != nil {
		return err
	}

	r.opts.Id.Id = shello.UAID
	if !firstTime {
		return nil
	}

	err = r.socket.WriteJSON(&receiverRegister{
		MessageType: "register",
		ChannelID:   DefaultChannelId,
		Key:         r.opts.PublicKey,
	})
	if err != nil {
		return err
	}

	sregister := serviceRegister{}
	err = r.socket.ReadJSON(&sregister)
	if err != nil {
		return err
	}

	r.opts.Id.Endpoint = sregister.PushEndpoint
	return nil
}
