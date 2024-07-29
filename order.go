package whatsmeow

import waBinary "go.mau.fi/whatsmeow/binary"

//////////////////////////////
/// Order | Structs & Types
//////////////////////////////

type GetOrderParams struct {
	ID    string
	Token string
}

//////////////////////////////
/// Order | Functions
//////////////////////////////

func (cli *Client) GetOrder(params GetOrderParams) (*waBinary.Node, error) {

	return nil, nil
}
