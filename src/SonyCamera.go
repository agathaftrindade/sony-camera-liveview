package liveview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type SonyCamera struct {
	URL             string
	LiveviewStarted bool
	LiveviewURL     *string
}

func NewSonyCamera(url string) SonyCamera {
	return SonyCamera{
		URL:             url,
		LiveviewStarted: false,
	}
}

func (s SonyCamera) DoAPIRequest(api string, method string, params ...string) ([]any, error) {
	if params == nil {
		params = make([]string, 0)
	}

	reqBody := map[string]any{
		"method":  method,
		"params":  params,
		"id":      1,
		"version": "1.0",
	}

	reqURL, _ := url.JoinPath(s.URL, api)
	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(reqURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseBody map[string]any
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}

	if responseBody["error"] != nil {
		return nil, fmt.Errorf("error in response: %v", responseBody["error"])
	}

	r := responseBody["result"].([]any)
	return r, nil
}

func (s *SonyCamera) StartLiveView() error {
	res, err := s.DoAPIRequest("camera", "startLiveview")
	if err != nil {
		return err
	}

	liveviewURL := res[0].(string)
	log.Printf("Started Liveview at %s\n", liveviewURL)
	s.LiveviewURL = &liveviewURL
	return nil
}
