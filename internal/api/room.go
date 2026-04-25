package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RoomMessage represents a message in a room.
type RoomMessage struct {
	ID          string `json:"id"`
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`
	Timestamp   string `json:"timestamp"`
}

// RoomMessagesResponse is the response from reading room messages.
type RoomMessagesResponse struct {
	Messages []RoomMessage `json:"messages"`
	LastID   string        `json:"lastId,omitempty"`
}

// RoomStatusResponse is the response from the room status endpoint.
type RoomStatusResponse struct {
	Exists       bool  `json:"exists"`
	Archived     bool  `json:"archived"`
	MessageCount int64 `json:"messageCount"`
}

// SendRoomMessageRequest is the request body for posting a room message.
type SendRoomMessageRequest struct {
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`
}

// SendRoomMessageResponse is the response after sending a message.
type SendRoomMessageResponse struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
}

// roomURL builds the base URL for room operations.
func (c *Client) roomURL(entityType, entityID string) string {
	return fmt.Sprintf("%s/api/v1/rooms/%s/%s", c.baseURL, entityType, entityID)
}

// SendRoomMessage posts a message to a room.
func (c *Client) SendRoomMessage(ctx context.Context, entityType, entityID string, msg SendRoomMessageRequest) (*SendRoomMessageResponse, error) {
	url := c.roomURL(entityType, entityID) + "/messages"
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal room message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create room message request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send room message: %w", err)
	}
	defer resp.Body.Close()

	var result SendRoomMessageResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("send room message: %w", err)
	}
	return &result, nil
}

// ReadRoomMessages reads messages from a room.
// Pass since="" for full history. Pass since=lastId for delta reads.
func (c *Client) ReadRoomMessages(ctx context.Context, entityType, entityID, since string, limit int, order string) (*RoomMessagesResponse, error) {
	url := c.roomURL(entityType, entityID) + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create read room request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	q := req.URL.Query()
	if since != "" {
		q.Set("since", since)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if order != "" {
		q.Set("order", order)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("read room messages: %w", err)
	}
	defer resp.Body.Close()

	var result RoomMessagesResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("read room messages: %w", err)
	}
	return &result, nil
}

// GetRoomStatus returns the status of a room.
func (c *Client) GetRoomStatus(ctx context.Context, entityType, entityID string) (*RoomStatusResponse, error) {
	url := c.roomURL(entityType, entityID) + "/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create room status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get room status: %w", err)
	}
	defer resp.Body.Close()

	var result RoomStatusResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get room status: %w", err)
	}
	return &result, nil
}

// ArchiveRoom archives a room. Reads still work, writes return 409.
func (c *Client) ArchiveRoom(ctx context.Context, entityType, entityID string) error {
	url := c.roomURL(entityType, entityID) + "/archive"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create archive room request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("archive room: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var raw json.RawMessage
		if err := decode(resp, &raw); err != nil {
			return fmt.Errorf("archive room: %w", err)
		}
	}
	return nil
}

// UnarchiveRoom unarchives a room, re-enabling writes.
func (c *Client) UnarchiveRoom(ctx context.Context, entityType, entityID string) error {
	url := c.roomURL(entityType, entityID) + "/unarchive"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create unarchive room request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unarchive room: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var raw json.RawMessage
		if err := decode(resp, &raw); err != nil {
			return fmt.Errorf("unarchive room: %w", err)
		}
	}
	return nil
}
