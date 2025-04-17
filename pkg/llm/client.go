package llm

import (
	"context"
	"fmt"
)

// Client defines the interface for LLM interactions
type Client interface {
	// Analyze generates an analysis of the provided code context
	Analyze(ctx context.Context, input AnalyzeInput, progress ProgressCallback) (*AnalyzeOutput, error)

	// ExplainFile generates an explanation of a specific file
	ExplainFile(ctx context.Context, input ExplainInput) (*ExplainOutput, error)
}

// AnalyzeInput contains the input for code analysis
type AnalyzeInput struct {
	Files        map[string]string // filename -> content
	Languages    map[string]float64
	ContextSize  int
	DirStructure string // Tree-like directory structure
	IsDetailed   bool   // Whether to perform detailed analysis
}

// AnalyzeOutput contains the analysis results
type AnalyzeOutput struct {
	Description  string
	Architecture string
	Components   []Component
	Setup        string
	FlowDiagram  string
}

// ExplainInput contains the input for file explanation
type ExplainInput struct {
	Filename    string
	Content     string
	ContextSize int
}

// ExplainOutput contains the file explanation
type ExplainOutput struct {
	Explanation string
	Purpose     string
	Components  []string
}

// Component represents a code component identified by the LLM
type Component struct {
	Name        string
	Type        string
	Description string
	Path        string
}

// Config contains LLM client configuration
type Config struct {
	OpenAIKey string
	APIBase   string
	Model     string
}

// NewClient creates a new LLM client based on the configuration
func NewClient(config Config) (Client, error) {
	if config.OpenAIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.APIBase == "" {
		config.APIBase = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	return newOpenAIClient(config)
}

// Template for the analysis prompt
const analyzePrompt = `Analyze the following codebase and provide a comprehensive overview:

Repository Contents:
%s

Please provide:
1. A brief description of what this codebase does
2. The main architectural components and their relationships
3. Key components and their purposes
4. Setup instructions if available
5. A mermaid diagram showing the main components and their interactions

Focus on the most important aspects and keep the response clear and concise.`

// Template for the file explanation prompt
const explainPrompt = `Explain the following file in detail:

Filename: %s

Content:
%s

Please provide:
1. What this file does
2. Its main purpose in the codebase
3. Key components/functions and their roles
4. Any important patterns or considerations

Keep the explanation clear and focused on the most important aspects.`
