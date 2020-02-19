package webpushy

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type Sender struct {
	opts *SenderOptions
}

type SenderOptions struct {
	Keys       SenderKeys
	Identifier string
	Serializer func(interface{}) ([]byte, error)
}

type SenderKeys struct {
	Public  string
	Private string
}

func NewSender(opts *SenderOptions) *Sender {
	return &Sender{opts: opts}
}

func (s *Sender) Send(endpoint string, payload interface{}, ttl time.Duration) error {
	body, err := s.opts.Serializer(payload)
	if err != nil {
		return err
	}

	req, err := makeRequest(s.opts, endpoint, ttl, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return &SendError{Response: resp}
	}

	return nil
}

type SendError struct {
	Response *http.Response
}

func (s *SendError) Error() string {
	return fmt.Sprintf("Got HTTP response status: %s", s.Response.Status)
}
