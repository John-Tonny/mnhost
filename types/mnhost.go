package types

// AuthConfig contains authorization information for connecting to a Registry

type VpsInfo struct {
	InstanceId      string
	RegionName      string
	AllocationId    string
	AllocationState bool
	PublicIp        string
	PrivateIp       string
	VolumeId        string
	VolumeState     bool
}

type Volume struct {
	Mountpoint string `json:"Mountpoint"`
	Name       string `json:"Name"`
}

type ResourceInfo struct {
	Code       string
	CodeMsg    string
	MemPercert float32 //`json:"cpuPercent"`
	CpuPercert float32 //`json:"memPercent"`
	Role       string
}

type DiskInfo struct {
	Code    string
	CodeMsg string
	Total   int64
	Free    int64
}

type CoinConf struct {
	CoinName   string
	RpcPort    int
	MnKey      string
	ExternalIp string
	FileName   string
}

type NameRequest struct {
	//必须的大写开头
	Name string
}

type MountRequest struct {
	//必须的大写开头
	DeviceName string
	NodeName   string
}

type BasicResponse struct {
	Code    string
	CodeMsg string
}

type NodeIpResponse struct {
	Code    string
	CodeMsg string
	Name    string
}

type Producer struct {
	ClusterIp string
}

type Consumer struct {
	ClusterIp string
}

type NodeInfo struct {
	PublicIp  string
	PrivateIp string
	Role      string
	Status    bool
}

type NodeMap struct {
	Node map[string]*NodeInfo
}

type CoinInfo struct {
	Status bool
}

type CoinMap struct {
	Coin map[string]*CoinInfo
}

type ServiceInfo struct {
	CoinName string
	RpcPort  int
	DockerId string
}

type ServiceMap struct {
	Service map[string]*ServiceInfo
}

type RetrysInfo struct {
	Nums int
}

type RetrysMap struct {
	Retrys map[string]*RetrysInfo
}

const INIT_MANAGER_NUMS = 3
const PUBLIC_IP_ENABLED = 1
const ROLE_MANAGER = "manager"
const ROLE_WORKER = "worker"
const ROLE_NFS = "nfs"
const NODE_PREFIX = "node"
const LOGIN_USER = "root"
const SYS_MONITOR_PORT = 8844

const TOPIC_NEWVPS_SUCCESS = "Vircle.Mnhost.TOPIC.VpsNew.Success"
const TOPIC_NEWVPS_FAIL = "Vircle.Mnhost.TOPIC.VpsNew.Fail"
const TOPIC_NEWVPS_START = "Vircle.Mnhost.TOPIC.VpsNew.Start"

const TOPIC_DELVPS_SUCCESS = "Vircle.Mnhost.TOPIC.VpsDel.Success"
const TOPIC_DELVPS_FAIL = "Vircle.Mnhost.TOPIC.VpsDel.Fail"
const TOPIC_DELVPS_START = "Vircle.Mnhost.TOPIC.VpsDel.Start"

const TOPIC_NEWNODE_SUCCESS = "Vircle.Mnhost.TOPIC.NodeNew.Success"
const TOPIC_NEWNODE_FAIL = "Vircle.Mnhost.TOPIC.NodeNew.Fail"
const TOPIC_NEWNODE_START = "Vircle.Mnhost.TOPIC.NodeNew.Start"

const TOPIC_DELNODE_SUCCESS = "Vircle.Mnhost.TOPIC.NodeDel.Success"
const TOPIC_DELNODE_FAIL = "Vircle.Mnhost.TOPIC.NodeDel.Fail"
const TOPIC_DELNODE_START = "Vircle.Mnhost.TOPIC.NodeDel.Start"

const TOPIC_EXPANDVOLUME_SUCCESS = "Vircle.Mnhost.TOPIC.ExpandVolume.Success"
const TOPIC_EXPANDVOLUME_FAIL = "Vircle.Mnhost.TOPIC.ExpandVolume.Fail"
const TOPIC_EXPANDVOLUME_START = "Vircle.Mnhost.TOPIC.ExpandVolume.Start"

const TOPIC_RESTARTNODE_SUCCESS = "Vircle.Mnhost.TOPIC.RestartNode.Success"
const TOPIC_RESTARTNODE_FAIL = "Vircle.Mnhost.TOPIC.RestartNode.Fail"
const TOPIC_RESTARTNODE_START = "Vircle.Mnhost.TOPIC.RestartNode.Start"

const TOPIC_UPDATENODE_SUCCESS = "Vircle.Mnhost.TOPIC.UpdateNode.Success"
const TOPIC_UPDATENODE_FAIL = "Vircle.Mnhost.TOPIC.UpdateNode.Fail"
const TOPIC_UPDATENODE_START = "Vircle.Mnhost.TOPIC.UpdateNode.Start"

const SSH_PASSWORD = "vpub$999$000"
const RPC_USER = "vpub"
const RPC_PASSWORD = "vpub999000"
const PORT_FROM = 10000
const PORT_TO = 65530
const S_PORT = "9998"
const S_RPCPROT = "9999"
const S_WORKDIR = "vircle"
const DEVICE_NAME_FROM = "a"
const DEVICE_NAME_TO = "z"
const DEVICE_NAME_PREFIX = "xvda"

const TEST_VOLUME_SIZE = 1

const MOUNT_PATH = "/mnt/vircle"
const NFS_HOST = "172.31.43.253"
const NFS_PATH = "/mnt/efs"
const DOCKER_API_VERSION = "1.40"
const AWS_ACCOUNT = "test-account"
const SYSTEM_IMAGE = "ami-0cc887bcfcf25bccd" //"ami-0987f4b3af1ef6791" //"ami-05c4f64b6f704720f" // "ami-0815b98db1e19417a" //"ami-0b0426f6bc13cbfe4"
const ZONE_DEFAULT = "us-east-2"
const INSTANCE_TYPE_DEFAULT = "t2.small" //"t2.micro"
const VOLUME_SIZE_DEFAULT = 20
const PROVIDER_NAME = "amazon"
const GROUP_NAME = "vcl-mngroup"
const GROUP_DESC = "basic masternode group"
const KEY_PAIR_NAME = "vcl-keypair"
const DEVICE_NAME = "xvdk"
const DEVICE_NAME1 = "xvdk1"
const CORE_NUMS = 1
const MEMORY_SIZE = 1
const MASTERNODE_MAX_NUMS = 3

const MOUNT_POINT = "/var/lib/docker/volumes"

const VPS_RETRYS = 3
const NODE_RETRYS = 3
const INSTANCE_RETRYS = 30
const VOLUME_RETRYS = 10
