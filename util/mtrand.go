package util

const (
	MT_N               = 624
	MT_M               = 397
	MT_MATRIX_A   uint = 0x9908b0df
	MT_UPPER_MASK uint = 0x80000000
	MT_LOWER_MASK uint = 0x7fffffff
)

type MTRand struct {
	mag01    []uint
	mt       []uint
	mti      int
	rndTimes int
}

func NewMTRand(seed uint) *MTRand {
	mt := make([]uint, MT_N, MT_N)
	mt[0] = seed & 0xffffffff
	for i := 1; i < MT_N; i++ {
		mt[i] = uint(1812433253*(mt[i-1]^(mt[i-1]>>30)) + uint(i))
		mt[i] &= 0xffffffff
	}
	return &MTRand{
		mag01: []uint{0x0, MT_MATRIX_A},
		mt:    mt,
		mti:   MT_N,
	}
}

func (self *MTRand) Random(min, max int) int {
	if min > max {
		min, max = max, min
	}
	ret := min + self.next()%(max-min)
	self.rndTimes++
	return ret
}

// 梅林旋转伪随机算法
func (self *MTRand) next() int {
	var y uint
	if self.mti >= MT_N {
		var k int
		for k = 0; k < MT_N-MT_M; k++ {
			y = (self.mt[k] & MT_UPPER_MASK) | (self.mt[k+1] & MT_LOWER_MASK)
			self.mt[k] = self.mt[k+MT_M] ^ (y >> 1) ^ self.mag01[y&0x1]
		}
		for ; k < MT_N-1; k++ {
			y = (self.mt[k] & MT_UPPER_MASK) | (self.mt[k+1] & MT_LOWER_MASK)
			self.mt[k] = self.mt[k+(MT_M-MT_N)] ^ (y >> 1) ^ self.mag01[y&0x1]
		}
		y = (self.mt[MT_N-1] & MT_UPPER_MASK) | (self.mt[0] & MT_LOWER_MASK)
		self.mt[MT_N-1] = self.mt[MT_M-1] ^ (y >> 1) ^ self.mag01[y&0x1]
		self.mti = 0
	}
	y = self.mt[self.mti]
	self.mti++
	y ^= y >> 11
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= y >> 18
	return int(y >> 1)
}
