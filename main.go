package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"flag"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

type Config struct {
	Prefix string   `mapstructure:"prefix"`
	Exhost []string `mapstructure:"exhost"`
	Host   []string `mapstructure:"host"`
	Path   []string `mapstructure:"path"`
}

var Conf = new(Config)

var (
	proxy  = flag.String("proxy", "http://192.168.2.170:80", "proxy")
	port   = flag.String("port", "80", "port")
	config = flag.String("config", "gor-proxy.yaml", "config")
)

func init() {
	flag.Parse()
	//https://www.liwenzhou.com/posts/Go/viper_tutorial/
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	var rconfig string
	if *config == "gor-proxy.yaml" {
		rconfig = dir + "/" + *config
	} else {
		rconfig = *config
	}
	viper.SetConfigFile(rconfig)
	err := viper.ReadInConfig() // 查找并读取配置文件
	if err != nil {             // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// 将读取的配置信息保存至全局变量Conf
	if err := viper.Unmarshal(Conf); err != nil {
		panic(fmt.Errorf("unmarshal conf failed, err:%s \n", err))
	}
	// 监控配置文件变化
	viper.WatchConfig()
	// 注意！！！配置文件发生变化后要同步到全局变量Conf
	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := viper.Unmarshal(Conf); err != nil {
			panic(fmt.Errorf("unmarshal conf failed, err:%s \n", err))
		}
	})

	log.Println(dir, Conf.Exhost, Conf.Host, Conf.Path, Conf.Prefix)
}

func main() {
	//flag.Parse()

	log.Println("proxy:", *proxy, "port:", *port)
	targetUrl, err := url.Parse(*proxy)
	if err != nil {
		log.Fatal("err")
	}
	proxy := NewSingleHostReverseProxy(targetUrl)
	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		panic(err)
	}
}

func sliceContains(elems []string, v string) bool {
	for _, s := range elems {
		if strings.Contains(v, s) {
			return true
		}
	}
	return false
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Println("r.URL:", r.URL, "r.Host:", r.Host, "r.URL.Host:", r.URL.Host, "r.URL.Scheme:", r.URL.Scheme, "r.URL.Path:", r.URL.Path, "r.URL.RawQuery:", r.URL.RawQuery, "r.URL.Hostname:", r.URL.Hostname())
		//排除不相关的域名或后缀
		if sliceContains(Conf.Exhost, r.Host) || sliceContains(Conf.Path, r.URL.Path) {
			w.Header().Set("X-CACHE", "MISS1")
			fmt.Fprintln(w, "")
			log.Println(fmt.Sprintf("%s%s", r.Host, r.URL.Path), "pass1")
		} else if sliceContains(Conf.Host, r.Host) {
			w.Header().Set("X-CACHE", "HIT")
			p.ServeHTTP(w, r)
		} else {
			w.Header().Set("X-CACHE", "MISS2")
			fmt.Fprintln(w, "")
			log.Println(fmt.Sprintf("%s%s", r.Host, r.URL.Path), "pass2")
		}

	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func sliceContainsReplace(elems []string, v string) string {
	for _, s := range elems {
		if strings.Contains(v, s) {
			return strings.Replace(v, s, fmt.Sprintf("%s.%s", Conf.Prefix, s), -1)
		}
	}
	return v
}
func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		//替换掉原来的域名
		req.Host = sliceContainsReplace(Conf.Host, req.Host)
		//log.Println("asdfasdf", target.Host)
		uri := req.Host + req.URL.RequestURI()
		green := color.FgGreen.Render
		red := color.FgRed.Render
		log.Println(red("uri:"), green(uri))
		//req.Method = "GET"

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	// 自定义ModifyResponse
	modifyResp := func(resp *http.Response) error {
		//log.Println("mmm", uri, surls, key)

		var err error
		var oldData, newData []byte
		oldData, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		// 根据不同状态码修改返回内容
		if resp.StatusCode == 200 {
			newData = []byte("" + string(oldData))
		} else {
			newData = []byte("[ERROR] " + string(oldData))
		}

		// 修改返回内容及ContentLength
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(newData))
		resp.ContentLength = int64(len(newData))
		resp.Header.Set("Content-Length", fmt.Sprint(len(newData)))
		return nil
	}
	// 传入自定义的ModifyResponse
	return &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResp}
}
