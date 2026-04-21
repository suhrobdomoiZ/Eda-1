package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// gRPC интерцептор для сбора метрик из входящих запросов
func (m *Metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// 1. До вызова: увеличиваем счётчик активных запросов
		m.RequestsInFlight.Inc()
		defer m.RequestsInFlight.Dec() // Уменьшим при выходе

		// 2. Засекаем время начала
		start := time.Now()

		// 3. Вызываем настоящий хендлер
		resp, err := handler(ctx, req)

		// 4. После вызова: считаем длительность
		duration := time.Since(start).Seconds()

		// 5. Определяем статус (success / error)
		statusCode := "success"
		if err != nil {
			st, _ := status.FromError(err)
			statusCode = st.Code().String()
		}

		// 6. Извлекаем имя сервиса из полного метода
		// info.FullMethod = "/customer.CustomerAPI/CreateOrder"
		serviceName := extractServiceName(info.FullMethod)
		methodName := extractMethodName(info.FullMethod)

		// 7. Записываем метрики
		m.RequestsTotal.WithLabelValues(serviceName, methodName, statusCode).Inc()
		m.RequestDuration.WithLabelValues(serviceName, methodName).Observe(duration)

		return resp, err
	}
}

// Для клиентских вызовов и запросов между микросервисами
func (m *Metrics) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start).Seconds()

		statusCode := "success"
		if err != nil {
			st, _ := status.FromError(err)
			statusCode = st.Code().String()
		}

		serviceName := extractServiceName(method)
		methodName := extractMethodName(method)

		m.RequestsTotal.WithLabelValues(serviceName+"_client", methodName, statusCode).Inc()
		m.RequestDuration.WithLabelValues(serviceName+"_client", methodName).Observe(duration)

		return err
	}
}

// Вспомогательные функции
func extractServiceName(fullMethod string) string {
	// "/customer.CustomerAPI/CreateOrder" → "customer"
	if len(fullMethod) > 1 && fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}
	for i := 0; i < len(fullMethod); i++ {
		if fullMethod[i] == '.' || fullMethod[i] == '/' {
			return fullMethod[:i]
		}
	}
	return "unknown"
}

func extractMethodName(fullMethod string) string {
	// "/customer.CustomerAPI/CreateOrder" → "CreateOrder"
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return fullMethod[i+1:]
		}
	}
	return fullMethod
}
