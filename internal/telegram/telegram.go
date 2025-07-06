package telegram

type MessageType string

const (
	MessageInfo  MessageType = "info"
	MessageWarn  MessageType = "warn"
	MessageError MessageType = "error"
)

type ITelegramClient interface {
	// Observer will send a Message on notifiy()
	// with MessageType `MessageInfo`.
	//
	// Deployer will send a Message either
	// on deployment success or failure.
	SendMsg(Message) error
}

type Message struct {
	Type    MessageType
	Title   string
	Content string
}
