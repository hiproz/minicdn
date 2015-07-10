package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang/groupcache"
)

var thumbNails = groupcache.NewGroup("thumbnail", 512<<20, groupcache.GetterFunc(
	func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		fileName := key
		bytes, err := generateThumbnail(fileName)
		if err != nil {
			return err
		}
		dest.SetBytes(bytes)
		return nil
	}))

func generateThumbnail(key string) ([]byte, error) {
	u, _ := url.Parse(*mirror)
	u.Path = key
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path

	state.addActiveDownload(1)
	defer state.addActiveDownload(-1)

	if *upstream == "" { // Master
		if slaveAddr, err := slaveMap.PeekSlave(); err == nil {
			u, _ := url.Parse(slaveAddr)
			u.Path = r.URL.Path
			u.RawQuery = r.URL.RawQuery
			http.Redirect(w, r, u.String(), 302)
			return
		}
	}
	fmt.Println("KEY:", key)
	var data []byte
	var ctx groupcache.Context
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
	token    = flag.String("token", "1234567890ABCDEFG", "slave and master token should be same")
)

func InitSignal() {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for {
			s := <-sig
			fmt.Println("Got signal:", s)
			if state.Closed {
				fmt.Println("Cold close !!!")
				os.Exit(1)
			}
			fmt.Println("Warm close, waiting ...")
			go func() {
				state.Close()
				os.Exit(0)
			}()
		}
	}()
}

func main() {
	flag.Parse()

	if *mirror != "" && *upstream != "" {
		log.Fatal("Can't set both -mirror and -upstream")
	}
	if *mirror == "" && *upstream == "" {
		log.Fatal("Must set one of -mirror and -upstream")
	}
	if *upstream != "" {
		if err := InitSlave(); err != nil {
			log.Fatal(err)
		}
	}
	if *mirror != "" {
		if _, err := url.Parse(*mirror); err != nil {
			log.Fatal(err)
		}
		if err := InitMaster(); err != nil {
			log.Fatal(err)
		}
	}

	InitSignal()
	fmt.Println("Hello CDN")
	http.HandleFunc("/", FileHandler)
	log.Printf("Listening on %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
