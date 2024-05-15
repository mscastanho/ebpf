// This program demonstrates attaching a fentry eBPF program to
// tcp_close and reading the RTT from the TCP socket using CO-RE helpers.
// It prints the IPs/ports/RTT information
// once the host closes a TCP connection.
// It supports only IPv4 for this example.
//
// Sample output:
//
// examples# go run -exec sudo ./tcprtt
// 2022/03/19 22:30:34 Src addr        Port   -> Dest addr       Port   RTT
// 2022/03/19 22:30:36 10.0.1.205      50578  -> 117.102.109.186 5201   195
// 2022/03/19 22:30:53 10.0.1.205      0      -> 89.84.1.178     9200   30
// 2022/03/19 22:30:53 10.0.1.205      36022  -> 89.84.1.178     9200   28

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mscastanho/ebpf/internal"
	"github.com/mscastanho/ebpf/link"
	"github.com/mscastanho/ebpf/ringbuf"
	"github.com/mscastanho/ebpf/rlimit"
)

//go:generate go run github.com/mscastanho/ebpf/cmd/bpf2go -type event bpf tcprtt.c -- -I../headers

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	link, err := link.AttachTracing(link.TracingOptions{
		Program: objs.bpfPrograms.TcpClose,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer link.Close()

	rd, err := ringbuf.NewReader(objs.bpfMaps.Events)
	if err != nil {
		log.Fatalf("opening ringbuf reader: %s", err)
	}
	defer rd.Close()

	log.Printf("%-15s %-6s -> %-15s %-6s %-6s",
		"Src addr",
		"Port",
		"Dest addr",
		"Port",
		"RTT",
	)
	go readLoop(rd)

	// Wait
	<-stopper
}

func readLoop(rd *ringbuf.Reader) {
	// bpfEvent is generated by bpf2go.
	var event bpfEvent
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				log.Println("received signal, exiting..")
				return
			}
			log.Printf("reading from reader: %s", err)
			continue
		}

		// Parse the ringbuf event entry into a bpfEvent structure.
		if err := binary.Read(bytes.NewBuffer(record.RawSample), internal.NativeEndian, &event); err != nil {
			log.Printf("parsing ringbuf event: %s", err)
			continue
		}

		log.Printf("%-15s %-6d -> %-15s %-6d %-6d",
			intToIP(event.Saddr),
			event.Sport,
			intToIP(event.Daddr),
			event.Dport,
			event.Srtt,
		)
	}
}

// intToIP converts IPv4 number to net.IP
func intToIP(ipNum uint32) net.IP {
	ip := make(net.IP, 4)
	internal.NativeEndian.PutUint32(ip, ipNum)
	return ip
}
