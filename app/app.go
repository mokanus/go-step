package app

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/golang/protobuf/proto"
	"github.com/mokanus/go-step/base36"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"github.com/mokanus/go-step/uid"
	"github.com/mokanus/go-step/util"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	exit = make(chan os.Signal)

	// 这三个在MD5()函数中组合成这个App的运行时校验码，格式为：CodeMD5-DataMD5-ConfMD5
	// CodeMD5是编译时设置；DataMD5和ConfMD5是App启动时计算
	CodeMD5 = "noset"
	DataMD5 = "noset"
	ConfMD5 = "noset"

	// Env、Type和ID是直接通过解析exe名而得。
	// 对于ServiceApp而言，appType就是serviceName。各个GameApp将通过serviceName来访问ServiceApp。而ID就是同一service下的多个索引。
	// 对于GameApp而言，appType固定就是game。而ID就是regionID。如果合服了，ID就是主服的regionID。
	Env  string
	Type string
	ID   int32

	Conf *Config

	// 用户可注册：启停管理器
	managers []Manager

	// 用户可注册：rpc和req协议号的枚举名
	protoNames = make(map[uint16]string)

	// 用户可注册：rpc消息处理
	rpcHandlers       = make(map[uint16]func(*RPC))
	defaultRpcHandler func(*RPC)

	// 用户可注册：lap消息处理
	lapHandler func(*LAP)

	// 用户可注册：req消息处理
	reqHandlers       = make(map[uint16]func(*REQ))
	defaultReqHandler func(*REQ)
	cnnlostReqHandler func(*REQ)

	// 用户可注册：qry消息处理
	qryPublicHandlers        = make(map[string]func(r *QRY))
	defaultQryPublicHandler  func(r *QRY)
	qryPrivateHandlers       = make(map[string]func(r *QRY))
	defaultQryPrivateHandler func(r *QRY)

	// App收到的rpc请求（call、cast）、lap请求，会根据channelUid放入对应的channel队列中，每个channel由一个协程串行地处理属于该channel的请求。
	channelMap       = make(map[string]*Channel)
	channelMapLocker = new(sync.RWMutex)

	// 只有ServiceApp会有defaultDbConn
	// ServiceApp的regionDbConnMap代表服务与这些区服的数据库有关联；GameApp的regionDbConnMap代表这些区服已合到本GameApp。
	// 注意：在regionDbConnMap中，不是一个region对应一个session，而是一个DBAddr对应一个session
	defaultDbConn         *DbConn
	regionDbConnMap       map[string]*DbConn
	regionDbConnMapLocker = new(sync.RWMutex)

	defaultRedisConn *RedisConn

	// 只有ServiceApp会有regionRpcConnMap。ServiceApp主动去连接GameApp从而得到的regionRpcConnMap并主动维护regionRpcConnMap，确保regionRpcConnMap是最新且可用的。
	// regionRpcConnMap的key是区服的RPCAddr，合服时，多个区服是可以共用RPCAddr的，不必一个区服一个连接。
	// 只有GameApp会有serviceRpcConnMap。GameApp是被动接收到来自ServiceApp的连接，从而得到ServiceRpcConnMap，如果连接断开了，就记connected为false，不做多余维护，即使有废弃的也依然会在内存中。
	// serviceRpcConnMap的key是season。GameApp要使用服务时，要指定season从而找到serviceRpcConn。
	// 每个season一定只会对应一个ServiceApp。在出现错误一个season有多个ServiceApp的情况下，以后来的为准。（所以如果配错了，纠正过后，正确的那个ServiceApp要重启一下）
	regionRpcConnMap       = make(map[string]*RpcConn)
	regionRpcConnMapLocker = new(sync.RWMutex)

	serviceRpcConnMap       = make(map[string]*RpcConn)
	serviceRpcConnMapLocker = new(sync.RWMutex)
)

func Init() {
	Env, Type, ID = parseExeName()

	confMD5, err := LoadConf()
	if err != nil {
		panic(err)
	}
	ConfMD5 = confMD5

	log.GetLogger().Info("服务器初始化", field.String("env", Env), field.String("type", Type), field.Int32("Id", ID))

	if err := LoadSvrToken(); err != nil {
		panic(err)
	}
	log.GetLogger().Info("当前运行令牌", field.Uint32("token", GetSvrToken()))
	log.GetLogger().Info("当前所处赛季", field.Int32("season", Conf.Season))
}

