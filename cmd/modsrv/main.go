package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/dfcfw/goproxy/launch"
)

func main() {
	args := os.Args
	name := filepath.Base(args[0])
	set := flag.NewFlagSet(name, flag.ExitOnError)
	cfg := set.String("c", "resources/config/application.jsonc", "配置文件")
	_ = set.Parse(args[1:])

	// https://github.com/golang/go/issues/67182
	for _, fp := range []string{"resources/.crash.txt", ".crash.txt"} {
		if f, _ := os.Create(fp); f != nil {
			_ = debug.SetCrashOutput(f, debug.CrashOptions{})
			_ = f.Close()
			break
		}
	}

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT}
	ctx, cancel := signal.NotifyContext(context.Background(), signals...)
	defer cancel()

	opt := &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	log := slog.New(slog.NewJSONHandler(os.Stdout, opt))

	if err := launch.Run(ctx, *cfg); err != nil {
		log.Error("服务运行错误", slog.Any("error", err))
	} else {
		log.Info("服务停止运行")
	}
}
