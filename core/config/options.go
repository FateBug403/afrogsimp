package config

import (
	"github.com/FateBug403/afrogsimp/core/poc"
	"github.com/FateBug403/afrogsimp/core/result"
	"github.com/zan8in/goflags"
	fileutil "github.com/zan8in/pins/file"
	sliceutil "github.com/zan8in/pins/slice"
	"sort"
	"strings"
)

type OnResult func(*result.Result)

type Options struct {
	Targets sliceutil.SafeSlice

	// 配置 http/socks5 代理
	Proxy string

	// 等待超时时间 (default 10)
	Timeout int

	MaxRespBodySize int

	Cookie string

	// number of times to retry a failed request (default 1)
	Retries int

	// 检测目标是否存活的最大次数
	MaxHostError int

	// PoC file or directory to scan
	PocFile string

	// sort
	// -sort severity (default low, info, medium, high, critical)
	// -sort a-z
	Sort string

	// 下面四个全是加载poc时的过滤选项，可以不用管
	Search string // search PoC by keyword , eg: -s tomcat
	SearchKeywords []string
	Severity string // pocs to run based on severity. Possible values: info, low, medium, high, critical
	SeverityKeywords []string

	// 下面两个是扩展poc的选项
	ExcludePocs     goflags.StringSlice
	ExcludePocsFile string

}

var DefaultOptions = &Options{
	Proxy:            "",
	Timeout:          10,
	Retries:          1,
	MaxRespBodySize:  2,
	Cookie:           "",
	MaxHostError:     3,
}

// 定义包含 POC 结构的切片
type POCSlices []poc.Poc

// 实现 sort.Interface 接口的 Len、Less 和 Swap 方法
func (s POCSlices) Len() int {
	return len(s)
}

func (s POCSlices) Less(i, j int) bool {
	// 比较两个 poc.Id 字段的首字母
	return s[i].Id < s[j].Id
}

func (s POCSlices) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// CreatePocList 加载poc
func (o *Options) CreatePocList() []poc.Poc {
	var pocSlice []poc.Poc

	if len(o.PocFile) > 0 && len(poc.LocalTestList) > 0 {
		for _, pocYaml := range poc.LocalTestList {
			if p, err := poc.LocalReadPocByPath(pocYaml); err == nil {
				pocSlice = append(pocSlice, p)
			}
		}
		return pocSlice
	}
	// 获取本地Home目录下的afrog-poc的所有poc路径
	for _, pocYaml := range poc.LocalAppendList {
		if p, err := poc.LocalReadPocByPath(pocYaml); err == nil {
			pocSlice = append(pocSlice, p)
		}
	}

	for _, pocYaml := range poc.LocalFileList {
		if p, err := poc.LocalReadPocByPath(pocYaml); err == nil {
			pocSlice = append(pocSlice, p)

		}
	}
	// 获取程序内嵌的pocs下的afrog-poc的所有poc路径
	//for _, pocEmbedYaml := range pocs.EmbedFileList {
	//	if p, err := pocs.EmbedReadPocByPath(pocEmbedYaml); err == nil {
	//		pocSlice = append(pocSlice, p)
	//	}
	//}

	newPocSlice := []poc.Poc{}
	for _, poc := range pocSlice {
		if o.FilterPocSeveritySearch(poc.Id, poc.Info.Name, poc.Info.Severity) {
			newPocSlice = append(newPocSlice, poc)
		}
	}

	latestPocSlice := []poc.Poc{}
	order := []string{"info", "low", "medium", "high", "critical"}
	for _, o := range order {
		for _, s := range newPocSlice {
			if o == strings.ToLower(s.Info.Severity) {
				latestPocSlice = append(latestPocSlice, s)
			}
		}
	}

	// exclude pocs
	excludePocs, _ := o.parseExcludePocs()
	finalPocSlice := []poc.Poc{}
	for _, poc := range latestPocSlice {
		if !isExcludePoc(poc, excludePocs) {
			finalPocSlice = append(finalPocSlice, poc)
		}
	}

	if o.Sort == "a-z" {
		sort.Sort(POCSlices(finalPocSlice))
	}

	return finalPocSlice
}



func (o *Options) FilterPocSeveritySearch(pocId, pocInfoName, severity string) bool {
	var isShow bool
	if len(o.Search) > 0 && o.SetSearchKeyword() && len(o.Severity) > 0 && o.SetSeverityKeyword() {
		if o.CheckPocKeywords(pocId, pocInfoName) && o.CheckPocSeverityKeywords(severity) {
			isShow = true
		}
	} else if len(o.Severity) > 0 && o.SetSeverityKeyword() {
		if o.CheckPocSeverityKeywords(severity) {
			isShow = true
		}
	} else if len(o.Search) > 0 && o.SetSearchKeyword() {
		if o.CheckPocKeywords(pocId, pocInfoName) {
			isShow = true
		}
	} else {
		isShow = true
	}
	return isShow
}

func (o *Options) CheckPocKeywords(id, name string) bool {
	if len(o.SearchKeywords) > 0 {
		for _, v := range o.SearchKeywords {
			v = strings.ToLower(v)
			if strings.Contains(strings.ToLower(id), v) || strings.Contains(strings.ToLower(name), v) {
				return true
			}
		}
	}
	return false
}

func (o *Options) CheckPocSeverityKeywords(severity string) bool {
	if len(o.SeverityKeywords) > 0 {
		for _, v := range o.SeverityKeywords {
			if strings.EqualFold(severity, v) {
				return true
			}
		}
	}
	return false
}

func (o *Options) SetSeverityKeyword() bool {
	if len(o.Severity) > 0 {
		arr := strings.Split(o.Severity, ",")
		if len(arr) > 0 {
			for _, v := range arr {
				o.SeverityKeywords = append(o.SeverityKeywords, strings.TrimSpace(v))
			}
			return true
		}
	}
	return false
}

func (o *Options) SetSearchKeyword() bool {
	if len(o.Search) > 0 {
		arr := strings.Split(o.Search, ",")
		if len(arr) > 0 {
			for _, v := range arr {
				o.SearchKeywords = append(o.SearchKeywords, strings.TrimSpace(v))
			}
			return true
		}
	}
	return false
}

func (o *Options) parseExcludePocs() ([]string, error) {
	var excludePocs []string
	if len(o.ExcludePocs) > 0 {
		excludePocs = append(excludePocs, o.ExcludePocs...)
	}

	if len(o.ExcludePocsFile) > 0 {
		cdata, err := fileutil.ReadFile(o.ExcludePocsFile)
		if err != nil {
			if len(excludePocs) > 0 {
				return excludePocs, nil
			} else {
				return excludePocs, err
			}
		}
		for poc := range cdata {
			excludePocs = append(excludePocs, poc)
		}
	}
	return excludePocs, nil
}

func isExcludePoc(poc poc.Poc, excludePocs []string) bool {
	if len(excludePocs) == 0 {
		return false
	}
	for _, ep := range excludePocs {
		v := strings.ToLower(ep)
		if strings.Contains(strings.ToLower(poc.Id), v) || strings.Contains(strings.ToLower(poc.Info.Name), v) {
			return true
		}
	}
	return false
}