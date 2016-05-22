package config

import (
	"io/ioutil"
	"github.com/bitly/go-simplejson"
)

type config struct {
	From string
	To string
	Limit int
}

var C config

func init() {
	bytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("error")
	}
	js, err := simplejson.NewJson(bytes)
	if err != nil {
		panic("error")
	}
	C.From = js.Get("from").MustString()
	C.To=js.Get("to").MustString()
	C.Limit=js.Get("routineLimit").MustInt()
}

