package gateway

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	http_helper "github.com/zoenion/common/http-helper"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

type Gateway struct {
	gs                         *grpc.Server
	hs                         *http.Server
	config                     *Config
	router                     *mux.Router
	listenerGRPC, listenerHTTP net.Listener
}

func New(config *Config) *Gateway {
	return &Gateway{
		config: config,
	}
}

func (g *Gateway) Start() error {
	err := g.listen()

	if err != nil {
		return err
	}

	if g.config.GRPC != nil {
		go g.startGRPC()
	}

	if g.config.HTTP != nil {
		go g.startHTTP()
	}
	return nil
}

func (g *Gateway) Stop() {
	g.gs.GracefulStop()
	if err := g.listenerGRPC.Close(); err != nil {
		log.Println("stopped grpc listener with error:", err)
	}

	ctx := context.Background()
	if err := g.hs.Shutdown(ctx); err != nil {
		log.Println("stopped grpc listener with error:", err)
	}

	if err := g.listenerHTTP.Close(); err != nil {
		log.Println("stopped http listener with error:", err)
	}
}

func (g *Gateway) listen() (err error) {
	if g.config.GRPC != nil {
		if g.config.GRPC.Address == "" {
			g.config.GRPC.Address = ":"
		}
		g.listenerGRPC, err = net.Listen("tcp", g.config.GRPC.Address)
		if err != nil {
			return err
		}
		g.config.GRPC.Address = g.listenerGRPC.Addr().String()
	}

	if g.config.HTTP != nil {
		if g.config.HTTP.Address == "" {
			g.config.HTTP.Address = ":"
		}
		g.listenerHTTP, err = net.Listen("tcp", g.config.HTTP.Address)
		if err != nil {
			return err
		}
		g.config.HTTP.Address = g.listenerHTTP.Addr().String()
	}
	return nil
}

func (g *Gateway) startGRPC() {
	log.Printf("starting gRPC server at %s", g.config.GRPC.Address)

	g.gs = grpc.NewServer()
	g.config.GRPC.RegisterHandlerFunc(g.gs)
	if err := g.gs.Serve(g.listenerGRPC); err != nil {
		log.Println("gRPC server stopped, cause:", err)
	}
}

func (g *Gateway) startHTTP() {
	log.Printf("starting HTTP server at %s", g.config.HTTP.Address)

	ctx := context.Background()
	endpoint := flag.String("grpc-server-endpoint", g.config.GRPC.Address, "gRPC server endpoint")

	serverMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := g.config.HTTP.WireGRPCFunc(ctx, serverMux, *endpoint, opts); err != nil {
		log.Println("failed to start HTTP gateway, cause: ", err)
		return
	}

	handler := http_helper.HttpBasicMiddlewareStack(context.Background(), serverMux.ServeHTTP, nil)
	for _, m := range g.config.HTTP.MiddlewareList {
		handler = m(handler)
	}

	g.hs = &http.Server{
		Addr:    g.config.HTTP.Address,
		Handler: handler,
	}

	if err := g.hs.Serve(g.listenerHTTP); err != nil {
		log.Println("HTTP server stopped, cause:", err)
	}
}
