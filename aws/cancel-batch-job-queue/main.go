package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"log"
	"sync"
	"time"
)

const (
	region       = ""
	jobQueueName = ""
	jobStatue    = "RUNNABLE"
)

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("failed to create session : %v", err)
	}

	svc := batch.New(sess)

	listJobsInput := &batch.ListJobsInput{
		JobQueue:  aws.String(jobQueueName),
		JobStatus: aws.String(jobStatue),
	}

	result, err := svc.ListJobs(listJobsInput)
	if err != nil {
		log.Fatalf("failed to list jobs: %v", err)
	}

	for len(result.JobSummaryList) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(result.JobSummaryList))

		for _, job := range result.JobSummaryList {
			// to avoid request limit
			time.Sleep(time.Millisecond * 50)
			go func(jobId *string) {
				cancelInput := &batch.CancelJobInput{
					JobId:  jobId,
					Reason: aws.String("Delete all jobs"),
				}
				_, err := svc.CancelJob(cancelInput)
				if err != nil {
					log.Printf("failed to cancel job %s: %v", *jobId, err)
				} else {
					log.Printf("Job %s has been canceled", *jobId)
				}
				wg.Done()
			}(job.JobId)
		}
		wg.Wait()

		result, err = svc.ListJobs(listJobsInput)
		if err != nil {
			log.Fatalf("failed to list jobs: %v", err)
		}
	}

	log.Println("All pending jobs have been canceled")
}
