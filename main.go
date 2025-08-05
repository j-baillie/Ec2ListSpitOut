package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func createFile(name string) {
	os.Create(name)
}

func buildFullDeets() ([]byte, *ec2.DescribeInstancesOutput) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		fmt.Printf("Error was returned, session not created: %v", err)
		finishProgErr(err)
	} // session created with the creds found in default credential location

	// Create an EC2 service client
	ec2Session := ec2.New(sess)

	// Call DescribeInstances to get a list of instances
	instanceList, err := ec2Session.DescribeInstances(nil)
	if err != nil {
		fmt.Printf("Error was returned, failed to get instance list: %v", err)
		finishProgErr(err)
	}

	jsonData, err := json.MarshalIndent(instanceList, "", "    ")
	if err != nil {
		fmt.Printf("Error was returned, failed to marshall json data: %v", err)
		finishProgErr(err)
	}
	return jsonData, instanceList
}

func writeFullInstancesDeets(j []byte) {
	// throw full payload into file for prosperity
	var err = os.WriteFile("FullInstances.json", j, 0644)
	if err != nil {
		fmt.Printf("Error was returned, failed to write to file: %v", err)
		finishProgErr(err)
		return
	} else {
		fmt.Println("FULL Instance data succesfully written to FullInstances.json")
	}
}

func finishProgErr(err error) {
	// close the application after giving the user a chance to see the error message
	input := bufio.NewScanner(os.Stdin)
	fmt.Println("Program will now exit. Please investigate any error message shown")
	fmt.Println(err)
	input.Scan()
	os.Exit(0)
}

func finishProgSucc() {
	// close the application with short explanation - Vorrangig for success
	input := bufio.NewScanner(os.Stdin)
	fmt.Println("\nTwo files should now be present. The raw payload from AWS: FullInstances.json and a slimmer list of instances: TrimmedInstances.txt")
	fmt.Println("Both stopped and running instances are returned. The stopped instances are noted and the reason why they are not running.")
	fmt.Println("\nAll finished. Press any key to close this application.")
	input.Scan()
	os.Exit(0)
}

func main() {

	// build the full dataset
	jsonDataSet, instanceList := buildFullDeets()

	// write the full dataset to a file for backup
	createFile("FullInstances.json")
	writeFullInstancesDeets(jsonDataSet)

	// prepare new file for trimmed down results
	createFile("TrimmedInstances.txt")

	// Iterate through the instances and print their IDs and tags
	for _, reservation := range instanceList.Reservations { //range meaning interate over given list for the amount which is in the list

		// in here we have _, which is a "blank identifier"
		// The main use cases for this identifier is to ignore some of the values returned by
		// a function or for import side-effects. The blank identifier ignores any value returned by a function

		for _, instance := range reservation.Instances {
			content := fmt.Sprintf("\nInstance ID: %s | State is currently: %s\n", *instance.InstanceId, *instance.State.Name)
			file, err := os.OpenFile("TrimmedInstances.txt", os.O_APPEND, 0644)
			if err != nil {
				fmt.Printf("Error occured %s\n", err)
				finishProgErr(err)
			}
			file.WriteString(content)
			if instance.StateReason != nil {
				content := fmt.Sprintf("Is not running because: %s\n", *instance.StateReason.Message)
				file.WriteString(content)
			}
			if instance.Tags != nil {
				for _, tag := range instance.Tags {
					content := fmt.Sprintf("  %s: %s\n", *tag.Key, *tag.Value)
					file.WriteString(content)
				}
			} else {
				fmt.Println("  No tags found.")
				finishProgErr(err)
			}
		}
	}

	finishProgSucc()
}
