package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/kms"
)

//encryptFile takes in a file path and returns the KMS encrypted data
func encryptFile(targetFile *string, kmsID *string, client kms.KMS, pipe chan<- []byte) {
	text, err := ioutil.ReadFile(*targetFile)
	if err != nil {
		log.Fatal("Cannot read file: ", *targetFile)
	}

	var input kms.EncryptInput
	input.KeyId = kmsID
	input.Plaintext = text

	output, err := client.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(*kmsID),
		Plaintext: text})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output.CiphertextBlob))
	pipe <- output.CiphertextBlob
}

//writeEncryptedFile writes the encrypted data to disk and creates the folder to hold them
func writeEncryptedFile(outputFolder *string, osPerms *int, path *string, pipe chan []byte) {
	file := <-pipe
	perms := os.FileMode(*osPerms)
	filename := *outputFolder + filepath.Base(*path)

	err := ioutil.WriteFile(filename, file, perms)
	if err != nil {
		fmt.Println(os.Mkdir(filepath.Base(*outputFolder), perms))
	}
}

func main() {
	outputFolder := flag.String("output", "./encrypted/", "folder to output encrytped files to")
	kmsID := flag.String("kms", "", "KMS Key to use to encrypt the file")
	region := flag.String("region", "us-west-1", "region with KMS key")
	flag.Parse()
	files := flag.Args()
	if len(files) < 1 {
		log.Fatal("usage: ./cfgcrpyt -o=./encryptedOutPut/ /path1/file1 /path2/file2")
	}
	sess := session.Must(session.NewSession())

	client := *kms.New(sess, aws.NewConfig().WithRegion(*region))
	osPerms := int(0667)
	pipe := make(chan []byte)
	for x := range files {
		fmt.Println("Encrypting: ", files[x])
		go encryptFile(&files[x], kmsID, client, pipe)
		writeEncryptedFile(outputFolder, &osPerms, &files[x], pipe)

	}

}
