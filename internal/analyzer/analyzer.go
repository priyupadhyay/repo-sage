package analyzer

import "time"

// RepoInfo contains the analyzed repository information
type RepoInfo struct {
	Name         string
	Description  string
	Languages    map[string]float64 // language -> percentage
	Components   []Component
	EntryPoints  []string
	Dependencies map[string]string // dependency -> version
}

// Component represents a major component in the codebase
type Component struct {
	Name        string
	Type        string // "API", "CLI", "Service", "Utility", etc.
	Path        string
	Description string
	Files       []string
}

// AnalysisResult contains the complete analysis output
type AnalysisResult struct {
	RepoInfo      RepoInfo
	Architecture  string
	Setup         string
	FlowDiagram   string
	AnalyzedAt    time.Time
	GeneratedWith string
}

// Analyzer defines the interface for repository analysis
type Analyzer interface {
	// Analyze performs the complete repository analysis
	Analyze(repoPath string, options AnalyzeOptions) (*AnalysisResult, error)

	// ExplainFile generates a detailed explanation of a specific file
	ExplainFile(filePath string, options ExplainOptions) (string, error)
}

// AnalyzeOptions contains configuration for the analysis
type AnalyzeOptions struct {
	ContextSize int
	OpenAIKey   string
	APIBase     string
	Model       string
	OutputPath  string
	Detailed    bool // If true, perform detailed code analysis
}

// ExplainOptions contains configuration for file explanation
type ExplainOptions struct {
	ContextSize int
	OpenAIKey   string
	APIBase     string
	Model       string
}
