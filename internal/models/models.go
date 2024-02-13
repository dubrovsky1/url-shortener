package models

import (
	"bytes"
	"github.com/golang/mock/gomock"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	URL           string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Want struct {
	ExpectedCode        int
	ExpectedContentType string
	ExpectedLocation    string
	ExpectedShortURL    string
}

type RequestParams struct {
	Method           string
	URL              string
	Body             string
	JSONBody         *bytes.Buffer
	ConnectionString string
}

type MockStorage struct {
	Ctrl        *gomock.Controller
	OriginalURL string
	ShortURL    string
	Error       error
}

type TestCase struct {
	Name string
	Ms   MockStorage
	Rp   RequestParams
	Want Want
}
