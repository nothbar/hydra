package middleware

// Licensed to SkyAPM org under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. SkyAPM org licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/micro-plat/hydra/components"
	"github.com/micro-plat/hydra/components/pkgs/apm"
	"github.com/micro-plat/hydra/components/pkgs/apm/apmtypes"
	"github.com/micro-plat/hydra/global"
	"github.com/micro-plat/lib4go/net"
)

//APMHttp 调用链
func APMHttp() Handler {

	return func(ctx IMiddleContext) {
		//fmt.Println("middleware.apm")
		//获取apm配置
		apmconf := ctx.ServerConf().GetAPMConf()
		if !apmconf.GetEnable() {
			ctx.Next()
			return
		}
		ctx.Response().AddSpecial("apm.http")

		octx := ctx.Context()
		//fmt.Println("middleware.apm-1", octx)
		oreq, _ := ctx.GetHttpReqResp()
		instance := fmt.Sprintf("%s_%s", global.Def.PlatName, net.GetLocalIPAddress())

		apmInstance := components.Def.APM().GetRegularAPM(instance, apmconf.GetConfig())

		tracer, err := apmInstance.CreateTracer(global.Def.GetAPMService())
		if err != nil {
			ctx.Log().Warnf("APM.CreateTracer:%+v", err)
			ctx.Next()
			return
		}

		//fmt.Println("middleware.apm-2", tracer, err)
		span, rootctx, err := tracer.CreateEntrySpan(octx, gethttpOperationName("", oreq), func() (string, error) {
			return oreq.Header.Get(apm.Header), nil
		})
		if err != nil {
			ctx.Log().Warnf("APM.CreateEntrySpan:%+v", err)
			ctx.Next()
			return
		}
		ctx.Meta().Set(apm.TraceInfo, &apm.APMInfo{
			Tracer:  tracer,
			RootCtx: rootctx,
		})
		//fmt.Println("middleware.apm-3", oreq.Header.Get("X-Request-Id"))
		span.SetComponent(apmtypes.ComponentIDGOHttpServer)
		span.Tag("X-Request-Id", oreq.Header.Get("X-Request-Id"))
		// for k, v := range h.extraTags {
		//
		// }
		span.Tag(apm.TagHTTPMethod, oreq.Method)
		span.Tag(apm.TagURL, fmt.Sprintf("%s%s", oreq.Host, oreq.URL.Path))
		span.SetSpanLayer(apm.SpanLayer_Http)

		defer func() {
			statusCode, _ := ctx.Response().GetRawResponse()
			code := statusCode
			if code >= 400 {
				span.Error(time.Now(), "Error on handling request,code:"+strconv.Itoa(code))
			}
			//fmt.Println("middleware.apm-4", statusCode)
			span.Tag(apm.TagStatusCode, strconv.Itoa(code))
			span.End()
		}()

		ctx.Next()
	}
}

func gethttpOperationName(name string, r *http.Request) string {
	if name == "" {
		return fmt.Sprintf("%s", r.URL.Path)
	}
	return name
}