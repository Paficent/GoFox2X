package protocol

// ControllerID selects where a message is routed to.
// Encoded as the byte field "c" of the envelope.
type ControllerID byte

const (
	System    ControllerID = 0
	Extension ControllerID = 1
)

func (c ControllerID) String() string {
	switch c {
	case System:
		return "system"
	case Extension:
		return "extension"
	default:
		return "controller(" + itoa(int(c)) + ")"
	}
}

type Action = int16

// TODO: More than just connect, handshake, and login
const (
	ActionHandshake Action = 0
	ActionLogin     Action = 1
	ActionLogout    Action = 2
)

// Some SFS2X builds use 13 for the extension CALL action; confirm against the client you are targeting
const (
	ActionExtensionCall Action = 12
)

// int->string for the String method
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
