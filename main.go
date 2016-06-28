package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"reflect"
	"sort"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	cfg = struct {
		AutoscalingGroup string        `flag:"autoscaling-group-name,a" env:"AUTOSCALING_GROUP_NAME" default:"" description:"Name of the AutoScalingGroup to fetch instances from"`
		Port             int64         `flag:"port" env:"PORT" default:"" description:"Port to register in SRV record"`
		PollInterval     time.Duration `flag:"interval,i" env:"INTERVAL" default:"30s" description:"Interval to poll for changes in the ASG"`
		PublishTo        string        `flag:"publish-to,p" env:"PUBLISH_TO" default:"file://discover.json" description:"Where to write the discovery file"`
		VersionAndExit   bool          `flag:"version" description:"Print version and exit"`
	}{}
	version = "dev"

	publishProviders = map[string]publishProvider{}

	ipCache []string
)

type publishProvider interface {
	Write(*url.URL, io.Reader) error
}

type staticConfig struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels,omitempty"`
}

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline flags: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("autoscaling-dns-sd %s\n", version)
		os.Exit(0)
	}

	if cfg.AutoscalingGroup == "" || cfg.Port == 0 || cfg.PublishTo == "" {
		rconfig.Usage()
		os.Exit(1)
	}
}

func main() {
	for range time.Tick(cfg.PollInterval) {
		if err := doUpdate(); err != nil {
			log.Printf("Failed to update discovery file for ASG %s: %s", cfg.AutoscalingGroup, err)
		}
	}
}

func doUpdate() error {
	instanceIDs, err := getInstanceIDsFromASG()
	if err != nil {
		return fmt.Errorf("Unable to retrieve InstanceIDs from ASG: %s", err)
	}
	if len(instanceIDs) == 0 {
		return fmt.Errorf("Got 0 InstanceIDs from ASG. Are you sure you have the right ASG?")
	}

	ips, err := getIPsFromInstanceIDs(instanceIDs)
	if err != nil {
		return fmt.Errorf("Unable to retrieve InstanceIDs from ASG: %s", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("Got 0 IPs from instances.")
	}

	sort.Strings(ips)
	if reflect.DeepEqual(ipCache, ips) {
		// We got the same IP set as last time, no need to change the DNS record
		log.Printf("No news in ASG, not touching DNS record")
		return nil
	}

	ipCache = ips

	targets := []string{}
	for _, ip := range ips {
		targets = append(targets, fmt.Sprintf("%s:%d", ip, cfg.Port))
	}

	if err := writeDiscoveryFile(cfg.PublishTo, targets); err != nil {
		return fmt.Errorf("Unable to publish IPs: %s", err)
	}

	log.Printf("Finished successfully, published %d IPs", len(ips))
	return nil
}

func writeDiscoveryFile(publishTo string, targets []string) error {
	targetURL, err := url.Parse(cfg.PublishTo)
	if err != nil {
		return err
	}

	provider, ok := publishProviders[targetURL.Scheme]
	if !ok {
		return fmt.Errorf("No handler found for scheme %s", targetURL.Scheme)
	}

	data := []staticConfig{
		{Targets: targets},
	}

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return err
	}

	return provider.Write(targetURL, buf)
}

func getIPsFromInstanceIDs(instanceIDs []*string) (ips []string, err error) {
	ec2Svc := ec2.New(session.New())

	var resp *ec2.DescribeInstancesOutput
	resp, err = ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})

	if err != nil {
		return
	}

	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			if i.PrivateIpAddress == nil {
				log.Printf("Machine %s has nil-ip, probably its not ready yet or its terminating.", *i.InstanceId)
				continue
			}
			ips = append(ips, *i.PrivateIpAddress)
		}
	}

	return
}

func getInstanceIDsFromASG() (instanceIDs []*string, err error) {
	asSvc := autoscaling.New(session.New())

	var resp *autoscaling.DescribeAutoScalingGroupsOutput
	resp, err = asSvc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(cfg.AutoscalingGroup)},
	})

	if err != nil || len(resp.AutoScalingGroups) != 1 {
		return
	}

	for _, i := range resp.AutoScalingGroups[0].Instances {
		instanceIDs = append(instanceIDs, i.InstanceId)
	}

	return
}
