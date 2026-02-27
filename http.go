package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

func routeDefault(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<h3>Send some text and read it back</h3>
		<code>
		$ echo just testing! | nc %s %d </br>
		%s/test </br>
		$ curl %s/test </br>
		just testing! </br>
		</code>`, webURL, *tcpPort, *baseURL, *baseURL)
}

func routeID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filePath := filepath.Join(*directory, id)

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		http.NotFound(w, r)
		return
	}

	slog.Info("Serving paste", "client", r.RemoteAddr, "id", id)
	w.Header().Set("Content-Type", "text/plain")
	http.ServeFile(w, r, filePath)
}

func startWebServer(isTLS bool, cerr chan error) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", routeDefault)
	router.HandleFunc("/{id}", routeID)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *address, *webPort),
		Handler: router,
	}

	go func() {
		slog.Info(fmt.Sprintf("Web server listening on %s:%d", *address, *webPort))
		var err error
		if isTLS {
			err = srv.ListenAndServeTLS(*certPath, *pkeyPath)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Web server failed", "error", err)
			cerr <- err
		}
	}()

	return srv
}
