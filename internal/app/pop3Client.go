package app

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-message"

	"github.com/reviashko/remail2/model"

	"github.com/knadh/go-pop3"
)

// POP3ClientInterface interface
type POP3ClientInterface interface {
	GetUnreadMessages() ([]model.MessageInfo, error)
	Quit()
	PrintStat() error
}

// POP3Client struct
type POP3Client struct {
	MsgID   int
	Conn    *pop3.Conn
	Decoder *mime.WordDecoder
}

// NewPOP3Client func
func NewPOP3Client(server string, port int, tlsEnabled bool, login string, pswd string, msgID int, ttl int64) POP3Client {

	p := pop3.New(pop3.Opt{
		DialTimeout: time.Duration(time.Duration(ttl) * time.Second),
		Host:        server,
		Port:        port,
		TLSEnabled:  tlsEnabled,
	})

	// Don't forget exec Quit before app close!
	c, err := p.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Auth(login, pswd); err != nil {
		log.Fatal(err)
	}

	dec := mime.WordDecoder{}

	return POP3Client{Conn: c, MsgID: msgID, Decoder: &dec}
}

// Quit func
func (p *POP3Client) Quit() {
	p.Conn.Quit()
}

// PrintStat func
func (p *POP3Client) PrintStat() error {
	count, size, err := p.Conn.Stat()
	if err != nil {
		return err
	}
	fmt.Println("LastID=", p.MsgID, "total messages=", count, "size=", size)

	return nil
}

// GetCC func
func (p *POP3Client) GetCC(mr message.MultipartReader) []string {
	var retval []string
	if mr != nil {
		loop := 0
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				//log.Fatal(err)
				return retval
			}

			if loop == 0 {
				if b, err := io.ReadAll(p.Body); err == nil {
					bodyString := fmt.Sprintf("%s", b)
					startFrom := strings.Index(bodyString, "Копия:")
					if startFrom >= 0 {
						re := regexp.MustCompile(`[a-z]+@dc2b.ru`)
						retval = re.FindAllString(bodyString[startFrom:], -1)
					}
				}
			}

			loop++
		}
	}
	return retval
}

// GetUnreadMessages func
func (p *POP3Client) GetUnreadMessages() ([]model.MessageInfo, error) {

	retval := make([]model.MessageInfo, 0)
	dec := mime.WordDecoder{}

	msgs, _ := p.Conn.List(p.MsgID)
	for _, m := range msgs {
		if m.ID <= p.MsgID {
			continue
		}

		p.MsgID = m.ID
		msg, _ := p.Conn.Retr(m.ID)

		subj, err := dec.DecodeHeader(msg.Header.Get("Subject"))
		if err != nil {
			//TODO: need to manage to this
			continue
		}

		from := msg.Header.Get("From")
		multiPart := msg.MultipartReader()
		cc := p.GetCC(multiPart)

		// multi-part body:
		// https://github.com/emersion/go-message/blob/master/example_test.go#L12
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(msg.Body)
		if err != nil {
			log.Fatal(err)
		}

		retval = append(retval, model.MessageInfo{MsgID: m.ID, Subject: subj, IsMultiPart: multiPart != nil, From: from, Cc: cc, Body: buf.Bytes()})
	}

	return retval, nil
}
