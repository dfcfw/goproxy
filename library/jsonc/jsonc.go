// MIT License

// Copyright (c) 2019 Muhammad Muzzammil

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package jsonc

import (
	"encoding/json"
	"io"
	"os"
)

// Unmarshal parses the JSONC-encoded data and stores the result in the value pointed to by v.
// Equivalent of calling `json.Unmarshal(jsonc.ToJSON(data), v)`
func Unmarshal(data []byte, v any) error {
	j := Translate(data)
	return json.Unmarshal(j, v)
}

// ReadFile 读取文件内容并反序列化。
func ReadFile(fp string, v any, limit ...int64) error {
	fd, err := os.Open(fp)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fd.Close()

	rd := io.Reader(fd)
	if len(limit) > 0 && limit[0] > 0 {
		rd = io.LimitReader(rd, limit[0])
	}

	raw, err := io.ReadAll(rd)
	if err != nil {
		return err
	}

	return Unmarshal(raw, v)
}
