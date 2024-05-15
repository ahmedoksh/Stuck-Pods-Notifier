/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/slack-go/slack"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	monitorv1 "github.com/Stuck-Pods-Notifier/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const defaultRequeueDuration = 1 * time.Minute
const defaultPodWaitThreshold = 2 * time.Minute

// StuckPodsNotifierReconciler reconciles a StuckPodsNotifier object
type StuckPodsNotifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=monitor.my.domain,resources=stuckpodsnotifiers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.my.domain,resources=stuckpodsnotifiers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.my.domain,resources=stuckpodsnotifiers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StuckPodsNotifier object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *StuckPodsNotifierReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	stuckPodsNotifier := &monitorv1.StuckPodsNotifier{}
	err := r.Get(ctx, req.NamespacedName, stuckPodsNotifier)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("StuckPodsNotifier resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		} else {
			log.Error(err, "Failed to get StuckPodsNotifier")
			return reconcile.Result{}, err
		}
	}

	selector, requeueDuration, podWaitThreshold, slackToken, slackChannel := parseStuckPodsNotifierSpec(&stuckPodsNotifier.Spec, log)

	labelSelector, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		log.Error(err, "Failed to create label selector")
		return reconcile.Result{}, err
	}

	podList := &corev1.PodList{}
	err = r.List(ctx, podList, &client.ListOptions{Namespace: req.Namespace, LabelSelector: labelSelector})
	if err != nil {
		log.Error(err, "Failed to list pods")
		return reconcile.Result{}, err
	}

	slackApi := slack.New(slackToken)

	for _, pod := range podList.Items {
		timeSinceCreation := time.Since(pod.ObjectMeta.CreationTimestamp.Time).Round(time.Second)
		if pod.Status.Phase == corev1.PodPending && timeSinceCreation >= podWaitThreshold {
			msg := fmt.Sprintf(
				"Pod *%v* in namespace *%v* hasen't been scheduled for *%v*",
				pod.Name,
				pod.Namespace,
				timeSinceCreation,
			)
			_, _, err := slackApi.PostMessage(
				slackChannel,
				slack.MsgOptionText(
					msg,
					false,
				),
			)
			if err != nil {
				log.Error(err, "Failed to send Slack message")
				return reconcile.Result{}, err
			} else {
				log.Info("Slack message sent successfully:\n\t" + msg)
			}

		}
	}

	return reconcile.Result{RequeueAfter: requeueDuration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StuckPodsNotifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorv1.StuckPodsNotifier{}).
		Complete(r)
}

func parseStuckPodsNotifierSpec(spec *monitorv1.StuckPodsNotifierSpec, log logr.Logger) (metav1.LabelSelector, time.Duration, time.Duration, string, string) {
	var err error
	var requeueDuration time.Duration
	if spec.RequeueAfter != "" {
		requeueDuration, err = time.ParseDuration(spec.RequeueAfter)
		if err != nil {
			log.Error(err, fmt.Sprintf("Failed to parse requeue duration, using default duration %v", defaultRequeueDuration))
			requeueDuration = defaultRequeueDuration
		}
	} else {
		requeueDuration = defaultRequeueDuration
	}

	var podWaitThreshold time.Duration
	if spec.PodWaitThreshold != "" {
		podWaitThreshold, err = time.ParseDuration(spec.PodWaitThreshold)
		if err != nil {
			log.Error(err, fmt.Sprintf("Failed to parse pod wait threshold, using default duration %v", defaultPodWaitThreshold))
			podWaitThreshold = defaultPodWaitThreshold
		}
	} else {
		podWaitThreshold = defaultPodWaitThreshold
	}
	return spec.Selector, requeueDuration, podWaitThreshold, spec.SlackToken, spec.SlackChannel
}
