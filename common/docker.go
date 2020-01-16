package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/John-Tonny/mnhost/utils"
	"github.com/docker/docker/api/types"

	//"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"

	"github.com/John-Tonny/mnhost/conf"
	"github.com/John-Tonny/mnhost/model"
	mnhostTypes "github.com/John-Tonny/mnhost/types"

	"github.com/astaxie/beego/orm"
)

type DockerClient struct {
	*client.Client
}

/*const PUBLIC_IP_ENABLED = 1
const nfs_path = "/mnt/efs"
const NODE_PREFIX = "node"*/

func DockerNewClient(publicIp, privateIp string) (*DockerClient, string, error) {
	ipAddress := privateIp
	if config.GetMyConst("publicIpEnabled") == "1" {
		//if mnhostTypes.PUBLIC_IP_ENABLED == 1 {
		ipAddress = publicIp
	}

	dockerHost := fmt.Sprintf("tcp://%s:2375", ipAddress)
	os.Setenv("DOCKER_HOST", dockerHost)
	os.Setenv("DOCKER_API_VERSION", "1.40")
	c := new(DockerClient)
	var err error
	c.Client, err = client.NewEnvClient()
	if err != nil {
		return c, utils.VPS_CONNECT_ERR, err
	}

	return c, utils.RECODE_OK, nil
}

func (c *DockerClient) SwarmInitA(ipAddress, advertiseAddr string, forceNewCluster bool) (string, error) {
	result, err := c.SwarmInit(context.Background(), swarm.InitRequest{
		ListenAddr:      fmt.Sprintf("%s:%d", "0.0.0.0", 2377),
		AdvertiseAddr:   fmt.Sprintf("%s:%d", advertiseAddr, 2377),
		ForceNewCluster: forceNewCluster,
	})
	if err != nil {
		log.Printf("result:%+v,err:%+v\n", result, err)
		return utils.SWARM_INIT_ERR, err
	}
	fmt.Printf("swarm init %s\n", result)
	return result, nil
}

func (c *DockerClient) SwarmInspectA() (string, string, error) {
	result, err := c.SwarmInspect(context.Background())
	if err != nil {
		return "", "", err
	}
	return result.JoinTokens.Manager, result.JoinTokens.Worker, nil
}

func (c *DockerClient) SwarmJoinA(advertiseAddr, remoteAddrs, token string) error {
	err := c.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:    fmt.Sprintf("%s:%d", "0.0.0.0", 2377),
		AdvertiseAddr: fmt.Sprintf("%s:%d", advertiseAddr, 2377),
		RemoteAddrs: []string{
			fmt.Sprintf("%s:%d", remoteAddrs, 2377),
		},
		JoinToken: token,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *DockerClient) ServiceCreateA(coinName string, rpcport int, dockerId, privateIp string) error {
	privateIp = strings.Replace(privateIp, ".", "-", -1)
	placement := fmt.Sprintf("node.hostname==ip-%s", privateIp)
	replicas := uint64(1)
	delay := time.Duration(10000000000)
	maxAttempts := uint64(0)
	nodeName := fmt.Sprintf("%s%d", coinName, rpcport)
	log.Printf("****create service:%s##%s\n", nodeName, placement)
	serviceSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: nodeName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image: dockerId,
				Mounts: []mount.Mount{
					{
						ReadOnly: false,
						Source:   fmt.Sprintf("%s/%s%d", mnhostTypes.MOUNT_PATH, coinName, rpcport),
						//Source:   fmt.Sprintf("%s/%s/%s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcport),
						Target: "/vircle",
						Type:   "bind",
					},
				},
				/*Labels: map[string]string{
					"role": "node1",
				},*/
			},
			LogDriver: &swarm.Driver{
				Name: "json-file",
				Options: map[string]string{
					"max-file": "3",
					"max-size": "10M",
				},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition:   "on-failure",
				Delay:       &delay,
				MaxAttempts: &maxAttempts,
			},
			Placement: &swarm.Placement{
				Constraints: []string{
					placement,
					//"node.hostname==ip-172-31-47-252",
				},
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:      "tcp",
					PublishedPort: uint32(rpcport),
					TargetPort:    uint32(rpcport),
				}, {
					Protocol:      "tcp",
					PublishedPort: uint32(rpcport + 1),
					TargetPort:    uint32(rpcport + 1),
				},
			},
		},
	}
	_, err := c.ServiceCreate(context.Background(), serviceSpec, types.ServiceCreateOptions{})
	return err
}

func (c *DockerClient) ServiceRemoveA(coinName string, rpcport int) error {
	nodeName := fmt.Sprintf("%s%d", coinName, rpcport)
	return c.ServiceRemove(context.Background(), nodeName)
}

