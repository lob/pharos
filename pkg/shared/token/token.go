// Package token is heavily inspired in aws-iam-authenticator's token package.
// The package was modified to better work with pharo's authentication scheme
// https://github.com/kubernetes-sigs/aws-iam-authenticator/blob/1097f929eb323964ccc2f1af3f26f493e2756f7d/pkg/token/token.go
package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	awsarn "github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

const (
	pharosPrefix = "pharos-v1."
	hostRegexp   = `^sts(\.[a-z1-9\-]+)?\.amazonaws\.com$`
)

// Generator provides new tokens to be used for authenticating with the pharos-api-server.
type Generator interface {
	GetSTSToken() (string, error)
}

type generator struct {
	STSClient stsiface.STSAPI
}

// NewGenerator creates a Generator and returns it.
func NewGenerator(stsClient stsiface.STSAPI) Generator {
	return generator{stsClient}
}

// GetSTSToken returns a token that contains a presigned AWS STS request.
func (g generator) GetSTSToken() (string, error) {
	request, _ := g.STSClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})

	// Sign the request.  The expires parameter (sets the x-amz-expires header) is
	// currently ignored by STS, and the token expires 15 minutes after the x-amz-date
	// timestamp regardless.
	// https://github.com/aws/aws-sdk-go/issues/2167
	presignedURLString, err := request.Presign(60)
	if err != nil {
		return "", errors.Wrap(err, "failed to presign request")
	}

	return pharosPrefix + base64.RawURLEncoding.EncodeToString([]byte(presignedURLString)), nil
}

// Identity is returned on successful Verify() results. It contains a parsed
// version of the AWS identity used to create the token.
type Identity struct {
	// ARN is the raw Amazon Resource Name returned by sts:GetCallerIdentity
	ARN string

	// CanonicalARN is the Amazon Resource Name converted to a more canonical
	// representation. In particular, STS assumed role ARNs like
	// "arn:aws:sts::ACCOUNTID:assumed-role/ROLENAME/SESSIONNAME" are converted
	// to their IAM ARN equivalent "arn:aws:iam::ACCOUNTID:role/NAME"
	CanonicalARN string

	// AccountID is the 12 digit AWS account number.
	AccountID string

	// UserID is the unique user/role ID (e.g., "AROAAAAAAAAAAAAAAAAAA").
	UserID string

	// SessionName is the STS session name (or "" if this is not a
	// session-based identity). For EC2 instance roles, this will be the EC2
	// instance ID (e.g., "i-0123456789abcdef0"). You should only rely on it
	// if you trust that _only_ EC2 is allowed to assume the IAM Role. If IAM
	// users or other roles are allowed to assume the role, they can provide
	// (nearly) arbitrary strings here.
	SessionName string
}

// Verifier validates tokens by calling STS and returning the associated identity.
type Verifier interface {
	Verify(token string) (*Identity, error)
}

type callerIdentity struct {
	GetCallerIdentityResponse struct {
		GetCallerIdentityResult struct {
			Account string `json:"Account"`
			Arn     string `json:"Arn"`
			UserID  string `json:"UserId"`
		} `json:"GetCallerIdentityResult"`
		ResponseMetadata struct {
			RequestID string `json:"RequestId"`
		} `json:"ResponseMetadata"`
	} `json:"GetCallerIdentityResponse"`
}

type tokenVerifier struct {
	client *http.Client
}

// NewVerifier creates a Verifier that is able to verify the pharos tokens
func NewVerifier() Verifier {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	return tokenVerifier{c}
}

// verify a sts host, doc: http://docs.amazonaws.cn/en_us/general/latest/gr/rande.html#sts_region
func (v tokenVerifier) verifyHost(host string) error {
	if match, _ := regexp.MatchString(hostRegexp, host); !match {
		return errors.New(fmt.Sprintf("unexpected hostname %q in pre-signed URL", host))
	}

	return nil
}

