package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// CreateChatSession starts a new Story Forge Chat interview.
func (c *Client) CreateChatSession(ctx context.Context) (*ChatSession, error) {
	resp, err := c.raw.PostChatSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("create chat session: %w", err)
	}
	defer resp.Body.Close()

	var result ChatSession
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create chat session: %w", err)
	}
	return &result, nil
}

// ListChatSessions returns the user's Story Forge Chat sessions.
func (c *Client) ListChatSessions(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetChatSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list chat sessions: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list chat sessions: %w", err)
	}
	return json.RawMessage(body), nil
}

// GetChatSession returns a chat session with its messages.
func (c *Client) GetChatSession(ctx context.Context, id string) (*ChatSession, error) {
	resp, err := c.raw.GetChatSessionsId(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get chat session %s: %w", id, err)
	}
	defer resp.Body.Close()

	var result ChatSession
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get chat session %s: %w", id, err)
	}
	return &result, nil
}

// SendChatMessage sends a message and gets the AI response.
func (c *Client) SendChatMessage(ctx context.Context, id string, req ChatSendReq) (*ChatSendResp, error) {
	resp, err := c.raw.PostChatSessionsIdMessages(ctx, id, gen.PostChatSessionsIdMessagesJSONRequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("send chat message in %s: %w", id, err)
	}
	defer resp.Body.Close()

	var result ChatSendResp
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("send chat message in %s: %w", id, err)
	}
	return &result, nil
}

// FinalizeChatSession finalizes a chat interview and triggers story generation.
func (c *Client) FinalizeChatSession(ctx context.Context, id string) (*ChatFinalizeResp, error) {
	resp, err := c.raw.PostChatSessionsIdFinalize(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finalize chat session %s: %w", id, err)
	}
	defer resp.Body.Close()

	var result ChatFinalizeResp
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("finalize chat session %s: %w", id, err)
	}
	return &result, nil
}

// GetGenerationStatus polls the generation pipeline status for a story.
func (c *Client) GetGenerationStatus(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdGenerationStatus(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get generation status %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get generation status %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// GetStoryMeta returns the generated outline (story.md, characters.md, plot.md).
func (c *Client) GetStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.GetStoryIdMeta(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("get story meta %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get story meta %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// ApproveStoryMeta approves the generated outline and starts section generation.
func (c *Client) ApproveStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.PostStoryIdMetaApprove(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("approve story meta %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("approve story meta %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// RegenerateStoryMeta triggers a free retry on outline generation.
func (c *Client) RegenerateStoryMeta(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.PostStoryIdMetaRegenerate(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("regenerate story meta %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("regenerate story meta %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}

// ResumeGeneration resumes a failed or paused generation.
func (c *Client) ResumeGeneration(ctx context.Context, storyID string) (json.RawMessage, error) {
	resp, err := c.raw.PostStoryIdGenerationResume(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("resume generation %s: %w", storyID, err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("resume generation %s: %w", storyID, err)
	}
	return json.RawMessage(body), nil
}
