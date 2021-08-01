package main

import (
	"GoTrainingCamp3_Works/work02/app"
	"log"
	"os"
	"testing"
	"time"
)

// TestApp01 两个应用启动成功
func TestApp01(t *testing.T) {
	mgr := app.NewManager()
	mgr.Signals(os.Interrupt, os.Kill)
	mgr.Run(NewHttpApp("app01", ":28001"))
	mgr.Run(NewHttpApp("app02", ":28002"))
	if err := mgr.Wait(); err != nil {
		log.Printf("apps exit, %+v", err)
	}
}

// TestApp02 两个应用端口相同, 启动失败
func TestApp02(t *testing.T) {
	mgr := app.NewManager()
	mgr.Signals(os.Interrupt, os.Kill)
	mgr.Run(NewHttpApp("app01", ":28001"))
	mgr.Run(NewHttpApp("app02", ":28001"))
	if err := mgr.Wait(); err != nil {
		log.Printf("apps exit, %+v", err)
	}
}

// TestApp03 10s 后关闭所有应用
func TestApp03(t *testing.T) {
	mgr := app.NewManager()
	mgr.Signals(os.Interrupt, os.Kill)
	mgr.Run(NewHttpApp("app01", ":28001"))
	mgr.Run(NewHttpApp("app02", ":28002"))

	go func() {
		time.Sleep(time.Second * 10)
		mgr.StopAll()
	}()

	if err := mgr.Wait(); err != nil {
		t.Fatalf("apps exit, %+v", err)
	}
}
