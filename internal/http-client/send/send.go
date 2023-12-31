package send

import (
	"fmt"
	"net/http"
)

// Client структура возвращаемая для работы, клиент
type Client struct {
	URL    string
	Client *http.Client
}

// New возвращает эксземпляр клиента
func New(url string) (*Client, error) {
	var client Client
	client.URL = url
	client.Client = &http.Client{}
	return &client, nil
}

// Send отправляет метрику на сервер
func (c *Client) Send(typ string, name string, value string) error {

	req, err := http.NewRequest("POST", c.URL+"/update/"+typ+"/"+name+"/"+value, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}
	return nil
}
