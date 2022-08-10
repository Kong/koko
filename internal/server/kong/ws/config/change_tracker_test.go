package config

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestChangeTracker_Track(t *testing.T) {
	t.Run("tracks a valid change", func(t *testing.T) {
		tracker := NewChangeTracker()
		err := tracker.Track("T420")
		require.NoError(t, err)
	})
	t.Run("tracking the same change is a no-op", func(t *testing.T) {
		tracker := NewChangeTracker()
		err := tracker.Track("T420")
		require.NoError(t, err)
		err = tracker.Track("T420")
		require.NoError(t, err)

		require.Len(t, tracker.changes, 1)
	})
	t.Run("tracking an invalid change returns an error", func(t *testing.T) {
		tracker := NewChangeTracker()
		err := tracker.Track("T42")
		require.ErrorContains(t, err, "invalid change ID")
	})
	t.Run("tracking more than MaxChanges is a no-op", func(t *testing.T) {
		tracker := NewChangeTracker()
		for i := 0; i < MaxChanges+1; i++ {
			err := tracker.Track(ChangeID(fmt.Sprintf("T%03d", i)))
			require.NoError(t, err)
		}
		require.Len(t, tracker.changes, MaxChanges)
	})
}

func TestChangeTracker_TrackForResource(t *testing.T) {
	t.Run("tracks a valid change", func(t *testing.T) {
		tracker := NewChangeTracker()
		err := tracker.TrackForResource("T420", ResourceInfo{
			Type: "route",
			ID:   uuid.NewString(),
		})
		require.NoError(t, err)
	})
	t.Run("tracking the same resource with same change is a no-op", func(t *testing.T) {
		tracker := NewChangeTracker()
		rid := uuid.NewString()

		err := tracker.TrackForResource("T420", ResourceInfo{
			Type: "route",
			ID:   rid,
		})
		require.NoError(t, err)
		err = tracker.TrackForResource("T420", ResourceInfo{
			Type: "route",
			ID:   rid,
		})
		require.NoError(t, err)

		require.Len(t, tracker.changes["T420"], 1,
			"only a single resource is tracked")
	})
	t.Run("tracking an invalid change returns an error", func(t *testing.T) {
		tracker := NewChangeTracker()
		err := tracker.TrackForResource("T42", ResourceInfo{
			Type: "route",
			ID:   uuid.NewString(),
		})
		require.ErrorContains(t, err, "invalid change ID")
	})
	t.Run("tracking an invalid resource returns an error", func(t *testing.T) {
		tracker := NewChangeTracker()
		err := tracker.TrackForResource("T042", ResourceInfo{
			Type: "",
			ID:   uuid.NewString(),
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid type: ''")
		err = tracker.TrackForResource("T042", ResourceInfo{
			Type: "route",
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "id cannot be empty")
	})
	t.Run("tracking more than MaxChanges is a no-op", func(t *testing.T) {
		tracker := NewChangeTracker()
		for i := 0; i < MaxChanges+1; i++ {
			err := tracker.TrackForResource(ChangeID(fmt.Sprintf("T%03d", i)), ResourceInfo{
				Type: "route",
				ID:   uuid.NewString(),
			})
			require.NoError(t, err)
		}
		require.Len(t, tracker.changes, MaxChanges)
	})
	t.Run("tracking more than MaxResourcesPerChange is a no-op", func(t *testing.T) {
		tracker := NewChangeTracker()
		for i := 0; i < MaxResourcesPerChange+1; i++ {
			err := tracker.TrackForResource("T042", ResourceInfo{
				Type: "route",
				ID:   uuid.NewString(),
			})
			require.NoError(t, err)
		}
		require.Len(t, tracker.changes["T042"], MaxResourcesPerChange)
	})
}

func TestChangeTracker_Get(t *testing.T) {
	tracker := NewChangeTracker()
	for i := 0; i < MaxChanges+1; i++ {
		err := tracker.Track(ChangeID(fmt.Sprintf("T%03d", i)))
		require.NoError(t, err)
	}

	for i := 0; i < MaxResourcesPerChange+1; i++ {
		err := tracker.TrackForResource("T042", ResourceInfo{
			Type: "route",
			ID:   uuid.NewString(),
		})
		require.NoError(t, err)
	}

	changes := tracker.Get()
	var resourcesForChange []ResourceInfo
	for _, r := range changes.ChangeDetails {
		if r.ID == "T042" {
			resourcesForChange = r.Resources
		}
	}

	t.Run("returns all tracked changes", func(t *testing.T) {
		require.Len(t, changes.ChangeDetails, MaxChanges)
		require.Len(t, resourcesForChange, MaxResourcesPerChange)
	})
	t.Run("changes are sorted", func(t *testing.T) {
		require.True(t, slices.IsSortedFunc(changes.ChangeDetails,
			func(a, b ChangeDetail) bool {
				return a.ID < b.ID
			}),
		)
	})
	t.Run("resources within a change are sorted", func(t *testing.T) {
		require.True(t, slices.IsSortedFunc(resourcesForChange,
			func(a, b ResourceInfo) bool {
				return a.Type < b.Type || a.ID < b.ID
			}),
		)
	})
}
