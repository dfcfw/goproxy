package service

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/mod/module"
)

type Gomod struct {
	dir string
	log *slog.Logger
}

func NewGomod(dir string, log *slog.Logger) *Gomod {
	return &Gomod{
		dir: dir,
		log: log,
	}
}

func (gmd *Gomod) Browse(ctx context.Context, node string) error {
	if node != "" {
		escaped, err := module.EscapePath(node)
		if err != nil {
			return err
		}
		node = escaped
	}

	dir := filepath.Join(gmd.dir, node)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var mod bool
	for _, ent := range entries {
		name := ent.Name()
		if name == "@v" && ent.IsDir() {
			mod = true
			continue
		}

		enode := path.Join(node, name)
		rnode, err := module.UnescapePath(enode)
		if err != nil {
			continue
		}

		if err = module.CheckPath(rnode); err == nil {
			gmd.log.Info(rnode)
		} else {
			gmd.log.Error("BAD " + rnode)
		}
	}
	if mod {
	}

	return nil
}

func (gmd *Gomod) View(ctx context.Context, node, version string) error {
	return nil
}

func (gmd *Gomod) File(ctx context.Context, node, name string) (io.ReadCloser, error) {
	if err := module.CheckPath(node); err == nil {
		return nil, err
	}

	escaped, err := module.EscapePath(node)
	if err != nil {
		return nil, err
	}

	fp := filepath.Join(gmd.dir, escaped, "@v", name)

	return os.Open(fp)
}

type Node struct {
	Name     string
	FullName string
}

type Mod struct {
	Version string
}