func (c *DockerClient) ServiceUpdateA(coinName string, rpcport int, dockerId string, version swarm.Version) error {
	replicas := uint64(1)
	delay := time.Duration(10000000000)
	maxAttempts := uint64(0)
	nodeName := fmt.Sprintf("%s%d", coinName, rpcport)
	serviceSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: nodeName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image: dockerId,
				Mounts: []mount.Mount{
					{
						ReadOnly: false,
						Source:   fmt.Sprintf("%s/%s%d", mnhostTypes.MOUNT_PATH, coinName, rpcport),
						//Source:   fmt.Sprintf("%s/%s/%s%d", mnhostTypes.NFS_PATH, coinName, mnhostTypes.NODE_PREFIX, rpcport),
						Target: "/vircle",
						Type:   "bind",
					},
				},
				/*Labels: map[string]string{
					"role": "node1",
				},*/
			},
			LogDriver: &swarm.Driver{
				Name: "json-file",
				Options: map[string]string{
					"max-file": "3",
					"max-size": "10M",
				},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition:   "on-failure",
				Delay:       &delay,
				MaxAttempts: &maxAttempts,
			},
			/*Placement: &swarm.Placement{
				Constraints: []string{
					"role==node1",
				},
			},*/
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:      "tcp",
					PublishedPort: uint32(rpcport),
					TargetPort:    uint32(rpcport),
				}, {
					Protocol:      "tcp",
					PublishedPort: uint32(rpcport + 1),
					TargetPort:    uint32(rpcport + 1),
				},
			},
		},
	}
	result, err := c.ServiceUpdate(context.Background(), nodeName, version, serviceSpec, types.ServiceUpdateOptions{
		RegistryAuthFrom: types.RegistryAuthFromSpec,
	})
	log.Printf("%+v\n", result)
	return err
}

func (c *DockerClient) ServiceInspectA(serviceID string) (swarm.Service, error) {
	service, _, err := c.ServiceInspectWithRaw(context.Background(), serviceID)
	return service, err
}

func (c DockerClient) ServiceListA(options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.ServiceList(context.Background(), options)
}

func (c *DockerClient) NodeListA(options types.NodeListOptions) ([]swarm.Node, error) {
	//f := filters.NewArgs()
	//f.Add("role", typeStr)
	return c.NodeList(context.Background(), options)
}

func (c *DockerClient) NodeInspectA(nodeId string) (swarm.Node, []byte, error) {
	return c.NodeInspectWithRaw(context.Background(), nodeId)
}

func GetVpsIp(clusterName string) (string, string, error) {
	var tvpss []models.TVps
	o := orm.NewOrm()
	qs := o.QueryTable("t_vps")
	nums, err := qs.Filter("clusterName", clusterName).Filter("vps_role", mnhostTypes.ROLE_MANAGER).All(&tvpss)
	if err != nil {
		return "", "", err
	}
	if nums == 0 {
		return "", "", errors.New("no manager")
	}

	publicIp := ""
	privateIp := ""
	for _, tvps := range tvpss {
		mc, _, err := DockerNewClient(tvps.PublicIp, tvps.PrivateIp)
		if err != nil {
			continue
		}
		defer mc.Close()

		nodes, err := mc.NodeListA(types.NodeListOptions{})
		if err != nil {
			continue
		}

		for _, node := range nodes {
			if node.Status.State == "down" {
				continue
			}
			if node.Spec.Role != "manager" {
				continue
			}
			if node.ManagerStatus == nil {
				continue
			}
			if node.Status.State == "ready" && node.ManagerStatus.Reachability == "reachable" && node.ManagerStatus.Leader == true {
				privateIp = strings.Split(node.ManagerStatus.Addr, ":")[0]
				err = UpdateVpsLeader("cluster1", privateIp)
				if err != nil {
					continue
				}
				publicIp, _, _, err = GetPublicIpFromVps(privateIp)
				if err != nil {
					return "", privateIp, err
				}
				//publicIp = common. tvps.PublicIp
				log.Printf("ip:%s-%s\n", publicIp, privateIp)
				return publicIp, privateIp, nil
			}
		}
	}
	return "", "", errors.New("no manager")
}

/*func StartService(managerPublicIp, managerPrivateIp, dockerId string) error {
	//启动服务
	var mc *common.DockerClient
	mc, _, err = common.DockerNewClient(managerPublicIp, managerPrivateIp)
	if err != nil {
		return err
	}
	if mc == nil {
		return errors.New("client no connect")
	}
	defer mc.Close()
	mc.ServiceCreateA(serviceInfo.CoinName, serviceInfo.RpcPort, serviceInfo.DockerId)
}*/
