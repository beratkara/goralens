package goralens

import (
	"context"

	"github.com/goravel/framework/contracts/mail"
)

func WrapMail(m mail.Mail) mail.Mail {
	if m == nil {
		return nil
	}
	if _, ok := m.(*watchedMail); ok {
		return m
	}
	return &watchedMail{inner: m}
}

type watchedMail struct {
	inner mail.Mail

	attachments []string
	bcc, cc, to []string
	content     mail.Content
	from        mail.Address
	headers     map[string]string
	subject     string
}

func (m *watchedMail) Attach(files []string) mail.Mail {
	m.inner = m.inner.Attach(files)
	m.attachments = files
	return m
}

func (m *watchedMail) Bcc(addresses []string) mail.Mail {
	m.inner = m.inner.Bcc(addresses)
	m.bcc = addresses
	return m
}

func (m *watchedMail) Cc(addresses []string) mail.Mail {
	m.inner = m.inner.Cc(addresses)
	m.cc = addresses
	return m
}

func (m *watchedMail) Content(content mail.Content) mail.Mail {
	m.inner = m.inner.Content(content)
	m.content = content
	return m
}

func (m *watchedMail) From(address mail.Address) mail.Mail {
	m.inner = m.inner.From(address)
	m.from = address
	return m
}

func (m *watchedMail) Headers(headers map[string]string) mail.Mail {
	m.inner = m.inner.Headers(headers)
	m.headers = headers
	return m
}

func (m *watchedMail) Subject(subject string) mail.Mail {
	m.inner = m.inner.Subject(subject)
	m.subject = subject
	return m
}

func (m *watchedMail) To(addresses []string) mail.Mail {
	m.inner = m.inner.To(addresses)
	m.to = addresses
	return m
}

func (m *watchedMail) Send(mailable ...mail.Mailable) error {
	err := m.inner.Send(mailable...)
	m.record(false, err, mailable)
	return err
}

func (m *watchedMail) Queue(mailable ...mail.Mailable) error {
	err := m.inner.Queue(mailable...)
	m.record(true, err, mailable)
	return err
}

func (m *watchedMail) record(queued bool, err error, mailables []mail.Mailable) {
	r := Instance()
	if r == nil || !r.config.watching(TypeMail) {
		return
	}
	subject, from, to, cc, bcc := m.subject, m.from, m.to, m.cc, m.bcc
	content, attachments, headers := m.content, m.attachments, m.headers
	mailableName := ""

	if len(mailables) > 0 && mailables[0] != nil {
		ml := mailables[0]
		mailableName = typeName(ml)
		if env := ml.Envelope(); env != nil {
			if env.Subject != "" {
				subject = env.Subject
			}
			if env.From.Address != "" {
				from = env.From
			}
			if len(env.To) > 0 {
				to = env.To
			}
			if len(env.Cc) > 0 {
				cc = env.Cc
			}
			if len(env.Bcc) > 0 {
				bcc = env.Bcc
			}
		}
		if c := ml.Content(); c != nil {
			content = *c
		}
		if a := ml.Attachments(); len(a) > 0 {
			attachments = a
		}
		if h := ml.Headers(); len(h) > 0 {
			headers = h
		}
	}
	entryContent := map[string]any{
		"mailable":    mailableName,
		"queued":      queued,
		"subject":     subject,
		"from":        from,
		"to":          to,
		"cc":          cc,
		"bcc":         bcc,
		"html":        content.Html,
		"text":        content.Text,
		"html_view":   content.HtmlView,
		"text_view":   content.TextView,
		"attachments": attachments,
		"headers":     headers,
	}
	if err != nil {
		entryContent["error"] = err.Error()
	}
	e := newEntry(TypeMail, entryContent)
	for _, addr := range to {
		e.AddTags("mail:" + addr)
	}
	r.Record(context.Background(), e)
}
