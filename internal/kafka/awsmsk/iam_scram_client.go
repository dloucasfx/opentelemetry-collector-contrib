// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awsmsk // import "github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka/awsmsk"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/aws"
	sign "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"go.uber.org/multierr"
)

const (
	Mechanism = "AWS_MSK_IAM"

	service          = "kafka-cluster"
	supportedVersion = "2020_10_22"
	scopeFormat      = `%s/%s/%s/kafka-cluster/aws4_request`
)

const (
	_ int32 = iota // Ignoring the zero value to ensure we start up correctly
	initMessage
	serverResponse
	complete
	failed
)

var (
	ErrFailedServerChallenge = errors.New("failed server challenge")
	ErrBadChallenge          = errors.New("invalid challenge data provided")
	ErrInvalidStateReached   = errors.New("invalid state reached")
)

type IAMSASLClient struct {
	MSKHostname string
	Region      string
	UserAgent   string

	signer      *sign.StreamSigner
	credentials aws.Credentials

	state     int32
	accessKey string
	secretKey string
}

type payload struct {
	Version       string `json:"version"`
	BrokerHost    string `json:"host"`
	UserAgent     string `json:"user-agent"`
	Action        string `json:"action"`
	Algorithm     string `json:"x-amz-algorithm"`
	Credentials   string `json:"x-amz-credential"`
	Date          string `json:"x-amz-date"`
	Expires       string `json:"x-amz-expires"`
	SignedHeaders string `json:"x-amz-signedheaders"`
	Signature     string `json:"x-amz-signature"`
}

type response struct {
	Version   string `json:"version"`
	RequestID string `json:"request-id"`
}

var _ sarama.SCRAMClient = (*IAMSASLClient)(nil)

func NewIAMSASLClient(mskhostname, region, useragent string) sarama.SCRAMClient {
	return &IAMSASLClient{
		MSKHostname: mskhostname,
		Region:      region,
		UserAgent:   useragent,
	}
}

func (sc *IAMSASLClient) Begin(username, password, _ string) error {
	if sc.MSKHostname == "" {
		return errors.New("missing required MSK Broker hostname")
	}

	if sc.Region == "" {
		return errors.New("missing MSK cluster region")
	}

	if sc.UserAgent == "" {
		return errors.New("missing value for MSK user agent")
	}

	var err error
	ctx := context.Background()

	credentialsProvider := credentials.NewStaticCredentialsProvider(username, password, "")

	sc.credentials, err = credentialsProvider.Retrieve(ctx)
	if err != nil {
		return err
	}
	sc.signer = sign.NewStreamSigner(sc.credentials, service, sc.Region, nil)

	sc.accessKey = username
	sc.secretKey = password
	sc.state = initMessage
	return nil
}

func (sc *IAMSASLClient) Step(challenge string) (string, error) {
	var resp string

	switch sc.state {
	case initMessage:
		if challenge != "" {
			sc.state = failed
			return "", fmt.Errorf("challenge must be empty for initial request: %w", ErrBadChallenge)
		}
		payload, err := sc.getAuthPayload()
		if err != nil {
			sc.state = failed
			return "", err
		}
		resp = string(payload)
		sc.state = serverResponse
	case serverResponse:
		if challenge == "" {
			sc.state = failed
			return "", fmt.Errorf("challenge must not be empty for server resposne: %w", ErrBadChallenge)
		}

		var resp response
		if err := json.NewDecoder(strings.NewReader(challenge)).Decode(&resp); err != nil {
			sc.state = failed
			return "", fmt.Errorf("unable to process msk challenge response: %w", multierr.Combine(err, ErrFailedServerChallenge))
		}

		if resp.Version != supportedVersion {
			sc.state = failed
			return "", fmt.Errorf("unknown version found in response: %w", ErrFailedServerChallenge)
		}

		sc.state = complete
	default:
		return "", fmt.Errorf("invalid invocation: %w", ErrInvalidStateReached)
	}

	return resp, nil
}

func (sc *IAMSASLClient) Done() bool { return sc.state == complete }

func (sc *IAMSASLClient) getAuthPayload() ([]byte, error) {
	ts := time.Now().UTC()
	ctx := context.Background()

	headers := []byte("host:" + sc.MSKHostname)

	sig, err := sc.signer.GetSignature(ctx, headers, nil, ts)
	if err != nil {
		return nil, err
	}

	// Creating a timestamp in the form of: yyyyMMdd'T'HHmmss'Z'
	date := ts.Format("20060102T150405Z")

	return json.Marshal(&payload{
		Version:       supportedVersion,
		BrokerHost:    sc.MSKHostname,
		UserAgent:     sc.UserAgent,
		Action:        "kafka-cluster:Connect",
		Algorithm:     "AWS4-HMAC-SHA256",
		Credentials:   fmt.Sprintf(scopeFormat, sc.accessKey, date[:8], sc.Region),
		Date:          date,
		SignedHeaders: "host",
		Expires:       "300", // Seconds => 5 Minutes
		Signature:     string(sig),
	})
}
