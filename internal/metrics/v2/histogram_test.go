package v2

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHistogram(t *testing.T) {
	type args struct {
		registry prometheus.Registerer
		opts     HistogramOpts
	}
	tests := []struct {
		name        string
		args        args
		wantBucket  []float64
		shouldPanic bool
	}{
		{
			name: "should successfully create a histogram",
			args: args{
				registry: prometheus.NewRegistry(),
				opts: HistogramOpts{
					Subsystem:  "test_subsystem",
					Name:       "histogram_test_total",
					Help:       "histogram_test help",
					LabelNames: []string{"histogram_test_label"},
				},
			},
		},
		{
			name: "should successfully create a histogram with default bucket values",
			args: args{
				registry: prometheus.NewRegistry(),
				opts: HistogramOpts{
					Subsystem:  "test_subsystem",
					Name:       "histogram_test_total",
					Help:       "histogram_test help",
					LabelNames: []string{"histogram_test_label"},
				},
			},
		},
		{
			name: "should successfully create a histogram with custom bucket values",
			args: args{
				registry: prometheus.NewRegistry(),
				opts: HistogramOpts{
					Subsystem:  "test_subsystem",
					Name:       "histogram_test_total",
					Help:       "histogram_test help",
					Buckets:    []float64{1, 2, 3, 4, 5, 6},
					LabelNames: []string{"histogram_test_label"},
				},
			},
		},
		{
			name: "should panic when HistogramOpts is empty",
			args: args{
				registry: prometheus.NewRegistry(),
				opts:     HistogramOpts{},
			},
			shouldPanic: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.shouldPanic {
				require.Panics(t, func() {
					NewHistogram(test.args.registry, test.args.opts)
				})
			} else {
				require.IsType(t, &prometheusHistogram{}, NewHistogram(test.args.registry, test.args.opts))
			}
		})
	}
}

func Test_prometheusHistogram_Observe(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     HistogramOpts
	}
	type args struct {
		floatVals []float64
		label     []Label
	}
	type expected struct {
		sampleSum   float64
		sampleCount uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect expected
	}{
		{
			name: "should observe",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: HistogramOpts{
					Subsystem:  "cp",
					Name:       "histogram_test_total",
					Help:       "histogram_test help",
					Buckets:    []float64{1, 2, 3, 4, 5, 6},
					LabelNames: []string{},
				},
			},
			args: args{
				floatVals: []float64{1, 1, 2, 2},
				label:     []Label{},
			},
			expect: expected{
				sampleSum:   float64(6),
				sampleCount: uint64(4),
			},
		},
		{
			name: "should observe with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: HistogramOpts{
					Subsystem:  "cp",
					Name:       "histogram_test_total",
					Help:       "histogram_test help",
					Buckets:    []float64{1, 2, 3, 4, 5, 6},
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVals: []float64{1, 1, 2, 2},
				label:     []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: expected{
				sampleSum:   float64(6),
				sampleCount: uint64(4),
			},
		},
		{
			name: "should observe when input is < 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: HistogramOpts{
					Subsystem:  "cp",
					Name:       "histogram_test_total",
					Help:       "histogram_test help",
					Buckets:    []float64{1, 2, 3, 4, 5, 6},
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVals: []float64{-1, -1, -2, -2},
				label:     []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: expected{
				sampleSum:   float64(-6),
				sampleCount: uint64(4),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := NewHistogram(test.fields.registry, test.fields.opts)
			for _, val := range test.args.floatVals {
				h.Observe(val, test.args.label...)
			}
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			assert.Greater(t, len(family), 0)
			assert.Greater(t, len(family[0].Metric), 0)
			histogram := family[0].Metric[0].Histogram
			require.Equal(t, test.expect.sampleSum, histogram.GetSampleSum())
			require.Equal(t, test.expect.sampleCount, histogram.GetSampleCount())
		})
	}
}
