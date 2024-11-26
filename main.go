package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

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
	// EC2 インスタンス一覧を取得
	fetchInstances()

	if len(instances) == 0 {
		fmt.Println("No instances found.")
		return
	}

	// インタラクティブな選択肢を表示
	fmt.Println("Select an EC2 instance to connect:")
	for i, instance := range instances {
		fmt.Printf("[%d] ID: %s, Name: %s\n", i, instance.InstanceID, instance.InstanceName)
	}

	var choice int
	fmt.Print("Enter the number of the instance to connect: ")
	_, err := fmt.Scan(&choice)
	if err != nil || choice < 0 || choice >= len(instances) {
		fmt.Println("Invalid selection.")
		return
	}

	selected := instances[choice]
	fmt.Printf("Connecting to Instance ID: %s, Name: %s\n", selected.InstanceID, selected.InstanceName)

	// `aws ssm start-session` コマンドを実行
	command := exec.Command("aws", "ssm", "start-session", "--target", selected.InstanceID)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	// コマンドを実行してセッションを開始
	err = command.Run()
	if err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}

	fmt.Println("Session ended.")
}
