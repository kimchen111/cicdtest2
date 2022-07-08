package common

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/vishvananda/netlink"
)

func VpeAsNO(vni int) int {
	return 50000 + vni
}

func HubAsNO(vni int) int {
	return 60000 + vni
}

func CpeAsNO(vni int) int {
	return 70000 + vni
}

// MergeArray 合并数组
func MergeArray(dest []string, src []string) (result []string) {
	result = make([]string, len(dest)+len(src))
	//将第一个数组传入result
	copy(result, dest)
	//将第二个数组接在尾部，也就是 len(dest):
	copy(result[len(dest):], src)
	return
}

func LoadConfig(path string, pointer any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buff := make([]byte, 4096)
	if n, err := f.Read(buff[:]); err == nil {
		// log.Printf("len: %d", n)
		if err2 := json.Unmarshal(buff[:n], pointer); err2 == nil { //反序列化
			return nil
		} else {
			return err2
		}
	} else {
		return err
	}
}

func Preparebgp() {
	basedir := "/sdwan"
	bgpdir := fmt.Sprintf("%s/%s", basedir, "bgp")
	if !FileExists(bgpdir) {
		exec.Command("mkdir", "-p", bgpdir).Run()
	}
	birdPath := fmt.Sprintf("%s/%s", basedir, "bird")
	content := `
#!/bin/sh
if [ $# != 2 ]; then
  echo "Usage: bird [start|stop] ns50"
  exit
fi

action=$1
name=$2

[ ! -d /var/run/bird ] && {
	mkdir /var/run/bird
}

start() {
  ip netns exec ${name} /usr/sbin/bird -c /sdwan/bgp/${name}.conf -s /run/${name}.ctl -P /var/run/bird/${name}.pid
  exit 0
}

stop() {
  pid=$(cat /var/run/bird/${name}.pid)
  kill ${pid}
}

[ "${action}" == "start" ] && {
  start
}

[ "${action}" == "stop" ] && {
  stop
}
`
	WriteFile(birdPath, content)
	exec.Command("chmod", "+x", birdPath).Run()
	bircPath := fmt.Sprintf("%s/%s", basedir, "birc")
	if !FileExists(bircPath) {
		content := `
#!/bin/sh
if [ $# == 1 ]; then
  name=$1
  ip netns exec ${name} /usr/sbin/birdc -s /run/${name}.ctl
else
  echo "Usage: birdc ns50"
fi
`
		WriteFile(bircPath, content)
		exec.Command("chmod", "+x", bircPath).Run()
	}
}

func LoadDirlinkIntf(name string) (MstpEndpointVO, error) {
	path := fmt.Sprintf("%s/%s", VpeDirlinkConfDir, name)
	dle := MstpEndpointVO{}
	err := LoadConfig(path, &dle)
	return dle, err
}

func LoadVxlanTunnel(name string) (VxlanTunnelVO, error) {
	path := fmt.Sprintf("%s/%s", VpeVxlanConfDir, name)
	vtvo := VxlanTunnelVO{}
	err := LoadConfig(path, &vtvo)
	return vtvo, err
}

func WriteConfig(path string, pointer any) bool {
	barr, err := json.Marshal(pointer)
	if err == nil {
		content := string(barr)
		return WriteFile(path, content)
	}
	return false
}

func EnsureBridge(brname string) (netlink.Link, error) {
	br, err := netlink.LinkByName(brname)
	if err != nil {
		la := netlink.NewLinkAttrs()
		la.Name = brname
		err2 := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: la})
		if err2 != nil {
			log.Printf("add bridge failed %v ", err2)
			return nil, err2
		}
		br, _ = netlink.LinkByName(brname)
		netlink.LinkSetUp(br)
		exec.Command("brctl", "stp", brname, "on").Run()
		return br, nil
	} else {
		return br, nil
	}
}

func AddIpaddr(link netlink.Link, ipaddr string) bool {
	addr, err := netlink.ParseAddr(ipaddr)
	if err != nil {
		log.Printf("parese addr %s failed.", ipaddr)
		return false
	}
	err1 := netlink.AddrAdd(link, addr)
	if err1 != nil {
		log.Printf("set ipaddr failed %s ", ipaddr)
		return false
	}
	return true
}

func LoadBody(body any, pointer any) error {
	if arr, err := json.Marshal(body); err == nil { //序列化
		if err2 := json.Unmarshal(arr, pointer); err2 == nil { //反序列化
			return nil
		} else {
			return err2
		}
	} else {
		return err
	}
}

