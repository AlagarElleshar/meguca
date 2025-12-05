// Package assets manages imager file asset allocation and deallocation
package assets

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/config"
	"github.com/bakape/meguca/util"
)

// Storage interface for different storage backends
type Storage interface {
	Write(SHA1 string, fileType, thumbType uint8, src, thumb io.ReadSeeker) error
	Delete(SHA1 string, fileType, thumbType uint8) error
	CreateDirs() error
	DeleteDirs() error
	ResetDirs() error
}

var (
	storage  Storage
	s3Client *s3.Client
	s3Bucket string
)

// Only used in tests, but we still need them exported
var (
	//  StdJPEG is a JPEG sample image standard struct. Only used in tests.
	StdJPEG = common.Image{
		ImageCommon: common.ImageCommon{
			SHA1:      "012a2f912c9ee93ceb0ccb8684a29ec571990a94",
			FileType:  common.JPEG,
			ThumbType: common.WEBP,
			Dims:      StdDims["jpeg"],
			MD5:       "YOQQklgfezKbBXuEAsqopw",
			Size:      300792,
		},
		Name: "sample.jpg",
	}

	// StdDims contains resulting dimentions after thumbnailing sample images.
	// Only used in tests.
	StdDims = map[string][4]uint16{
		"jpeg": {0x43c, 0x371, 0x96, 0x79},
		"png":  {0x500, 0x2d0, 0x96, 0x54},
		"gif":  {0x248, 0x2d0, 0x7a, 0x96},
		"pdf":  {0x253, 0x34a, 0x69, 0x96},
	}
)

// Initialize storage backend based on environment
func init() {
	// Check if S3 storage should be used
	useS3 := "true"                    //os.Getenv("USE_S3_STORAGE")
	s3Bucket = "geckochen-test-bucket" //os.Getenv("S3_BUCKET_NAME")

	if useS3 == "true" && s3Bucket != "" {
		cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic(err)
		}
		s3Client = s3.NewFromConfig(cfg)
		storage = &S3Storage{}
	} else {
		storage = &LocalStorage{}
	}
}

// SetStorage allows manual override of storage backend (useful for testing)
func SetStorage(s Storage) {
	storage = s
}

// ============================================================================
// S3 Storage Implementation
// ============================================================================

type S3Storage struct{}

func (s *S3Storage) Write(SHA1 string, fileType, thumbType uint8, src, thumb io.ReadSeeker) error {
	paths := GetFilePaths(SHA1, fileType, thumbType)

	// Upload source file
	if _, err := src.Seek(0, 0); err != nil {
		return err
	}

	key := strings.TrimPrefix(filepath.ToSlash(paths[0]), "/")
	if err := uploadToS3(key, src); err != nil {
		return err
	}

	// Upload thumbnail if present
	if thumb != nil {
		if _, err := thumb.Seek(0, 0); err != nil {
			return err
		}

		thumbKey := strings.TrimPrefix(filepath.ToSlash(paths[1]), "/")
		if err := uploadToS3(thumbKey, thumb); err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Storage) Delete(SHA1 string, fileType, thumbType uint8) error {
	paths := GetFilePaths(SHA1, fileType, thumbType)

	for _, path := range paths {
		key := strings.TrimPrefix(filepath.ToSlash(path), "/")
		if err := deleteFromS3(key); err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Storage) CreateDirs() error {
	// S3 doesn't require directory creation
	return nil
}

func (s *S3Storage) DeleteDirs() error {
	// For S3, you'd need to list and delete all objects with the prefix
	// This is a simplified version - add pagination for production
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Bucket),
		Prefix: aws.String("images/"),
	}

	result, err := s3Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return err
	}

	for _, obj := range result.Contents {
		if err := deleteFromS3(*obj.Key); err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Storage) ResetDirs() error {
	if err := s.DeleteDirs(); err != nil {
		return err
	}
	return s.CreateDirs()
}

func uploadToS3(key string, body io.Reader) error {
	// Determine content type from file extension
	contentType := getContentType(key)

	input := &s3.PutObjectInput{
		Bucket:             aws.String(s3Bucket),
		Key:                aws.String(key),
		Body:               body,
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String("inline"), // Display in browser instead of download
	}

	_, err := s3Client.PutObject(context.TODO(), input)
	return err
}

func getContentType(key string) string {
	ext := filepath.Ext(key)
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".pdf":  "application/pdf",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func deleteFromS3(key string) error {
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(key),
	})
	return err
}

// ============================================================================
// Local Filesystem Storage Implementation
// ============================================================================

type LocalStorage struct{}

