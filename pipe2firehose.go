// Copyright 2018 Bray Almini. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// pipe2firehose is a simple tool to pipe data from stdin and push into
// AWS Kinesis Firehose or Kinesis Stream.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

var (
	awsRegion          = flag.String("region", "", "AWS Region of Firehose")
	delayLine          = flag.String("delay", "", "Time in ms to wait between each line read")
	printVersion       = flag.Bool("version", false, "Print the version")
	batchSize          = flag.Int("batch-size", 500, "Modify the number of records included per PutRecordBatch")
	awsSession         *session.Session
	awsFirehose        *firehose.Firehose
	firehoseRecords    []*firehose.Record
	deliveryStreamName string
	totalRecords       = 0
	lastPushCount      = 0
	startTime          = time.Now()
	version            = "master"
)

var usageText = `pipe2firehose version %s

Usage:
  cat data.json | pipe2firehose [options] [firehose-name]
or
  pipe2firehose [options] [firehose-name] < data.json

Options:
`

func connectFirehose() {
	// fmt.Println("connectFirehose()")

	// Create AWS Session using custom blank config
	awsConfig := &aws.Config{}
	if len(*awsRegion) != 0 {
		awsConfig.Region = aws.String(*awsRegion)
	}
	awsSession = session.Must(session.NewSession(awsConfig))
	// Create Firehose client using session
	awsFirehose = firehose.New(awsSession)
}

func flushFirehose() {
	// fmt.Println("flushFirehose()")
	_, err := awsFirehose.PutRecordBatch(
		&firehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String(deliveryStreamName),
			Records:            firehoseRecords,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed firehose.PutRecordBatch:\n%s\n", err)
		os.Exit(1)
	}
	pushed := len(firehoseRecords)
	totalRecords += pushed
	// fmt.Printf("flushFirehose() pushed %v OK\n", pushed)
	firehoseRecords = firehoseRecords[:0]
}

func pushFirehose(line string) {
	firehoseRecords = append(
		firehoseRecords,
		&firehose.Record{Data: []byte(line + "\n")},
	)
	numRecords := len(firehoseRecords)
	// fmt.Printf("pushFirehose() numRecords: %v\n", numRecords)
	if numRecords == *batchSize {
		flushFirehose()
	}
}

func echoStats() {
	time.Sleep(1 * time.Second)
	uptime := time.Since(startTime) / time.Second
	perSec := totalRecords / int(uptime)
	fmt.Printf(
		"Uptime: %v, InBuffer: %v, Pushed %v, %v/sec\n",
		int(uptime),
		len(firehoseRecords),
		totalRecords,
		perSec,
	)
	lastPushCount = totalRecords
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usageText, version))
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	if *printVersion {
		fmt.Println("pipe2firehose version " + version)
		os.Exit(0)
	}

	if len(args) == 0 {
		fmt.Println("Firehose name required")
		os.Exit(1)
	}
	deliveryStreamName = args[0]

	connectFirehose()

	go func() {
		for {
			echoStats()
		}
	}()

	// Slow down for debugging
	var delay time.Duration
	if *delayLine != "" {
		delayTmp, _ := strconv.Atoi(*delayLine)
		delay = time.Millisecond * time.Duration(delayTmp)
	}

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		// fmt.Println("main() got a line")
		pushFirehose(s.Text())
		// Slow down for debugging
		if delay != 0 {
			time.Sleep(delay)
		}
	}
	if len(firehoseRecords) > 0 {
		flushFirehose()
	}
	echoStats()
}
