package main

import (
	"github.com/parnurzeal/gorequest"
	"log"
)

func StartZsk() {
	var client = gorequest.New()
	if LoginZsk(client) {
		QueryZsk(client, c.RetryTime)
		//mp := map[string]string{
		//	"zlcode":  "0000000002015051300007",
		//	"zlflag":  "5",
		//	"zltitle": "航空运输企业分支机构传递单",
		//	"zltype":  "表证单书->申报征收->增值税纳税申报",
		//	"zllrsj":  "2015-05-13 14:59:40.0000",
		//}
		//QueryZskDetail(client, mp, c.RetryTime)
	}
	log.Println("delete data with tables...")
	if err := DeleteNullData(); err != nil {
		log.Printf("delete data err:%v", err)
	}
	log.Println("----that's ok!---")
}
