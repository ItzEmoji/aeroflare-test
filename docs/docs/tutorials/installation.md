---
sidebar_position: 3
title: Installation
---

# Installation

Aeroflare distributes binaries via Nix flakes and standard Go toolchains. 

## Option 1: Nix Flakes (Recommended)

If you have a modern Nix installation with flakes enabled, you can run Aeroflare directly without installing it globally.

```bash
nix run github:ItzEmoji/aeroflare -- --help
```

To install it into your user profile permanently:

```bash
nix profile install github:ItzEmoji/aeroflare
```

## Option 2: Building from Source (Go)

Aeroflare requires Go 1.21 or later. 

1. Clone the repository:
   ```bash
   git clone https://github.com/ItzEmoji/aeroflare.git
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


## Verification

Verify the installation by checking the version output.

```bash
aeroflare version
```
