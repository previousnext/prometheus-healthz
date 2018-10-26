package prometheus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type Client struct {
	address string
}

func New(address string) (Client, error) {
	return Client{address}, nil
}

func (c Client) Rules() (RulesResponse, error) {
	var data RulesResponse

	resp, err := httpClient.Get(fmt.Sprintf("%s/api/v1/rules", c.address))
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}