// 如 24 对应的子网掩码地址为 255.255.255.0
func LenToNetMask(subnet int) string {
	//TODO CHECK
	buff := bytes.Buffer{}
	for i := 0; i < subnet; i++ {
		buff.WriteString("1")
	}
	for i := subnet; i < 32; i++ {
		buff.WriteString("0")
	}
	masker := buff.String()
	a, _ := strconv.ParseUint(masker[:8], 2, 64)
	b, _ := strconv.ParseUint(masker[8:16], 2, 64)
	c, _ := strconv.ParseUint(masker[16:24], 2, 64)
	d, _ := strconv.ParseUint(masker[24:32], 2, 64)
	resultMask := fmt.Sprintf("%v.%v.%v.%v", a, b, c, d)
	return resultMask
}

func Cidr2AddrMask(cidr string) (ipaddr string, netmask string) {
	arr := strings.Split(cidr, "/")
	netmask = "255.255.255.0"
	if len(arr) >= 2 {
		ipaddr = arr[0]
		if masklen, err := strconv.Atoi(arr[1]); err == nil {
			netmask = LenToNetMask(masklen)
		}
	} else {
		return "", ""
	}
	return
}

func GetLoopBackAddr() string {
	if loAddr := GetInterfaceAddr("dummy0"); loAddr != "" {
		log.Printf("Using address of dummy0 as RouterID: %s", loAddr)
		return loAddr
	} else if eth0Addr := GetInterfaceAddr("eth0"); eth0Addr != "" {
		log.Printf("Using address of eth0 as RouterID: %s", eth0Addr)
		return eth0Addr
	} else {
		log.Panic("Cound't find router id addr, please check dummy0 interface.")
		return ""
	}
}

func GetInterfaceAddr(name string) string {
	nl, _ := netlink.LinkByName(name)
	addrs, err := netlink.AddrList(nl, netlink.FAMILY_V4)
	if err == nil {
		if len(addrs) == 1 {
			addr := addrs[0]
			str := addr.IPNet.String()
			arr := strings.Split(str, "/")
			return arr[0]
		}
	}
	return ""
}

func GetIntfMacaddr(name string) string {
	nl, err := netlink.LinkByName(name)
	if err == nil {
		return nl.Attrs().HardwareAddr.String()
	}
	return ""
}

func IntfExists(name string) bool {
	_, err := netlink.LinkByName(name)
	return err == nil
}

func Netmask(cidr string) int {
	_, netx, _ := net.ParseCIDR(cidr)
	x, _ := netx.Mask.Size()
	return x
}

func Esn2RID(esn string) string {
	xa, err := hex.DecodeString(esn)
	if err != nil {
		return "1.1.1.1"
	}
	ip := net.IPv4(xa[2], xa[5], xa[1], xa[3])
	return ip.String()
}

//只截取IP地址部分
func PureIp(cidr string) string {
	ip, _, _ := net.ParseCIDR(cidr)
	return ip.String()
}

//返回255.255.255.0的形式
func NetmaskStr(cidr string) string {
	_, netx, _ := net.ParseCIDR(cidr)
	a, _ := netx.Mask.Size()
	return LenToNetMask(a)
}

//计算地址的网络描述，192.168.1.1/24 -> 192.168.1.0/24
func Network(addr string) string {
	_, netx, _ := net.ParseCIDR(addr)
	return netx.String()
}

//取一个网段的第一个IP
func FirstIp(cidr string) string {
	_, netx, _ := net.ParseCIDR(cidr)
	prefix, _ := netip.ParsePrefix(netx.String())
	return prefix.Addr().Next().String()
}

//取一个网段的第二个IP
func SecondIp(cidr string) string {
	_, netx, _ := net.ParseCIDR(cidr)
	prefix, _ := netip.ParsePrefix(netx.String())
	return prefix.Addr().Next().Next().String()
}

func RecalcCidr(cidr string, mask int) string {
	ip, _, _ := net.ParseCIDR(cidr)
	temp := fmt.Sprintf("%s/%d", ip.String(), mask)
	_, netx, _ := net.ParseCIDR(temp)
	return netx.String()
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func ReadFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	buff := make([]byte, 4096)
	if n, err := f.Read(buff[:]); err == nil {
		return string(buff[:n]), nil
	} else {
		return "", err
	}
}

