package procutil

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForSigterm() os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		sig := <-ch
		if sig == syscall.SIGHUP {
			// Prevent from the program stop on SIGHUP
			continue
		}
		// Stop listening for SIGINT and SIGTERM signals,
		// so the app could be interrupted by sending these signals again
		// in the case if the caller doesn't finish the app gracefully.
		signal.Stop(ch)
		return sig
	}
}
