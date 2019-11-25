package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

//Connection input
func input(con *net.Conn, in *os.File, c <-chan bool) {
	reader := bufio.NewReader(in)
	scanner := bufio.NewScanner(reader)

	input := make(chan bool)
	//Look for input
	go func() {
		for {
			input <- scanner.Scan()
		}
	}()

	//Check to see if the channel is still open
	for {
		select {
		case closed := <-c:
			if closed {
				return
			}
		case <-input:
			_, err := (*con).Write(scanner.Bytes())
			if err != nil {
				return
			}
      in.Sync()
			if err != nil {
				return
			}
			(*con).Write([]byte("\n"))
		}
	}
	return
}

//Connection output
func output(con *net.Conn, out *os.File, closed chan<- bool) {
	b := make([]byte, 4096)
	defer (*con).Close()
	for true {
		n, err := (*con).Read(b)
		if err != nil {
			closed <- true
			return
		}
		fmt.Fprint(out, string(b))
		//Zero the buffer out again
		ssetn(b, 0, n)
	}
}

//Set the contents of a byte array to something
func sset(b []byte, e byte) {
	for i := 0; i < len(b); i++ {
		b[i] = e
	}
}
//Connection output

//Set the first n elements in byte array to something
func ssetn(b []byte, e byte, n int) {
	for i := 0; i < n; i++ {
		b[i] = e
	}
}

func isPort(p int) bool {
	if p < 0 || p > 65535 {
		return false
	}
	return true
}

func main() {
	//Get command line flags
	portPtr := flag.Int("p", 0, "Set the source or listening port")
	listenPtr := flag.Bool("l", false, "Listen for incoming connections")
	udpPtr := flag.Bool("u", false, "Uses UDP instead of TCP")
	verbosePtr := flag.Bool("v", false, "Verbose mode")

	flag.Parse()

	sport := *portPtr
	listen := *listenPtr
	udp := *udpPtr
	verbose := *verbosePtr

	ip := ""
	var dport int
	var err error

	//Connection
	var c net.Conn

	//Get ip and port of thing to connect to
	if !listen {
		if flag.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "Missing destination and port")
			return
		} else if flag.NArg() == 1 {
			fmt.Fprintln(os.Stderr, "Missing destination port")
			return
		} else {
			//Look up host
			addrs, err := net.LookupHost(flag.Arg(0))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			//Set ip and port
			ip = addrs[0]
			dport, err = strconv.Atoi(flag.Arg(1))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			} else if !isPort(dport) {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		//Start UDP or TCP connection
		if udp {
			c, err = net.Dial("udp", net.JoinHostPort(ip, strconv.Itoa(dport)))
		} else {
			c, err = net.Dial("tcp", net.JoinHostPort(ip, strconv.Itoa(dport)))
		}
		//Check to see if the connection is valid
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to connect")
			return
		}
		if verbose {
			fmt.Println("Connected to " + c.RemoteAddr().String() + " successfully!")
		}
	} else {
		//Make sure port is valid
		if !isPort(sport) {
			fmt.Fprintln(os.Stderr, "Port number is invalid")
			return
		}

		//Check args
		if flag.NArg() > 0 {
			ip = flag.Arg(0)
		}

		//UDP vs TCP listener
		var l net.Listener
		if udp {
			//TODO fix this
			addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, strconv.Itoa(sport)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			c, _ = net.ListenUDP("udp", addr)
		} else {
			l, err = net.Listen("tcp", net.JoinHostPort(ip, strconv.Itoa(sport)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			defer l.Close()

			//Only allow one connection for now
			c, err = l.Accept()
			if verbose {
				fmt.Println("Connection from " + c.RemoteAddr().String() + " succeeded!")
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	closed := make(chan bool)
	go output(&c, os.Stdout, closed)
	input(&c, os.Stdin, closed)
}
