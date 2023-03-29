package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Get the S3 action from the user
	fmt.Println("Select S3 action:")
	fmt.Println("1. Create bucket")
	fmt.Println("2. Delete bucket")
	fmt.Print("Enter action number: ")
	actionStr, _ := reader.ReadString('\n')
	actionStr = strings.TrimSpace(actionStr)

	action, err := strconv.Atoi(actionStr)
	if err != nil {
		fmt.Println("Invalid action number. Please enter 1 or 2.")
		os.Exit(1)
	}

	// Get the S3 bucket name from the user
	fmt.Print("Enter S3 bucket name: ")
	bucketName, _ := reader.ReadString('\n')
	bucketName = strings.TrimSpace(bucketName)

	// Get the AWS region from the user
	fmt.Print("Enter AWS region (default is us-east-1): ")
	region, _ := reader.ReadString('\n')
	region = strings.TrimSpace(region)
	if region == "" {
		region = "us-east-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		os.Exit(1)
	}

	svc := s3.New(sess)

	switch action {
	case 1:
		// Check if the bucket already exists
		_, err := svc.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err == nil {
			fmt.Printf("Bucket %s already exists.\n", bucketName)
			os.Exit(1)
		}

		// Create the S3 bucket
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			fmt.Println("Error creating bucket:", err)
			os.Exit(1)
		}

		// Wait for the bucket to be created
		err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			fmt.Println("Error waiting for bucket to be created:", err)
			os.Exit(1)
		}

		fmt.Printf("Bucket %s created successfully.\n", bucketName)

	case 2:
		// Check if the bucket exists
		_, err := svc.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			fmt.Printf("Bucket %s does not exist.\n", bucketName)
			os.Exit(1)
		}

		// Check if the bucket is empty
		params := &s3.ListObjectsInput{
			Bucket:  aws.String(bucketName),
			MaxKeys: aws.Int64(1),
		}
		resp, err := svc.ListObjects(params)
		if err != nil {
			fmt.Println("Error checking bucket contents:", err)
			os.Exit(1)
		}
		if len(resp.Contents) > 0 {
			fmt.Printf("Bucket %s is not empty.\n", bucketName)

			// Check if versioning is enabled or suspended
			vparams := &s3.GetBucketVersioningInput{
				Bucket: aws.String(bucketName),
			}
			vresp, err := svc.GetBucketVersioning(vparams)
			if err != nil {
				fmt.Println("Error getting versioning status:", err)
				os.Exit(1)
			}
			if vresp.Status != nil && (*vresp.Status == "Enabled" || *vresp.Status == "Suspended") {
				fmt.Println("Bucket versioning is enabled or suspended.")

				// Ask the user to confirm before deleting objects and versions
				fmt.Printf("Are you sure you want to delete all objects and versions in bucket %s? (y/n): ", bucketName)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer != "y" {
					fmt.Println("Aborting.")
					os.Exit(0)
				}

				// Delete all objects and versions
				dparams := &s3.ListObjectVersionsInput{
					Bucket: aws.String(bucketName),
				}
				dresp, err := svc.ListObjectVersions(dparams)
				if err != nil {
					fmt.Println("Error listing object versions:", err)
					os.Exit(1)
				}
				for _, v := range dresp.Versions {
					_, err := svc.DeleteObject(&s3.DeleteObjectInput{
						Bucket:    aws.String(bucketName),
						Key:       v.Key,
						VersionId: v.VersionId,
					})
					if err != nil {
						fmt.Printf("Error deleting object version %s: %s\n", *v.VersionId, err)
						os.Exit(1)
					}
				}
				for _, v := range dresp.DeleteMarkers {
					_, err := svc.DeleteObject(&s3.DeleteObjectInput{
						Bucket:    aws.String(bucketName),
						Key:       v.Key,
						VersionId: v.VersionId,
					})
					if err != nil {
						fmt.Printf("Error deleting object version %s: %s\n", *v.VersionId, err)
						os.Exit(1)
					}
				}

				// Wait for all objects and versions to be deleted
				err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(""),
				})
				if err != nil {
					fmt.Println("Error waiting for objects and versions to be deleted:", err)
					os.Exit(1)
				}

				fmt.Printf("All objects and versions in bucket %s deleted successfully.\n", bucketName)
			} else {
				fmt.Println("Bucket versioning is not enabled.")
				fmt.Printf("Please delete all objects in bucket %s manually.\n", bucketName)
			}
			os.Exit(0)
		}

		// Delete the bucket
		_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			fmt.Println("Error deleting bucket:", err)
			os.Exit(1)
		}

		fmt.Printf("Bucket %s deleted successfully.\n", bucketName)

	default:
		fmt.Println("Invalid action number. Please enter 1 or 2.")
		os.Exit(1)
	}
}
