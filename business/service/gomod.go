package service

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dfcfw/goproxy/contract/response"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	modzip "golang.org/x/mod/zip"
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

func (gmd *Gomod) Walk(_ context.Context, rawpath string) (*response.GomodWalk, error) {
	var escpath string
	if rawpath != "" {
		escaped, err := module.EscapePath(rawpath)
		if err != nil {
			return nil, err
		}
		escpath = escaped
	}

	dir := filepath.Join(gmd.dir, escpath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var hasmod bool
	ret := new(response.GomodWalk)
	for _, ent := range entries {
		name := ent.Name()
		if name == "@v" && ent.IsDir() {
			hasmod = true
			continue
		}

		epath := path.Join(escpath, name)
		rpath, err := module.UnescapePath(epath)
		if err != nil {
			continue
		}
		if err = module.CheckPath(rpath); err != nil {
			continue
		}

		rname := path.Base(rpath)

		gmp := &response.GomodPath{Path: rpath, Name: rname}
		ret.Paths = append(ret.Paths, gmp)
	}
	if !hasmod {
		return ret, nil
	}

	flist := filepath.Join(gmd.dir, escpath, "@v", "list")
	file, err := os.Open(flist)
	if err != nil {
		return ret, nil
	}
	defer file.Close()

	buf := bufio.NewScanner(file)
	for buf.Scan() {
		line := buf.Text()
		if !semver.IsValid(line) {
			continue
		}
		version, err := module.UnescapeVersion(line)
		if err != nil {
			continue
		}
		gmv := &response.GomodModule{Version: version}
		ret.Modules = append(ret.Modules, gmv)
	}

	return ret, nil
}

func (gmd *Gomod) Stat(_ context.Context, modpath, version string) (response.GomodFiles, error) {
	var err error
	if modpath, err = module.EscapePath(modpath); err != nil {
		return nil, err
	}
	if version, err = module.EscapeVersion(version); err != nil {
		return nil, err
	}

	fpath := filepath.Join(gmd.dir, modpath, "@v")
	entries, err := os.ReadDir(fpath)
	if err != nil {
		return nil, err
	}

	ret := make(response.GomodFiles, 0, 10)
	for _, ent := range entries {
		ename := ent.Name()
		if ent.IsDir() {
			continue
		}
		if !strings.HasPrefix(ename, version) {
			continue
		}
		ext := path.Ext(ename)
		if ext == "" {
			continue
		}
		if ename != version+ext {
			continue
		}

		gmf := &response.GomodFile{Name: ename}
		if inf, _ := ent.Info(); inf != nil {
			gmf.Size = inf.Size()
			gmf.Mode = inf.Mode().String()
			gmf.ModifiedAt = inf.ModTime()
		}

		ret = append(ret, gmf)
	}

	return ret, nil
}

func (gmd *Gomod) Sniff(mf multipart.File, size int64) (*response.GomodSniff, error) {
	zr, err := zip.NewReader(mf, size)
	if err != nil {
		return nil, err
	}

	ret := new(response.GomodSniff)
	for _, zf := range zr.File {
		name := zf.Name
		if !strings.Contains(name, "@v") {
			continue
		}
		before, after, _ := strings.Cut(name, "@")
		after, _, _ = strings.Cut(after, "/")
		if module.CheckPath(before) == nil &&
			semver.IsValid(after) {
			ret.Path = before
			ret.Version = after
			break
		}
	}

	return ret, nil
}

func (gmd *Gomod) Format(w io.Writer, zr *zip.Reader, modpath, version string) error {
	var err error
	rawpath := modpath
	if modpath, err = module.EscapePath(modpath); err != nil {
		return err
	}
	if version, err = module.EscapeVersion(version); err != nil {
		return err
	}

	var files []modzip.File
	for _, zf := range zr.File {
		files = append(files, &zipFile{f: zf})
		if zf.Name == "go.mod" {
			modf, err := zf.Open()
			if err != nil {
				continue
			}
			buf, _ := io.ReadAll(modf)
			pf, err := modfile.Parse(zf.Name, buf, nil)
			if err != nil {
				continue
			}
			m := pf.Module
			if m == nil {
				continue
			}
			if mp := m.Mod.Path; mp != rawpath {
				return fmt.Errorf("模块名不匹配 (输入为 %s, 检测到 %s)", rawpath, mp)
			}
		}
	}
	mdv := module.Version{Path: modpath, Version: version}

	return modzip.Create(w, mdv, files)
}

//goland:noinspection GoUnhandledErrorResult
func (gmd *Gomod) Upload(mf multipart.File, modpath, version string) error {
	now := time.Now()
	minf := &moduleInfo{Version: version, Time: now}
	mdv := module.Version{Path: modpath, Version: version}
	var err error
	if modpath, err = module.EscapePath(modpath); err != nil {
		return err
	}
	if version, err = module.EscapeVersion(version); err != nil {
		return err
	}

	dir := filepath.Join(gmd.dir, modpath, "@v")
	if _, exx := os.Stat(dir); exx != nil {
		if !os.IsNotExist(exx) {
			return exx
		}
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	temp, err := os.CreateTemp(os.TempDir(), "gomod_*.zip")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	_, err = io.Copy(temp, mf)
	_ = temp.Close()
	defer os.Remove(tempName)
	if err != nil {
		return err
	}
	if _, err = modzip.CheckZip(mdv, tempName); err != nil {
		return errors.New("不是合法的 go 模块")
	}

	// 读取 go.mod 文件
	zr, err := zip.OpenReader(tempName)
	if err != nil {
		return err
	}
	defer zr.Close()

	var modok, mdok bool
	modbuf := new(bytes.Buffer)
	mdbuf := new(bytes.Buffer)

	zdir := path.Join(mdv.Path + "@" + mdv.Version)
	gomodName := path.Join(zdir, "go.mod")
	for _, zf := range zr.File {
		if zf.Name == gomodName && !modok {
			if buf := dumpZip(zf); len(buf) != 0 {
				modok = true
				modbuf.Write(buf)
			}
		}

		lower := strings.ToLower(strings.TrimPrefix(zf.Name, zdir+"/"))
		fmt.Println(lower)
		if lower == "readme.md" || lower == "readme.markdown" {
			if buf := dumpZip(zf); len(buf) != 0 {
				mdok = true
				mdbuf.Write(buf)
			}
		}
	}
	if !modok {
		modbuf.Reset()
		modbuf.WriteString("module " + mdv.Path + "\n")
	}

	{
		// 存放 .zip
		zfile, err := os.OpenFile(filepath.Join(dir, version+".zip"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		tfile, err := os.Open(tempName)
		if err != nil {
			_ = zfile.Close()
			return err
		}
		_, err = io.Copy(zfile, tfile)
		_ = tfile.Close()
		_ = zfile.Close()
		if err != nil {
			return err
		}
	}
	{
		// 提取 go.mod
		mfile, err := os.OpenFile(filepath.Join(dir, version+".mod"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		_, err = io.Copy(mfile, modbuf)
		_ = mfile.Close()
		if err != nil {
			return err
		}
	}
	if mdok && mdbuf.Len() != 0 {
		// 提取 go.mod
		mfile, err := os.OpenFile(filepath.Join(dir, version+".markdown"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err == nil {
			_, _ = io.Copy(mfile, mdbuf)
			_ = mfile.Close()
		}
	}
	{
		// 提取 go.mod
		mfile, err := os.OpenFile(filepath.Join(dir, version+".mod"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		_, err = io.Copy(mfile, modbuf)
		_ = mfile.Close()
		if err != nil {
			return err
		}
	}
	{
		ifile, err := os.OpenFile(filepath.Join(dir, version+".info"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		err = json.NewEncoder(ifile).Encode(minf)
		_ = ifile.Close()
		if err != nil {
			return err
		}
	}
	{
		// 更新 list 文件
		index := make(map[string]struct{}, 16)
		var versions []string
		fstr := filepath.Join(dir, "list")
		if inf, exx := os.Stat(fstr); exx == nil && !inf.IsDir() {
			old, err := os.Open(fstr)
			if err != nil {
				return err
			}
			sc := bufio.NewScanner(old)
			for sc.Scan() {
				line := sc.Text()
				if _, exists := index[line]; exists {
					continue
				}
				index[line] = struct{}{}
				versions = append(versions, line)
			}
			_ = old.Close()
		}
		if _, exists := index[mdv.Version]; !exists {
			versions = append(versions, mdv.Version)
		}
		semver.Sort(versions)
		fd, err := os.OpenFile(fstr, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		for _, ver := range versions {
			fd.WriteString(ver + "\n")
		}
		_ = fd.Close()
	}

	return nil
}

func (gmd *Gomod) Open(rawpath string, filename string) (*os.File, error) {
	escpath, err := module.EscapePath(rawpath)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(filename)
	if ext == "" {
		return nil, os.ErrNotExist
	}
	suffix := strings.TrimSuffix(filename, ext)
	if _, err := module.UnescapeVersion(suffix); err != nil {
		return nil, err
	}
	fpath := filepath.Join(gmd.dir, escpath, "@v", filename)

	return os.Open(fpath)
}

type moduleInfo struct {
	Version string    `json:",omitempty"`
	Time    time.Time `json:",omitempty"`
}

type zipFile struct {
	f *zip.File
}

func (z *zipFile) Path() string {
	return strings.TrimRight(z.f.Name, "/")
}

func (z *zipFile) Lstat() (os.FileInfo, error) {
	inf := z.f.FileInfo()
	return inf, nil
}

func (z *zipFile) Open() (io.ReadCloser, error) {
	return z.f.Open()
}

func dumpZip(f *zip.File) []byte {
	zf, err := f.Open()
	if err != nil {
		return nil
	}

	buf, _ := io.ReadAll(zf)
	_ = zf.Close()

	return buf
}
