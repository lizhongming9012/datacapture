package main

import (
	"crypto/rand"
	"github.com/PuerkitoBio/goquery"
	r "math/rand"
	"strings"
	"time"
)

func RandomString(n int, alphabets ...byte) string {
	const alphanum = "0123456789"
	var bytes = make([]byte, n)
	var randby bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randby = true
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			if randby {
				bytes[i] = alphanum[r.Intn(len(alphanum))]
			} else {
				bytes[i] = alphanum[b%byte(len(alphanum))]
			}
		} else {
			if randby {
				bytes[i] = alphabets[r.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return string(bytes)
}

func getNodeNameValue(s *goquery.Selection) (string, string) {
	var k, vv string
	for _, v := range s.Nodes[0].Attr {
		if strings.Compare(v.Key, "name") == 0 {
			k = v.Val
		}
		if strings.Compare(v.Key, "value") == 0 {
			vv = v.Val
		}
	}
	return k, vv
}
