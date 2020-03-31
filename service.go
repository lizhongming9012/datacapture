package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
	"log"
	"strconv"
	"strings"
	"sync"
)

var IeHeader = "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Win64; x64; Trident/4.0; .NET CLR 2.0.limit727; SLCC2; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET4.0C; .NET4.0E)"

func LoginZsk(client *gorequest.SuperAgent) bool {
	var b bool
	var rt = c.RetryTime
	uri := "http://76.12.128.189/szzsk/sword?ctrl=LoginCtrl_logon&tokenLog=ODAzMiYxMjM0NTY="
	_, body, errs := client.Get(uri).Set("Accept", "*/*").
		Set("Accept-Language", "zh-CN").Set("User-Agent", IeHeader).End()
	if len(errs) != 0 {
		log.Println("zsk login err!")
		log.Printf("retry times left %d:\n", rt)
		if rt > 0 {
			b = LoginZsk(client)
		}
		return false
	}
	if strings.Contains(body, "登录成功！") {
		b = true
	}
	return b
}

func QueryZsk(client *gorequest.SuperAgent, rt int) {
	uri := "http://76.12.128.189/szzsk/ajax.sword?r=0." + RandomString(16)
	pMap := make(map[string]string)
	pMap["postData"] = fmt.Sprintf(`{"tid":"ZskkjsBLH_findMoreHots","pageNum":1,"rows":15,"queryType":"page","widgetname":"hotsGridID","sortName":"zlfz","sortFlag":"desc"}`)
	log.Printf("begin query zsk data")
	_, body, errs := client.Post(uri).Set("Accept-Language", "zh-CN").
		Set("Accept", "text/javascript, text/html, application/xml, text/xml, */*").
		Set("Referer", "http://76.12.128.189/szzsk/sword?tid=ZskkjsBLH_findMoreHots").
		Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		Set("User-Agent", IeHeader).
		Type(gorequest.TypeForm).
		SendMap(pMap).End()
	if len(errs) != 0 {
		log.Printf("query error:\n %v\n", errs[0])
		log.Printf("retry times left %d:\n", rt)
		if rt > 0 {
			QueryZsk(client, rt-1)
		}
		return
	}
	tot := firstParseZsk(body)
	log.Printf("totalRows is %d", tot)
	pages := 1
	if tot <= 0 {
		log.Println("query totalRows < 0")
		return
	}
	if tot%c.Limit == 0 {
		pages = tot / c.Limit
	} else {
		pages = tot/c.Limit + 1
	}
	var wg sync.WaitGroup
	var seg int
	if pages%c.GoroutineNum == 0 {
		seg = pages / c.GoroutineNum
	} else {
		seg = pages/c.GoroutineNum + 1
	}
	log.Printf("启用多线程技术获取zlcode，当前启动线程%d个.", c.GoroutineNum)
	var zlcodes []map[string]string
	for i := 1; i <= c.GoroutineNum; i++ {
		wg.Add(1)
		cliClone := client.Clone()
		beg := (i - 1) * seg
		end := i * seg
		if end == pages {
			end++
			i = c.GoroutineNum + 1
		}
		go func() {
			zs := segZskPageQueryPost(cliClone, beg, end, pages, rt, &wg)
			zlcodes = append(zlcodes, zs...)
		}()
	}
	wg.Wait()
	log.Println("zsk list query success")
	log.Printf("zlcodes num is %d", len(zlcodes))
	segZskQueryDetail(client, c.GoroutineNum, c.RetryTime, zlcodes)
}

func PageZskQuery(client *gorequest.SuperAgent, pageNum, rows, rt int) []map[string]string {
	var rst []map[string]string
	uri := "http://76.12.128.189/szzsk/ajax.sword?r=0." + RandomString(16)
	pMap := make(map[string]string)
	pMap["postData"] = fmt.Sprintf(`{"tid":"ZskkjsBLH_findMoreHots","pageNum":` +
		strconv.Itoa(pageNum) + `,"rows":` + strconv.Itoa(rows) +
		`,"queryType":"page","widgetname":"hotsGridID","sortName":"zlfz","sortFlag":"desc"}`)
	//log.Println(pMap)
	_, body, errs := client.Post(uri).Set("Accept-Language", "zh-CN").
		Set("Accept", "text/javascript, text/html, application/xml, text/xml, */*").
		Set("Referer", "http://76.12.128.189/szzsk/sword?tid=ZskkjsBLH_findMoreHots").
		Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		Set("User-Agent", IeHeader).
		Type(gorequest.TypeForm).
		SendMap(pMap).End()
	if len(errs) != 0 {
		log.Printf("query error:\n %v\n", errs[0])
		log.Printf("retry times left %d:\n", rt)
		if rt > 0 {
			rst = PageZskQuery(client, pageNum, rows, rt-1)
		}
		return nil
	}
	rst = pageParseZsk(body)
	return rst
}

