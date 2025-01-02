/*
TODO
Read the request line and headers.
Maintain the connection open for multiple requests unless the Connection: close header is present.
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
    "strings"
    "errors"
)

func read_request(buf []byte, length int) error{

  var builder strings.Builder
  n, err := builder.Write(buf[:length])
  if err != nil {
    return errors.New("Error writing to builder")
  }
  req_str := builder.String()

  fmt.Println(req_str, n)

  path, after_path, _ := strings.Cut(req_str,"HTTP")
  ver, after_ver, _ := strings.Cut(after_path,"Host")
  host, after_host, _ := strings.Cut(after_ver,"User")
  user, after_user, _ := strings.Cut(after_host,"Accept")
  accept, after_accept, _ := strings.Cut(after_user,"Accept")
  lang, after_lang, _ := strings.Cut(after_accept,"Accept")
  encoding, after_enc, _ := strings.Cut(after_lang,"Connection:")
  connection, after_conn, _ := strings.Cut(after_enc,"Upgrade")
  _, priority, _ := strings.Cut(after_conn,"Priority")

  if strings.EqualFold(strings.TrimSpace(connection), "close"){
    return errors.New("Connection close received!")
  }

  fmt.Printf(path)
  fmt.Println("HTTP"+ver)
  fmt.Println("Host"+host)
  fmt.Println("User"+user)
  fmt.Println("Accept"+accept)
  fmt.Println("Accept"+lang)
  fmt.Println("Accept"+encoding)
  fmt.Println("connection"+connection)
  fmt.Println("Priority"+priority)
  return nil 
}

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

func client_handler(client_socket int){
  var recv_buf = make([]byte, 1024)
  
  for{
   length, _, recv_err := syscall.Recvfrom(client_socket, recv_buf, 0)

   if recv_err != nil{
      log.Print(recv_err)
   }
   if length == 0{
      log.Print("Connection closed by peer!")
      break	
   }else if length < 0{
      log.Print("Connection reset by peer!")
      break	
   }else{
      err := read_request(recv_buf, length)
      if err != nil{
        log.Print(err)
        response := "HTTP/1.1 200 OK\r\n" +
        "Content-Type: text/plain\r\n" +
        "Content-Length: 14\r\n" +
        "\r\n" +
        "Connection close received, closing...\n"

        _, msg_err := syscall.Write(client_socket, []byte(response))
       if msg_err != nil {
         if msg_err == syscall.EPIPE {
           log.Print("Broken pipe: Client has disconnected.")
           break
         }else {
           log.Print(msg_err)
         }
       }
       break
      }
   }

   response := "HTTP/1.1 200 OK\r\n" +
    "Content-Type: text/plain\r\n" +
    "Content-Length: 14\r\n" +
    "\r\n" +
    "Replying to Mon chacu\n"

    _, msg_err := syscall.Write(client_socket, []byte(response))
    if msg_err != nil {
      if msg_err == syscall.EPIPE {
        log.Print("Broken pipe: Client has disconnected.")
        break
      }else {
        log.Print(msg_err)
      }
    }
  }
  syscall.Close(client_socket)
}

// To Try: AF_UNSPEC
func main() {
  var sa syscall.SockaddrInet4

  fd, sock_err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
  if sock_err != nil{
    log.Fatal(sock_err)
  }

  sa.Port = 3232
  sa.Addr = [4]byte{192,168,100,43} //{127,0,0,1}
  bind_err := syscall.Bind(fd, &sa)
  if bind_err != nil{
    log.Fatal(bind_err)
  }

  listen_err := syscall.Listen(fd, 10)
  if listen_err != nil{
    log.Fatal(listen_err)
  }
  fmt.Println("Server is listening on 192.168.100.43")

  for {
    nfd, _, accept_err := syscall.Accept(fd)
    if accept_err != nil{
        log.Fatal(accept_err)
    }

    addr, port, getPeerErr := getpeer(nfd)
    if getPeerErr != nil {
      log.Println("Getpeername error:", getPeerErr)
    } else {
      fmt.Println("Client connected: ", addr, port)
    }

    response := "HTTP/1.1 200 OK\r\n" +
    "Content-Type: text/plain\r\n" +
    "Content-Length: 14\r\n" +
    "\r\n" +
    "Hey Mon Chachos\n"

    _, msg_err := syscall.Write(nfd, []byte(response))
    if msg_err != nil{
        log.Fatal(msg_err)
        syscall.Close(nfd)
    }
    
    go client_handler(nfd)
    
    //syscall.Close(nfd)
  }
}
