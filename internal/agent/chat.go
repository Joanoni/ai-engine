package agent

import "github.com/swarmit/ai-engine/internal/llm"

// Chat manages the message history for a single conversation between two entities.
type Chat struct {
	messages []llm.Message
}

// NewChat creates a new empty Chat.
func NewChat() *Chat {
	return &Chat{}
}

// AddMessage appends a pre-built Message to the chat history.
func (c *Chat) AddMessage(m llm.Message) {
	c.messages = append(c.messages, m)
}

// AddText appends a simple text message to the chat history.
func (c *Chat) AddText(role llm.Role, text string) {
	c.messages = append(c.messages, llm.NewTextMessage(role, text))
}

// Messages returns a copy of the current message history.
func (c *Chat) Messages() []llm.Message {
	out := make([]llm.Message, len(c.messages))
	copy(out, c.messages)
	return out
}

// Reset clears the message history.
func (c *Chat) Reset() {
	c.messages = nil
}
