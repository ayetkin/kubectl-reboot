package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	CS     *kubernetes.Clientset
	logger *log.Logger
}

func New(kubeconfigPath, contextName string, logger *log.Logger) (*Client, error) {
	var restCfg *rest.Config
	var err error
	if kubeconfigPath != "" {
		abs := kubeconfigPath
		if !filepath.IsAbs(abs) {
			if abs, err = filepath.Abs(abs); err != nil {
				return nil, err
			}
		}
		restCfg, err = clientcmd.BuildConfigFromFlags("", abs)
	} else {
		kubeconfig := clientcmd.RecommendedHomeFile
		if _, statErr := os.Stat(kubeconfig); statErr == nil {
			restCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		} else {
			restCfg, err = rest.InClusterConfig()
		}
	}
	if err != nil {
		return nil, err
	}
	if contextName != "" && kubeconfigPath != "" {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
		cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{CurrentContext: contextName})
		restCfg, err = cc.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	return &Client{CS: cs, logger: logger}, nil
}

func (c *Client) ListNodeNames(excludeControlPlane bool) ([]string, error) {
	ctx := context.Background()
	opts := metav1.ListOptions{}
	if excludeControlPlane {
		opts.LabelSelector = "!node-role.kubernetes.io/control-plane,!node-role.kubernetes.io/master"
	}
	list, err := c.CS.CoreV1().Nodes().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(list.Items))
	for _, n := range list.Items {
		names = append(names, n.Name)
	}
	return names, nil
}

func (c *Client) Cordon(ctx context.Context, nodeName string) error {
	patch := []byte(`{"spec":{"unschedulable":true}}`)
	_, err := c.CS.CoreV1().Nodes().Patch(ctx, nodeName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *Client) Uncordon(ctx context.Context, nodeName string) error {
	patch := []byte(`{"spec":{"unschedulable":false}}`)
	_, err := c.CS.CoreV1().Nodes().Patch(ctx, nodeName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *Client) GetNode(ctx context.Context, nodeName string) (*corev1.Node, error) {
	return c.CS.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
}

func (c *Client) WaitForCondition(ctx context.Context, node string, pred func(*corev1.Node) bool, timeout, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		n, err := c.GetNode(ctx, node)
		if err == nil && pred(n) {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

func (c *Client) WaitForBootIDChange(ctx context.Context, node, before string, timeout, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		n, err := c.GetNode(ctx, node)
		if err == nil {
			if n.Status.NodeInfo.BootID != "" && n.Status.NodeInfo.BootID != before {
				return true
			}
		}
		time.Sleep(interval)
	}
	return false
}

func (c *Client) EvictPods(ctx context.Context, node string, pollInterval time.Duration, timeout time.Duration, dryRun bool) error {
	pods, err := c.CS.CoreV1().Pods("").List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.nodeName=%s", node)})
	if err != nil {
		return err
	}
	for _, p := range pods.Items {
		if isMirrorPod(&p) || hasOwnerKind(&p, "DaemonSet") || (p.Namespace == "kube-system" && hasCritical(&p)) || p.DeletionTimestamp != nil {
			continue
		}
		if dryRun {
			if c.logger != nil {
				c.logger.Info("🧪 DRY-RUN: Would evict pod", "namespace", p.Namespace, "pod", p.Name)
			}
			continue
		}
		eviction := &policyv1.Eviction{ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace}, DeleteOptions: &metav1.DeleteOptions{GracePeriodSeconds: int64Ptr(30)}}
		if err := c.CS.CoreV1().Pods(p.Namespace).EvictV1(ctx, eviction); err != nil {
			if c.logger != nil {
				c.logger.Warn("⚠️  Failed to evict pod", "namespace", p.Namespace, "pod", p.Name, "error", err)
			}
		} else {
			if c.logger != nil {
				c.logger.Info("🏃 Eviction sent for pod", "namespace", p.Namespace, "pod", p.Name)
			}
		}
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		left, err := c.CS.CoreV1().Pods("").List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.nodeName=%s", node)})
		if err != nil {
			return err
		}
		evictable := 0
		for _, p := range left.Items {
			if isMirrorPod(&p) || hasOwnerKind(&p, "DaemonSet") || (p.Namespace == "kube-system" && hasCritical(&p)) {
				continue
			}
			evictable++
		}
		if evictable == 0 {
			return nil
		}
		time.Sleep(pollInterval)
	}
	return fmt.Errorf("timeout waiting for pods eviction on %s", node)
}

func IsNodeReady(n *corev1.Node) bool {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isMirrorPod(p *corev1.Pod) bool { return p.Annotations[corev1.MirrorPodAnnotationKey] != "" }

func hasOwnerKind(p *corev1.Pod, kind string) bool {
	for _, o := range p.OwnerReferences {
		if o.Kind == kind {
			return true
		}
	}
	return false
}

func hasCritical(p *corev1.Pod) bool {
	if p.Annotations["kubernetes.io/config.mirror"] != "" {
		return true
	}
	if p.Annotations["scheduler.alpha.kubernetes.io/critical-pod"] != "" {
		return true
	}
	return false
}

func int64Ptr(i int64) *int64 { return &i }
