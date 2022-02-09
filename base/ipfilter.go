package base

import (
	"strings"
	"net"
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
	ipnet	string
}

type iprange struct {
	start	string
	end	string
}

var BlackListedIPs BlackListed

func InitBlacklistedIPs(){
        BlackListedIPs.ips = make(map[string]bool)
        viper.SetConfigName("blacklisted")
        viper.SetConfigType("yaml")
        viper.AddConfigPath("/usr/local/production/config/")

        err := viper.ReadInConfig()
        if err != nil {
                Zlog.Errorf("Falied to Initialise the Blacklisted domain data: %s", err.Error())
        }
        blacklistedIPs := viper.Get("BLACKLISTED_IP").(string)
        UpdateBlacklistedIPs(blacklistedIPs)
        viper.OnConfigChange(func(e fsnotify.Event){
                Zlog.Infof("Config file chnaged")
                blacklistedIPs := viper.Get("BLACKLISTED_IP").(string)
                UpdateBlacklistedIPs(blacklistedIPs)
        })
        viper.WatchConfig()
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
                } else {
                        Zlog.Infof(checkpoint)
                        ip := net.ParseIP(checkpoint)
                        ips[ip.String()] = true
                }
        }
        BlackListedIPs.ips = ips
        BlackListedIPs.subnets = subnets
        BlackListedIPs.ranges = ranges
}
