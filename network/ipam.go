package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const ipamDefaultAllocatorPath = "/var/lib/tiny-docker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string             // 分配文件存放位置
	Subnets             *map[string]string // 网段和位图算法的数组 map, key 是网段， value 是分配的位图数组
}

// 初始化一个IPAM的对象，默认使用/var/lib/tiny-docker/network/ipam/subnet.json作为分配信息存储位置
var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// Allocate 在网段中分配一个可用的 IP 地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 初始化数组
	ipam.Subnets = &map[string]string{}

	// 加载信息
	err = ipam.load()
	if err != nil {
		return nil, errors.Wrap(err, "load subnet allocation info error")
	}
	// net.IPNet.Mask.Size函数会返回网段的子网掩码的总长度和网段前面的固定位的长度
	// 比如“127.0.0.0/8”网段的子网掩码是“255.0.0.0”
	// 那么subnet.Mask.Size()的返回值就是前面255所对应的位数和总位数，即8和24
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, size := subnet.Mask.Size()
	// 如果之前没有分配过这个网段，则初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// ／用“0”填满这个网段的配置，uint8(size - one)表示这个网段中有多少个可用地址
		// size - one是子网掩码后面的网络位数，2^(size - one)表示网段中的可用IP数
		// 而2^(size - one)等价于1 << (size - one) - 广播地址 - 网络地址
		// size <= 32
		ipCount := 1 << (size - one)
		// 减去 0和 255这俩个不可分配地址，所有可分配的 IP 地址
		ipalloc := strings.Repeat("0", ipCount-2)
		// 初始化分配配置，标记 .0 和 .255 位置为不可分配，直接置为 1
		(*ipam.Subnets)[subnet.String()] = fmt.Sprintf("1%s1", ipalloc)
	}
	// 遍历网段的位图数组
	for c := range (*ipam.Subnets)[subnet.String()] {
		// 找到数组中为“0”的项和数组序号，即可以分配的 IP
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// 设置这个为“0”的序号值为“1” 即标记这个IP已经分配过了
			// Go 的字符串，创建之后就不能修改 所以通过转换成 byte 数组，修改后再转换成字符串赋值
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			// 这里的 subnet.IP只是初始IP，比如对于网段192 168.0.0/16 ，这里就是192.168.0.0
			ip = subnet.IP
			/*
				还需要通过网段的IP与上面的偏移相加计算出分配的IP地址，由于IP地址是uint的一个数组，
				需要通过数组中的每一项加所需要的值，比如网段是172.16.0.0/12，数组序号是65555,
				那么在[172,16,0,0] 上依次加[uint8(65555 >> 24)、uint8(65555 >> 16)、
				uint8(65555 >> 8)、uint8(65555 >> 0)]， 即[0, 1, 0, 19]， 那么获得的IP就
				是172.17.0.19.
			*/
			// 转换成256进制再加

			// 由于此处IP是从1开始分配的（0被网关占了），所以最后再加1，最终得到分配的IP 172.17.0.20
			for t := uint(4); t > 0; t-- {
				// uint8 = 只看c二进制后面8位
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			break
		}
	}
	// 最后调用dump将分配结果保存到文件中
	err = ipam.dump()
	if err != nil {
		logrus.Error("Allocate: dump ipam error", err)
	}
	return
}

// Release 回收分配的ip地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		return errors.Wrap(err, "load subnet allocation info error")
	}

	c := 0
	releaseIP := ipaddr.To4()
	for t := uint(4); t > 0; t-- {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	// 对应位置0
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	// 最后调用dump将分配结果保存到文件中
	err = ipam.dump()
	if err != nil {
		logrus.Error("Allocate: dump ipam error", err)
	}
	return nil
}

// load 加载网段地址分配信息
func (ipam *IPAM) load() error {
	// 检查存储文件状态，如果不存在，则说明之前没有分配，则不需要加载
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()

	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	return errors.Wrap(err, "err dump allocation info")
}

// dump 存储网段信息
func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	// 判断是否存在文件夹，如果不存在则创建
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(ipamConfigFileDir, constant.Perm0644); err != nil {
			return err
		}
	}
	// 打开存储文件 O_TRUNC 表示如果存在则消空， os O_CREATE 表示如果不存在则创建
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, constant.Perm0644)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	_, err = subnetConfigFile.Write(ipamConfigJson)
	return err
}
