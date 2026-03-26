package outrunner

import (
	"context"
	"fmt"
	"log/slog"
)

// MultiProvisioner routes runner requests to the appropriate backend
// based on the matched image's provider type.
type MultiProvisioner struct {
	logger   *slog.Logger
	config   *Config
	backends map[string]Provisioner
}

func NewMultiProvisioner(logger *slog.Logger, config *Config) *MultiProvisioner {
	return &MultiProvisioner{
		logger:   logger,
		config:   config,
		backends: make(map[string]Provisioner),
	}
}

// Register adds a provisioner backend by name.
func (m *MultiProvisioner) Register(name string, prov Provisioner) {
	m.backends[name] = prov
}

func (m *MultiProvisioner) Start(ctx context.Context, req *RunnerRequest) error {
	img, err := m.config.MatchImage(req.Labels)
	if err != nil {
		return err
	}

	provType := img.ProviderType()
	prov, ok := m.backends[provType]
	if !ok {
		return fmt.Errorf("no %s provisioner registered", provType)
	}

	// Attach the matched image to the request so the backend can use it
	req.Image = img

	m.logger.Debug("Routing to provisioner",
		slog.String("type", provType),
		slog.String("name", req.Name),
		slog.Any("labels", req.Labels),
	)

	return prov.Start(ctx, req)
}

func (m *MultiProvisioner) Stop(ctx context.Context, name string) error {
	// Try all backends — only one will have this runner
	for _, prov := range m.backends {
		if err := prov.Stop(ctx, name); err == nil {
			return nil
		}
	}
	return nil
}

func (m *MultiProvisioner) Close() error {
	var firstErr error
	for name, prov := range m.backends {
		if err := prov.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close %s: %w", name, err)
		}
	}
	return firstErr
}
