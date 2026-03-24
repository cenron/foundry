package broker

import "fmt"

const (
	ExchangeCommands = "foundry.commands"
	ExchangeEvents   = "foundry.events"
	ExchangeLogs     = "foundry.logs"
)

func (c *Client) declareExchanges() error {
	exchanges := []string{ExchangeCommands, ExchangeEvents, ExchangeLogs}

	for _, name := range exchanges {
		if err := c.ch.ExchangeDeclare(name, "topic", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declaring exchange %q: %w", name, err)
		}
	}

	return nil
}
