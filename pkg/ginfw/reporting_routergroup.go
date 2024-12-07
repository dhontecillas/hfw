package ginfw

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/gin-gonic/gin"
)

type ReportingRouterGroup struct {
	ins     *obs.Insighter
	wrapped *gin.RouterGroup
}

func NewGroup(r *gin.RouterGroup, ins *obs.Insighter) *ReportingRouterGroup {
	return &ReportingRouterGroup{
		wrapped: r,
		ins:     ins,
	}
}

func (r *ReportingRouterGroup) report(method string, route string) {
	var fullRoute string
	if strings.HasPrefix(route, "/") {
		fullRoute = r.wrapped.BasePath() + route
	} else {
		fullRoute = r.wrapped.BasePath() + "/" + route
	}
	r.ins.L.Info("route registered", map[string]interface{}{
		"method": method,
		"route":  fullRoute,
	})
}

func (r *ReportingRouterGroup) Use(hfs ...gin.HandlerFunc) gin.IRoutes {
	r.ins.L.Info(fmt.Sprintf("Using Middleware %#v", hfs), nil)
	return r.wrapped.Use(hfs...)
}

func (r *ReportingRouterGroup) Handle(method string, route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report(method, route)
	return r.wrapped.Handle(method, route, hfs...)
}

func (r *ReportingRouterGroup) Any(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("ANY", route)
	return r.wrapped.Any(route, hfs...)
}

func (r *ReportingRouterGroup) GET(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("GET", route)
	return r.wrapped.GET(route, hfs...)
}

func (r *ReportingRouterGroup) POST(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("POST", route)
	return r.wrapped.POST(route, hfs...)
}

func (r *ReportingRouterGroup) DELETE(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("DELETE", route)
	return r.wrapped.DELETE(route, hfs...)
}

func (r *ReportingRouterGroup) PATCH(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("PATCH", route)
	return r.wrapped.PATCH(route, hfs...)
}

func (r *ReportingRouterGroup) PUT(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("PUT", route)
	return r.wrapped.PUT(route, hfs...)
}

func (r *ReportingRouterGroup) OPTIONS(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("OPTIONS", route)
	return r.wrapped.OPTIONS(route, hfs...)
}

func (r *ReportingRouterGroup) HEAD(route string, hfs ...gin.HandlerFunc) gin.IRoutes {
	r.report("HEAD", route)
	return r.wrapped.HEAD(route, hfs...)

}

func (r *ReportingRouterGroup) Match(methods []string, relativePaths string, hfs ...gin.HandlerFunc) gin.IRoutes {
	return r.wrapped.Match(methods, relativePaths, hfs...)
}

func (r *ReportingRouterGroup) StaticFile(relativePath string, filePath string) gin.IRoutes {
	return r.wrapped.StaticFile(relativePath, filePath)
}

func (r *ReportingRouterGroup) StaticFileFS(relativePath string, filePath string, fs http.FileSystem) gin.IRoutes {
	return r.wrapped.StaticFileFS(relativePath, filePath, fs)
}

func (r *ReportingRouterGroup) Static(relativePath string, root string) gin.IRoutes {
	return r.wrapped.Static(relativePath, root)
}

func (r *ReportingRouterGroup) StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	return r.wrapped.StaticFS(relativePath, fs)
}

func (r *ReportingRouterGroup) Group(route string, hfs ...gin.HandlerFunc) *gin.RouterGroup {
	return r.wrapped.Group(route, hfs...)
}
