package payloads

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
)

const maxPayloadSize = 100 * 1024 * 1024

// Request configures payload generation.
type Request struct {
	Type     string
	Size     int
	Pattern  string
	Filename string
}

// Generate writes a payload file via the files manager.
func Generate(fm *files.Manager, req Request) (map[string]any, error) {
	if fm == nil {
		return nil, fmt.Errorf("files manager not configured")
	}
	if req.Size <= 0 {
		req.Size = 1024
	}
	if req.Size > maxPayloadSize {
		return nil, fmt.Errorf("payload size too large (max 100MB)")
	}
	if req.Filename == "" {
		req.Filename = fmt.Sprintf("payload_%d.txt", time.Now().Unix())
	}
	if req.Pattern == "" {
		req.Pattern = "A"
	}
	content, err := buildContent(req)
	if err != nil {
		return nil, err
	}
	res, err := fm.Create(req.Filename, content, false)
	if err != nil {
		return nil, err
	}
	res["payload_info"] = map[string]any{
		"type":    req.Type,
		"size":    req.Size,
		"pattern": req.Pattern,
	}
	return res, nil
}

func buildContent(req Request) (string, error) {
	switch req.Type {
	case "", "buffer":
		if len(req.Pattern) == 0 {
			return "", fmt.Errorf("pattern required")
		}
		repeat := req.Size / len(req.Pattern)
		if repeat < 1 {
			repeat = 1
		}
		return strings.Repeat(req.Pattern, repeat)[:req.Size], nil
	case "cyclic":
		alpha := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		var b strings.Builder
		for i := 0; i < req.Size; i++ {
			b.WriteByte(alpha[i%len(alpha)])
		}
		return b.String(), nil
	case "random":
		const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		var b strings.Builder
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < req.Size; i++ {
			b.WriteByte(chars[rng.Intn(len(chars))])
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("invalid payload type %q", req.Type)
	}
}
