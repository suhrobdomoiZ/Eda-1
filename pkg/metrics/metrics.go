package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Общий список метрик
type Metrics struct {
	// HTTP/gRPC метрики
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge

	// Бизнес-метрики
	OrdersCreated    prometheus.Counter
	OrdersDelivered  prometheus.Counter
	OrdersCancelled  prometheus.Counter
	OrdersInProgress prometheus.Gauge
	OrdersWaiting    prometheus.Gauge
	OrdersDelivering prometheus.Gauge
	OrdersPrices     prometheus.Histogram
}

// Создание и регистрация всех метрик
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of requests",
			},
			[]string{"service", "method", "status"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "request_duration_seconds",
				Help:      "Request duration in seconds",
				Buckets:   prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
			},
			[]string{"service", "method"},
		),

		RequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "requests_in_flight",
				Help:      "Current number of requests being processed",
			},
		),

		OrdersCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "orders_created_total",
				Help:      "Total number of orders created",
			},
		),

		OrdersDelivered: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "orders_delivered_total",
				Help:      "Total number of orders delivered",
			},
		),

		OrdersCancelled: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "orders_cancelled_total",
				Help:      "Total number of orders cancelled",
			},
		),

		OrdersInProgress: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "orders_in_progress",
				Help:      "Current number of orders in progress",
			},
		),

		OrdersWaiting: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "orders_waiting",
				Help:      "Current number of orders waiting for courier",
			},
		),

		OrdersDelivering: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "orders_delivering",
				Help:      "Current number of orders being delivered",
			},
		),

		OrdersPrices: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "orders_prices_rub",
				Help:      "Distribution of order prices in rubles",
				Buckets:   []float64{100, 250, 500, 750, 1000, 1500, 2000, 3000, 5000},
			},
		),
	}
}

// HTTP handler для Prometheus scrape endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}

// Вызывается при создании заказа
func (m *Metrics) OnOrderCreated(price float64) {
	m.OrdersCreated.Inc()
	m.OrdersInProgress.Inc()
	m.OrdersPrices.Observe(price)
}

// Вызывается при создании заказа
func (m *Metrics) OnOrderCooked() {
	m.OrdersWaiting.Inc()
}

// Вызывается при создании заказа
func (m *Metrics) OnOrderPickUp() {
	m.OrdersWaiting.Dec()
	m.OrdersDelivering.Inc()
}

// Вызывается при завершении доставки
func (m *Metrics) OnOrderDelivered() {
	m.OrdersDelivering.Dec()
	m.OrdersDelivered.Inc()
	m.OrdersInProgress.Dec()
}

// Вызывается при отмене заказа
func (m *Metrics) OnOrderCancelled() {
	m.OrdersCancelled.Inc()
	m.OrdersInProgress.Dec()
}
