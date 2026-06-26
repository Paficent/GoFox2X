package protocol

import "github.com/Paficent/GoFox2X/data"

func ParseExtensionRequest(m *Message) (string, *data.GFSObject, bool) {
	if m.Controller != Extension || m.Action != ActionExtensionCall {
		return "", nil, false
	}

	command, ok := m.Payload.GetUtfString("c")
	if !ok {
		return "", nil, false
	}

	params, ok := m.Payload.GetGFSObject("p")
	if !ok || params == nil {
		params = data.MakeGFSObject()
	}

	return command, params, true
}

func ExtensionResponse(command string, params *data.GFSObject) *Message {
	if params == nil {
		params = data.MakeGFSObject()
	}

	payload := data.MakeGFSObject().
		PutUtfString("c", command).
		PutInt("r", -1).
		PutGFSObject("p", params)

	return NewMessage(Extension, ActionExtensionCall, payload)
}
