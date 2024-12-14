package transport

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	grpcTransport "github.com/go-kit/kit/transport/grpc"
	grpcProm "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker"
	"github.com/thyyl/chatr/pkg/common"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

func InitializeGrpcServer(name string, logger common.GrpcLog) *grpc.Server {
	grpcOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 8), // increase to 8 MB (default: 4 MB)
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // terminate the connection if a client pings more than once every 5 seconds
			PermitWithoutStream: true,            // allow pings even when there are no active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,  // if a client is idle for 15 seconds, send a GOAWAY
			MaxConnectionAge:      600 * time.Second, // if any connection is alive for more than maxConnectionAge, send a GOAWAY
			MaxConnectionAgeGrace: 5 * time.Second,   // allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:                  5 * time.Second,   // ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout:               1 * time.Second,   // wait 1 second for the ping ack before assuming the connection is dead
		}),
	}

	serverMetrics := grpcProm.NewServerMetrics(
		grpcProm.WithServerCounterOptions(
			func(o *prometheus.CounterOpts) {
				o.Namespace = name
			},
			grpcProm.WithConstLabels(prometheus.Labels{"serviceID": name}),
		),
		grpcProm.WithServerHandlingTimeHistogram(
			grpcProm.WithHistogramConstLabels(prometheus.Labels{"serviceID": name}),
			grpcProm.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	prometheus.MustRegister(serverMetrics)
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}
	panicsTotal := promauto.NewCounter(prometheus.CounterOpts{
		Namespace:   name,
		Name:        "grpc_req_panics_recovered_total",
		Help:        "Total number of gRPC requests recovered from internal panic.",
		ConstLabels: prometheus.Labels{"serviceID": name},
	})
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		logger.Error("recovered from panic, stack: " + string(debug.Stack()))
		return status.Errorf(codes.Internal, "%s", p)
	}

	logTraceId := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}
	logOptions := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		logging.WithDurationField(logging.DurationToTimeMillisFields),
		logging.WithFieldsFromContext(logTraceId),
	}

	grpcOptions = append(grpcOptions,
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			serverMetrics.StreamServerInterceptor(grpcProm.WithExemplarFromContext(exemplarFromContext)),
			logging.StreamServerInterceptor(interceptorLogger(logger), logOptions...),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			serverMetrics.UnaryServerInterceptor(grpcProm.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryServerInterceptor(interceptorLogger(logger), logging.WithFieldsFromContext(logTraceId)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	)

	grpcServer := grpc.NewServer(grpcOptions...)
	serverMetrics.InitializeMetrics(grpcServer)
	return grpcServer
}

func InitializeGrpcClient(serviceHost string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	scheme := "dns"

	retryOpts := []retry.CallOption{
		// generate waits between 900ms to 1100ms
		retry.WithBackoff(retry.BackoffLinearWithJitter(1*time.Second, 0.1)),
		retry.WithMax(3),
		retry.WithCodes(codes.Unavailable, codes.Aborted),
		retry.WithPerRetryTimeout(3 * time.Second),
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	dialOptions = append(dialOptions,
		grpc.WithDisableServiceConfig(),
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingPolicy": "round_robin"
		}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
		grpc.WithChainStreamInterceptor(
			otelgrpc.StreamClientInterceptor(),
			retry.StreamClientInterceptor(retryOpts...),
		),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
			retry.UnaryClientInterceptor(retryOpts...),
		),
		//grpc.WithBlock(),
	)

	slog.Info("connecting to grpc host: " + serviceHost)
	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:///%s", scheme, serviceHost),
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewGrpcEndpoint(conn *grpc.ClientConn, serviceID, serviceName, method string, grpcReply interface{}) endpoint.Endpoint {
	var options []grpcTransport.ClientOption
	var (
		ep         endpoint.Endpoint
		endpointer sd.FixedEndpointer
	)

	ep = grpcTransport.NewClient(
		conn,
		serviceName,
		method,
		encodeGRPCRequest,
		decodeGRPCResponse,
		grpcReply,
		append(options, grpcTransport.ClientBefore(grpcTransport.SetRequestHeader(common.ServiceIdHeader, serviceID)))...,
	).Endpoint()
	ep = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    serviceName + "." + method,
		Timeout: 60 * time.Second,
	}))(ep)
	endpointer = append(endpointer, ep)
	// timeout for the whole invocation
	ep = lb.Retry(1, 15*time.Second, lb.NewRoundRobin(endpointer))

	return ep
}
