package ymp3d

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/knadh/go-get-youtube/youtube"
	"github.com/sirupsen/logrus"
)

type Page struct {
	Title   string
	Message string
}

type Server struct {
	conf    *Config
	log     *logrus.Logger
	tempDir string
}

func NewServer() (s *Server) {
	var err error
	s = new(Server)
	s.conf = newConfig()
	s.setupLogger()
	s.tempDir, err = ioutil.TempDir("", ".ymp3d-")
	if err != nil {
		panic(err)
	}
	s.tempDir += "/"
	return s
}

func (s *Server) Run() {
	s.log.Info("Start ymp3d Server")
	defer os.RemoveAll(s.tempDir)
	http.HandleFunc("/youtube/", s.getMp3Handler)
	http.ListenAndServe(":"+s.conf.Server.Port, nil)
}

func (s *Server) setupLogger() {
	s.log = logrus.New()
	f, err := os.OpenFile(s.conf.Log.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	s.log.Out = f
	level := s.conf.Log.Level
	switch level {
	case "info":
		s.log.Level = logrus.InfoLevel
	case "warn":
		s.log.Level = logrus.WarnLevel
	case "errror":
		s.log.Level = logrus.ErrorLevel
	case "panic":
		s.log.Level = logrus.PanicLevel
	case "debug":
		s.log.Level = logrus.DebugLevel
	default:
		s.log.Level = logrus.InfoLevel
	}
}

func (s *Server) getMp3Handler(w http.ResponseWriter, r *http.Request) {
	const tmpl = `
<!DOCTYPE html>
<html>
<head>
<meta http-equiv="content-type" content="text/html; charset=utf-8">
<title>{{.Title}}</title>
</head>
<body>
{{.Message}}
</body>
<html>
`
	videoId := strings.Split(r.RequestURI, "/")[2]
	go s.getMp3(videoId)
	p := Page{"ymp3d", "start downloading " + videoId}
	t, _ := template.New("download").Parse(tmpl)
	t.Execute(w, p)
	return
}

func (s *Server) getMp3(videoId string) {
	l := s.log.WithField("videoId", videoId)
	l.Info("start get mp3 file")
	l.Debug("get metadata")
	video, filename, err := s.getVideoInfo(videoId)
	if err != nil {
		l.Warn("failed: get meta data")
		return
	}
	l.Debug("success: get metadata")
	if _, err = os.Stat(s.tempDir + filename); err == nil {
		l.Debug("already start downloading")
		return
	}
	l.Debug("download video")
	err = video.Download(0, s.tempDir+filename)
	defer os.Remove(filename)
	if err != nil {
		l.Warn("failed: download")
		return
	}
	l.Debug("success: download")
	l.Debug("convert to mp3")
	err = s.convertToMp3(filename)
	if err != nil {
		l.Warn("failed: comvert to mp3")
		return
	}
	l.Info("success: get mp3 file")
}

func (s *Server) getVideoInfo(videoId string) (video youtube.Video, filename string, err error) {
	video, err = youtube.Get(videoId)
	if err != nil {
		return video, "", err
	}
	codec := video.GetExtension(0)
	filename = "." + video.Title + "." + codec
	return video, filename, nil
}

func (s *Server) convertToMp3(videoname string) (err error) {
	a := strings.Split(videoname, ".")
	name := strings.Join(a[0:len(a)-1], "")
	mp3file := name + ".mp3"
	tmpfile := "." + mp3file
	l := s.log.WithFields(logrus.Fields{
		"tmppath": s.tempDir + tmpfile,
		"mp3path": s.conf.Server.DownloadDir + "/" + mp3file,
	})

	l.Debug("start comvert video file -> tmp mp3file ")
	cmd := exec.Command("avconv", "-i", s.tempDir+videoname, s.tempDir+tmpfile)
	defer os.Remove(s.tempDir + tmpfile)
	err = cmd.Run()
	if err != nil {
		return err
	}
	l.Debug("rename tmp tmppath -> mp3path")
	err = os.Rename(s.tempDir+tmpfile, s.conf.Server.DownloadDir+"/"+mp3file)
	if err != nil {
		return err
	}
	return nil
}
