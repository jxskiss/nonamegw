package protocol

type Headers []*KVEntry

func (p Headers) Get(key string) string {
	for _, kv := range p {
		if kv.Key == key {
			return kv.Value
		}
	}
	return ""
}

func (x *Packet) GetHeaderSlice() Headers {
	return x.Headers
}

func (x *Packet) GetHeaderMap() map[string]string {
	if len(x.Headers) == 0 {
		return nil
	}
	headers := make(map[string]string, len(x.Headers))
	for _, kv := range x.Headers {
		headers[kv.Key] = headers[kv.Value]
	}
	return headers
}

func (x *Packet) GetContent() *Content {
	content := &Content{
		BizFlag: x.BizFlag,
		Headers: x.GetHeaderMap(),
		Payload: x.Payload,
	}
	return content
}

func (x *Content) GetHeaderSlice() Headers {
	if len(x.Headers) == 0 {
		return nil
	}
	headers := make([]*KVEntry, 0, len(x.Headers))
	for k, v := range x.Headers {
		headers = append(headers, &KVEntry{
			Key:   k,
			Value: v,
		})
	}
	return headers
}

func (x *Content) ToPacket() *Packet {
	packet := &Packet{
		BizFlag: x.BizFlag,
		Headers: x.GetHeaderSlice(),
		Payload: x.Payload,
	}
	return packet
}
