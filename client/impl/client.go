package impl

// Client is data structor for client
type Client struct {
}

// NewClient creates a client instance
func NewClient() *Client {
	return nil
}

// Dial is called to initiate connections
func (cli *Client) Dial(address string, guid string) *MsgConnection {
	return nil
}

// MsgConnection the data struct for underlying connection
type MsgConnection struct {
}

// OnResponse is data struct for callback
type OnResponse struct {
}

// Send a message to the underlying connection
func (conn *MsgConnection) Send(msg string) error {
	return nil
}

// Subscribe messages from the underlying connection
func (conn *MsgConnection) Subscribe(msg string) error {
	return nil
}

// Call is called when message arived
func (cb *OnResponse) Call(msg string) error {
	return nil
}
