package v2

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name        string
		opts        CounterOpts
		shouldPanic bool
	}{
		{
			name: "should successfully create a counter",
			opts: CounterOpts{
				Subsystem:  "cp",
				Name:       "websocket_connection_closed_total",
				Help:       "Number of data-plane websocket connections closed",
				LabelNames: []string{"ws_close_code"},
			},
		},
		{
			name:        "should panic when CounterOpts is empty",
			opts:        CounterOpts{},
			shouldPanic: true,
		},
	}
	for _, test := range tests {
		r := prometheus.NewRegistry()
		t.Run(test.name, func(t *testing.T) {
			if test.shouldPanic {
				require.Panics(t, func() {
					NewCounter(r, test.opts)
				})
			} else {
				require.IsType(t, &prometheusCounter{}, NewCounter(r, test.opts))
			}
		})
	}
}

func TestPrometheusCounterAdd(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     CounterOpts
	}
	type args struct {
		floatVal float64
		label    []Label
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expect      float64
		shouldPanic bool
	}{
		{
			name: "should add",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: CounterOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "Number of data-plane websocket connections closed",
					LabelNames: []string{},
				},
			},
			args: args{
				floatVal: float64(1),
				label:    []Label{},
			},
			expect: float64(1),
		},
		{
			name: "should add with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: CounterOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "Number of data-plane websocket connections closed",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(20),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: float64(20),
		},
		{
			name: "should panic when input is < 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: CounterOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "Number of data-plane websocket connections closed",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(-2),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			shouldPanic: true,
		},
		{
			name: "should succeed when when input is 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: CounterOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "Number of data-plane websocket connections closed",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				label: []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewCounter(test.fields.registry, test.fields.opts)
			if test.shouldPanic {
				require.Panics(t, func() {
					c.Add(test.args.floatVal, test.args.label...)
				})
			} else {
				c.Add(test.args.floatVal, test.args.label...)
				family, err := test.fields.registry.Gather()
				require.NoError(t, err)
				require.Equal(t, test.expect, family[0].Metric[0].Counter.GetValue())
				require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
			}
		})
	}
}

func TestPrometheusCounterInc(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     CounterOpts
	}
	type args struct {
		label []Label
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "should increment",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: CounterOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "Number of data-plane websocket connections closed",
					LabelNames: []string{},
				},
			},
		},
		{
			name: "should increment with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: CounterOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "Number of data-plane websocket connections closed",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				[]Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewCounter(test.fields.registry, test.fields.opts)
			c.Inc(test.args.label...)
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			require.Equal(t, float64(1), family[0].Metric[0].Counter.GetValue())
			require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
		})
	}
}
