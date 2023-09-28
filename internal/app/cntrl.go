package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/reviashko/remail2/model"
)

// ControllerInterface interface
type ControllerInterface interface {
	GetUnreadMessages() ([]model.MessageInfo, error)
	MakeSubjectIfSutable(msg *model.MessageInfo) (string, bool)
	Quit()
	PrintStat()
}

// Controller struct
type Controller struct {
	Config        *model.RemailConfig
	EmailReceiver POP3ClientInterface
	EmailSender   SMTPClientInterface
}

// NewController func
func NewController(emailReceiver POP3ClientInterface, config *model.RemailConfig, emailSender SMTPClientInterface) Controller {
	return Controller{EmailReceiver: emailReceiver, Config: config, EmailSender: emailSender}
}

// MakeSubjectIfSutable func
func (c *Controller) MakeSubjectIfSutable(msg *model.MessageInfo) (string, bool) {
	if msg.IsMultiPart {
		return "", false
	}

	if strings.Contains(msg.From, "<FSM-OUTSOURCE@megafon.ru>") {
		return msg.Subject, true
	}

	if strings.Contains(msg.From, "<sd@direct-credit.ru>") {
		//msg := append([]byte(fmt.Sprintf("Subject: %s\n", m.Subject)+c.Config.MIMEHeader), m.Body...)
		return msg.Subject, false
	}

	return "", false
}

// Run func
func (c *Controller) Run() {

	/*
		to := []string{
			//"devers@inbox.ru",
			"3ce2744b695b43d3a6c82f7ea0c1ff5b@webim-mail.ru",
		}
	*/

	for true {

		time.Sleep(time.Duration(c.Config.LoopDelaySec) * time.Second)

		msgs, err := c.EmailReceiver.GetUnreadMessages()
		if err != nil {
			fmt.Println(err.Error())
		}

		if len(msgs) == 0 {
			continue
		}

		for _, m := range msgs {

			if !strings.Contains(m.From, "devers@inbox.ru") {
				continue
			}

			if len(m.Cc) == 0 {
				continue
			}

			//fmt.Println("id=", m.MsgID, "Subject=", m.Subject, "from=", m.From, "Cc=", m.Cc)
			err := c.EmailSender.SendEmail(m.Cc, m.Body)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}
