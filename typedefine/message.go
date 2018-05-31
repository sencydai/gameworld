package typedefine

type MessageType byte

const (
	MTClient       MessageType = 1
	MTSystem       MessageType = 2
	MTSystemAsynCB MessageType = 3
)

type Message struct {
	MTType MessageType
	CBFunc interface{}
	CBArgs interface{}
}
