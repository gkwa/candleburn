package myec2

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func compareInstances(instances1, instances2 []Instance) bool {
	if len(instances1) != len(instances2) {
		return false
	}

	for i := range instances1 {
		if instances1[i] != instances2[i] {
			return false
		}
	}

	return true
}

func CompareContainerSlices(slice1, slice2 []RegionInstances) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := 0; i < len(slice1); i++ {
		if !reflect.DeepEqual(slice1[i], slice2[i]) {
			return false
		}
	}

	return true
}

func CompareInstanceSlices(slice1, slice2 []Instance) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := 0; i < len(slice1); i++ {
		if !reflect.DeepEqual(slice1[i], slice2[i]) {
			return false
		}
	}

	return true
}

func TestLoad(t *testing.T) {
	expected := `
	[
		{
		  "InstanceID": "i-west1aabbbccc",
		  "Name": "myhost1",
		  "Region": "us-west-1",
		  "Running": false
		},
		{
		  "InstanceID": "i-east1eeefffggg",
		  "Name": "myhost2",
		  "Region": "us-east-1",
		  "Running": false
		},
		{
		  "InstanceID": "i-east1hhhiiijjj",
		  "Name": "myhost3",
		  "Region": "us-east-1",
		  "Running": false
		}
	  ]
	  `

	got, err := LoadInstancesFromYAML()
	if err != nil {
		panic(err)
	}

	// Unmarshal the JSON string into a struct
	var instanceList []Instance
	err = json.Unmarshal([]byte(expected), &instanceList)
	if err != nil {
		t.Error(err)
	}

	want := instanceList

	equal := CompareInstanceSlices(want, got)

	if !equal {
		t.Errorf("Expected %v but got %v", want, got)
	}
}

func TestMyFunction(t *testing.T) {
	hosts, err := LoadInstancesFromYAML()
	if err != nil {
		panic(err)
	}
	want := 3
	got := len(hosts)

	if got != want {
		t.Errorf("Expected %d hosts, but got %d", want, got)
	}
}

func TestInstancesByRegion(t *testing.T) {
	expected := `
	{
		"us-east-1": [
		  {
			"InstanceID": "i-east1eeefffggg",
			"Name": "myhost2",
			"Region": "us-east-1",
			"Running": false
		  },
		  {
			"InstanceID": "i-east1hhhiiijjj",
			"Name": "myhost3",
			"Region": "us-east-1",
			"Running": false
		  }
		],
		"us-west-1": [
		  {
			"InstanceID": "i-west1aabbbccc",
			"Name": "myhost1",
			"Region": "us-west-1",
			"Running": false
		  }
		]
	  }
	`

	var want map[string][]Instance
	err := json.Unmarshal([]byte(expected), &want)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	instanceListAsString := `
	[
		{
		  "InstanceID": "i-west1aabbbccc",
		  "Name": "myhost1",
		  "Region": "us-west-1",
		  "Running": false
		},
		{
		  "InstanceID": "i-east1eeefffggg",
		  "Name": "myhost2",
		  "Region": "us-east-1",
		  "Running": false
		},
		{
		  "InstanceID": "i-east1hhhiiijjj",
		  "Name": "myhost3",
		  "Region": "us-east-1",
		  "Running": false
		}
	  ]
	`

	// Unmarshal the JSON string into a struct
	var instanceList []Instance
	err = json.Unmarshal([]byte(instanceListAsString), &instanceList)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	got := instancesByRegion(instanceList)

	// Compare the maps
	areEqual := true
	for key := range want {
		if !compareInstances(want[key], got[key]) {
			areEqual = false
			break
		}
	}

	if !areEqual {
		t.Errorf("Expected %v but got %v", want, got)
	}
}

func TestContainerSlice(t *testing.T) {
	expected := `
	[
		{
		  "InstanceList": [
			{
			  "InstanceID": "i-west1aabbbccc",
			  "Name": "myhost1",
			  "Region": "us-west-1",
			  "Running": false
			}
		  ],
		  "InstanceIDs": [
			"i-west1aabbbccc"
		  ],
		  "Region": "us-west-1"
		},
		{
		  "InstanceList": [
			{
			  "InstanceID": "i-east1eeefffggg",
			  "Name": "myhost2",
			  "Region": "us-east-1",
			  "Running": false
			},
			{
			  "InstanceID": "i-east1hhhiiijjj",
			  "Name": "myhost3",
			  "Region": "us-east-1",
			  "Running": false
			}
		  ],
		  "InstanceIDs": [
			"i-east1eeefffggg",
			"i-east1hhhiiijjj"
		  ],
		  "Region": "us-east-1"
		}
	  ]
	  `
	instances, err := LoadInstancesFromYAML()
	if err != nil {
		panic(err)
	}
	instancesByRegion := instancesByRegion(instances)
	got := generateInstanceSlice(instancesByRegion)

	var containerSlice []RegionInstances
	err = json.Unmarshal([]byte(expected), &containerSlice)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	want := containerSlice

	equal := CompareContainerSlices(want, got)

	if !equal {
		t.Errorf("Expected %v but got %v", want, got)
	}
}
