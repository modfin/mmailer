package mmailer

import (
	"fmt"
)

type Address struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (a Address) String() string {
	if len(a.Name) == 0 {
		return a.Email
	}
	return fmt.Sprintf("\"%s\" <%s>", a.Name, a.Email)
}

type Email struct {
	Headers map[string]string `json:"headers"`
	From    Address           `json:"from"`
	To      []Address         `json:"to"`
	Cc      []Address         `json:"cc"`
	Subject string            `json:"subject"`
	Text    string            `json:"text"`
	Html    string            `json:"html"`
}

func NewEmail() Email {
	return Email{
		Headers: map[string]string{},
	}
}

type Response struct {
	Service   string `json:"service"`
	MessageId string `json:"message_id"`
	Email     string `json:"email"`
}

func (r Response) Id() string {
	return fmt.Sprintf("%s:%s", r.Service, r.MessageId)
}

type PosthookEvent string

func (p PosthookEvent) String() string {
	return string(p)
}

// Message has been successfully delivered to the receiving server.
const EventDelivered PosthookEvent = "delivered"

//Recipient's email server temporarily rejected message.
const EventDeferred PosthookEvent = "deferred"

//Receiving server could not or would not accept message.
const EventBounce PosthookEvent = "bounce"

const EventDropped PosthookEvent = "dropped"

//Recipient has opened the HTML message. You need to enable Open Tracking for getting this type of event.
const EventOpen PosthookEvent = "open"

//Recipient clicked on a link within the message. You need to enable Click Tracking for getting this type of event.
const EventClick PosthookEvent = "click"

//Recipient marked a message as spam.
const EventSpam PosthookEvent = "spam"

//Recipient clicked on message's subscription management link. You need to enable Subscription Tracking for getting this type of event.
const EventUnsubscribe PosthookEvent = "unsubscribe"

const EventUnknown PosthookEvent = "unknown"

type Posthook struct {
	Service   string        `json:"service"`
	MessageId string        `json:"message_id"`
	Email     string        `json:"email"`
	Event     PosthookEvent `json:"event"`
	Info      string        `json:"info,omitempty"`
}

func (r Posthook) Id() string {
	return fmt.Sprintf("%s:%s", r.Service, r.MessageId)
}
