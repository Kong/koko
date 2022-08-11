package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChange_valid(t *testing.T) {
	t.Run("valid change returns no error", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F001",
				Severity:    ChangeSeverityWarning,
				Description: "some description",
				Resolution:  "some resolution",
			},
			SemverRange: "< 2.8.0",
		}
		err := change.valid()
		require.NoError(t, err)
	})
	t.Run("invalid severity errors", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F001",
				Severity:    "borked",
				Description: "some description",
				Resolution:  "some resolution",
			},
			SemverRange: "< 2.8.0",
		}
		err := change.valid()
		require.ErrorContains(t, err, "invalid change severity")
	})
	t.Run("invalid id errors", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F1234",
				Severity:    ChangeSeverityWarning,
				Description: "some description",
				Resolution:  "some resolution",
			},
			SemverRange: "< 2.8.0",
		}
		err := change.valid()
		require.ErrorContains(t, err, "invalid change ID")
	})
	t.Run("change without description errors", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F123",
				Severity:    ChangeSeverityWarning,
				Description: "",
				Resolution:  "some resolution",
			},
			SemverRange: "< 2.8.0",
		}
		err := change.valid()
		require.ErrorContains(t, err, "no description or resolution")
	})
	t.Run("change without resolution errors", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F123",
				Severity:    ChangeSeverityWarning,
				Description: "some description",
				Resolution:  "",
			},
			SemverRange: "< 2.8.0",
		}
		err := change.valid()
		require.ErrorContains(t, err, "no description or resolution")
	})
	t.Run("change without version errors", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F123",
				Severity:    ChangeSeverityWarning,
				Description: "some description",
				Resolution:  "some resolution",
			},
		}
		err := change.valid()
		require.ErrorContains(t, err, "invalid version")
	})
	t.Run("change with invalid version range errors", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F123",
				Severity:    ChangeSeverityWarning,
				Description: "some description",
				Resolution:  "some resolution",
			},
			SemverRange: "< a.b.c",
		}
		err := change.valid()
		require.ErrorContains(t, err, "invalid range")
	})
	t.Run("change with enterprise version returns no error", func(t *testing.T) {
		change := Change{
			Metadata: ChangeMetadata{
				ID:          "F123",
				Severity:    ChangeSeverityWarning,
				Description: "some description",
				Resolution:  "some resolution",
			},
			SemverRange: "< 3.0.0.0",
		}
		err := change.valid()
		require.NoError(t, err)
	})
}

func TestCompatChangeRegistryImpl_Register(t *testing.T) {
	registry := newCompatChangeRegistry()

	t.Run("invalid change errors", func(t *testing.T) {
		err := registry.Register(Change{
			Metadata:    ChangeMetadata{},
			SemverRange: "< 2.8.0",
		})
		require.ErrorContains(t, err, "invalid change severity")
	})
	t.Run("valid change doesn't error", func(t *testing.T) {
		err := registry.Register(Change{
			Metadata: ChangeMetadata{
				ID:          "T042",
				Severity:    ChangeSeverityWarning,
				Description: "42 is not the answer",
				Resolution:  "Make it so",
			},
			SemverRange: "< 2.8.0",
		})
		require.NoError(t, err)
	})
	t.Run("registering with same ID again errors", func(t *testing.T) {
		err := registry.Register(Change{
			Metadata: ChangeMetadata{
				ID:          "T042",
				Severity:    ChangeSeverityError,
				Description: "compile time errors",
				Resolution:  "yay",
			},
			SemverRange: "< 2.8.0",
		})
		require.ErrorContains(t, err, "already registered")
	})
}

func TestCompatChangeRegistryImpl_GetMetadata(t *testing.T) {
	registry := newCompatChangeRegistry()
	require.NoError(t, registry.Register(Change{
		Metadata: ChangeMetadata{
			ID:          "T042",
			Severity:    ChangeSeverityWarning,
			Description: "42 is not the answer",
			Resolution:  "Make it so",
		},
		SemverRange: "< 2.8.0",
	}))

	t.Run("get for an existing ID doesn't return an error", func(t *testing.T) {
		metadata, err := registry.GetMetadata("T042")
		require.NoError(t, err)
		require.Equal(t, "42 is not the answer", metadata.Description)
	})
	t.Run("get for an non-existing returns an error", func(t *testing.T) {
		metadata, err := registry.GetMetadata("T044")
		require.ErrorContains(t, err, "not found")
		require.Empty(t, metadata)
		require.Equal(t, ErrRegistryEntryNotFound, err)
	})
}

func TestCompatChangeRegistryImpl_GetPluginUpdates(t *testing.T) {
	registry := newCompatChangeRegistry()
	require.NoError(t, registry.Register(Change{
		Metadata: ChangeMetadata{
			ID:          "T042",
			Severity:    ChangeSeverityWarning,
			Description: "42 is not the answer",
			Resolution:  "Make it so",
		},
		SemverRange: "< 2.7.0",
		Update: ConfigTableUpdates{
			Name:   "opentelemetry",
			Type:   Plugin,
			Remove: true,
		},
	}))
	require.NoError(t, registry.Register(Change{
		Metadata: ChangeMetadata{
			ID:          "T043",
			Severity:    ChangeSeverityWarning,
			Description: "New version introduced a field.",
			Resolution:  "Please upgrade.",
		},
		SemverRange: "< 2.7.0",
		Update: ConfigTableUpdates{
			Name: "response-ratelimiting",
			Type: Plugin,
			RemoveFields: []string{
				"redis_username",
			},
		},
	}))
	require.NoError(t, registry.Register(Change{
		Metadata: ChangeMetadata{
			ID:          "T044",
			Severity:    ChangeSeverityWarning,
			Description: "New version introduced a field.",
			Resolution:  "Please upgrade.",
		},
		SemverRange: "< 2.8.0",
		Update: ConfigTableUpdates{
			Name: "ip-restriction",
			Type: Plugin,
			RemoveFields: []string{
				"status",
				"message",
			},
		},
	}))

	updates := registry.GetUpdates()

	require.Equal(t, 2, len(updates))

	require.Equal(t, 2, len(updates["< 2.7.0"]))
	require.Equal(t, 1, len(updates["< 2.8.0"]))
}
