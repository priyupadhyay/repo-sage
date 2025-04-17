package llm

import (
	"context"
	"fmt"
)

type ollamaClient struct{}

func newOllamaClient() (Client, error) {
	return &ollamaClient{}, nil
}

func (c *ollamaClient) Analyze(ctx context.Context, input AnalyzeInput, progress ProgressCallback) (*AnalyzeOutput, error) {
	if progress != nil {
		progress("Not implemented", 0, 1, "")
	}
	return nil, fmt.Errorf("Ollama integration not implemented yet")
}

func (c *ollamaClient) ExplainFile(ctx context.Context, input ExplainInput) (*ExplainOutput, error) {
	return nil, fmt.Errorf("Ollama integration not implemented yet")
}