// Verify a token is valid for the specified clusterID. On success, returns an
// Identity that contains information about the AWS principal that created the
// token. On failure, returns nil and a non-nil error.
func (v tokenVerifier) Verify(token string) (*Identity, error) {
	if !strings.HasPrefix(token, pharosPrefix) {
		return nil, errors.New(fmt.Sprintf("token is missing expected %q prefix", pharosPrefix))
	}

	tokenBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, pharosPrefix))
	if err != nil {
		return nil, errors.Wrap(err, "failed to base64 decode token")
	}

	parsedURL, err := url.Parse(string(tokenBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse STS request URL")
	}

	if parsedURL.Scheme != "https" {
		return nil, errors.New(fmt.Sprintf("unexpected scheme %q in pre-signed URL", parsedURL.Scheme))
	}

	if err = v.verifyHost(parsedURL.Host); err != nil {
		return nil, err
	}

	if parsedURL.Path != "/" {
		return nil, errors.New("unexpected path in pre-signed URL")
	}

	queryParamsLower := make(url.Values)
	queryParams := parsedURL.Query()
	for key, values := range queryParams {
		queryParamsLower.Set(strings.ToLower(key), values[0])
	}

	if queryParamsLower.Get("action") != "GetCallerIdentity" {
		return nil, errors.New("unexpected action parameter in pre-signed URL")
	}

	if queryParamsLower.Get("x-amz-date") == "" {
		return nil, errors.New("X-Amz-Date parameter must be present in pre-signed URL")
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GET request")
	}

	req.Header.Set("accept", "application/json")

	response, err := v.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error performing AWS STS GET request")
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading HTTP response")
	}

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("AWS STS error (expected HTTP 200, got HTTP %d)", response.StatusCode))
	}

	ci := &callerIdentity{}
	err = json.Unmarshal(responseBody, ci)
	if err != nil {
		return nil, err
	}

	// parse the response into an Identity
	id := &Identity{
		ARN:       ci.GetCallerIdentityResponse.GetCallerIdentityResult.Arn,
		AccountID: ci.GetCallerIdentityResponse.GetCallerIdentityResult.Account,
	}
	id.CanonicalARN, err = canonicalizeARN(id.ARN)
	if err != nil {
		return nil, err
	}

	// The user ID is either UserID:SessionName (for assumed roles) or just
	// UserID (for IAM User principals).
	userIDParts := strings.Split(ci.GetCallerIdentityResponse.GetCallerIdentityResult.UserID, ":")
	switch len(userIDParts) {
	case 2:
		id.UserID = userIDParts[0]
		id.SessionName = userIDParts[1]
	case 1:
		id.UserID = userIDParts[0]
	default:
		return nil, errors.New(fmt.Sprintf("malformed UserID %s", ci.GetCallerIdentityResponse.GetCallerIdentityResult.UserID))
	}

	return id, nil
}

// canonicalizeARN validates IAM resources are appropriate for the authenticator
// and converts STS assumed roles into the IAM role resource.
//
// Supported IAM resources are:
//   * AWS account: arn:aws:iam::123456789012:root
//   * IAM user: arn:aws:iam::123456789012:user/Bob
//   * IAM role: arn:aws:iam::123456789012:role/S3Access
//   * IAM Assumed role: arn:aws:sts::123456789012:assumed-role/Accounting-Role/Mary (converted to IAM role)
//   * Federated user: arn:aws:sts::123456789012:federated-user/Bob
//
// This function has been copied over from
// https://github.com/kubernetes-sigs/aws-iam-authenticator/blob/ed1ce8bd6af7e648f2f12bdce0c725a084fc7db7/pkg/arn/arn.go
func canonicalizeARN(arn string) (string, error) {
	parsed, err := awsarn.Parse(arn)
	if err != nil {
		return "", errors.New(fmt.Sprintf("arn '%s' is invalid: '%v'", arn, err))
	}

	if err := checkPartition(parsed.Partition); err != nil {
		return "", errors.New(fmt.Sprintf("arn '%s' does not have a recognized partition", arn))
	}

	parts := strings.Split(parsed.Resource, "/")
	resource := parts[0]

	switch parsed.Service {
	case "sts":
		switch resource {
		case "federated-user":
			return arn, nil
		case "assumed-role":
			if len(parts) < 3 {
				return "", errors.New(fmt.Sprintf("assumed-role arn '%s' does not have a role", arn))
			}
			// IAM ARNs can contain paths, part[0] is resource, parts[len(parts)] is the SessionName.
			role := strings.Join(parts[1:len(parts)-1], "/")
			return fmt.Sprintf("arn:%s:iam::%s:role/%s", parsed.Partition, parsed.AccountID, role), nil
		default:
			return "", errors.New(fmt.Sprintf("unrecognized resource %s for service sts", parsed.Resource))
		}
	case "iam":
		switch resource {
		case "role", "user", "root":
			return arn, nil
		default:
			return "", errors.New(fmt.Sprintf("unrecognized resource %s for service iam", parsed.Resource))
		}
	}

	return "", errors.New(fmt.Sprintf("service %s in arn %s is not a valid service for identities", parsed.Service, arn))
}

// This function has been copied over from
// https://github.com/kubernetes-sigs/aws-iam-authenticator/blob/ed1ce8bd6af7e648f2f12bdce0c725a084fc7db7/pkg/arn/arn.go
func checkPartition(partition string) error {
	switch partition {
	case endpoints.AwsPartitionID:
	case endpoints.AwsCnPartitionID:
	case endpoints.AwsUsGovPartitionID:
	default:
		return errors.New(fmt.Sprintf("partion %s is not recognized", partition))
	}
	return nil
}
