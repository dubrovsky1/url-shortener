package models

import "bytes"

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
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
