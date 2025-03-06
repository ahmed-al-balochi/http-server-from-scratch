package main

import(
    "fmt"
    "syscall"
    "log"
    "strings"
    "errors"
    "strconv"
    "os"
)

func send_img(conn_socket int, path string) error{
  img_file, err := os.ReadFile(path)
  if err != nil{
    body :="Internal Sever error while loading the img file"
    send_msg("500 Internal Server Error", body, "close", conn_socket)
    return errors.New("ERROR: Internal Sever error while loading the img file")
  }

  response := fmt.Sprintf("HTTP/1.1 200 OK\r\n" +
    "Host: www.monchaco.tech-ahmed.com\r\n" +
    "Content-Type: img/jpg\r\n" +
    "Content-Length: %d\r\n" +
    "Connection: close\r\n" +
    "\r\n%s",
    len(img_file),string(img_file))

  _, msg_err := syscall.Write(conn_socket, []byte(response))
  if msg_err != nil {
    if msg_err == syscall.EPIPE {
      return errors.New("ERROR: Broken pipe: Client has disconnected.")
    }else {
      return msg_err
    }
  }
  return nil
}

func send_html(conn_socket int, path string) error{
  html_file, err := os.ReadFile(path)
  if err != nil{
    body :="Internal Sever error while loading the HTML file"
    send_msg("500 Internal Server Error", body, "close", conn_socket)
    return errors.New("ERROR: Internal Sever error while loading the HTML file")
  }

  response := fmt.Sprintf("HTTP/1.1 200 OK\r\n" +
    "Host: www.monchaco.tech-ahmed.com\r\n" +
    "Content-Type: text/html\r\n" +
    "Content-Length: %d\r\n" +
    "Connection: close\r\n" +
    "\r\n%s",
    len(html_file),string(html_file))

  _, msg_err := syscall.Write(conn_socket, []byte(response))
  if msg_err != nil {
    if msg_err == syscall.EPIPE {
      return errors.New("ERROR: Broken pipe: Client has disconnected.")
    }else {
      return msg_err
    }
  }
  return nil
} 

func send_msg(code string, body string, conn_state string, conn_socket int) error{
  body_len := strconv.Itoa(len(body))

  response := "HTTP/1.1 " + code + "\r\n" +
  "Content-Type: text/plain\r\n" +
  "Content-Length: " + body_len + "\r\n" +
  "Connection: " + conn_state  + "\r\n" +
  "\r\n" +
  body + "\n"

  _, msg_err := syscall.Write(conn_socket, []byte(response))
  if msg_err != nil {
    if msg_err == syscall.EPIPE {
      return errors.New("ERROR: Broken pipe: Client has disconnected.")
    }else {
      return msg_err
    }
  }
  return nil
}

func get_peer(conn_socket int) (Addr [4]byte, Port int, err error){
  peer, err := syscall.Getpeername(conn_socket)
  if err != nil{
    log.Fatal(err)
  }
  peer_data, ok := peer.(*syscall.SockaddrInet4)
  if !ok {
   return [4]byte{}, 0, errors.New("ERROR: unexpected address type")
  }
  Addr = peer_data.Addr
  Port = peer_data.Port
  return Addr, Port, nil
}

func client_handler(conn_socket int){
  var recv_buf = make([]byte, 1024)
  
  for{
   length, _, recv_err := syscall.Recvfrom(conn_socket, recv_buf, 0)

   if recv_err != nil{
      log.Print("ERROR: ", recv_err)
   }

   if length == 0{
      body := "Connection closed by peer!" 
      send_msg("400", body, "close", conn_socket)
      break	
   }else if length < 0{
      body := "Connection reset by peer!"
      send_msg("400", body, "close", conn_socket)
      break	
   }else{
      err := read_request(conn_socket, recv_buf, length)
      if err != nil{
        log.Print("ERROR: ", err)
	body := "Detected an error, closing connection..."
        send_msg("400", body, "close", conn_socket)
        break
      }
   }
  }
  syscall.Close(conn_socket)
}

func read_request(conn_socket int, buf []byte, length int) error{

  var builder strings.Builder
  _, err := builder.Write(buf[:length])
  if err != nil {
    return errors.New("ERROR: Error writing to builder")
  }
  req_str := builder.String()
  lines := strings.Split(req_str, "\n")

  for _, line := range lines {
     line = strings.TrimSpace(line)

     switch line {
       case "Connection: close":
	 log.Println("INFO:" + line)
         return errors.New("ERROR: Connection close received!")
       case "GET / HTTP/1.1":
	 log.Println("INFO:" + line)
         return send_html(conn_socket, "pages/index.html")
       case "GET /favicon.ico HTTP/1.1":
	 log.Println("INFO:" + line)
         return send_img(conn_socket, "pages/img/favicon.ico")
       case "GET /hey.jpg HTTP/1.1":
	 log.Println("INFO:" + line)
         return send_img(conn_socket, "pages/img/hey.jpg")
       case "GET /about HTTP/1.1":
	 log.Println("INFO:" + line)
         return send_html(conn_socket, "pages/about.html")
       case "GET /prime.jpg HTTP/1.1":
	 log.Println("INFO:" + line)
         return send_img(conn_socket, "pages/img/prime.jpg")
       default:
	 log.Println("ERROR:" + line)
         send_msg("404 Not Found", "Page doesn't exist", "close", conn_socket)
         return errors.New("ERROR: 404 Not Found, page doesn't exist")
     }
  }
  
  if !strings.Contains(strings.TrimSpace(req_str), "/1.1"){
    return errors.New("ERROR: This Server expects HTTP/1.1!")
  }

  return nil 
}

func main() {
  var sa syscall.SockaddrInet4

  socket, sock_err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
  if sock_err != nil{
    log.Fatal(sock_err)
  }

  sa.Port = 3232
  sa.Addr = [4]byte{127,0,0,1}
  bind_err := syscall.Bind(socket, &sa)
  if bind_err != nil{
    log.Fatal(bind_err)
  }

  listen_err := syscall.Listen(socket, 10)
  if listen_err != nil{
    log.Fatal(listen_err)
  }
  log.Println("INFO: Server is listening on 127.0.0.1")

  for {
    connection_socket, _, accept_err := syscall.Accept(socket)
    if accept_err != nil{
        errors.New("ERROR: Failed to accept the connection, please try again")
    }

    addr, port, getPeerErr := get_peer(connection_socket)
    if getPeerErr != nil {
	    log.Println("ERROR: Getpeername error:", getPeerErr)
    } else {
	    log.Println("INFO: Client connected: ", addr, port)
    }

    go client_handler(connection_socket)

  }
}
