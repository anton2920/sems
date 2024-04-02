package main

type URL struct {
	Path  string
	Query string
}

type URLValue struct {
	Key    string
	Values []string
}

type URLValues []URLValue

func (vs URLValues) Get(key string) string {
	for i := 0; i < len(vs); i++ {
		if key == vs[i].Key {
			values := vs[i].Values
			if len(values) == 0 {
				return ""
			}
			return values[0]
		}
	}
	return ""
}

func (vs URLValues) GetMany(key string) []string {
	for i := 0; i < len(vs); i++ {
		if key == vs[i].Key {
			return vs[i].Values
		}
	}
	return nil
}

/* CharToByte returns ASCII-decoded character. For example, 'A' yields '\x0A'. */
func CharToByte(c byte) (byte, bool) {
	if c >= '0' && c <= '9' {
		return c - '0', true
	} else if c >= 'A' && c <= 'F' {
		return 10 + c - 'A', true
	} else {
		return '\x00', false
	}
}

func URLDecode(decoded []byte, encoded string) (int, bool) {
	var hi, lo byte
	var ok bool
	var n int

	for i := 0; i < len(encoded); i++ {
		if encoded[i] == '%' {
			hi = encoded[i+1]
			hi, ok = CharToByte(hi)
			if !ok {
				return 0, false
			}

			lo = encoded[i+2]
			lo, ok = CharToByte(lo)
			if !ok {
				return 0, false
			}

			decoded[n] = byte(hi<<4 | lo)
			i += 2
		} else if encoded[i] == '+' {
			decoded[n] = ' '
		} else {
			decoded[n] = encoded[i]
		}
		n++
	}
	return n, true
}
