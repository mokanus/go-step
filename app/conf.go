package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"github.com/mokanus/go-step/pkg/github.com/globalsign/mgo"
	"github.com/mokanus/go-step/pkg/github.com/gomodule/redigo/redis"
	"github.com/mokanus/go-step/stat"
	"github.com/mokanus/go-step/util"
	"io/ioutil"
	"runtime/debug"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// 配置的结构定义
// ----------------------------------------------------------------------------
type Config struct {
	Area             string          `json:",omitempty"` // 区域
	IP               string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	Port             int32           `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	DataPath         string          `json:",omitempty"` // 生成conf时，ServiceApp不填，GameApp必填
	StatPath         string          `json:",omitempty"` // 生成conf时，ServiceApp不填，GameApp必填
	LogPath          string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	LogLevel         string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	DefaultDBAddr    string          `json:",omitempty"` // 生成conf时，ServiceApp必填，GameApp不填
	DefaultDBName    string          `json:",omitempty"` // 生成conf时，ServiceApp必填，GameApp不填
	DefaultRedisAddr string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	DefaultRedisPass string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	Maintain         *MaintainConfig `json:",omitempty"` // 生成conf时，只有FrontApp会填
	CreateTime       string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	Season           int32           `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	SeasonTime       string          `json:",omitempty"` // 生成conf时，ServiceApp和GameApp必填
	ServiceName      string          `json:",omitempty"` // 生成conf时，部分ServiceApp会填
	ServiceTurn      int32           `json:",omitempty"` // 生成conf时，部分ServiceApp会填
	RegionList       []*RegionConfig `json:",omitempty"` // 生成conf时，ServiceApp用来填关联的区服且只关心区服部分字段，GameApp用来填下辖区服且只关心ID;Name;DB三个字段且必有本ID区服

	Custom map[string]string `json:",omitempty"` // 非自动生成的conf可用的自定义配置字段（目前只有AdminApp有用）

	RegionMap       map[int32]*RegionConfig `json:"-"`
	CreateTimeStamp uint32                  `json:"-"` // 加载配置时初始化填充，会确保为0点
	SeasonTimeStamp uint32                  `json:"-"` // 加载配置时初始化填充，会确保为0点
}

type MaintainConfig struct {
	AuditAddr      string
	AuditRegionID  int32
	Ver            string
	WhiteList      string
	MaintainNotice string
	FullNotice     string
}

type RegionConfig struct {
	ID      int32           `json:",omitempty"`
	Name    string          `json:",omitempty"`
	State   int32           `json:",omitempty"`
	MergeID int32           `json:",omitempty"`
	Addr    string          `json:",omitempty"`
	QRYAddr string          `json:",omitempty"`
	RPCAddr string          `json:",omitempty"`
	DB      *RegionConfigDB `json:",omitempty"`
}

type RegionConfigDB struct {
	Addr string `json:",omitempty"`
	Name string `json:",omitempty"`
}

// ----------------------------------------------------------------------------
// 配置的加载、检查与刷新相关状态
// ----------------------------------------------------------------------------
// 注意：后台可以通知App调用loadConfig函数来重新载入配置的！
func LoadConf() (string, error) {
	// 读取配置文件
	content, err := ioutil.ReadFile("./conf")
	if err != nil {
		return "error", err
	}

	// 解析配置文件
	config := new(Config)
	if err := json.Unmarshal(content, config); err != nil {
		return "error", err
	}
	if err := checkAndFillConfig(config); err != nil {
		return "error", err
	}

	Conf = config

	var outStd = config.LogLevel != "Debug_CloseStdout"

	// 将配置项应用到log的设置去
	log.Init(
		log.WithAppName(fmt.Sprintf("%s-%s", Env, Type)),
		log.WithRegionId(ID),
		log.WithLevel(log.Level(strings.ToLower(Conf.LogLevel))),
		log.WithStdout(outStd, "console"),
		log.WithFileOut(true, Conf.LogPath, true),
	)

	// 将配置项应用到stat的设置去
	stat.Config(Conf.StatPath)

	// 如果配置了DefaultDB，就去连接；如果新配置DefaultDB与旧配置DefaultDB不一致，更新连接状态
	refreshDefaultDbConn()

	// 如果配置了DefaultRedis，就去连接；如果新配置DefaultRedis与旧配置DefaultRedis不一致，更新连接状态
	refreshDefaultRedisPool()

	// 如果配置了有DB关联的区服，就去连接相应的DB
	refreshRegionDbConnMap()

	// 加载配置成功，返回最新配置的md5
	return util.MD5Conf(content), nil
}

