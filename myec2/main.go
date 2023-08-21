package myec2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taylormonacelli/candleburn/logging"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type RegionInstances struct {
	InstanceList []Instance
	InstanceIDs  []string
	Region       string `yaml:"region"`
}

type Instance struct {
	InstanceID string `yaml:"instance_id"`
	Name       string
	Region     string `yaml:"region"`
	State      string
	Type       string
}

func LoadInstancesFromYAML() ([]Instance, error) {
	file, err := os.Open("hosts.yaml")
	if err != nil {
		logging.Logger.Error("failed to open hosts.yaml")
		return []Instance{}, err
	}
	defer file.Close()

	var data map[string][]Instance
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		logging.Logger.Fatal(err.Error())
	}

	hosts := data["hosts"]

	return hosts, nil
}

func InstancesByRegion(instances []Instance) map[string][]Instance {
	instancesByRegion := make(map[string][]Instance)
	for _, i := range instances {
		instancesByRegion[i.Region] = append(instancesByRegion[i.Region], i)
	}
	return instancesByRegion
}

func GetInstancesState() ([]Instance, error) {
	instances, err := LoadInstancesFromYAML()
	if err != nil {
		return []Instance{}, fmt.Errorf("failed to load instances from yaml: %w", err)
	}

	var wg sync.WaitGroup
	var instanceQueryResults []Instance
	instChannel := make(chan Instance)

	instancesByRegion := InstancesByRegion(instances)
	instSlice := generateInstanceSlice(instancesByRegion)

	for _, inst := range instSlice {
		wg.Add(1)
		go CheckRegionInstanceState(inst, instChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(instChannel)
	}()

	for inst := range instChannel {
		instanceQueryResults = append(instanceQueryResults, inst)
	}

	return instanceQueryResults, nil
}

func CheckRegionInstanceState(ris RegionInstances, regionInstancesChannel chan Instance, wg *sync.WaitGroup) {
	defer wg.Done()

	input := &ec2.DescribeInstancesInput{InstanceIds: ris.InstanceIDs}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(ris.Region))
	if err != nil {
		panic(err)
	}
	client := ec2.NewFromConfig(cfg)

	resp, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		s := zap.String("instance_ids", strings.Join(ris.InstanceIDs, ","))
		logging.Logger.Error("failed to describe instances", s, zap.Error(err))
		return
	}

	// Extract and print the desired information
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceName := ""
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instanceName = *tag.Value
					break
				}
			}
			i := Instance{
				Name:       instanceName,
				InstanceID: *instance.InstanceId,
				Region:     ris.Region,
				State:      string(instance.State.Name),
				Type:       string(instance.InstanceType),
			}
			regionInstancesChannel <- i
		}
	}
}

func generateInstanceSlice(instancesByRegion map[string][]Instance) []RegionInstances {
	var containers []RegionInstances

	for region := range instancesByRegion {
		var hostIds []string
		var container RegionInstances
		instances := instancesByRegion[region]
		for _, instance := range instances {
			hostIds = append(hostIds, instance.InstanceID)
		}
		container.InstanceList = instancesByRegion[region]
		container.InstanceIDs = hostIds
		container.Region = region
		containers = append(containers, container)
	}

	return containers
}

func ExportInstancesQuery(query []Instance, outfile string) error {
	var writer *os.File
	if outfile == "-" {
		writer = os.Stdout
	} else {
		file, err := os.Create(outfile)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	jsonData, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal LaunchTemplateData to JSON: %w", err)
	}

	_, err = writer.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write request response to log file: %w", err)
	}

	return nil
}
