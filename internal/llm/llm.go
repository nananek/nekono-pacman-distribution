// Package llm is a minimal OpenAI-compatible chat client (ocgo endpoint).
//
// It exists so the harness does not depend on OpenCode: the agent loop, the
// tool surface and the phase isolation are ours, and the only thing that comes
// from outside is /chat/completions. Streaming is used because the host side
// needs progress while a phase runs (an OpenCode run only explains itself at
// the end).
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ToolCall is one function call requested by the model.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Msg is one chat message.
type Msg struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// FunctionDef is the schema of one callable tool.
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// Tool registers one function with the model.
type Tool struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// Usage is the token accounting returned by the provider.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Completion is the assembled result of one assistant turn.
type Completion struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
	Model     string
	Truncated bool
}

// Client calls {BaseURL}/chat/completions.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
	// SessionID is sent as x-opencode-session. The ocgo endpoint rejects
	// requests without it (it routes per session), so it is required there
	// and ignored by plain OpenAI-compatible servers.
	SessionID string
	// Stream enables SSE; the non-streaming path exists for tests.
	Stream bool
	// OnDelta, when set, receives content deltas as they arrive.
	OnDelta func(string)
}

// New builds a client for an OpenAI-compatible endpoint.
func New(baseURL, apiKey, model string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTP:    &http.Client{Timeout: 15 * time.Minute},
		Stream:  true,
	}
}

type request struct {
	Model         string          `json:"model"`
	Messages      []Msg           `json:"messages"`
	Tools         []Tool          `json:"tools,omitempty"`
	Stream        bool            `json:"stream"`
	StreamOptions *streamOptions  `json:"stream_options,omitempty"`
	Temperature   *float64        `json:"temperature,omitempty"`
	Extra         json.RawMessage `json:"-"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// Complete sends the conversation and returns the assistant turn.
func (c *Client) Complete(ctx context.Context, msgs []Msg, tools []Tool) (Completion, error) {
	body := request{Model: c.Model, Messages: msgs, Tools: tools, Stream: c.Stream}
	if c.Stream {
		body.StreamOptions = &streamOptions{IncludeUsage: true}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return Completion{}, err
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Completion{}, ctx.Err()
			case <-time.After(time.Duration(attempt) * 3 * time.Second):
			}
		}
		out, err := c.do(ctx, raw)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !retryable(err) {
			break
		}
	}
	return Completion{}, lastErr
}

type httpError struct {
	Status int
	Body   string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("llm: http %d: %s", e.Status, e.Body)
}

func retryable(err error) bool {
	if e, ok := err.(*httpError); ok {
		return e.Status == 429 || e.Status >= 500
	}
	return true
}

func (c *Client) do(ctx context.Context, raw []byte) (Completion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return Completion{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if c.SessionID != "" {
		req.Header.Set("x-opencode-session", c.SessionID)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Completion{}, &httpError{Status: resp.StatusCode, Body: string(b)}
	}
	if !c.Stream {
		return parseNonStream(resp.Body)
	}
	return c.parseStream(resp.Body)
}

type chunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
}

func (c *Client) parseStream(body io.Reader) (Completion, error) {
	out := Completion{}
	var content strings.Builder
	calls := map[int]*ToolCall{}
	order := []int{}

	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var ch chunk
		if err := json.Unmarshal([]byte(data), &ch); err != nil {
			// Providers occasionally emit keep-alives / partial frames.
			continue
		}
		if ch.Model != "" {
			out.Model = ch.Model
		}
		if ch.Usage != nil {
			out.Usage = *ch.Usage
		}
		for _, choice := range ch.Choices {
			if d := choice.Delta.Content; d != "" {
				content.WriteString(d)
				if c.OnDelta != nil {
					c.OnDelta(d)
				}
			}
			for _, tc := range choice.Delta.ToolCalls {
				cur, ok := calls[tc.Index]
				if !ok {
					cur = &ToolCall{Type: "function"}
					calls[tc.Index] = cur
					order = append(order, tc.Index)
				}
				if tc.ID != "" {
					cur.ID = tc.ID
				}
				if tc.Function.Name != "" {
					cur.Function.Name += tc.Function.Name
				}
				cur.Function.Arguments += tc.Function.Arguments
			}
		}
	}
	if err := sc.Err(); err != nil && err != io.EOF {
		return out, err
	}
	out.Content = content.String()
	for _, i := range order {
		out.ToolCalls = append(out.ToolCalls, *calls[i])
	}
	return out, nil
}

func parseNonStream(body io.Reader) (Completion, error) {
	var resp struct {
		Model   string `json:"model"`
		Choices []struct {
			Message Msg `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return Completion{}, err
	}
	if len(resp.Choices) == 0 {
		return Completion{}, fmt.Errorf("llm: no choices in response")
	}
	return Completion{
		Content:   resp.Choices[0].Message.Content,
		ToolCalls: resp.Choices[0].Message.ToolCalls,
		Usage:     resp.Usage,
		Model:     resp.Model,
	}, nil
}
