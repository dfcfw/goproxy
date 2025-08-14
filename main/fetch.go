package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

type diskFetch struct {
	moddir string
	log    *slog.Logger
}

func (df *diskFetch) Query(ctx context.Context, modpath, query string) (version string, mtime time.Time, err error) {
	df.log.Info("查询 info 信息", "path", modpath, "query", query)
	return "", time.Time{}, errors.ErrUnsupported
}

func (df *diskFetch) List(ctx context.Context, modpath string) ([]string, error) {
	epath, err := module.EscapePath(modpath)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(df.moddir, epath, "@v", "list")
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	versions := make([]string, 0, 20)
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		line := scan.Text()
		if semver.IsValid(line) {
			versions = append(versions, line)
		}
	}
	if len(versions) == 0 {
		return nil, errors.ErrUnsupported
	}
	semver.Sort(versions)
	df.log.Info("获取列表", "path", modpath)

	return versions, nil
}

func (df *diskFetch) Download(ctx context.Context, modpath, version string) (info, mod, zip io.ReadSeekCloser, err error) {
	epath, err := module.EscapePath(modpath)
	if err != nil {
		return nil, nil, nil, err
	}

	defer func() {
		if err != nil {
			if info != nil {
				_ = info.Close()
			}
			if mod != nil {
				_ = mod.Close()
			}
			if zip != nil {
				_ = mod.Close()
			}
		}
	}()

	dir := filepath.Join(df.moddir, epath, "@v", version+".info")
	if info, err = os.Open(dir); err != nil {
		return nil, nil, nil, err
	}
	dir = filepath.Join(df.moddir, epath, "@v", version+".mod")
	if mod, err = os.Open(dir); err != nil {
		return nil, nil, nil, err
	}
	dir = filepath.Join(df.moddir, epath, "@v", version+".zip")
	if zip, err = os.Open(dir); err != nil {
		return nil, nil, nil, err
	}

	return
}
