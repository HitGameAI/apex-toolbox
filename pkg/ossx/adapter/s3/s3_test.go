package s3_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	ossx "github.com/HitGameAI/apex-toolbox/pkg/ossx"
	s3adapter "github.com/HitGameAI/apex-toolbox/pkg/ossx/adapter/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	fixtureAccess = "fixture-access"
	fixtureSecret = "fixture-only-secret"
)

var _ ossx.Adapter = (*s3adapter.Adapter)(nil)

type fakeS3 func(*http.Request) (*http.Response, error)

func (f fakeS3) Do(req *http.Request) (*http.Response, error) { return f(req) }

func newAdapter(t *testing.T, region, endpoint, prefix string, pathStyle bool, client fakeS3) *s3adapter.Adapter {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(fixtureAccess, fixtureSecret, "")),
		config.WithHTTPClient(client), config.WithRetryMaxAttempts(1))
	if err != nil {
		t.Fatal(err)
	}
	sdk := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = pathStyle
	})
	var options []s3adapter.Option
	if prefix != "" {
		options = append(options, s3adapter.WithPrefix(prefix))
	}
	a, err := s3adapter.NewAdapter(sdk, options...)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestPresignUsesBucketPrefixAndEndpointMode(t *testing.T) {
	const key = "year 2026/报表.json"
	const endpoint = "https://objects.invalid"
	for _, tc := range []struct {
		name, region, prefix, host, path string
		pathStyle                        bool
	}{
		{"s3-path", "us-west-2", "tenant-", "objects.invalid", "/tenant-reports/" + key, true},
		{"s3-virtual-host", "us-west-2", "tenant-", "tenant-reports.objects.invalid", "/" + key, false},
		{"r2-style", "auto", "", "objects.invalid", "/reports/" + key, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			adapter := newAdapter(t, tc.region, endpoint, tc.prefix, tc.pathStyle, func(req *http.Request) (*http.Response, error) {
				t.Fatalf("presigning unexpectedly sent an HTTP request: %s", req.Method)
				return nil, errors.New("unexpected HTTP request")
			})
			if adapter.Provider() != ossx.ProviderS3 {
				t.Fatalf("provider is %v", adapter.Provider())
			}
			for method, fetch := range map[string]func(context.Context, string, string, time.Duration) (*ossx.PresignedResult, error){
				http.MethodGet: adapter.PresignedDownloadURL,
				http.MethodPut: adapter.PresignedUploadURL,
			} {
				result, err := fetch(context.Background(), "reports", key, 4*time.Minute)
				if err != nil {
					t.Fatal(err)
				}
				signedURL, err := url.Parse(result.URL)
				if err != nil || signedURL.Host != tc.host || signedURL.Path != tc.path || result.Method != method || signedURL.Scheme != "https" {
					t.Fatalf("incorrect presigned request routing for %s", method)
				}
				q := signedURL.Query()
				if q.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" || q.Get("X-Amz-Expires") != "240" ||
					!strings.HasPrefix(q.Get("X-Amz-Credential"), fixtureAccess+"/") ||
					!strings.HasSuffix(q.Get("X-Amz-Credential"), "/"+tc.region+"/s3/aws4_request") ||
					len(q.Get("X-Amz-Signature")) != 64 || !strings.Contains(q.Get("X-Amz-SignedHeaders"), "host") ||
					!strings.Contains(signedURL.EscapedPath(), "%20") {
					t.Fatalf("incorrect signed URL fields for %s", method)
				}
				if result.SignedHeader["Host"] != tc.host {
					t.Fatalf("incorrect signed host returned for %s", method)
				}
			}
		})
	}
}

