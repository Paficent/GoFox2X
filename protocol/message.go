package protocol

import "paficent/GoFox2X/data"

const (
	keyController = "c" // (byte)
	keyAction     = "a" // (short)
	keyPayload    = "p"
)

type Message struct {
	Controller ControllerID
	Action     Action
	Payload    *data.GFSObject
}

func NewMessage(controller ControllerID, action Action, payload *data.GFSObject) *Message {
	return &Message{Controller: controller, Action: action, Payload: payload}
}

func (m *Message) toEnvelope() *data.GFSObject {
	payload := m.Payload
	if payload == nil {
		payload = data.MakeGFSObject()
	}
	return data.MakeGFSObject().
		PutByte(keyController, byte(m.Controller)).
		PutShort(keyAction, m.Action).
		PutGFSObject(keyPayload, payload)
}

func (m *Message) MarshalBinary() ([]byte, error) {
	return m.toEnvelope().MarshalBinary()
}

func DecodeMessage(body []byte) (*Message, error) {
	envelope := data.MakeGFSObject()
	if err := envelope.UnmarshalBinary(body); err != nil {
		return nil, err
	}

	controller, _ := envelope.GetByte(keyController)
	action, _ := envelope.GetShort(keyAction)
	payload, ok := envelope.GetGFSObject(keyPayload)
	if !ok || payload == nil {
		payload = data.MakeGFSObject()
	}

	return &Message{
		Controller: ControllerID(controller),
		Action:     action,
		Payload:    payload,
	}, nil
}
