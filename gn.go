package main

import (
  "bufio"
  "fmt"
  "os"
  "net"
)

//Get user input
func input(con *net.Conn, c chan bool) {
  reader := bufio.NewReader(os.Stdin)
  scanner := bufio.NewScanner(reader)

  //Look for input
  go func () {
    for {
      good := scanner.Scan()
      if good {
        c <- false
      } else {
        c <- true
        return
      }
    }
  }()

  //Check to see if the channel is still open
  for {
    closed := <-c
    if !closed {
      _, err := (*con).Write(scanner.Bytes())
      if err != nil {
        return
      }
      (*con).Write([]byte("\n"))
    } else {
      break
    }
  }
  return
}

//Set the contents of a byte array to something
func sset(b []byte, e byte) {
  for i := 0; i < len(b); i++ {
    b[i] = e
  }
}

//Set the first n elements in byte array to something
func ssetn(b []byte, e byte, n int) {
  for i := 0; i < n; i++ {
    b[i] = e
  }
}

func main() {
  c, err := net.Dial("tcp", "127.0.0.1:1234")
  if err != nil {
    fmt.Println("Failed to connect")
    return
  }
  fmt.Println("Connected")
  b := make([]byte, 4096)
  closed := make(chan bool)
  go input(&c, closed)
  for true {
    n, err := c.Read(b)
    if err != nil {
      closed <- true
      return
    }
    fmt.Print(string(b))
    //Zero the buffer out again
    ssetn(b, 0, n)
  }
  c.Close()
}