func TestObjectOperationsAndErrors(t *testing.T) {
	const in = "fixture data"
	const object = "year 2026/报表.json"
	ctx := context.Background()
	var operations []string
	adapter := newAdapter(t, "us-west-2", "https://objects.invalid", "tenant-", true, func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "objects.invalid" || !strings.HasPrefix(req.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential="+fixtureAccess+"/") {
			t.Fatalf("expected a signed request to the fixture endpoint for %s", req.Method)
		}
		target := req.Method + " " + req.URL.Path
		operations = append(operations, target)
		switch target {
		case "PUT /tenant-reports/" + object:
			var body io.Reader = req.Body
			if strings.Contains(req.Header.Get("Content-Encoding"), "aws-chunked") {
				if req.Header.Get("X-Amz-Decoded-Content-Length") != "12" {
					t.Fatal("incorrect AWS chunked payload length")
				}
				body = httputil.NewChunkedReader(body)
			}
			payload, err := io.ReadAll(body)
			if err != nil || string(payload) != in || req.Header.Get("Content-Type") != "application/octet-stream" {
				t.Fatalf("upload fixture mismatch: content type %q, encoding %q, decoded bytes %d, error %v", req.Header.Get("Content-Type"), req.Header.Get("Content-Encoding"), len(payload), err)
			}
			return response(req, 200, "", http.Header{"ETag": {`"fixture"`}}), nil
		case "HEAD /tenant-reports/" + object:
			return response(req, 200, "", http.Header{"Content-Length": {"12"}, "Content-Type": {"application/json"}, "X-Amz-Meta-Example": {"fixture"}}), nil
		case "GET /tenant-reports/" + object:
			return response(req, 200, in, nil), nil
		case "DELETE /tenant-reports/" + object:
			return response(req, 204, "", nil), nil
		case "GET /tenant-reports":
			q := req.URL.Query()
			if q.Get("prefix") != "year 2026/" || q.Get("max-keys") != "2" || q.Get("list-type") != "2" {
				t.Fatal("list request lost the object prefix or limit")
			}
			return response(req, 200, `<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><IsTruncated>true</IsTruncated><Contents><Key>year 2026/one</Key><Size>12</Size></Contents><NextContinuationToken>page-two</NextContinuationToken></ListBucketResult>`, nil), nil
		case "HEAD /tenant-reports", "HEAD /health-check":
			return response(req, 200, "", nil), nil
		case "HEAD /tenant-reports/missing", "HEAD /tenant-missing":
			return response(req, 404, "", nil), nil
		default:
			t.Fatalf("unexpected SDK request: %s", target)
			return nil, errors.New("unexpected request")
		}
	})
	if _, err := s3adapter.NewAdapter(nil); !errors.Is(err, ossx.ErrNilClient) {
		t.Fatalf("unexpected nil SDK result: %v", err)
	}
	put, err := adapter.PutObjectFromBytes(ctx, "reports", object, []byte(in), "")
	if err != nil || put.Key != object || put.Size != 12 || put.ETag != `"fixture"` {
		t.Fatalf("incorrect upload result: %+v %v", put, err)
	}
	metadata, err := adapter.GetObject(ctx, "reports", object)
	if err != nil || metadata.Size != 12 || metadata.ContentType != "application/json" || metadata.Metadata["example"] != "fixture" {
		t.Fatalf("incorrect head result: %+v %v", metadata, err)
	}
	blob, err := adapter.GetObjectBlob(ctx, "reports", object)
	if err != nil || string(blob) != in {
		t.Fatalf("incorrect download result: %v", err)
	}
	if err := adapter.DeleteObject(ctx, "reports", object); err != nil {
		t.Fatal(err)
	}
	objects, next, err := adapter.ListObjects(ctx, "reports", "year 2026/", 2)
	if err != nil || len(objects) != 1 || objects[0].Key != "year 2026/one" || next != "page-two" {
		t.Fatalf("incorrect list result: %+v %q %v", objects, next, err)
	}
	if found, err := adapter.BucketExists(ctx, "reports"); !found || err != nil {
		t.Fatalf("expected an existing bucket: %v %v", found, err)
	}
	if found, err := adapter.BucketExists(ctx, "missing"); found || err != nil {
		t.Fatalf("expected old missing-bucket result: %v %v", found, err)
	}
	if _, err := adapter.GetObject(ctx, "reports", "missing"); !errors.Is(err, ossx.ErrObjectNotFound) {
		t.Fatalf("expected an object-not-found error, got %v", err)
	}
	if !adapter.IsConnected() || len(operations) != 9 || operations[len(operations)-1] != "HEAD /health-check" {
		t.Fatal("health check should use the unprefixed health-check bucket")
	}
}

func response(req *http.Request, status int, text string, headers http.Header) *http.Response {
	header := make(http.Header)
	for name, values := range headers {
		for _, value := range values {
			header.Add(name, value)
		}
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(text)),
		Request:    req,
	}
}
