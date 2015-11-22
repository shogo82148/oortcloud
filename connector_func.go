package oortcloud

type FuncConnection struct {
	SendFunc func(data []byte) error
}

func (c *FuncConnection) Send(data []byte) error {
	if c.SendFunc == nil {
		return nil
	}
	return c.SendFunc(data)
}
