package main

import (
	"fmt"
	"os"

	"github.com/priyupadhyay/repo-sage/internal/analyzer"
	"github.com/priyupadhyay/repo-sage/internal/config"
	"github.com/priyupadhyay/repo-sage/internal/generator"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repo-sage",
	Short: "repo-sage - Let your codebase speak its truth",
	Long: `repo-sage is a powerful CLI tool that analyzes Git repositories and generates 
comprehensive documentation using AI. It helps you understand codebases by identifying
key components, architecture, and generating clear documentation.`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a Git repository",
	Long: `Analyze a Git repository and generate comprehensive documentation.
By default, performs a quick analysis of the repository structure and key files.
Use --detailed for in-depth code analysis.

Example: repo-sage analyze --repo /path/to/repo --output docs/overview.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath, _ := cmd.Flags().GetString("repo")
		outputPath, _ := cmd.Flags().GetString("output")
		profileName, _ := cmd.Flags().GetString("profile")
		contextSize, _ := cmd.Flags().GetInt("context")
		detailed, _ := cmd.Flags().GetBool("detailed")

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get profile
		var profile config.Profile
		if profileName != "" {
			p, exists := cfg.GetProfile(profileName)
			if !exists {
				return fmt.Errorf("profile %q not found", profileName)
			}
			profile = p
		} else {
			p, _, err := cfg.GetDefaultProfile()
			if err != nil {
				return fmt.Errorf("no profile configured. Run 'repo-sage config add-profile' to get started")
			}
			profile = p
		}

		// Create analyzer
		a, err := analyzer.NewAnalyzer(analyzer.AnalyzeOptions{
			OpenAIKey:   profile.APIKey,
			APIBase:     profile.APIBase,
			Model:       profile.Model,
			ContextSize: contextSize,
			Detailed:    detailed,
		})
		if err != nil {
			return fmt.Errorf("failed to create analyzer: %w", err)
		}

		// Analyze repository
		result, err := a.Analyze(repoPath, analyzer.AnalyzeOptions{
			OpenAIKey:   profile.APIKey,
			APIBase:     profile.APIBase,
			Model:       profile.Model,
			ContextSize: contextSize,
			Detailed:    detailed,
			OutputPath:  outputPath,
		})
		if err != nil {
			return fmt.Errorf("failed to analyze repository: %w", err)
		}

		// Generate documentation
		gen, err := generator.New()
		if err != nil {
			return fmt.Errorf("failed to create generator: %w", err)
		}

		doc, err := gen.Generate(result)
		if err != nil {
			return fmt.Errorf("failed to generate documentation: %w", err)
		}

		// Write output
		if err := os.WriteFile(outputPath, []byte(doc), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}

		fmt.Printf("✨ Analysis complete! Documentation saved to %s\n", outputPath)
		return nil
	},
}

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain a specific file",
	Long: `Generate a detailed explanation of a specific file in the repository.
Example: repo-sage explain --file path/to/file.go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		profileName, _ := cmd.Flags().GetString("profile")
		contextSize, _ := cmd.Flags().GetInt("context")

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get profile
		var profile config.Profile
		if profileName != "" {
			p, exists := cfg.GetProfile(profileName)
			if !exists {
				return fmt.Errorf("profile %q not found", profileName)
			}
			profile = p
		} else {
			p, _, err := cfg.GetDefaultProfile()
			if err != nil {
				return fmt.Errorf("no profile configured. Run 'repo-sage config add-profile' to get started")
			}
			profile = p
		}

		// Create analyzer
		a, err := analyzer.NewAnalyzer(analyzer.AnalyzeOptions{
			OpenAIKey:   profile.APIKey,
			APIBase:     profile.APIBase,
			Model:       profile.Model,
			ContextSize: contextSize,
		})
		if err != nil {
			return fmt.Errorf("failed to create analyzer: %w", err)
		}

		// Explain file
		explanation, err := a.ExplainFile(filePath, analyzer.ExplainOptions{
			ContextSize: contextSize,
			OpenAIKey:   profile.APIKey,
			APIBase:     profile.APIBase,
			Model:       profile.Model,
		})
		if err != nil {
			return fmt.Errorf("failed to explain file: %w", err)
		}

		fmt.Println(explanation)
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage repo-sage configuration",
	Long:  `Configure LLM endpoints and profiles for repo-sage.`,
}

var addProfileCmd = &cobra.Command{
	Use:   "add-profile [name]",
	Short: "Add or update a profile",
	Long:  `Add a new profile or update an existing profile with LLM endpoint configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		apiBase, _ := cmd.Flags().GetString("api-base")
		apiKey, _ := cmd.Flags().GetString("api-key")
		model, _ := cmd.Flags().GetString("model")

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		profile := config.Profile{
			APIBase: apiBase,
			APIKey:  apiKey,
			Model:   model,
		}

		cfg.AddProfile(name, profile)

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Profile %q added successfully\n", name)
		return nil
	},
}

var listProfilesCmd = &cobra.Command{
	Use:   "list-profiles",
	Short: "List all configured profiles",
	Long:  `Display all configured profiles and their settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles configured.")
			fmt.Println("Run 'repo-sage config add-profile' to add a profile.")
			return nil
		}

		fmt.Println("Configured profiles:")
		for name, profile := range cfg.Profiles {
			defaultMark := " "
			if name == cfg.DefaultProfile {
				defaultMark = "*"
			}
			fmt.Printf("%s %s:\n", defaultMark, name)
			fmt.Printf("  API Base: %s\n", profile.APIBase)
			fmt.Printf("  Model: %s\n", profile.Model)
			fmt.Printf("  API Key: %s\n", maskAPIKey(profile.APIKey))
			fmt.Println()
		}

		return nil
	},
}

var useProfileCmd = &cobra.Command{
	Use:   "use-profile [name]",
	Short: "Set the default profile",
	Long:  `Set which profile should be used by default for LLM operations.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.SetDefaultProfile(name); err != nil {
			return err
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Default profile set to %q\n", name)
		return nil
	},
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func init() {
	// Analyze command flags
	analyzeCmd.Flags().StringP("repo", "r", "", "Path to the Git repository")
	analyzeCmd.Flags().StringP("output", "o", "SUMMARY.md", "Output file path")
	analyzeCmd.Flags().String("profile", "", "Profile to use for LLM operations")
	analyzeCmd.Flags().Int("context", 4000, "Context size for AI analysis")
	analyzeCmd.Flags().Bool("detailed", false, "Perform detailed code analysis")
	analyzeCmd.MarkFlagRequired("repo")

	// Explain command flags
	explainCmd.Flags().StringP("file", "f", "", "Path to the file to explain")
	explainCmd.Flags().String("profile", "", "Profile to use for LLM operations")
	explainCmd.Flags().Int("context", 4000, "Context size for AI analysis")
	explainCmd.MarkFlagRequired("file")

	// Add commands to root
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(explainCmd)

	// Add config commands
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(addProfileCmd)
	configCmd.AddCommand(listProfilesCmd)
	configCmd.AddCommand(useProfileCmd)

	addProfileCmd.Flags().String("api-base", "", "API base URL for the LLM endpoint")
	addProfileCmd.Flags().String("api-key", "", "API key for authentication")
	addProfileCmd.Flags().String("model", "", "Model name to use")

	addProfileCmd.MarkFlagRequired("api-base")
	addProfileCmd.MarkFlagRequired("api-key")
	addProfileCmd.MarkFlagRequired("model")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
