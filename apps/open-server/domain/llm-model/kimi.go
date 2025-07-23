package llm

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	libqdrant "github.com/eviltomorrow/open-terminal/lib/qdrant"
	"github.com/eviltomorrow/open-terminal/lib/snowflake"
	"github.com/eviltomorrow/open-terminal/lib/zlog"
	"github.com/qdrant/go-client/qdrant"
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
	sync.RWMutex

	Id        string
	ModelName string

	client       *KimiClient
	alreadyStart bool
	num          uint64
}

func (c *KimiClient) NewSession(modelName string) (*KimiSession, error) {
	id := fmt.Sprintf("Kimi-%v", snowflake.GenerateID())

	session := &KimiSession{
		Id:        id,
		ModelName: modelName,

		client: c,
	}
	return session, nil
}

func (s *KimiSession) getNum() uint64 {
	s.Lock()
	defer s.Unlock()

	s.num = s.num + 1
	return s.num
}

func (s *KimiSession) isAlreadyStart() bool {
	s.RLock()
	defer s.RUnlock()

	return s.alreadyStart
}

func (s *KimiSession) setAlreadyStart() {
	s.Lock()
	defer s.Unlock()

	s.alreadyStart = true
}

func (s *KimiSession) embeddings(content string) ([]float32, error) {
	resp, err := s.client.ai.CreateEmbeddings(context.Background(),
		openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{content},
		})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) > 0 {
		return resp.Data[0].Embedding, nil
	}
	return nil, fmt.Errorf("panic: no embeddings result")
}

func (s *KimiSession) cache(id uint64, content string) error {
	vec, err := s.embeddings(content)
	if err != nil {
		return err
	}

	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDNum(id),
		Vectors: qdrant.NewVectors(vec...),
		Payload: qdrant.NewValueMap(map[string]interface{}{"content": content}),
	}

	_, err = libqdrant.Client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: s.Id,
		Points:         []*qdrant.PointStruct{point},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *KimiSession) search(content string) ([]string, error) {
	vec, err := s.embeddings(content)
	if err != nil {
		return nil, err
	}

	points, err := libqdrant.Client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: s.Id,
		Query:          qdrant.NewQuery(vec...),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	data := make([]string, 0, len(points))
	for _, point := range points {
		text := point.Payload["content"].GetStringValue()
		data = append(data, text)
	}
	return data, nil
}

func (s *KimiSession) StartChat(ctx context.Context, content string, opts ...func(*openai.ChatCompletionRequest)) (chan string, error) {
	if s.isAlreadyStart() {
		return nil, fmt.Errorf("session'chat already start")
	}

	if err := libqdrant.Client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: s.Id,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1536,
			Distance: qdrant.Distance_Cosine,
		}),
	}); err != nil {
		return nil, err
	}

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

	// ch, err := s.sendRequest(ctx, req)
	// if err != nil {
	// 	return nil, err
	// }
	ch := make(chan string, 1)
	ch <- content

	s.setAlreadyStart()

	if err := s.cache(s.getNum(), content); err != nil {
		zlog.Error("Cache content failure", zap.Error(err), zap.String("content", content), zap.String("sessionId", s.Id))
	}
	return ch, nil
}

func (s *KimiSession) Send(ctx context.Context, content string, opts ...func(*openai.ChatCompletionRequest)) (chan string, error) {
	relevant, err := s.search(content)
	if err != nil {
		return nil, fmt.Errorf("search history content failure, nest error: %v", err)
	}

	msg := strings.Join(relevant, "\n---\n")

	req := openai.ChatCompletionRequest{
		Model:  s.ModelName,
		Stream: true,

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: msg,
			},
		},
	}

	for _, opt := range opts {
		opt(&req)
	}

	ch := make(chan string, 1)
	ch <- content
	return ch, nil
	// return s.sendRequest(ctx, req)
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

func (s *KimiSession) Close() error {
	return nil
}
