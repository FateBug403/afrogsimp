package main

import (
	"github.com/FateBug403/afrogsimp"
	"github.com/FateBug403/afrogsimp/core/config"
	"log"
)

func main() {
	options := config.DefaultOptions
	options.Proxy="http://127.0.0.1:8080"
	runClient := afrogsimp.NewAfrogSimp(options)
	vlueList,err:=runClient.CheckPoc("http://www.example.com","druid")
	if err != nil {
		log.Println(err)
	}
	if len(vlueList)>0{
		for _,value := range vlueList{
			log.Println(value.PocInfo.Id)
		}
	}
}


