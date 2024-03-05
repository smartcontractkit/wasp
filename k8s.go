package wasp

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	K8sStatePollInterval = 3 * time.Second
)

// K8sClient high level k8s client
type K8sClient struct {
	ClientSet  *kubernetes.Clientset
	RESTConfig *rest.Config
}

// GetLocalK8sDeps get local k8s context config
func GetLocalK8sDeps() (*kubernetes.Clientset, *rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	k8sConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, err
	}
	return k8sClient, k8sConfig, nil
}

// NewK8sClient creates a new k8s client with a REST config
func NewK8sClient() *K8sClient {
	cs, cfg, err := GetLocalK8sDeps()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	return &K8sClient{
		ClientSet:  cs,
		RESTConfig: cfg,
	}
}

func (m *K8sClient) jobPods(ctx context.Context, nsName, syncLabel string) (*v1.PodList, error) {
	var pods *v1.PodList
	var lastError error
	maxRetries := 5               // Maximum number of retries
	retryDelay := 2 * time.Second // Constant delay interval

	retryPolicy := wait.Backoff{
		Steps:    maxRetries, // Max retry attempts
		Duration: retryDelay, // Constant delay interval
		Factor:   1.0,        // No increase in delay, making it constant
		Jitter:   0,          // No jitter
	}

	err := wait.ExponentialBackoff(retryPolicy, func() (bool, error) {
		timeout := int64(30)
		pods, lastError = m.ClientSet.CoreV1().Pods(nsName).List(ctx, metaV1.ListOptions{
			LabelSelector:  syncSelector(syncLabel), // Assuming syncSelector is a function that formats the label selector
			TimeoutSeconds: &timeout,
		})
		if lastError != nil {
			log.Warn().Msgf("Error retrieving pods, will retry: %v\n", lastError)
			return false, nil // Return false to trigger a retry
		}
		return true, nil // Success, stop retrying
	})

	if err != nil {
		// Handle the case where retries are exhausted
		return nil, fmt.Errorf("after %d attempts, last error: %s", maxRetries, lastError)
	}

	return pods, nil
}

func (m *K8sClient) jobs(ctx context.Context, nsName, syncLabel string) (*batchV1.JobList, error) {
	var jobs *batchV1.JobList
	var lastError error
	maxRetries := 5               // Maximum number of retries
	retryDelay := 2 * time.Second // Constant delay interval

	retryPolicy := wait.Backoff{
		Steps:    maxRetries, // Max retry attempts
		Duration: retryDelay, // Constant delay interval
		Factor:   1.0,        // No increase in delay, making it constant
		Jitter:   0,          // No jitter
	}

	err := wait.ExponentialBackoff(retryPolicy, func() (bool, error) {
		timeout := int64(30)
		jobs, lastError = m.ClientSet.BatchV1().Jobs(nsName).List(ctx, metaV1.ListOptions{
			LabelSelector:  syncSelector(syncLabel), // Assuming syncSelector is a function that formats the label selector
			TimeoutSeconds: &timeout,
		})
		if lastError != nil {
			log.Warn().Msgf("Error retrieving jobs, will retry: %v", lastError)
			return false, nil // Return false to trigger a retry
		}
		return true, nil // Success, stop retrying
	})

	if err != nil {
		// Handle the case where retries are exhausted
		return nil, fmt.Errorf("after %d attempts, last error: %s", maxRetries, lastError)
	}

	return jobs, nil
}

func syncSelector(s string) string {
	return fmt.Sprintf("sync=%s", s)
}

func (m *K8sClient) removeJobs(ctx context.Context, nsName string, jobs *batchV1.JobList) error {
	log.Info().Msg("Removing jobs")
	for _, j := range jobs.Items {
		dp := metaV1.DeletePropagationForeground
		if err := m.ClientSet.BatchV1().Jobs(nsName).Delete(ctx, j.Name, metaV1.DeleteOptions{
			PropagationPolicy: &dp,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *K8sClient) waitSyncGroup(ctx context.Context, nsName string, syncLabel string, jobNum int) error {
outer:
	for {
		time.Sleep(K8sStatePollInterval)
		log.Info().Str("SyncLabel", syncLabel).Msg("Awaiting group sync")
		pods, err := m.jobPods(ctx, nsName, syncLabel)
		if err != nil {
			return err
		}
		if len(pods.Items) != jobNum {
			log.Info().Str("SyncLabel", syncLabel).Msg("Awaiting pods")
			continue
		}
		for _, p := range pods.Items {
			if p.Status.Phase != v1.PodRunning {
				continue outer
			}
		}
		return nil
	}
}

// TrackJobs tracks both jobs and their pods until they succeed or fail
func (m *K8sClient) TrackJobs(ctx context.Context, nsName, syncLabel string, jobNum int, keepJobs bool) error {
	log.Debug().Str("LabelSelector", syncSelector(syncLabel)).Msg("Searching for jobs/pods")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Cluster context finished")
			return nil
		default:
			time.Sleep(K8sStatePollInterval)
			jobs, err := m.jobs(ctx, nsName, syncLabel)
			if err != nil {
				return errors.Wrapf(err, "failed to get jobs")
			}
			jobPods, err := m.jobPods(ctx, nsName, syncLabel)
			if err != nil {
				return errors.Wrapf(err, "failed to get job pods")
			}
			if len(jobPods.Items) != jobNum {
				log.Info().Int("JobPods", jobNum).Msg("Awaiting job pods")
				continue
			}
			for _, jp := range jobPods.Items {
				log.Debug().Interface("Phase", jp.Status.Phase).Msg("Job status")
			}
			var successfulJobs int
			for _, j := range jobs.Items {
				log.Debug().Interface("Status", j.Status).Str("Name", j.Name).Msg("Pod status")
				if j.Status.Failed > 0 {
					log.Warn().Str("Name", j.Name).Msg("Job has failed")
					if !keepJobs {
						if err := m.removeJobs(ctx, nsName, jobs); err != nil {
							return err
						}
					}
					return fmt.Errorf("job %s has failed", j.Name)
				}
				if j.Status.Succeeded > 0 {
					successfulJobs += 1
				}
			}
			if successfulJobs == jobNum {
				log.Info().Msg("Test ended")
				if !keepJobs {
					return m.removeJobs(ctx, nsName, jobs)
				}
				return nil
			}
		}
	}
}
