package base36

const base = 36

var src = []byte("CFGPD0MN1WX2RS3TU4IJ5HK6AB7YZ8QV9ELO")
var dst = map[byte]uint64{
	'C': 0,
	'F': 1,
	'G': 2,
	'P': 3,
	'D': 4,
	'0': 5,
	'M': 6,
	'N': 7,
	'1': 8,
	'W': 9,
	'X': 10,
	'2': 11,
	'R': 12,
	'S': 13,
	'3': 14,
	'T': 15,
	'U': 16,
	'4': 17,
	'I': 18,
	'J': 19,
	'5': 20,
	'H': 21,
	'K': 22,
	'6': 23,
	'A': 24,
	'B': 25,
	'7': 26,
	'Y': 27,
	'Z': 28,
	'8': 29,
	'Q': 30,
	'V': 31,
	'9': 32,
	'E': 33,
	'L': 34,
	'O': 35,
}

// 将一个数值编码成36进制字符串，因为是36进制的，所以：
// 1位最大值O    -> 36-1          = 35
// 2位最大值OO   -> 36*36-1       = 1295
// 3位最大值OOO  -> 36*36*36-1    = 46655
// 4位最大值OOOO -> 36*36*36*36-1 = 1679615
func Encode(n uint64) string {
	if n == 0 {
		return string(src[0])
	}

	// 这时的code列表，前面的是小位，后面的是大位
	code := make([]byte, 0)
	for n > 0 {
		i := n % base
		code = append(code, src[i])
		n = n / base
	}

	// code列表逆序后拼成字符串，就是36进制编码出来的正式值了
	l := len(code)
	for i := 0; i < l/2; i++ {
		code[i], code[l-i-1] = code[l-i-1], code[i]
	}
	return string(code)
}

func Decode(code string) (uint64, bool) {
	bytes := []byte(code)
	l := len(bytes)

	if l == 0 {
		return 0, false
	}

	var n uint64
	var x uint64 = 1
	for i := l - 1; i >= 0; i-- {
		ch := bytes[i]
		v, ok := dst[ch]
		if !ok {
			return 0, false
		}
		n += v * x
		x *= base
	}

	return n, true
}
