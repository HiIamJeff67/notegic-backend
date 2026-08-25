package metrics

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MeterInterface interface {
	Count(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue)
	Duration(ctx context.Context, name string, value time.Duration, attributes ...attribute.KeyValue)
	Bytes(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue)
	UpDown(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue)
	Value(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue)
}

type Meter struct {
	meter              metric.Meter
	byteCounters       sync.Map
	counters           sync.Map
	durationHistograms sync.Map
	upDownCounters     sync.Map
	valueHistograms    sync.Map
}

func NewMeter(meter metric.Meter) MeterInterface {
	return &Meter{meter: meter}
}

var (
	NotegicMeter MeterInterface
)

func (m *Meter) Count(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue) {
	counter, ok := m.counters.Load(name)
	if !ok {
		created, err := m.meter.Int64Counter(name)
		if err != nil {
			return
		}

		counter, _ = m.counters.LoadOrStore(name, created)
	}

	counter.(metric.Int64Counter).Add(ctx, value, metric.WithAttributes(attributes...))
}

func (m *Meter) Duration(ctx context.Context, name string, value time.Duration, attributes ...attribute.KeyValue) {
	histogram, ok := m.durationHistograms.Load(name)
	if !ok {
		created, err := m.meter.Float64Histogram(name, metric.WithUnit("s"))
		if err != nil {
			return
		}

		histogram, _ = m.durationHistograms.LoadOrStore(name, created)
	}

	histogram.(metric.Float64Histogram).Record(ctx, value.Seconds(), metric.WithAttributes(attributes...))
}

func (m *Meter) Bytes(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue) {
	counter, ok := m.byteCounters.Load(name)
	if !ok {
		created, err := m.meter.Int64Counter(name, metric.WithUnit("By"))
		if err != nil {
			return
		}

		counter, _ = m.byteCounters.LoadOrStore(name, created)
	}

	counter.(metric.Int64Counter).Add(ctx, value, metric.WithAttributes(attributes...))
}

func (m *Meter) UpDown(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue) {
	counter, ok := m.upDownCounters.Load(name)
	if !ok {
		created, err := m.meter.Int64UpDownCounter(name)
		if err != nil {
			return
		}

		counter, _ = m.upDownCounters.LoadOrStore(name, created)
	}

	counter.(metric.Int64UpDownCounter).Add(ctx, value, metric.WithAttributes(attributes...))
}

func (m *Meter) Value(ctx context.Context, name string, value int64, attributes ...attribute.KeyValue) {
	histogram, ok := m.valueHistograms.Load(name)
	if !ok {
		created, err := m.meter.Int64Histogram(name)
		if err != nil {
			return
		}

		histogram, _ = m.valueHistograms.LoadOrStore(name, created)
	}

	histogram.(metric.Int64Histogram).Record(ctx, value, metric.WithAttributes(attributes...))
}
