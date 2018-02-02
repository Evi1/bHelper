package main

import (
	"io/ioutil"
	"log"
	"container/list"
	"os"
	"strings"
	"github.com/bitly/go-simplejson"
	"bufio"
	"github.com/Evi1/bHelper/config"
	"io"
	"github.com/Evi1/bHelper/tools"
)

var routineNum chan int
var routineLimit chan int

func init() {
	routineNum = make(chan int)
	routineLimit = make(chan int, config.C.Limit)
}

func main() {
	bPath := config.C.From + "download/"
	cPath := bPath
	files, _ := ioutil.ReadDir(cPath)
	l1 := list.New()
	for _, f := range files {
		if f.IsDir() {
			log.Println(f.Name())
			l1.PushBack(f)
		}
	}
	num := 0
	for l1.Len() > 0 {
		t := l1.Front()
		l1.Remove(t)
		x := t.Value.(os.FileInfo)
		cPath = bPath + x.Name() + "/"
		files, _ := ioutil.ReadDir(cPath)
		for _, f := range files {
			if f.IsDir() {
				num++
				routineLimit <- 1
				log.Println(cPath, f.Name())
				go handle(cPath, f)
			}
		}
	}
	for i := 0; i < num; i++ {
		<-routineNum
	}
}

func handle(cPath string, f os.FileInfo) {
	defer func() {
		<-routineLimit
		routineNum <- 1
	}()
	path := cPath + f.Name() + "/"
	files, _ := ioutil.ReadDir(path)
	title := "";
	part := "";
	v := "";
	thisTitle := "";
	inPath := ""
	gotMp4 := false
	gotConfig := false
	gotXlv := false
	var xlvs *list.List
	for _, f := range files {
		if (gotMp4 || gotXlv) && gotConfig {
			break
		}
		if f.IsDir() {
			videos, _ := ioutil.ReadDir(path + f.Name() + "/")
			l, r := tools.CheckFLV(videos)
			if r {
				log.Println(path + f.Name() + " BLV/FLV")
				inPath = f.Name() + "/"
				gotXlv = true
				xlvs = l
			}
		} else if strings.HasSuffix(f.Name(), ".json") {
			var e error
			title, part, thisTitle, e = handleJSON(path + f.Name())
			if e != nil {
				log.Println("An error occurred with json file:" + f.Name() + " : " + e.Error())
				//<-routineLimit
				//routineNum <- 1
				return
			}
			gotConfig = true;
		} else if strings.HasSuffix(f.Name(), ".mp4") {
			v = f.Name()
			inPath = ""
			gotMp4 = true
		}
	}
	inputFile := path + inPath + v;
	part = strings.TrimSpace(part)
	if len(part) <= 1 {
		part = "0" + part
	}
	outputFile := ""
	if len(thisTitle) > 0 {
		outputFile = config.C.To + title + "/" + part + "-" + thisTitle + ".mp4"
	} else {
		outputFile = config.C.To + title + "/" + part + ".mp4"
	}
	if _, err := os.Stat(outputFile); err == nil {
		// path/to/whatever exists
		log.Println(outputFile + " file done !")
		return
	}
	err := os.MkdirAll(config.C.To+title+"/", os.ModePerm)
	if err != nil {
		log.Println("mkdir error" + config.C.To + title + "/")
		return
	}
	if gotMp4 && gotConfig {
		copyVideo(inputFile,outputFile)
	} else if gotXlv && gotConfig {
		tools.MakeMp4(xlvs, path+inPath, outputFile)
	}

}

func handleJSON(filename string) (title, part, thisTitle string, e error) {
	title = "";
	part = "";
	thisTitle = ""
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		e = err
		return
	}
	datajson := []byte(data)
	js, err := simplejson.NewJson(datajson)
	if err != nil {
		e = err
		return
	}
	title = js.Get("title").MustString("")
	title = strings.Replace(title, "\\", "", -1)
	title = strings.Replace(title, "/", "", -1)
	title = strings.Replace(title, ":", "", -1)
	title = strings.Replace(title, "?", "", -1)
	title = strings.Replace(title, "*", "", -1)
	title = strings.Replace(title, "\"", "", -1)
	title = strings.Replace(title, "<", "", -1)
	title = strings.Replace(title, ">", "", -1)
	title = strings.Replace(title, "|", "", -1)
	part = js.Get("page_data").Get("part").MustString("")
	if part == "" {
		part = js.Get("ep").Get("index").MustString("")
		thisTitle = js.Get("ep").Get("index_title").MustString("")
		thisTitle = strings.Replace(thisTitle, "\\", "", -1)
		thisTitle = strings.Replace(thisTitle, "/", "", -1)
		thisTitle = strings.Replace(thisTitle, ":", "", -1)
		thisTitle = strings.Replace(thisTitle, "?", "", -1)
		thisTitle = strings.Replace(thisTitle, "*", "", -1)
		thisTitle = strings.Replace(thisTitle, "\"", "", -1)
		thisTitle = strings.Replace(thisTitle, "<", "", -1)
		thisTitle = strings.Replace(thisTitle, ">", "", -1)
		thisTitle = strings.Replace(thisTitle, "|", "", -1)
	}
	return
}

func copyVideo(inputFile, outputFile string) {
	log.Println(inputFile + "  ------>  " + outputFile)
	if config.C.Buf <= 0 {
		buf, err := ioutil.ReadFile(inputFile)
		if err != nil {
			log.Println("An error occurred with read:" + inputFile)
			return
		}
		out, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println("An error occurred with file opening or creation:" + outputFile)
			return
		}
		defer out.Close()
		outputWriter := bufio.NewWriter(out)
		outputWriter.Write(buf)
		outputWriter.Flush()
	} else {
		in, err := os.OpenFile(inputFile, os.O_RDONLY, 0666)
		if err != nil {
			log.Println("An error occurred with file opening:" + outputFile)
			return
		}
		defer in.Close()
		out, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println("An error occurred with file opening or creation:" + outputFile)
			return
		}
		defer out.Close()
		buf := make([]byte, config.C.Buf) //一次读取多少个字节
		bfRd := bufio.NewReader(in)
		outputWriter := bufio.NewWriter(out)
		for {
			n, err := bfRd.Read(buf)
			outputWriter.Write(buf[:n]) // n 是成功读取字节数
			if err != nil {
				//遇到任何错误立即返回，并忽略 EOF 错误信息
				if err == io.EOF {
					break
				}
				log.Println(err)
			}
		}
		outputWriter.Flush()
	}
}
