package token

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/stretchr/testify/assert"
)

const (
	account = "123456789012"
	userID  = "Alice"
)

type mockClient struct {
	stsiface.STSAPI
}

func (m mockClient) GetCallerIdentityRequest(input *sts.GetCallerIdentityInput) (*request.Request, *sts.GetCallerIdentityOutput) {
	return &request.Request{
		HTTPRequest: &http.Request{
			URL: &url.URL{
				Scheme: "https",
				Host:   "localhost",
				Path:   "/",
			},
		},
		Operation: &request.Operation{},
	}, &sts.GetCallerIdentityOutput{}
}

func TestGetSTSToken(t *testing.T) {
	g := NewGenerator(&mockClient{})

	token, err := g.GetSTSToken()
	assert.NoError(t, err)
	assert.Equal(t, "pharos-v1.aHR0cHM6Ly9sb2NhbGhvc3Qv", token)
}

func validationErrorTest(t *testing.T, token string, expectedErr string) {
	t.Helper()
	_, err := tokenVerifier{}.Verify(token)
	errorContains(t, err, expectedErr)
}

func errorContains(t *testing.T, err error, expectedErr string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("err should have contained '%s' was '%s'", expectedErr, err)
	}
}

var (
	now        = time.Now()
	timeStr    = now.UTC().Format("20060102T150405Z")
	validToken = toToken(validURL)
	validURL   = fmt.Sprintf("https://sts.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-expires=60&x-amz-date=%s", timeStr)
)

func toToken(url string) string {
	return pharosPrefix + base64.RawURLEncoding.EncodeToString([]byte(url))
}

func newVerifier(statusCode int, body string, err error) Verifier {
	var rc io.ReadCloser
	if body != "" {
		rc = ioutil.NopCloser(bytes.NewReader([]byte(body)))
	}
	return tokenVerifier{
		client: &http.Client{
			Transport: &roundTripper{
				err: err,
				resp: &http.Response{
					StatusCode: statusCode,
					Body:       rc,
				},
			},
		},
	}
}

type roundTripper struct {
	err  error
	resp *http.Response
}

type errorReadCloser struct {
}

func (r errorReadCloser) Read(b []byte) (int, error) {
	return 0, errors.New("An Error")
}

func (r errorReadCloser) Close() error {
	return nil
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.resp, rt.err
}

func jsonResponse(arn, account, userid string) string {
	response := callerIdentity{}
	response.GetCallerIdentityResponse.GetCallerIdentityResult.Account = account
	response.GetCallerIdentityResponse.GetCallerIdentityResult.Arn = arn
	response.GetCallerIdentityResponse.GetCallerIdentityResult.UserID = userid
	data, _ := json.Marshal(response)
	return string(data)
}

func TestSTSEndpoints(t *testing.T) {
	verifier := tokenVerifier{}
	globalR := "sts.amazonaws.com"
	usEast1R := "sts.us-east-1.amazonaws.com"
	usEast2R := "sts.us-east-2.amazonaws.com"
	usWest1R := "sts.us-west-1.amazonaws.com"
	usWest2R := "sts.us-west-2.amazonaws.com"
	apSouth1R := "sts.ap-south-1.amazonaws.com"
	apNorthEast1R := "sts.ap-northeast-1.amazonaws.com"
	apNorthEast2R := "sts.ap-northeast-2.amazonaws.com"
	apSouthEast1R := "sts.ap-southeast-1.amazonaws.com"
	apSouthEast2R := "sts.ap-southeast-2.amazonaws.com"
	caCentral1R := "sts.ca-central-1.amazonaws.com"
	euCenteral1R := "sts.eu-central-1.amazonaws.com"
	euWest1R := "sts.eu-west-1.amazonaws.com"
	euWest2R := "sts.eu-west-2.amazonaws.com"
	euWest3R := "sts.eu-west-3.amazonaws.com"
	euNorth1R := "sts.eu-north-1.amazonaws.com"
	saEast1R := "sts.sa-east-1.amazonaws.com"

	hosts := []string{globalR, usEast1R, usEast2R, usWest1R, usWest2R, apSouth1R, apNorthEast1R, apNorthEast2R, apSouthEast1R, apSouthEast2R, caCentral1R, euCenteral1R, euWest1R, euWest2R, euWest3R, euNorth1R, saEast1R}

	for _, host := range hosts {
		if err := verifier.verifyHost(host); err != nil {
			t.Errorf("%s is not valid endpoints host", host)
		}
	}
}

func TestVerifyTokenPreSTSValidations(t *testing.T) {
	validationErrorTest(t, "pharos-v2.asdfasdfa", "token is missing expected \"pharos-v1.\" prefix")
	validationErrorTest(t, "pharos-v1.decodingerror", "illegal base64 data")
	validationErrorTest(t, toToken(":ab:cd.af:/asda"), "missing protocol scheme")
	validationErrorTest(t, toToken("http://"), "unexpected scheme")
	validationErrorTest(t, toToken("https://google.com"), fmt.Sprintf("unexpected hostname %q in pre-signed URL", "google.com"))
	validationErrorTest(t, toToken("https://sts.amazonaws.com/abc"), "unexpected path in pre-signed URL")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=NotGetCallerIdenity"), "unexpected action parameter in pre-signed URL")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=GetCallerIdentity"), "X-Amz-Date parameter must be present in pre-signed URL")
}

