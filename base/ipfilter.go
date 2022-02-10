package base

import (
	"net"
	"bytes"
	"strings"
	"github.com/spf13/viper"
	"github.com/fsnotify/fsnotify"
)

type BlackListed struct {
	ips	map[string]bool
	subnets	[]*subnet
	ranges	[]*iprange
}

type subnet struct {
	ip 	string
	ipnet	*net.IPNet
}

type iprange struct {
	start	net.IP
	end	net.IP
}

var BlackListedIPs BlackListed

func InitBlacklistedIPs(){
        BlackListedIPs.ips = make(map[string]bool)
	config := viper.New()
        config.SetConfigName("blacklisted")
        config.SetConfigType("yaml")
        config.AddConfigPath("/usr/local/production/config/")

        err := config.ReadInConfig()
        if err != nil {
                Zlog.Errorf("Falied to Initialise the Blacklisted domain data: %s", err.Error())
        }
	blacklistedIPs := config.GetString("BLACKLISTED_IP")
        UpdateBlacklistedIPs(blacklistedIPs)
        config.OnConfigChange(func(e fsnotify.Event){
                Zlog.Infof("Config file chnaged")
                blacklistedIPs = config.GetString("BLACKLISTED_IP")
                UpdateBlacklistedIPs(blacklistedIPs)
        })
        config.WatchConfig()
}

func UpdateBlacklistedIPs(blacklistedIPs string){
        ips := make(map[string]bool)
        var subnets []*subnet
        var ranges  []*iprange
        networkCheckpoints :=  strings.Split(strings.ReplaceAll(blacklistedIPs, " ", ""), ",")
        for _, checkpoint := range networkCheckpoints{
		Zlog.Infof("Processing:%s", checkpoint)
                if strings.Index(checkpoint, "-" ) != -1{
                        Zlog.Infof("Matching -", checkpoint)
                        ipRange := strings.Split(checkpoint, "-")
                        startIPnet := net.ParseIP(ipRange[0])
			endIPnet := net.ParseIP(ipRange[1])
			if endIPnet == nil || startIPnet == nil {
				Zlog.Errorf("Invalid IP Range [%s-%s]", ipRange[0], ipRange[1])
				continue
			}
                        ranges = append(ranges, &iprange{
                                start:          startIPnet,
                                end:            endIPnet,
                        })
                }else if strings.Index(checkpoint, "/" ) != -1{
                        Zlog.Infof("Matching /", checkpoint)
                        _, network, err :=  net.ParseCIDR(checkpoint)
                        if err != nil{
                                Zlog.Errorf("Invalid Subnet [%s]", checkpoint)
                                continue
                        }
                        subnets = append(subnets, &subnet{
                                ip:     checkpoint,
                                ipnet:  network,
                        })
                } else if len(checkpoint) > 0 {
                        Zlog.Infof(checkpoint)
                        ip := net.ParseIP(checkpoint)
			if ip != nil{
				ips[ip.String()] = true
			}else{
				Zlog.Errorf("Invalid IP [%s]", checkpoint)
			}
                }
        }
        BlackListedIPs.ips = ips
        BlackListedIPs.subnets = subnets
        BlackListedIPs.ranges = ranges
}

func ValidateClientIP(clientIP string) (bool){
	Zlog.Infof("Checking if the Client IP [%s] belongs to blacklisted", clientIP)
	clientIPnet := net.ParseIP(clientIP)
	if _, found := BlackListedIPs.ips[clientIPnet.String()]; found {
		Zlog.Infof("IP address [%s] belongs to blacklisted IPs", clientIP)
		return false
	}
	for _, subnet := range BlackListedIPs.subnets {
		if subnet.ipnet.Contains(clientIPnet){
			Zlog.Infof("IP address [%s] belongs to blacklisted Subnet [%s]", clientIP, subnet.ip) 
			return false
		}
	}
	for _, iprange := range BlackListedIPs.ranges{
		if bytes.Compare(clientIPnet, iprange.start) >= 0 && bytes.Compare(clientIPnet, iprange.end) <= 0{
			Zlog.Infof("IP address [%s] belongs to blacklisted IP Range [%s-%s]", clientIP, iprange.start.String(), iprange.end.String())
			return false
		}
	}
	return true
}

