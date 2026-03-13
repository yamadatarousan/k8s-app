package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestLoadConfig_環境変数から設定を読む(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("TODO_API_NAME", "todo-api-test")
	t.Setenv("SHUTDOWN_TIMEOUT_SECONDS", "7")

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("設定読み込みに失敗: %v", err)
	}

	if config.Port != "9090" {
		t.Fatalf("期待したポートは9090、実際は%q", config.Port)
	}
	if config.ApplicationName != "todo-api-test" {
		t.Fatalf("期待したアプリ名と異なる: %q", config.ApplicationName)
	}
	if config.ShutdownTimeout != 7*time.Second {
		t.Fatalf("期待した停止猶予は7秒、実際は%v", config.ShutdownTimeout)
	}
}

func TestWaitForShutdownSignal_SIGTERMでShutdownを呼ぶ(t *testing.T) {
	t.Parallel()

	signalContext, stop := signalContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownStarted := make(chan struct{}, 1)

	go func() {
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()

	err := waitForShutdownSignal(signalContext, func(context.Context) error {
		shutdownStarted <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("停止処理が失敗: %v", err)
	}

	select {
	case <-shutdownStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("SIGTERM受信後にShutdownが呼ばれなかった")
	}
}
