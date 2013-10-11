package vitali

import (
    "net/http"
)

type Ctx struct {
    Username    string
    Request *http.Request
    ResponseWriter  http.ResponseWriter

    pathParams map[string]string
}

func (c Ctx) AddHeader(key string, value string) {
    c.ResponseWriter.Header().Add(key, value)
}

func (c Ctx) SetCookie(cookie *http.Cookie) {
    http.SetCookie(c.ResponseWriter, cookie)
}

func (c Ctx) Param(key string) string {
    return c.Request.Form.Get(key)
}

func (c Ctx) ParamArray(key string) []string {
    return c.Request.Form[key]
}

func (c Ctx) PathParam(key string) string {
    return c.pathParams[key]
}