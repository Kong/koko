package config

import (
	"fmt"

	"github.com/kong/koko/internal/model"
	"golang.org/x/exp/slices"
)

const (
	// MaxChanges is the maximum number of changes tracked by ChangeTracker.
	MaxChanges = 128
	// MaxResourcesPerChange is the maximum number of resource that
	// ChangeTracker will record for a specific change.
	MaxResourcesPerChange = 128
)

// ChangeTracker tracks changes for a given configuration.
// Tracking is bounded i.e. it there is an upper limits to the amount of
// changes that are tracked. Tracking changes or resources beyond MaxChanges
// or MaxResourcesPerChange is a no-op.
// Duplicated changes are de-duplicated.
// ChangeTracker is not thread-safe.
type ChangeTracker struct {
	// changes is a map of changeID to resource.
	changes map[ChangeID][]ResourceInfo
}

// NewChangeTracker returns a new ChangeTracker.
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		changes: map[ChangeID][]ResourceInfo{},
	}
}

// ResourceInfo serves as a pointer to a resource of Type and ID within the
// entire configuration of Kong gateway.
type ResourceInfo struct {
	Type string
	ID   string
}

func (r ResourceInfo) Valid() error {
	if !model.ValidType(model.Type(r.Type)) {
		return fmt.Errorf("invalid type: '%v'", r.Type)
	}
	if r.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	return nil
}

// Track tracks a change with id.
// The change is not associated with any resource.
// If ChangeTracker's capacity has been reached, the call is a no-op.
func (c *ChangeTracker) Track(id ChangeID) error {
	if err := id.Valid(); err != nil {
		return err
	}
	if _, tracked := c.changes[id]; tracked {
		return nil
	}
	if len(c.changes) == MaxChanges {
		// past limit
		return nil
	}
	c.changes[id] = []ResourceInfo{}
	return nil
}

// TrackForResource tracks a change with id for resource referenced by r.
// If ChangeTracker's capacity has been reached, the call is a no-op.
func (c *ChangeTracker) TrackForResource(id ChangeID, r ResourceInfo) error {
	if err := id.Valid(); err != nil {
		return err
	}
	if err := r.Valid(); err != nil {
		return fmt.Errorf("invalid resource: %w", err)
	}

	trackedResources, tracked := c.changes[id]
	if !tracked && len(c.changes) == MaxChanges {
		// honor max changes
		return nil
	}
	if len(trackedResources) == MaxResourcesPerChange {
		// honor max resources per change
		return nil
	}

	for _, trackedResource := range trackedResources {
		if r.Type == trackedResource.Type && r.ID == trackedResource.ID {
			return nil
		}
	}

	c.changes[id] = append(trackedResources, r)
	return nil
}

// ChangeDetail is a change ID that applies to multiple resources.
type ChangeDetail struct {
	ID        ChangeID
	Resources []ResourceInfo
}

// TrackedChanges contains all changes recorded by ChangeTracker.
type TrackedChanges struct {
	// Changes are resource-specific changes.
	ChangeDetails []ChangeDetail
}

// Get returns all the changes tracked up until this point.
// The response is sorted alphabetically and deterministic.
func (c *ChangeTracker) Get() TrackedChanges {
	if len(c.changes) == 0 {
		return TrackedChanges{}
	}
	sortedChanges := make([]ChangeDetail, 0, len(c.changes))
	for changeID, resources := range c.changes {
		slices.SortStableFunc(resources, func(a, b ResourceInfo) bool {
			return a.Type < b.Type || a.ID < b.ID
		})
		sortedChanges = append(sortedChanges, ChangeDetail{
			ID:        changeID,
			Resources: resources,
		})
	}
	slices.SortStableFunc(sortedChanges, func(a, b ChangeDetail) bool {
		return a.ID < b.ID
	})

	return TrackedChanges{
		ChangeDetails: sortedChanges,
	}
}
