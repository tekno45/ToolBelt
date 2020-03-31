package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sqs"
)

//QueueURL is the sqs endpoint to deliver configuration files to
//Region is the region the sqs queue lives in
//CredPath is the path to the aws credentials file
//CredProfile is the profile to authenticate with from the credentials file
const (
	QueueURL     = "https://sqs.us-west-2.amazonaws.com/203310025890/testq.fifo"
	Region       = "us-west-2"
	CredProfile  = "zapproved-login"
	InstanceType = "t2.micro"
)

func runJob(client *ec2.EC2, instancePayload *ec2.RunInstancesInput, jobSpecs map[string]*string) (*ec2.Reservation, error) {
	instancePayload.UserData = jobSpecs["script"]
	resv, err := client.RunInstances(instancePayload)
	log.Println(resv, jobSpecs["hash"])
	return resv, err
}

func parseJob(message *sqs.Message) map[string]*string {
	script := fmt.Sprintf(`#!/bin/bash
	echo "%d" > /etc/tool/config.json
	/path/to/tool --Config /etc/tool/config.json && shutdown -t 1
	`, message.Body)
	script64 := base64.StdEncoding.EncodeToString([]byte(script))

	return map[string]*string{
		"body":     message.Body,
		"hash":     message.MD5OfBody,
		"script64": &script64,
	}
}

func consumeQueue(sqs *sqs.SQS, request *sqs.ReceiveMessageInput) []*sqs.Message {
	data, err := sqs.ReceiveMessage(request)
	if err != nil {
		log.Fatal("Cannot reach SQS at:", QueueURL)
	}
	return data.Messages
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		//Get MFA Token from user, automatically prompts at stdin before calling AWS
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
		Profile:                 "it-sec-admin",
	}))

	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QueueURL),
		MaxNumberOfMessages: aws.Int64(3),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(20),
	}

	ec2Payload := &ec2.RunInstancesInput{
		LaunchTemplate: &ec2.LaunchTemplateSpecification{LaunchTemplateName: aws.String("Migrator-Node")},
	}

	sqsClient := sqs.New(sess)
	assignments := consumeQueue(sqsClient, receiveParams)
	ec2Client := new(ec2.EC2)

	for i := range assignments {
		job := parseJob(assignments[i])
		resv, err := runJob(ec2Client, ec2Payload, job)
		if err != nil {
			log.Println("Unable to start job ", err.Error(), job["hash"])
			log.Println("Failed Reservation ", resv.Instances[0].MetadataOptions)
		}
	}

}
