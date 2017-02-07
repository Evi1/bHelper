package tools

import (
	"strings"
	"os"
	"container/list"
	"strconv"
	"log"
	"os/exec"
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
	l = list.New(); ll := list.New()
	for _, video := range videos {
		if strings.HasSuffix(video.Name(), ".flv") {
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

func MakeMp4(l *list.List, p string) {
	out, err := os.OpenFile(p + "file", os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		log.Println("An error occurred with file opening or creation:" + p + "file")
		return
	}
	defer out.Close()
	for i := l.Front(); i != nil; i = i.Next() {
		log.Println("file '" + p + i.Value.(string) + "'\n")
		out.WriteString("file '" + p + i.Value.(string) + "'\n")
	}
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", p + "file", "-c", "copy", p + "output.mp4")
	log.Println("ffmpeg -f concat -safe 0 -i " + p + "file -c copy " + p + "output.mp4")
	o, e := cmd.CombinedOutput()
	if e != nil {
		log.Println(err)
	}
	log.Println(string(o))
}

func getNum(n string) (x int) {
	var e error
	m := strings.Split(n, ".")[0]
	x, e = strconv.Atoi(m)
	if e != nil {
		log.Println("An error occurred with vlc file name:" + n + " : " + e.Error())
	}
	return
}