func Exec() {
	log.GetLogger().Info("服务器启动完成！", field.String("type", Type), field.Int32("id", ID))

	// 开启多核
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 配置表加载
	if dataMap, errorLog := LoadData(); errorLog != "" {
		panic(errorLog)
		return
	} else {
		DataMD5 = util.MD5Data(dataMap)
	}

	// 计算程序、数据、配置的校验码
	log.GetLogger().Info("运行时校验码", field.String("md5", MD5()))

	// 启动网络服务
	startWebServer()

	// 启动一个循环去维护regionRpcConnMap
	startRefreshRegionRpcConnMap()

	// 管理器启动
	for _, m := range managers {
		m.Init()
	}

	// 等待退出信号
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit

	// 管理器停止
	for _, m := range managers {
		m.Stop()
	}

	log.GetLogger().Info("服务器退出！", field.String("type", Type), field.Int32("id", ID))
}

// ----------------------------------------------------------------------------
// 注册接口
// ----------------------------------------------------------------------------
func RegisterRegionDBIndex(collection string, indexName string, isUnique bool) {
	index := mgo.Index{
		Key:    []string{indexName},
		Unique: isUnique,
	}
	for _, region := range Conf.RegionList {
		if region.DB != nil {
			dbConn := getRegionDbConn(region.DB.Addr)
			dbConn.session.DB(region.DB.Name).C(collection).EnsureIndex(index)
		}
	}
}

func RegisterDefaultDBIndex(collection string, indexName string, isUnique bool) {
	if defaultDbConn == nil {
		panic("RegisterDBIndex失败！数据库未配置！")
	}
	defaultDbConn.session.DB(Conf.DefaultDBName).C(collection).EnsureIndex(mgo.Index{
		Key:    []string{indexName},
		Unique: isUnique,
	})
}

func RegisterDataLoader(name string, load interface{}) {
	if loadFunc, ok := load.(func([]byte) error); ok {
		flatDataLoaders = append(flatDataLoaders, FlatDataLoader{name: name, load: loadFunc})
		return
	}
	if loadFunc, ok := load.(func([]*SubDataFile) error); ok {
		nestDataLoaders = append(nestDataLoaders, NestDataLoader{name: name, load: loadFunc})
		return
	}
	panic("RegisterDataLoader失败！load函数应为【func([]byte) error】或【func([]*SubDataFile) error】")
}

func RegisterManager(manager Manager) {
	managers = append(managers, manager)
}

func RegisterProtoNames(exterProtoNames, innerProtoNames map[int32]string) {
	for code, name := range exterProtoNames {
		protoNames[uint16(code)] = name
	}
	for code, name := range innerProtoNames {
		protoNames[uint16(code)] = name
	}
}

func RegisterReqHandler(route interface{}, handler func(*REQ)) {
	if r, ok := route.(string); ok {
		switch r {
		case "*":
			defaultReqHandler = handler
		case "lost":
			cnnlostReqHandler = handler
		default:
			panic("RegisterReqHandler失败！协议号应为ProtoCode或*")
		}
		return
	}
	if r, ok := route.(ProtoCode); ok {
		reqHandlers[r.Value()] = handler
		return
	}
	panic("RegisterReqHandler失败！协议号应为ProtoCode或*")
}

func RegisterRpcHandler(route interface{}, handler func(*RPC)) {
	if r, ok := route.(string); ok {
		if r != "*" {
			panic("RegisterRpcHandler失败！协议号应为ProtoCode或*")
		}
		defaultRpcHandler = handler
		return
	}
	if r, ok := route.(ProtoCode); ok {
		rpcHandlers[r.Value()] = handler
		return
	}
	panic("RegisterRpcHandler失败！协议号应为ProtoCode或*")
}

func RegisterLapHandler(handler func(*LAP)) {
	lapHandler = handler
}

// 注意！！只有AdminApp有提供web页面服务所以要注册defaultQryPublicHandler。其余App都不应该注册defaultQryPublicHandler，
// 否则所有privateHandler都不会生效。而AdminApp是必定不需要privateHandler的。
func RegisterQryPublicHandler(r string, handler func(r *QRY)) {
	if r == "*" {
		defaultQryPublicHandler = handler
	} else {
		qryPublicHandlers["/"+r] = handler
	}
}

func RegisterQryPrivateHandler(r string, handler func(r *QRY)) {
	if r == "*" {
		defaultQryPrivateHandler = handler
	} else {
		qryPrivateHandlers["/"+r] = handler
	}
}

