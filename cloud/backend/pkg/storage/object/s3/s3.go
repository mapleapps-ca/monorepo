package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"go.uber.org/zap"
)

// ACL constants for public and private objects
const (
	ACLPrivate    = "private"
	ACLPublicRead = "public-read"
)

type S3ObjectStorage interface {
	UploadContent(ctx context.Context, objectKey string, content []byte) error
	UploadContentWithVisibility(ctx context.Context, objectKey string, content []byte, isPublic bool) error
	UploadContentFromMulipart(ctx context.Context, objectKey string, file multipart.File) error
	UploadContentFromMulipartWithVisibility(ctx context.Context, objectKey string, file multipart.File, isPublic bool) error
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	GetDownloadablePresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)
	GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)
	DeleteByKeys(ctx context.Context, key []string) error
	Cut(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error
	CutWithVisibility(ctx context.Context, sourceObjectKey string, destinationObjectKey string, isPublic bool) error
	Copy(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error
	CopyWithVisibility(ctx context.Context, sourceObjectKey string, destinationObjectKey string, isPublic bool) error
	GetBinaryData(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadToLocalfile(ctx context.Context, objectKey string, filePath string) (string, error)
	ListAllObjects(ctx context.Context) (*s3.ListObjectsOutput, error)
	FindMatchingObjectKey(s3Objects *s3.ListObjectsOutput, partialKey string) string
	IsPublicBucket() bool
	// GeneratePresignedUploadURL creates a presigned URL for uploading objects
	GeneratePresignedUploadURL(ctx context.Context, key string, duration time.Duration) (string, error)
	ObjectExists(ctx context.Context, key string) (bool, error)
	GetObjectSize(ctx context.Context, key string) (int64, error)
}

type s3ObjectStorage struct {
	S3Client      *s3.Client
	PresignClient *s3.PresignClient
	Logger        *zap.Logger
	BucketName    string
	IsPublic      bool
}

// NewObjectStorage connects to a specific S3 bucket instance and returns a connected
// instance structure.
func NewObjectStorage(s3Config S3ObjectStorageConfigurationProvider, logger *zap.Logger) S3ObjectStorage {
	// DEVELOPERS NOTE:
	// How can I use the AWS SDK v2 for Go with DigitalOcean Spaces? via https://stackoverflow.com/a/74284205
	logger = logger.With(zap.String("component", "☁️🗄️ s3-object-storage"))
	logger.Debug("s3 initializing...")

	// STEP 1: initialize the custom `endpoint` we will connect to.
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: s3Config.GetEndpoint(),
		}, nil
	})

	// STEP 2: Configure.
	sdkConfig, err := config.LoadDefaultConfig(
		context.TODO(), config.WithRegion(s3Config.GetRegion()),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Config.GetAccessKey(), s3Config.GetSecretKey(), "")),
	)
	if err != nil {
		log.Fatal(err) // We need to crash the program at start to satisfy google wire requirement of having no errors.
	}

	// STEP 3\: Load up s3 instance.
	s3Client := s3.NewFromConfig(sdkConfig)

	// Create our storage handler.
	s3Storage := &s3ObjectStorage{
		S3Client:      s3Client,
		PresignClient: s3.NewPresignClient(s3Client),
		Logger:        logger,
		BucketName:    s3Config.GetBucketName(),
		IsPublic:      s3Config.GetIsPublicBucket(),
	}

	logger.Debug("s3 checking remote connection...")

	// STEP 4: Connect to the s3 bucket instance and confirm that bucket exists.
	doesExist, err := s3Storage.BucketExists(context.TODO(), s3Config.GetBucketName())
	if err != nil {
		log.Fatal(err) // We need to crash the program at start to satisfy google wire requirement of having no errors.
	}
	if !doesExist {
		log.Fatal("bucket name does not exist") // We need to crash the program at start to satisfy google wire requirement of having no errors.
	}

	logger.Debug("s3 initialized")

	// Return our s3 storage handler.
	return s3Storage
}

// IsPublicBucket returns whether the bucket is configured as public by default
func (s *s3ObjectStorage) IsPublicBucket() bool {
	return s.IsPublic
}

// UploadContent uploads content using the default bucket visibility setting
func (s *s3ObjectStorage) UploadContent(ctx context.Context, objectKey string, content []byte) error {
	return s.UploadContentWithVisibility(ctx, objectKey, content, s.IsPublic)
}

// UploadContentWithVisibility uploads content with specified visibility (public or private)
func (s *s3ObjectStorage) UploadContentWithVisibility(ctx context.Context, objectKey string, content []byte, isPublic bool) error {
	acl := ACLPrivate
	if isPublic {
		acl = ACLPublicRead
	}

	s.Logger.Debug("Uploading content with visibility",
		zap.String("objectKey", objectKey),
		zap.Bool("isPublic", isPublic),
		zap.String("acl", acl))

	_, err := s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(content),
		ACL:    types.ObjectCannedACL(acl),
	})
	if err != nil {
		s.Logger.Error("Failed to upload content",
			zap.String("objectKey", objectKey),
			zap.Bool("isPublic", isPublic),
			zap.Any("error", err))
		return err
	}
	return nil
}

