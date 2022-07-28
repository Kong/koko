package config

import (
	"errors"
	"fmt"
	"regexp"
)

type ChangeSeverity string

const (
	// ChangeSeverityWarning is a configuration warning.
	// This is used to highlight the use of a deprecated configuration field.
	// This severity implies that there is no change in functionality of the
	// gateway.
	ChangeSeverityWarning ChangeSeverity = "warning"
	// ChangeSeverityError is a configuration error.
	// This is used to highlight a compatibility change that result in the
	// behavior of the gateway to be different from what the intention of the
	// user.
	ChangeSeverityError ChangeSeverity = "error"
)

var ErrRegistryEntryNotFound = errors.New("not found")

var changeIDRegex = regexp.MustCompile(`^[A-Z][A-Z\d]{3}$`)

// ChangeID is a globally unique identifier for each compatibility change
// definition.
type ChangeID string

func (c ChangeID) Valid() error {
	if !changeIDRegex.MatchString(string(c)) {
		return fmt.Errorf("invalid change ID: '%v'", c)
	}
	return nil
}

// ChangeMetadata holds metadata for a specific change.
type ChangeMetadata struct {
	// ID identifies a change uniquely. Required.
	ID ChangeID
	// Severity identifies the severity of a change. Required.
	// Read Severities defined in this package for more details.
	Severity ChangeSeverity
	// Description is a human-readable sentence describing the impact of the
	// change. Required.
	Description string
	// Resolution is a human-readable sentence describing the path to
	// resolution. Required.
	Resolution string
	// DocumentationURL is an optional web URL that further details the
	// change or its resolution.
	DocumentationURL string
}

// Change is a configuration change that is done automatically by the
// control-plane in order to achieve compatibility with a Kong data-plane.
type Change struct {
	// Metadata holds metadata associated with a change.
	Metadata ChangeMetadata
	// Version is the last version for which this change must be executed to
	// guarantee compatibility.
	Version uint64
	// Update holds a declarative definition of the change that must be
	// applied to a specific schema.
	Update ConfigTableUpdates
}

func (c *Change) valid() error {
	switch c.Metadata.Severity {
	case ChangeSeverityWarning:
	case ChangeSeverityError:
	default:
		return fmt.Errorf("invalid change severity: '%v'", c.Metadata.Severity)
	}

	if err := c.Metadata.ID.Valid(); err != nil {
		return err
	}

	if c.Metadata.Description == "" || c.Metadata.Resolution == "" {
		return fmt.Errorf("change has no description or resolution")
	}

	if c.Version == 0 {
		return fmt.Errorf("invalid version '%v'", c.Version)
	}
	return nil
}

// ChangeRegistry holds all changes.
var ChangeRegistry = newCompatChangeRegistry()

// CompatChangeRegistry holds compatibility changes.
// Implementations may not be thread safe.
type CompatChangeRegistry interface {
	// Register registers a change.
	// If a change is already registers or if a change is invalid,
	// it returns an error.
	Register(c Change) error
	// GetMetadata returns ChangeMetadata for an id.
	// It returns ErrRegistryEntryNotFound if a change with id has not been previously
	// registered.
	GetMetadata(id ChangeID) (ChangeMetadata, error)
	// GetPluginUpdates returns configuration updates for all registered
	// changes. The order of updates within a version is not deterministic.
	GetPluginUpdates() VersionedConfigUpdates
}

// compatChangeRegistryImpl implements the CompatChangeRegistry interface.
// It is not thread-safe.
type compatChangeRegistryImpl struct {
	changes map[ChangeID]Change
}

func newCompatChangeRegistry() CompatChangeRegistry {
	return &compatChangeRegistryImpl{
		changes: map[ChangeID]Change{},
	}
}

func (c *compatChangeRegistryImpl) Register(change Change) error {
	if err := change.valid(); err != nil {
		return fmt.Errorf("invalid change: %w", err)
	}
	id := change.Metadata.ID
	if _, ok := c.changes[id]; ok {
		return fmt.Errorf("change '%s' already registered", id)
	}
	c.changes[id] = change
	return nil
}

func (c *compatChangeRegistryImpl) GetMetadata(id ChangeID) (ChangeMetadata, error) {
	res, ok := c.changes[id]
	if !ok {
		return ChangeMetadata{}, ErrRegistryEntryNotFound
	}
	return res.Metadata, nil
}

func (c *compatChangeRegistryImpl) GetPluginUpdates() VersionedConfigUpdates {
	res := make(VersionedConfigUpdates, len(c.changes))
	for _, change := range c.changes {
		res[change.Version] = append(res[change.Version], change.Update)
	}
	return res
}
