package nacos

import (
	"bytes"
	"github.com/cnfree/props/v3/kvs"
	"github.com/cnfree/props/v3/util"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

var _ kvs.ConfigSource = new(NacosClientPropsConfigSource)

// 通过key/value来组织，过滤root prefix后，替换/为.作为properties key
type NacosClientPropsConfigSource struct {
	kvs.MapProperties
	// Required Configuration ID. Use a naming rule similar to package.class (for example, com.taobao.tc.refund.log.level) to ensure global uniqueness. It is recommended to indicate business meaning of the configuration in the "class" section. Use lower case for all characters. Use alphabetical letters and these four special characters (".", ":", "-", "_") only. Up to 256 characters are allowed.
	DataId string
	// Required Configuration group. To ensure uniqueness, format such as product name: module name (for example, Nacos:Test) is preferred. Use alphabetical letters and these four special characters (".", ":", "-", "_") only. Up to 128 characters are allowed.
	Group string

	//Tenant information. It corresponds to the Namespace field in Nacos.
	//Tenant      string
	NamespaceId string
	ContentType string
	AppName     string

	LineSeparator string
	KVSeparator   string
	//
	name          string
	lastCt        uint32
	ClientConfig  *constant.ClientConfig
	ServerConfigs []constant.ServerConfig
	Client        config_client.IConfigClient
}

func newNacosClientPropsConfigSource(address, group, dataId, namespaceId string) *NacosClientPropsConfigSource {
	dir := util.GetExecuteFilePath()
	s := &NacosClientPropsConfigSource{}
	name := strings.Join([]string{"Nacos", address, namespaceId, dataId}, ":")
	s.name = name
	s.DataId = dataId
	s.Group = group
	s.ClientConfig = &constant.ClientConfig{
		TimeoutMs:            10 * 1000, //请求Nacos服务端的超时时间，默认是10000ms
		BeatInterval:         5 * 1000,  //心跳间隔时间，单位毫秒（仅在ServiceClient中有效）
		LogDir:               dir + "/../data/nacos/logs",
		CacheDir:             dir + "/../data/nacos/cache",
		UpdateThreadNum:      20,    //更新服务的线程数
		NotLoadCacheAtStart:  true,  //在启动时不读取本地缓存数据，true--不读取，false--读取
		UpdateCacheWhenEmpty: false, //当服务列表为空时是否更新本地缓存，true--更新,false--不更新,当service返回的实例列表为空时，不更新缓存，用于推空保护
		//RotateTime:           "1h",            // 日志轮转周期，比如：30m, 1h, 24h, 默认是24h
		//MaxAge:               3,               // 日志最大文件数，默认3
	}
	if len(namespaceId) > 0 {
		s.ClientConfig.NamespaceId = namespaceId
		s.NamespaceId = namespaceId
	}
	s.ServerConfigs = make([]constant.ServerConfig, 0)
	addrs := strings.Split(address, ",")
	for _, addr := range addrs {
		a := strings.Split(addr, ":")
		port := 80
		if len(a) == 2 {
			var err error
			port, err = strconv.Atoi(a[1])
			if err != nil {
				log.Error("error config nacos address:", addr)
				continue
			}
		}

		s.ServerConfigs = append(s.ServerConfigs, constant.ServerConfig{
			IpAddr:      a[0],
			ContextPath: "/nacos",
			Port:        uint64(port),
		})
	}

	s.Values = make(map[string]string)

	var err error
	//s.Client, err = clients.CreateConfigClient(map[string]interface{}{
	//	constant.KEY_SERVER_CONFIGS: s.ServerConfigs,
	//	constant.KEY_CLIENT_CONFIG:  *s.ClientConfig,
	//})
	s.Client, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  s.ClientConfig,
			ServerConfigs: s.ServerConfigs,
		},
	)
	if err != nil {
		log.Panic("error create ConfigClient: ", err)
	}

	return s
}
func NewNacosClientPropsConfigSource(address, group, dataId, namespaceId string) *NacosClientPropsConfigSource {
	s := newNacosClientPropsConfigSource(address, group, dataId, namespaceId)
	s.init()

	return s
}

func NewNacosClientPropsCompositeConfigSource(address, group, tenant string, dataIds []string) *kvs.CompositeConfigSource {
	s := kvs.NewEmptyNoSystemEnvCompositeConfigSource()
	s.ConfName = "NacosKevValue"
	for _, dataId := range dataIds {
		c := NewNacosClientPropsConfigSource(address, group, dataId, tenant)
		s.Add(c)
	}

	return s
}

func (s *NacosClientPropsConfigSource) init() {
	s.listenConfig()
	s.findProperties()
}

func (s *NacosClientPropsConfigSource) listenConfig() {
	cp := vo.ConfigParam{
		DataId: s.DataId,
		Group:  s.Group,
	}
	//if len(s.AppName) > 0 {
	//	cp.AppName = s.AppName
	//}
	cp.OnChange = func(namespace, group, dataId, data string) {
		s.parseAndRegisterProps([]byte(data))
		log.Info("changed config:", namespace, group, dataId)
	}
	err := s.Client.ListenConfig(cp)
	if err != nil {
		log.Error("listen config： ", err)
	}

}

func (s *NacosClientPropsConfigSource) Close() {
}

func (s *NacosClientPropsConfigSource) findProperties() {

	data, err := s.get()
	if err != nil {
		log.Error(err)
		return
	}
	s.parseAndRegisterProps(data)

}

func (s *NacosClientPropsConfigSource) parseAndRegisterProps(data []byte) {
	sep := s.LineSeparator
	if sep == "" {
		sep = NACOS_LINE_SEPARATOR
	}
	kvsep := s.KVSeparator
	if kvsep == "" {
		kvsep = NACOS_KV_SEPARATOR
	}
	lines := bytes.Split(data, []byte(sep))

	for _, l := range lines {

		i := bytes.Index(l, []byte(kvsep))
		if i <= 0 {
			continue
		}
		key := string(l[:i])
		value := string(l[i+1:])
		s.registerProps(key, value)
		//log.Info(key,"=",value)
	}
}

func (s *NacosClientPropsConfigSource) registerProps(key, value string) {
	s.Set(strings.TrimSpace(key), strings.TrimSpace(value))

}

func (s *NacosClientPropsConfigSource) Name() string {
	return s.name
}

func (h *NacosClientPropsConfigSource) get() (body []byte, err error) {
	cp := vo.ConfigParam{
		DataId: h.DataId,
		Group:  h.Group,
	}
	content, err := h.Client.GetConfig(cp)
	if err != nil {
		log.Error("get config: ", err)
		return nil, err
	}
	return []byte(content), err
}
