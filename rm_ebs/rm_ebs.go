package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		os.Exit(1)
	}

	svc := ec2.New(sess)

	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("status"),
				Values: []*string{aws.String("available")},
			},
		},
	}

	volumes, err := svc.DescribeVolumes(input)
	if err != nil {
		fmt.Println("Error describing volumes:", err)
		os.Exit(1)
	}

	for _, vol := range volumes.Volumes {
		fmt.Printf("Deleting volume %s...\n", *vol.VolumeId)
		_, err := svc.DeleteVolume(&ec2.DeleteVolumeInput{
			VolumeId: vol.VolumeId,
		})
		if err != nil {
			fmt.Printf("Error deleting volume %s: %s\n", *vol.VolumeId, err)
		}
	}

	fmt.Println("Done!")
}
