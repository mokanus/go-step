package uid

import (
	"go-step/base36"
	"sync"
	"time"
)

const (
	uidBeginTime = 1635696000000 // 2024-04-01 00:00:00（毫秒）
)

var (
	generatorMap       = make(map[string]*Generator)
	generatorMapLocker = new(sync.Mutex)
)

// domain+prefix用来确定一个域，域内的uid正常情况下都是不冲突的。其中，prefix会拼入uid中。

// 常用用法1：用来生成playerUid。domain放空，prefix填regionID的36进制串。这样得到的playerUid就是确保区服内唯一，且可以从playerUid反解出regionID。
// 常用用法2：用来生成游戏内itemUid。domain为playerUid，prefix为item类型前缀，这样得到的itemUid就是确保该玩家内唯一，且可以通过item类型前缀区分itemUid。prefix也可以放空不用。

// 生成uid的逻辑是：
// 每10毫秒一个分片，在这个分片内增长inc，这样分片+inc就得到唯一ID了。
// inc最大可达uint64，如果10毫秒内inc超过uint64了，就会得到重复的uid了。但这个情况不处理，业务逻辑应该去保证不会在10毫秒内生成这么多uid。

// 另外两种导致uid冲突情况为：
// 1. 服务器重启时间小于10毫秒（这个可以对服务器启动时间做最小时间控制）
// 2. 操作系统时间回退（这个问题无法处理，只能是运维上约定不能回退操作系统时间）

// 注意1：该方法生成的uid是不定长的。
// 注意2：如果uid有反解的需求，那prefix和period编码成36进制后，都不应该超过36^35-1，否则只用一个字母不够标识prefix和period的长度。
// 正常情况下，prefix和period都不可能达到这样的长度，真出现这样的情况下，后果仅仅是反解结果错误。业务逻辑自己考虑反解错误的后果。

func Gen(domain string, prefix string) string {
	generatorMapLocker.Lock()
	defer generatorMapLocker.Unlock()
	k := domain + prefix
	if generatorMap[k] == nil {
		generatorMap[k] = &Generator{
			prefix: prefix,
		}
	}
	return generatorMap[k].gen()
}

func Prefix(uid string) string {
	uidLen := uint64(len(uid))

	if uidLen == 0 {
		return ""
	}

	prefixLen, ok := base36.Decode(string(uid[0]))
	if !ok {
		return ""
	}

	if prefixLen > uidLen-1 {
		return ""
	}

	return uid[1 : 1+prefixLen]
}

type Generator struct {
	prefix   string
	lastTime uint64
	inc      uint64
}

func (self *Generator) gen() string {
	now := uint64(time.Now().UnixNano() / 1000000)
	t := (now - uidBeginTime) / 10

	if t != self.lastTime {
		self.inc = 0
		self.lastTime = t
	} else {
		self.inc++
	}

	period := base36.Encode(t)

	if self.prefix != "" {
		return base36.Encode(uint64(len(self.prefix))) + // prefix的长度的36进制
			self.prefix + // prefix
			base36.Encode(uint64(len(period))) + // period分片的长度的36进度
			period + // period分片
			base36.Encode(self.inc) // period分片内的增长序号
	} else {
		return base36.Encode(uint64(len(period))) + // period分片的长度的36进度
			period + // period分片
			base36.Encode(self.inc) // period分片内的增长序号
	}
}
