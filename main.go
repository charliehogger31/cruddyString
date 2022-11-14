package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/ini.v1"
)

type master struct{}

var cache []string
var config conf
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
				_, err := fmt.Fprintf(w, html.EscapeString(cache[i]))
				cacheMutex.Unlock()
				if err != nil {
					w.WriteHeader(500)
					return
				}
			}
		}
	} else if req.Method == "POST" {
		cacheMutex.Lock()
		if config.maxNumResources != 0 && len(cache) >= config.maxNumResources {
			w.WriteHeader(500)
			cacheMutex.Unlock()
			return
		}
		cacheMutex.Unlock()
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

			if len(inp) > config.maxResourceSize && config.maxResourceSize != 0 {
				w.WriteHeader(500)
				cacheMutex.Unlock()
				return
			}

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
				_, err := fmt.Fprintf(w, html.EscapeString(old))
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

						if len(inp) > config.maxResourceSize && config.maxResourceSize != 0 {
							w.WriteHeader(500)
							cacheMutex.Unlock()
							return
						}

						cache[i] = inp
						cacheMutex.Unlock()
						_, err := fmt.Fprintf(w, html.EscapeString(old))
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

		config.maxResourceSize, err = cfg.Section("memory").Key("maxresourcesize").Int()
		config.maxNumResources, err = cfg.Section("memory").Key("maxnumresources").Int()

		config.useTLS, err = cfg.Section("tls").Key("usetls").Bool()
		config.selfSign, err = cfg.Section("tls").Key("selfsign").Bool()

		if !config.selfSign {
			config.pemFile = cfg.Section("tls").Key("certfile").String()
			config.keyFile = cfg.Section("tls").Key("keyfile").String()
		}

		if err != nil {
			fmt.Println("Error with ini file: ", err)
			os.Exit(1)
		}
	} else {
		config.maxResourceSize = 0
		config.maxNumResources = 0 // Set to infinity
		config.useTLS = false
		config.selfSign = false
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
			inp := fScan.Text()
			if len(inp) <= config.maxResourceSize && config.maxResourceSize != 0 {
				cache = append(cache, inp)
			}
			if len(cache) >= config.maxNumResources && config.maxNumResources != 0 {
				break
			}
		}
		cacheMutex.Unlock()

		errf := f.Close()
		if errf != nil {
			fmt.Println("Error on config file close: ", errf)
		}
	}

	if config.selfSign {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			fmt.Println("Error generating ECDSA privatekey", err)
			os.Exit(1)
		}
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			fmt.Println("Error with serial number generation", err)

		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"Test Corp"},
			},
			DNSNames:  []string{"localhost"},
			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(3 * time.Hour),

			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		derByte, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			fmt.Println("Error creating self signed certificate", err)
			os.Exit(1)
		}

		pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derByte})
		if pemCert == nil {
			fmt.Println("Failed PEM Encoding")
			os.Exit(1)
		}

		if err := os.WriteFile("cert.pem", pemCert, 0644); err != nil {
			fmt.Println("Error writing PEM cert", err)
			os.Exit(1)
		}

		config.pemFile = "cert.pem"

		privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			fmt.Println("Error on key marshal", err)
			os.Exit(1)
		}

		pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
		if pemKey == nil {
			fmt.Println("Pem Fail for key")
		}

		if err = os.WriteFile("key.pem", pemKey, 0600); err != nil {
			fmt.Println("Error writing PEM key", err)
		}
		config.keyFile = "key.pem"

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
	sv1.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	sv1.ReadTimeout = 30 * time.Second
	sv1.WriteTimeout = 30 * time.Second
	sv1.IdleTimeout = 30 * time.Second
	sv1.MaxHeaderBytes = 0
	sv1.TLSNextProto = nil
	sv1.ConnState = nil
	sv1.ErrorLog = nil
	sv1.BaseContext = nil
	sv1.ConnState = nil

	if config.useTLS {
		err := sv1.ListenAndServeTLS(config.pemFile, config.keyFile)
		if err != nil {
			fmt.Println("Listen and Server Error: ", err)
		}
	} else {
		err := sv1.ListenAndServe()
		if err != nil {
			fmt.Println("Listen and Server Error: ", err)
		}
	}
}
