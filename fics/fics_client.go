package fics

type FicsClient struct {
	telnet *TelnetClient
}

func NewFicsClient() (*FicsClient, error) {
	telnet, err := NewTelnetClient("freechess.org:23")
	if err != nil {
		return nil, err
	}

	return &FicsClient{
		telnet,
	}, nil
}

func (c *FicsClient) Close() {
	c.telnet.Close()
}
