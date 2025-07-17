package infrastructure

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// NotifyPredictResult gửi webhook tới UI với payload là map bất kỳ
func NotifyPredictResult(webhookURL string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
