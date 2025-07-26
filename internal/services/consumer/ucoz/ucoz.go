package ucoz

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dghubble/oauth1"
)

type Client struct {
	authConfig *oauth1.Config
	authToken  *oauth1.Token
	httpClient *http.Client
	log        *slog.Logger
}

func New(log *slog.Logger, consumerKey, consumerSecret, token, tokenSecret string) *Client {
	authConfig := oauth1.NewConfig(consumerKey, consumerSecret)
	authToken := oauth1.NewToken(token, tokenSecret)

	client := authConfig.Client(context.TODO(), authToken)

	return &Client{
		authConfig: authConfig,
		authToken:  authToken,
		httpClient: client,
		log:        log,
	}

}
