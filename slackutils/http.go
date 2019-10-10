package slackutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

type HttpClient interface {
	PostJsonContent(ctx context.Context, message, url string)
}

type DefaultHttpClient struct{}

func (*DefaultHttpClient) PostJsonContent(ctx context.Context, message, url string) {
	type Payload struct {
		Text string `json:"text"`
	}

	data := Payload{
		Text: message,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Notifying slack failed", zap.Error(err))
		return
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.Post(url, "application/json", body)
	defer req.Body.Close()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Notifying slack failed", zap.Error(err))
		return
	}
}
