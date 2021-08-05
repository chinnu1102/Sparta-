package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	validator "gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

/******************************************************************************/
// Global options
type optionsLinkStruct struct {
	StackName       string `validate:"required"`
	OutputDirectory string `validate:"required"`
}

var optionsLink optionsLinkStruct

// RootCmd represents the root Cobra command invoked for the discovery
// and serialization of an existing CloudFormation stack
var RootCmd = &cobra.Command{
	Use:   "link",
	Short: "Link is a tool to discover and serialize a prexisting CloudFormation stack",
	Long:  "",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		validateErr := validate.Struct(optionsLink)
		if nil != validateErr {
			return validateErr
		}
		// Make sure the output value is a directory
		osStat, osStatErr := os.Stat(optionsLink.OutputDirectory)
		if nil != osStatErr {
			return osStatErr
		}
		if !osStat.IsDir() {
			return errors.Errorf("--output (%s) is not a valid directory", optionsLink.OutputDirectory)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the output and stuff it to a file
		sess, err := session.NewSession()
		if err != nil {
			return errors.Wrap(err, "Attempting to create session")
		}

		svc := cloudformation.New(sess)

		params := &cloudformation.DescribeStacksInput{
			StackName: aws.String(optionsLink.StackName),
		}
		describeStacksResponse, describeStacksResponseErr := svc.DescribeStacks(params)

		if describeStacksResponseErr != nil {
			return describeStacksResponseErr
		}

		stackInfo, stackInfoErr := json.Marshal(describeStacksResponse)
		if stackInfoErr != nil {
			return errors.Wrapf(stackInfoErr, "Failed to describe stacks")
		}
		outputFilepath := filepath.Join(optionsLink.OutputDirectory, fmt.Sprintf("%s.json", optionsLink.StackName))
		err = ioutil.WriteFile(outputFilepath, stackInfo, 0644)
		if nil != err {
			return errors.Wrap(err, "Attempting to write output file")
		}
		fmt.Println("Created file: " + outputFilepath)
		fmt.Println(describeStacksResponse)
		return nil
	},
}

func init() {
	validate = validator.New()
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringVar(&optionsLink.StackName, "stackName", "", "CloudFormation Stack Name/ID to query")
	RootCmd.PersistentFlags().StringVar(&optionsLink.OutputDirectory, "output", "", "Output directory")
}

func main() {
	// Take a stack name and an output file...
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
