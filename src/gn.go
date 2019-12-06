package main

import (
  "bufio"
  "flag"
  "io"
  "fmt"
  "net"
  //"github.com/gdamore/tcell"
  "os"
  "strconv"
)

//Connection input
func input(con *net.Conn, in *os.File, conChan <-chan *net.Conn, closedChan <-chan bool) {
  reader := bufio.NewReader(in)
  scanner := bufio.NewScanner(reader)

  inputChan := make(chan bool)
  sync := make(chan bool)
  //Look for input
  go func() {
    for {
      inputChan <- scanner.Scan()
      //Sync the threads back up
      <-sync
    }
  }()

  //Check to see if the channel is still open
  for {
    select {
    case newConn := <-conChan:
      con = newConn
    case closed := <-closedChan:
      if closed {
        return
      }
    case <-inputChan:
      _, err := (*con).Write(scanner.Bytes())
      if err != nil {
        return
      }
      in.Sync()
      //Something used to sync the two threads
      sync <- true
      if err != nil {
        return
      }
      (*con).Write([]byte("\n"))
    }
  }
  return
}

//Connection output
func output(con *net.Conn, out *os.File, conChan <-chan *net.Conn, closedChan chan<- bool) {
  //Remember to close the connection
  defer (*con).Close()

  reader := bufio.NewReader(out)
  inputChan := make(chan int)
  sync := make(chan bool)
  b := make([]byte, 4096)

  //Look for input
  go func() {
    for {
      n, err := reader.Read(b)
      fmt.Println("test")
      //Catch EOF in case I handle this eventually
      if err == io.EOF {
        fmt.Fprint(out, "EOF")
      }
      inputChan <- n
      //Sync the threads back up
      <-sync
    }
  }()

  //Loop until done
  for {
    select {
    case newConn := <-conChan:
      con = newConn
    case n := <-inputChan:
      if n == 0 {
        //Connection kill the connection or the connection already died
        closedChan <- true
        return
      }
      fmt.Println("Something")
      fmt.Fprint(out, string(b))
      //Zero the buffer out again
      ssetn(b, 0, n)
      //Resync the thread back up
      sync <- true
    }
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

  //Flags
  sport := *portPtr
  listen := *listenPtr
  udp := *udpPtr
  verbose := *verbosePtr


  //Used to notify io listeners if a connection has changed
  conChan := make(chan *net.Conn)

  //Connection
  var c net.Conn

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
        c, err = net.DialUDP("udp", laddr, raddr)
      } else {
        c, err = net.Dial("udp", net.JoinHostPort(ip, strconv.Itoa(dport)))
      }
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
  go output(&c, os.Stdout, conChan, closed)
  input(&c, os.Stdin, conChan, closed)
}
