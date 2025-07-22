package llm

import "github.com/sashabaranov/go-openai"

func WithChatCompletionRequestForTemperature(val float32) func(*openai.ChatCompletionRequest) {
	return func(ccr *openai.ChatCompletionRequest) {
		ccr.Temperature = val
	}
}
