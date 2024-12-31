/*
TODO
Maintain the connection open for multiple requests unless the Connection: close header is present.
Read the request line and headers.
Ensure the presence of mandatory headers like Host.
Generate responses based on parsed requests, including proper headers like Content-Type and Content-Length.
Support at least the common HTTP methods (GET, POST, etc.).
Send appropriate HTTP error codes (400 Bad Request, 404 Not Found, etc.) when necessary.
*/
package main

import(
    "fmt"
    "syscall"
    "log"
)

func getpeer(nfd int) (Addr [4]byte, Port int, err error){
  peer, err := syscall.Getpeername(nfd)
  if err != nil{
    log.Fatal(err)
  }
  peer_data, ok := peer.(*syscall.SockaddrInet4)
  if !ok {
     return [4]byte{}, 0, fmt.Errorf("unexpected address type")
  }
  Addr = peer_data.Addr
  Port = peer_data.Port
  return Addr, Port, nil
}

// To Try: AF_UNSPEC
func main() {
  fd, sock_err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
  if sock_err != nil{
    log.Fatal(sock_err)
  }

  var sa syscall.SockaddrInet4
  sa.Port = 3232
  sa.Addr = [4]byte{127,0,0,1}
  bind_err := syscall.Bind(fd, &sa)
  if bind_err != nil{
    log.Fatal(bind_err)
  }

  listen_err := syscall.Listen(fd, 10)
  if listen_err != nil{
    log.Fatal(listen_err)
  }
  fmt.Println("Server is listening on 127.0.0.1:3232")

  for {
    nfd, _, accept_err := syscall.Accept(fd)
    if accept_err != nil{
        log.Fatal(accept_err)
    }

    addr, port, getPeerErr := getpeer(nfd)
    if getPeerErr != nil {
      log.Println("Getpeername error:", getPeerErr)
    } else {
      fmt.Printf("Client connected: %v:%d\n", addr, port)
    }

    response := "HTTP/1.1 200 OK\r\n" +
    "Content-Type: text/plain\r\n" +
    "Content-Length: 14\r\n" +
    "\r\n" +
    "Hey Mon Chachos"

    _, msg_err := syscall.Write(nfd, []byte(response))
    if msg_err != nil{
        log.Fatal(msg_err)
        syscall.Close(nfd)
    }

    syscall.Close(nfd)
    
  }
}
