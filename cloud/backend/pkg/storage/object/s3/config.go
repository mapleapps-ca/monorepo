package s3

type S3ObjectStorageConfigurationProvider interface {
	GetAccessKey() string
	GetSecretKey() string
	GetEndpoint() string
	GetRegion() string
	GetBucketName() string
	GetIsPublicBucket() bool
}

type s3ObjectStorageConfigurationProviderImpl struct {
	accessKey      string `env:"AWS_ACCESS_KEY,required"`
	secretKey      string `env:"AWS_SECRET_KEY,required"`
	endpoint       string `env:"AWS_ENDPOINT,required"`
	region         string `env:"AWS_REGION,required"`
	bucketName     string `env:"AWS_BUCKET_NAME,required"`
	isPublicBucket bool   `env:"AWS_IS_PUBLIC_BUCKET"`
}

func NewS3ObjectStorageConfigurationProvider(accessKey, secretKey, endpoint, region, bucketName string, isPublicBucket bool) S3ObjectStorageConfigurationProvider {
	return &s3ObjectStorageConfigurationProviderImpl{
		accessKey:      accessKey,
		secretKey:      secretKey,
		endpoint:       endpoint,
		region:         region,
		bucketName:     bucketName,
		isPublicBucket: isPublicBucket,
	}
}

func (me *s3ObjectStorageConfigurationProviderImpl) GetAccessKey() string {
	return me.accessKey
}

func (me *s3ObjectStorageConfigurationProviderImpl) GetSecretKey() string {
	return me.secretKey
}

func (me *s3ObjectStorageConfigurationProviderImpl) GetEndpoint() string {
	return me.endpoint
}

func (me *s3ObjectStorageConfigurationProviderImpl) GetRegion() string {
	return me.region
}

func (me *s3ObjectStorageConfigurationProviderImpl) GetBucketName() string {
	return me.bucketName
}

func (me *s3ObjectStorageConfigurationProviderImpl) GetIsPublicBucket() bool {
	return me.isPublicBucket
}