// UploadContentFromMulipart uploads file using the default bucket visibility setting
func (s *s3ObjectStorage) UploadContentFromMulipart(ctx context.Context, objectKey string, file multipart.File) error {
	return s.UploadContentFromMulipartWithVisibility(ctx, objectKey, file, s.IsPublic)
}

// UploadContentFromMulipartWithVisibility uploads a multipart file with specified visibility
func (s *s3ObjectStorage) UploadContentFromMulipartWithVisibility(ctx context.Context, objectKey string, file multipart.File, isPublic bool) error {
	acl := ACLPrivate
	if isPublic {
		acl = ACLPublicRead
	}

	s.Logger.Debug("Uploading multipart file with visibility",
		zap.String("objectKey", objectKey),
		zap.Bool("isPublic", isPublic),
		zap.String("acl", acl))

	// Create the S3 upload input parameters
	params := &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
		ACL:    types.ObjectCannedACL(acl),
	}

	// Perform the file upload to S3
	_, err := s.S3Client.PutObject(ctx, params)
	if err != nil {
		s.Logger.Error("Failed to upload multipart file",
			zap.String("objectKey", objectKey),
			zap.Bool("isPublic", isPublic),
			zap.Any("error", err))
		return err
	}
	return nil
}

func (s *s3ObjectStorage) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	// Note: https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html#actions

	_, err := s.S3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				log.Printf("Bucket %v is available.\n", bucketName)
				exists = false
				err = nil
			default:
				log.Printf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", bucketName, err)
			}
		}
	}

	return exists, err
}

func (s *s3ObjectStorage) GetDownloadablePresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	// DEVELOPERS NOTE:
	// AWS S3 Bucket — presigned URL APIs with Go (2022) via https://ronen-niv.medium.com/aws-s3-handling-presigned-urls-2718ab247d57

	presignedUrl, err := s.PresignClient.PresignGetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket:                     aws.String(s.BucketName),
			Key:                        aws.String(key),
			ResponseContentDisposition: aws.String("attachment"), // This field allows the file to download it directly from your browser
		},
		s3.WithPresignExpires(duration))
	if err != nil {
		return "", err
	}
	return presignedUrl.URL, nil
}

func (s *s3ObjectStorage) GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	// DEVELOPERS NOTE:
	// AWS S3 Bucket — presigned URL APIs with Go (2022) via https://ronen-niv.medium.com/aws-s3-handling-presigned-urls-2718ab247d57

	bkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	presignedUrl, err := s.PresignClient.PresignGetObject(bkCtx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(duration))
	if err != nil {
		return "", err
	}
	return presignedUrl.URL, nil
}

func (s *s3ObjectStorage) DeleteByKeys(ctx context.Context, objectKeys []string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	_, err := s.S3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.BucketName),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		log.Printf("Couldn't delete objects from bucket %v. Here's why: %v\n", s.BucketName, err)
	}
	return err
}

// Cut moves a file using the default bucket visibility setting
func (s *s3ObjectStorage) Cut(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error {
	return s.CutWithVisibility(ctx, sourceObjectKey, destinationObjectKey, s.IsPublic)
}

// CutWithVisibility moves a file with specified visibility
func (s *s3ObjectStorage) CutWithVisibility(ctx context.Context, sourceObjectKey string, destinationObjectKey string, isPublic bool) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // Increase timout so it runs longer then usual to handle this unique case.
	defer cancel()

	// First copy the object with the desired visibility
	if err := s.CopyWithVisibility(ctx, sourceObjectKey, destinationObjectKey, isPublic); err != nil {
		return err
	}

	// Delete the original object
	_, deleteErr := s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(sourceObjectKey),
	})
	if deleteErr != nil {
		s.Logger.Error("Failed to delete original object:", zap.Any("deleteErr", deleteErr))
		return deleteErr
	}

	s.Logger.Debug("Original object deleted.")

	return nil
}

