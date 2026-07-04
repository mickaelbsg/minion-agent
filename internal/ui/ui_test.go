package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"minion/internal/admin"
	"minion/internal/config"
)

func TestEnsureInteractiveRejectsPipe(t *testing.T) {
	inReader, _, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	defer inReader.Close()

	_, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	defer outWriter.Close()

	err = ensureInteractive(inReader, outWriter)
	if err == nil {
		t.Fatal("expected TTY validation to fail")
	}
}

func TestNewModelSetupSectionRequiresRoot(t *testing.T) {
	service := admin.NewService(filepath.Join(t.TempDir(), "config.json"))
	service.IsRoot = func() bool { return false }

	model := NewModel(service, "setup")
	if model.current != messageScreen {
		t.Fatalf("expected message screen, got %v", model.current)
	}
	if !strings.Contains(model.message, "requer root") {
		t.Fatalf("unexpected message: %q", model.message)
	}
}

func TestConfigCancelDoesNotApplyChanges(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	service := admin.NewService(configPath)
	service.IsRoot = func() bool { return true }

	model := NewModel(service, "config")
	model.inputs[0].SetValue("127.0.0.1:9999")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.current != configConfirmScreen {
		t.Fatalf("expected config confirm screen, got %v", model.current)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	current, err := config.Read(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if current.API.Bind != cfg.API.Bind {
		t.Fatalf("expected bind to remain %q, got %q", cfg.API.Bind, current.API.Bind)
	}
}

func TestClientCreateFlowCreatesClient(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	cfg := config.Default()
	cfg.DBPath = filepath.Join(tempDir, "minion.db")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	service := admin.NewService(configPath)
	service.IsRoot = func() bool { return true }

	model := NewModel(service, "clients")
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(Model)

	model.inputs[0].SetValue("api")
	model.inputs[1].SetValue("127.0.0.1/32")

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.current != messageScreen || !model.messageOK {
		t.Fatalf("expected success message screen, got screen=%v ok=%v message=%q", model.current, model.messageOK, model.message)
	}
	if !strings.Contains(model.message, "Cliente criado") {
		t.Fatalf("unexpected message: %q", model.message)
	}
}
