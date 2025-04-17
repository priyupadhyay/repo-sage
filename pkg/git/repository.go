package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository
type Repository struct {
	Path string
}

// New creates a new Repository instance
func New(path string) (*Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("repository path does not exist: %w", err)
	}

	// Check if it's a Git repository
	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	return &Repository{
		Path: absPath,
	}, nil
}

// ListFiles returns all tracked files in the repository
func (r *Repository) ListFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(r.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip common dependency directories
		if isInDependencyDir(path) {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(r.Path, path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk repository: %w", err)
	}

	return files, nil
}

// isInDependencyDir checks if the path is in a common dependency directory
func isInDependencyDir(path string) bool {
	deps := []string{
		"node_modules",
		"vendor",
		"dist",
		"build",
		".venv",
		"venv",
		"env",
	}

	for _, dep := range deps {
		if filepath.Base(path) == dep {
			return true
		}
	}

	return false
}

// ReadFile reads the contents of a file in the repository
func (r *Repository) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(r.Path, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return content, nil
}

// GetLanguages returns a map of languages and their usage percentages
func (r *Repository) GetLanguages() (map[string]float64, error) {
	files, err := r.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Count bytes per language
	langBytes := make(map[string]int64)
	totalBytes := int64(0)

	for _, file := range files {
		lang := detectLanguage(file)
		if lang == "" {
			continue
		}

		content, err := r.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		bytes := int64(len(content))
		langBytes[lang] += bytes
		totalBytes += bytes
	}

	// Convert byte counts to percentages
	result := make(map[string]float64)
	if totalBytes > 0 {
		for lang, bytes := range langBytes {
			result[lang] = float64(bytes) / float64(totalBytes) * 100
		}
	}

	return result, nil
}

// detectLanguage returns the programming language based on file extension
func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "Go"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".jsx":
		return "React"
	case ".tsx":
		return "React/TypeScript"
	case ".py":
		return "Python"
	case ".java":
		return "Java"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".rs":
		return "Rust"
	case ".c":
		return "C"
	case ".cpp", ".cc", ".cxx":
		return "C++"
	case ".h", ".hpp":
		return "C/C++ Header"
	case ".cs":
		return "C#"
	case ".swift":
		return "Swift"
	case ".kt":
		return "Kotlin"
	case ".scala":
		return "Scala"
	case ".html":
		return "HTML"
	case ".css":
		return "CSS"
	case ".scss", ".sass":
		return "SASS"
	case ".md", ".markdown":
		return "Markdown"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".xml":
		return "XML"
	case ".sql":
		return "SQL"
	case ".sh", ".bash":
		return "Shell"
	case ".ps1":
		return "PowerShell"
	case ".bat", ".cmd":
		return "Batch"
	case ".dockerfile", ".containerfile":
		return "Dockerfile"
	case ".vue":
		return "Vue"
	case ".svelte":
		return "Svelte"
	case ".proto":
		return "Protocol Buffer"
	case ".graphql", ".gql":
		return "GraphQL"
	}

	return ""
}
