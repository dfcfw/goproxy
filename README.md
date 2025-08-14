<!-- TOC -->
* [goproxy protocol](#goproxy-protocol)
  * [协议分析](#协议分析)
    * [获取版本列表 `$base/$module/@v/list`](#获取版本列表-basemodulevlist)
      * [响应结果（为了防止文档太长影响阅读，中间的响应数据被 ... 代替）:](#响应结果为了防止文档太长影响阅读中间的响应数据被--代替)
    * [获取特定版本信息 `$base/$module/@v/$version.info`](#获取特定版本信息-basemodulevversioninfo)
    * [获取特定版本的 go.mod 文件 `$base/$module/@v/$version.mod`](#获取特定版本的-gomod-文件-basemodulevversionmod)
    * [获取特定版本的源码 zip 文件 `$base/$module/@v/$version.zip`](#获取特定版本的源码-zip-文件-basemodulevversionzip)
    * [【可选】获取最新版本 `$base/$module/@latest`](#可选获取最新版本-basemodulelatest)
<!-- TOC -->

# goproxy protocol

## 协议分析

依照 [goproxy protocol](https://go.dev/ref/mod#goproxy-protocol)

- `$base`: 以公开的 https://goproxy.cn 为例。

- `$module`: 以 github.com/labstack/echo/v4 为例。

### 获取版本列表 `$base/$module/@v/list`

GET https://goproxy.cn/github.com/labstack/echo/v4/@v/list

#### 响应结果（为了防止文档太长影响阅读，中间的响应数据被 ... 代替）:

```text
v4.0.0
v4.1.0
v4.1.1
v4.1.2
v4.1.3
...
...
...
v4.13.0
v4.13.1
v4.13.2
v4.13.3
v4.13.4
```

### 获取特定版本信息 `$base/$module/@v/$version.info`

GET https://goproxy.cn/github.com/labstack/echo/v4/@v/v4.13.4.info

响应结果：

```json
{
  "Version": "v4.13.4",
  "Time": "2025-05-22T11:18:29Z"
}
```

### 获取特定版本的 go.mod 文件 `$base/$module/@v/$version.mod`

GET https://goproxy.cn/github.com/labstack/echo/v4/@v/v4.13.4.mod

响应结果：

```text
module github.com/labstack/echo/v4

go 1.23.0

require (
	github.com/labstack/gommon v0.4.2
	github.com/stretchr/testify v1.10.0
	github.com/valyala/fasttemplate v1.2.2
	golang.org/x/crypto v0.38.0
	golang.org/x/net v0.40.0
	golang.org/x/time v0.11.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

### 获取特定版本的源码 zip 文件 `$base/$module/@v/$version.zip`

GET https://goproxy.cn/github.com/labstack/echo/v4/@v/v4.13.4.zip

响应结果即为文件下载流，此处不作展示。

### 【可选】获取最新版本 `$base/$module/@latest`

GET https://goproxy.cn/github.com/labstack/echo/v4/@latest

响应结果：

```json
{
  "Version": "v4.13.4",
  "Time": "2025-05-22T11:18:29Z"
}
```
