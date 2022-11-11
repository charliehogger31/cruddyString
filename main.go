package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/ini.v1"
)

type master struct{}

var cache []string
var maxRsSize int
var maxNRs int
var cacheMutex sync.Mutex

func (m *master) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		i, err := strconv.Atoi(req.RequestURI[1:])
		cacheMutex.Lock()
		clen := len(cache)
		cacheMutex.Unlock()
		if err != nil {
			w.WriteHeader(400)
			return
		} else {
			if i >= clen {
				w.WriteHeader(400)
				return
			} else {
				cacheMutex.Lock()
				_, err := fmt.Fprintf(w, cache[i])
				cacheMutex.Unlock()
				if err != nil {
					w.WriteHeader(500)
					return
				}
			}
		}
	} else if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			w.WriteHeader(500)
			return
		}

		var inp = req.FormValue("inputdata")
		if inp == "" {
			w.WriteHeader(400)
			return
		} else {
			cacheMutex.Lock()
			cache = append(cache, inp)
			_, err := fmt.Fprintf(w, strconv.Itoa(len(cache)-1))
			if err != nil {
				w.WriteHeader(500)
				return
			}
			cacheMutex.Unlock()
		}
	} else if req.Method == "DELETE" {
		i, err := strconv.Atoi(req.RequestURI[1:])
		cacheMutex.Lock()
		clen := len(cache)
		cacheMutex.Unlock()
		if err != nil {
			w.WriteHeader(400)
		} else {
			if i >= clen {
				w.WriteHeader(400)
				return
			} else {
				cacheMutex.Lock()
				old := cache[i]
				cache[i] = ""
				cacheMutex.Unlock()
				_, err := fmt.Fprintf(w, old)
				if err != nil {
					w.WriteHeader(500)
					return
				}
			}
		}
	} else if req.Method == "PATCH" {
		var ix int
		for ix = 0; ix < len(req.RequestURI[1:]); ix++ {
			if req.RequestURI[1:][ix] == '?' {
				break
			}
		}

		i, err := strconv.Atoi(req.RequestURI[1 : ix+1])
		cacheMutex.Lock()
		clen := len(cache)
		cacheMutex.Unlock()
		if err != nil {
			w.WriteHeader(400)
			return
		} else {
			if i >= clen {
				w.WriteHeader(400)
				return
			} else {
				err := req.ParseForm()
				if err != nil {
					w.WriteHeader(400)
					return
				} else {
					inp := req.FormValue("inputdata")
					if inp == "" {
						w.WriteHeader(400)
					} else {
						cacheMutex.Lock()
						old := cache[i]
						cache[i] = inp
						cacheMutex.Unlock()
						_, err := fmt.Fprintf(w, old)
						if err != nil {
							w.WriteHeader(500)
							return
						}
					}
				}
			}
		}
	}
}

func main() {
	// Initiation
	if len(os.Args) >= 2 {
		cfg, err := ini.Load(os.Args[1])
		if err != nil {
			fmt.Println("Error loading ini: ", err)
			os.Exit(1)
		}

		fmt.Println("Starting with max resource size: ", cfg.Section("memory").Key("maxresourcesize").String())
		fmt.Println("Starting with max # of resources: ", cfg.Section("memory").Key("maxnumresources").String())

		maxRsSize, err = cfg.Section("memory").Key("maxresourcesize").Int()
		maxNRs, err = cfg.Section("memory").Key("maxnumresources").Int() // TODO: Implement into checks and balances

		if err != nil {
			fmt.Println("Error with ini file: ", err)
			os.Exit(1)
		}
	}

	if len(os.Args) >= 3 {
		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Println("Error on config file open: ", err)
			os.Exit(1)
		}

		fScan := bufio.NewScanner(f)
		fScan.Split(bufio.ScanLines)

		cacheMutex.Lock()
		for fScan.Scan() {
			cache = append(cache, fScan.Text())
		}
		cacheMutex.Unlock()

		errf := f.Close()
		if errf != nil {
			fmt.Println("Error on config file close: ", errf)
		}
	}

	// Server Domain
	var sv1 http.Server
	defer func(sv1 *http.Server) {
		err := sv1.Close()
		if err != nil {
			return
		}
	}(&sv1)
	sv1.Addr = ":8080"
	sv1.Handler = &master{}
	sv1.TLSConfig = nil
	sv1.ReadTimeout = 30 * time.Second
	sv1.WriteTimeout = 30 * time.Second
	sv1.IdleTimeout = 30 * time.Second
	sv1.MaxHeaderBytes = 0
	sv1.TLSNextProto = nil
	sv1.ConnState = nil
	sv1.ErrorLog = nil
	sv1.BaseContext = nil
	sv1.ConnState = nil

	err := sv1.ListenAndServe()

	if err != nil {
		fmt.Println("Listen and Server Error: ", err)
	}
}
