# Armada

> The reproducible security toolbox. One YAML, every platform, zero sudo.

[![Build](https://github.com/DiegoDev2/armada/actions/workflows/build.yml/badge.svg)](https://github.com/DiegoDev2/armada/actions/workflows/build.yml)
[![Test](https://github.com/DiegoDev2/armada/actions/workflows/test.yml/badge.svg)](https://github.com/DiegoDev2/armada/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/DiegoDev2/armada)](https://goreportcard.com/report/github.com/DiegoDev2/armada)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Armada is a user-space package manager for security tools, CLI utilities and
reproducible developer toolchains. You describe a tool once in a YAML manifest
and Armada downloads the right asset for your OS and architecture, verifies
the SHA-256 checksum and links the binaries into `~/.armada/bin`. No `sudo`,
no system package manager, no surprise mutations to `/usr`.

> **Note on the name:** this repository used to be called *Fleet*. It has
> been renamed to **Armada** to avoid trademark and SEO collisions with
> JetBrains Fleet and FleetDM. GitHub keeps a permanent redirect from the old
> URL, and the Go module path is now `github.com/DiegoDev2/armada`.

## Why Armada

- **Security-first workflow.** Pentesters and CTF players need `ffuf`,
  `nuclei`, `httpx`, `subfinder`, `gobuster`, `amass`, and a dozen other
  Go/Rust tools on every new VM. Armada makes that a one-liner.
- **Reproducible.** Every install is pinned to a version and a SHA-256 digest.
  `armada list` shows exactly what is on the machine.
- **Cross-platform.** A single manifest covers `linux/amd64`, `linux/arm64`,
  `darwin/amd64`, `darwin/arm64`, `windows/amd64` and `windows/arm64`.
- **No sudo.** Everything lives under `~/.armada/`. Uninstall is `rm -rf`.
- **Simple manifest format.** One YAML, a map of `"<os>/<arch>"` keys to asset
  URLs and checksums. No DSL, no Ruby, no scripting required.

## Install

### One-liner (Linux and macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/DiegoDev2/armada/main/scripts/install.sh | sh
```

The installer drops the `armada` binary into `~/.armada/bin`. Add that
directory to your `PATH`:

```sh
echo 'export PATH="$HOME/.armada/bin:$PATH"' >> ~/.bashrc   # or ~/.zshrc
```

### From source

```sh
go install github.com/DiegoDev2/armada/cmd/armada@latest
```

### Pre-built binaries

See the [Releases](https://github.com/DiegoDev2/armada/releases) page for
`tar.gz` / `zip` archives and `.sha256` files.

## Quickstart

```sh
# Preview the install plan for a local manifest
armada simulate examples/ffuf.yaml

# Install from a local manifest
armada install --from examples/ffuf.yaml

# What's installed?
armada list

# Register the default registry and search it
armada repo add armada-default https://github.com/DiegoDev2/armada-registry --type git --priority 100
armada repo sync
armada search scan

# Install a tool by name
armada install ffuf

# Uninstall
armada uninstall ffuf
```

## Manifest format

A manifest describes a single tool. The only required fields are `name`,
`version` and at least one asset.

```yaml
name: ffuf
version: 2.1.0
description: Fast web fuzzer written in Go
homepage: https://github.com/ffuf/ffuf
license: MIT
categories: [security, web, fuzzing]

binaries: [ffuf]

assets:
  linux/amd64:
    url: https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_linux_amd64.tar.gz
    checksum: sha256:<hex>
    type: tar.gz
  linux/arm64:
    url: https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_linux_arm64.tar.gz
    checksum: sha256:<hex>
    type: tar.gz
  darwin/amd64:
    url: https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_macOS_amd64.tar.gz
    checksum: sha256:<hex>
    type: tar.gz
  darwin/arm64:
    url: https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_macOS_arm64.tar.gz
    checksum: sha256:<hex>
    type: tar.gz
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Canonical tool name, used for lookups and the on-disk package directory. |
| `version` | yes | Version string. Armada records this verbatim in the state file. |
| `description` | no | One-line description shown by `armada list` and `armada search`. |
| `homepage`, `license`, `categories` | no | Purely informational. |
| `binaries` | no | Explicit list of binary names to symlink. Defaults to `[name]`. |
| `assets` | at least one | Map of `"<os>/<arch>"` to asset definitions. |
| `assets.<key>.url` | yes | HTTPS URL to the archive or raw binary. |
| `assets.<key>.checksum` | yes | `<algo>:<hex>`. Currently `sha256` only. |
| `assets.<key>.type` | no | `tar.gz`, `tar.bz2`, `tar`, `zip`, or `raw`. Inferred from the URL when omitted. |
| `assets.<key>.strip_components` | no | Number of leading path components to strip when extracting. |

### Layout on disk

```
~/.armada/
├── bin/                    # symlinks into packages (add to PATH)
├── pkgs/<name>/<version>/  # extracted package contents
├── cache/                  # downloaded archives
├── repos/                  # cloned/fetched registries
├── config.yaml             # repository list
└── state.json              # what is installed
```

Override the root with `ARMADA_HOME=/path/to/dir`.

## Repositories

Armada resolves tool names by walking a list of registries. Each registry is a
git repo or an HTTP index that contains manifest files.

```sh
armada repo add armada-default https://github.com/DiegoDev2/armada-registry --type git --priority 100
armada repo list
armada repo sync
```

Higher `--priority` wins when multiple registries provide the same tool.

## Security posture

- All downloads go over HTTPS.
- Every archive is verified against the manifest's SHA-256 digest before
  extraction. A mismatch aborts the install and leaves the cache file on disk
  for inspection.
- Archive extraction rejects any entry that would escape the destination
  directory (no Zip-Slip).
- Nothing runs as root. Post-install hooks are **not** executed in v0.1.

Future versions will add Sigstore / minisign signatures on manifests and SBOM
generation.

## Roadmap

v0.1 (this release) delivers the MVP: `install`, `uninstall`, `list`,
`search`, `repo`, `simulate`, and a modest default registry.

Planned next:

- v0.2 — Dependency resolution, `armada upgrade`, `armada doctor`.
- v0.3 — Manifest signing (Sigstore / minisign), SBOM output, `armada audit`.
- v0.4 — Team lockfiles (`armada.lock`) for reproducible CI caches.
- v0.5 — Optional `armada agent` that keeps a pinned toolchain in sync.

## Contributing

Issues and pull requests are welcome — especially new manifests for the
[default registry](https://github.com/DiegoDev2/armada-registry). See
[CONTRIBUTING.md](CONTRIBUTING.md) for the local development workflow.

## License

[MIT](LICENSE). See [SECURITY.md](SECURITY.md) for responsible disclosure.
