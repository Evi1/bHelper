package main

import (
	"io/ioutil"
	"log"
	"container/list"
	"os"
	"strings"
	"github.com/bitly/go-simplejson"
	"bufio"
	"path/filepath"
	"github.com/Evi1/bHelper/config"
	"io"
)

var routineNum chan int
var routineLimit chan int

func init() {
	routineNum = make(chan int)
	routineLimit = make(chan int, config.C.Limit)
}

func main() {
	bPath := config.C.From + "tv.danmaku.bili/download/"
	cPath := bPath
	files, _ := ioutil.ReadDir(filepath.FromSlash(cPath))
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
	files, _ := ioutil.ReadDir(filepath.FromSlash(path))
	title := ""
	part := ""
	v := ""
	inPath := ""
	for _, f := range files {
		if f.IsDir() {
			videos, _ := ioutil.ReadDir(filepath.FromSlash(path + f.Name() + "/"))
			for _, video := range videos {
				if strings.HasSuffix(video.Name(), ".mp4") {
					v = video.Name()
					inPath = f.Name() + "/"
				}
			}
		} else if strings.HasSuffix(f.Name(), ".json") {
			title, part = handleJSON(path + f.Name())
		} else if strings.HasSuffix(f.Name(), ".mp4") {
			v = f.Name()
			inPath = ""
		}
	}
	if v != "" {
		copyVideo(title, part, path, inPath, v)
	}
	<-routineLimit
	routineNum <- 1
}

func handleJSON(filename string) (string, string) {
	data, err := ioutil.ReadFile(filepath.FromSlash(filename))
	if err != nil {
		return "", ""
	}
	datajson := []byte(data)
	js, err := simplejson.NewJson(datajson)
	if err != nil {
		return "", ""
	}
	title := js.Get("title").MustString()
	part := js.Get("page_data").Get("part").MustString()
	if part == "" {
		part = js.Get("ep").Get("index").MustString()
	}
	return title, part
}

func copyVideo(title string, part string, path string, inPath string, v string) {
	inputFile := path + inPath + v;
	part = strings.TrimSpace(part)
	if len(part) <= 1 {
		part = "0" + part
	}
	outputFile := config.C.To + title + "/" + part + ".mp4"
	//oldMask := syscall.Umask(0)
	err := os.MkdirAll(filepath.FromSlash(config.C.To + title + "/"), os.ModePerm)
	if err != nil {
		log.Println("mkdir error" + config.C.To + title + "/")
		return
	}
	//syscall.Umask(oldMask)
	log.Println(inputFile + "  ------>  " + outputFile)
	if config.C.Buf <= 0 {
		buf, err := ioutil.ReadFile(filepath.FromSlash(inputFile))
		if err != nil {
			log.Println("An error occurred with read:" + v)
			return
			// panic(err.Error())
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

	/*buf, err := ioutil.ReadFile(filepath.FromSlash(inputFile))
	if err != nil {
		log.Println("An error occurred with read:" + v)
		return
		// panic(err.Error())
	}
	out, err := os.OpenFile(outputFile, os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		log.Println("An error occurred with file opening or creation:" + part + ".mp4")
		return
	}
	defer out.Close()
	outputWriter := bufio.NewWriter(out)
	outputWriter.Write(buf)
	outputWriter.Flush()*/
}
