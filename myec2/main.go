package myec2

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

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

func LoadInstancesFromYAML(filePath string) ([]Instance, error) {
	err := checkFileExists(filePath)
	if err != nil {
		return []Instance{}, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("open file", "path", filePath, "error", err)
		return []Instance{}, err
	}
	defer file.Close()

	data := make(map[string][]Instance)
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		slog.Error("decoding failure", "error", err)
		return []Instance{}, err
	}

	hosts := data["hosts"]

	return hosts, nil
}

func instancesByRegion(instances []Instance) map[string][]Instance {
	instancesByRegion := make(map[string][]Instance)
	for _, i := range instances {
		instancesByRegion[i.Region] = append(instancesByRegion[i.Region], i)
	}
	return instancesByRegion
}

func GetInstancesState(instances []Instance) ([]Instance, error) {
	var wg sync.WaitGroup
	var instanceQueryResults []Instance
	instChannel := make(chan Instance)

	instancesByRegion := instancesByRegion(instances)
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
		slog.Error("failed to describe instances", "instance_ids", ris.InstanceIDs, "error", err)
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
		var ri RegionInstances
		instances := instancesByRegion[region]
		for _, instance := range instances {
			hostIds = append(hostIds, instance.InstanceID)
		}

		sort.Slice(hostIds, func(i, j int) bool {
			return hostIds[i] < hostIds[j]
		})

		ri.InstanceList = instancesByRegion[region]
		ri.InstanceIDs = hostIds
		ri.Region = region
		containers = append(containers, ri)
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

func checkFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist")
		}
		return err
	}
	return nil
}
