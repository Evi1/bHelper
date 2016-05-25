package config

import (
	"io/ioutil"
	"github.com/bitly/go-simplejson"
	"os"
	"fmt"
	"bufio"
	"path/filepath"
)

type config struct {
	From  string
	To    string
	Limit int
}

var C config

func init() {
	bytes, err := ioutil.ReadFile(filepath.FromSlash("./config.json"))
	if err != nil {
		out, err := os.OpenFile(filepath.FromSlash("./config.json"), os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("An error occurred with file opening or creation:config.json")
			return
		}
		defer out.Close()
		outputWriter := bufio.NewWriter(out)
		outputWriter.WriteString("{\"from\": \"./\",\"to\": \"./bilibili/\",\"routineLimit\": 4}")
		outputWriter.Flush()
		C.From = "./"
		C.To = "./bilibili/"
		C.Limit = 4
		return
	}
	js, err := simplejson.NewJson(bytes)
	if err != nil {
		panic(err)
	}
	C.From = js.Get("from").MustString()
	C.To = js.Get("to").MustString()
	C.Limit = js.Get("routineLimit").MustInt()
}

