package main

import(
    "fmt"
    "syscall"
    "log"
)

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

    peer_name, _ := syscall.Getpeername(nfd)
    fmt.Println(peer_name)

    response := "Hey Mon Chachos"
    _, msg_err := syscall.Write(nfd, []byte(response))
    if msg_err != nil{
        log.Fatal(msg_err)
    }
    syscall.Close(nfd)
  }
}