func RegisterQryAuthUsers(authFilePath string) {
	fileBytes, err := ioutil.ReadFile(authFilePath)
	if err != nil {
		panic(err)
	}
	authUsers := make(map[string]string)
	if err := json.Unmarshal(fileBytes, &authUsers); err != nil {
		panic(fmt.Sprintf("解析auth文件失败：%v", err))
	}
	qryAuthUserMap = authUsers
}

// ----------------------------------------------------------------------------
// 数据库接口
// ----------------------------------------------------------------------------
func DefaultDB(collection string) *DbAgent {
	if defaultDbConn == nil {
		return &DbAgent{
			err: fmt.Errorf("DefaultDB[%s]未连接！", Conf.DefaultDBName),
		}
	}
	defaultDbConn.session.Refresh()
	return &DbAgent{
		collection: defaultDbConn.session.DB(Conf.DefaultDBName).C(collection),
	}
}

func RegionDB(regionID int32, collection string) *DbAgent {
	regionConfig := Conf.RegionMap[regionID]
	if regionConfig == nil {
		return &DbAgent{
			err: fmt.Errorf("区服[%d]的配置不存在！", regionID),
		}
	}
	if regionConfig.DB == nil {
		return &DbAgent{
			err: fmt.Errorf("区服[%d]的数据库配置不存在！", regionID),
		}
	}
	dbConn := getRegionDbConn(regionConfig.DB.Addr)
	if dbConn == nil {
		return &DbAgent{
			err: fmt.Errorf("区服[%d]的数据库连接未建立！", regionID),
		}
	}
	dbConn.session.Refresh()
	return &DbAgent{
		collection: dbConn.session.DB(regionConfig.DB.Name).C(collection),
	}
}

// ----------------------------------------------------------------------------
// 向战区发RPC请求
// ----------------------------------------------------------------------------
func SeasonRPC() *RpcAgent {
	return ServiceRPC(fmt.Sprintf("%s%d", ServiceSeason, Conf.Season))
}

func ServiceRPC(serviceName string) *RpcAgent {
	conn := getServiceRpcConn(serviceName)
	if conn == nil {
		return &RpcAgent{
			err: fmt.Errorf("服务[%s]的RPC连接不存在！", serviceName),
		}
	}
	return &RpcAgent{
		name: fmt.Sprintf("Service:%s", serviceName),
		conn: conn,
	}
}

// ----------------------------------------------------------------------------
// 向区服发RPC请求
// ----------------------------------------------------------------------------
func RegionRPC(regionID int32) *RpcAgent {
	regionConfig := Conf.RegionMap[regionID]
	if regionConfig == nil {
		return &RpcAgent{
			err: fmt.Errorf("区服[%d]的配置不存在！", regionID),
		}
	}
	if regionConfig.RPCAddr == "" {
		return &RpcAgent{
			err: fmt.Errorf("区服[%d]的RPC配置不存在！", regionID),
		}
	}
	conn := getRegionRpcConn(regionConfig.RPCAddr)
	if conn == nil {
		return &RpcAgent{
			err: fmt.Errorf("区服[%d]的RPC连接不存在！", regionID),
		}
	}
	return &RpcAgent{
		name: fmt.Sprintf("Region:%d", regionID),
		conn: conn,
	}
}

func RegionPlayerRPC(playerUid string) *RpcAgent {
	regionID := ParseRUID(playerUid)
	if regionID == 0 {
		return &RpcAgent{
			err: fmt.Errorf("无法从[%s]中解析出regionID！", playerUid),
		}
	}
	regionConfig := Conf.RegionMap[regionID]
	if regionConfig == nil {
		return &RpcAgent{
			err: fmt.Errorf("区服[%d]的配置不存在！", regionID),
		}
	}
	if regionConfig.RPCAddr == "" {
		return &RpcAgent{
			err: fmt.Errorf("区服[%d]的RPC配置不存在！", regionID),
		}
	}
	conn := getRegionRpcConn(regionConfig.RPCAddr)
	if conn == nil {
		return &RpcAgent{
			err: fmt.Errorf("区服[%d]的RPC连接不存在！", regionID),
		}
	}
	return &RpcAgent{
		name:       fmt.Sprintf("Region:%d:%s", regionID, playerUid),
		conn:       conn,
		channelUid: playerUid,
	}
}

