package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"aeroflare/internal/oci"
	"aeroflare/internal/proxy"

	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the cache proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		registry, repository := oci.GetRegistryAndRepository()

		// Settings below are read from NIXCACHE_* env vars rather than flags so
		// the proxy can be configured the same way whether it's run directly
		// or deployed as a systemd service / container.
		port := 37515
		if pStr := os.Getenv("NIXCACHE_PORT"); pStr != "" {
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		listenAddr := os.Getenv("NIXCACHE_LISTEN")
		if listenAddr == "" {
			listenAddr = "127.0.0.1"
		}

		var upstreams []string
		if ups := os.Getenv("NIXCACHE_UPSTREAM"); ups != "" {
			upstreams = strings.Fields(ups)
		} else {
			upstreams = []string{"https://cache.nixos.org"}
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cancel the proxy's context on SIGINT/SIGTERM so StartProxy can shut
		// down its listener cleanly instead of the process being killed mid-request.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cancel()
		}()

		token := getOptionalTokenForRegistry(registry)
		actualPort, err := proxy.StartProxy(ctx, port, listenAddr, registry, repository, upstreams, token)
		if err != nil {
			PrintError(fmt.Sprintf("Proxy server failed: %v", err))
			os.Exit(1)
		}
		PrintInfo(fmt.Sprintf("Started proxy on %s:%d...", listenAddr, actualPort))

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
}
