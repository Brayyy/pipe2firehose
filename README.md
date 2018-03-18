# pipe2firehose

A simple tool to pipe data from stdin and push into AWS Kinesis Firehose or Kinesis Stream.

```
Usage:
  cat data.json | pipe2firehose [options] [firehose-name]
or
  pipe2firehose [options] [firehose-name] < data.json

Options:
  -batch-size int
    	Modify the number of records included per PutRecordBatch (default 500)
  -delay string
    	Time in ms to wait between each line read
  -region string
    	AWS Region of Firehose
  -version
    	Print the version
```

The `-delay` option is helpful if you are trying to slowly feed data into Kinesis over a long period of time rather than blast the whole file all at once.

As this project uses the official AWS SDK for Go, you can make use of [environment variables](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) to specify your AWS credentials and region.

```bash
# Linux, OS X, or Unix
export AWS_ACCESS_KEY_ID=YOUR_AKID
export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
export AWS_SESSION_TOKEN=TOKEN
export AWS_REGION=us-east-1
```

```
# Windows
set AWS_ACCESS_KEY_ID=YOUR_AKID
set AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
set AWS_SESSION_TOKEN=TOKEN
set AWS_REGION=us-east-1
```