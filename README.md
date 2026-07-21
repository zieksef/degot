# degot

Go service scaffolds-depot.

## Install

```sh
go install github.com/zieksef/degot@latest
```

## Usage

Generate the service directories and base files in the current directory:

```sh
degot gen --service orderservice
```

Generate the service and initialize `go.mod` at the same time:

```sh
degot gen --service orderservice --mod example.com/acme/orderservice
```

`--service` accepts any path; the last path segment becomes the service name.
You can generate into a subdirectory, or anywhere outside the current directory:

```sh
# nested under the current directory
degot gen --service services/orderservice

# an absolute path, outside the current directory
degot gen --service ~/work/orderservice --mod example.com/acme/orderservice
```

## Sync bundled config

`degot sync` copies bundled files onto the machine and overwrites managed
destinations; extra files already at the destinations are left untouched.

```sh
degot sync instructions   # agent instructions → ~/.claude, ~/.codex
degot sync skills         # agent skills → ~/.claude, ~/.codex
degot sync linters        # golangci-lint config → ~/.golangci.yml
degot sync all            # all of the above
```

`--claude-only` / `--codex-only` limit the target for `instructions` and
`skills` (mutually exclusive; not allowed with `all`).

## Generated structure

`<service>` is the name passed to `--service`:

```text
<service>/
├── api/
├── cmd/
│   └── main.go
├── configs/
├── deployments/
├── internal/
│   ├── infra/
│   │   └── conf/
│   ├── pkg/
│   ├── repo/
│   └── svc/
├── migrations/
├── scripts/
├── .gitignore
├── AGENTS.md
├── CLAUDE.md
└── README.md
```
