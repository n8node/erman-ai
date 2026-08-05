package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SpeechKitTranscriptionAsyncMaxBytes int64   = 50 << 20
	SpeechKitTranscriptionSyncMaxBytes  int64   = 900 << 10
	SpeechKitTranscriptionSyncMaxSec    float64 = 25
	SpeechKitTranscriptionSegmentSec    float64 = 600
)

type speechKitTranscriptionSettings struct {
	Model                      string
	LanguageCode               string
	TextNormalizationEnabled   bool
	LiteratureText             bool
	ProfanityFilter            bool
}

type speechKitTranscriptionResult struct {
	Text        string
	DurationSec *float64
	ChunkCount  int
}

func transcribeSpeechKitAsset(
	ctx context.Context,
	llm *LLMService,
	assetPath string,
	workDir string,
	settings speechKitTranscriptionSettings,
	params YandexSpeechKitParams,
) (speechKitTranscriptionResult, error) {
	normalizedMedia, err := normalizeAudioForSpeechKit(ctx, assetPath, workDir)
	if err != nil {
		return speechKitTranscriptionResult{}, fmt.Errorf("audio normalize: %w", err)
	}

	duration, err := ffprobeDuration(ctx, normalizedMedia.path)
	if err != nil {
		duration = nil
	}

	normalizedData, err := os.ReadFile(normalizedMedia.path)
	if err != nil {
		return speechKitTranscriptionResult{}, err
	}

	chunks, err := chunkSpeechKitAudio(ctx, normalizedMedia, normalizedData, workDir)
	if err != nil {
		return speechKitTranscriptionResult{}, err
	}

	var transcriptParts []string
	for i, chunk := range chunks {
		text, err := transcribeSpeechKitChunk(ctx, llm, params, chunk, normalizedMedia.contentType, normalizedMedia.syncFormat)
		if err != nil {
			return speechKitTranscriptionResult{}, fmt.Errorf("chunk %d: %w", i+1, err)
		}
		transcriptParts = append(transcriptParts, strings.TrimSpace(text))
	}

	fullText := strings.TrimSpace(strings.Join(transcriptParts, "\n\n"))
	if fullText == "" {
		return speechKitTranscriptionResult{}, ErrYandexSpeechKitEmptyResult
	}

	return speechKitTranscriptionResult{
		Text:        fullText,
		DurationSec: duration,
		ChunkCount:  len(chunks),
	}, nil
}

func chunkSpeechKitAudio(
	ctx context.Context,
	normalizedMedia normalizedAudioMedia,
	normalizedData []byte,
	workDir string,
) ([][]byte, error) {
	if int64(len(normalizedData)) <= SpeechKitTranscriptionAsyncMaxBytes {
		return [][]byte{normalizedData}, nil
	}

	segmentDir := filepath.Join(workDir, "segments")
	if err := ffmpegSegmentAudio(ctx, normalizedMedia.path, segmentDir, SpeechKitTranscriptionSegmentSec, normalizedMedia.segmentCopy); err != nil {
		return nil, fmt.Errorf("audio split: %w", err)
	}
	entries, err := os.ReadDir(segmentDir)
	if err != nil {
		return nil, err
	}

	var chunks [][]byte
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), normalizedMedia.segmentExt) {
			continue
		}
		part, err := os.ReadFile(filepath.Join(segmentDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if int64(len(part)) > SpeechKitTranscriptionAsyncMaxBytes {
			subParts, err := splitSpeechKitSyncChunks(ctx, part, workDir, normalizedMedia)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, subParts...)
			continue
		}
		chunks = append(chunks, part)
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no audio segments produced")
	}
	return chunks, nil
}

func transcribeSpeechKitChunk(ctx context.Context, llm *LLMService, params YandexSpeechKitParams, chunk []byte, contentType, syncFormat string) (string, error) {
	if syncFormat == "" {
		syncFormat = "oggopus"
	}
	if contentType == "" {
		contentType = "audio/ogg"
	}
	if !params.PreferAsyncRecognition() && int64(len(chunk)) <= SpeechKitTranscriptionSyncMaxBytes {
		text, syncErr := llm.TranscribeYandexSpeechKitSync(ctx, params, chunk, syncFormat)
		if syncErr == nil {
			return text, nil
		}
	}
	return llm.TranscribeYandexSpeechKitAsync(ctx, params, chunk, contentType)
}

func splitSpeechKitSyncChunks(ctx context.Context, part []byte, workDir string, media normalizedAudioMedia) ([][]byte, error) {
	partPath := filepath.Join(workDir, "oversized"+media.segmentExt)
	if err := os.WriteFile(partPath, part, 0o640); err != nil {
		return nil, err
	}
	syncDir := filepath.Join(workDir, "sync_segments")
	if err := ffmpegSegmentAudio(ctx, partPath, syncDir, SpeechKitTranscriptionSyncMaxSec, media.segmentCopy); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), media.segmentExt) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(syncDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return out, nil
}

func ffmpegExtractAudioFromVideo(ctx context.Context, input, output string) error {
	hasAudio, err := ffprobeHasAudioStream(ctx, input)
	if err != nil {
		return err
	}
	if !hasAudio {
		return fmt.Errorf("video file has no audio track")
	}
	return ffmpegNormalizeAudioWithArgs(ctx, input, output,
		"-map", "0:a:0",
		"-ac", "1", "-ar", "16000",
		"-c:a", "pcm_s16le",
		"-f", "wav",
	)
}

