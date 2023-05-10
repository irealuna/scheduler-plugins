package annotationmatch

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// 定义plugin struct
type AnnotationMatch struct {
	handle framework.Handle
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	klog.V(3).Infof("create annotationmatch plugin")
	return &AnnotationMatch{handle: h}, nil
}

// 用来保证AnnotationMatch实现了FilterPlugin的所有接口
var _ = framework.FilterPlugin(&AnnotationMatch{})

const Name = "AnnotationMatch"
const targetAnnotation = "annotation/annotationmatch"

// plugin注册和配置时使用的Name
func (am *AnnotationMatch) Name() string {
	return Name
}

func (am *AnnotationMatch) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {

	if pod == nil {
		return framework.NewStatus(framework.Error, "pod is nil")
	}
	klog.InfoS("annotationmatch plugin get pod", "pod", pod.ObjectMeta.Name)

	// 检查pod是否有c
	var targetValue string = ""
	for key, value := range pod.ObjectMeta.Annotations {
		if key == targetAnnotation {
			targetValue = value
		}
	}
	// 如果没有检查pod是否有targetAnnotation则直接返回success
	if len(targetValue) == 0 {
		return framework.NewStatus(framework.Success)
	}

	// 如果有则判断pod的annotation同node annotation是否相同，如果相同则返回success，否则返回UnschedulableAndUnresolvable
	node := nodeInfo.Node()
	if node == nil {
		return framework.NewStatus(framework.Error, "node not found")
	}
	for key, value := range node.ObjectMeta.Annotations {
		if key == targetAnnotation && value == targetValue {
			return framework.NewStatus(framework.Success)
		}
	}
	return framework.NewStatus(framework.UnschedulableAndUnresolvable, "annotation not match")
}
