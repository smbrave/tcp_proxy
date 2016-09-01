package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strings"
	"sync/atomic"
)

type TCPProxy struct {
	localAddress  string
	remoteAddress string
	exitChan      chan struct{}
	listener      net.Listener
	maxCopy       uint32
}

func NewTCPProxy(local, retmote string) *TCPProxy {
	return &TCPProxy{
		localAddress:  local,
		remoteAddress: retmote,
		exitChan:      make(chan struct{}),
		maxCopy:       1024,
	}
}

func (tp *TCPProxy) Start() {
	go tp.Accept()
}

func (tp *TCPProxy) Stop() {
	if tp.exitChan != nil {
		close(tp.exitChan)
		tp.exitChan = nil
	}
	tp.listener.Close()
}

func (tp *TCPProxy) Accept() {
	serverAddr, err := net.ResolveTCPAddr("tcp", tp.localAddress)
	if err != nil {
		panic(fmt.Sprintf("ERROR: ResolveTCPAddr err:%s", err.Error()))
	}

	listener, err := net.ListenTCP("tcp", serverAddr)
	if err != nil {
		panic(fmt.Sprintf("ERROR: ListenTCP err:%s", err.Error()))
	}
	tp.listener = listener

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.Printf("NOTICE: temporary Accept() failure - %s", err.Error())
				runtime.Gosched()
				continue
			}

			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("ERROR: listener.Accept() - %s", err.Error())
			}
			break
		}

		clientConn.(*net.TCPConn).SetNoDelay(true)
		go tp.IOLoop(clientConn)
	}
}

func (tp *TCPProxy) Transform(src, dst net.Conn, exitChan chan int) {
	maxCopy := int64(atomic.LoadUint32(&tp.maxCopy))
	for {
		select {
		case <-exitChan:
			goto exit
		default:
		}
		_, err := io.CopyN(dst, src, maxCopy)
		if err != nil {
			log.Printf("ERROR: %s -> %s read error:%s", src.RemoteAddr(), dst.RemoteAddr(), err.Error())
			goto exit
		}
		//log.Printf("INFO: transform %s -> %s (%d)bytes", src.RemoteAddr(), dst.RemoteAddr(), number)
	}
exit:
	exitChan <- 1
}

func (tp *TCPProxy) IOLoop(localConn net.Conn) {
	defer localConn.Close()
	remoteConn, err := net.Dial("tcp", tp.remoteAddress)
	if err != nil {
		log.Printf("ERROR: Dial tcp %s err:%s", tp.remoteAddress, err.Error())
		return
	}
	defer remoteConn.Close()

	log.Printf("net tcp conn:%s->%s", localConn.RemoteAddr().String(), remoteConn.RemoteAddr().String())

	exit := make(chan int, 1)

	go tp.Transform(localConn, remoteConn, exit)
	go tp.Transform(remoteConn, localConn, exit)

	<-exit
}