func WriteFile(path string, content string) bool {
	//打开文件,新建文件
	f, err := os.Create(path)
	if err != nil {
		log.Println("err = ", err)
		return false
	}
	//使用完毕,需要关闭文件
	defer f.Close()
	if _, err = f.WriteString(content); err != nil {
		return false
	}
	return true
}

func AppendFile(path string, content string) bool {
	//打开文件
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		log.Println("err = ", err)
		return false
	}
	//使用完毕,需要关闭文件
	defer f.Close()
	if _, err = f.WriteString(content); err != nil {
		return false
	}
	return true
}

func GenWgPrivKey() string {
	command := exec.Command("wg", "genkey")
	out, _ := command.CombinedOutput()
	rest := string(out)
	return strings.Replace(rest, "\n", "", -1)
}

func GenWgPubKey(privKey string) string {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = bytes.NewBuffer([]byte(privKey))
	out, _ := cmd.CombinedOutput()
	rest := string(out)
	return strings.Replace(rest, "\n", "", -1)
}

//把mapy中的内容合并到mapx中
func MapConcat(mapx map[string]string, mapy map[string]string) {
	for k, v := range mapy {
		if _, ok := mapx[k]; !ok {
			mapx[k] = v
		}
	}
}

func CheckTools() (string, bool) {
	if !FileExists("/usr/bin/wg") {
		return "/usr/bin/wg", false
	}
	return "", true
}

func Netns() []string {
	command := exec.Command("ip", "netns")
	out, _ := command.CombinedOutput()
	output := string(out)
	lines := strings.Split(output, "\n")
	len := len(lines)
	rest := make([]string, len)
	for i, line := range lines {
		arr := strings.Split(line, " ")
		rest[i] = arr[0]
	}
	return rest[:len]
}

func NetnsExists(name string) bool {
	list := Netns()
	for _, ns := range list {
		if name == ns {
			return true
		}
	}
	return false
}

func HttpPostWithRetry(url string, data []byte, retry int) {
	reader := bytes.NewReader(data)
	log.Printf("Request URL: %s", url)
	log.Printf("Body: %s", string(data))

	for i := 0; i < retry; i++ {
		resp, err := http.Post(url, "application/json", reader)
		if err != nil {
			log.Printf("Error: %s", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				//Request success
				return
			}
			log.Printf("Resp Status Code: %d", resp.StatusCode)
			// body, _ := ioutil.ReadAll(resp.Body)
			// log.Printf("Resp Body: %s", body)
		}
		s := 5 * i
		log.Printf("Retry after %d seconds", s)
		tchan := time.After(time.Second * time.Duration(s))
		<-tchan
	}
	log.Printf("Failed push data after 3 times retry.")
}

//默认重试3次
func HttpPost(url string, data []byte, disable bool) {
	if disable {
		return
	}
	HttpPostWithRetry(url, data, 3)
}

func BytesToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}
func CompressStr(str string) string {
	reg, _ := regexp.Compile(`\s+`)
	return reg.ReplaceAllString(str, " ")
}

func GenRandBuff() []byte {
	buff := make([]byte, 64)
	rand.Seed(time.Now().Unix())
	for i := 4; i < 64; i++ {
		r := rand.Intn(94)
		buff[i] = byte(r + 32)
	}
	return buff[:]
}

func CreateVlanIntf(intf string, vlanid int, name string) bool {
	_, err := netlink.LinkByName(name)
	if err == nil {
		return true
	} else {
		link, err := netlink.LinkByName(intf)
		if err != nil {
			log.Printf("Can not found interface %s", intf)
			return false
		}
		la := netlink.LinkAttrs{
			Name:        name,
			ParentIndex: link.Attrs().Index,
			TxQLen:      -1,
		}
		err1 := netlink.LinkAdd(&netlink.Vlan{
			LinkAttrs: la,
			VlanId:    vlanid,
		})
		if err1 != nil {
			log.Printf("Create failed.")
			return false
		}
		return true
	}
}

func ListFiles(path string) (bool, []string) {
	arr := make([]string, 500)
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("ReadDir failed, error: %s\n", err)
		return false, nil
	}
	i := 0
	for _, info := range fileInfos {
		if !info.IsDir() {
			arr[i] = fmt.Sprintf("%s/%s", path, info.Name())
			i++
		}
	}
	return true, arr[:i]
}
