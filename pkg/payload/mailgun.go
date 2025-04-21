package payload

type MessageConfig struct {
	Sender    string
	Recipient string
	CC        string
	BCC       string
	Body      []byte
	Subject   string
}

func (c *PolicyReportResultPayload) ToEmail() (EmailMsg, error) {
	return EmailMsg{}, nil
}
