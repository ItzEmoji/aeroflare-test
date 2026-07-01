---
sidebar_position: 3
title: Installation
---

# Installation

Aeroflare distributes binaries via Nix flakes, standard Go toolchains, and pre-compiled GitHub releases. 

## Option 1: Nix Flakes (Recommended)

If you have a modern Nix installation with flakes enabled, you can run Aeroflare directly without installing it globally.

```bash
nix run github:aeroflare/aeroflare -- --help
```

To install it into your user profile permanently:

```bash
nix profile install github:aeroflare/aeroflare
```

## Option 2: Building from Source (Go)

Aeroflare requires Go 1.21 or later. 

1. Clone the repository:
   ```bash
   git clone https://github.com/aeroflare/aeroflare.git
   cd aeroflare
   ```

2. Compile the binary:
   ```bash
   go build -o aeroflare main.go
   ```

3. Move the binary to your `PATH`:
   ```bash
   sudo mv aeroflare /usr/local/bin/
   ```

## Option 3: Pre-compiled Binaries

Statically linked binaries for Linux and macOS are attached to every GitHub Release.

1. Navigate to the [Releases page](https://github.com/aeroflare/aeroflare/releases).
2. Download the archive corresponding to your OS and architecture.
3. Extract and place the binary in your `PATH`.

## Verification

Verify the installation by checking the version output.

```bash
aeroflare version
```
