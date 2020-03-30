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

type job struct {
	hash     *string
	script64 *string
}

func (j *job) Run(client *ec2.EC2, instancePayload *ec2.RunInstancesInput) (*ec2.Reservation, error) {
	resv, err := client.RunInstances(instancePayload)
	log.Println(resv, j.hash)
	return resv, err
}

func parseJob(message *sqs.Message) job {
	body := message.Body
	hash := message.MD5OfBody
	script := fmt.Sprintf(`#!/bin/bash
	echo "%d" > /etc/tool/config.json
	/path/to/tool --Config /etc/tool/config.json && shutdown -t 1
	`, body)

	script64 := base64.StdEncoding.EncodeToString([]byte(script))
	return job{
		hash:     hash,
		script64: &script64,
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
	client := new(ec2.EC2)

	for i := range assignments {
		job := parseJob(assignments[i])
		ec2Payload.UserData = job.script64
		resv, err := job.Run(client, ec2Payload)
		if err != nil {
			log.Println("Unable to start job ", err.Error())
			log.Println("Failed Reservation ", resv.Instances[0].MetadataOptions)
		}
	}

}