func ffprobeHasAudioStream(ctx context.Context, input string) (bool, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_type",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "audio", nil
}

func ensureReadableMediaAsset(path, missingMsg string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s", missingMsg)
		}
		return fmt.Errorf("uploaded file unavailable: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("uploaded file is empty")
	}
	return nil
}

func ensureReadableAudioAsset(path string) error {
	return ensureReadableMediaAsset(path, "uploaded audio file is missing on server; upload the file again")
}

func ensureReadableVideoAsset(path string) error {
	return ensureReadableMediaAsset(path, "uploaded video file is missing on server; upload the file again")
}

func ffmpegNormalizeAudio(ctx context.Context, input, output string) error {
	return ffmpegNormalizeAudioWithArgs(ctx, input, output,
		"-map", "0:a:0",
		"-ac", "1", "-ar", "16000",
		"-c:a", "libopus", "-application", "voip", "-b:a", "32k",
		"-f", "ogg",
	)
}

func ffmpegNormalizeAudioWAV(ctx context.Context, input, output string) error {
	return ffmpegNormalizeAudioWithArgs(ctx, input, output,
		"-map", "0:a:0",
		"-ac", "1", "-ar", "16000",
		"-c:a", "pcm_s16le",
		"-f", "wav",
	)
}

type normalizedAudioMedia struct {
	path        string
	contentType string
	syncFormat  string
	segmentExt  string
	segmentCopy bool
}

func normalizeAudioForSpeechKit(ctx context.Context, input, workDir string) (normalizedAudioMedia, error) {
	oggPath := filepath.Join(workDir, "normalized.ogg")
	oggErr := ffmpegNormalizeAudio(ctx, input, oggPath)
	if oggErr == nil {
		return normalizedAudioMedia{
			path:        oggPath,
			contentType: "audio/ogg",
			syncFormat:  "oggopus",
			segmentExt:  ".ogg",
			segmentCopy: true,
		}, nil
	}

	wavPath := filepath.Join(workDir, "normalized.wav")
	if wavErr := ffmpegNormalizeAudioWAV(ctx, input, wavPath); wavErr == nil {
		return normalizedAudioMedia{
			path:        wavPath,
			contentType: "audio/wav",
			syncFormat:  "lpcm",
			segmentExt:  ".wav",
			segmentCopy: false,
		}, nil
	} else if wavErr != nil {
		return normalizedAudioMedia{}, fmt.Errorf("ogg: %v; wav: %w", oggErr, wavErr)
	}
	return normalizedAudioMedia{}, fmt.Errorf("ogg: %w", oggErr)
}

func ffmpegNormalizeAudioWithArgs(ctx context.Context, input, output string, encodeArgs ...string) error {
	info, err := os.Stat(input)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input media file is missing")
		}
		return err
	}
	if info.Size() == 0 {
		return fmt.Errorf("input media file is empty")
	}
	args := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error",
		"-probesize", "50M", "-analyzeduration", "50M",
		"-y", "-i", input,
		"-vn", "-sn", "-dn",
	}
	args = append(args, encodeArgs...)
	args = append(args, output)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ffmpegCommandError(err, out)
	}
	stat, statErr := os.Stat(output)
	if statErr != nil || stat.Size() == 0 {
		return fmt.Errorf("ffmpeg produced an empty output file")
	}
	return nil
}

func ffmpegSegmentAudio(ctx context.Context, input, outputPattern string, segmentSec float64, useCopy bool) error {
	pattern := filepath.Join(outputPattern, "chunk_%03d.ogg")
	if err := os.MkdirAll(outputPattern, 0o750); err != nil {
		return err
	}
	args := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error",
		"-y", "-i", input,
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%.0f", segmentSec),
	}
	if useCopy {
		args = append(args, "-c", "copy")
	} else {
		pattern = filepath.Join(outputPattern, "chunk_%03d.wav")
		args = append(args, "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le")
	}
	args = append(args, pattern)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ffmpegCommandError(err, out)
	}
	return nil
}

func ffmpegCommandError(err error, output []byte) error {
	if err == nil {
		return nil
	}
	msg := extractFFmpegError(output)
	if msg == "" {
		msg = strings.TrimSpace(string(output))
	}
	if len(msg) > 500 {
		msg = msg[len(msg)-500:]
	}
	if msg == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, msg)
}

func extractFFmpegError(output []byte) string {
	var parts []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "invalid") ||
			strings.Contains(lower, "no such file") ||
			strings.Contains(lower, "does not contain") ||
			strings.Contains(lower, "could not") {
			parts = append(parts, line)
		}
	}
	return strings.Join(parts, "; ")
}

func ffprobeDuration(ctx context.Context, input string) (*float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || val <= 0 {
		return nil, fmt.Errorf("invalid duration")
	}
	return &val, nil
}

func randomMediaAssetName(ext string) string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	if ext == "" {
		ext = ".bin"
	}
	return hex.EncodeToString(buf) + ext
}
