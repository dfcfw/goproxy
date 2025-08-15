package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"golang.org/x/mod/zip"
)

func main() {
	var output string
	var modpath string
	var version string
	var directory string

	fset := flag.NewFlagSet("modzip", flag.ExitOnError)
	fset.StringVar(&directory, "d", "", "源代码目录")
	fset.StringVar(&output, "o", "go.src.zip", "打包后的 zip 文件名")
	fset.StringVar(&modpath, "m", "", "模块名（不填写自动检测），例如：github.com/gin-gonic/gin")
	fset.StringVar(&version, "v", "", "版本号，例如：v1.2.3-beta")
	_ = fset.Parse(os.Args[1:])

	if directory == "" || version == "" {
		fset.PrintDefaults()
		return
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	lstat, err := os.Lstat(directory)
	if err != nil || !lstat.IsDir() {
		_, _ = fmt.Fprintf(os.Stderr, "源码目录无效: %v\n", err)
		os.Exit(1)
	}

	var detect bool
	if modpath == "" {
		modpath = detectGoModFile(directory)
		detect = true
	}
	if modpath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "请输入一个模块名")
		os.Exit(1)
	}
	if err = module.CheckPath(modpath); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "模块名不合法\n")
		os.Exit(1)
	}
	if detect {
		_, _ = fmt.Fprintf(os.Stdout, "检测到模块名: %s\n", modpath)
	}

	if !semver.IsValid(version) ||
		module.IsPseudoVersion(version) {
		_, _ = fmt.Fprintf(os.Stderr, "版本号不合法\n")
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(output))
	if ext != ".zip" {
		output += ".zip"
	}

	out, err := os.Create(output)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "创建输出文件错误: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	mv := module.Version{
		Path:    modpath,
		Version: version,
	}
	if err = zip.CreateFromDir(out, mv, directory); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "执行错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("执行成功")
}

func detectGoModFile(dir string) string {
	const name = "go.mod"
	fp := filepath.Join(dir, name)
	f, err := os.Open(fp)
	if err != nil {
		return ""
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, 1<<20))
	if err != nil {
		return ""
	}
	pf, err := modfile.Parse(name, raw, nil)
	if err != nil {
		return ""
	}

	if m := pf.Module; m != nil {
		return m.Mod.Path
	}

	return ""
}
