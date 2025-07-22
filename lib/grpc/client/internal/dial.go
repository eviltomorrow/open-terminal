package internal

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

func DialWithTarget(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithNoProxy(),
	)
}

func DialWithEtcd(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target,
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithNoProxy(),
	)
}
