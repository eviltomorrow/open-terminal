package llm

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	libqdrant "github.com/eviltomorrow/open-terminal/lib/qdrant"
)

func init() {
	if _, err := libqdrant.InitQdrant(&libqdrant.Config{
		StartupRetryPeriod: 3 * time.Second,
		StartupRetryTimes:  3,
		ConnectTimeout:     5 * time.Second,

		Host: "localhost",
		Port: 6334,
	}); err != nil {
		log.Fatal(err)
	}
}

func TestKimiStream(t *testing.T) {
	apiKey := os.Getenv("KIMI_API_KEY")
	if apiKey == "" {
		t.Fatal("KIMI_API_KEY is nil")
	}

	client := NewKimiClient("https://api.moonshot.cn/v1", apiKey)

	session, err := client.NewSession("moonshot-v1-32k")
	if err != nil {
		t.Fatalf("NewSession failure, nest error: %v", err)
	}
	defer session.Close()

	ch, err := session.StartChat(context.Background(), "你好")
	if err != nil {
		t.Fatalf("StartChat failure, nest error: %v", err)
	}
	for c := range ch {
		fmt.Println(c)
	}
	// ch, err := client.ChatStream(context.Background(), []*chat.Message{
	// 	{
	// 		Role:    openai.ChatMessageRoleUser,
	// 		Content: "帮我介绍一下 Rust 语言，并分析一下日后趋势",
	// 	},
	// })
	// if err != nil {
	// 	t.Fatalf("ChatStream failure, nest error: %v", err)
	// }

	// for r := range ch {
	// 	fmt.Printf("%s", r)
	// }
}
