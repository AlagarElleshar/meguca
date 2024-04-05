package websockets

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bakape/meguca/config"
	"io"
	"net/http"
)

func streamMessages(model string, systemPrompt string, maxTokens int, message string, callback func(string)) error {
	apiKey := config.Server.AnthropicApiKey

	url := "https://api.anthropic.com/v1/messages"

	body := requestData{
		model,
		[]messageParam{messageParam{"user", message}},
		maxTokens,
		true,
		systemPrompt,
	}

	jsonBody, err := json.Marshal(body)
	fmt.Println(string(jsonBody))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "messages-2023-12-15")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte(": "), 2)
		if len(parts) != 2 {
			continue
		}

		eventType := parts[0]
		if !bytes.Equal(eventType, []byte("data")) {
			continue
		}

		var event contentBlockDeltaEvent
		err = json.Unmarshal(parts[1], &event)
		if err != nil {
			return err
		}

		callback(event.Delta.Text)
	}

	return nil
}

type requestData struct {
	Model        string         `json:"model"`
	Messages     []messageParam `json:"messages"`
	MaxTokens    int            `json:"max_tokens"`
	Stream       bool           `json:"stream"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
}

type contentBlock struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type contentBlockDeltaEvent struct {
	Delta textDelta `json:"delta"`
	Index int       `json:"index"`
	Type  string    `json:"type"`
}

type contentBlockStartEvent struct {
	ContentBlock contentBlock `json:"content_block"`
	Index        int          `json:"index"`
	Type         string       `json:"type"`
}

type contentBlockStopEvent struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
}

type imageBlockParam struct {
	Source imageBlockParamSource `json:"source"`
	Type   string                `json:"type,omitempty"`
}

type imageBlockParamSource struct {
	Data      string `json:"data"`
	MediaType string `json:"media_type"`
	Type      string `json:"type,omitempty"`
}

type message struct {
	ID           string         `json:"id"`
	Content      []contentBlock `json:"content"`
	Model        string         `json:"model"`
	Role         string         `json:"role"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Type         string         `json:"type"`
	Usage        usage          `json:"usage"`
}

type messageDeltaEvent struct {
	Delta messageDeltaEventDelta `json:"delta"`
	Type  string                 `json:"type"`
	Usage messageDeltaUsage      `json:"usage"`
}

type messageDeltaEventDelta struct {
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
}

type messageDeltaUsage struct {
	OutputTokens int `json:"output_tokens"`
}

type messageParam struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messageStartEvent struct {
	Message message `json:"message"`
	Type    string  `json:"type"`
}

type messageStopEvent struct {
	Type string `json:"type"`
}

type textBlock struct {
	Text string `json:"text"`
	Type string `json:"type,omitempty"`
}

type textDelta struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
