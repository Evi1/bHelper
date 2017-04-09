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
				go handle(cPath, f)
			}
		}
	}
	for i := 0; i < num; i++ {
		<-routineNum
	}
}

func handle(cPath string, f os.FileInfo) {
	path := cPath + f.Name() + "/"
	files, _ := ioutil.ReadDir(path)
	title := ""; part := ""; v := ""; thisTitle := ""; inPath := ""
	isV := false; isM := false
	for _, f := range files {
		if isV&&isM {
			break;
		}
		if f.IsDir() {
			videos, _ := ioutil.ReadDir(path + f.Name() + "/")
			l, r := tools.CheckFLV(videos)
			if r {
				tools.MakeMp4(l, path + f.Name() + "/")
				videos, _ = ioutil.ReadDir(path + f.Name() + "/")
			}

			for _, video := range videos {
				if strings.HasSuffix(video.Name(), ".mp4") {
					v = video.Name()
					inPath = f.Name() + "/"
					isV = true
					break
				}
			}
		} else if strings.HasSuffix(f.Name(), ".json") {
			var e error
			title, part, thisTitle, e = handleJSON(path + f.Name())
			if e != nil {
				log.Println("An error occurred with json file:" + f.Name() + " : " + e.Error())
				<-routineLimit
				routineNum <- 1
				return
			}
			isM = true;
		} else if strings.HasSuffix(f.Name(), ".mp4") {
			v = f.Name()
			inPath = ""
			isV = true
		}
	}
	if isV&&isM {
		copyVideo(title, part, path, inPath, v, thisTitle)
	}
	<-routineLimit
	routineNum <- 1
}

func handleJSON(filename string) (title, part, thisTitle string, e error) {
	title = ""; part = ""; thisTitle = ""
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
	part = js.Get("page_data").Get("part").MustString("")
	if part == "" {
		part = js.Get("ep").Get("index").MustString("")
		thisTitle = js.Get("ep").Get("index_title").MustString("")
	}
	return
}

func copyVideo(title, part, path, inPath, v, thisTitle string) {
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
		log.Println(outputFile + " file exitsts !")
		return
	}
	//oldMask := syscall.Umask(0)
	err := os.MkdirAll(config.C.To + title + "/", os.ModePerm)
	if err != nil {
		log.Println("mkdir error" + config.C.To + title + "/")
		return
	}
	//syscall.Umask(oldMask)
	log.Println(inputFile + "  ------>  " + outputFile)
	if config.C.Buf <= 0 {
		buf, err := ioutil.ReadFile(inputFile)
		if err != nil {
			log.Println("An error occurred with read:" + v)
			return
		}
		out, err := os.OpenFile(outputFile, os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			log.Println("An error occurred with file opening or creation:" + part + ".mp4")
			return
		}
		defer out.Close()
		outputWriter := bufio.NewWriter(out)
		outputWriter.Write(buf)
		outputWriter.Flush()
	} else {
		in, err := os.OpenFile(inputFile, os.O_RDONLY, 0666)
		if err != nil {
			log.Println("An error occurred with file opening:" + part + ".mp4")
			return
		}
		defer in.Close()
		out, err := os.OpenFile(outputFile, os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			log.Println("An error occurred with file opening or creation:" + part + ".mp4")
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


