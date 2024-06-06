package encoder

type Response struct {
	CorrelationID int32
	ClientID      string
	//body          protocolBody
}

func (r *Response) Encode(pe packetEncoder) error {
	pe.push(&lengthField{})
	pe.putInt16(16)

	err := pe.putString(r.ClientID)
	if err != nil {
		return err
	}
	// we don't use tag headers at the moment so we just put an array length of       0
	pe.putUVarint(0)

	return pe.pop()
}
