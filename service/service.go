package service

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/middleware"
	"github.com/ONSdigital/dp-geodata-api/api"
	"github.com/ONSdigital/dp-geodata-api/cache"
	"github.com/ONSdigital/dp-geodata-api/cantabular"
	"github.com/ONSdigital/dp-geodata-api/config"
	"github.com/ONSdigital/dp-geodata-api/handlers"
	"github.com/ONSdigital/dp-geodata-api/metadata"
	"github.com/ONSdigital/dp-geodata-api/pkg/database"
	"github.com/ONSdigital/dp-geodata-api/pkg/geodata"
	"github.com/ONSdigital/dp-geodata-api/postcode"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Service contains all the configs, server and clients to run the dp-topic-api API
type Service struct {
	Config      *config.Config
	Server      HTTPServer
	ServiceList *ExternalServiceList
	HealthCheck HealthChecker
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {

	log.Info(ctx, "config on startup", log.Data{"config": cfg, "build_time": buildTime, "git-commit": gitCommit})
	log.Info(ctx, "running service")

	var cant *cantabular.Client
	if cfg.EnableCantabular {
		cant = cantabular.New(cfg.CantabularURL, cfg.CantabularUser, os.Getenv("CANT_PW"))
	}

	var db *database.Database
	var queryGeodata *geodata.Geodata
	var md *metadata.Metadata
	var pc *postcode.Postcode
	var err error
	if cfg.EnableDatabase {
		// figure out postgres password
		pgpwd := os.Getenv("PGPASSWORD")
		if pgpwd == "" {
			aws, err := serviceList.GetAWS()
			if err != nil {
				return nil, err
			}
			pgpwd, err = aws.GetSecret(ctx, os.Getenv("FI_PG_SECRET_ID"))
			if err != nil {
				return nil, err
			}
		}

		// open postgres connection

		db, err = database.Open("pgx", database.GetDSN(pgpwd))
		if err != nil {
			return nil, err
		}

		// set up our query functionality if we have a db
		queryGeodata, err = geodata.New(db, cant, cfg.MaxMetrics)
		if err != nil {
			return nil, err
		}

		// metadata.New can set up gorm itself, but it calls GetDSN without an
		// argument, so it cannot know about passwords held in AWS secrets.
		//
		// We loop here in case the db isn't up yet (happens when using docker compose).
		// (Looping doesn't seem to be needed for the pgx connection, for some reason.)
		var gdb *gorm.DB
		for try := 0; try < 5; try++ {
			gdb, err = gorm.Open(postgres.Open(database.GetDSN(pgpwd)), &gorm.Config{
				//	Logger: logger.Default.LogMode(logger.Info), // display SQL
			})
			if err == nil {
				break
			}
			log.Info(ctx, "gorm opening")
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			return nil, err
		}

		md, err = metadata.New(gdb)
		if err != nil {
			return nil, err
		}

		pc = postcode.New(gdb)

	}

	cm, err := cache.New(cfg.CacheTTL, cfg.CacheSize)
	if err != nil {
		return nil, err
	}

	// Who am I?
	baseurl, prefix, err := ParseBaseURL(cfg.BindAddr, cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	if prefix == "" {
		prefix = "/v1/geodata" // for backward compatibility
	}

	// Setup the API
	a := handlers.New(
		cfg.APIToken,
		cfg.BindAddr,
		baseurl+prefix,
		cfg.EnableHeaderAuth,
		cfg.DoCors,
		true, // always include private handlers for now
		queryGeodata,
		md,
		cm,
		pc,
	)

	// Setup health checks
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}
	if err := registerCheckers(ctx, hc, db, md, cant); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}
	hc.Start(ctx)

	clientInfo := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info(r.Context(), "client info", log.Data{"addr": r.RemoteAddr})
			h.ServeHTTP(w, r)
		})
	}

	timeoutHandler := func(h http.Handler) http.Handler {
		return http.TimeoutHandler(h, cfg.WriteTimeout, "operation timed out\n")
	}

	stripOptionalPrefix := func(h http.Handler) http.Handler {
		strip := func(w http.ResponseWriter, r *http.Request) {
			p := strings.TrimPrefix(r.URL.Path, prefix)
			rp := strings.TrimPrefix(r.URL.RawPath, prefix)
			if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p
				r2.URL.RawPath = rp
				h.ServeHTTP(w, r2)
			} else {
				h.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(strip)
	}

	// build handler chain
	chain := alice.New(
		clientInfo,
		middleware.Whitelist(middleware.HealthcheckFilter(hc.Handler)),
		timeoutHandler,
		stripOptionalPrefix,
	).Then(api.Handler(a))

	// bind router handler to http server
	s := serviceList.GetHTTPServer(cfg.BindAddr, chain)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		Config:      cfg,
		HealthCheck: hc,
		ServiceList: serviceList,
		Server:      s,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutown gracefully closes up
	var hasShutdownError bool

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// TODO: Close other dependencies, in the expected order
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func registerCheckers(ctx context.Context,
	hc HealthChecker,
	db *database.Database,
	md *metadata.Metadata,
	cant *cantabular.Client,
) (err error) {
	if db != nil {
		err = hc.AddCheck("postgres", db.Checker)
	}
	if md != nil {
		err = hc.AddCheck("gorm", md.Checker)
	}
	if cant != nil {
		err = hc.AddCheck("cantabular", cant.Checker)
	}
	return err
}

// ParseBaseURL figures out this instance's scheme, host, port and endpoint prefix.
// These pieces of info are used to strip optional endpoint prefixes from incomimg
// requests and to help swaggerui construct self-referential URLs.
//
// If baseurl is set, then use it exclusively.
// Else if bindaddr has a missing or 0.0.0.0 host, use os.Hostname()
// Else use host and port from bindaddr.
func ParseBaseURL(bindaddr, baseurl string) (server, prefix string, err error) {
	if baseurl != "" {
		u, err := url.Parse(baseurl)
		if err != nil {
			return "", "", err
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return "", "", errors.New("BASEURL: http or https expected")
		}
		if u.User != nil {
			return "", "", errors.New("BASEURL: user/pass not allowed")
		}
		if u.Hostname() == "" {
			return "", "", errors.New("BASEURL: host expected")
		}
		if u.RawQuery != "" {
			return "", "", errors.New("BASEURL: query string now allowed")
		}
		if u.Fragment != "" {
			return "", "", errors.New("BASEURL: fragment not allowed")
		}
		prefix = u.Path
		u.Path = ""
		return u.String(), prefix, nil
	}

	host, port, err := net.SplitHostPort(bindaddr)
	if err != nil {
		return "", "", err
	}
	if host == "" || host == "0.0.0.0" {
		host, err = os.Hostname()
		if err != nil {
			return "", "", err
		}
	}

	return "http://" + net.JoinHostPort(host, port), prefix, nil
}
