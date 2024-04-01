package main

func SlicePutInt(buf []byte, x int) int {
	var ndigits int
	var rx, i int

	if x == 0 {
		buf[0] = '0'
		return 1
	}

	if x < 0 {
		x = -x
		buf[0] = '-'
		i++
	}

	for x > 0 {
		rx = (10 * rx) + (x % 10)
		x /= 10
		ndigits++
	}

	for ndigits > 0 {
		buf[i] = byte((rx % 10) + '0')
		i++

		rx /= 10
		ndigits--
	}
	return i
}

func SlicePutTm(buf []byte, tm Tm) int {
	var n, ndigits int

	if tm.Mday < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Mday)
	n += ndigits
	buf[n] = '.'
	n++

	if tm.Mon+1 < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Mon+1)
	n += ndigits
	buf[n] = '.'
	n++

	ndigits = SlicePutInt(buf[n:], tm.Year+1900)
	n += ndigits
	buf[n] = ' '
	n++

	if tm.Hour < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Hour)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Min < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Min)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Sec < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Sec)
	n += ndigits
	buf[n] = ' '
	n++

	buf[n] = 'M'
	buf[n+1] = 'S'
	buf[n+2] = 'K'

	return n + 3
}

func SlicePutTmRFC822(buf []byte, tm Tm) int {
	var n, ndigits int

	var wdays = [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	var months = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	n += copy(buf[n:], wdays[tm.Wday])
	buf[n] = ','
	buf[n+1] = ' '
	n += 2

	if tm.Mday < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Mday)
	n += ndigits
	buf[n] = ' '
	n++

	n += copy(buf[n:], months[tm.Mon])
	buf[n] = ' '
	n++

	ndigits = SlicePutInt(buf[n:], tm.Year+1900)
	n += ndigits
	buf[n] = ' '
	n++

	if tm.Hour < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Hour)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Min < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Min)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Sec < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutInt(buf[n:], tm.Sec)
	n += ndigits
	buf[n] = ' '
	n++

	buf[n] = '+'
	buf[n+1] = '0'
	buf[n+2] = '3'
	buf[n+3] = '0'
	buf[n+4] = '0'

	return n + 5
}
