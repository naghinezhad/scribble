package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"strings"
	"time"
)

func (h *Handler) funcs() template.FuncMap {
	return template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
		"formatTime": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"hashed": h.getAssetHashedURL,
	}
}

// calculateAssetHash calculates a hash for the given asset file for cache busting.
func (h *Handler) calculateAssetHash(asset string) (string, error) {
	file, err := h.static.Open(strings.TrimLeft(asset, "/"))
	if err != nil {
		return "", fmt.Errorf("failed to open asset %s: %w", asset, err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			slog.Error("failed to close asset file", "asset", asset, "error", err)
		}
	}()

	hash := sha256.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("failed to hash %s: %w", asset, err)
	}

	// Use first 8 characters of hash for brevity
	return hex.EncodeToString(hash.Sum(nil))[:8], nil
}

func (h *Handler) getAssetHashedURL(asset string) string {
	hash, ok := h.assetHashes[asset]
	if !ok {
		var err error

		hash, err = h.calculateAssetHash(asset)
		if err != nil {
			slog.Error("failed to calculate asset hash", "asset", asset, "error", err)

			return asset
		}

		h.assetHashes[asset] = hash
	}

	return fmt.Sprintf("%s?v=%s", asset, hash)
}
