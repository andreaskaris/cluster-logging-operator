package cloudwatch

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	. "github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/source"
	genhelper "github.com/openshift/cluster-logging-operator/pkg/generator/helpers"
	corev1 "k8s.io/api/core/v1"
)

type CloudWatch struct {
	Region         string
	SecurityConfig Element
}

func (cw CloudWatch) Name() string {
	return "cloudwatchTemplate"
}

func (cw CloudWatch) Template() string {
	return `{{define "` + cw.Name() + `" -}}
@type cloudwatch_logs
auto_create_stream true
region {{.Region }}
log_group_name_key cw_group_name
log_stream_name_key cw_stream_name
remove_log_stream_name_key true
remove_log_group_name_key true
auto_create_stream true
concurrency 2
{{compose_one .SecurityConfig}}
include_time_key true
log_rejected_request true
{{end}}`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	logGroupPrefix := LogGroupPrefix(o)
	logGroupName := LogGroupName(o)
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				GroupNameStreamName(fmt.Sprintf("%s%s", logGroupPrefix, logGroupName),
					"${tag}",
					source.ApplicationTags),
				GroupNameStreamName(fmt.Sprintf("%sinfrastructure", logGroupPrefix),
					"${record['hostname']}.${tag}",
					source.InfraTags),
				GroupNameStreamName(fmt.Sprintf("%saudit", logGroupPrefix),
					"${record['hostname']}.${tag}",
					source.AuditTags),
				OutputConf(bufspec, secret, o, op),
			},
		},
	}
}

func OutputConf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	return Match{
		MatchTags: "**",
		MatchElement: CloudWatch{
			Region:         o.Cloudwatch.Region,
			SecurityConfig: SecurityConfig(o, secret),
		},
	}
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) Element {
	return AWSKey{
		KeyIDPath: security.SecretPath(o.Secret.Name, "aws_access_key_id"),
		KeyPath:   security.SecretPath(o.Secret.Name, "aws_secret_access_key"),
	}
}

func GroupNameStreamName(groupName, streamName, tag string) Element {
	return Filter{
		MatchTags: tag,
		Element: RecordModifier{
			Records: []Record{
				{
					Key:        "cw_group_name",
					Expression: groupName,
				},
				{
					Key:        "cw_stream_name",
					Expression: streamName,
				},
			},
		},
	}
}

func LogGroupPrefix(o logging.OutputSpec) string {
	if o.Cloudwatch != nil {
		prefix := o.Cloudwatch.GroupPrefix
		if prefix != nil && strings.TrimSpace(*prefix) != "" {
			return fmt.Sprintf("%s.", *prefix)
		}
	}
	return ""
}

func LogGroupName(o logging.OutputSpec) string {
	if o.Cloudwatch != nil {
		switch o.Cloudwatch.GroupBy {
		case logging.LogGroupByNamespaceName:
			return "${record['kubernetes']['namespace_name']}"
		case logging.LogGroupByNamespaceUUID:
			return "${record['kubernetes']['namespace_id']}"
		default:
			return logging.InputNameApplication
		}
	}
	return ""
}
