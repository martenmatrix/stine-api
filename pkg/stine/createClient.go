package stine

import (
	"net/http"
)

type myTransport struct{}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	return http.DefaultTransport.RoundTrip(req)
}

func getClient() *http.Client {
	return &http.Client{
		Transport: &myTransport{},
	}
}
