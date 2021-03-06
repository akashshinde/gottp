// +build !appengine

package gottp

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	conf "github.com/Simversity/gottp/conf"
)

func cleanAddr(addr string) {
	err := os.Remove(addr)
	if err != nil {
		log.Panic(err)
	}
}

func interrupt_cleanup(addr string) {
	if strings.Index(addr, "/") != 0 {
		return
	}

	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	//NOTE: Capture every Signal right now.
	//signal.Notify(sigchan)

	s := <-sigchan
	log.Println("Exiting Program. Got Signal: ", s)

	// do last actions and wait for all write operations to end
	cleanAddr(addr)
	os.Exit(0)
}

var SysInitChan = make(chan bool, 1)

var settings conf.Config

func parseCLI() {
	cfgPath, unixAddr := conf.CliArgs()
	settings.MakeConfig(cfgPath)

	if unixAddr != "" {
		settings.Gottp.Listen = unixAddr
	}
}

func MakeServer(cfg conf.Configurer) {
	cfgPath, unixAddr := conf.CliArgs()
	cfg.MakeConfig(cfgPath)

	settings.Gottp = *cfg.GetGottpConfig()

	if unixAddr != "" {
		settings.Gottp.Listen = unixAddr
	}

	makeServer()
}

func DefaultServer() {
	parseCLI()
	makeServer()
}

func makeServer() {
	addr := settings.Gottp.Listen

	SysInitChan <- true

	var serverError error
	if addr != "" {
		log.Println("Listening on " + addr)
	}

	if strings.Index(addr, "/") == 0 {
		listener, err := net.Listen("unix", addr)
		if err != nil {
			c, err := net.Dial("unix", addr)

			if c != nil {
				defer c.Close()
			}

			if err != nil {
				log.Println("The socket does not look like consumed. Erase ?")
				cleanAddr(addr)
				listener, err = net.Listen("unix", addr)
			} else {
				log.Fatal("Cannot start Server. Address is already in Use.", err)
				os.Exit(0)
			}
		}

		go interrupt_cleanup(addr)
		os.Chmod(addr, 0770)
		serverError = http.Serve(listener, nil)
	} else {
		serverError = http.ListenAndServe(addr, nil)
	}

	if serverError != nil {
		log.Println(serverError)
	}
}
