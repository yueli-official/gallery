package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

func Healthz(r *ghttp.Request) {
	r.Response.WriteJson(map[string]any{"status": "ok"})
}
