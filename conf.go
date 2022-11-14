package main

type conf struct {
	maxResourceSize int
	maxNumResources int
	useTLS          bool
	selfSign        bool
	pemFile         string
	keyFile         string
}