func (l *LocalStorage) Write(SHA1 string, fileType, thumbType uint8, src, thumb io.ReadSeeker) error {
	// Assert at least 100 MB of free disk space is available
	if !common.IsCI {
		var free uint64
		free, err := freeSpace()
		if err != nil {
			return err
		}
		if free < (100 << 20) {
			return errors.New("not enough disk space")
		}
	}

	paths := GetFilePaths(SHA1, fileType, thumbType)

	err := writeFile(paths[0], src)
	if err != nil {
		return err
	}
	if thumb != nil {
		err = writeFile(paths[1], thumb)
	}
	return err
}

func (l *LocalStorage) Delete(SHA1 string, fileType, thumbType uint8) error {
	for _, path := range GetFilePaths(SHA1, fileType, thumbType) {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (l *LocalStorage) CreateDirs() error {
	for _, dir := range [...]string{"src", "thumb"} {
		path := filepath.Join("images", dir)
		if err := os.MkdirAll(path, 0705); err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalStorage) DeleteDirs() error {
	return os.RemoveAll("images")
}

func (l *LocalStorage) ResetDirs() error {
	if err := l.DeleteDirs(); err != nil {
		return err
	}
	return l.CreateDirs()
}

// ============================================================================
// Public API Functions (delegates to current storage backend)
// ============================================================================

// GetFilePaths generates file paths of the source file and its thumbnail
func GetFilePaths(SHA1 string, fileType, thumbType uint8) (paths [2]string) {
	paths[0] = util.ConcatStrings(
		"/images/src/",
		SHA1,
		".",
		common.Extensions[fileType],
	)
	paths[1] = util.ConcatStrings(
		"/images/thumb/",
		SHA1,
		".",
		common.Extensions[thumbType],
	)
	for i := range paths {
		paths[i] = filepath.FromSlash(paths[i][1:])
	}

	return
}

// RelativeSourcePath returns a file's source path relative to the root path
func RelativeSourcePath(fileType uint8, SHA1 string) string {
	return util.ConcatStrings(
		"/assets/images/src/",
		SHA1,
		".",
		common.Extensions[fileType],
	)
}

// RelativeThumbPath returns a thumbnail's path relative to the root path
func RelativeThumbPath(thumbType uint8, SHA1 string) string {
	return util.ConcatStrings(
		"/assets/images/thumb/",
		SHA1,
		".",
		common.Extensions[thumbType],
	)
}

// ImageSearchPath returns the relative path used for image search file lookups
func ImageSearchPath(img common.ImageCommon) string {
	switch img.FileType {
	case common.JPEG, common.PNG, common.GIF:
		if img.Size < 8<<20 {
			return RelativeSourcePath(img.FileType, img.SHA1)
		}
	}
	return RelativeThumbPath(img.ThumbType, img.SHA1)
}

func imageRoot() string {
	r := config.Get().ImageRootOverride
	if r != "" {
		return r
	}
	return "/assets/images"
}

// ThumbPath returns the path to the thumbnail of an image
func ThumbPath(thumbType uint8, SHA1 string) string {
	return util.ConcatStrings(
		imageRoot(),
		"/thumb/",
		SHA1,
		".",
		common.Extensions[thumbType],
	)
}

// SourcePath returns the path to the source file on an image
func SourcePath(fileType uint8, SHA1 string) string {
	return util.ConcatStrings(
		imageRoot(),
		"/src/",
		SHA1,
		".",
		common.Extensions[fileType],
	)
}

// Return free space on image storage device
func freeSpace() (n uint64, err error) {
	var stats syscall.Statfs_t
	path, err := filepath.Abs("images/src")
	if err != nil {
		return
	}
	err = syscall.Statfs(path, &stats)
	return stats.Bavail * uint64(stats.Bsize), err
}

// Write writes file assets to disk or S3
func Write(SHA1 string, fileType, thumbType uint8, src, thumb io.ReadSeeker) error {
	return storage.Write(SHA1, fileType, thumbType, src, thumb)
}

// Write a single file to disk with the appropriate permissions and flags
func writeFile(path string, src io.ReadSeeker) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = src.Seek(0, 0)
	if err != nil {
		return
	}
	_, err = io.Copy(file, src)
	return
}

// Delete deletes file assets belonging to a single upload
func Delete(SHA1 string, fileType, thumbType uint8) error {
	return storage.Delete(SHA1, fileType, thumbType)
}

// CreateDirs creates directories for processed image storage
func CreateDirs() error {
	return storage.CreateDirs()
}

// DeleteDirs recursively deletes the image storage folder
func DeleteDirs() error {
	return storage.DeleteDirs()
}

// ResetDirs removes all contents from the image storage directories
func ResetDirs() error {
	return storage.ResetDirs()
}
