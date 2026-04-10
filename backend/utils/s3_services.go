package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// DeleteS3Namespace deletes all files under the given namespace folder in S3
func DeleteS3Namespace(bucket, namespace, region string) error {
	// Build the prefix path
	prefix := fmt.Sprintf("uploads/kb/%s/", namespace)

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := s3.NewFromConfig(cfg)

	// List objects under prefix
	listOut, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	if len(listOut.Contents) == 0 {
		log.Printf("No objects found under prefix %s", prefix)
		return nil
	}

	// Build delete identifiers
	var objects []types.ObjectIdentifier
	for _, obj := range listOut.Contents {
		objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
	}

	// Delete objects
	_, err = client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}

	log.Printf("✅ Deleted folder %s from bucket %s", prefix, bucket)
	return nil
}

// DeleteS3Document deletes a single file from the given knowledge base namespace in S3
func DeleteS3Document(bucket, namespace, fileName, region string) error {
	// Build the file path (same structure as uploads/kb/{namespace}/{fileName})
	key := fmt.Sprintf("uploads/kb/%s/%s", namespace, fileName)

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := s3.NewFromConfig(cfg)

	// Delete the file
	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	log.Printf("✅ Deleted file %s from bucket %s", key, bucket)
	return nil
}

// ReadS3FileAsJSON reads an S3 object and unmarshals its content into a provided target variable.
func ReadS3FileAsJSON(bucket, key, region string, target interface{}) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Get the object
	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get S3 object %s from bucket %s: %w", key, bucket, err)
	}
	defer output.Body.Close()

	// Read the file content
	body, err := io.ReadAll(output.Body)
	if err != nil {
		return fmt.Errorf("failed to read S3 object body: %w", err)
	}

	// Try to unmarshal JSON into the target struct/map
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	log.Printf("✅ Successfully read and parsed S3 file: s3://%s/%s", bucket, key)
	return nil
}

func S3ObjectExists(bucket, key, region string) (bool, error) {
	ctx := context.TODO()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return false, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// HEAD request (no object download)
	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err == nil {
		return true, nil
	}

	// Object does not exist
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return false, nil
	}

	// Permission or other AWS error
	return false, fmt.Errorf("failed to check S3 object %s/%s: %w", bucket, key, err)
}

func ReadS3ObjectBytes(bucket, key, region string) ([]byte, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	return io.ReadAll(out.Body)
}

// ReadS3JSONLines reads a JSONL (one JSON object per line) file from S3,
// unmarshals each line into an element, and returns a slice of results.
func ReadS3JSONLines[T any](bucket, key, region string) ([]T, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Fetch the object
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer resp.Body.Close()

	// Prepare scanner for line-by-line reading
	scanner := bufio.NewScanner(resp.Body)
	var results []T

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		if len(line) == 0 {
			continue // skip empty lines
		}

		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			log.Printf("⚠️ Failed to parse line %d: %v\n", lineNum, err)
			continue // skip bad lines instead of failing whole read
		}

		results = append(results, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file line by line: %w", err)
	}

	log.Printf("✅ Successfully read %d JSON lines from s3://%s/%s", len(results), bucket, key)
	return results, nil
}

func SaveToS3(localFilePath, bucket, key, region string) (string, error) {
	// Load AWS config (from env vars, ~/.aws/credentials, or IAM role)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Open local file
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload file
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
		ACL:    types.ObjectCannedACLPrivate, // use PublicRead if you want public
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Build the file URL
	// Public URL if bucket is public
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, cfg.Region, key)

	return url, nil
}

// UploadBase64ToS3 uploads base64 encoded image to S3 and returns the public URL
func UploadBase64ToS3(base64Str, userId, bucket, region string) (string, error) {
	// Load AWS config (from env vars, ~/.aws/credentials, or IAM role)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Split "data:image/png;base64,xxxx"
	commaIndex := strings.Index(base64Str, ",")
	if commaIndex < 0 {
		return "", nil
	}

	header := base64Str[:commaIndex]
	data := base64Str[commaIndex+1:]

	// Get file extension from MIME type
	mimeType := strings.TrimSuffix(strings.TrimPrefix(header, "data:"), ";base64")
	exts, _ := mime.ExtensionsByType(mimeType)
	ext := ".png"
	if len(exts) > 0 {
		ext = exts[0]
	}

	// Decode image data
	imageData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	key := "profiles/" + userId + ext

	// Upload to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(mimeType),
		ACL:         types.ObjectCannedACLPrivate, // use PublicRead if you want public
	})
	if err != nil {
		return "", err
	}

	// Return public URL
	return "https://" + bucket + ".s3." + region + ".amazonaws.com/" + key, nil
}

