/*
TODO
Need to make better error handling and logging.
Maintain the connection open for multiple requests unless the Connection: close header is present.
*/
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
    return errors.New("Internal Sever error while loading the img file")
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
      return errors.New("Broken pipe: Client has disconnected.")
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
    return errors.New("Internal Sever error while loading the HTML file")
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
      return errors.New("Broken pipe: Client has disconnected.")
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
      return errors.New("Broken pipe: Client has disconnected.")
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
     return [4]byte{}, 0, fmt.Errorf("unexpected address type")
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
      log.Print(recv_err)
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
        log.Print(err)
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
    return errors.New("Error writing to builder")
  }
  req_str := builder.String()

  path, after_path, _ := strings.Cut(req_str,"HTTP")
  ver, after_ver, _ := strings.Cut(after_path,"Host")
  host, after_host, _ := strings.Cut(after_ver,"User")
  user, after_user, _ := strings.Cut(after_host,"Accept")
  accept, after_accept, _ := strings.Cut(after_user,"Accept")
  lang, after_lang, _ := strings.Cut(after_accept,"Accept")
  encoding, after_enc, _ := strings.Cut(after_lang,"Connection:")
  connection, after_conn, _ := strings.Cut(after_enc,"Upgrade")
  _, priority, _ := strings.Cut(after_conn,"Priority")

  if (host == ""){
    return errors.New("Host header is empty, closing socket!")
  }

  if strings.EqualFold(strings.TrimSpace(connection), "close"){
    return errors.New("Connection close received!")
  }

  if !strings.Contains(strings.TrimSpace(ver), "/1.1"){
    return errors.New("This Server expects HTTP/1.1!")
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

  if strings.EqualFold(strings.TrimSpace(path), "GET /") {
      return send_html(conn_socket, "pages/index.html")
  }else if strings.EqualFold(strings.TrimSpace(path), "GET /hey.jpg") {
      return send_img(conn_socket, "pages/img/hey.jpg")
  } else if strings.EqualFold(strings.TrimSpace(path), "GET /about") {
      return send_html(conn_socket, "pages/about.html")
  } else if strings.EqualFold(strings.TrimSpace(path), "GET /prime.jpg") {
      return send_img(conn_socket, "pages/img/prime.jpg")
  } else {
      send_msg("404 Not Found", "Page doesn't exist", "close", conn_socket)
      return errors.New("404 Not Found, page doesn't exist")
  }

  return nil 
}

// To Try: AF_UNSPEC
func main() {
  var sa syscall.SockaddrInet4

  socket, sock_err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
  if sock_err != nil{
    log.Fatal(sock_err)
  }

  sa.Port = 3232
  sa.Addr = [4]byte{192,168,100,74}
  bind_err := syscall.Bind(socket, &sa)
  if bind_err != nil{
    log.Fatal(bind_err)
  }

  listen_err := syscall.Listen(socket, 10)
  if listen_err != nil{
    log.Fatal(listen_err)
  }
  fmt.Println("Server is listening on 192.168.100.74")

  for {
    connection_socket, _, accept_err := syscall.Accept(socket)
    if accept_err != nil{
        log.Fatal(accept_err)
    }

    addr, port, getPeerErr := get_peer(connection_socket)
    if getPeerErr != nil {
      log.Println("Getpeername error:", getPeerErr)
    } else {
      fmt.Println("Client connected: ", addr, port)
    }

    go client_handler(connection_socket)

  }
}
