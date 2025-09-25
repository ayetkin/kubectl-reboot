package kube

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsNodeReady(t *testing.T) {
	tests := []struct {
		name       string
		conditions []corev1.NodeCondition
		expected   bool
	}{
		{
			name: "ready node",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
			expected: true,
		},
		{
			name: "not ready node",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			},
			expected: false,
		},
		{
			name: "unknown ready status",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionUnknown,
				},
			},
			expected: false,
		},
		{
			name: "no ready condition",
			conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionFalse,
				},
			},
			expected: false,
		},
		{
			name:       "empty conditions",
			conditions: []corev1.NodeCondition{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: tt.conditions,
				},
			}
			result := IsNodeReady(node)
			if result != tt.expected {
				t.Errorf("IsNodeReady() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsMirrorPod(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name: "mirror pod",
			annotations: map[string]string{
				corev1.MirrorPodAnnotationKey: "mirror-host",
			},
			expected: true,
		},
		{
			name: "regular pod",
			annotations: map[string]string{
				"app": "test",
			},
			expected: false,
		},
		{
			name:        "no annotations",
			annotations: nil,
			expected:    false,
		},
		{
			name:        "empty annotations",
			annotations: map[string]string{},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}
			result := isMirrorPod(pod)
			if result != tt.expected {
				t.Errorf("isMirrorPod() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasOwnerKind(t *testing.T) {
	tests := []struct {
		name            string
		ownerReferences []metav1.OwnerReference
		kind            string
		expected        bool
	}{
		{
			name: "has daemonset owner",
			ownerReferences: []metav1.OwnerReference{
				{
					Kind: "DaemonSet",
					Name: "test-ds",
				},
			},
			kind:     "DaemonSet",
			expected: true,
		},
		{
			name: "has deployment owner",
			ownerReferences: []metav1.OwnerReference{
				{
					Kind: "ReplicaSet",
					Name: "test-rs",
				},
			},
			kind:     "DaemonSet",
			expected: false,
		},
		{
			name: "multiple owners, has target kind",
			ownerReferences: []metav1.OwnerReference{
				{
					Kind: "ReplicaSet",
					Name: "test-rs",
				},
				{
					Kind: "DaemonSet",
					Name: "test-ds",
				},
			},
			kind:     "DaemonSet",
			expected: true,
		},
		{
			name:            "no owners",
			ownerReferences: []metav1.OwnerReference{},
			kind:            "DaemonSet",
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: tt.ownerReferences,
				},
			}
			result := hasOwnerKind(pod, tt.kind)
			if result != tt.expected {
				t.Errorf("hasOwnerKind() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasCritical(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name: "config mirror annotation",
			annotations: map[string]string{
				"kubernetes.io/config.mirror": "true",
			},
			expected: true,
		},
		{
			name: "critical pod annotation",
			annotations: map[string]string{
				"scheduler.alpha.kubernetes.io/critical-pod": "true",
			},
			expected: true,
		},
		{
			name: "both critical annotations",
			annotations: map[string]string{
				"kubernetes.io/config.mirror":                "true",
				"scheduler.alpha.kubernetes.io/critical-pod": "true",
			},
			expected: true,
		},
		{
			name: "regular pod",
			annotations: map[string]string{
				"app": "test",
			},
			expected: false,
		},
		{
			name:        "no annotations",
			annotations: nil,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}
			result := hasCritical(pod)
			if result != tt.expected {
				t.Errorf("hasCritical() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldSkipPod(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
		reason   string
	}{
		{
			name: "mirror pod should be skipped",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						corev1.MirrorPodAnnotationKey: "mirror-host",
					},
				},
			},
			expected: true,
			reason:   "mirror pod",
		},
		{
			name: "daemonset pod should be skipped",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: "DaemonSet", Name: "test-ds"},
					},
				},
			},
			expected: true,
			reason:   "daemonset pod",
		},
		{
			name: "critical kube-system pod should be skipped",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Annotations: map[string]string{
						"kubernetes.io/config.mirror": "true",
					},
				},
			},
			expected: true,
			reason:   "critical kube-system pod",
		},
		{
			name: "pod with deletion timestamp should be skipped",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{},
				},
			},
			expected: true,
			reason:   "already being deleted",
		},
		{
			name: "regular pod should not be skipped",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test-pod",
				},
			},
			expected: false,
			reason:   "regular pod",
		},
		{
			name: "non-critical kube-system pod should not be skipped",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "non-critical-pod",
				},
			},
			expected: false,
			reason:   "non-critical kube-system pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldSkipPod(tt.pod)
			if result != tt.expected {
				t.Errorf("shouldSkipPod() = %v, want %v for %s", result, tt.expected, tt.reason)
			}
		})
	}
}

func TestCountEvictablePods(t *testing.T) {
	client := &Client{}

	pods := []corev1.Pod{
		// Regular pod - evictable
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "regular-pod",
				Namespace: "default",
			},
		},
		// Mirror pod - not evictable
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mirror-pod",
				Namespace: "kube-system",
				Annotations: map[string]string{
					corev1.MirrorPodAnnotationKey: "mirror-host",
				},
			},
		},
		// DaemonSet pod - not evictable
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds-pod",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "DaemonSet", Name: "test-ds"},
				},
			},
		},
		// Another regular pod - evictable
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "another-pod",
				Namespace: "app",
			},
		},
	}

	result := client.countEvictablePods(pods)
	expected := 2 // Only the two regular pods should be evictable

	if result != expected {
		t.Errorf("countEvictablePods() = %d, want %d", result, expected)
	}
}
