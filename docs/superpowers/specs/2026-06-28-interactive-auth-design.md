# Interactive Auth Wizard Design

## Overview
The `aeroflare auth` command currently requires users to pass secrets via flags (e.g., `--github-token`). This project adds a fully interactive wizard using the `charmbracelet/huh` TUI library to make onboarding and secret management seamless. 

## Requirements
- Provide an interactive main menu to select which service to authenticate.
- Support a GitHub Device Auth flow (via OAuth) or manual token entry.
- Support direct entry of Cloudflare credentials.
- Support arbitrary OCI registries.
- Automatically save GitHub tokens in a way that implies usage for `ghcr.io` and inform the user.
- Persist all credentials using the existing `secrets.Manager`.

## User Interface & Flow

The wizard uses `huh` forms to present the following flow:

### Main Menu
- Prompt: "What do you want to authenticate?"
- Options: 
  - `GitHub / GitLab`
  - `Cloudflare`
  - `Custom OCI Registry`

### GitHub Flow
- **Sub-prompt**: "How would you like to authenticate?" -> [`Device Auth Flow (Browser)`, `Enter Token Manually`]
- **Device Flow**:
  - Request a device code from `https://github.com/login/device/code` using Client ID `Ov23liIJyLpd2Cse5gne`.
  - Display the `user_code` and the `verification_uri`.
  - Poll `https://github.com/login/oauth/access_token` until the user completes the browser flow (handling `authorization_pending`, `slow_down`, etc).
  - Extract the access token.
- **Save**: Save token as `github-token`.
- **Confirmation**: Print "Success! Token saved. This will automatically be used for GitHub APIs and the ghcr.io container registry."

### Cloudflare Flow
- Prompt 1 (Input): "Cloudflare API Token"
- Prompt 2 (Input): "Cloudflare User ID"
- Save as `cf-token` and `cf-user-id`.

### Custom OCI Registry Flow
- Prompt 1 (Input): "Registry URL (e.g. registry.gitlab.com)"
- Prompt 2 (Input): "Username"
- Prompt 3 (Password): "Token / Password"
- Save as `oci-<registry>-username` and `oci-<registry>-token`.

## Architecture & Code Organization
- **CLI Command (`cmd/auth.go`)**: Modify the `authCmd` to invoke the interactive wizard when no flags are passed.
- **Wizard Logic (`cmd/auth_wizard.go`)**: Create a new file to encapsulate the `huh` form logic and flows.
- **GitHub Auth (`src/auth/github.go`)**: Create a new package or file to isolate the GitHub Device Flow HTTP logic (requesting device code, polling).
- **Dependencies**: Add `github.com/charmbracelet/huh` for the interactive TUI.

## Error Handling
- Errors during GitHub HTTP polling (e.g., timeouts, expired tokens) must be handled gracefully and presented to the user.
- If saving to `secrets.Manager` fails (e.g., keychain locked), the error must be logged clearly.

## Testing Strategy
- Unit tests for the GitHub Device API logic using `httptest` to mock the GitHub endpoints.
- Ensure the interactive flow gracefully skips or can be bypassed in non-interactive CI environments if needed.
