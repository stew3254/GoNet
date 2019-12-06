package main

import (
  "flag"
  "fmt"
  "net"
  //"github.com/gdamore/tcell"
  "os"
  "strconv"
)

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

//Checks valid port
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

  //Flags
  sport := *portPtr
  listen := *listenPtr
  udp := *udpPtr
  verbose := *verbosePtr

  //Client
  var client Client

  //Initialize network vars
  ip := ""
  var dport int
  var err error

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
      //Check if source port is set
      if isPort(sport) {
        laddr, _ := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(sport))
        raddr := &net.UDPAddr {IP: net.ParseIP(ip), Port: dport, Zone: ""}
        client.Con, err = net.DialUDP("udp", laddr, raddr)
      } else {
        client.Con, err = net.Dial("udp", net.JoinHostPort(ip, strconv.Itoa(dport)))
      }
    } else {
      client.Con, err = net.Dial("tcp", net.JoinHostPort(ip, strconv.Itoa(dport)))
    }
    //Check to see if the connection is valid
    if err != nil {
      fmt.Fprintln(os.Stderr, "Failed to connect")
      return
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
      client.Con, _ = net.ListenUDP("udp", addr)
    } else {
      l, err = net.Listen("tcp", net.JoinHostPort(ip, strconv.Itoa(sport)))
      if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
      }
      defer l.Close()

      //Only allow one connection for now
      client.Con, err = l.Accept()
      if verbose {
        fmt.Println("Connection from " + client.Con.RemoteAddr().String() + " succeeded!")
      }
      if err != nil {
        fmt.Fprintln(os.Stderr, err)
      }
    }
  }

  client.Closed = make(chan bool)
  client.In = os.Stdin
  client.Out = os.Stdout
  client.Handle()
}
