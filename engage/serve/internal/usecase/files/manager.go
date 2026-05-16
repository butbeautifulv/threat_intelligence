package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manager stores engage artifacts under a chrooted root directory.
type Manager struct {
	Root string
}

func NewManager(root string) (*Manager, error) {
	if root == "" {
		root = "/var/veil/engage/files"
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	return &Manager{Root: root}, nil
}

func (m *Manager) resolve(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("filename is required")
	}
	clean := filepath.Clean(name)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path")
	}
	abs, err := filepath.Abs(filepath.Join(m.Root, clean))
	if err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(m.Root)
	if err != nil {
		return "", err
	}
	if abs != rootAbs && !strings.HasPrefix(abs, rootAbs+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes files root")
	}
	return abs, nil
}

func (m *Manager) CreateBytes(filename string, data []byte) (map[string]any, error) {
	path, err := m.resolve(filename)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, err
	}
	return map[string]any{"path": path, "created": true, "size": len(data)}, nil
}

func (m *Manager) Create(filename, content string, binary bool) (map[string]any, error) {
	path, err := m.resolve(filename)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	_ = binary
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return nil, err
	}
	return map[string]any{"path": path, "created": true}, nil
}

func (m *Manager) Modify(filename, content string, append bool) (map[string]any, error) {
	path, err := m.resolve(filename)
	if err != nil {
		return nil, err
	}
	flag := os.O_CREATE | os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	f, err := os.OpenFile(path, flag, 0o600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return nil, err
	}
	return map[string]any{"path": path, "modified": true}, nil
}

func (m *Manager) Delete(filename string) (map[string]any, error) {
	path, err := m.resolve(filename)
	if err != nil {
		return nil, err
	}
	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}
	return map[string]any{"deleted": filename}, nil
}

func (m *Manager) List(directory string) (map[string]any, error) {
	dir := directory
	if dir == "" || dir == "." {
		dir = "."
	}
	path, err := m.resolve(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return map[string]any{"directory": directory, "files": names}, nil
}
