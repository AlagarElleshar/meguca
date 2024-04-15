package imager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	sess       *session.Session
	svc        *mediaconvert.MediaConvert
	downloader *s3manager.Downloader
	bucket     = "nekochen"
	jobs       = make(map[string]ConvertJob)
)

type ConvertJob struct {
	ID      *string
	success chan bool
}

type JobEvent struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Password string `json:"password"`
}

func init() {
	var err error
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Replace with your desired region
	})
	if err != nil {
		panic(fmt.Errorf("Error creating session: %v", err))
	}
	svc = mediaconvert.New(sess)

	downloader = s3manager.NewDownloader(sess)
}

func HandleJobEvent(event *JobEvent) {
	if event.Status == "COMPLETE" {
		jobs[event.JobID].success <- true
	} else if event.Status == "ERROR" {
		jobs[event.JobID].success <- false
	}
}

func startJob(url string, id *string, rotation int) (*mediaconvert.CreateJobOutput, error) {
	var rotationString string
	if rotation == 0 {
		rotationString = "DEGREE_0"
	} else {
		n := strconv.Itoa(rotation)
		rotationString = "DEGREES_" + n
	}
	jobSettings := &mediaconvert.JobSettings{
		Inputs: []*mediaconvert.Input{
			{
				FileInput:     aws.String(url),
				VideoSelector: &mediaconvert.VideoSelector{Rotate: &rotationString},
				AudioSelectors: map[string]*mediaconvert.AudioSelector{
					"Audio Selector 1": &mediaconvert.AudioSelector{
						DefaultSelection: aws.String("DEFAULT"),
					},
				},
			},
		},
		TimecodeConfig: &mediaconvert.TimecodeConfig{
			Source: aws.String("ZEROBASED"),
		},
		OutputGroups: []*mediaconvert.OutputGroup{
			{
				Name: aws.String("File Group"),
				Outputs: []*mediaconvert.Output{
					{
						ContainerSettings: &mediaconvert.ContainerSettings{
							Container:   aws.String("MP4"),
							Mp4Settings: &mediaconvert.Mp4Settings{
								// Add any additional MP4 settings here
							},
						},
						VideoDescription: &mediaconvert.VideoDescription{
							CodecSettings: &mediaconvert.VideoCodecSettings{
								Codec: aws.String("H_264"),
								H264Settings: &mediaconvert.H264Settings{
									MaxBitrate:         aws.Int64(15000000),
									RateControlMode:    aws.String("QVBR"),
									SceneChangeDetect:  aws.String("TRANSITION_DETECTION"),
									QualityTuningLevel: aws.String("SINGLE_PASS"),
								},
							},
						},
						AudioDescriptions: []*mediaconvert.AudioDescription{
							{
								AudioSourceName: aws.String("Audio Selector 1"),
								CodecSettings: &mediaconvert.AudioCodecSettings{
									Codec: aws.String("AAC"),
									AacSettings: &mediaconvert.AacSettings{
										Bitrate:    aws.Int64(160000),
										CodingMode: aws.String("CODING_MODE_2_0"),
										SampleRate: aws.Int64(48000),
									},
								},
							},
						},
					},
				},
				OutputGroupSettings: &mediaconvert.OutputGroupSettings{
					Type: aws.String("FILE_GROUP_SETTINGS"),
					FileGroupSettings: &mediaconvert.FileGroupSettings{
						Destination: aws.String("s3://nekochen/" + *id),
						DestinationSettings: &mediaconvert.DestinationSettings{
							S3Settings: &mediaconvert.S3DestinationSettings{
								StorageClass: aws.String("STANDARD"),
							},
						},
					},
				},
			},
		},
	}

	input := &mediaconvert.CreateJobInput{
		Role:                 aws.String("arn:aws:iam::058264480997:role/service-role/MediaConvert_Default_Role"), // Replace with your MediaConvert role ARN
		Settings:             jobSettings,
		StatusUpdateInterval: aws.String("SECONDS_60"),
		Queue:                aws.String("arn:aws:mediaconvert:us-east-1:058264480997:queues/Default"), // Replace with your MediaConvert queue ARN
	}

	output, err := svc.CreateJob(input)
	if err != nil {
		return nil, err
	}
	jobs[*output.Job.Id] = ConvertJob{
		id, make(chan bool, 1),
	}
	return output, err
}
func downloadConverted(url string, id *string, file string, rotation int) (fileSize int64, err error) {
	startTime := time.Now()

	job, err := startJob(url, id, rotation)
	if err != nil {
		return 0, err
	}
	result := <-jobs[*job.Job.Id].success
	delete(jobs, *job.Job.Id)
	if !result {
		return 0, fmt.Errorf("job failed")
	}
	outputFile, err := os.Create(file)
	if err != nil {
		return
	}
	defer outputFile.Close()
	key := *id + ".mp4"
	fileSize, err = downloader.Download(outputFile, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return
	}
	_, err = s3.New(sess).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to delete file from S3: %v", err)
	}

	elapsedTime := time.Since(startTime)
	log.Printf("Total time for downloadConverted: %s", elapsedTime)

	return
}
