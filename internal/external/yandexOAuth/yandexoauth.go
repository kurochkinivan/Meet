package yandexoauth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type OAuthResponse struct {
	
}

func GetInfoByToken(ctx context.Context, OAuthToken string) error {
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
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
