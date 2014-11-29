package ymp3d

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/knadh/go-get-youtube/youtube"
)

func convertToMp3(filename string) (err error) {
	a := strings.Split(filename, ".")
	name := strings.Join(a[0:len(a)-1], "")
	fmt.Printf("convert %s to mp3", name)
	cmd := exec.Command("avconv", "-i", filename, name+".mp3")
	return cmd.Run()
}

func getMp3Handler(w http.ResponseWriter, r *http.Request) {
	videoId := strings.Split(r.RequestURI, "/")[2]
	filename, err := download(videoId)
	if err != nil {
		fmt.Printf("failed: download %s", videoId)
		return
	}
	err = convertToMp3(filename)
	if err != nil {
		fmt.Printf("failed: comvert %s to mp3", videoId)
		return
	}
}

func download(videoId string) (filename string, err error) {
	fmt.Printf("start download %s\n", videoId)
	video, err := youtube.Get(videoId)
	if err != nil {
		return "", err
	}
	fmt.Printf("success meta data %s\n", videoId)

	codec := video.GetExtension(0)
	filename = video.Title + "." + codec

	err = video.Download(0, filename)
	if err != nil {
		return "", err
	}
	fmt.Printf("finish download\n")

	return filename, nil
}

func Run() {
	http.HandleFunc("/youtube/", getMp3Handler)
	http.ListenAndServe(":80", nil)
}
