package models

import "bytes"

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
}

type RequestParams struct {
	Name             string
	Method           string
	URL              string
	Body             string
	JSONBody         *bytes.Buffer
	ConnectionString string
	Want             Want
}
