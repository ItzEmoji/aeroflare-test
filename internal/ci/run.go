package ci

import (
	"fmt"
	"io"

	"github.com/itzemoji/aeroflare/internal/push"
)

// summaryLine renders the final one-line roll-up.
func summaryLine(buildsTotal, buildsOK, pushesTotal, pushesOK, uploaded, skipped int) string {
	status := "OK"
	if buildsOK != buildsTotal || pushesOK != pushesTotal {
		status = "FAILED"
	}
	return fmt.Sprintf("summary: builds %d/%d  pushes %d/%d  uploaded %d  skipped-upstream %d  —  %s",
		buildsOK, buildsTotal, pushesOK, pushesTotal, uploaded, skipped, status)
}

// Run executes the full spec: build every installable, push each result to every
// cache, continue past failures, and return true iff everything succeeded.
func Run(spec RunSpec, w io.Writer) bool {
	fmt.Fprintf(w, "aeroflare-ci: %d builds, %d caches\n\n", len(spec.Builds), len(spec.Caches))

	keyPath, cleanup, err := ResolveSigningKey(spec.SigningKey)
	if err != nil {
		fmt.Fprintf(w, "✗ signing-key: %v\n", err)
		return false
	}
	defer cleanup()

	upstream := spec.UpstreamCache
	if upstream == "none" {
		upstream = ""
	}

	buildsTotal := len(spec.Builds)
	buildsOK := 0
	pushesTotal := 0
	pushesOK := 0
	totalUploaded := 0
	totalSkipped := 0

	for _, installable := range spec.Builds {
		paths, err := BuildInstallable(installable, 0)
		if err != nil {
			fmt.Fprintf(w, "✗ build  %s   %v\n", installable, err)
			continue
		}
		buildsOK++
		fmt.Fprintf(w, "✓ build  %s   (%d paths)\n", installable, len(paths))

		for _, cache := range spec.Caches {
			pushesTotal++
			token := ResolveToken(cache.Registry)
			if token == "" {
				fmt.Fprintf(w, "✗ push   %s → %s   auth: no token (set %s)\n",
					installable, cache.Raw, TokenEnvVar(cache.Registry))
				continue
			}

			cfg := &push.PushConfig{
				TargetPaths: paths,
				Compression: spec.Compression,
				CacheURL:    upstream,
				Workers:     spec.Workers,
				PrepareRefs: true,
				SigningKey:  keyPath,
				ForcePush:   false,
				Verbosity:   1,
			}
			plan, err := push.Preflight(cfg)
			if err != nil {
				fmt.Fprintf(w, "✗ push   %s → %s   %v\n", installable, cache.Raw, err)
				continue
			}

			fmt.Fprintf(w, "  push %s → %s\n", installable, cache.Raw)
			reporter := NewPlainReporter(w, "  ")
			target := push.Target{Registry: cache.Registry, Repository: cache.Repository, Token: token}
			res, err := push.RunPushTo(plan, target, reporter)
			if res != nil {
				totalUploaded += res.Uploaded
				totalSkipped += res.SkippedUpstream
			}
			if err != nil {
				fmt.Fprintf(w, "✗ push   %s → %s   %v\n", installable, cache.Raw, err)
				continue
			}
			pushesOK++
			fmt.Fprintf(w, "✓ push   %s → %s   (%d uploaded, %d skipped)\n",
				installable, cache.Raw, res.Uploaded, res.SkippedUpstream)
		}
	}

	fmt.Fprintf(w, "\n%s\n", summaryLine(buildsTotal, buildsOK, pushesTotal, pushesOK, totalUploaded, totalSkipped))
	return buildsOK == buildsTotal && pushesOK == pushesTotal
}
