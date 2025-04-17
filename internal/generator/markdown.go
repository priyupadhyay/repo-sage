package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/priyupadhyay/repo-sage/internal/analyzer"
)

const markdownTemplate = `# Project Overview: {{.RepoInfo.Name}}

## 📌 Purpose
{{.RepoInfo.Description}}

## 🧠 Architecture
{{.Architecture}}

## 🔍 Components
{{range .RepoInfo.Components}}
### {{.Name}} ({{.Type}})
{{.Description}}
Location: ` + "`" + `{{.Path}}` + "`" + `
{{end}}

## 🚀 Entry Points
{{range .RepoInfo.EntryPoints}}
- ` + "`" + `{{.}}` + "`" + `
{{end}}

## 📦 Dependencies
{{range $dep, $ver := .RepoInfo.Dependencies}}
- {{$dep}}: {{$ver}}
{{end}}

## 🛠 Setup Instructions
{{.Setup}}

{{if .FlowDiagram}}
## 🌀 Flow Diagram
` + "```mermaid" + `
{{.FlowDiagram}}
` + "```" + `
{{end}}

## 📊 Language Statistics
{{range $lang, $pct := .RepoInfo.Languages}}
- {{$lang}}: {{printf "%.1f%%" $pct}}
{{end}}

---
Generated with ❤️ by repo-sage at {{.GeneratedAt}}`

// Generator generates documentation from analysis results
type Generator struct {
	tmpl *template.Template
}

// New creates a new Generator instance
func New() (*Generator, error) {
	tmpl, err := template.New("markdown").Parse(markdownTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Generator{
		tmpl: tmpl,
	}, nil
}

// templateData adds additional fields needed for the template
type templateData struct {
	*analyzer.AnalysisResult
	GeneratedAt string
}

// Generate creates a Markdown document from the analysis results
func (g *Generator) Generate(result *analyzer.AnalysisResult) (string, error) {
	// Sort components by type
	sort.Slice(result.RepoInfo.Components, func(i, j int) bool {
		if result.RepoInfo.Components[i].Type == result.RepoInfo.Components[j].Type {
			return result.RepoInfo.Components[i].Name < result.RepoInfo.Components[j].Name
		}
		return result.RepoInfo.Components[i].Type < result.RepoInfo.Components[j].Type
	})

	// Sort entry points
	sort.Strings(result.RepoInfo.EntryPoints)

	// Sort languages by percentage
	languages := make([]struct {
		Name       string
		Percentage float64
	}, 0, len(result.RepoInfo.Languages))
	for lang, pct := range result.RepoInfo.Languages {
		languages = append(languages, struct {
			Name       string
			Percentage float64
		}{lang, pct})
	}
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Percentage > languages[j].Percentage
	})

	data := templateData{
		AnalysisResult: result,
		GeneratedAt:    time.Now().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Clean up empty sections
	lines := strings.Split(buf.String(), "\n")
	var cleanLines []string
	skipNext := false

	for i, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}

		// Skip empty sections
		if i < len(lines)-1 && strings.HasPrefix(line, "##") && lines[i+1] == "" {
			skipNext = true
			continue
		}

		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n"), nil
}
