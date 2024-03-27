package models

import (
	"bytes"
	"github.com/golang/mock/gomock"
)

type Want struct {
	ExpectedCode        int
	ExpectedContentType string
	ExpectedLocation    string
	ExpectedShortURL    string
	ExpectedJSONBody    string
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
	OriginalURL OriginalURL
	ShortURL    ShortURL
	BatchResp   []BatchResponse
	List        []ShortenURL
	Error       error
}

type TestCase struct {
	Name string
	Ms   MockStorage
	Rp   RequestParams
	Want Want
}
