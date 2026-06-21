package vrchatosc

func buildOSCPacket(address, typeTag string, intArg *int32) []byte {
	var buf []byte
	buf = append(buf, []byte(address)...)
	buf = append(buf, 0)
	for len(buf)%4 != 0 {
		buf = append(buf, 0)
	}
	buf = append(buf, ',')
	buf = append(buf, []byte(typeTag)...)
	buf = append(buf, 0)
	for len(buf)%4 != 0 {
		buf = append(buf, 0)
	}
	if intArg != nil {
		var b [4]byte
		b[0] = byte(*intArg >> 24)
		b[1] = byte(*intArg >> 16)
		b[2] = byte(*intArg >> 8)
		b[3] = byte(*intArg)
		buf = append(buf, b[:]...)
	}
	return buf
}
