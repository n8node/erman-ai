package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

var ErrMultipartFileMissing = errors.New("multipart file field missing")

type multipartFilePart struct {
	Reader      io.ReadCloser
	Filename    string
	ContentType string
}

func openStreamingMultipartFile(w http.ResponseWriter, r *http.Request, maxBodyBytes int64, fieldNames ...string) (*multipartFilePart, error) {
	if len(fieldNames) == 0 {
		fieldNames = []string{"file"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}
	want := make(map[string]struct{}, len(fieldNames))
	for _, name := range fieldNames {
		want[name] = struct{}{}
	}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, ok := want[part.FormName()]; !ok {
			_ = part.Close()
			continue
		}
		contentType := strings.TrimSpace(part.Header.Get("Content-Type"))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		filename := strings.TrimSpace(part.FileName())
		if filename == "" {
			filename = "upload"
		}
		return &multipartFilePart{
			Reader:      part,
			Filename:    filename,
			ContentType: contentType,
		}, nil
	}
	return nil, ErrMultipartFileMissing
}
