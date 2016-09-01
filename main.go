package main

import (
	"flag"
	"fmt"
)

var buildTime string
var binaryVersion string
var gitRevision string
var appName string = "tcp_proxy"

func version() string {
	return fmt.Sprintf("%s v%s built(%s) git(%s)", appName, binaryVersion, buildTime, gitRevision)
}

var (
	flag_version = flag.Bool("v", false, "show version info")
	flag_local   = flag.String("local_addr", "0.0.0.0:9799", "local address")
	flag_remote  = flag.String("remote_addr", "127.0.0.1:60086", "remote address")
)

func main() {
	flag.Parse()
	if *flag_version {
		fmt.Printf("%s\n", version())
		return
	}
	tcp_proxy := NewTCPProxy(*flag_local, *flag_remote)
	tcp_proxy.Start()

	select {}
}