// Copy copies a file using the default bucket visibility setting
func (s *s3ObjectStorage) Copy(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error {
	return s.CopyWithVisibility(ctx, sourceObjectKey, destinationObjectKey, s.IsPublic)
}

// CopyWithVisibility copies a file with specified visibility
func (s *s3ObjectStorage) CopyWithVisibility(ctx context.Context, sourceObjectKey string, destinationObjectKey string, isPublic bool) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // Increase timout so it runs longer then usual to handle this unique case.
	defer cancel()

	acl := ACLPrivate
	if isPublic {
		acl = ACLPublicRead
	}

	s.Logger.Debug("Copying object with visibility",
		zap.String("sourceKey", sourceObjectKey),
		zap.String("destinationKey", destinationObjectKey),
		zap.Bool("isPublic", isPublic),
		zap.String("acl", acl))

	_, copyErr := s.S3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.BucketName),
		CopySource: aws.String(s.BucketName + "/" + sourceObjectKey),
		Key:        aws.String(destinationObjectKey),
		ACL:        types.ObjectCannedACL(acl),
	})
	if copyErr != nil {
		s.Logger.Error("Failed to copy object:",
			zap.String("sourceKey", sourceObjectKey),
			zap.String("destinationKey", destinationObjectKey),
			zap.Bool("isPublic", isPublic),
			zap.Any("copyErr", copyErr))
		return copyErr
	}

	s.Logger.Debug("Object copied successfully.")

	return nil
}

// GetBinaryData function will return the binary data for the particular key.
func (s *s3ObjectStorage) GetBinaryData(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}

	s3object, err := s.S3Client.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	return s3object.Body, nil
}

func (s *s3ObjectStorage) DownloadToLocalfile(ctx context.Context, objectKey string, filePath string) (string, error) {
	responseBin, err := s.GetBinaryData(ctx, objectKey)
	if err != nil {
		return filePath, err
	}
	out, err := os.Create(filePath)
	if err != nil {
		return filePath, err
	}
	defer out.Close()

	_, err = io.Copy(out, responseBin)
	if err != nil {
		return "", err
	}
	return filePath, err
}

func (s *s3ObjectStorage) ListAllObjects(ctx context.Context) (*s3.ListObjectsOutput, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(s.BucketName),
	}

	objects, err := s.S3Client.ListObjects(ctx, input)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

// Function will iterate over all the s3 objects to match the partial key with
// the actual key found in the S3 bucket.
func (s *s3ObjectStorage) FindMatchingObjectKey(s3Objects *s3.ListObjectsOutput, partialKey string) string {
	for _, obj := range s3Objects.Contents {

		match := strings.Contains(*obj.Key, partialKey)

		// If a match happens then it means we have found the ACTUAL KEY in the
		// s3 objects inside the bucket.
		if match == true {
			return *obj.Key
		}
	}
	return ""
}

// GeneratePresignedUploadURL creates a presigned URL for uploading objects to S3
func (s *s3ObjectStorage) GeneratePresignedUploadURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	presignedUrl, err := s.PresignClient.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(key),
			ACL:    types.ObjectCannedACL(ACLPrivate), // Always private for file uploads
		},
		s3.WithPresignExpires(duration))
	if err != nil {
		s.Logger.Error("Failed to generate presigned upload URL",
			zap.String("key", key),
			zap.Duration("duration", duration),
			zap.Error(err))
		return "", err
	}

	s.Logger.Debug("Generated presigned upload URL",
		zap.String("key", key),
		zap.Duration("duration", duration))

	return presignedUrl.URL, nil
}

// ObjectExists checks if an object exists at the given key using HeadObject
func (s *s3ObjectStorage) ObjectExists(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := s.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				// Object doesn't exist
				s.Logger.Debug("Object does not exist",
					zap.String("key", key))
				return false, nil
			case *types.NoSuchKey:
				// Object doesn't exist
				s.Logger.Debug("Object does not exist (NoSuchKey)",
					zap.String("key", key))
				return false, nil
			default:
				// Some other error occurred
				s.Logger.Error("Error checking object existence",
					zap.String("key", key),
					zap.Error(err))
				return false, err
			}
		}
		// Non-API error
		s.Logger.Error("Error checking object existence",
			zap.String("key", key),
			zap.Error(err))
		return false, err
	}

	s.Logger.Debug("Object exists",
		zap.String("key", key))
	return true, nil
}

// GetObjectSize returns the size of an object at the given key using HeadObject
func (s *s3ObjectStorage) GetObjectSize(ctx context.Context, key string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				s.Logger.Debug("Object not found when getting size",
					zap.String("key", key))
				return 0, errors.New("object not found")
			case *types.NoSuchKey:
				s.Logger.Debug("Object not found when getting size (NoSuchKey)",
					zap.String("key", key))
				return 0, errors.New("object not found")
			default:
				s.Logger.Error("Error getting object size",
					zap.String("key", key),
					zap.Error(err))
				return 0, err
			}
		}
		s.Logger.Error("Error getting object size",
			zap.String("key", key),
			zap.Error(err))
		return 0, err
	}

	if result.ContentLength == nil {
		s.Logger.Warn("Object size is nil",
			zap.String("key", key))
		return 0, errors.New("object size unavailable")
	}

	size := *result.ContentLength
	s.Logger.Debug("Retrieved object size",
		zap.String("key", key),
		zap.Int64("size", size))

	return size, nil
}
