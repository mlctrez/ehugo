package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/mlctrez/ehugo/hueapi"
	"github.com/mlctrez/ehugo/ssdp"
	"github.com/mlctrez/servicego"
	"go.etcd.io/bbolt"
	"net/http"
	"os"
	"time"
)

var _ servicego.Service = (*svc)(nil)

type svc struct {
	servicego.Defaults
	addr       string
	ssdpServer *ssdp.SSDP
	apiServer  *http.Server
	hueApi     *hueapi.HueApi
	boltDb     *bbolt.DB
}

func New() servicego.Service {
	return &svc{}
}

func (g *svc) Start(s service.Service) (err error) {
	g.Infof("starting")
	defer g.Infof("started")

	options := &bbolt.Options{
		Timeout:      time.Second * 5,
		NoGrowSync:   false,
		FreelistType: bbolt.FreelistArrayType,
	}
	if g.boltDb, err = bbolt.Open("database.db", 0600, options); err != nil {
		return err
	}

	g.addr = os.Getenv("ADDRESS")
	if g.addr == "" {
		return fmt.Errorf("ADDRESS environment variable not set")
	}
	bridgeOne := &ssdp.BridgeInfo{
		SerialNumber: "001788FFFE23BFC1",
		UUID:         "2f402f80-da50-11e1-9b23-001788255acc",
	}
	g.hueApi = hueapi.New(g, g.boltDb, g.addr, bridgeOne)
	if err = g.hueApi.SetupBolt(); err != nil {
		return err
	}

	g.apiServer = &http.Server{Addr: g.addr, Handler: g.hueApi.Handler()}
	go g.serveHttp()

	g.ssdpServer = ssdp.New(ssdp.WithCallback(g.hueApi.SSDPCallback))
	if err = g.ssdpServer.Listen(); err != nil {
		return err
	}
	go g.ssdpServer.Read()

	return nil
}

func (g *svc) Stop(s service.Service) error {
	g.Infof("stopping")
	defer g.Infof("stopped")
	if g.apiServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := g.apiServer.Shutdown(ctx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				g.Errorf("api server shutdown error: %v", err)
			}
		}
	}
	if g.ssdpServer != nil {
		g.ssdpServer.Shutdown()
	}
	if g.boltDb != nil {
		if err := g.boltDb.Close(); err != nil {
			g.Errorf("error closing database: %v", err)
		}
	}
	return nil
}

func (g *svc) serveHttp() {
	if err := g.apiServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			g.Errorf("apiServer.ListenAndServe() error: %s", err)
		}
	}
}
