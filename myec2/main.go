package myec2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	log "github.com/taylormonacelli/ivytoe"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var logger log.Logger

func init() {
	logger = log.Logger{}
}

type Bar struct {
	Logger log.Logger
}

func (b *Bar) Something() {
	b.Logger.Debug("starting something")
}

func DoSomething(logger log.Logger) {
	b := Bar{Logger: logger}
	b.Something()
}

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
		logger.Error("failed to open " + filePath)
		return []Instance{}, err
	}
	defer file.Close()

	data := make(map[string][]Instance)
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		logger.Fatal(err.Error())
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

func GetInstancesState() ([]Instance, error) {
	filePath := "hosts.yaml"

	instances, err := LoadInstancesFromYAML(filePath)
	if err != nil {
		return []Instance{}, fmt.Errorf("failed to load instances from yaml: %w", err)
	}

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

	logger := log.Logger{}

	b := Bar{Logger: logger}
	b.Something()

	resp, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		s := zap.String("instance_ids", strings.Join(ris.InstanceIDs, ","))
		logger.Error("failed to describe instances", s, zap.Error(err))
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
