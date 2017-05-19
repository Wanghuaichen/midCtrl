package main

import (
	"net/http"
	"net/url"
)

func main() {
	http.PostForm("http://39.108.5.184/smart/api/saveElectricityData", url.Values{"hdId": {"1"}, "time": {"123456464"}, "record": {"123456464"}, "kw": {"123456464"}, "pt": {"123456464"}, "ct": {"123456464"}})
}
