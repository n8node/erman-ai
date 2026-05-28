package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const startImageBasename = "start_image"

var allowedStartImageMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type TelegramAssets struct {
	dir string
}

func NewTelegramAssets(dir string) *TelegramAssets {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = "/app/data/telegram"
	}
	return &TelegramAssets{dir: dir}
}

func (a *TelegramAssets) EnsureDir() error {
	return os.MkdirAll(a.dir, 0o750)
}

func (a *TelegramAssets) SaveStartImage(contentType string, r io.Reader, maxBytes int64) (filename string, err error) {
	ext, ok := allowedStartImageMIME[strings.ToLower(strings.TrimSpace(contentType))]
	if !ok {
		return "", fmt.Errorf("unsupported image type %q (use JPEG, PNG or WebP)", contentType)
	}
	if err := a.EnsureDir(); err != nil {
		return "", err
	}
	_ = a.DeleteStartImage()

	filename = startImageBasename + ext
	path := filepath.Join(a.dir, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return "", err
	}
	defer f.Close()

	limited := io.LimitReader(r, maxBytes+1)
	n, err := io.Copy(f, limited)
	if err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if n > maxBytes {
		_ = os.Remove(path)
		return "", fmt.Errorf("image exceeds %d MB limit", maxBytes/(1024*1024))
	}
	if n == 0 {
		_ = os.Remove(path)
		return "", fmt.Errorf("empty image file")
	}
	return filename, nil
}

func (a *TelegramAssets) StartImagePath(filename string) string {
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "" {
		return ""
	}
	return filepath.Join(a.dir, filename)
}

func (a *TelegramAssets) StartImageExists(filename string) bool {
	path := a.StartImagePath(filename)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (a *TelegramAssets) DeleteStartImage() error {
	if err := a.EnsureDir(); err != nil {
		return err
	}
	entries, err := os.ReadDir(a.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), startImageBasename) {
			_ = os.Remove(filepath.Join(a.dir, e.Name()))
		}
	}
	return nil
}

func (a *TelegramAssets) OpenStartImage(filename string) (*os.File, error) {
	path := a.StartImagePath(filename)
	if path == "" {
		return nil, os.ErrNotExist
	}
	return os.Open(path)
}
