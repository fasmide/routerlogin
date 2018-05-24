package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/fasmide/routerlogin/conntrack"
	"github.com/fasmide/routerlogin/daemon"
)

func main() {

	d := daemon.Daemon{}

	listener, err := net.Listen("unix", "/tmp/hello")
	if err != nil {
		panic(err)
	}
	d.AddStore(&conntrack.StateStore{})

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		// Block until a signal is received.
		s := <-c
		fmt.Println("Interrupted:", s)
		listener.Close()
	}()

	d.Accept(listener)

}
