package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	address     = flag.String("address", "127.0.0.1", "bind to this address")
	baseURL     = flag.String("base-url", "http://127.0.0.1", "base url to show in the template")
	tcpPort     = flag.Int("tcp-port", 9999, "bind to this port (tcp server)")
	webPort     = flag.Int("web-port", 4433, "bind to this port (web server)")
	directory   = flag.String("directory", os.TempDir(), "directory to save/serve the pastes")
	maxDirSize  = flag.Int64("max-dir-size", 104857600, "max directory size allowed in byte")
	maxFileSize = flag.Int64("max-file-size", 10485760, "max file size allowed in byte")
	timeout     = flag.Int("timeout", 1, "timeout in second to receive a paste")
	pkeyPath    = flag.String("pkey-path", "", "private key path for tls")
	certPath    = flag.String("cert-path", "", "certificate path for tls")
	loglevel    = flag.String("loglevel", "INFO", "log message up to that level")
	randLen     = flag.Int("randlen", 4, "IDs length")
	gc          = flag.Bool("gc", false, "delete old paste if the pool is full")

	fsMutex        sync.Mutex
	idPool         []string
	currentDirSize int64
	tmplTcpAddr        string // used to show the ip/domain in the template
	tmplTcpPort        string // used to show the port in the template
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func main() {
	flag.Parse()

	if *directory == "" {
		fmt.Println("Error: --directory is mandatory.")
		flag.Usage()
		os.Exit(1)
	}

	if err := setupLogger(); err != nil {
		slog.Error("Failed to setup logger", "error", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*directory, 0755); err != nil {
		slog.Error("Failed to create directory", "error", err)
		os.Exit(1)
	}

	if err := initStorageState(); err != nil {
		slog.Error("Failed to init the storage state", "error", err)
		os.Exit(1)
	}

	webURL := strings.SplitN(*baseURL, "/", 3)[2]
	tmplTcpAddr = "tcp." + webURL
	if proxy_port := os.Getenv("RAILWAY_TCP_PROXY_PORT"); proxy_port != "" {
		tmplTcpPort = proxy_port
	}

	isTLS := *pkeyPath != "" && *certPath != ""
	showPort := !((*webPort == 80 && !isTLS) || (*webPort == 443 && isTLS))
	if showPort {
		*baseURL += fmt.Sprintf(":%d", *webPort)
	}

	slog.Info("Starting service...")
	slog.Debug("Config", "tls", isTLS, "web_url", baseURL, "gc", *gc, "pool_size", len(idPool))

	httpErr := make(chan error)
	srv := startWebServer(isTLS, httpErr)
	listener := startTCPServer()

	// handle ctrl-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		break
	case <-httpErr:
		os.Exit(1)
	}

	fmt.Println("\rBye !")
	srv.Shutdown(context.Background())
	listener.Close()
}

// setupLogger logs to stdout whats happening depending on the level give by cli arguments
func setupLogger() error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(*loglevel)); err != nil {
		return err
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	return nil
}
