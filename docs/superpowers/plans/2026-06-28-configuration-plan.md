# Aeroflare Configuration Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a unified configuration system using Viper that supports CLI flags, environment variables, and a YAML configuration file.

**Architecture:** We will initialize Viper in `cmd/root.go`'s `init` function (or `Execute`), load a YAML config from `$XDG_CONFIG_HOME/aeroflare/aeroflare.yaml`, and automatically bind environment variables prefixed with `AEROFLARE`. If the file doesn't exist, we will create it with commented defaults. We will update `cmd/init.go`, `src/init/wizard.go`, and `src/init/theme.go` to consume these variables and skip interactive prompts when applicable.

**Tech Stack:** Go, Cobra, Viper, Huh (for terminal UI)

## Global Constraints

- Configuration file must be generated at `$XDG_CONFIG_HOME/aeroflare/aeroflare.yaml` (defaulting to `$HOME/.config/aeroflare/aeroflare.yaml`) if it does not exist.
- Environment variables must be prefixed with `AEROFLARE_`.
- `AEROFLARE_CACHE_URL` takes precedence over `AEROFLARE_CACHE`.

---

### Task 1: Add Viper Dependency

**Files:**
- Modify: `go.mod`, `go.sum`

**Interfaces:**
- Consumes: N/A
- Produces: `github.com/spf13/viper` available for imports.

- [ ] **Step 1: Get Viper**

```bash
go get github.com/spf13/viper@v1.19.0
```

- [ ] **Step 2: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add viper dependency"
```

### Task 2: Viper Initialization and Auto-generation in Root Command

**Files:**
- Modify: `cmd/root.go`

**Interfaces:**
- Consumes: `cobra` package.
- Produces: Viper initialized, config file generated if missing, env vars bound, `--cache-url` flag added.

- [ ] **Step 1: Write Viper Initialization Code in `root.go`**

```go
// Add import: "github.com/spf13/viper"
// Add import: "path/filepath"

// Add flag variables at the top of cmd/root.go
var cacheURL string

func initConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("HOME") + "/.config"
	}
	aeroDir := filepath.Join(configDir, "aeroflare")
	
	if err := os.MkdirAll(aeroDir, 0755); err != nil {
		PrintError("Could not create config directory: " + err.Error())
	}

	configFile := filepath.Join(aeroDir, "aeroflare.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := []byte(`# Aeroflare Configuration
# theme: catppuccin
# cache-url: oci://docker.io/my-org/my-cache
# backend: r2
`)
		os.WriteFile(configFile, defaultConfig, 0644)
	}

	viper.SetConfigFile(configFile)
	viper.SetEnvPrefix("AEROFLARE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			PrintError("Error reading config file: " + err.Error())
		}
	}
}

// Modify init() in cmd/root.go
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().CountVarP(&VerboseCount, "verbose", "v", "Enable verbose output (-v for packages, -vv for requests)")
	rootCmd.PersistentFlags().StringVar(&cacheURL, "cache-url", "", "OCI registry URL for the cache")
	viper.BindPFlag("cache-url", rootCmd.PersistentFlags().Lookup("cache-url"))
}
```

- [ ] **Step 2: Run build to verify it compiles**

Run: `go build -o aeroflare main.go`
Expected: Compile succeeds

- [ ] **Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "feat: initialize viper and auto-generate config"
```

### Task 3: Theme Configuration Integration

**Files:**
- Modify: `src/init/theme.go`

**Interfaces:**
- Consumes: `viper.GetString("theme")`
- Produces: Custom `huh.Theme` based on config.

- [ ] **Step 1: Modify `AeroflareTheme()` to use Viper**

```go
// Add import "github.com/spf13/viper"
// In src/init/theme.go
func AeroflareTheme() *huh.Theme {
	t := huh.ThemeBase()
	themeName := viper.GetString("theme")

	var primaryColor lipgloss.Color
	var secondaryColor lipgloss.Color

	switch themeName {
	case "catppuccin":
		primaryColor = lipgloss.Color("#cba6f7") // Mauve
		secondaryColor = lipgloss.Color("#585b70") // Surface2
	case "gruvbox-dark":
		primaryColor = lipgloss.Color("#fe8019") // Orange
		secondaryColor = lipgloss.Color("#504945") // Bg2
	case "gruvbox-light":
		primaryColor = lipgloss.Color("#af3a03") // Orange
		secondaryColor = lipgloss.Color("#ebdbb2") // Bg1
	default:
		primaryColor = lipgloss.Color("#00FFFF") // Cyan
		secondaryColor = lipgloss.Color("#555555") // Gray
	}

	t.Focused.Base = t.Focused.Base.Border(lipgloss.RoundedBorder()).BorderForeground(primaryColor)
	t.Focused.Title = t.Focused.Title.Foreground(primaryColor).Bold(true)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(primaryColor)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(primaryColor)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(primaryColor)

	t.Blurred.Base = t.Blurred.Base.Border(lipgloss.RoundedBorder()).BorderForeground(secondaryColor)
	t.Blurred.Title = t.Blurred.Title.Foreground(secondaryColor)

	return t
}
```

- [ ] **Step 2: Run build to verify**

Run: `go build -o aeroflare main.go`
Expected: Compile succeeds

- [ ] **Step 3: Commit**

```bash
git add src/init/theme.go
git commit -m "feat: integrate themes via viper config"
```

### Task 4: Wizard Configuration Skipping

**Files:**
- Modify: `src/init/wizard.go`

**Interfaces:**
- Consumes: Viper configuration values (`viper.GetString()`).
- Produces: `InitConfig` struct filled either from user prompt or direct config.

- [ ] **Step 1: Modify `RunWizard` to read Viper config and bypass prompts**

```go
// In src/init/wizard.go
// Add import "github.com/spf13/viper"
// Modify RunWizard to check viper before prompting. For each field, if viper has a value, set it and skip the prompt.
// Example for CacheName:
/*
func RunWizard() (*InitConfig, error) {
	cfg := &InitConfig{}

	cacheURL := viper.GetString("cache-url")
	cacheName := viper.GetString("cache")

	if cacheURL != "" {
		// Parse URL or use directly
		cfg.Registry = "custom" // simplified, extract from URL logic
		cfg.CacheName = cacheURL
	} else if cacheName != "" {
		cfg.CacheName = cacheName
		cfg.Registry = "ghcr.io"
	}

	// For inputs, we must conditionally add them to huh.NewForm if not provided
	var groups []*huh.Group
	
	if cfg.CacheName == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("What is your GitHub org / username?").
				Value(&cfg.CacheName),
		))
	}
	
	// Example for Backend
	backendVal := viper.GetString("backend")
	if backendVal != "" {
		cfg.Backend = BackendType(backendVal)
	} else {
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[BackendType]().
				Title("Choose index storage backend").
				Options(
					huh.NewOption("Cloudflare R2", BackendR2),
					huh.NewOption("Native OCI Tags", BackendNative),
				).
				Value(&cfg.Backend),
		))
	}

	if len(groups) > 0 {
		form := huh.NewForm(groups...).WithTheme(AeroflareTheme())
		if err := form.Run(); err != nil {
			return nil, err
		}
	}
	
	cfg.DeriveDefaults()
	return cfg, nil
}
*/
```

- [ ] **Step 2: Run build to verify**

Run: `go build -o aeroflare main.go`
Expected: Compile succeeds

- [ ] **Step 3: Commit**

```bash
git add src/init/wizard.go
git commit -m "feat: use viper config to skip wizard prompts"
```
