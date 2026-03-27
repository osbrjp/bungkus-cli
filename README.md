# bungkus-cli

A CLI tool to **scaffold and configure modern frontend projects** quickly.

`bungkus-cli` sets up projects with common stacks (Astro, Tailwind, Prettier, ESLint) by combining:

- a **Go-based CLI** (with TUI coming soon)
- a **Node/TypeScript AST patch engine** using ts-morph for safe config modifications

The patcher JS is embedded in the Go binary — single binary distribution, no extra files needed.

---

## Usage

```bash
bungkus init . --css tailwindcss --fmt prettier --linter eslint
bungkus init my-app --css tailwindcss --fmt prettier
bungkus init my-app -b astro --css tailwindcss
```

### Flags

| Flag       | Default      | Description              |
|------------|--------------|--------------------------|
| `--base`   | `astro`      | Base framework           |
| `--css`    | `tailwindcss`| CSS framework            |
| `--fmt`    | `prettier`   | Code formatter           |
| `--linter` | `eslint`     | Linter                   |

### What it does

1. Creates the project directory (or uses `.` for current dir)
2. Runs `npm init -y` and patches `package.json` with framework scripts
3. Writes the base framework config (e.g. `astro.config.mjs`)
4. Scaffolds starter files (e.g. `src/pages/index.astro`)
5. Copies extra templates and AST-patches them with framework-specific plugins
6. Runs `npm install` with all required packages

---

## Project Structure

```
bungkus-cli/
├── main.go                         # Entrypoint, embeds patcher JS
├── cmd/
│   ├── root.go                     # Root cobra command
│   └── init.go                     # Init command
├── internal/
│   └── patcher/
│       └── patcher.go              # Patcher cache + runner
├── config/
│   ├── config.go                   # Embed + config loaders
│   ├── setup.json                  # Base framework config (astro)
│   ├── extras/                     # Extra configs (per tool)
│   │   ├── tailwindcss.json
│   │   ├── prettier.json
│   │   └── eslint.json
│   └── templates/                  # Template files (one per tool)
│       ├── astro/
│       │   └── src/pages/index.astro
│       ├── tailwindcss/
│       │   └── global.css
│       ├── prettier/
│       │   └── .prettierrc
│       └── eslint/
│           └── eslint.config.mjs
├── patcher/                        # Node/TS AST patch engine
│   ├── package.json
│   ├── build.ts                    # Bun build config
│   └── src/
│       └── index.ts                # ts-morph patcher (JSON + JS/TS)
├── go.mod
└── go.sum
```

---

## Development

### Prerequisites

- Go 1.24+
- Node.js 22+
- Bun (for building the patcher)

### Build

```bash
# Build the patcher first
cd patcher && bun install && bun run build.ts && cd ..

# Build the CLI (embeds patcher/dist/index.js)
go build -o bungkus-cli .
```

### Install globally

```bash
# Build patcher first, then:
go install .
```

The patcher JS is embedded in the binary and cached at `~/.bungkus/cache/` on first run.

---

## Adding a new base framework

1. Create `config/setup.json` for the framework (or extend to support multiple bases)
2. Add template files under `config/templates/<framework>/`
3. Add framework support in each extras JSON under `config/extras/`

## Adding a new extra (tool)

1. Create `config/extras/<tool>.json` with packages and patch instructions per base
2. Add a template under `config/templates/<tool>/` if the tool needs a config file
3. The patcher handles:
   - **JSON files**: merges arrays via `jsonMerge` (e.g. prettier plugins)
   - **JS/TS files**: injects imports, vite plugins, or array spreads

---

## Tech Stack

- **Go** — CLI, orchestration, binary distribution
- **Node / TypeScript** — AST-based config patching (ts-morph)
- **Bun** — patcher build tooling

---

## Roadmap

- [x] CLI scaffolding flow
- [x] Config patching (JSON + JS/TS AST)
- [x] Embedded patcher (single binary)
- [ ] TUI interface
- [ ] More base frameworks (Nuxt, Next.js, Vite vanilla)
- [ ] More extras (Biome, oxfmt)

---

## License

MIT
