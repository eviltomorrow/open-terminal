package server

import (
	"context"
	"fmt"
	"net"
	"path/filepath"

	"github.com/eviltomorrow/king/lib/certificate"
	"github.com/eviltomorrow/king/lib/etcd"
	"github.com/eviltomorrow/open-terminal/lib/buildinfo"
	"github.com/eviltomorrow/open-terminal/lib/finalizer"
	"github.com/eviltomorrow/open-terminal/lib/grpc/middleware"
	"github.com/eviltomorrow/open-terminal/lib/log"
	"github.com/eviltomorrow/open-terminal/lib/network"
	"github.com/eviltomorrow/open-terminal/lib/system"
	"github.com/eviltomorrow/open-terminal/lib/zlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type GRPC struct {
	network *network.Config
	log     *log.Config

	server     *grpc.Server
	ctx        context.Context
	cancel     func()
	revokeFunc func() error

	RegisteredAPI []func(*grpc.Server)
}

func NewGRPC(network *network.Config, log *log.Config, supported ...func(*grpc.Server)) *GRPC {
	return &GRPC{
		network: network,
		log:     log,

		RegisteredAPI: supported,
	}
}

func (g *GRPC) Serve() error {
	midlog, err := middleware.InitLogger(&zlog.Config{
		Level:  g.log.Level,
		Format: "json",
		File: zlog.FileLogConfig{
			Filename:    filepath.Join(system.Directory.LogDir, "access.log"),
			MaxSize:     100,
			MaxDays:     30,
			MaxBackups:  90,
			Compression: "gzip",
		},
		DisableStacktrace: true,
		DisableStdlog:     g.log.DisableStdlog,
	})
	if err != nil {
		return fmt.Errorf("init middleware log failure, nest error: %v", err)
	}
	finalizer.RegisterCleanupFuncs(midlog)

	var creds credentials.TransportCredentials
	if !g.network.DisableTLS {
		ipList := make([]string, 0, 4)
		ipList = append(ipList, system.Network.BindIP)
		ipList = append(ipList, g.network.BindIP)
		if system.Network.AccessIP == "" {
			ipList = append(ipList, system.Network.AccessIP)
		}

		err := certificate.CreateOrOverrideFile(certificate.BuildDefaultAppInfo(ipList), &certificate.Config{
			CaCertFile:     filepath.Join(system.Directory.UsrDir, "certs/ca.crt"),
			CaKeyFile:      filepath.Join(system.Directory.UsrDir, "certs/ca.key"),
			ClientCertFile: filepath.Join(system.Directory.VarDir, "certs/client.crt"),
			ClientKeyFile:  filepath.Join(system.Directory.VarDir, "certs/client.pem"),
			ServerCertFile: filepath.Join(system.Directory.VarDir, "certs/server.crt"),
			ServerKeyFile:  filepath.Join(system.Directory.VarDir, "certs/server.pem"),
		})
		if err != nil {
			return err
		}

		creds, err = certificate.LoadServerCredentials(&certificate.Config{
			CaCertFile:     filepath.Join(system.Directory.UsrDir, "certs/ca.crt"),
			ServerCertFile: filepath.Join(system.Directory.VarDir, "certs/server.crt"),
			ServerKeyFile:  filepath.Join(system.Directory.VarDir, "certs/server.pem"),
		})
		if err != nil {
			return err
		}
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", g.network.BindIP, g.network.BindPort))
	if err != nil {
		return err
	}

	g.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecoveryInterceptor,
			middleware.UnaryServerLogInterceptor,
		),
		grpc.ChainStreamInterceptor(
			middleware.StreamServerRecoveryInterceptor,
			// middleware.StreamServerLogInterceptor,
		),
		grpc.Creds(creds),
		// grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	reflection.Register(g.server)
	for _, register := range g.RegisteredAPI {
		register(g.server)
	}

	go func() {
		if err := g.server.Serve(listen); err != nil {
			zlog.Fatal("server(grpc) startup failure", zap.Error(err))
		}
	}()

	g.ctx, g.cancel = context.WithCancel(context.Background())
	if etcd.Client != nil {
		g.revokeFunc, err = etcd.RegisterService(g.ctx, buildinfo.AppName, system.Network.AccessIP, g.network.BindPort, 10)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GRPC) Stop() error {
	if g.revokeFunc != nil {
		g.revokeFunc()
	}
	if g.server != nil {
		g.server.GracefulStop()
	}
	g.cancel()

	return nil
}
