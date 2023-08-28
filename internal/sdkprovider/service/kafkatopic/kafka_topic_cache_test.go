package kafkatopic

import (
	"reflect"
	"testing"

	"github.com/aiven/aiven-go-client"
)

func TestTopicCache_LoadByProjectAndServiceName(t1 *testing.T) {
	type args struct {
		projectName string
		serviceName string
	}
	tests := []struct {
		name        string
		doSomething func(*kafkaTopicCache)
		args        args
		want        map[string]aiven.KafkaTopic
		want1       bool
	}{
		{
			"not_found",
			func(*kafkaTopicCache) {
			},
			args{
				projectName: "test-pr1",
				serviceName: "test-sr1",
			},
			nil,
			false,
		},
		{
			"basic",
			testAddTwoTopicsToCache,
			args{
				projectName: "test-pr1",
				serviceName: "test-sr1",
			},
			map[string]aiven.KafkaTopic{
				"topic-1": {
					Replication: 3,
					State:       "AVAILABLE",
					TopicName:   "topic-1",
				},
				"topic-2": {
					Replication: 1,
					State:       "AVAILABLE",
					TopicName:   "topic-2",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t := newTopicCache()
		tt.doSomething(t)

		t1.Run(tt.name, func(t1 *testing.T) {
			got, got1 := t.LoadByProjectAndServiceName(tt.args.projectName, tt.args.serviceName)
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("LoadByProjectAndServiceName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t1.Errorf("LoadByProjectAndServiceName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTopicCache_LoadByTopicName(t1 *testing.T) {
	type args struct {
		projectName string
		serviceName string
		topicName   string
	}
	tests := []struct {
		name        string
		doSomething func(*kafkaTopicCache)
		args        args
		want        aiven.KafkaTopic
		want1       bool
	}{
		{
			"not_found",
			func(*kafkaTopicCache) {

			},
			args{
				projectName: "test-pr1",
				serviceName: "test-sr1",
				topicName:   "topic-1",
			},
			aiven.KafkaTopic{
				State: "CONFIGURING",
			},
			false,
		},
		{
			"basic",
			testAddTwoTopicsToCache,
			args{
				projectName: "test-pr1",
				serviceName: "test-sr1",
				topicName:   "topic-1",
			},
			aiven.KafkaTopic{
				Replication: 3,
				State:       "AVAILABLE",
				TopicName:   "topic-1",
			},
			true,
		},
	}
	for _, tt := range tests {
		t := newTopicCache()
		tt.doSomething(t)

		t1.Run(tt.name, func(t1 *testing.T) {
			got, got1 := t.LoadByTopicName(tt.args.projectName, tt.args.serviceName, tt.args.topicName)
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("LoadByTopicName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t1.Errorf("LoadByTopicName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTopicCache_DeleteByProjectAndServiceName(t1 *testing.T) {
	type args struct {
		projectName string
		serviceName string
	}
	tests := []struct {
		name        string
		doSomething func(*kafkaTopicCache)
		args        args
	}{
		{
			"basic",
			testAddTwoTopicsToCache,
			args{
				projectName: "test-pr1",
				serviceName: "test-sr1",
			},
		},
	}
	for _, tt := range tests {
		t := newTopicCache()
		tt.doSomething(t)

		t1.Run(tt.name, func(t1 *testing.T) {
			got, got1 := t.LoadByProjectAndServiceName(tt.args.projectName, tt.args.serviceName)
			if len(got) == 0 {
				t1.Errorf("LoadByProjectAndServiceName() got = %v", got)
			}
			if got1 != true {
				t1.Errorf("LoadByProjectAndServiceName() got1 = %v", got1)
			}

			t.DeleteByProjectAndServiceName(tt.args.projectName, tt.args.serviceName)

			got, got1 = t.LoadByProjectAndServiceName(tt.args.projectName, tt.args.serviceName)
			if len(got) != 0 {
				t1.Errorf("After deletion LoadByProjectAndServiceName() should be empty, got = %v", got)
			}
			if got1 != false {
				t1.Errorf("After deletion LoadByProjectAndServiceName() got1 whould be false = %v", got1)
			}
		})
	}
}

func testAddTwoTopicsToCache(c *kafkaTopicCache) {
	c.StoreByProjectAndServiceName(
		"test-pr1",
		"test-sr1",
		[]*aiven.KafkaTopic{
			{
				Replication: 3,
				State:       "AVAILABLE",
				TopicName:   "topic-1",
			},
			{
				Replication: 1,
				State:       "AVAILABLE",
				TopicName:   "topic-2",
			},
		})
}
