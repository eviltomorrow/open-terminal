package llm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/eviltomorrow/open-terminal/apps/open-server/domain/chat"
	"github.com/sashabaranov/go-openai"
)

func TestKimiStream(t *testing.T) {
	apiKey := os.Getenv("KIMI_API_KEY")
	if apiKey == "" {
		t.Fatal("KIMI_API_KEY is nil")
	}

	client := NewKimiClient("https://api.moonshot.cn/v1", apiKey, "moonshot-v1-32k")

	ch, err := client.ChatStream(context.Background(), []*chat.Message{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "帮我介绍一下 Rust 语言，并分析一下日后趋势",
		},
	})
	if err != nil {
		t.Fatalf("ChatStream failure, nest error: %v", err)
	}

	for r := range ch {
		fmt.Printf("%s", r)
	}
}
