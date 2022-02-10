package base

import (
	"strings"
	"net"
	"encoding/json"
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
	start	string
	end	string
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
                Zlog.Infof(checkpoint)
                if strings.Index(checkpoint, "-" ) != -1{
                        Zlog.Infof("Matching -", checkpoint)
                        ipRange := strings.Split(checkpoint, "-")
                        if len(ipRange) < 2 || len(ipRange) > 2 {
                                continue
                        }
                        ranges = append(ranges, &iprange{
                                start:          ipRange[0],
                                end:            ipRange[1],
                        })
                }else if strings.Index(checkpoint, "/" ) != -1{
                        Zlog.Infof("Matching /", checkpoint)
                        _, network, err :=  net.ParseCIDR(checkpoint)
                        if err != nil{
                                Zlog.Infof("Invalid subnet")
                                continue
                        }
                        subnets = append(subnets, &subnet{
                                ip:     checkpoint,
                                ipnet:  network,
                        })
                } else if len(checkpoint) > 0 {
                        Zlog.Infof(checkpoint)
                        ip := net.ParseIP(checkpoint)
                        ips[ip.String()] = true
                }
        }
        BlackListedIPs.ips = ips
        BlackListedIPs.subnets = subnets
        BlackListedIPs.ranges = ranges
}

func ValidateClientIP(clientIP string) (bool){
	base.Zlog.Infof("Checking if the Client IP [%s] belongs to blacklisted", clientIP)
	clientIPnet := net.ParseIP(clientIP)
	if _, found := BlackListedIPs.ips[clientIPnet.String()]; found {
		Zlog.Infof("IP address [%s] belongs to blacklisted IPs", clientIP)
		return false
	}
	for _, subnet := range BlackListedIPs.subnets {
		if subnet.ipnet.Contain(clientIPnet){
			Zlog.Infof("IP address [%s] belongs to Subnet [%s]", clientIP, subnet.ip) 
			return false
		}
	}
	return true
}