// SaveFileStreamToS3 uploads a file stream (any type) to S3 and returns its public URL (if ACL PublicRead)
func SaveFileStreamToS3(file io.Reader, bucket, key, region string) (string, error) {
	// Load AWS config (reads env vars or ~/.aws/credentials)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Read file stream into buffer (since S3 PutObject requires io.Reader)
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file stream: %w", err)
	}

	// Upload to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    types.ObjectCannedACLPrivate, // change to PublicRead if needed
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return S3 object URL (not presigned, just bucket+key path)
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	return url, nil
}

// Save plain string content into S3
func SaveStringToS3(content, bucket, key, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload text: %v", err)
	}

	// Return the S3 object URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key), nil
}

func UploadFileToS3(filename, bucket, region string, file multipart.File) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", err
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	key := "support_attachments/" + time.Now().Format("20060102_150405_") + filepath.Base(filename)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key), nil
}

// PreSignURL generates a presigned URL for downloading an S3 object
func PreSignURL(bucket, s3URL, region string) (string, string, error) {
	// Example s3URL: https://trugen.s3.us-east-2.amazonaws.com/uploads/kb/.../sample.txt
	parsed, err := url.Parse(s3URL)
	if err != nil {
		return "", "", fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Extract bucket and key
	// Host format: <bucket>.s3.<region>.amazonaws.com
	parts := strings.Split(parsed.Host, ".")
	if len(parts) < 4 || parts[1] != "s3" {
		return "", "", fmt.Errorf("invalid S3 host format: %s", parsed.Host)
	}
	//bucket := parts[0]                        // trugen
	key := strings.TrimLeft(parsed.Path, "/") // uploads/kb/.../sample.txt

	// Detect MIME type from file extension
	mimeType := mime.TypeByExtension(path.Ext(key))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", "", fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	// Generate presigned URL (60 min)
	presigned, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket:                     aws.String(bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String("inline"),
		ResponseContentType:        aws.String(mimeType),
	}, s3.WithPresignExpires(60*time.Minute))
	if err != nil {
		return "", "", fmt.Errorf("failed to presign URL: %w", err)
	}

	// Generate presigned URL (60 min)
	presigneddownload, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(60*time.Minute))
	if err != nil {
		return "", "", fmt.Errorf("failed to presign URL: %w", err)
	}

	return presigned.URL, presigneddownload.URL, nil
}

// CheckS3FileAvailability checks if a file exists at the given S3 presigned URL.
// It tries HEAD first, then falls back to GET if HEAD is not allowed.
func CheckS3FileAvailability(presignedURL string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// First, try HEAD
	req, err := http.NewRequest("HEAD", presignedURL, nil)
	if err != nil {
		fmt.Println("Error creating HEAD request:", err)
		return false
	}

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	// If HEAD fails, try GET (some presigned URLs only allow GET)
	req, err = http.NewRequest("GET", presignedURL, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
		return false
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	fmt.Printf("File not available. HTTP status: %d\n", resp.StatusCode)
	return false
}

func GenerateVideoPresignedURL(s3BaseURL, region, videoBucket string) (string, error) {
	parsed, err := url.Parse(s3BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid S3 base URL: %w", err)
	}

	key := strings.TrimLeft(parsed.Path, "/")
	if key == "" {
		return "", fmt.Errorf("could not extract S3 key from URL: %s", s3BaseURL)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	presigner := s3.NewPresignClient(s3.NewFromConfig(cfg))

	out, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(videoBucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(30*time.Minute))
	if err != nil {
		return "", fmt.Errorf("failed to presign video URL: %w", err)
	}

	return out.URL, nil
}

func GeneratePresignedUploadURL(bucket, region, key, contentType string, contentLength int64, expiry time.Duration) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	presigner := s3.NewPresignClient(s3.NewFromConfig(cfg))

	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(contentLength),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	out, err := presigner.PresignPutObject(context.TODO(), input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign PUT URL: %w", err)
	}

	return out.URL, nil
}
