package main

import (
	"fmt"
	cns "github.com/moriyoshi/cfnetservices-go"
	"os"
	"time"
)

func main() {
	ns := cns.CFNetServiceCreate("local.", "_test._tcp", "test", 65535)
	if !cns.CFNetServiceSetTXTData(ns, []byte{11, 'h', 'e', 'l', 'l', 'o', '=', 'w', 'o', 'r', 'l', 'd'}) {
		fmt.Fprintf(os.Stderr, "Failed to set TXT data\n")
	}
	d, _ := time.ParseDuration("10s")
	t := time.AfterFunc(d, func() { cns.CFNetServiceCancel(ns) })
	ch := make(chan struct{})
	go func() {
		<-ch
		fmt.Printf("Registered!\n")
	}()
	err := cns.CFNetServiceRegisterWithOptions(ns, 0, ch)
	if err != nil {
		t.Stop()
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer cns.CFNetServiceRelease(ns)
}