func TestVerifyHTTPError(t *testing.T) {
	_, err := newVerifier(0, "", errors.New("")).Verify(validToken)
	errorContains(t, err, "error performing AWS STS GET request")
	assert.Error(t, err)
}

func TestVerifyHTTP403(t *testing.T) {
	_, err := newVerifier(403, " ", nil).Verify(validToken)
	errorContains(t, err, "AWS STS error (expected HTTP 200, got")
	assert.Error(t, err)
}

func TestVerifyBodyReadError(t *testing.T) {
	verifier := tokenVerifier{
		client: &http.Client{
			Transport: &roundTripper{
				err: nil,
				resp: &http.Response{
					StatusCode: 200,
					Body:       errorReadCloser{},
				},
			},
		},
	}
	_, err := verifier.Verify(validToken)
	errorContains(t, err, "error reading HTTP response")
	assert.Error(t, err)
}

func TestVerifyUnmarshalJSONError(t *testing.T) {
	_, err := newVerifier(200, "xxxx", nil).Verify(validToken)
	errorContains(t, err, "invalid character")
	assert.Error(t, err)
}

func TestVerifyInvalidCanonicalARNError(t *testing.T) {
	_, err := newVerifier(200, jsonResponse("arn", "1000", "userid"), nil).Verify(validToken)
	errorContains(t, err, "arn 'arn' is invalid:")
	assert.Error(t, err)
}

func TestVerifyInvalidUserIDError(t *testing.T) {
	_, err := newVerifier(200, jsonResponse("arn:aws:iam::123456789012:user/Alice", "123456789012", "not:valid:userid"), nil).Verify(validToken)
	errorContains(t, err, "malformed UserID")
	assert.Error(t, err)
}

func TestVerifyNoSession(t *testing.T) {
	arn := "arn:aws:iam::123456789012:user/Alice"
	identity, err := newVerifier(200, jsonResponse(arn, account, userID), nil).Verify(validToken)
	if err != nil {
		t.Errorf("expected error to be nil was %q", err)
	}
	if identity.ARN != arn {
		t.Errorf("expected ARN to be %q but was %q", arn, identity.ARN)
	}
	if identity.CanonicalARN != arn {
		t.Errorf("expected CannonicalARN to be %q but was %q", arn, identity.CanonicalARN)
	}
	if identity.UserID != userID {
		t.Errorf("expected Username to be %q but was %q", userID, identity.UserID)
	}
}

func TestVerifySessionName(t *testing.T) {
	arn := "arn:aws:iam::123456789012:user/Alice"
	account := "123456789012"
	userID := "Alice"
	session := "session-name"
	identity, err := newVerifier(200, jsonResponse(arn, account, userID+":"+session), nil).Verify(validToken)
	if err != nil {
		t.Errorf("expected error to be nil was %q", err)
	}
	if identity.UserID != userID {
		t.Errorf("expected Username to be %q but was %q", userID, identity.UserID)
	}
	if identity.SessionName != session {
		t.Errorf("expected Session to be %q but was %q", session, identity.SessionName)
	}
}

func TestVerifyCanonicalARN(t *testing.T) {
	arn := "arn:aws:sts::123456789012:assumed-role/Alice/extra"
	canonicalARN := "arn:aws:iam::123456789012:role/Alice"
	account := "123456789012"
	userID := "Alice"
	session := "session-name"
	identity, err := newVerifier(200, jsonResponse(arn, account, userID+":"+session), nil).Verify(validToken)
	if err != nil {
		t.Errorf("expected error to be nil was %q", err)
	}
	if identity.ARN != arn {
		t.Errorf("expected ARN to be %q but was %q", arn, identity.ARN)
	}
	if identity.CanonicalARN != canonicalARN {
		t.Errorf("expected CannonicalARN to be %q but was %q", canonicalARN, identity.CanonicalARN)
	}
}

func TestCanonicalizeARN(t *testing.T) {
	var arnTests = []struct {
		arn      string // input arn
		expected string // canonacalized arn
		err      error  // expected error value
	}{
		{"NOT AN ARN", "", fmt.Errorf("Not an arn")},
		{"arn:aws:iam::123456789012:user/Alice", "arn:aws:iam::123456789012:user/Alice", nil},
		{"arn:aws:iam::123456789012:role/Users", "arn:aws:iam::123456789012:role/Users", nil},
		{"arn:aws:sts::123456789012:assumed-role/Admin/Session", "arn:aws:iam::123456789012:role/Admin", nil},
		{"arn:aws:sts::123456789012:federated-user/Bob", "arn:aws:sts::123456789012:federated-user/Bob", nil},
		{"arn:aws:iam::123456789012:root", "arn:aws:iam::123456789012:root", nil},
		{"arn:aws:sts::123456789012:assumed-role/Org/Team/Admin/Session", "arn:aws:iam::123456789012:role/Org/Team/Admin", nil},
	}

	for _, tc := range arnTests {
		actual, err := canonicalizeARN(tc.arn)
		if err != nil && tc.err == nil || err == nil && tc.err != nil {
			t.Errorf("Canoncialize(%s) expected err: %v, actual err: %v", tc.arn, tc.err, err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Canonicalize(%s) expected: %s, actual: %s", tc.arn, tc.expected, actual)
		}
	}
}