// 加载配置时的操作：检查并完善配置项
func checkAndFillConfig(config *Config) error {
	// 如果配置了DefaultDBAddr，那一定要配置DefaultDBName的
	if (config.DefaultDBAddr == "" && config.DefaultDBName != "") || (config.DefaultDBAddr != "" && config.DefaultDBName == "") {
		return errors.New("DefaultDBAddr和DefaultDBAddr配置错误！")
	}

	if config.DataPath == "" {
		config.DataPath = "./data"
	}

	if config.LogLevel == "" {
		config.LogLevel = "DEBUG"
	}

	if config.CreateTime != "" {
		createTime, err := time.ParseInLocation(util.GoBeginTimeString, config.CreateTime, time.Local)
		if err != nil {
			return fmt.Errorf("开服时间[%s]格式错误！", config.CreateTime)
		}
		// 确保为0点的时间戳
		createTimeStamp := time.Date(createTime.Year(), createTime.Month(), createTime.Day(), 0, 0, 0, 0, createTime.Location()).Unix()
		if createTimeStamp < 0 {
			return fmt.Errorf("开服时间[%s]格式错误！", config.CreateTime)
		}
		config.CreateTimeStamp = uint32(createTimeStamp)
	}

	if config.SeasonTime != "" {
		seasonTime, err := time.ParseInLocation(util.GoBeginTimeString, config.SeasonTime, time.Local)
		if err != nil {
			return fmt.Errorf("赛季时间[%s]格式错误！", config.SeasonTime)
		}
		// 确保为0点的时间戳
		seasonTimeStamp := time.Date(seasonTime.Year(), seasonTime.Month(), seasonTime.Day(), 0, 0, 0, 0, seasonTime.Location()).Unix()
		if seasonTimeStamp < 0 {
			return fmt.Errorf("赛季时间[%s]格式错误！", config.SeasonTime)
		}
		config.SeasonTimeStamp = uint32(seasonTimeStamp)
	}

	// game和zone要求一定有CreateTime
	if Type == TypeGame || Type == TypeZone {
		if config.CreateTimeStamp <= 0 {
			return fmt.Errorf("%s服必定配置正确的开服时间！", Type)
		}
	}

	// game一定要有正确的SeasonTime、Season
	if Type == TypeGame {
		if config.SeasonTimeStamp <= 0 {
			return fmt.Errorf("%s服必定配置正确的赛季开始时间！", Type)
		}
		if config.Season <= 0 {
			return errors.New("game必须配置Season字段！")
		}
	}

	// zone的ServiceName加载时处理为一定不空，为Season则是赛季zone，其余是玩法zone
	if Type == TypeZone {
		// ServiceName为空的转为Season（兼容线上旧配置）
		if config.ServiceName == "" {
			config.ServiceName = ServiceSeason
		}
		if config.ServiceName == ServiceSeason {
			// 赛季zone，会确保Season一定大于0
			if config.Season <= 0 {
				config.Season = 1
			}
		} else {
			// 玩法zone，会确保Season为0
			config.Season = 0
		}
	}

	// 自定义项如果为nil，初始化成空map
	if config.Custom == nil {
		config.Custom = make(map[string]string)
	}

	// 区服列表中的区服配置字段完善一下，然后构造出一份RegionMap来
	regionMap := make(map[int32]*RegionConfig)
	for _, region := range config.RegionList {
		if region.MergeID == 0 {
			region.MergeID = region.ID
		}
		if region.DB != nil {
			if region.DB.Addr == "" || region.DB.Name == "" {
				return errors.New("RegionList.DB配置错误！")
			}
		}
		regionMap[region.ID] = region
	}
	// 如果GameApp，本ID的配置要求一定存在
	if Type == TypeGame {
		if regionMap[ID] == nil {
			return fmt.Errorf("本GameApp（ID=%d）的RegionConfig不存在！", ID)
		}
	}

	config.RegionMap = regionMap

	return nil
}

