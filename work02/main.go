package main

import (
	"GoTrainingCamp3_Works/work02/app"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

type httpApp struct {
	name   string
	addr   string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHttpApp(name, addr string) *httpApp {
	return &httpApp{name: name, addr: addr}
}

func (a httpApp) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	_, _ = response.Write([]byte(fmt.Sprintf("hello, %s.%s", a.Name(), a.Version())))
}

func (a httpApp) Name() string {
	return a.name
}

func (a httpApp) Version() string {
	return "v1"
}

func (a *httpApp) Run(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)
	server := &http.Server{Addr: a.addr, Handler: a}
	go func() {
		select {
		case <-a.ctx.Done():
			err := server.Close()
			if err != nil {
				log.Printf("close %s.%s failed, %+v", a.Name(), a.Version(), err)
			}
		}
	}()
	return server.ListenAndServe()
}

func (a httpApp) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

func main() {
	mgr := app.NewManager()
	mgr.Signals(os.Interrupt, os.Kill)
	mgr.Run(NewHttpApp("app01", ":28001"))
	mgr.Run(NewHttpApp("app02", ":28002"))
	if err := mgr.Wait(); err != nil {
		log.Printf("apps exit, %+v", err)
	}
}
