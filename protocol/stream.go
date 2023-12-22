package protocol

type PDU interface {
	Marshal([]byte)
	Unmarshal([]byte)
}
