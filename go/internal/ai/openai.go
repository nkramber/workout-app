package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// OpenAIEndpoint is the endpoint of the Responses API of OpenAI.
const OpenAIEndpoint = "https://api.openai.com/v1/responses"

// maxReply is the largest body of a reply that the provider reads.
const maxReply = 4 << 20

// OpenAI is the provider of OpenAI, with the Responses API. Each call
// turns off the response store (`docs/design.md` section 6). The key
// comes from the caller, and Phase 3 gives no key to the API. Each
// call of this provider costs money, so a development run needs the
// approval of the owner (D-25).
type OpenAI struct {
	key      string
	endpoint string
	client   *http.Client
}

// NewOpenAI gives the provider of OpenAI. It refuses an empty key. An
// empty endpoint gives OpenAIEndpoint, and a nil client gives
// http.DefaultClient. A test gives the endpoint of a local server.
func NewOpenAI(key, endpoint string, client *http.Client) (*OpenAI, error) {
	if key == "" {
		return nil, errors.New("ai: the OpenAI key is empty")
	}
	if endpoint == "" {
		endpoint = OpenAIEndpoint
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &OpenAI{key, endpoint, client}, nil
}

type openAIRequest struct {
	Model     string `json:"model"`
	Reasoning struct {
		Effort string `json:"effort"`
	} `json:"reasoning"`
	Instructions string `json:"instructions"`
	Input        string `json:"input"`
	Text         struct {
		Format struct {
			Type   string          `json:"type"`
			Name   string          `json:"name"`
			Schema json.RawMessage `json:"schema"`
			Strict bool            `json:"strict"`
		} `json:"format"`
	} `json:"text"`
	MaxOutputTokens int  `json:"max_output_tokens"`
	Store           bool `json:"store"`
}

type openAIResponse struct {
	Status string `json:"status"`
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type    string  `json:"type"`
			Text    string  `json:"text"`
			Refusal *string `json:"refusal"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens        int64 `json:"input_tokens"`
		InputTokensDetails struct {
			CachedTokens     int64 `json:"cached_tokens"`
			CacheWriteTokens int64 `json:"cache_write_tokens"`
		} `json:"input_tokens_details"`
		OutputTokens        int64 `json:"output_tokens"`
		OutputTokensDetails struct {
			ReasoningTokens int64 `json:"reasoning_tokens"`
		} `json:"output_tokens_details"`
	} `json:"usage"`
}

// Send sends one call to OpenAI. An error names the HTTP status alone.
func (o *OpenAI) Send(ctx context.Context, c Call) (Reply, error) {
	var body openAIRequest
	body.Model = c.Role.Model
	body.Reasoning.Effort = c.Role.Effort
	body.Instructions = c.Instructions
	body.Input = string(c.Input)
	body.Text.Format.Type = "json_schema"
	body.Text.Format.Name = c.SchemaName
	body.Text.Format.Schema = c.Schema
	body.Text.Format.Strict = true
	body.MaxOutputTokens = c.Role.MaxOutputTokens
	data, err := json.Marshal(body)
	if err != nil {
		return Reply{}, fmt.Errorf("openai: request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.endpoint, bytes.NewReader(data))
	if err != nil {
		return Reply{}, errors.New("openai: request: bad endpoint")
	}
	req.Header.Set("Authorization", "Bearer "+o.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return Reply{}, ctx.Err()
		}
		return Reply{}, errors.New("openai: the request failed")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Reply{}, fmt.Errorf("openai: HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxReply))
	if err != nil {
		if ctx.Err() != nil {
			return Reply{}, ctx.Err()
		}
		return Reply{}, errors.New("openai: the reply failed")
	}
	var r openAIResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return Reply{}, errors.New("openai: the reply is not JSON")
	}
	out := Reply{Usage: Usage{
		InputTokens:       r.Usage.InputTokens,
		CachedInputTokens: r.Usage.InputTokensDetails.CachedTokens,
		CacheWriteTokens:  r.Usage.InputTokensDetails.CacheWriteTokens,
		OutputTokens:      r.Usage.OutputTokens,
		ReasoningTokens:   r.Usage.OutputTokensDetails.ReasoningTokens,
	}}
	for _, item := range r.Output {
		if item.Type != "message" {
			continue
		}
		for _, part := range item.Content {
			switch part.Type {
			case "output_text":
				out.Text += part.Text
			case "refusal":
				out.Refusal = true
			}
		}
	}
	switch {
	case out.Refusal || r.Status == "completed":
	case r.Status == "incomplete":
		out.Incomplete = true
	default:
		return out, errors.New("openai: the reply did not complete")
	}
	return out, nil
}
