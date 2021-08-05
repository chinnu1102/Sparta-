package sparta

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Utility discovery information that is necessary for compilation
// in both local and AWS Binary mode

const (
	nodeColorService     = "#720502"
	nodeColorEventSource = "#BF2803"
	nodeColorLambda      = "#F35B05"
	nodeColorAPIGateway  = "#06B5F5"
	nodeNameAPIGateway   = "API Gateway"
)

type cytoscapeData struct {
	ID               string `json:"id"`
	Image            string `json:"image"`
	BackgroundColor  string `json:"backgroundColor,omitempty"`
	Source           string `json:"source,omitempty"`
	Target           string `json:"target,omitempty"`
	Label            string `json:"label,omitempty"`
	DegreeCentrality int    `json:"degreeCentrality"`
}
type cytoscapeNode struct {
	Data    cytoscapeData `json:"data"`
	Classes string        `json:"classes,omitempty"`
}
type templateResource struct {
	KeyName string
	Data    string
}

func cytoscapeNodeID(rawData interface{}) (string, error) {
	bytes, bytesErr := json.Marshal(rawData)
	if bytesErr != nil {
		return "", bytesErr
	}
	hash := sha1.New()
	_, writeErr := hash.Write(bytes)
	if writeErr != nil {
		return "", writeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type descriptionWriter struct {
	nodes  []*cytoscapeNode
	logger *logrus.Logger
}

func (dw *descriptionWriter) writeNode(nodeName string,
	nodeColor string,
	nodeImage string) error {

	nodeID, nodeErr := cytoscapeNodeID(nodeName)
	if nodeErr != nil {
		return errors.Wrapf(nodeErr,
			"Failed to create nodeID for entry: %s",
			nodeName)
	}
	appendNode := &cytoscapeNode{
		Data: cytoscapeData{
			ID:    nodeID,
			Label: strings.Trim(nodeName, "\""),
		},
	}
	if nodeImage != "" {
		resourceItem := templateResourceForKey(nodeImage, dw.logger)
		if resourceItem != nil {
			appendNode.Data.Image = fmt.Sprintf("data:image/svg+xml;base64,%s",
				base64.StdEncoding.EncodeToString([]byte(resourceItem.Data)))
		}
	}
	dw.nodes = append(dw.nodes, appendNode)
	return nil
}

func (dw *descriptionWriter) writeEdge(fromNode string,
	toNode string,
	label string) error {

	nodeSource, nodeSourceErr := cytoscapeNodeID(fromNode)
	if nodeSourceErr != nil {
		return errors.Wrapf(nodeSourceErr,
			"Failed to create nodeID for entry: %s",
			fromNode)
	}
	nodeTarget, nodeTargetErr := cytoscapeNodeID(toNode)
	if nodeTargetErr != nil {
		return errors.Wrapf(nodeSourceErr,
			"Failed to create nodeID for entry: %s",
			toNode)
	}

	dw.nodes = append(dw.nodes, &cytoscapeNode{
		Data: cytoscapeData{
			ID:     fmt.Sprintf("%d", rand.Uint64()),
			Source: nodeSource,
			Target: nodeTarget,
			Label:  label,
		},
	})
	return nil
}

func templateResourceForKey(resourceKeyName string, logger *logrus.Logger) *templateResource {
	var resource *templateResource
	resourcePath := fmt.Sprintf("/resources/describe/%s",
		strings.TrimLeft(resourceKeyName, "/"))
	data, dataErr := _escFSString(false, resourcePath)
	if dataErr == nil {
		keyParts := strings.Split(resourcePath, "/")
		keyName := keyParts[len(keyParts)-1]
		resource = &templateResource{
			KeyName: keyName,
			Data:    data,
		}
		logger.WithFields(logrus.Fields{
			"Path":    resourcePath,
			"KeyName": keyName,
		}).Debug("Embedded resource")

	} else {
		logger.WithFields(logrus.Fields{
			"Path": resourcePath,
		}).Warn("Failed to embed resource")
	}
	return resource
}
func templateResourcesForKeys(resourceKeyNames []string, logger *logrus.Logger) []*templateResource {
	var resources []*templateResource

	for _, eachKey := range resourceKeyNames {
		loadedResource := templateResourceForKey(eachKey, logger)
		if loadedResource != nil {
			resources = append(resources, loadedResource)
		}
	}
	return resources
}

func templateCSSFiles(logger *logrus.Logger) []*templateResource {
	cssFiles := []string{"bootstrap-4.0.0/dist/css/bootstrap.min.css",
		"highlight.js/styles/xcode.css",
	}
	return templateResourcesForKeys(cssFiles, logger)
}

func templateJSFiles(logger *logrus.Logger) []*templateResource {
	jsFiles := []string{"jquery/jquery-3.3.1.min.js",
		"popper/popper.min.js",
		"bootstrap-4.0.0/dist/js/bootstrap.min.js",
		"highlight.js/highlight.pack.js",
		"dagre-0.8.4/dist/dagre.min.js",
		"cytoscape.js/dist/cytoscape.min.js",
		"cytoscape.js-dagre/cytoscape-dagre.js",
		"sparta.js",
	}
	return templateResourcesForKeys(jsFiles, logger)
}

func templateImageMap(logger *logrus.Logger) map[string]string {
	images := []string{"SpartaHelmet256.png",
		"AWS-Architecture-Icons_SVG_20200131/SVG Light/Compute/AWS-Lambda_Lambda-Function_light-bg.svg",
		"AWS-Architecture-Icons_SVG_20200131/SVG Light/Management & Governance/AWS-CloudFormation_light-bg.svg",
	}
	resources := templateResourcesForKeys(images, logger)
	imageMap := make(map[string]string)
	for _, eachResource := range resources {
		imageMap[eachResource.KeyName] = base64.StdEncoding.EncodeToString([]byte(eachResource.Data))
	}
	return imageMap
}

// TODO - this should really be smarter, including
// looking at the referred resource to understand it's
// type
func iconForAWSResource(rawEmitter interface{}) string {
	jsonBytes, jsonBytesErr := json.Marshal(rawEmitter)
	if jsonBytesErr != nil {
		jsonBytes = make([]byte, 0)
	}
	canonicalRaw := strings.ToLower(string(jsonBytes))
	iconMappings := map[string]string{
		"dynamodb":   "AWS-Architecture-Icons_SVG_20200131/SVG Light/Database/Amazon-DynamoDB_Table_light-bg.svg",
		"sqs":        "AWS-Architecture-Icons_SVG_20200131/SVG Light/Application Integration/Amazon-Simple-Queue-Service-SQS_light-bg.svg",
		"sns":        "AWS-Architecture-Icons_SVG_20200131/SVG Light/Application Integration/Amazon-Simple-Notification-Service-SNS_light-bg.svg",
		"cloudwatch": "AWS-Architecture-Icons_SVG_20200131/SVG Light/Management & Governance/Amazon-CloudWatch.svg",
		"kinesis":    "AWS-Architecture-Icons_SVG_20200131/SVG Light/Analytics/Amazon-Kinesis_light-bg.svg",
		//lint:ignore ST1018 This is the name of the icon
		"s3": "AWS-Architecture-Icons_SVG_20200131/SVG Light/Storage/Amazon-Simple-Storage-Service-S3.svg",
		"codecommit": "AWS-Architecture-Icons_SVG_20200131/SVG Light/Developer Tools/AWS-CodeCommit_light-bg.svg",
	}
	// Return it if we have it...
	for eachKey, eachPath := range iconMappings {
		if strings.Contains(canonicalRaw, eachKey) {
			return eachPath
		}
	}
	return "AWS-Architecture-Icons_SVG_20200131/SVG Light/_General/General_light-bg.svg"
}
