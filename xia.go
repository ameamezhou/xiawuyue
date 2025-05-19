package xiawuyue

import (
	"github.com/ameamezhou/xiawuyue/xconfig"
	"net/http"
	"path"
	"strings"

	"github.com/ameamezhou/xiawuyue/xlog"
)

type HandlerFunc func(c *Context)

type Xia struct {
	*RouterGroup
	addr   string
	router *router
	groups []*RouterGroup
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	xia         *Xia          // all groups share a Engine instance
}

type Handler struct {
	Method  string
	Prepare func(w http.ResponseWriter, r *http.Request) (err error)
	Do      func(w http.ResponseWriter, r *http.Request)
	URL     string
	CmdID   int
}

/*
type Handler struct {
        Method  string
        Prepare func(w http.ResponseWriter, r *http.Request) (err error)
        Do      func(w http.ResponseWriter, r *http.Request)
        URL     string
        CmdID   int
}

// HandlerMap 路由规则
var HandlerMap = map[string]Handler{
        "main": {
                Method: "GET",
                // 权限需要自己申请，详见 comm/wwauth/checkauth.go
                // 测试环境会跳过权限要求
                Prepare: nil,
                Do:      handler.GetPageHandler,
                URL:     "/wego/osswwlocalticketlogic/page",
                // Cmd ID 需要唯一
                CmdID: 2,
        },
}
*/

// func New() *Engine {
// 	engine := &Engine{router: newRouter()}
// 	engine.RouterGroup = &RouterGroup{engine: engine}
// 	engine.groups = []*RouterGroup{engine.RouterGroup}
// 	return engine
// }

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	xiaWuYue := group.xia
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		xia:    xiaWuYue,
	}
	xiaWuYue.groups = append(xiaWuYue.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRouter(method string, comp string, handler HandlerFunc) {
	if len(comp) > 0 && comp[0] != '/' {
		comp = "/" + comp
	}
	pattern := group.prefix + comp
	group.xia.router.addRouter(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRouter("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRouter("POST", pattern, handler)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		xlog.Debug(file)
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}

// New is the constructor of xia.Xia
func New() *Xia {
	xiaWuYue := &Xia{
		router: newRouter(),
	}
	xiaWuYue.RouterGroup = &RouterGroup{xia: xiaWuYue}
	xiaWuYue.groups = []*RouterGroup{xiaWuYue.RouterGroup}
	g := xiaWuYue.Group("/")
	g.Use(TimeLogger)
	return xiaWuYue
}

func (x *Xia) SET(method, pattern string, handler HandlerFunc) {
	x.router.addRouter(method, pattern, handler)
}

func (x *Xia) GET(pattern string, handler HandlerFunc) {
	// pattern = "GET-" + pattern
	x.router.addRouter("GET", pattern, handler)
}

func (x *Xia) POST(pattern string, handler HandlerFunc) {
	// pattern = "POST-" + pattern
	x.router.addRouter("POST", pattern, handler)
}

func (x *Xia) PUT(pattern string, handler HandlerFunc) {
	// pattern = "PUT-" + pattern
	x.router.addRouter("PUT", pattern, handler)
}

func (x *Xia) DELETE(pattern string, handler HandlerFunc) {
	// pattern = "DELETE-" + pattern
	x.router.addRouter("DELETE", pattern, handler)
}

// create static handler
func (x *Xia) createStaticHandler(absolutePath string, fs http.FileSystem) HandlerFunc {
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		xlog.Info(file)
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
func (x *Xia) Static(absolutePath string, root string) {
	handler := x.createStaticHandler(absolutePath, http.Dir(root))
	urlPattern := path.Join(absolutePath, "/*filepath")
	// Register GET handlers
	x.GET(urlPattern, handler)
}

func (x *Xia) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares = []HandlerFunc{}
	for _, g := range x.groups {
		if strings.HasPrefix(r.URL.Path, g.prefix) {
			middlewares = append(middlewares, g.middlewares...)
		}
	}
	c := newContext(w, r)
	c.middlewares = middlewares
	c.Req.ParseForm()
	x.router.handle(c)
}

func (x *Xia) SetAddr(addr string) {
	x.addr = addr
}

func (x *Xia) ServerStart() {
	if x.addr == "" {
		// 设置默认启动地址
		x.addr = ":9999"
	}
	xlog.Infof("listen localhost %s", x.addr)
	err := http.ListenAndServe(x.addr, x)
	if err != nil {
		xlog.Errorf("server start file, %v", err)
	}
}

func (x *Xia) EnvInit(c *xconfig.WeConfig){
	// env init --> 标准化命名

	// log init
	xlog.InitLogerProject(c)
}

// ------------------------- Response 返回值的结构体

type ResponseXia struct {
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}
