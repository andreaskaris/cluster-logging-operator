package client

import (
	"fmt"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	validationFailureMsg = "is dependent on a ClusterLogging instance with a valid Collector configuration"
)

func ClusterLogForwarderReady(e watch.Event) (bool, error) {
	clf := e.Object.(*loggingv1.ClusterLogForwarder)
	cond := clf.Status.Conditions
	switch {
	case cond.IsTrueFor(loggingv1.ConditionReady):
		return true, nil
	case cond.IsFalseFor(loggingv1.ConditionReady), cond.IsTrueFor(loggingv1.ConditionDegraded):
		return false, fmt.Errorf("ClusterLogForwarder unexpected condition: %v", test.YAMLString(clf.Status))
	default:
		return false, nil
	}
}

// ClusterLogForwarderValidationFailure expects condition type "Validation" to be set on the ClusterLogForwarder
// resource. If no such condition can be found, it returns false, and a nil error (so that c.WaitFor can wait until
// the condition is set, or time out if the condition is never set). If the condition is set, we expect its message
// to match validationFailureMsg and we expect it to be "True". We also expect the "Ready" condition to be "False".
// In that case, we return true and no error. In case of the contrary, we return false and an error.
func ClusterLogForwarderValidationFailure(e watch.Event) (bool, error) {
	clf := e.Object.(*loggingv1.ClusterLogForwarder)
	cond := clf.Status.Conditions

	validationCondition := cond.GetCondition(loggingv1.ValidationCondition)
	if validationCondition == nil {
		return false, nil
	}

	if strings.Contains(validationCondition.Message, validationFailureMsg) &&
		validationCondition.Status == v1.ConditionTrue && cond.IsFalseFor(loggingv1.ConditionReady) {
		return true, nil
	}
	return false, fmt.Errorf("ClusterLogForwarder unexpected condition: %v", test.YAMLString(clf.Status))
}
