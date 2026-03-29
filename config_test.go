package outrunner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// TODO: test valid config parsing
	// TODO: test missing label error
	// TODO: test missing provider error
	// TODO: test default values for libvirt/tart
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `images:
  - label: linux
    docker:
      image: runner:latest
  - label: windows
    libvirt:
      path: /tmp/win.qcow2
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if len(cfg.Images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(cfg.Images))
	}
	if cfg.Images[0].Label != "linux" {
		t.Errorf("expected label linux, got %s", cfg.Images[0].Label)
	}
	if cfg.Images[0].Docker == nil {
		t.Error("expected docker config")
	}
	if cfg.Images[1].Libvirt == nil {
		t.Error("expected libvirt config")
	}
	// Check defaults applied
	if cfg.Images[1].Libvirt.CPUs != 4 {
		t.Errorf("expected default CPUs 4, got %d", cfg.Images[1].Libvirt.CPUs)
	}
	if cfg.Images[1].Libvirt.MemoryMB != 8192 {
		t.Errorf("expected default memory 8192, got %d", cfg.Images[1].Libvirt.MemoryMB)
	}
}

func TestMatchImage(t *testing.T) {
	cfg := &Config{
		Images: []ImageConfig{
			{Label: "linux", Docker: &DockerImage{Image: "runner:latest"}},
			{Label: "windows", Libvirt: &LibvirtImage{Path: "/tmp/win.qcow2"}},
		},
	}

	// Match by label
	img, err := cfg.MatchImage([]string{"windows"})
	if err != nil {
		t.Fatalf("MatchImage: %v", err)
	}
	if img.Label != "windows" {
		t.Errorf("expected windows, got %s", img.Label)
	}

	// Empty labels → first image
	img, err = cfg.MatchImage(nil)
	if err != nil {
		t.Fatalf("MatchImage(nil): %v", err)
	}
	if img.Label != "linux" {
		t.Errorf("expected linux fallback, got %s", img.Label)
	}

	// No match
	_, err = cfg.MatchImage([]string{"nonexistent"})
	if err == nil {
		t.Error("expected error for no match")
	}
}

func TestAllLabels(t *testing.T) {
	cfg := &Config{
		Images: []ImageConfig{
			{Label: "linux"},
			{Label: "windows"},
			{Label: "macos"},
		},
	}

	labels := cfg.AllLabels()
	if len(labels) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(labels))
	}
}

func TestNeedsBackend(t *testing.T) {
	cfg := &Config{
		Images: []ImageConfig{
			{Label: "linux", Docker: &DockerImage{Image: "runner:latest"}},
		},
	}

	if !cfg.NeedsDocker() {
		t.Error("expected NeedsDocker=true")
	}
	if cfg.NeedsLibvirt() {
		t.Error("expected NeedsLibvirt=false")
	}
	if cfg.NeedsTart() {
		t.Error("expected NeedsTart=false")
	}
}
