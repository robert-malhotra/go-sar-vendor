# go-sar-vendor

**One CLI + lightweight Go SDKs for tasking & ordering synthetic‑aperture‑radar (SAR) imagery across multiple commercial vendors.**

> *Airbus OneAtlas • Capella Space • ICEYE v2 • Umbra Space*

---

## ✨ Highlights

* **Unified CLI –** `gosar` wraps four vendor APIs behind a single command‑line interface with discoverable sub‑commands.
* **Minimal, typed SDKs –** idiomatic Go packages (`pkg/*`) with zero external runtime dependencies.
* **Auth & pagination built‑in –** transparent OAuth2 / API‑key handling and lazy iterators using Go 1.23 `iter.Seq2`.
* **Battle‑tested** – exhaustive unit tests via `httptest`; no network calls leave the process.
* **Configurable endpoints** – point the SDKs or CLI at staging environments with a single flag.

---

## Directory layout

```
go-sar-vendor/
├── cmd/           # CLI entry‑points (one per vendor + root)
├── pkg/           # Re‑usable SDK packages
├── makefile       # Lint / vet / test helpers
└── go.{mod,sum}   # Go modules
```

<details>
<summary>Sub‑packages overview</summary>

| Path          | Description                             |
| ------------- | --------------------------------------- |
| `pkg/airbus`  | Airbus OneAtlas Radar SAR SDK & helpers |
| `pkg/capella` | Capella Space Tasking & Access SDK      |
| `pkg/iceye`   | ICEYE Tasking v2 SDK                    |
| `pkg/umbra`   | Umbra Space Tasking SDK                 |

</details>

---

## Requirements

* **Go ≥ 1.24** (uses `iter` & `context` enhancements)
* One or more vendor credentials:

  * `AIRBUS_API_KEY`
  * `CAPELLA_API_KEY`
  * `ICEYE_CLIENT_ID`, `ICEYE_CLIENT_SECRET`
  * `UMBRA_API_KEY` (bearer token)

---

## Installation

### Module consumer

```bash
go get github.com/robert-malhotra/go-sar-vendor@latest
```

### Stand‑alone CLI

```bash
go install github.com/robert-malhotra/go-sar-vendor/cmd/go-sar-vendor@latest
# binary ➝ $GOBIN/gosar
```

---

## Quick start

#### List help

```bash
gosar --help           # global flags + vendor list
gosar capella --help   # vendor‑specific help
```

#### Submit a Capella access‑request

```bash
export CAPELLA_API_KEY="<your key>"
cat request.json | gosar capella access create > access.json
```

#### Price & schedule an ICEYE task

```bash
export ICEYE_CLIENT_ID=… ICEYE_CLIENT_SECRET=…
cat task.json | gosar iceye tasks price
```

---

## Development

```bash
make tidy   # go mod tidy
make vet    # go vet ./...
make lint   # golangci‑lint run
make test   # go test -v ./...
make ci     # run all of the above
```

All unit tests spin up local `httptest.Server` instances – no external connectivity is required.

---

## Versioning

This repository follows **Semantic Versioning 2.0**.  SDKs may evolve at different paces; breaking changes bump the MAJOR version.

---

## Contributing

Pull‑requests and GitHub issues are welcome!  Please run `make ci` before submitting and ensure new code includes unit tests.

---

## License

Distributed under the **MIT License**.  See `LICENSE` for full text.