// 加载配置时的操作：刷新与默认数据库的连接状态
func refreshDefaultDbConn() {
	// 如果原先已有数据库连接，检查地址是否有变。如果有变，要断开旧的
	if defaultDbConn != nil {
		if Conf.DefaultDBAddr == "" || Conf.DefaultDBAddr != defaultDbConn.addr {
			defaultDbConn.session.Close()
			defaultDbConn = nil
			log.GetLogger().Info("加载配置时，关闭默认数据库连接！", field.String("addr", defaultDbConn.addr))
		}
	}

	// 原先没有数据库连接，或有连接被替换了，则建立新连接
	if Conf.DefaultDBAddr != "" && defaultDbConn == nil {
		if session, err := mgo.Dial(Conf.DefaultDBAddr); err == nil {
			defaultDbConn = &DbConn{Conf.DefaultDBAddr, session}
			log.GetLogger().Info("加载配置时，连接默认数据库成功！", field.String("addr", Conf.DefaultDBAddr))
		} else {
			log.GetLogger().Error("加载配置时，新建默认数据库连接失败!", field.String("addr", Conf.DefaultDBAddr), field.Error(err))
		}
	}
}

// 加载配置时的操作：刷新与默认Redis的连接状态
func refreshDefaultRedisPool() {
	// 如果原先已有Redis连接，检查地址是否有变。如果有变，要断开旧的
	if defaultRedisConn != nil {
		if Conf.DefaultRedisAddr == "" || Conf.DefaultRedisAddr != defaultRedisConn.addr {
			defaultRedisConn.pool.Close()
			defaultRedisConn = nil
			log.GetLogger().Info("加载配置时，关闭默认Redis连接！", field.String("redis_addr", defaultRedisConn.addr))
		}
	}

	// 原先没有Redis连接，或有连接被替换了，则建立新连接
	if Conf.DefaultRedisAddr != "" && defaultRedisConn == nil {
		defaultRedisConn = &RedisConn{
			addr: Conf.DefaultRedisAddr,
			pool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					conn, err := redis.Dial("tcp", Conf.DefaultRedisAddr)
					if err != nil {
						return nil, err
					}
					if Conf.DefaultRedisPass == "" {
						return conn, nil
					}
					if _, err := conn.Do("AUTH", Conf.DefaultRedisPass); err != nil {
						conn.Close()
						return nil, err
					}
					return conn, nil
				},
				MaxIdle:   1,
				MaxActive: 1,
				Wait:      true,
			},
		}
		log.GetLogger().Info("加载配置时，连接默认Redis成功！", field.String("redis_addr", Conf.DefaultRedisAddr))
	}
}

