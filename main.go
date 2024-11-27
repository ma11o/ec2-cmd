package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

type EC2Instance struct {
	InstanceID   string
	InstanceName string
}

var instances []EC2Instance

func main() {
	rootCmd := &cobra.Command{
		Use:   "ec2connect",
		Short: "EC2 Session Manager Connection Tool",
	}

	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Interactively connect to an EC2 instance using Session Manager",
		Run:   connectToInstance,
	}

	rootCmd.AddCommand(connectCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fetchInstances() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	client := ec2.NewFromConfig(cfg)

	// EC2 インスタンス一覧を取得
	result, err := client.DescribeInstances(context.TODO(), nil)
	if err != nil {
		log.Fatalf("failed to describe instances: %v", err)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			name := "Unknown"
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
					break
				}
			}
			instanceInfo := EC2Instance{
				InstanceID:   *instance.InstanceId,
				InstanceName: name,
			}
			instances = append(instances, instanceInfo)
		}
	}
}

func connectToInstance(cmd *cobra.Command, args []string) {
	fetchInstances()

	if len(instances) == 0 {
		fmt.Println("No instances found.")
		return
	}

	var choices []string
	for _, instance := range instances {
		choices = append(choices, instance.InstanceID+":"+instance.InstanceName)
	}

	var selectedInstance string
	selectPrompt := &survey.Select{
		Message: "Select an EC2 instance to connect:",
		Options: choices,
		Default: choices[0],
	}

	err := survey.AskOne(selectPrompt, &selectedInstance)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	split := strings.Split(selectedInstance, ":")
	fmt.Printf("Connecting to Instance ID: %s\n", split[0])

	command := exec.Command("aws", "ssm", "start-session", "--target", split[0])
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	err = command.Run()
	if err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}

	fmt.Println("Session ended.")
}
