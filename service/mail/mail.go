package mail

import (
	"context"
	"crypto/tls"
	"net/smtp"
	"net/textproto"

	"github.com/jordan-wright/email"
	"github.com/pkg/errors"
)

// Mail struct holds necessary data to send emails.
type Mail struct {
	usePlainText      bool
	senderAddress     string
	smtpHostAddr      string
	smtpAuth          smtp.Auth
	receiverAddresses []string
	useTLS            bool
	useStartTLS       bool
	tlsConfig         *tls.Config
}

// New returns a new instance of a Mail notification service.
func New(senderAddress, smtpHostAddress string) *Mail {
	return &Mail{
		usePlainText:      false,
		senderAddress:     senderAddress,
		smtpHostAddr:      smtpHostAddress,
		receiverAddresses: []string{},
		useTLS:            false,
		useStartTLS:       false,
	}
}

// BodyType is used to specify the format of the body.
type BodyType int

const (
	// PlainText is used to specify that the body is plain text.
	PlainText BodyType = iota
	// HTML is used to specify that the body is HTML.
	HTML
)

// AuthenticateSMTP authenticates you to send emails via smtp.
// Example values: "", "test@gmail.com", "password123", "smtp.gmail.com"
// For more information about smtp authentication, see here:
//
//	-> https://pkg.go.dev/net/smtp#PlainAuth
func (m *Mail) AuthenticateSMTP(identity, userName, password, host string) {
	m.smtpAuth = smtp.PlainAuth(identity, userName, password, host)
}

// smtp.office365.com use AUTH LOGIN
// solve the problem: 504 5.7.4 Unrecognized authentication type
// https://github.com/go-gomail/gomail/issues/16#issuecomment-73672398
func (m *Mail) AuthenticateSMTPWithLoginAuth(identity, userName, password, host string) {
	m.smtpAuth = LoginAuth(userName, password)
}

// AddReceivers takes email addresses and adds them to the internal address list. The Send method will send
// a given message to all those addresses.
func (m *Mail) AddReceivers(addresses ...string) {
	m.receiverAddresses = append(m.receiverAddresses, addresses...)
}

// BodyFormat can be used to specify the format of the body.
// Default BodyType is HTML.
func (m *Mail) BodyFormat(format BodyType) {
	switch format {
	case PlainText:
		m.usePlainText = true
	default:
		m.usePlainText = false
	}
}

// SetTLS can be used to send email over tls with an optional TLS config.
func (m *Mail) SetTLS(tlsConfig *tls.Config) {
	m.useTLS = true
	m.tlsConfig = tlsConfig
}

// UnSetTLS can be used to send email without tls.
func (m *Mail) UnSetTLS() {
	m.useTLS = false
	m.tlsConfig = nil
}

func (m *Mail) SetStartTLS(tlsConfig *tls.Config) {
	m.useStartTLS = true
	m.tlsConfig = tlsConfig
}

func (m *Mail) UnSetStartTLS() {
	m.useStartTLS = false
	m.tlsConfig = nil
}

func (m *Mail) newEmail(subject, message string) *email.Email {
	msg := &email.Email{
		To:      m.receiverAddresses,
		From:    m.senderAddress,
		Subject: subject,
		Headers: textproto.MIMEHeader{},
	}

	if m.usePlainText {
		msg.Text = []byte(message)
	} else {
		msg.HTML = []byte(message)
	}
	return msg
}

// Send takes a message subject and a message body and sends them to all previously set chats. Message body supports
// html as markup language.
func (m Mail) Send(ctx context.Context, subject, message string) error {
	msg := m.newEmail(subject, message)

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		if m.useStartTLS {
			err = msg.SendWithStartTLS(m.smtpHostAddr, m.smtpAuth, m.tlsConfig)
		} else if m.useTLS {
			err = msg.SendWithTLS(m.smtpHostAddr, m.smtpAuth, m.tlsConfig)
		} else {
			err = msg.Send(m.smtpHostAddr, m.smtpAuth)
		}
		if err != nil {
			err = errors.Wrap(err, "failed to send mail")
		}
	}

	return err
}
