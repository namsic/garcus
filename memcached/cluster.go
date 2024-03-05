package memcached

// [Operator] implementation for multiple arcus-memcached nodes.
type cluster struct {
	distributor Distributor
	protocol    Protocol
}

func NewCluster(distributor Distributor, protocol Protocol) Operator {
	return &cluster{
		distributor: distributor,
		protocol:    protocol,
	}
}

func (c *cluster) ProcessCommand(command []byte) (<-chan []byte, <-chan error) {
	key, err := c.protocol.GetKey(command)
	if err != nil {
		ch := make(chan error, 1)
		ch <- err
		close(ch)
		return nil, ch
	}
	return c.distributor.Distribute(key).ProcessCommand(command)
}

func (c *cluster) Identifier() []byte {
	return nil
}
