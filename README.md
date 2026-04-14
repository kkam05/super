# super

An npm-style project manager for Go. Scaffold new projects, define scripts, and run them — all from one CLI.

```
 ____  _   _ ____  _____ ____
/ ___|| | | |  _ \| ____|  _ \
\___ \| | | | |_) |  _| | |_) |
 ___) | |_| |  __/| |___|  _ <
|____/ \___/|_|   |_____|_| \_\
```

---

## Installation

### From a GitHub release (recommended)

Download the latest release zip for your platform from the [releases page](https://github.com/kkam05/super/releases), then run `super update --local` from inside the super source directory, or manually extract the binary to `~/.super/bin/`.

Once installed, run `super path` to add `~/.super/bin` to your PATH automatically.

### From source

```bash
git clone https://github.com/kkam05/super.git
cd super
go build -o build/super src/*.go
./build/super update --local
./build/super path
```

---

## Updating

Pull and install the latest release from GitHub:

```bash
super update
```

Install from a local build (useful when developing super itself):

```bash
super run build          # compiles to build/super
super update --local     # installs build/super to ~/.super/bin
```

---

## Commands

### `super new [project-name] [-y]`

Scaffold a new Go project. If `project-name` is omitted, the current directory is used.

```bash
super new myapp       # creates myapp/ and scaffolds inside it
super new             # scaffolds in the current directory
super new myapp -y    # overwrite even if the directory is not empty
```

The scaffolded project includes:

```
myapp/
├── src/
│   └── main.go
├── build/
├── lib/
├── .super/
│   └── scripts/
│       ├── build.sh
│       ├── run.sh
│       └── dev.sh
├── project.settings
└── go.mod
```

### `super run <script> [args...]`

Run a script defined in `project.settings`. Args are forwarded to the script.

```bash
super run build
super run dev
super run lint --fix
```

Scripts are resolved in order:

1. **Directory** — if the value is a path to a directory (e.g. `.super/scripts`), super looks for `<dir>/<script>.sh` inside it.
2. **File** — if the value is a direct path to a script file, it is executed directly.
3. **Inline** — if the value contains spaces or does not resolve to a path, it is run as a shell command.

If a script is not registered in `project.settings` but a matching `.sh` file exists in `.super/scripts/`, super will auto-register it and run it.

### `super fix`

Repair a project to match the expected super structure. Ensures dirs, `project.settings`, scripts, and the version file are all present and correct.

```bash
super fix
```

### `super update [--local]`

Pull and install the latest release from GitHub, or install from a local build.

```bash
super update           # download and install the latest release from GitHub
super update --local   # install from build/super in the current project
```

The binary is installed to `~/.super/bin/super`. A backup of the previous version is kept in `~/.super/backup/`.

### `super path`

Check whether `~/.super/bin` is on your `PATH`. If it is not, super adds the entry to your shell config file automatically.

```bash
super path
```

Supported shells: zsh, bash, fish. On Windows, the user `PATH` is updated in the registry via PowerShell.

### `super dev <subcommand>`

Developer utilities for working on super itself.

#### `super dev release --local`

Package the local build binary into a release zip ready to upload to GitHub Releases.

```bash
super run build               # build the binary first
super dev release --local     # creates release/super-<GOOS>-<GOARCH>.zip
```

The zip name matches what `super update` expects, so you can upload it directly to a GitHub release tagged `v<version>`.

### `super version`

Print the installed version.

```bash
super version   # super v0.1.4
```

### `super help`

Show the help message.

---

## project.settings

Projects are configured via a `project.settings` file at the project root, written in TOML.

```toml
[project]
name = "myapp"
version = "0.1.0"
description = "my Go application"

[scripts]
build = ".super/scripts"
run   = ".super/scripts"
dev   = ".super/scripts"
lint  = "go vet ./..."
```

Script values can be:

| Value | Behaviour |
|---|---|
| `.super/scripts` | Runs `.super/scripts/<script>.sh` |
| `path/to/script.sh` | Runs the script file directly |
| `go vet ./...` | Runs as an inline shell command |

---

## Default Scripts

When you run `super new`, the following scripts are generated inside `.super/scripts/`:

- **`build.sh`** — compiles `src/*.go` to `build/<name>`, auto-incrementing the patch version in `project.settings`, `src/main.go`, and `.super/version`
- **`run.sh`** — executes the compiled binary from `build/`
- **`dev.sh`** — runs `go run src/main.go` directly

---

## Requirements

- Go 1.21 or later
- `bash` (macOS / Linux) or PowerShell (Windows)

---

## License

MIT
