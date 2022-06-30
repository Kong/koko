package seed

import (
	"errors"
	"sort"
	"sync"

	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
)

// Result defines a single resource, with some common fields
// shared across resources & its underling Protobuf message.
type Result struct {
	ID       string
	Tags     []string
	Resource proto.Message
}

// Results is a list of Result objects.
type Results struct {
	mu     sync.RWMutex
	all    []*Result
	byType map[model.Type][]*Result
}

// ToMap returns a JSON-based representation of the resource.
// This is safe for concurrent use.
func (r *Result) ToMap() (map[string]interface{}, error) {
	if r == nil {
		return nil, errors.New("unable to covert to map from nil result")
	}
	resourceJSON, err := json.ProtoJSONMarshal(r.Resource)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resourceJSON, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetByID attempts to find the given result by its ID. Returns nil when not found.
// This is safe for concurrent use.
func (r *Results) GetByID(id string) *Result {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return lo.FindOrElse(r.all, nil, func(rr *Result) bool { return rr.ID == id })
}

// IDs returns a sorted list of all resource IDs on the given results object.
// This is safe for concurrent use.
func (r *Results) IDs() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := lo.Map(r.all, func(rr *Result, _ int) string { return rr.ID })
	// Sorting for predictability.
	sort.Strings(ids)
	return ids
}

// ByType returns the given results for a particular resource type.
// In the event no results were created for the type, nil is returned.
// This is safe for concurrent use.
func (r *Results) ByType(typ model.Type) *Results {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.byType == nil {
		return nil
	}
	results, ok := r.byType[typ]
	if !ok {
		return nil
	}
	return &Results{
		all:    results,
		byType: map[model.Type][]*Result{typ: results},
	}
}

// AllByType returns all results keyed by its resource type.
// This is safe for concurrent use.
func (r *Results) AllByType() map[model.Type]*Results {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return lo.MapValues(r.byType, func(results []*Result, typ model.Type) *Results {
		return &Results{
			all:    results,
			byType: map[model.Type][]*Result{typ: results},
		}
	})
}

// All returns all results, regardless of resource type.
// This is safe for concurrent use.
func (r *Results) All() []*Result {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.all
}

// Add is used to insert new results for a specific type.
// This is safe for concurrent use.
func (r *Results) Add(typ model.Type, results *Results) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.all, r.byType[typ] = append(r.all, results.All()...), append(r.byType[typ], results.All()...)
}
