package response

import "time"

type GomodWalk struct {
	Paths   GomodPaths   `json:"paths,omitzero"`
	Modules GomodModules `json:"modules,omitzero"`
}

type GomodPath struct {
	Name string `json:"name,omitzero"`
	Path string `json:"path,omitzero"`
}

type GomodPaths []*GomodPath

type GomodModule struct {
	Version string `json:"version,omitzero"`
}

type GomodModules []*GomodModule

type GomodFile struct {
	Name       string    `json:"name,omitzero"`
	Mode       string    `json:"mode,omitzero"`
	Size       int64     `json:"size,omitzero"`
	ModifiedAt time.Time `json:"modified_at,omitzero"`
}

type GomodFiles []*GomodFile

type GomodSniff struct {
	Path    string `json:"path,omitzero"`
	Version string `json:"version,omitzero"`
}
