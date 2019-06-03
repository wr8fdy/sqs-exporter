# AWS SQS exporter

A Prometheus SQS metrics exporter

## Metrics

| Metric  | Description |
| ------  | ----------- |
| aws\_sqs\_approximate\_number\_of\_messages | Number of messages available |
| aws\_sqs\_approximate\_number\_of\_messages\_delayed | Number of messages delayed |
| aws\_sqs\_approximate\_number\_of\_messages\_not\_visible | Number of messages in flight |

## Lables

- queue name 
- tags - you can disable with flag

For more information see the [AWS SQS Documentation](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-message-attributes.html)

## Configuration

Flags:

- interval - how often to update queue list, env INTERVAL
- prefix - filter queue list by prefix (filtered on AWS API side), env PREFIX
- regex - filter queues by regex (filtered in app), env REGEX
- tags - add tags as lables, env TAGS

Credentials to AWS are provided in the following order:

- Environment variables (AWS\_ACCESS\_KEY\_ID and AWS\_SECRET\_ACCESS\_KEY)
- Shared credentials file (~/.aws/credentials)
- IAM role for Amazon EC2

For more information see the [AWS SDK Documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)

### AWS IAM permissions

The app needs sqs list and read access to the sqs policies

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "sqs:GetQueueAttributes",
                "sqs:GetQueueUrl",
                "sqs:ListDeadLetterSourceQueues",
                "sqs:ListQueueTags"
                "sqs:ListQueues",
            ],
            "Resource": "*"
        }
    ]
}
```

## Running

**You need to specify the region you to connect to**
Running on an ec2 machine using IAM roles:
`docker run -e AWS_REGION=<region> -d -p 9108:9108 sqs-exporter`

Or running it externally:
`docker run -d -p 9108:9108 -e AWS_ACCESS_KEY_ID=<access_key> -e AWS_SECRET_ACCESS_KEY=<secret_key> -e AWS_REGION=<region>  sqs-exporter`
