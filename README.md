# degot

Go service scaffolds-depot.

## Install

```sh
go install github.com/zieksef/degot@latest
```

## Usage

Generate the service directories and base files in the current directory:

```sh
degot --service orderservice
```

Generate the service and initialize `go.mod` at the same time:

```sh
degot --service orderservice --mod example.com/acme/orderservice
```

`--service` accepts any path; the last path segment becomes the service name.
You can generate into a subdirectory, or anywhere outside the current directory:

```sh
# nested under the current directory
degot --service services/orderservice

# an absolute path, outside the current directory
degot --service ~/work/orderservice --mod example.com/acme/orderservice
```

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
│   ├── cfg/
│   ├── infra/
│   ├── repo/
│   └── svc/
├── migrations/
├── scripts/
├── .gitignore
├── AGENTS.md
└── README.md
```
