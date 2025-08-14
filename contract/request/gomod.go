package request

import "mime/multipart"

type GomodWalk struct {
	Path string `json:"path" query:"path" validate:"omitempty"`
}

type GomodStat struct {
	Path    string `json:"path,omitzero"    query:"path"    validate:"required"`
	Version string `json:"version,omitzero" query:"version" validate:"required"`
}

type GomodSniff struct {
	File *multipart.FileHeader `json:"file" form:"file" validate:"required"`
}

type GomodUpload struct {
	File    *multipart.FileHeader `json:"file"    form:"file"    validate:"required"`
	Path    string                `json:"path"    form:"path"    validate:"required"`
	Version string                `json:"version" form:"version" validate:"required"`
}

type GomodFile struct {
	Path string `json:"path" query:"path" validate:"required"`
	Name string `json:"name" query:"name" validate:"required"`
}
