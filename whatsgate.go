package whatsgate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type transport struct {
	rt      http.RoundTripper
	xApiKey string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Api-Key", t.xApiKey)
	req.Header.Add("Content-type", "application/json")
	return t.rt.RoundTrip(req)
}

type Client struct {
	httpClient *http.Client
	url        string
	WhatsappID string
}

func NewClient(apiKey, whatsappID string) *Client {
	httpClient := &http.Client{Transport: &transport{rt: http.DefaultTransport, xApiKey: apiKey}}

	return &Client{httpClient: httpClient, url: "https://whatsgate.ru/api/v1", WhatsappID: whatsappID}
}

type MessageResponse struct {
	Result struct {
		Id          string `json:"_id"`
		Id1         string `json:"id"`
		Ack         int    `json:"ack"`
		HasMedia    bool   `json:"hasMedia"`
		MediaKey    string `json:"mediaKey"`
		Body        string `json:"body"`
		Type        string `json:"type"`
		Timestamp   int    `json:"timestamp"`
		From        string `json:"from"`
		FromName    string `json:"from_name"`
		To          string `json:"to"`
		IsForwarded bool   `json:"isForwarded"`
	} `json:"result"`
}

type MessageRequest struct {
	WhatsappID string    `json:"WhatsappID"`
	Async      bool      `json:"async"`
	Recipient  Recipient `json:"recipient"`
	Message    Message   `json:"message"`
}

type Recipient struct {
	Number string `json:"number"`
}

type Message struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

func (c *Client) SendMessage(recipientPhone, text string) (MessageResponse, error) {
	body, err := json.Marshal(MessageRequest{
		WhatsappID: c.WhatsappID,
		Async:      false,
		Recipient:  Recipient{Number: recipientPhone},
		Message:    Message{Type: "text", Body: text},
	})
	if err != nil {
		return MessageResponse{}, err
	}

	r, err := http.NewRequest("POST", c.url+"/send", bytes.NewBuffer(body))
	if err != nil {
		slog.Error(err.Error())
		return MessageResponse{}, err
	}
	r.Close = true

	req, err := c.httpClient.Do(r)
	if err != nil {
		slog.Error(err.Error())
		return MessageResponse{}, err
	}

	defer req.Body.Close()

	respBody, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error(err.Error())
		return MessageResponse{}, err
	}

	if req.StatusCode != http.StatusOK {
		return MessageResponse{}, errors.New(req.Status)
	}

	fmt.Println(string(respBody))

	var message MessageResponse

	err = json.Unmarshal(respBody, &message)
	if err != nil {
		slog.Error(err.Error())
		return MessageResponse{}, err
	}

	return message, nil
}
