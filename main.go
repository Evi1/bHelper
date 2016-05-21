package main

import (
	"io/ioutil"
	"fmt"
	"container/list"
	"os"
	"strings"
	"github.com/bitly/go-simplejson"
	"bufio"
	//"syscall"
	"path/filepath"
)

var bPath string
var cPath string
var l1 *list.List

func main(){
	bPath = "./tv.danmaku.bili/download/"
	cPath = bPath
	files, _ := ioutil.ReadDir(filepath.FromSlash(cPath))
	l1 = list.New()
	for _, f := range files {
		if f.IsDir(){
			l1.PushBack(f)
		}
	}
	for l1.Len() > 0 {
		t := l1.Front()
		l1.Remove(t)
		x := t.Value.(os.FileInfo)
		cPath=bPath+x.Name()+"/"
		files, _ := ioutil.ReadDir(cPath)
		for _, f := range files {
			if f.IsDir(){
				handle(f)
			}
		}
	}
}

func handle(f os.FileInfo)  {
	path := cPath +f.Name() + "/"
	files, _ := ioutil.ReadDir(filepath.FromSlash(path))
	var title string
	var part string
	var v string
	var inPath string
	for _, f := range files {
		if f.IsDir(){
			videos,_:= ioutil.ReadDir(filepath.FromSlash(path+f.Name()+"/"))
			for _, video := range videos{
				if strings.HasSuffix(video.Name(),".mp4"){
					v = video.Name()
					inPath = f.Name()+"/"
				}
			}
		}else if strings.HasSuffix(f.Name(),".json"){
			title,part=handleJSON(path+f.Name())
		}else if strings.HasSuffix(f.Name(),".mp4"){
			v = f.Name()
			inPath=""
		}
	}
	copyVideo(title,part,path,inPath,v)
}

func handleJSON(filename string) (string, string)  {
	data, err := ioutil.ReadFile(filepath.FromSlash(filename))
	if err != nil{
		return "",""
	}
	datajson := []byte(data)
	js, err := simplejson.NewJson(datajson)
	if err != nil {
		return "",""
	}
	title:=js.Get("title").MustString()
	part:=js.Get("page_data").Get("part").MustString()
	if part == ""{
		part=js.Get("ep").Get("index").MustString()
	}
	return title,part
}

func copyVideo(title string, part string,path string,inPath string, v string){
	inputFile := path+inPath+v;
	outputFile := "./bilibili/"+title+"/"+part+".mp4"
	//oldMask := syscall.Umask(0)
	os.MkdirAll(filepath.FromSlash("./bilibili/"+title+"/"),os.ModePerm)
	//syscall.Umask(oldMask)
	fmt.Println(inputFile+"  ------>  "+outputFile)
	buf, err := ioutil.ReadFile(filepath.FromSlash(inputFile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
		// panic(err.Error())
	}
	out,err:=os.OpenFile(outputFile,os.O_WRONLY|os.O_CREATE,0666)
	if err!=nil{
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer out.Close()
	outputWriter := bufio.NewWriter(out)
	outputWriter.Write(buf)
	outputWriter.Flush()
}
