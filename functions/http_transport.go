package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type HttpFunctionsClient struct {
	URL string
}

func NewHttpTransport(url string) Transport {
	return func(ctx context.Context, req *FunctionsRuntimeRequest) (*FunctionsRuntimeResponse, error) {
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(b))
		if err != nil {
			return nil, err
		}

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var response FunctionsRuntimeResponse

		err = json.Unmarshal(b, &response)
		if err != nil {
			return nil, errors.New("invalid json: " + string(b))
		}

		return &response, nil
	}
}
