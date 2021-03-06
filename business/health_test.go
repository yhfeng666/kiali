package business

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes/kubetest"
	"github.com/kiali/kiali/prometheus/prometheustest"
)

func TestGetServiceHealth(t *testing.T) {
	assert := assert.New(t)

	// Setup mocks
	k8s := new(kubetest.K8SClientMock)
	prom := new(prometheustest.PromClientMock)
	conf := config.NewConfig()
	config.Set(conf)
	hs := HealthService{k8s: k8s, prom: prom}

	queryTime := time.Date(2017, 01, 15, 0, 0, 0, 0, time.UTC)
	prom.MockServiceRequestRates("ns", "httpbin", serviceRates)

	health, _ := hs.GetServiceHealth("ns", "httpbin", "1m", queryTime)

	prom.AssertNumberOfCalls(t, "GetServiceRequestRates", 1)
	// 1.4 / 15.4 = 0.09
	assert.InDelta(float64(0.09), health.Requests.ErrorRatio, 0.01)
	assert.Equal(float64(1.4)/float64(15.4), health.Requests.InboundErrorRatio)
	assert.Equal(float64(-1), health.Requests.OutboundErrorRatio)
}

func TestGetAppHealth(t *testing.T) {
	assert := assert.New(t)

	// Setup mocks
	k8s := new(kubetest.K8SClientMock)
	prom := new(prometheustest.PromClientMock)
	conf := config.NewConfig()
	config.Set(conf)
	hs := HealthService{k8s: k8s, prom: prom}

	k8s.On("IsOpenShift").Return(true)
	k8s.MockEmptyWorkloads("ns")
	k8s.On("GetDeployments", "ns").Return(fakeDeploymentsHealthReview(), nil)
	k8s.On("GetPods", "ns", "app=reviews").Return(fakePodsHealthReview(), nil)

	queryTime := time.Date(2017, 01, 15, 0, 0, 0, 0, time.UTC)
	prom.MockAppRequestRates("ns", "reviews", otherRatesIn, otherRatesOut)

	health, _ := hs.GetAppHealth("ns", "reviews", "1m", queryTime)

	prom.AssertNumberOfCalls(t, "GetAppRequestRates", 1)
	// 1.6 / 6.6 = 0.24
	assert.Equal(float64((1.6+3.5)/(1.6+5+3.5)), health.Requests.ErrorRatio)
	assert.Equal(float64(1), health.Requests.InboundErrorRatio)
	assert.Equal(float64(3.5/(5+3.5)), health.Requests.OutboundErrorRatio)
}

func TestGetWorkloadHealth(t *testing.T) {
	assert := assert.New(t)

	// Setup mocks
	k8s := new(kubetest.K8SClientMock)
	prom := new(prometheustest.PromClientMock)
	conf := config.NewConfig()
	config.Set(conf)
	hs := HealthService{k8s: k8s, prom: prom}
	k8s.On("IsOpenShift").Return(true)
	k8s.MockEmptyWorkload("ns", "reviews-v1")
	k8s.On("GetDeployment", "ns", "reviews-v1").Return(&fakeDeploymentsHealthReview()[0], nil)
	k8s.On("GetPods", "ns", "").Return(fakePodsHealthReview(), nil)

	queryTime := time.Date(2017, 01, 15, 0, 0, 0, 0, time.UTC)
	prom.MockWorkloadRequestRates("ns", "reviews-v1", otherRatesIn, otherRatesOut)

	health, _ := hs.GetWorkloadHealth("ns", "reviews-v1", "1m", queryTime)

	k8s.AssertNumberOfCalls(t, "GetDeployment", 1)
	prom.AssertNumberOfCalls(t, "GetWorkloadRequestRates", 1)
	// 1.6 / 6.6 = 0.24
	assert.Equal(float64((1.6+3.5)/(1.6+5+3.5)), health.Requests.ErrorRatio)
	assert.Equal(float64(1), health.Requests.InboundErrorRatio)
	assert.Equal(float64(3.5/(5+3.5)), health.Requests.OutboundErrorRatio)
}

