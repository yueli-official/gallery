package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/response"
)

func Healthz(r *ghttp.Request) {
	r.Response.WriteJson(response.OK(map[string]any{"status": "ok"}))
}
