package main

import (
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func initStorageState() error {
	// Calculate existing dir size and collect used IDs
	entries, err := os.ReadDir(*directory)
	if err != nil {
		return err
	}
	usedIDs := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			currentDirSize += info.Size()
			usedIDs[entry.Name()] = true
		}
	}

	// Generate all possible ID combinations using backtracking
	var generate func(prefix string, depth int)
	generate = func(prefix string, depth int) {
		if depth == *randLen {
			if !usedIDs[prefix] {
				idPool = append(idPool, prefix)
			}
			return
		}
		for i := range len(charset) {
			generate(prefix+string(charset[i]), depth+1)
		}
	}

	slog.Info("Generating ID pool (this may take a moment if randlen is high)...")
	generate("", 0)

	// Shuffle the pool
	rand.Shuffle(len(idPool), func(i, j int) {
		idPool[i], idPool[j] = idPool[j], idPool[i]
	})
	return nil
}

// acquireIDAndSpace returns an id if maxDirSize isnt reached and if the idPool isnt empty
func acquireIDAndSpace() (string, error) {
	fsMutex.Lock()
	defer fsMutex.Unlock()

	if currentDirSize >= *maxDirSize {
		if !*gc {
			return "", errors.New("storage full")
		}
		for currentDirSize >= *maxDirSize {
			if err := freeOldestFile(); err != nil {
				return "", errors.New("failed to run GC for size")
			}
		}
	}

	if len(idPool) == 0 {
		if !*gc {
			return "", errors.New("ID pool empty")
		}
		if err := freeOldestFile(); err != nil {
			return "", errors.New("failed to run GC for IDs")
		}
	}

	id := idPool[len(idPool)-1]
	idPool = idPool[:len(idPool)-1]
	return id, nil
}

// releaseID add id back into the idPool
func releaseID(id string) {
	fsMutex.Lock()
	idPool = append(idPool, id)
	fsMutex.Unlock()
}

// commitPasteSize add file size to currentDirSize aftre a successful transaction
func commitPasteSize(size int64) {
	fsMutex.Lock()
	currentDirSize += size
	fsMutex.Unlock()
}

func freeOldestFile() error {
	entries, err := os.ReadDir(*directory)
	if err != nil || len(entries) == 0 {
		return errors.New("nothing to clean up")
	}

	var oldestName string
	var oldestTime time.Time
	var oldestSize int64

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if oldestName == "" || info.ModTime().Before(oldestTime) {
				oldestName = entry.Name()
				oldestTime = info.ModTime()
				oldestSize = info.Size()
			}
		}
	}

	if oldestName != "" {
		slog.Warn("GC triggered: deleting file", "file", oldestName)
		os.Remove(filepath.Join(*directory, oldestName))
		currentDirSize -= oldestSize
		idPool = append(idPool, oldestName)
		return nil
	}
	return errors.New("no valid files found to GC")
}
