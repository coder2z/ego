package ehttp

import (
	"github.com/go-resty/resty/v2"
	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/util/xdebug"
	"time"
)

const PackageName = "client.ehttp"

type Component struct {
	name   string
	config *Config
	logger *elog.Component
	*resty.Client
}

func newComponent(name string, config *Config, logger *elog.Component) *Component {
	restyClient := resty.New().SetDebug(config.RawDebug).SetTimeout(config.ReadTimeout).OnBeforeRequest(func(client *resty.Client, request *resty.Request) error {
		client.Header.Set("app", eapp.Name())
		return nil
	}).OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		rr := response.Request.RawRequest
		if eapp.IsDevelopmentMode() {
			xdebug.Info(name, config.Address, response.Time(), response.Request.Method+"."+rr.URL.RequestURI(), string(response.Body()))
		}

		isSlowLog := false
		var fields = make([]elog.Field, 0, 8)

		if config.SlowLogThreshold > time.Duration(0) && response.Time() > config.SlowLogThreshold {
			fields = append(fields,
				elog.FieldEvent("slow"),
				elog.FieldMethod(response.Request.Method+"."+rr.URL.RequestURI()), // GET./hello
				elog.FieldName(name),
				elog.FieldCost(response.Time()),
				elog.FieldAddr(rr.URL.Host),
			)
			if config.EnableAccessInterceptorReply {
				fields = append(fields, elog.FieldValueAny(string(response.Body())))
			}
			logger.Warn("access", fields...)
			isSlowLog = true
		}

		if config.EnableAccessInterceptor && !isSlowLog {
			fields = append(fields,
				elog.FieldEvent("normal"),
				elog.FieldMethod(response.Request.Method+"."+rr.URL.RequestURI()), // GET./hello
				elog.FieldName(name),
				elog.FieldCost(response.Time()),
				elog.FieldAddr(rr.URL.Host),
			)
			if config.EnableAccessInterceptorReply {
				fields = append(fields, elog.FieldValueAny(string(response.Body())))
			}
			logger.Info("access", fields...)
		}

		return nil
	}).SetHostURL(config.Address)
	return &Component{
		name:   name,
		config: config,
		logger: logger,
		Client: restyClient,
	}
}
