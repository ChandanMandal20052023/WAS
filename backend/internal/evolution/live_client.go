package evolution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LiveClient struct {
	baseURL      string
	globalAPIKey string
	httpClient   *http.Client
}

func NewLiveClient(baseURL, globalAPIKey string) *LiveClient {
	return &LiveClient{
		baseURL:      baseURL,
		globalAPIKey: globalAPIKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *LiveClient) doRequest(ctx context.Context, method, endpoint string, body interface{}, apiKey string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("apikey", apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("evolution API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("evolution API error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}

func (c *LiveClient) CreateInstance(ctx context.Context, name, token string) error {
	payload := map[string]string{
		"name":  name,
		"token": token,
	}
	_, err := c.doRequest(ctx, http.MethodPost, "/instance/create", payload, c.globalAPIKey)
	return err
}

func (c *LiveClient) ConnectInstance(ctx context.Context, token, webhookURL string) error {
	payload := map[string]interface{}{
		"webhook": webhookURL,
		"events":  []string{"ALL"},
	}
	_, err := c.doRequest(ctx, http.MethodPost, "/instance/connect", payload, token)
	return err
}

func (c *LiveClient) GetQRCode(ctx context.Context, token string) (string, error) {
	respBytes, err := c.doRequest(ctx, http.MethodGet, "/instance/qr", nil, token)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse QR response: %w", err)
	}

	switch v := parsed.Data.(type) {
	case string:
		return v, nil
	case map[string]interface{}:
		if qr, ok := v["qrcode"].(string); ok {
			return qr, nil
		}
		if base64Str, ok := v["base64"].(string); ok {
			return base64Str, nil
		}
	}

	return "", errors.New("no valid QR code found in evolution response")
}

func (c *LiveClient) GetStatus(ctx context.Context, token string) (bool, string, error) {
	respBytes, err := c.doRequest(ctx, http.MethodGet, "/instance/status", nil, token)
	if err != nil {
		return false, "", err
	}

	var parsed struct {
		Message string `json:"message"`
		Data    struct {
			Connected bool   `json:"connected"`
			JID       string `json:"jid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return false, "", fmt.Errorf("failed to parse status response: %w", err)
	}

	return parsed.Data.Connected, parsed.Data.JID, nil
}

func (c *LiveClient) Disconnect(ctx context.Context, token string) error {
	_, err := c.doRequest(ctx, http.MethodPost, "/instance/disconnect", nil, token)
	return err
}

func (c *LiveClient) Reconnect(ctx context.Context, token string) error {
	_, err := c.doRequest(ctx, http.MethodPost, "/instance/reconnect", nil, token)
	return err
}

func (c *LiveClient) DeleteInstance(ctx context.Context, instanceID string) error {
	endpoint := fmt.Sprintf("/instance/delete/%s", instanceID)
	_, err := c.doRequest(ctx, http.MethodDelete, endpoint, nil, c.globalAPIKey)
	return err
}

func (c *LiveClient) PairPhone(ctx context.Context, token, phone string) (string, error) {
	payload := map[string]string{"phone": phone}
	respBytes, err := c.doRequest(ctx, http.MethodPost, "/instance/pair", payload, token)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", nil
	}
	return parsed.Data, nil
}

func (c *LiveClient) SendTextMessage(ctx context.Context, token, phone, message string) (string, error) {
	payload := map[string]interface{}{
		"number": phone,
		"text":   message,
	}
	respBytes, err := c.doRequest(ctx, http.MethodPost, "/send/text", payload, token)
	if err != nil {
		return "", fmt.Errorf("failed to send text message via evolution-go: %w", err)
	}

	var parsed struct {
		ID  string `json:"id"`
		Key struct {
			ID string `json:"id"`
		} `json:"key"`
	}
	_ = json.Unmarshal(respBytes, &parsed)
	msgID := parsed.ID
	if msgID == "" {
		msgID = parsed.Key.ID
	}
	if msgID == "" {
		msgID = fmt.Sprintf("WA-%d", time.Now().UnixNano())
	}
	return msgID, nil
}
