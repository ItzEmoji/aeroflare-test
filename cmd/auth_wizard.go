package cmd

import (
	"fmt"
	"aeroflare/src/auth"
	"github.com/charmbracelet/huh"
)

// githubClientID identifies the Aeroflare OAuth app used for the GitHub
// device authorization flow below; it's a public client ID, not a secret.
const githubClientID = "Ov23liIJyLpd2Cse5gne"

// runInteractiveGithubAuth prompts the user to choose between the OAuth
// device flow (opens a browser) or pasting a personal access token
// manually, then saves whichever token results via the secrets manager.
func runInteractiveGithubAuth() string {
	manager := getSecretsManager()
	var ghMethod string
	err := huh.NewSelect[string]().
		Title("How would you like to authenticate with GitHub?").
		Options(
			huh.NewOption("Device Auth Flow (Browser)", "device"),
			huh.NewOption("Enter Token Manually", "manual"),
		).
		Value(&ghMethod).
		Run()
	if err != nil {
		return ""
	}

	var token string
	if ghMethod == "device" {
		fmt.Println("Requesting device code...")
		res, err := auth.RequestDeviceCode(githubClientID)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to request code: %v", err))
			return ""
		}
		fmt.Printf("Please go to %s and enter the code: %s\n", res.VerificationURI, res.UserCode)
		fmt.Println("Waiting for authorization...")
		
		token, err = auth.PollAccessToken(githubClientID, res.DeviceCode, res.Interval)
		if err != nil {
			PrintError(fmt.Sprintf("Authorization failed: %v", err))
			return ""
		}
	} else {
		err = huh.NewInput().Title("GitHub Token").EchoMode(huh.EchoModePassword).Value(&token).Run()
		if err != nil {
			return ""
		}
	}
	
	if token != "" {
		if err := manager.Set("github-token", token); err != nil {
			PrintError(fmt.Sprintf("Failed to save token: %v", err))
			return ""
		}
		fmt.Println("Success! GitHub token saved. This will automatically be used for GitHub APIs and the ghcr.io container registry.")
	}
	return token
}

// runInteractiveGitlabAuth prompts for a GitLab personal access token and
// saves it via the secrets manager.
func runInteractiveGitlabAuth() string {
	manager := getSecretsManager()
	var token string
	err := huh.NewInput().Title("GitLab Personal Access Token").EchoMode(huh.EchoModePassword).Value(&token).Run()
	if err != nil {
		return ""
	}
	if token != "" {
		if err := manager.Set("gitlab-token", token); err != nil {
			PrintError(fmt.Sprintf("Failed to save token: %v", err))
			return ""
		}
		fmt.Println("Success! GitLab token saved.")
	}
	return token
}

// runInteractiveCloudflareAuth prompts for a Cloudflare API token and
// account ID and saves whichever of the two the user entered.
func runInteractiveCloudflareAuth() (string, string) {
	manager := getSecretsManager()
	var apiToken, userID string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Cloudflare API Token").EchoMode(huh.EchoModePassword).Value(&apiToken),
			huh.NewInput().Title("Cloudflare Account ID").Value(&userID),
		),
	).Run()
	if err != nil {
		return "", ""
	}

	if apiToken != "" {
		if err := manager.Set("cf-token", apiToken); err != nil {
			PrintError(fmt.Sprintf("Failed to save Cloudflare API token: %v", err))
			return "", ""
		}
	}
	if userID != "" {
		if err := manager.Set("cf-user-id", userID); err != nil {
			PrintError(fmt.Sprintf("Failed to save Cloudflare user ID: %v", err))
			return "", ""
		}
	}
	if apiToken != "" || userID != "" {
		fmt.Println("Cloudflare credentials saved.")
	}
	return apiToken, userID
}

// runInteractiveOCIAuth prompts for a registry hostname (if not already
// known) plus a username/token pair, and saves them under
// "oci-<registry>-username" / "oci-<registry>-token".
func runInteractiveOCIAuth(registry string) (string, string) {
	manager := getSecretsManager()
	var user, pass string
	
	if registry == "" {
		err := huh.NewInput().Title("Registry URL (e.g. registry.gitlab.com)").Value(&registry).Run()
		if err != nil || registry == "" {
			return "", ""
		}
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Username for " + registry).Value(&user),
			huh.NewInput().Title("Token / Password").EchoMode(huh.EchoModePassword).Value(&pass),
		),
	).Run()
	if err != nil {
		return "", ""
	}

	if registry != "" {
		if err := manager.Set(fmt.Sprintf("oci-%s-username", registry), user); err != nil {
			PrintError(fmt.Sprintf("Failed to save OCI username: %v", err))
			return "", ""
		}
		if err := manager.Set(fmt.Sprintf("oci-%s-token", registry), pass); err != nil {
			PrintError(fmt.Sprintf("Failed to save OCI token: %v", err))
			return "", ""
		}
		fmt.Println("OCI credentials saved.")
	}
	return user, pass
}

// runInteractiveAuth is the entry point for `aeroflare auth login` when no
// token flags were passed: it asks which service to authenticate and
// dispatches to the matching runInteractive*Auth helper.
func runInteractiveAuth() {
	var service string
	err := huh.NewSelect[string]().
		Title("What do you want to authenticate?").
		Options(
			huh.NewOption("GitHub", "github"),
			huh.NewOption("GitLab", "gitlab"),
			huh.NewOption("Cloudflare", "cloudflare"),
			huh.NewOption("Custom OCI Registry", "oci"),
		).
		Value(&service).
		Run()
		
	if err != nil {
		PrintError("Cancelled")
		return
	}

	switch service {
	case "github":
		runInteractiveGithubAuth()
	case "gitlab":
		runInteractiveGitlabAuth()
	case "cloudflare":
		runInteractiveCloudflareAuth()
	case "oci":
		runInteractiveOCIAuth("")
	}
}
