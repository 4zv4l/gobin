package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"
)

func startTCPServer() net.Listener {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *address, *tcpPort))
	if err != nil {
		slog.Error("Failed to start TCP server", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info(fmt.Sprintf("TCP Server listening on %s:%d", *address, *tcpPort))
		for {
			conn, err := listener.Accept()
			if err != nil {
				// if tcpserver is closed
				if errors.Is(err, net.ErrClosed) {
					break
				}
				slog.Error("TCP accept error", "error", err)
				continue
			}
			go handleTCPClient(conn)
		}
	}()

	return listener
}

func handleTCPClient(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()
	slog.Info("New client", "address", clientAddr)

	conn.SetReadDeadline(time.Now().Add(time.Duration(*timeout) * time.Second))

	id, err := acquireIDAndSpace()
	if err != nil {
		slog.Warn("Rejected paste", "reason", err)
		fmt.Fprintf(conn, "Server error: %v\n", err)
		return
	}

	path := filepath.Join(*directory, id)
	file, err := os.Create(path)
	if err != nil {
		releaseID(id)
		slog.Error("Failed to create file", "error", err)
		return
	}

	// +1 to detect client sending too much data
	// we will return an error rather than pasting a truncated file
	limitReader := io.LimitReader(conn, *maxFileSize+1)
	written, err := io.Copy(file, limitReader)
	file.Close()

	if written > *maxFileSize {
		os.Remove(path)
		releaseID(id)
		slog.Warn("Client tried to send too much data", "address", clientAddr)
		conn.Write([]byte("Too much data, try smaller :)\n"))
		return
	}

	if written == 0 {
		os.Remove(path)
		releaseID(id)
		return
	}

	if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		slog.Debug("TCP stream error", "error", err)
	}

	commitPasteSize(written)
	slog.Info("Paste created", "client", clientAddr, "id", id)
	fmt.Fprintf(conn, "%s/%s\n", webURL, id)
}
