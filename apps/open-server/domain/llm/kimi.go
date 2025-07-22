package llm

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/eviltomorrow/open-terminal/apps/open-server/domain/chat"
	"github.com/eviltomorrow/open-terminal/lib/zlog"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type KimiClient struct {
	BaseURL   string
	APIKey    string
	ModelName string

	client *openai.Client
}

func NewKimiClient(baseURL string, apiKey string, modelName string) *KimiClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL

	client := &KimiClient{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ModelName: modelName,

		client: openai.NewClientWithConfig(cfg),
	}
	return client
}

func (c *KimiClient) ChatStream(ctx context.Context, msgs []*chat.Message) (chan string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("panic: openai client is nil")
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:     c.ModelName,
		Messages:  messages,
		MaxTokens: 30000,
		Stream:    true,
	}
	resp, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 64)
	go func() {
		for {
			stream, err := resp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				zlog.Error("Recv failure", zap.Error(err))
				break
			}

			if len(stream.Choices) > 0 {
				delta := stream.Choices[0].Delta.Content
				for _, r := range delta {
					ch <- string(r)
					time.Sleep(20 * time.Millisecond)
				}
			}
		}
		close(ch)
		resp.Close()
	}()

	return ch, nil
}
