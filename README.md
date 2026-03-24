# 🥡 bungkus-cli

A CLI tool to **scaffold and configure modern frontend projects** quickly.

`bungkus-cli` helps you set up projects with common stacks (e.g. Astro, Vite, Tailwind, GSAP) by combining:

- a **Go-based CLI/TUI**
- a **Node/TypeScript patch engine** for safe config modifications

---

## ✨ Features

- 🚀 Scaffold frontend projects fast
- 🧩 Add libraries and presets (Tailwind, GSAP, etc.)
- 🛠 Patch config files (`package.json`, `vite.config.ts`, etc.)
- 🖥 Interactive TUI (coming soon)
- ⚡ Native CLI (Go)

---

## 📦 Project Structure

```
bungkus-cli/
├─ cmd/
│  └─ bungkus/
│     └─ main.go          # CLI entrypoint
├─ internal/
│  ├─ runner/             # command + process runner
│  │  └─ runner.go
│  ├─ presets/            # preset definitions
│  │  └─ presets.go
│  └─ tui/                # TUI (future)
│     └─ tui.go
├─ patcher/               # Node/TS patch engine
│  ├─ package.json
│  ├─ tsconfig.json
│  └─ src/
│     ├─ index.ts
│     ├─ patch-package-json.ts
│     ├─ patch-vite-config.ts
│     └─ patch-astro-config.ts
├─ examples/              # example outputs / templates
├─ go.mod
└─ README.md
```

---

## 🧠 How it works

1. Go CLI handles:
   - user input
   - project creation
   - running package managers

2. Node patcher handles:
   - modifying config files safely
   - merging dependencies
   - injecting plugins into configs

---

## 🚀 Getting Started

### 1. Build the CLI

```bash
go build -o bungkus ./cmd/fe-init
```

### 2. Build the patcher

```bash
cd patcher
bun run build
```

### 3. Run

```bash
./bungkus
```

---

## ⚙️ Example Usage (planned)

```bash
bungkus new my-app --preset vite-react-tailwind-gsap
```

---

## 🧱 Tech Stack

- Go — CLI + orchestration  
- Node / TypeScript — config patching  
- Bun — build tooling for patcher  

---

## 🗺 Roadmap

- [ ] CLI scaffolding flow  
- [ ] Preset system  
- [ ] TUI interface  
- [ ] More framework support (Next, Astro, Svelte, etc.)  
- [ ] Plugin system  

---

## 📄 License

MIT
