# Module proxies 模块代理（GOPROXY 协议）

> AI 翻译，人工校正。

`module proxy` 是一种 HTTP 服务器，可以响应下面指定的路径的 GET 请求。这些请求没有查询参数，也不需要特定的标头，因此即使是从固定文件系统（包括
file:// URL）提供服务的站点也可以是模块代理。

成功的 HTTP 响应必须具有状态代码 200 （OK） 。遵循重定向 （3xx）。状态代码为 4xx 和 5xx 的响应被视为错误。错误代码 404 （Not
Found） 和 410 （Gone） 表示请求的模块或版本在代理上不可用，但可以在其他地方找到。错误响应的内容类型应为 text/plain 字符集
UTF-8 或 US-ASCII。

go 命令可以配置为使用 GOPROXY 环境变量联系代理或源代码控制服务器，该环境变量接受代理 URL 列表。该列表可能包括关键字 direct
或 off（有关详细信息，请参阅[环境变量](https://go.dev/ref/mod#environment-variables)）。列表元素可以用逗号 （，） 或管道 （|）
分隔，它们决定错误回退行为。当 URL 后跟逗号时，go 命令仅在 404（未找到）或 410（已消失）响应后回退到后续源。当 URL 后跟管道时，go
命令会在任何错误（包括超时等非 HTTP 错误）后回退到以后的源。这种错误处理行为允许代理充当未知模块的网守。例如，对于不在批准列表中的模块，代理可能会响应错误
403（禁止）（请参阅为[专用模块提供服务的专用代理](https://go.dev/ref/mod#private-module-proxy-private)）。

下表指定了模块代理必须响应的查询。对于每个路径，\$base 是代理 URL 的路径部分，\$module 是模块路径，并且 \$version
是一个版本。例如，如果代理 URL 是 https://example.com/mod，并且客户端正在请求 v0.3.2 版本的模块 golang.org/x/text 的 go.mod
文件，则客户端将发送一个 GET 请求 https://example.com/mod/golang.org/x/text/@v/v0.3.2.mod 。

为了避免从不区分大小写的文件系统提供服务时出现歧义，\$module 和 \$version 元素采用大小写编码，方法是将每个大写字母替换为感叹号，后跟相应的小写字母。这允许模块
example.com/M 和 example.com/m 都存储在磁盘上，因为前者编码为 example.com/!m。

| Path （路径）                        | Description （描述）                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `$base/$module/@v/list`          | 返回指定模块已知版本的纯文本列表，每行一个版本。该列表不应包含伪版本（pseudo-version）。                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `$base/$module/@v/$version.info` | 返回特定模块版本的 JSON 格式元数据。响应必须是一个与下面 Go 数据结构对应的 JSON 对象：<br/> `type Info struct {`<br>&nbsp;&nbsp;`Version string    // 版本号`<br>&nbsp;&nbsp;`Time    time.Time // 提交时间`<br>`}`<br/><br/> Version 字段是必填项，必须包含一个有效的、[规范化的版本](https://go.dev/ref/mod#glos-canonical-version)（参见[Versions](https://go.dev/ref/mod#versions)）。请求路径中的 \$version 不需要与 Version 字段的值相同，甚至不需要是一个有效版本；此端点可以用于通过分支名或修订标识符来查找版本。但是，如果 $version 是一个与 \$module 主版本兼容的规范化版本，则成功响应中的 Version 字段必须与其相同。<br/><br/>Time 字段是可选的。如果存在，必须是 RFC 3339 格式的字符串，表示该版本的创建时间。<br/><br/>将来可能会添加更多字段，因此其他字段名已被保留。 |
| `$base/$module/@v/$version.mod`  | 返回特定模块版本的 go.mod 文件。如果该模块在请求的版本中没有 go.mod 文件，则必须返回一个仅包含 module 声明且模块路径为请求的模块路径的文件。否则，必须返回原始且未修改的 go.mod 文件。                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `$base/$module/@v/$version.zip`  | 返回包含特定模块版本内容的 ZIP 文件。有关此 ZIP 文件的格式要求，请参阅 [Module zip files](https://go.dev/ref/mod#zip-files) 详细说明。                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `$base/$module/@latest`          | 返回指定模块最新已知版本的 JSON 格式元数据，格式与 \$base/\$module/@v/\$version.info 相同。最新版本应是当 \$base/\$module/@v/list 为空或列表中没有合适版本时，go 命令应使用的模块版本。 此端点是可选的，模块代理不要求必须实现。                                                                                                                                                                                                                                                                                                                                                                                                          |

在解析模块的最新版本时，go 命令会先请求 \$base/\$module/@v/list，如果没有找到合适的版本，再请求 \$base/\$module/@latest。
go 命令的选择优先级如下：

- 语义版本号中最高的正式发布版本（release version）；

- 语义版本号中最高的预发布版本（pre-release version）；

- 按时间顺序最新的伪版本（pseudo-version）。

在 Go 1.12 及更早版本中，go 命令会将 \$base/\$module/@v/list 中的伪版本视为预发布版本，但从 Go 1.13 开始，这种情况不再成立。

模块代理必须保证对 \$base/\$module/\$version.mod 和 \$base/\$module/\$version.zip 请求的成功响应内容完全一致。
这些内容通过 [go.sum 文件](https://go.dev/ref/mod#go-sum-files)
及默认的 [checksum 数据库](https://go.dev/ref/mod#checksum-database)
进行 [加密认证](https://go.dev/ref/mod#authenticating)。

go 命令会将从模块代理下载的大部分内容缓存到模块缓存目录 \$GOPATH/pkg/mod/cache/download 中。即使是直接从版本控制系统下载，go
命令也会生成对应的 info、mod 和 zip 文件并存储到该目录，行为与直接从代理下载完全一致。缓存目录的结构与代理 URL 空间相同，因此将
\$GOPATH/pkg/mod/cache/download 目录提供为静态服务（或复制到如 https://example.com/proxy ）后，用户通过设置
GOPROXY=https://example.com/proxy 即可访问缓存的模块版本。
