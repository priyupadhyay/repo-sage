package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type openAIClient struct {
	apiKey  string
	apiBase string
	model   string
	client  *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ProgressCallback is called to report progress during analysis
type ProgressCallback func(stage string, current, total int, response string)

func newOpenAIClient(config Config) (Client, error) {
	return &openAIClient{
		apiKey:  config.OpenAIKey,
		apiBase: config.APIBase,
		model:   config.Model,
		client:  &http.Client{},
	}, nil
}

func (c *openAIClient) makeRequest(ctx context.Context, prompt string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are a helpful AI assistant that analyzes and explains code."},
			{Role: "user", Content: prompt},
		},
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiBase+"/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *openAIClient) Analyze(ctx context.Context, input AnalyzeInput, progress ProgressCallback) (*AnalyzeOutput, error) {
	// For quick summary, use a single prompt with directory structure and important files
	if !input.IsDetailed {
		if progress != nil {
			progress("Preparing quick summary", 0, 1, "")
		}

		prompt := fmt.Sprintf(`Analyze this codebase and provide a quick overview:

Directory Structure:
%s

Languages:
%s

Please provide:
1. A brief description of what this codebase likely does
2. Main components and their purpose (based on directory structure)
3. Technologies used (based on file types and languages)
4. Setup/build system (based on manifest files)

Focus on high-level understanding and keep it concise.`, input.DirStructure, formatLanguages(input.Languages))

		response, err := c.makeRequest(ctx, prompt)
		if err != nil {
			return nil, err
		}

		if progress != nil {
			progress("Quick summary", 1, 1, response)
		}

		return &AnalyzeOutput{
			Description:  response,
			Architecture: "",
			Components:   nil,
			Setup:        "",
			FlowDiagram:  "",
		}, nil
	}

	// For detailed analysis, process all files in chunks
	// Sort files by size to process most important files first
	type fileInfo struct {
		name    string
		content string
	}
	files := make([]fileInfo, 0, len(input.Files))
	for name, content := range input.Files {
		files = append(files, fileInfo{name, content})
	}
	sort.Slice(files, func(i, j int) bool {
		// Prioritize main files and shorter files
		iMain := strings.Contains(files[i].name, "main.") || strings.Contains(files[i].name, "index.")
		jMain := strings.Contains(files[j].name, "main.") || strings.Contains(files[j].name, "index.")
		if iMain != jMain {
			return iMain
		}
		return len(files[i].content) < len(files[j].content)
	})

	// Process files in chunks
	const maxChunkSize = 1500 // characters per chunk
	var chunks []string
	currentChunk := strings.Builder{}

	for i, file := range files {
		if progress != nil {
			progress("Processing files", i+1, len(files), "")
		}

		fileContent := fmt.Sprintf("File: %s\n\n%s\n\n", file.name, file.content)
		if currentChunk.Len()+len(fileContent) > maxChunkSize {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}
			// If the file is too large, split it into smaller chunks
			if len(fileContent) > maxChunkSize {
				parts := splitLongContent(fileContent, maxChunkSize)
				chunks = append(chunks, parts...)
				continue
			}
		}
		currentChunk.WriteString(fileContent)
	}
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	// Analyze each chunk
	var descriptions []string
	for i, chunk := range chunks {
		if progress != nil {
			progress("Analyzing chunks", i+1, len(chunks), "")
		}

		prompt := fmt.Sprintf("Analyze this part of the codebase. Focus on key components, patterns, and functionality. Be concise:\n\n%s", chunk)
		response, err := c.makeRequest(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze chunk %d: %w", i+1, err)
		}

		if progress != nil {
			progress("Analysis response", i+1, len(chunks), response)
		}

		descriptions = append(descriptions, response)
	}

	// Combine the results
	if len(descriptions) > 1 {
		if progress != nil {
			progress("Generating summary", 0, 1, "")
		}

		summaryPrompt := fmt.Sprintf("Combine these analysis parts into a concise overview focusing on key components and architecture:\n\n%s", strings.Join(descriptions, "\n\n---\n\n"))
		finalResponse, err := c.makeRequest(ctx, summaryPrompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary: %w", err)
		}

		if progress != nil {
			progress("Final summary", 1, 1, finalResponse)
		}

		descriptions = []string{finalResponse}
	}

	return &AnalyzeOutput{
		Description:  descriptions[0],
		Architecture: "",
		Components:   nil,
		Setup:        "",
		FlowDiagram:  "",
	}, nil
}

func formatLanguages(langs map[string]float64) string {
	var result []string
	for lang, pct := range langs {
		result = append(result, fmt.Sprintf("%s (%.1f%%)", lang, pct))
	}
	sort.Strings(result)
	return strings.Join(result, ", ")
}

func splitLongContent(content string, maxSize int) []string {
	var chunks []string
	lines := strings.Split(content, "\n")
	currentChunk := strings.Builder{}

	for _, line := range lines {
		if currentChunk.Len()+len(line)+1 > maxSize {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}
			// If a single line is too long, split it
			if len(line) > maxSize {
				for i := 0; i < len(line); i += maxSize {
					end := i + maxSize
					if end > len(line) {
						end = len(line)
					}
					chunks = append(chunks, line[i:end])
				}
				continue
			}
		}
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)
	}
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}
	return chunks
}

func (c *openAIClient) ExplainFile(ctx context.Context, input ExplainInput) (*ExplainOutput, error) {
	prompt := fmt.Sprintf(explainPrompt, input.Filename, input.Content)
	response, err := c.makeRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &ExplainOutput{
		Explanation: response,
		Purpose:     "",
		Components:  nil,
	}, nil
}
