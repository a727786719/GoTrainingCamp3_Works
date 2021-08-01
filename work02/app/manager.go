package app

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"sync"
)

type Manager struct {
	rootCtx context.Context
	ctx     context.Context
	cancel  context.CancelFunc
	group   *errgroup.Group
	signals []os.Signal
	apps    map[string]App
	lock    *sync.RWMutex
}

func NewManager() *Manager {
	return NewManagerWithContext(context.Background())
}

func NewManagerWithContext(parent context.Context) *Manager {
	rootCtx, cancel := context.WithCancel(parent)
	group, ctx := errgroup.WithContext(rootCtx)
	return &Manager{
		rootCtx: rootCtx,
		ctx:     ctx,
		cancel:  cancel,
		group:   group,
		lock:    &sync.RWMutex{},
		apps:    map[string]App{},
	}
}

// Run 启动新应用
func (mgr *Manager) Run(app App) {
	mgr.group.Go(func() error {
		mgr.lock.Lock()
		name := app.Name()
		log.Printf("start app %s.%s", name, app.Version())
		if _, ok := mgr.apps[name]; ok {
			mgr.lock.Unlock()
			return fmt.Errorf("start app %s.%s, already exist", name, app.Version())
		}
		mgr.apps[name] = app
		mgr.lock.Unlock()
		err := app.Run(mgr.ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

// List 列出运行中的全部应用信息
func (mgr Manager) List() []App {
	if len(mgr.apps) == 0 {
		return []App{}
	}

	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	apps := make([]App, 0, len(mgr.apps))
	for _, app := range mgr.apps {
		apps = append(apps, app)
	}

	return apps
}

// Signals 等待停止的信号
func (mgr *Manager) Signals(signals ...os.Signal) {
	mgr.signals = signals
}

// Stop 停止指定应用
func (mgr *Manager) Stop(appName string) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	app, ok := mgr.apps[appName]
	if !ok {
		return fmt.Errorf("app %s.%s not exist", appName, app.Version())
	}
	delete(mgr.apps, appName)
	return app.Stop()
}

func (mgr *Manager) StopAll() {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	if mgr.cancel != nil {
		mgr.cancel()
	}

	mgr.apps = map[string]App{}
}

func (mgr Manager) Wait() error {
	if len(mgr.signals) != 0 {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, mgr.signals...)
		go func() {
			select {
			case sig := <-ch:
				log.Printf("receive signal %+v", sig)
				if mgr.cancel != nil {
					mgr.cancel()
				}
			case <-mgr.ctx.Done():
				log.Printf("context closed")
				close(ch)
				break
			}
		}()
	}
	return mgr.group.Wait()
}
