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

## Generate kitex code from a remote IDL repository

`kitexgen` wraps the [kitex](https://github.com/cloudwego/kitex) tool: it pulls
the IDL git repository (cached in `~/.kitex/cache`, refreshed with `git pull`)
and generates code into the current directory (`kitex_gen/`). The `kitex`
binary must be installed:

```sh
go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
```

Generate from a proto file inside a remote repository:

```sh
degot kitexgen --repo https://github.com/acme/idl.git --idl order/order.proto
```

`--idl` is the proto file path relative to the IDL repository root. The
repository's default branch is used, and the go module name is inferred from
go.mod. Re-running the same command refreshes `kitex_gen/` after the IDL
changes.

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
└── README.md
```
