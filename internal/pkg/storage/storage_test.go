package storage

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type recordingHTTPClient struct {
	request *http.Request
	body    []byte
}

func (c *recordingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.request = req
	var err error
	c.body, err = io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

type closeTrackingFile struct {
	*bytes.Reader
	closed bool
}

func (f *closeTrackingFile) Close() error {
	f.closed = true
	return nil
}

func TestInitS3ValidatesRequiredEnv(t *testing.T) {
	originalClient := S3Client
	S3Client = nil
	t.Cleanup(func() {
		S3Client = originalClient
	})

	tests := []struct {
		name    string
		region  string
		key     string
		secret  string
		wantErr string
	}{
		{name: "missing region", key: "access", secret: "secret", wantErr: "AWS_REGION is required"},
		{name: "missing access key", region: "ap-south-1", secret: "secret", wantErr: "AWS_S3_ACCESS_KEY_ID is required"},
		{name: "missing secret key", region: "ap-south-1", key: "access", wantErr: "AWS_S3_SECRET_ACCESS_KEY is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_REGION", tc.region)
			t.Setenv("AWS_S3_ACCESS_KEY_ID", tc.key)
			t.Setenv("AWS_S3_SECRET_ACCESS_KEY", tc.secret)

			err := InitS3()
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("expected %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestInitS3SetsClient(t *testing.T) {
	originalClient := S3Client
	S3Client = nil
	t.Cleanup(func() {
		S3Client = originalClient
	})

	t.Setenv("AWS_REGION", "ap-south-1")
	t.Setenv("AWS_S3_ACCESS_KEY_ID", "access")
	t.Setenv("AWS_S3_SECRET_ACCESS_KEY", "secret")

	if err := InitS3(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if S3Client == nil {
		t.Fatal("expected S3Client to be initialized")
	}
}

func TestUploadImageValidatesEnvAndClosesFile(t *testing.T) {
	originalClient := S3Client
	S3Client = nil
	t.Cleanup(func() {
		S3Client = originalClient
	})

	tests := []struct {
		name    string
		bucket  string
		region  string
		wantErr string
	}{
		{name: "missing bucket", region: "ap-south-1", wantErr: "AWS_BUCKET_NAME is required"},
		{name: "missing region", bucket: "booknest-test", wantErr: "AWS_REGION is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_BUCKET_NAME", tc.bucket)
			t.Setenv("AWS_REGION", tc.region)

			file := &closeTrackingFile{Reader: bytes.NewReader([]byte("image-bytes"))}
			header := &multipart.FileHeader{Filename: "cover.jpg"}

			got, err := UploadImage(file, header)
			if got != "" {
				t.Fatalf("expected empty URL, got %q", got)
			}
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("expected %q, got %v", tc.wantErr, err)
			}
			if !file.closed {
				t.Fatal("expected UploadImage to close the file")
			}
		})
	}
}

func TestUploadImagePutsObjectAndReturnsPublicURL(t *testing.T) {
	originalClient := S3Client
	t.Cleanup(func() {
		S3Client = originalClient
	})

	httpClient := &recordingHTTPClient{}
	S3Client = s3.NewFromConfig(aws.Config{
		Region:      "ap-south-1",
		Credentials: credentials.NewStaticCredentialsProvider("access", "secret", ""),
		HTTPClient:  httpClient,
	})

	t.Setenv("AWS_BUCKET_NAME", "booknest-test")
	t.Setenv("AWS_REGION", "ap-south-1")

	file := &closeTrackingFile{Reader: bytes.NewReader([]byte("image-bytes"))}
	header := &multipart.FileHeader{
		Filename: "../cover.png",
		Header:   make(textproto.MIMEHeader),
	}
	header.Header.Set("Content-Type", "image/png")

	url, err := UploadImage(file, header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !file.closed {
		t.Fatal("expected UploadImage to close the file")
	}
	if !strings.HasPrefix(url, "https://booknest-test.s3.ap-south-1.amazonaws.com/book-covers/") {
		t.Fatalf("expected S3 public URL, got %q", url)
	}
	if !strings.HasSuffix(url, "-cover.png") {
		t.Fatalf("expected URL to use base filename, got %q", url)
	}
	if httpClient.request == nil {
		t.Fatal("expected S3 PutObject request")
	}
	if httpClient.request.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("expected image/png content type, got %q", httpClient.request.Header.Get("Content-Type"))
	}
	if string(httpClient.body) != "image-bytes" {
		t.Fatalf("expected uploaded image bytes, got %q", string(httpClient.body))
	}
}
