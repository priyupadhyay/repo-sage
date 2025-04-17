package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/priyupadhyay/repo-sage/pkg/git"
	"github.com/priyupadhyay/repo-sage/pkg/llm"
)

type analyzer struct {
	llmClient llm.Client
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(options AnalyzeOptions) (Analyzer, error) {
	llmClient, err := llm.NewClient(llm.Config{
		OpenAIKey: options.OpenAIKey,
		APIBase:   options.APIBase,
		Model:     options.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &analyzer{
		llmClient: llmClient,
	}, nil
}

func (a *analyzer) Analyze(repoPath string, options AnalyzeOptions) (*AnalysisResult, error) {
	repo, err := git.New(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	fmt.Println("\n📂 Scanning repository files...")
	// Get repository files
	files, err := repo.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list repository files: %w", err)
	}

	fmt.Printf("Found %d files\n", len(files))
	fmt.Println("\n🔍 Analyzing languages...")
	// Get language statistics
	languages, err := repo.GetLanguages()
	if err != nil {
		return nil, fmt.Errorf("failed to get language statistics: %w", err)
	}

	fmt.Printf("Languages detected: %v\n", formatLanguages(languages))

	// Build directory structure
	dirStructure := buildDirStructure(files)

	// Read important files for quick summary
	importantFiles := make(map[string]string)

	// Always include README files
	for _, file := range files {
		base := strings.ToLower(filepath.Base(file))
		if strings.HasPrefix(base, "readme.") {
			content, err := repo.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", file, err)
			}
			importantFiles[file] = string(content)
			break // Only use the first README found
		}
	}

	// Add package manifests
	manifestFiles := []string{
		"go.mod", "package.json", "requirements.txt", "Cargo.toml",
		"Gemfile", "composer.json", "pom.xml", "build.gradle",
	}
	for _, manifest := range manifestFiles {
		for _, file := range files {
			if filepath.Base(file) == manifest {
				content, err := repo.ReadFile(file)
				if err != nil {
					return nil, fmt.Errorf("failed to read file %s: %w", file, err)
				}
				importantFiles[file] = string(content)
			}
		}
	}

	// Add main/index files if not in detailed mode
	if !options.Detailed {
		for _, file := range files {
			base := filepath.Base(file)
			if base == "main.go" || base == "index.js" || base == "index.ts" {
				content, err := repo.ReadFile(file)
				if err != nil {
					return nil, fmt.Errorf("failed to read file %s: %w", file, err)
				}
				importantFiles[file] = string(content)
			}
		}
	}

	var fileContents map[string]string
	if options.Detailed {
		fmt.Println("\n📖 Reading all files...")
		// Read all files for detailed analysis
		fileContents = make(map[string]string)
		for i, file := range files {
			fmt.Printf("\r%d/%d files processed", i+1, len(files))
			content, err := repo.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", file, err)
			}
			fileContents[file] = string(content)
		}
		fmt.Println()
	} else {
		fileContents = importantFiles
	}

	// Prepare analysis input with directory structure
	analysisInput := fmt.Sprintf("Directory Structure:\n%s\n\nFiles to analyze:\n", dirStructure)
	for name := range fileContents {
		analysisInput += fmt.Sprintf("- %s\n", name)
	}

	fmt.Println("\n🤖 Analyzing with AI...")
	// Analyze with LLM
	analysis, err := a.llmClient.Analyze(context.Background(), llm.AnalyzeInput{
		Files:        fileContents,
		Languages:    languages,
		ContextSize:  options.ContextSize,
		DirStructure: dirStructure,
		IsDetailed:   options.Detailed,
	}, func(stage string, current, total int, response string) {
		switch stage {
		case "Preparing files":
			fmt.Printf("\r⚙️  %s... %d/%d", stage, current, total)
		case "Processing files":
			fmt.Printf("\r📝 %s... %d/%d", stage, current, total)
		case "Analyzing chunks":
			fmt.Printf("\r🧠 %s... %d/%d", stage, current, total)
		case "Analysis response":
			fmt.Printf("\n\n🔹 Analysis part %d/%d:\n%s\n", current, total, response)
		case "Generating summary":
			fmt.Printf("\n\n📊 Generating final summary...\n")
		case "Final summary":
			fmt.Printf("\n✨ Final Analysis:\n%s\n", response)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze repository: %w", err)
	}

	// Convert components
	components := make([]Component, len(analysis.Components))
	for i, c := range analysis.Components {
		components[i] = Component{
			Name:        c.Name,
			Type:        c.Type,
			Path:        c.Path,
			Description: c.Description,
		}
	}

	return &AnalysisResult{
		RepoInfo: RepoInfo{
			Name:         filepath.Base(repoPath),
			Description:  analysis.Description,
			Languages:    languages,
			Components:   components,
			EntryPoints:  findEntryPoints(files),
			Dependencies: findDependencies(files, fileContents),
		},
		Architecture:  analysis.Architecture,
		Setup:         analysis.Setup,
		FlowDiagram:   analysis.FlowDiagram,
		GeneratedWith: "repo-sage",
	}, nil
}

func formatLanguages(langs map[string]float64) string {
	var result string
	for lang, pct := range langs {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%s (%.1f%%)", lang, pct)
	}
	return result
}

func (a *analyzer) ExplainFile(filePath string, options ExplainOptions) (string, error) {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find git repository by walking up the directory tree
	dir := filepath.Dir(absPath)
	var repo *git.Repository
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			repo, err = git.New(dir)
			if err != nil {
				return "", fmt.Errorf("failed to open repository: %w", err)
			}
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no git repository found in parent directories")
		}
		dir = parent
	}

	// Get the relative path within the repository
	relPath, err := filepath.Rel(repo.Path, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	content, err := repo.ReadFile(relPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	explanation, err := a.llmClient.ExplainFile(context.Background(), llm.ExplainInput{
		Filename:    filepath.Base(absPath),
		Content:     string(content),
		ContextSize: options.ContextSize,
	})
	if err != nil {
		return "", fmt.Errorf("failed to explain file: %w", err)
	}

	return explanation.Explanation, nil
}

// findEntryPoints identifies potential entry points in the repository
func findEntryPoints(files []string) []string {
	var entryPoints []string
	for _, file := range files {
		base := filepath.Base(file)
		if base == "main.go" || base == "index.js" || base == "app.py" {
			entryPoints = append(entryPoints, file)
		}
	}
	return entryPoints
}

// findDependencies extracts dependencies from common dependency files
func findDependencies(files []string, contents map[string]string) map[string]string {
	// TODO: Implement dependency parsing from package.json, go.mod, requirements.txt, etc.
	return map[string]string{}
}

func buildDirStructure(files []string) string {
	// Create a map to store directory structure
	dirs := make(map[string]bool)
	for _, file := range files {
		dir := filepath.Dir(file)
		for dir != "." && dir != "/" {
			dirs[dir] = true
			dir = filepath.Dir(dir)
		}
	}

	// Convert to tree structure
	var result strings.Builder
	result.WriteString(".\n")

	// Convert map to sorted slice for consistent output
	var paths []string
	for path := range dirs {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		depth := strings.Count(path, string(os.PathSeparator))
		result.WriteString(strings.Repeat("  ", depth))
		result.WriteString("└── ")
		result.WriteString(filepath.Base(path))
		result.WriteString("\n")
	}

	return result.String()
}
