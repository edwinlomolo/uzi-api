package ipinfo

import "github.com/ipinfo/go/v2/ipinfo"

type IpInfo interface {
	GetIpinfo(ip string) (*ipinfo.Core, error)
}