// ----------------------------------------------------------------------------
// 向本进程玩家发请求的两种方式：1、本地RPC；2、异步调用
// ----------------------------------------------------------------------------
func LocalCast(playerUid string, rpcType ProtoCode, rpcData proto.Message) error {
	rpcBody, err := util.PbMarshal(rpcData)
	if err != nil {
		log.GetLogger().Error("local cast error", field.Any("rpc_type", rpcType), field.Error(err))
		return ErrServer
	}

	if len(rpcBody) > rpcBodySizeLimit {
		log.GetLogger().Error("local cast error 包体过大", field.Any("rpc_type", rpcType))
		return ErrServer
	}

	channel := getChannel(playerUid)

	select {
	case channel.rpcQueue <- NewRPC(nil, 0, playerUid, rpcType.Value(), rpcBody):
		return nil
	default:
		log.GetLogger().Error("local cast(%v) Error: channel写入失败！", field.Any("rpc_type", rpcType))
		return ErrServer
	}
}

func LocalAsyncProcess(playerUid string, funName string, fun interface{}, arg interface{}, offline bool, offlineInit bool) {
	if funName != "" {
		log.GetLogger().Info("LocalAsyncProcess -> ", field.String("player_uid", playerUid), field.String("fun_name", funName))
	}

	channel := getChannel(playerUid)

	select {
	case channel.lapQueue <- NewLAP(playerUid, fun, arg, offline, offlineInit):
		return
	default:
		log.GetLogger().Error("lap channel full!", field.String("player_uid", playerUid))
	}
}

// ----------------------------------------------------------------------------
// 向指定地址发送QRY请求
// ----------------------------------------------------------------------------
func AddrQRY(qryAddr string) *QryAgent {
	return &QryAgent{
		addr: qryAddr,
	}
}

func AddrChannelQRY(qryAddr string, channelUid string) *QryAgent {
	return &QryAgent{
		addr:       qryAddr,
		channelUid: channelUid,
	}
}

// ----------------------------------------------------------------------------
// 杂散功能函数
// ----------------------------------------------------------------------------
// 获取本App的运行时检验码
func MD5() string {
	return fmt.Sprintf("%s-%s-%s", CodeMD5, DataMD5, ConfMD5)
}

// 构造包含RegionID信息的全局UID
// 特殊处理1：channel为"1"，代表抖音大区，抖音大区的RUid统一加上字母O做标识区分。（默认的RUid的首字母最大只会到base36的字母0）。后面有新的大区，就继续分配新的标识。
// 特殊处理2：channel为"100"，代表支付宝大区，支付宝大区的RUid统一加上字母L做标识区分。（默认的RUid的首字母最大只会到base36的字母0）。后面有新的大区，就继续分配新的标识。
// TODO: 因为涉及到业务层了，所以这两个函数应该移到业务层去（放common）。将渠道相关的处理归在一起。
func MakeRUID(channel string, regionID int32) string {
	switch channel {
	case "1":
		return "O" + uid.Gen("", base36.Encode(uint64(regionID)))
	case "100":
		return "L" + uid.Gen("", base36.Encode(uint64(regionID)))
	case "3048011", "3048010":
		return "E" + uid.Gen("", base36.Encode(uint64(regionID)))
	default:
		return uid.Gen("", base36.Encode(uint64(regionID)))
	}
}

// 从RUID中解析出regionID
// 特殊处理1：如果RUid的首字母是O，代表是抖音大区，去掉首字母O之后再解出RegionID
// 特殊处理2：如果RUid的首字母是L，代表是支付宝大区，去掉首字母L之后再解出RegionID
func ParseRUID(RUid string) int32 {
	if len(RUid) > 0 {
		firstChar := string(RUid[0])
		if firstChar == "O" || firstChar == "L" || firstChar == "E" {
			RUid = RUid[1:]
		}
	}

	prefix := uid.Prefix(RUid)
	if prefix == "" {
		return 0
	}
	v, ok := base36.Decode(prefix)
	if !ok {
		return 0
	}
	if v > math.MaxInt32 {
		return 0
	}
	return int32(v)
}

