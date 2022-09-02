package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promDeviceLastSeen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "netscan",
			Name:        "device_last_seen_timestamp",
			Help:        "List of when each device (hostname, ip) was last seen on the network.",
			ConstLabels: map[string]string{},
		},
		[]string{"hostname", "ip"},
	)

	nmapPrefix        = "Nmap scan report for"
	ipRx              = `[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`
	rxWithHostname    = regexp.MustCompile(fmt.Sprintf(`%s (?P<Hostname>[a-zA-Z0-9_-]+) \((?P<ipAddress>%s)\)`, nmapPrefix, ipRx))
	rxWithoutHostname = regexp.MustCompile(fmt.Sprintf(`%s (?P<ipAddress>%s)`, nmapPrefix, ipRx))
)

func pollNetwork(network string) {
	out, err := exec.Command("nmap", "-sn", network).Output()
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		match := rxWithHostname.FindStringSubmatch(line)
		if len(match) > 0 {
			host := match[rxWithHostname.SubexpIndex("Hostname")]
			ip := match[rxWithHostname.SubexpIndex("ipAddress")]
			fmt.Printf("Found %s at %s\n", host, ip)
			promDeviceLastSeen.With(prometheus.Labels{"hostname": host, "ip": ip}).SetToCurrentTime()
			continue
		}
		match = rxWithoutHostname.FindStringSubmatch(line)
		if len(match) > 0 {
			hostIp := match[rxWithoutHostname.SubexpIndex("ipAddress")]
			fmt.Printf("Found %s\n", hostIp)
			promDeviceLastSeen.With(prometheus.Labels{"hostname": hostIp, "ip": hostIp}).SetToCurrentTime()
			continue
		}
	}
}

func main() {
	// Serve Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":7004", nil)

	for {
		pollNetwork("192.168.1.0/24")
		time.Sleep(5 * time.Minute) // nmap itself takes more than a minute to run.
	}
}
