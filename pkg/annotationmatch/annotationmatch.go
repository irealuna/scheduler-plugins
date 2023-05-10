package annotationmatch

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// 1. 定义一个插件结构体
type AnnotationMatch struct {
	handle framework.Handle
}

// 用来保证AnnotationMatch实现了FilterPlugin的所有接口
var _ = framework.FilterPlugin(&AnnotationMatch{})

const Name = "AnnotationMatch"
const targetAnnotation = "annotation/annotationmatch"

// 2. 实现 Plugin 插件，即实现 Name 方法
func (am *AnnotationMatch) Name() string {
	return Name
}

// 3. 实现 Filter 函数
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

// 4. 实现 New 函数，返回该自定义插件对象
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	klog.V(3).Infof("create annotationmatch plugin")
	return &AnnotationMatch{handle: h}, nil
}
