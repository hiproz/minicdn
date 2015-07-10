package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/golang/groupcache"
)

var thumbNails = groupcache.NewGroup("thunbnail", 512<<20, groupcache.GetterFunc(
	func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		fileName := key
		bytes, err := generateThumbnail(fileName)
		if err != nil {
			return err
		}
		dest.SetBytes(bytes)
		return nil
	}))

func generateThumbnail(filename string) ([]byte, error) {
	resp, err := http.Get("http://10.246.13.180:5000" + filename)
	if err != nil {
		return nil, err
	}
	println("RRR")
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	var ctx groupcache.Context
	key := r.URL.Path
	fmt.Println("KEY:", key)
	var data []byte
	err := thumbNails.Get(ctx, key, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var modTime time.Time = time.Now()

	rd := bytes.NewReader(data)
	http.ServeContent(w, r, filepath.Base(key), modTime, rd)
}

var (
	mirror   = flag.String("mirror", "", "Mirror Web Base URL")
	logfile  = flag.String("log", "-", "Set log file, default STDOUT")
	upstream = flag.String("upstream", "", "Server base URL, conflict with -mirror")
	address  = flag.String("addr", ":5000", "Listen address")
)

func main() {
	flag.Parse()

	if *mirror != "" && *upstream != "" {
		log.Fatal("Can't set both -mirror and -upstream")
	}
	if *upstream != "" {
		if err := InitSlave(); err != nil {
			log.Fatal(err)
		}
	}
	if *mirror != "" {
		if err := InitMaster(); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Hello CDN")
	http.HandleFunc("/", FileHandler)
	log.Fatal(http.ListenAndServe(*address, nil))
}
