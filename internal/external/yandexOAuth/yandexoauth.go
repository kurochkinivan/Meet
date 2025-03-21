package yandexoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kurochkinivan/Meet/internal/apperr"
)

type YandexResponse struct {
	ID        string `json:"id"`
	PSUID     string `json:"psuid"`
	ClientID  string `json:"client_id"`
	Login     string `json:"login"`
	Birthday  string `json:"birthday"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RealName  string `json:"real_name"`
	Sex       string `json:"sex"`
	Phone     Phone  `json:"default_phone"`
}

type Phone struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
}

func ParseOAuthToken(ctx context.Context, OAuthToken string) (*YandexResponse, error) {
	req := &http.Request{
		Method: http.MethodGet,
		Header: http.Header{
			"Authorization": {
				fmt.Sprintf("OAuth %s", OAuthToken),
			},
		},
		URL: &url.URL{
			Scheme: "https",
			Host:   "login.yandex.ru",
			Path:   "/info",
			RawQuery: url.Values{
				"format": {
					"json",
				},
			}.Encode(),
		},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, apperr.WithHTTPStatus(fmt.Errorf("failed to send request to yandex oauth system: %w", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	var userData YandexResponse
	err = json.NewDecoder(resp.Body).Decode(&userData)
	if err != nil {
		return nil, apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return &userData, nil
}
