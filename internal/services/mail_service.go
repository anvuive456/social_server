package services

import (
	mail "github.com/xhit/go-simple-mail/v2"
)

type MailService struct {
	mail *mail.SMTPClient
}

func NewMailService() *MailService {
	server := mail.NewSMTPClient()
	server.Host = "smtp.gmail.com"
	server.Port = 587
	server.Username = "trannhatan2803@gmail.com"
	server.Password = "examplepass"

	mail, err := server.Connect()
	if err != nil {
		panic(err)
	}
	return &MailService{
		mail: mail,
	}
}

func (ms *MailService) SendEmailVerification(to string) error {
	msg := mail.NewMSG()
	msg.SetFrom("trannhatan2803@gmail.com")
	msg.AddTo(to)
	msg.SetSubject("Verify your email")
	msg.SetBody(mail.TextHTML, `
		<p>Nhấn vào link dưới đây để xác thực email của bạn:</p>
		<a href="http://example.com/verify">Xác thực Email</a>
	`)

	if err := msg.Send(ms.mail); err != nil {
		return err
	}
	return nil
}

func (ms *MailService) Close() error {
	return ms.mail.Close()
}
