package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type master struct{}

var cache []string
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
