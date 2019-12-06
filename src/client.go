package main

import (
  "bufio"
  "io"
  "fmt"
  "net"
  //"github.com/gdamore/tcell"
  "os"
)

type Client struct {
  Con     net.Conn
  Closed  chan bool
  In      *os.File
  Out     *os.File
}

//Implements io.Reader
func (c Client) Read(b []byte) (n int, err error) {
  return c.Con.Read(b)
}

//Implements io.Writer
func (c Client) Write(b []byte) (n int, err error) {
  return c.Con.Write(b)
}

func (c Client) Close() {
  c.Con.Close()
}

//Connection input
func (c Client) handleInput() {
  reader := bufio.NewReader(c.In)
  scanner := bufio.NewScanner(reader)

  input := make(chan bool)
  sync := make(chan bool)

  //Look for input
  go func() {
    for {
      input <- scanner.Scan()
      <-sync
    }
  }()

  //Check to see if the channel is still open
  for {
    select {
    case closed := <-c.Closed:
      if closed {
        return
      }
    case <-input:
      _, err := c.Write(append(scanner.Bytes(), byte('\n')))
      if err != nil {
        return
      }
      //Synchronize the states
      sync <- true
    }
  }
  return
}

//TODO add select to check for when this is closed
//Connection output
func (c Client) handleOutput() {
  buff := make([]byte, 4096)

  input := make(chan int)
  sync := make(chan bool)

  //Look for input
  go func() {
    for {
      n, err := c.Read(buff)
      input <- n
      if n == 0 {
        //Catch EOF in case I handle this eventually
        if err == io.EOF {
          fmt.Fprintln(c.Out, "EOF")
        }
        break
      }
      //Sync the goroutines back up
      <-sync
    }
  }()


  //Loop until done
  for {
    select {
    case n := <-input:
      if n == 0 {
        //Connection kill the connection or the connection already died
        c.Closed <- true
        c.Close()
        return
      }
      fmt.Fprint(c.Out, string(buff))
      //Zero the buffer out again
      ssetn(buff, 0, n)
      //Sync the goroutines
      sync <- true
    }
  }
}

//Handle connections
func (c Client) Handle() {
  go c.handleInput()
  c.handleOutput()
}

