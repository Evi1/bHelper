package tools

import (
	"strings"
	"os"
	"container/list"
	"strconv"
	"log"
	"os/exec"
	"runtime"
	"path/filepath"
)

func findMin(l *list.List) *list.Element {
	min := l.Front()
	for i := l.Front().Next(); i != nil; i = i.Next() {
		if (getNum(min.Value.(string)) > getNum(i.Value.(string))) {
			min = i
		}
	}
	return min
}

func CheckFLV(videos []os.FileInfo) (l *list.List, r bool) {
	r = false;
	l = list.New();
	ll := list.New()
	for _, video := range videos {
		if strings.HasSuffix(video.Name(), ".flv") || strings.HasSuffix(video.Name(), ".blv") {
			r = true;
			ll.PushBack(video.Name())
		}
	}
	for ll.Len() > 0 {
		t := findMin(ll)
		ll.Remove(t)
		l.PushBack(t.Value.(string))
	}
	return
}

func MakeMp4(l *list.List, p string, t string) {
	out, err := os.OpenFile(p+"file", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("An error occurred with file opening or creation:" + p + "file")
		return
	}
	defer out.Close()
	if runtime.GOOS == "windows" {
		//log.Println(runtime.GOOS)
		p = filepath.FromSlash(p)
	}
	//log.Println(p)
	for i := l.Front(); i != nil; i = i.Next() {
		log.Println("file '" + p + i.Value.(string) + "'\n")
		out.WriteString("file '" + p + i.Value.(string) + "'\n")
	}
	log.Println(p + "file" + "  ------>  " + t)
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", p+"file", "-c", "copy", t /*p + "output.mp4"*/)
	o, e := cmd.CombinedOutput()
	if e != nil {
		log.Println("cmdout:" + string(o) + "error:" + e.Error())
	}
	//log.Println(string(o))
}

func getNum(n string) (x int) {
	var e error
	m := strings.Split(n, ".")[0]
	x, e = strconv.Atoi(m)
	if e != nil {
		x = -1
		log.Println("An error occurred with vlc file name:" + n + " : " + e.Error())
	}
	return
}