// 加载配置时的操作：刷新与区服数据库的连接状态
func refreshRegionDbConnMap() {
	latestDbAddrMap := make(map[string]bool)
	for _, region := range Conf.RegionList {
		if region.DB != nil {
			latestDbAddrMap[region.DB.Addr] = true
		}
	}

	// 构造一份新的regionDBMap：遍历新配置，如果DBAddr有在老的DBMap里，session拿过来；如果没在老的DBMap里，则新建。
	newMap := make(map[string]*DbConn)
	for addr := range latestDbAddrMap {
		dbConn := getRegionDbConn(addr)
		if dbConn != nil {
			newMap[addr] = dbConn
		} else {
			if s, err := mgo.Dial(addr); err == nil {
				newMap[addr] = &DbConn{addr, s}
				log.GetLogger().Info("加载配置时，连接区服数据库成功！", field.String("addr", addr))
			} else {
				log.GetLogger().Error("加载配置时，新建区服数据库连接失败！", field.String("addr", addr), field.Error(err))
			}
		}
	}

	// 把老的DBMap的引用拿出来，然后把新的DBMap赋给regionDBMap。这样业务中就可以使用上新的regionDBMap了。
	regionDbConnMapLocker.Lock()
	oldMap := regionDbConnMap
	regionDbConnMap = newMap
	regionDbConnMapLocker.Unlock()

	// 遍历一遍老的DBMap，如果DBAddr已经不在新配置里了，那就可以关闭这个DBAddr对应的连接了。
	for addr, dbConn := range oldMap {
		if !latestDbAddrMap[addr] {
			dbConn.session.Close()
			log.GetLogger().Info("加载配置时，关闭区服数据库连接！", field.String("addr", addr))
		}
	}
}

// ----------------------------------------------------------------------------
// 配置的定时刷新相关状态
// ----------------------------------------------------------------------------
func startRefreshRegionRpcConnMap() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.GetLogger().Error("RefreshRegionRpcConnMap协程崩溃了", field.Any("stack", debug.Stack()), field.Any("error", err))
			}
		}()
		for {
			refreshRegionRpcConnMap()
			time.Sleep(time.Second)
		}
	}()
}

// App启动之后，每秒用最新配置去刷新当前rpc连接状态
func refreshRegionRpcConnMap() {
	// 根据serviceConfig.RegionsRPC，确定出服务要rpc连接这些区服，收集这些区服的RPCAddr，用来刷新regionRpcConnMap！
	// 注意：配置了RPCAddr，则一定要配置ServiceList，否则不会去建立对区服的RPC连接
	latestRpcAddrMap := make(map[string]bool)
	for _, region := range Conf.RegionList {
		if region.RPCAddr != "" {
			latestRpcAddrMap[region.RPCAddr] = true
		}
	}

	// 构造一份新的regionRpcMap：遍历新配置，如果RPCAddr有在老的RpcConnMap里，conn拿过来；如果没在老的RpcConnMap里，则新建。
	// 然后检查将connected为false的连接连上。这个过程发生在新的Map里，旧的Map业务层是仍可用的。
	newMap := make(map[string]*RpcConn)
	for addr := range latestRpcAddrMap {
		conn := getRegionRpcConn(addr)
		if conn == nil {
			conn = NewRpcConn(nil, false)
		}
		newMap[addr] = conn
		if !conn.connected {
			conn.dial(addr)
		}
	}

	// 把老的RpcConnMap的引用拿出来，然后把新的RpcConnMap赋给regionRpcConnMap。这样业务中就可以使用上新的regionRpcConnMap了。
	regionRpcConnMapLocker.Lock()
	oldMap := regionRpcConnMap
	regionRpcConnMap = newMap
	regionRpcConnMapLocker.Unlock()

	// 遍历一遍老的RpcConnMap，如果RPCAddr已经不在新配置里了，那就可以关闭这个RPCAddr对应的连接了。
	for addr, conn := range oldMap {
		if !latestRpcAddrMap[addr] {
			conn.close()
		}
	}
}

// 是否是特殊战区
func (self *Config) IsSpecialZone() bool {
	if Type == TypeZone && self.ServiceName != ServiceSeason {
		return true
	}
	return false
}

// 是否是特定功能性特殊战区
func (self *Config) IsSpecialZoneByName(name string) bool {
	if Type == TypeZone && self.ServiceName == name {
		return true
	}
	return false
}