func QueryZskDetail(cli *gorequest.SuperAgent, data map[string]string, rt int) {
	uri := "http://76.12.128.189/szzsk/sword?tid=ZskzjjsBLH_findBigNr&zlcode=" +
		data["zlcode"] + "&show=yes&zlflag=" + data["zlflag"]
	resp, body, errs := cli.Get(uri).Set("Accept-Language", "zh-CN").
		Set("Accept", "*/*").Set("User-Agent", IeHeader).End()
	if len(errs) != 0 {
		log.Printf("query error:\n %v\n", errs[0])
		log.Printf("retry times left %d:\n", rt)
		if rt > 0 {
			QueryZskDetail(cli, data, rt-1)
		}
		return
	}
	zsk := Zsk{
		Zlcode: data["zlcode"],
		Zlflag: data["zlflag"],
		Title:  data["zltitle"],
		Type:   data["zltypemc"],
		Uptime: data["zllrsj"],
	}
	if strings.Contains(body, "无效") || strings.Contains(body, "失效") {
		if e := DeleteData(zsk.Zlcode); e != nil {
			log.Printf("delete data err:%v", e)
		}
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Parse the body fail...%v", err)
		return
	}
	doc.Find("#moreInfo4 .yctd").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			zsk.Key = s.Text()
		}
	})
	zsk.Answer = strings.TrimSpace(doc.Find(".texts").Text())
	//fmt.Println(zsk)
	if IsDataExist(zsk.Zlcode) {
		if IsNewData(zsk.Zlcode, zsk.Uptime) {
			return
		}
		if e := DeleteData(zsk.Zlcode); e != nil {
			log.Printf("delete data err:%v", e)
		}
	}
	if err = AddData(&zsk); err != nil {
		log.Printf("create data fail：%v", err)
	}
}

func firstParseZsk(b string) (tot int) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(b), &m)
	if err != nil {
		log.Println("unmarshal json error in FirstParseZsk")
		return
	}
	data := m["data"].([]interface{})
	d1 := data[0].(map[string]interface{})
	tot = int(d1["totalRows"].(float64))
	//log.Printf("TOT IS :%d\n", tot)
	return tot
}

func pageParseZsk(body string) []map[string]string {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(body), &m)
	if err != nil {
		log.Fatalln("unmarshal json error in PageParseZsk")
		return nil
	}
	data := m["data"].([]interface{})
	d0 := data[0].(map[string]interface{})
	trsa := d0["trs"].([]interface{})
	rst := make([]map[string]string, 0)
	for i := range trsa {
		trsaRec0 := trsa[i].(map[string]interface{})
		trsaRec0Content := trsaRec0["tds"].(map[string]interface{})
		tempMap := make(map[string]string)
		for k, v := range trsaRec0Content {
			switch k {
			case "zlcode":
				mv := v.(map[string]interface{})
				tempMap["zlcode"] = mv["value"].(string)
			case "zlflag":
				mv := v.(map[string]interface{})
				tempMap["zlflag"] = mv["value"].(string)
			case "zltitle":
				mv := v.(map[string]interface{})
				tempMap["zltitle"] = mv["value"].(string)
			case "zltypemc":
				mv := v.(map[string]interface{})
				tempMap["zltypemc"] = mv["value"].(string)
			case "zllrsj":
				mv := v.(map[string]interface{})
				tempMap["zllrsj"] = mv["value"].(string)
			}

		}
		rst = append(rst, tempMap)
	}
	//log.Printf("zlcodes is :%v\n", rst)
	return rst
}

func segZskPageQueryPost(cli *gorequest.SuperAgent, beg, end, pages, rt int, wg *sync.WaitGroup) []map[string]string {
	defer wg.Done()
	zlcodes := make([]map[string]string, 0)
	if end == pages {
		end++
	}
	for j := beg; j < end; j++ {
		if j <= 0 {
			continue
		}
		if j <= pages {
			zs := PageZskQuery(cli, j, c.Limit, rt)
			zlcodes = append(zlcodes, zs...)
		}
	}
	return zlcodes
}

func segZskQueryDetail(cli *gorequest.SuperAgent, syncNum, rt int, zlcodes []map[string]string) {
	zlcodesChan := make(chan map[string]string, 100)
	cntChan := make(chan int)
	var seg int
	zlcodesNum := len(zlcodes)
	if zlcodesNum%8 == 0 {
		seg = zlcodesNum / 8
	} else {
		seg = (zlcodesNum / 8) + 1
	}
	for j := 0; j < 8; j++ {
		beg := j * seg
		if beg > zlcodesNum-1 {
			break
		}
		var end int
		if (j+1)*seg < zlcodesNum {
			end = (j + 1) * seg
		} else {
			end = zlcodesNum
		}
		segIds := zlcodes[beg:end]
		go func() {
			for i, zlId := range segIds {
				zlcodesChan <- zlId
				cntChan <- i
			}
		}()
	}
	go func() {
		var num int
		for range cntChan {
			num++
			if num == zlcodesNum {
				close(zlcodesChan)
			}
		}
	}()
	wg := &sync.WaitGroup{}
	log.Printf("启用多线程技术更新数据，当前启动线程%d个.", syncNum)
	for k := 0; k < syncNum; k++ {
		wg.Add(1)
		cliClone := cli.Clone()
		go func() {
			defer wg.Done()
			for zlcode := range zlcodesChan {
				//if strings.Contains(zlcode["zltypemc"], "问题解答") {
				QueryZskDetail(cliClone, zlcode, rt)
				//}
			}
		}()
	}
	wg.Wait()
	log.Println("query zsk detail end")
}
