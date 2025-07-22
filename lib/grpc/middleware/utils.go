package middleware

import (
	"context"
	"net"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	otel_codes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func spanInfo(fullMethod, peerAddress string) (string, []attribute.KeyValue) {
	attrs := []attribute.KeyValue{semconv.RPCSystemGRPC}
	name, mAttrs := parseFullMethod(fullMethod)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(peerAddress)...)
	return name, attrs
}

func extract(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	return propagators.Extract(ctx, &metadataSupplier{
		metadata: &md,
	})
}

type metadataSupplier struct {
	metadata *metadata.MD
}

// assert that metadataSupplier implements the TextMapCarrier interface.
var _ propagation.TextMapCarrier = &metadataSupplier{}

func (s *metadataSupplier) Get(key string) string {
	values := s.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (s *metadataSupplier) Set(key string, value string) {
	s.metadata.Set(key, value)
}

func (s *metadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.metadata))
	for key := range *s.metadata {
		out = append(out, key)
	}
	return out
}

func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	if !strings.HasPrefix(fullMethod, "/") {
		// Invalid format, does not follow `/package.service/method`.
		return fullMethod, nil
	}
	name := fullMethod[1:]
	pos := strings.LastIndex(name, "/")
	if pos < 0 {
		// Invalid format, does not follow `/package.service/method`.
		return name, nil
	}
	service, method := name[:pos], name[pos+1:]

	var attrs []attribute.KeyValue
	if service != "" {
		attrs = append(attrs, semconv.RPCService(service))
	}
	if method != "" {
		attrs = append(attrs, semconv.RPCMethod(method))
	}
	return name, attrs
}

func peerFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}

func peerAttr(addr string) []attribute.KeyValue {
	host, p, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	var attr []attribute.KeyValue
	if ip := net.ParseIP(host); ip != nil {
		attr = []attribute.KeyValue{
			semconv.NetSockPeerAddr(host),
			semconv.NetSockPeerPort(port),
		}
	} else {
		attr = []attribute.KeyValue{
			semconv.NetPeerName(host),
			semconv.NetPeerPort(port),
		}
	}

	return attr
}

func serverStatus(grpcStatus *status.Status) (otel_codes.Code, string) {
	switch grpcStatus.Code() {
	case codes.Unknown,
		codes.DeadlineExceeded,
		codes.Unimplemented,
		codes.Internal,
		codes.Unavailable,
		codes.DataLoss:
		return otel_codes.Error, grpcStatus.Message()
	default:
		return otel_codes.Unset, ""
	}
}

func statusCodeAttr(c codes.Code) attribute.KeyValue {
	return GRPCStatusCodeKey.Int64(int64(c))
}

const (
	GRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")
)
