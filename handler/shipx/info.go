package shipx

type Infer interface {
	// Name 路由名字。
	Name() string

	// Perm 访问该路由所需的权限。
	Perm() Permission
}

type Permission struct {
	// Anonymous 任何人都可以访问（无需登录认证）。
	Anonymous bool

	// UsePAT 是否使用 PAT 认证。
	UsePAT bool

	// Logon 任何已登录用户均可访问。
	Logon bool
}

var RouteInfoKey = routeInfoKey{}

type routeInfoKey struct{}

func (routeInfoKey) String() string {
	return "ship-route-info-key"
}

func NewRouteInfo(name string) *RouteInfo {
	return &RouteInfo{name: name}
}

type RouteInfo struct {
	name      string
	anonymous bool
	usePAT    bool
	logon     bool
}

func (ri RouteInfo) Name() string {
	return ri.name
}

func (ri RouteInfo) Perm() Permission {
	return Permission{
		Anonymous: ri.anonymous,
		UsePAT:    ri.usePAT,
		Logon:     ri.logon,
	}
}

func (ri RouteInfo) UsePAT() RouteInfo {
	ri.usePAT = true
	ri.anonymous = false
	ri.logon = false

	return ri
}

func (ri RouteInfo) Anonymous() RouteInfo {
	ri.anonymous = true
	ri.usePAT = false
	ri.logon = false

	return ri
}

func (ri RouteInfo) Logon() RouteInfo {
	ri.logon = true
	ri.anonymous = false
	ri.usePAT = false

	return ri
}

func (ri RouteInfo) Map() map[any]any {
	return map[any]any{
		RouteInfoKey: ri,
	}
}

func DetectRouteInfo(dat any) Infer {
	if dat == nil {
		return nil
	}

	if inf, ok := dat.(Infer); ok {
		return inf
	}

	key := RouteInfoKey
	skey := key.String()
	if amp, ok := dat.(map[any]any); ok {
		if val, exists := amp[key]; exists {
			if inf, yes := val.(Infer); yes {
				return inf
			}
		}
		if val, exists := amp[skey]; exists {
			if inf, yes := val.(Infer); yes {
				return inf
			}
		}
	}
	if amp, ok := dat.(map[any]Infer); ok {
		if inf, exists := amp[key]; exists {
			return inf
		}
		if inf, exists := amp[skey]; exists {
			return inf
		}
	}
	if amp, ok := dat.(map[string]any); ok {
		if val, exists := amp[skey]; exists {
			if inf, yes := val.(Infer); yes {
				return inf
			}
		}
	}
	if amp, ok := dat.(map[string]Infer); ok {
		if inf, exists := amp[skey]; exists {
			return inf
		}
	}

	return nil
}