func TestGetAppHealthWithoutIstio(t *testing.T) {
	assert := assert.New(t)

	// Setup mocks
	k8s := new(kubetest.K8SClientMock)
	prom := new(prometheustest.PromClientMock)
	conf := config.NewConfig()
	config.Set(conf)
	hs := HealthService{k8s: k8s, prom: prom}

	k8s.On("IsOpenShift").Return(true)
	k8s.MockEmptyWorkloads("ns")
	k8s.On("GetDeployments", "ns").Return(fakeDeploymentsHealthReview(), nil)
	k8s.On("GetPods", "ns", "app=reviews").Return(fakePodsHealthReviewWithoutIstio(), nil)

	queryTime := time.Date(2017, 01, 15, 0, 0, 0, 0, time.UTC)
	prom.MockAppRequestRates("ns", "reviews", otherRatesIn, otherRatesOut)

	health, _ := hs.GetAppHealth("ns", "reviews", "1m", queryTime)

	prom.AssertNumberOfCalls(t, "GetAppRequestRates", 0)
	assert.Equal(float64(-1), health.Requests.ErrorRatio)
}

func TestGetWorkloadHealthWithoutIstio(t *testing.T) {
	assert := assert.New(t)

	// Setup mocks
	k8s := new(kubetest.K8SClientMock)
	prom := new(prometheustest.PromClientMock)
	conf := config.NewConfig()
	config.Set(conf)
	hs := HealthService{k8s: k8s, prom: prom}
	k8s.On("IsOpenShift").Return(true)
	k8s.MockEmptyWorkload("ns", "reviews-v1")
	k8s.On("GetDeployment", "ns", "reviews-v1").Return(&fakeDeploymentsHealthReview()[0], nil)
	k8s.On("GetPods", "ns", "").Return(fakePodsHealthReviewWithoutIstio(), nil)

	queryTime := time.Date(2017, 01, 15, 0, 0, 0, 0, time.UTC)
	prom.MockWorkloadRequestRates("ns", "reviews-v1", otherRatesIn, otherRatesOut)

	health, _ := hs.GetWorkloadHealth("ns", "reviews-v1", "1m", queryTime)

	prom.AssertNumberOfCalls(t, "GetWorkloadRequestRates", 0)
	assert.Equal(float64(-1), health.Requests.ErrorRatio)
}

func TestGetNamespaceAppHealthWithoutIstio(t *testing.T) {
	// Setup mocks
	k8s := new(kubetest.K8SClientMock)
	prom := new(prometheustest.PromClientMock)
	conf := config.NewConfig()
	config.Set(conf)
	hs := HealthService{k8s: k8s, prom: prom}

	k8s.On("IsOpenShift").Return(false)
	k8s.MockEmptyWorkloads("ns")
	k8s.On("GetServices", "ns", mock.AnythingOfType("map[string]string")).Return([]v1.Service{}, nil)
	k8s.On("GetDeployments", "ns").Return(fakeDeploymentsHealthReview(), nil)
	k8s.On("GetPods", "ns", "app").Return(fakePodsHealthReviewWithoutIstio(), nil)

	hs.GetNamespaceAppHealth("ns", "1m", time.Date(2017, 01, 15, 0, 0, 0, 0, time.UTC))

	// Make sure unnecessary call isn't performed
	prom.AssertNumberOfCalls(t, "GetAllRequestRates", 0)
}

