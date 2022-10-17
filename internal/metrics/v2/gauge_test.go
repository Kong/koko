package v2

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewGauge(t *testing.T) {
	type args struct {
		registry prometheus.Registerer
		opts     GaugeOpts
	}
	tests := []struct {
		name        string
		args        args
		want        Gauge
		shouldPanic bool
	}{
		{
			name: "should successfully create a gauge",
			args: args{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "websocket_connection_closed_total",
					Help:       "gauge_test help",
					LabelNames: []string{"ws_close_code"},
				},
			},
		},
		{
			name: "should panic when GaugeOpts is empty",
			args: args{
				registry: prometheus.NewRegistry(),
				opts:     GaugeOpts{},
			},
			shouldPanic: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.shouldPanic {
				require.Panics(t, func() {
					NewGauge(test.args.registry, test.args.opts)
				})
			} else {
				require.IsType(t, &prometheusGauge{}, NewGauge(test.args.registry, test.args.opts))
			}
		})
	}
}

func TestPrometheusGaugeAdd(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     GaugeOpts
	}
	type args struct {
		floatVal float64
		label    []Label
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect float64
	}{
		{
			name: "should add",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
			name: "should behave like sub when input is < 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(-2),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: float64(-2),
		},
		{
			name: "should succeed when when input is 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
			g := NewGauge(test.fields.registry, test.fields.opts)
			g.Add(test.args.floatVal, test.args.label...)
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			require.Equal(t, test.expect, family[0].Metric[0].Gauge.GetValue())
			require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
		})
	}
}

func TestPrometheusGaugeDec(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     GaugeOpts
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
			name: "should decrement",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{},
				},
			},
			args: args{
				label: []Label{},
			},
		},
		{
			name: "should decrement with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
			g := NewGauge(test.fields.registry, test.fields.opts)
			g.Dec(test.args.label...)
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			require.Equal(t, float64(-1), family[0].Metric[0].Gauge.GetValue())
			require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
		})
	}
}

func TestPrometheusGaugeInc(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     GaugeOpts
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
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{},
				},
			},
			args: args{
				label: []Label{},
			},
		},
		{
			name: "should increment with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
			g := NewGauge(test.fields.registry, test.fields.opts)
			g.Inc(test.args.label...)
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			require.Equal(t, float64(1), family[0].Metric[0].Gauge.GetValue())
			require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
		})
	}
}

func TestPrometheusGaugeSet(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     GaugeOpts
	}
	type args struct {
		floatVal float64
		label    []Label
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect float64
	}{
		{
			name: "should set gauge to valid number",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{},
				},
			},
			args: args{
				floatVal: float64(20),
				label:    []Label{},
			},
			expect: float64(20),
		},
		{
			name: "should set gauge to valid number with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(100),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: float64(100),
		},
		{
			name: "should succeed when input is < 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(-2),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: float64(-2),
		},
		{
			name: "should succeed when when input is 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
			g := NewGauge(test.fields.registry, test.fields.opts)
			g.Set(test.args.floatVal, test.args.label...)
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			require.Equal(t, test.expect, family[0].Metric[0].Gauge.GetValue())
			require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
		})
	}
}

func TestPrometheusGaugeSub(t *testing.T) {
	type fields struct {
		registry *prometheus.Registry
		opts     GaugeOpts
	}
	type args struct {
		floatVal float64
		label    []Label
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect float64
	}{
		{
			name: "should subtract",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{},
				},
			},
			args: args{
				floatVal: float64(1),
				label:    []Label{},
			},
			expect: float64(-1),
		},
		{
			name: "should subtract with labels",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(20),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: float64(-20),
		},
		{
			name: "should behave like add when input is < 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
					LabelNames: []string{"foo", "bar"},
				},
			},
			args: args{
				floatVal: float64(-2),
				label:    []Label{{"foo", "fooval"}, {"bar", "barval"}},
			},
			expect: float64(2),
		},
		{
			name: "should succeed when when input is 0",
			fields: fields{
				registry: prometheus.NewRegistry(),
				opts: GaugeOpts{
					Subsystem:  "cp",
					Name:       "gauge_test_total",
					Help:       "gauge_test help",
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
			g := NewGauge(test.fields.registry, test.fields.opts)
			g.Sub(test.args.floatVal, test.args.label...)
			family, err := test.fields.registry.Gather()
			require.NoError(t, err)
			require.Equal(t, test.expect, family[0].Metric[0].Gauge.GetValue())
			require.Equal(t, len(test.fields.opts.LabelNames), len(family[0].Metric[0].GetLabel()))
		})
	}
}
