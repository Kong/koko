package v2

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestToPrometheusLabel(t *testing.T) {
	tests := []struct {
		name   string
		labels []Label
		want   prometheus.Labels
	}{
		{
			name:   "should return a valid list of labels",
			labels: []Label{{Key: "foo", Value: "bar"}},
			want:   prometheus.Labels{"foo": "bar"},
		},
		{
			name:   "should return an empty label list when input is empty",
			labels: []Label{},
			want:   prometheus.Labels{},
		},
		{
			name:   "should return an empty label list when input is nil",
			labels: nil,
			want:   prometheus.Labels{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := toPrometheusLabel(test.labels...); !reflect.DeepEqual(got, test.want) {
				t.Errorf("toPrometheusLabel() = %v, want %v", got, test.want)
			}
		})
	}
}