var (
	sampleReviewsToHttpbin200 = model.Sample{
		Metric: model.Metric{
			"source_service":      "reviews.tutorial.svc.cluster.local",
			"destination_service": "httpbin.tutorial.svc.cluster.local",
			"response_code":       "200",
		},
		Value:     model.SampleValue(5),
		Timestamp: model.Now(),
	}
	sampleReviewsToHttpbin400 = model.Sample{
		Metric: model.Metric{
			"source_service":      "reviews.tutorial.svc.cluster.local",
			"destination_service": "httpbin.tutorial.svc.cluster.local",
			"response_code":       "400",
		},
		Value:     model.SampleValue(3.5),
		Timestamp: model.Now(),
	}
	sampleUnknownToHttpbin200 = model.Sample{
		Metric: model.Metric{
			"destination_service": "httpbin.tutorial.svc.cluster.local",
			"source_service":      "unknown",
			"response_code":       "200",
		},
		Value:     model.SampleValue(14),
		Timestamp: model.Now(),
	}
	sampleUnknownToHttpbin404 = model.Sample{
		Metric: model.Metric{
			"destination_service": "httpbin.tutorial.svc.cluster.local",
			"source_service":      "unknown",
			"response_code":       "404",
		},
		Value:     model.SampleValue(1.4),
		Timestamp: model.Now(),
	}
	sampleUnknownToReviews500 = model.Sample{
		Metric: model.Metric{
			"destination_service": "reviews.tutorial.svc.cluster.local",
			"source_service":      "unknown",
			"response_code":       "500",
		},
		Value:     model.SampleValue(1.6),
		Timestamp: model.Now(),
	}
	serviceRates = model.Vector{
		&sampleUnknownToHttpbin200,
		&sampleUnknownToHttpbin404,
	}
	otherRatesIn = model.Vector{
		&sampleUnknownToReviews500,
	}
	otherRatesOut = model.Vector{
		&sampleReviewsToHttpbin200,
		&sampleReviewsToHttpbin400,
	}
)

func fakeServicesHealthReview() []v1.Service {
	return []v1.Service{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "reviews",
				Namespace: "tutorial",
				Labels: map[string]string{
					"app":     "reviews",
					"version": "v1"}},
			Spec: v1.ServiceSpec{
				ClusterIP: "fromservice",
				Type:      "ClusterIP",
				Selector:  map[string]string{"app": "reviews"},
				Ports: []v1.ServicePort{
					{
						Name:     "http",
						Protocol: "TCP",
						Port:     3001},
					{
						Name:     "http",
						Protocol: "TCP",
						Port:     3000}}}}}
}

func fakePodsHealthReview() []v1.Pod {
	return []v1.Pod{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:        "reviews-v1",
				Labels:      map[string]string{"app": "reviews", "version": "v1"},
				Annotations: kubetest.FakeIstioAnnotations(),
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:        "reviews-v2",
				Labels:      map[string]string{"app": "reviews", "version": "v2"},
				Annotations: kubetest.FakeIstioAnnotations(),
			},
		},
	}
}

func fakePodsHealthReviewWithoutIstio() []v1.Pod {
	return []v1.Pod{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:   "reviews-v1",
				Labels: map[string]string{"app": "reviews", "version": "v1"},
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:   "reviews-v2",
				Labels: map[string]string{"app": "reviews", "version": "v2"},
			},
		},
	}
}

func fakeDeploymentsHealthReview() []v1beta1.Deployment {
	return []v1beta1.Deployment{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "reviews-v1"},
			Status: v1beta1.DeploymentStatus{
				Replicas:            3,
				AvailableReplicas:   2,
				UnavailableReplicas: 1},
			Spec: v1beta1.DeploymentSpec{
				Selector: &meta_v1.LabelSelector{
					MatchLabels: map[string]string{"app": "reviews", "version": "v1"}}}},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "reviews-v2"},
			Status: v1beta1.DeploymentStatus{
				Replicas:            2,
				AvailableReplicas:   1,
				UnavailableReplicas: 1},
			Spec: v1beta1.DeploymentSpec{
				Selector: &meta_v1.LabelSelector{
					MatchLabels: map[string]string{"app": "reviews", "version": "v2"}}}}}
}
