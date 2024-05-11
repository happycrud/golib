package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"github.com/urfave/negroni"
	negroniprometheus "github.com/zbindenren/negroni-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type TLSConfig struct {
	Port     int
	KeyPath  string
	CertPath string
}
type CorsConfig struct {
}
type AppOptions struct {
	// if nil not run tls
	tls *TLSConfig
	// if nil not run tcp // if Port == TLS.Port port run on same port
	port *int
	//
	enableGRPCWeb bool

	corsOptions *cors.Options
}

type AppOption func(aos *AppOptions)

func WithTLSConfig(c *TLSConfig) AppOption {
	return func(aos *AppOptions) {
		aos.tls = c
	}
}
func WithPort(p int) AppOption {
	return func(aos *AppOptions) {
		aos.port = &p
	}
}
func WihtGrpcWeb(open bool) AppOption {
	return func(aos *AppOptions) {
		aos.enableGRPCWeb = open
	}
}

func WithCorsOptions(opt *cors.Options) AppOption {
	return func(aos *AppOptions) {
		aos.corsOptions = opt
	}
}

type App struct {
	options    *AppOptions
	bothTCPTLS net.Listener
	onlyTCP    net.Listener
	onlyTLS    net.Listener
	rpc        *grpc.Server
	h1         *http.Server
	wg         sync.WaitGroup
	mList      []cmux.CMux
	httpmux    *http.ServeMux
}

func New(options ...AppOption) *App {
	a := &App{
		options: &AppOptions{},
		rpc:     grpc.NewServer(),
		h1:      &http.Server{},
		wg:      sync.WaitGroup{},
		httpmux: http.NewServeMux(),
	}
	for _, opt := range options {
		opt(a.options)
	}
	grpc_health_v1.RegisterHealthServer(a.rpc, health.NewServer())
	reflection.Register(a.rpc)
	return a
}
func (a *App) listenTLS() {
	cert, err := tls.LoadX509KeyPair(a.options.tls.CertPath, a.options.tls.KeyPath)
	if err != nil {
		panic(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	a.onlyTLS, err = tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", a.options.tls.Port), cfg)
	if err != nil {
		panic(err)
	}
}
func (a *App) listenTCP() {
	var err error
	a.onlyTCP, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *a.options.port))
	if err != nil {
		panic(err)
	}
}
func (a *App) listens() {
	slog.Info("init listens")
	var err error
	if a.options.port == nil && a.options.tls == nil {
		panic("need listen port")
	}
	if a.options.port != nil && a.options.tls != nil {
		if *a.options.port == a.options.tls.Port {
			a.bothTCPTLS, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *a.options.port))
			if err != nil {
				panic(err)
			}
		} else {
			a.listenTCP()
			a.listenTLS()
		}
		return
	}
	if a.options.port != nil {
		a.listenTCP()
	}
	if a.options.tls != nil {
		a.listenTLS()
	}

}
func (a *App) startServe(l net.Listener) {
	slog.Info(" server ", "listen", l.Addr())
	m := cmux.New(l)
	httpL := m.Match(cmux.HTTP1Fast())
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	a.wg.Add(3)
	go func() {
		defer a.wg.Done()
		if err := a.rpc.Serve(grpcL); err != nil {
			slog.Error("grpc stop", "error", err)
		}
	}()

	go func() {
		defer a.wg.Done()
		if err := a.h1.Serve(httpL); err != nil {
			slog.Error("h1 stop", "error", err)
		}
	}()

	go func() {
		defer a.wg.Done()
		if err := m.Serve(); err != nil {
			slog.Error("cmux stop", "error", err)
		}
	}()
	a.mList = append(a.mList, m)
}
func (a *App) start() {
	slog.Info("mux ")
	if a.bothTCPTLS != nil {
		m := cmux.New(a.bothTCPTLS)
		grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
		httpL := m.Match(cmux.HTTP1Fast())
		other := m.Match(cmux.Any())
		a.wg.Add(3)
		go func() {
			defer a.wg.Done()
			a.h1.Serve(grpcL)
		}()
		go func() {
			defer a.wg.Done()
			a.rpc.Serve(httpL)
		}()
		go func() {
			defer a.wg.Done()
			m.Serve()
		}()
		a.mList = append(a.mList, m)
		cert, err := tls.LoadX509KeyPair(a.options.tls.CertPath, a.options.tls.KeyPath)
		if err != nil {
			panic(err)
		}
		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		tlsL := tls.NewListener(other, cfg)
		a.startServe(tlsL)
		return
	}
	if a.onlyTCP != nil {
		a.startServe(a.onlyTCP)
	}
	if a.onlyTLS != nil {
		a.startServe(a.onlyTLS)
	}

}
func (a *App) Run() {
	slog.Info("start server")
	a.listens()
	a.loadH1Handler()
	a.start()

	// wait sign to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for range ch {
		a.h1.Shutdown(context.Background())
		a.rpc.GracefulStop()
		for _, v := range a.mList {
			v.Close()
		}
		slog.Info("finish server")
		a.wg.Wait()
		return
	}

}

func (a *App) RegisteGrpcService(desc *grpc.ServiceDesc, s any) {
	a.rpc.RegisterService(desc, s)
}

func (a *App) loadH1Handler() {
	slog.Info(" initHTTPMux ")
	var hander http.Handler
	hander = a.httpmux
	if a.options.enableGRPCWeb {
		wrappedGrpc := grpcweb.WrapServer(a.rpc)
		hander = http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(resp, req)
				return
			}
			// Fall back to other servers.
			hander.ServeHTTP(resp, req)
		})

	}
	if a.options.corsOptions != nil {
		hander = cors.New(*a.options.corsOptions).Handler(hander)
	}
	// log metric trace recovery  hander
	n := negroni.Classic()
	p := negroniprometheus.NewMiddleware("app")
	n = n.With(p)
	n.UseHandler(hander)
	a.h1.Handler = n
	a.httpmux.Handle("/metrics", promhttp.Handler())
}

func (a *App) Get(path string, h http.Handler) {
	a.httpmux.Handle(path, h)
}
