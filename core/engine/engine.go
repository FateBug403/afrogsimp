package engine

import (
	mycel "github.com/FateBug403/afrogsimp/core/cel"
	"github.com/FateBug403/afrogsimp/core/checker"
	"github.com/FateBug403/afrogsimp/core/config"
	"github.com/FateBug403/afrogsimp/core/poc"
	"github.com/FateBug403/afrogsimp/core/protocols/http/retryhttpclient"
	"github.com/FateBug403/afrogsimp/core/result"
	"sync"
	"time"
)

type TransData struct {
	Target string
	Poc    poc.Poc
}

type Engine struct {
	options *config.Options
	ticker  *time.Ticker
}

func NewEngine(options *config.Options) *Engine {
	// 这里配置的参数主要用户出raw请求以外的http请求
	retryhttpclient.Init(&retryhttpclient.Options{
		Proxy:           options.Proxy,
		Timeout:         options.Timeout,
		Retries:         options.Retries,
		MaxRespBodySize: options.MaxRespBodySize,
	})
	// 这里的options主要用于raw请求
	engine := &Engine{
		options: options,
	}
	return engine
}

// CheckerPool 创建对象池，用于临时存储和重用对象，以提高性能
var CheckerPool = sync.Pool{
	New: func() any {
		return &checker.Checker{
			Options: &config.Options{},
			// OriginalRequest: &http.Request{},
			VariableMap: make(map[string]any),
			Result:      &result.Result{},
			CustomLib:   mycel.NewCustomLib(),
		}
	},
}

// AcquireChecker 从对象池获取一个Checker对象
func (e *Engine) AcquireChecker() *checker.Checker {
	c := CheckerPool.Get().(*checker.Checker)
	c.Options = e.options
	return c
}

// ReleaseChecker 释放指定Checker对象到对象池中
func (e *Engine) ReleaseChecker(c *checker.Checker) {
	// *c.OriginalRequest = http.Request{}
	c.VariableMap = make(map[string]any)
	c.Result = &result.Result{}
	c.CustomLib = mycel.NewCustomLib()
	CheckerPool.Put(c)
}