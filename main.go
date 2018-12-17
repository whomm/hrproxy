package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/whomm/hrproxy/httptool"
)

//限流item
type Limit struct {
	Path string
	Qps  float64
	Code int
	Msg  string
}

//配置
type Conf struct {
	Listen       string
	Backservices string
	Limitlist    []Limit
}

var conffile = flag.String("c", "", "confige file like def.yaml")

func main() {
	flag.Parse()
	if len(*conffile) < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	confbytes, err := ioutil.ReadFile(*conffile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conf := Conf{}

	err = yaml.Unmarshal(confbytes, &conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	serv := httptool.NewServer(conf.Listen, conf.Backservices)
	for _, i := range conf.Limitlist {
		serv.Handler(i.Path, i.Qps, i.Code, i.Msg, nil)
	}

	fmt.Println("start success")
	serv.ListenAndServe()
}
