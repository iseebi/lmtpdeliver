package main

import (
	"flag"
	"github.com/emersion/go-smtp"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type LMTPDeliverServer struct {
	server             string
	helloRequestServer string
}

func (s *LMTPDeliverServer) forwardMessage(from string, to string, contents io.Reader) error {
	defaultTimeout := 30 * time.Second
	conn, err := net.DialTimeout("tcp", s.server, defaultTimeout)
	if err != nil {
		return err
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	host, _, _ := net.SplitHostPort(s.server)
	c, err := smtp.NewClientLMTP(conn, host)
	if err != nil {
		return err
	}

	if err = c.Hello(s.helloRequestServer); err != nil {
		return err
	}

	if err = c.Mail(from, nil); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}

	var errors []*smtp.SMTPError
	w, err := c.LMTPData(func(rcpt string, status *smtp.SMTPError) {
		if status != nil {
			errors = append(errors, status)
		}
	})
	if err != nil {
		return err
	}
	if _, err = io.Copy(w, contents); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	if len(errors) > 0 {
		return errors[0]
	}
	_ = c.Quit()
	return nil
}

func (s *LMTPDeliverServer) DeliveryMessage(writer http.ResponseWriter, request *http.Request) {
	recipient := request.FormValue("to")
	sender := request.FormValue("from")
	file, fileHeader, err := request.FormFile("mail")
	if err != nil {
		log.Printf("file read error %s", err)
		writer.WriteHeader(400)
		return
	}
	if fileHeader.Size == 0 {
		log.Println("empty file")
		writer.WriteHeader(400)
		return
	}
	if recipient == "" {
		log.Println("recipient not set")
		writer.WriteHeader(400)
		return
	}
	var from string
	if sender == "" {
		from = "undisclosed-recipient"
	} else {
		from = sender
	}

	err = s.forwardMessage(from, recipient, file)
	if err != nil {
		log.Printf("LMTP error %s", err)
		writer.WriteHeader(500)
		return
	}
	log.Printf("Success: %s", recipient)
	writer.WriteHeader(204)
}

func NewServer(server string, helloRequestServer string) *LMTPDeliverServer {
	return &LMTPDeliverServer{
		server:             server,
		helloRequestServer: helloRequestServer,
	}
}

func main() {
	server := flag.String("server", "localhost:24", "LMTP Server")
	listen := flag.String("listen", ":8080", "Listen Port")
	helloServer := flag.String("helloServer", "localhost", "Hello server name")
	flag.Parse()

	lmtpServer := NewServer(*server, *helloServer)

	http.HandleFunc("/delivery", lmtpServer.DeliveryMessage)
	log.Println("Listening...")
	_ = http.ListenAndServe(*listen, nil)
}
