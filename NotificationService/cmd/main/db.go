package main
import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewDynamoDBClient(ctx context.Context) *dynamodb.Client {
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
            func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                return aws.Endpoint{
                    PartitionID:   "aws",
                    URL:           "http://localhost:8000", // URL for DynamoDB Local
                	SigningRegion: "us-west-2",
                }, nil
            })),
        config.WithRegion("us-west-2"),
    )
    if err != nil {
        panic("unable to load SDK config, " + err.Error())
    }

    client := dynamodb.NewFromConfig(cfg)
    return client
}