func SystemLoad() (sPid, sLoad, sNumCPU, sMem, sNumGoroutine string) {
	pid := os.Getpid()
	numCPU := runtime.NumCPU()
	numGoroutine := runtime.NumGoroutine()
	sLoad = "-"
	sMem = "-"

	// 获取load信息
	if result, err := exec.Command("cat", "/proc/loadavg").Output(); err == nil {
		frags := strings.Split(strings.TrimSpace(string(result)), " ")
		if len(frags) >= 3 {
			sLoad = frags[0] + " " + frags[1] + " " + frags[2]
		}
	}

	// 获取mem信息
	if result, err := exec.Command("free", "-m").Output(); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(result)), "\n") {
			line = strings.TrimSpace(line)
			if len(line) >= 4 && line[:4] == "Mem:" {
				frags := make([]string, 0)
				for _, frag := range strings.Split(line, " ") {
					if frag != "" {
						frags = append(frags, frag)
					}
				}
				if len(frags) >= 3 {
					total, err1 := strconv.Atoi(frags[1])
					used, err2 := strconv.Atoi(frags[2])
					if err1 == nil && err2 == nil && total > 0 {
						percent := float64(used) / float64(total) * 100
						sMem = fmt.Sprintf("%d/%d(%.2f%%)", used, total, percent)
					}
				}
				break
			}
		}
	}

	sPid = fmt.Sprintf("%d", pid)
	sNumCPU = fmt.Sprintf("%d", numCPU)
	sNumGoroutine = fmt.Sprintf("%d", numGoroutine)
	return
}

func ServiceInfo() string {
	serviceRpcConnMapLocker.Lock()
	defer serviceRpcConnMapLocker.Unlock()

	var infoList []string
	for season, serviceRpcConn := range serviceRpcConnMap {
		if !serviceRpcConn.connected {
			infoList = append(infoList, fmt.Sprintf("%s:notconnected", season))
		} else if serviceRpcConn.socket == nil {
			infoList = append(infoList, fmt.Sprintf("%s:nosocket", season))
		} else {
			infoList = append(infoList, fmt.Sprintf("%s:%s", season, serviceRpcConn.socket.RemoteAddr()))
		}
	}

	return strings.Join(infoList, "|")
}

// ----------------------------------------------------------------------------
// 内部辅助函数
// ----------------------------------------------------------------------------
// 直接从exe名中解析出AppType和AppID，exe名规则：[Env:loc/dev/rel]_[Type:serviceName/game]_[ID]
// 格式规定：serviceName不能带下划线
func parseExeName() (appEnv string, appType string, appID int32) {
	exeName := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	sp := strings.Split(exeName, "_")
	if len(sp) != 3 {
		panic(fmt.Sprintf("可执行文件%s的文件名格式异常！", exeName))
		return
	}

	appEnv = strings.TrimSpace(sp[0])
	if appEnv == "" {
		panic(fmt.Sprintf("可执行文件%s的文件名格式异常！", exeName))
		return
	}

	appType = strings.TrimSpace(sp[1])
	if appType == "" {
		panic(fmt.Sprintf("可执行文件%s的文件名格式异常！", exeName))
		return
	}

	if v, err := strconv.Atoi(strings.TrimSpace(sp[2])); err == nil {
		appID = int32(v)
	} else {
		panic(fmt.Sprintf("可执行文件%s的文件名格式异常！", exeName))
	}

	return
}

func getRegionDbConn(dbAddr string) *DbConn {
	regionDbConnMapLocker.RLock()
	defer regionDbConnMapLocker.RUnlock()
	return regionDbConnMap[dbAddr]
}

func getRegionRpcConn(rpcAddr string) *RpcConn {
	regionRpcConnMapLocker.RLock()
	defer regionRpcConnMapLocker.RUnlock()
	return regionRpcConnMap[rpcAddr]
}

func getChannel(uid string) *Channel {
	channelMapLocker.Lock()
	defer channelMapLocker.Unlock()
	channel := channelMap[uid]
	if channel == nil {
		channel = NewChannel(uid)
		channelMap[uid] = channel
	}
	return channel
}

func delChannel(uid string) {
	channelMapLocker.Lock()
	defer channelMapLocker.Unlock()
	delete(channelMap, uid)
}

func addServiceRpcConn(conn *RpcConn, serviceName string) {
	serviceRpcConnMapLocker.Lock()
	defer serviceRpcConnMapLocker.Unlock()
	serviceRpcConnMap[serviceName] = conn
}

func getServiceRpcConn(serviceName string) *RpcConn {
	serviceRpcConnMapLocker.Lock()
	defer serviceRpcConnMapLocker.Unlock()
	return serviceRpcConnMap[serviceName]
}
