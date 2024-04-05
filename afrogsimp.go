package afrogsimp

import (
	"github.com/FateBug403/afrogsimp/core/config"
	"github.com/FateBug403/afrogsimp/core/engine"
	"github.com/FateBug403/afrogsimp/core/result"
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
)

type AfrogSimp struct {
	Options *config.Options
}

func NewAfrogSimp(Options *config.Options)*AfrogSimp{
	return &AfrogSimp{Options: Options}
}

// CheckPoc 通过ants goroutine池,来实现对某个目标进行批量POC探测
func (receiver *AfrogSimp) CheckPoc(target string,pocSearch string) ([]*result.Result,error) {
	var err error
	options := config.DefaultOptions
	options.Search=pocSearch
	pocSlice := options.CreatePocList()
	ege := engine.NewEngine(options)

	var vulList []*result.Result
	var wg sync.WaitGroup
	p, err := ants.NewPoolWithFunc(10, func(p any) {
		tap := p.(*engine.TransData)
		execResult,err:=execPoc(ege,tap)
		if execResult.IsVul&&err==nil{
			vulList = append(vulList,execResult)
		}
		wg.Done()
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer p.Release()

	for _, poc := range pocSlice {
		wg.Add(1)
		p.Invoke(&engine.TransData{Target: target, Poc: poc})
	}
	wg.Wait()
	return vulList,err
}

//execPoc  调用Check来执行poc
func execPoc(ege *engine.Engine,transData *engine.TransData) (*result.Result,error)  {
	var err error
	c := ege.AcquireChecker()
	defer ege.ReleaseChecker(c)
	err =c.Check(transData.Target,&transData.Poc)
	return c.Result,err
}