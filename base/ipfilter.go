package base

import (
	"net"
	"bytes"
	"strings"
	"net/http"
	"github.com/spf13/viper"
	"github.com/fsnotify/fsnotify"
)

type Prohibited struct {
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

var ProhibitedIPs Prohibited 
var ProhibitedDomains string

func InitProhibitedIPs(){
        ProhibitedIPs.ips = make(map[string]bool)
	config := viper.New()
        config.SetConfigName("prohibited")
        config.SetConfigType("yaml")
        config.AddConfigPath("/usr/local/production/config/")

        err := config.ReadInConfig()
        if err != nil {
                Zlog.Errorf("Falied to Initialise the Prohibited domain data: %s", err.Error())
        }
	blockedIPs := config.GetString("BANNED_IP")
	ProhibitedDomains = config.GetString("BANNED_DOMAINS")
        UpdateProhibitedIPs(blockedIPs)
        config.OnConfigChange(func(e fsnotify.Event){
                Zlog.Infof("Config file chnaged")
                blockedIPs = config.GetString("BANNED_IP")
		ProhibitedDomains = config.GetString("BANNED_DOMAINS")
                UpdateProhibitedIPs(blockedIPs)
        })
        config.WatchConfig()
}

func UpdateProhibitedIPs(blockedIPs string){
        ips := make(map[string]bool)
        var subnets []*subnet
        var ranges  []*iprange
        networkCheckpoints :=  strings.Split(strings.ReplaceAll(blockedIPs, " ", ""), ",")
        for _, checkpoint := range networkCheckpoints{
		Zlog.Infof("Processing:%s", checkpoint)
                if strings.Index(checkpoint, "-" ) != -1{
			Zlog.Infof("Matching: %s", checkpoint)
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
        ProhibitedIPs.ips = ips
        ProhibitedIPs.subnets = subnets
        ProhibitedIPs.ranges = ranges
}

func ValidateClientIP(req *http.Request) (bool){
	clientIP := GetClientIP(req)
	Zlog.Infof("Checking if the Client IP [%s] belongs to blocked IP list", clientIP)
	clientIPnet := net.ParseIP(clientIP)
	if _, found := ProhibitedIPs.ips[clientIPnet.String()]; found {
		Zlog.Errorf("IP address [%s] belongs to blocked IPs", clientIP)
		return false
	}
	for _, subnet := range ProhibitedIPs.subnets {
		if subnet.ipnet.Contains(clientIPnet){
			Zlog.Errorf("IP address [%s] belongs to blocked Subnet [%s]", clientIP, subnet.ip) 
			return false
		}
	}
	for _, iprange := range ProhibitedIPs.ranges{
		if bytes.Compare(clientIPnet, iprange.start) >= 0 && bytes.Compare(clientIPnet, iprange.end) <= 0{
			Zlog.Errorf("IP address [%s] belongs to blocked IP Range [%s-%s]", clientIP, iprange.start.String(), iprange.end.String())
			return false
		}
	}
	return true
}

func validateDomain(userEmail string) bool{
	if len(ProhibitedDomains) == 0{
		Zlog.Infof("Blacklisted domains are not defined")
		return true
	}
	blockedDomains := strings.ReplaceAll(ProhibitedDomains, " ", "")
        blockedDomains = strings.Split(blockedDomains, ",")
        at := strings.LastIndex(userEmail, "@")
        userDomain := userEmail[at+1:]
        base.Zlog.Infof("Verifying if email Domain[%s] belongs to the Banned domains.", userDomain)
        for _, bdomain:= range blockedDomains{
		domain := strings.ReplaceAll(bdomain, ".", `\.`)
                domain = strings.ReplaceAll(domain, "*", ".*")
		base.Zlog.Debugf("Domains:[%s]:[%s]",bdomain, domain)
                emailRegex := regexp.MustCompile(domain)
                match := emailRegex.FindString(userDomain)
                if match != ""{
                        Zlog.Errorf("Banned domain found:%s", bdomain)
			Zlog.Errorf("User email Domain [%s] belongs to the Banned domains [%s]", userDomain, bdomain)
                        return false
                }
        }
        Zlog.Infof("User email Domain [%s] is safe to procced", userDomain)
        return true
}

func GetClientIP(r *http.Request)(string){
        ip := r.Header.Get("X-REAL-IP")
        netip := net.ParseIP(ip)
        if netip != nil {
                return ip
        }

        xfips := r.Header.Get("X-FORWARDED-FOR")
        ips := strings.Split(xfips, ",")
        for _, ip := range ips {
                netip = net.ParseIP(ip)
                if netip != nil {
                        return ip
                }
        }
        // Get the IP from remote Addr
        ip, _,  err := net.SplitHostPort(r.RemoteAddr)
        if err != nil{
                return ""
        }
        netip = net.ParseIP(ip)
        if netip != nil{
                return ip
        }
        return ""
}
