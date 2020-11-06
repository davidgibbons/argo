package sensor

import (
	"context"
	"encoding/json"

	sv1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sensorpkg "github.com/argoproj/argo/pkg/apiclient/sensor"
	"github.com/argoproj/argo/server/auth"
	"github.com/argoproj/argo/util/logs"
)

type sensorServer struct{}

func (s *sensorServer) ListSensors(ctx context.Context, in *sensorpkg.ListSensorsRequest) (*sv1.SensorList, error) {
	client := auth.GetSensorClient(ctx)
	list, err := client.ArgoprojV1alpha1().Sensors(in.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *sensorServer) SensorsLogs(in *sensorpkg.SensorsLogsRequest, svr sensorpkg.SensorService_SensorsLogsServer) error {
	listOptions := metav1.ListOptions{LabelSelector: "sensor-name"}
	if in.Name != "" {
		listOptions.LabelSelector += "=" + in.Name
	}
	return logs.LogLabelledPods(
		svr.Context(),
		in.Namespace,
		listOptions,
		in.PodLogOptions,
		func(pod *corev1.Pod, data []byte) error {
			now := metav1.Now()
			e := &sensorpkg.LogEntry{
				Namespace:  pod.Namespace,
				SensorName: pod.Labels["sensor-name"],
				Level:      "info",
				Time:       &now,
				Msg:        string(data),
			}
			_ = json.Unmarshal(data, e)
			if in.TriggerName != "" && in.TriggerName != e.TriggerName {
				return nil
			}
			return svr.Send(e)
		},
	)
}

func NewSensorServer() sensorpkg.SensorServiceServer {
	return &sensorServer{}
}