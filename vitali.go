package vitali

import (
    "os"
    "io"
    "log"
    "fmt"
    "net/http"
    "time"
    "regexp"
    "strings"
    "reflect"
)

const (
    PUBLIC = iota
    AUTHENTICATED
)

type RouteRule struct {
    Pattern string
    Resource interface{}
}

type PatternMapping struct {
    Re *regexp.Regexp
    Names []string
}

type UserProvider interface {
    AuthHeader(*http.Request) string
    User(*http.Request) string
}

type webApp struct {
    RouteRules []RouteRule
    PatternMappings []PatternMapping
    UserProvider UserProvider
    Settings map[string]string
}

func checkPermission(perm Perm, method string, user string) bool {
    required, exist := perm[method]
    if !exist {
        required = perm["*"]
    }
    return !(required==AUTHENTICATED && user == "")
}

func matchRules(c webApp, w *wrappedWriter, r *http.Request) {
    for i, routeRule := range c.RouteRules {
        params := c.PatternMappings[i].Re.FindStringSubmatch(r.URL.Path)
        if params != nil {
            pathParams := make(map[string]string)
            if len(params) > 1 {
                for j, param := range params[1:] {
                    pathParams[c.PatternMappings[i].Names[j]] = param
                }
            }

            ctx := Ctx {
                pathParams: pathParams,
                Username: c.UserProvider.User(r),
                Request: r,
                ResponseWriter: w,
            }

            vResource := reflect.ValueOf(routeRule.Resource)
            vNewResource := reflect.New(reflect.TypeOf(routeRule.Resource)).Elem()
            for i := 0; i < vResource.NumField(); i++ {
                srcField := vResource.Field(i)
                newField := vNewResource.Field(i)

                switch reflect.TypeOf(srcField.Interface()).Name() {
                case "Ctx":
                    newField.Set(reflect.ValueOf(ctx))
                case "Perm":
                    if !checkPermission(srcField.Interface().(Perm), r.Method, ctx.Username) {
                        if c.Settings["401_PAGE"] != "" {
                            w.Header().Set("Content-Type", "text/html")
                            w.Header()["WWW-Authenticate"] = []string{c.UserProvider.AuthHeader(r)}
                            w.WriteHeader(http.StatusUnauthorized)
                            f, err := os.Open(c.Settings["401_PAGE"])
                            if err != nil {
                                panic(err)
                            }
                            io.Copy(w, f)
                        } else {
                            http.Error(w, "unauthorized", http.StatusUnauthorized)
                        }
                        return
                    }
                default:
                    newField.Set(srcField)
                }
            }
            resource := vNewResource.Interface()

            result := getResult(r.Method, resource)
            writeResponse(w, r, result)
            return
        }
    }
    http.NotFound(w, r)
}

func getAllowed(resource interface{}) (allowed []string) {
    _, ok := resource.(Getter)
    if ok {
        allowed = append(allowed, "GET", "HEAD")
    }
    _, ok = resource.(Poster)
    if ok {
        allowed = append(allowed, "POST")
    }
    _, ok = resource.(Putter)
    if ok {
        allowed = append(allowed, "PUT")
    }
    _, ok = resource.(Deleter)
    if ok {
        allowed = append(allowed, "DELETE")
    }
    return
}

func getResult(method string, resource interface{}) (result interface{}) {
    defer func() {
        if r := recover(); r != nil {
            rstr := fmt.Sprintf("%s", r)
            result = internalError {
                where: lineInfo(3),
                why: rstr + fullTrace(5, "\n\t"),
                code: errorCode(rstr),
            }
        }
    }()

    switch method {
    case "HEAD", "GET":
        h, ok := resource.(Getter)
        if ok {
            result = h.Get()
        }
    case "POST":
        h, ok := resource.(Poster)
        if ok {
            result = h.Post()
        }
    case "PUT":
        h, ok := resource.(Putter)
        if ok {
            result = h.Put()
        }
    case "DELETE":
        h, ok := resource.(Deleter)
        if ok {
            result = h.Delete()
        }
    default:
        return notImplemented{}
    }

    return methodNotAllowed{getAllowed(resource)}
}

func (c webApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    ww := &wrappedWriter{
        status: 0,
        writer: w,
        inTime: time.Now(),
    }

    matchRules(c, ww, r)

    elapsedMs := float64(time.Now().UnixNano() - ww.inTime.UnixNano()) / 1000000
    if ww.status == 0 {
        log.Printf("%s %s %s Client Disconnected (%.2f ms)", r.RemoteAddr, r.Method,
            r.URL.Path, elapsedMs)
    } else {
        errMsg := ""
        if ww.err.why != "" {
            errMsg = fmt.Sprintf("%s #%d %s ", ww.err.where, ww.err.code, ww.err.why)
        }
        log.Printf("%s %s %s %s %s(%.2f ms, %d bytes)", r.RemoteAddr, r.Method, r.URL.Path,
            http.StatusText(ww.status), errMsg, elapsedMs, ww.written)
    }
}

type EmptyUserProvider struct {
}

func (c EmptyUserProvider) AuthHeader(r *http.Request) string {
    return ""
}

func (c EmptyUserProvider) User(r *http.Request) string {
    return ""
}

func CreateWebApp(rules []RouteRule) webApp {
    patternMappings := make([]PatternMapping, len(rules))
    for i, v := range rules {
        re := regexp.MustCompile("{[^}]*}")
        params := re.FindAllString(v.Pattern, -1)
        names := make([]string, len(params))

        transformedPattern := v.Pattern
        for j, param := range params {
            names[j] = param[1:len(param)-1]
            transformedPattern = strings.Replace(transformedPattern, param, "([^/]+)", -1)
        }
        patternMappings[i] = PatternMapping{regexp.MustCompile("^"+transformedPattern+"$"), names}
    }

    return webApp{
        RouteRules: rules,
        PatternMappings: patternMappings,
        UserProvider: EmptyUserProvider{},
        Settings: make(map[string]string),
    }
}

type Perm map[string]int