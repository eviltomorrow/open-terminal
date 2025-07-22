package llm

import (
	"context"
	"io"

	"github.com/eviltomorrow/open-terminal/lib/snowflake"
	"github.com/eviltomorrow/open-terminal/lib/zlog"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type KimiClient struct {
	BaseURL   string
	APIKey    string
	ModelName string

	ai *openai.Client
}

func NewKimiClient(baseURL string, apiKey string) *KimiClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL

	client := &KimiClient{
		BaseURL: baseURL,
		APIKey:  apiKey,

		ai: openai.NewClientWithConfig(cfg),
	}
	return client
}

type KimiSession struct {
	Id        string
	ModelName string

	client *KimiClient
}

func (c *KimiClient) NewSession(modelName string) *KimiSession {
	id := snowflake.GenerateID()

	session := &KimiSession{
		Id:        id,
		ModelName: modelName,
	}
	return session
}

func (s *KimiSession) StartChat(ctx context.Context, content string, opts ...func(*openai.ChatCompletionRequest)) (chan string, error) {
	req := openai.ChatCompletionRequest{
		Model:  s.ModelName,
		Stream: true,

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: content,
			},
		},
	}

	for _, opt := range opts {
		opt(&req)
	}

	return s.sendRequest(ctx, req)
}

func (s *KimiSession) Send(ctx context.Context, content string, opts ...func(*openai.ChatCompletionRequest)) (chan string, error) {
	req := openai.ChatCompletionRequest{
		Model:  s.ModelName,
		Stream: true,

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
	}

	for _, opt := range opts {
		opt(&req)
	}

	return s.sendRequest(ctx, req)
}

func (s *KimiSession) sendRequest(ctx context.Context, req openai.ChatCompletionRequest) (chan string, error) {
	resp, err := s.client.ai.CreateChatCompletionStream(ctx, req)
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
				ch <- delta
			}
		}
		close(ch)
		resp.Close()
	}()

	return ch, nil
}

func (c *KimiClient) Close() error {
	return nil
}
