# 🛠️ repo-sage  
> *Let your codebase speak its truth.*

**repo-sage** is a powerful and elegant CLI tool that analyzes your Git repository and generates a beautifully structured documentation summary using AI ✨

![demo](demo.svg)

---

## 🚀 Features

- 🔗 Connects to **OpenAI** or **Ollama** for AI-powered analysis
- 📂 Scans local Git repositories
- 🧠 Identifies:
  - Programming languages
  - Main components (API, CLI, services, utils, etc.)
  - Entry points and dependencies
  - Architecture and code flow
- 📝 Generates structured **Markdown** documentation:
  - Project overview and purpose
  - Architecture breakdown
  - Key file summaries
  - Setup instructions
  - Optional Mermaid diagrams

---

## 🔧 Installation
Clone the repository and build from source:

```bash
git clone https://github.com/priyupadhyay/repo-sage.git
cd repo-sage
go build -o repo-sage
```

Install the binary to your system:

```bash
sudo install repo-sage /usr/local/bin/
```

---

## 📚 Usage

### Basic analysis:
```bash
repo-sage analyze --repo /path/to/repo --output docs/overview.md
```

### Options:
```bash
# Use OpenAI
repo-sage analyze --repo ./my-project --openai-key sk-xxx

# Use Ollama
repo-sage analyze --repo ./my-project --ollama

# Customize token context size
repo-sage analyze --repo ./my-project --context 5000

# Explain a specific file
repo-sage explain --file path/to/file.go
```

---

## 🏗️ Architecture

- Written in **Go**
- (Optional) Uses **Tree-sitter** for advanced code parsing
- Integrates with **OpenAI** / **Ollama** APIs
- Template-based **Markdown** generation
- Intelligent repository scanning & analysis

---

## 🤝 Contributing

Contributions are welcome!  
Feel free to fork the project and submit a Pull Request 🤗

---

## 📝 License

[MIT](LICENSE)

---

Created with ❤️ by [@priyupadhyay](https://github.com/priyupadhyay)