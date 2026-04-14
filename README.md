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

```bash
git clone https://github.com/kkam05/super.git
cd super
go build -o super src/*.go
```

Move the binary somewhere on your `PATH`:

```bash
mv super /usr/local/bin/super
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
  src/
    main.go
  build/
  lib/
  .super/
    scripts/
      build.sh
      run.sh
      dev.sh
  project.settings
  go.mod
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

### `super version`

Print the installed version.

### `super help`

Show the help message.

---

## project.settings

Projects are configured via a `project.settings` file at the project root, written in TOML.

```toml
[project]
name = "myapp"
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

- **`build.sh`** — compiles `src/main.go` to `build/<name>`
- **`run.sh`** — executes the compiled binary from `build/`
- **`dev.sh`** — runs `go run src/main.go` directly

---

## Requirements

- Go 1.21 or later
- `bash` (macOS / Linux) or PowerShell (Windows)

---

## License

MIT
