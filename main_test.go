package main

import (
	"encoding/json"
	"testing"

	alerts "github.com/opsgenie/opsgenie-go-sdk/alertsv2"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestGetNote(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	eventJSON, err := json.Marshal(event)
	assert.NoError(t, err)
	note, err := getNote(event)
	assert.NoError(t, err)
	assert.Contains(t, note, "Event data update:\n\n")
	assert.Contains(t, note, string(eventJSON))
}

func TestFormatTags(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	om := v2.NewObjectMeta("foo", "default")
	om.Labels = map[string]string{"foo": "bar"}
	event.Check.ObjectMeta = om
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	tags := formatTags(event)
	assert.Contains(t, tags, "foo:bar")
}

func TestParseEventKeyTags(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	om := v2.NewObjectMeta("foo", "default")
	om.Labels = map[string]string{"foo": "bar"}
	event.Check.ObjectMeta = om
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	plugin.MessageTemplate = "{{.Entity.Name}}/{{.Check.Name}}"
	plugin.MessageLimit = 100
	title, alias, tags := parseEventKeyTags(event)
	assert.Contains(t, title, "foo")
	assert.Contains(t, alias, "foo")
	assert.Contains(t, tags, "foo:bar")
}

func TestParseDescription(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	event.Check.Output = "Check OK"
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	plugin.DescriptionTemplate = "{{.Check.Output}}"
	plugin.DescriptionLimit = 100
	description := parseDescription(event)
	assert.Equal(t, description, "Check OK")
}

func TestParseDetails(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	event.Check.Output = "Check OK"
	event.Check.State = "passing"
	event.Check.Status = 0
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	det := parseDetails(event)
	assert.Equal(t, det["output"], "Check OK")
	assert.Equal(t, det["state"], "passing")
	assert.Equal(t, det["status"], "0")
}

func TestEventPriority(t *testing.T) {
	testcases := []struct {
		myPriority         string
		mismatchedPriority string
		alertsPriority     alerts.Priority
	}{
		{"P1", "P2", alerts.P1},
		{"P2", "P3", alerts.P2},
		{"P3", "P4", alerts.P3},
		{"P4", "P5", alerts.P4},
		{"P5", "P1", alerts.P5},
		{"Default", "P4", alerts.P3},
	}

	for _, tc := range testcases {
		assert := assert.New(t)
		event := types.FixtureEvent("foo", "bar")
		event.Check.Annotations["opsgenie_priority"] = tc.myPriority
		priority := eventPriority(event)
		expectedValue := tc.alertsPriority
		assert.Equal(priority, expectedValue)
	}

	// The FixtureEntity in FixtureEvent lacks Annotations, hand roll event
	// Set the check priority to mismatch, and ensure that entity priortity
	// takes precedence
	for _, tc := range testcases {
		assert := assert.New(t)
		event := types.Event{
			Entity: &types.Entity{
				ObjectMeta: types.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						"opsgenie_priority": tc.myPriority,
					},
				},
			},
			Check: &types.Check{
				ObjectMeta: types.ObjectMeta{
					Name:      "test-check",
					Namespace: "default",
					Annotations: map[string]string{
						"opsgenie_priority": tc.mismatchedPriority,
					},
				},
				Output: "test output",
			},
		}
		priority := eventPriority(&event)
		expectedValue := tc.alertsPriority
		assert.Equal(priority, expectedValue)
	}

	assert := assert.New(t)
	event := types.FixtureEvent("foo", "bar")
	priority := eventPriority(event)
	expectedValue := alerts.P3
	assert.Equal(priority, expectedValue)
}

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")
	assert.Error(checkArgs(event))
	plugin.AuthToken = "Testing"
	assert.NoError(checkArgs(event))
}

func TestStringInSlice(t *testing.T) {
	testSlice := []string{"foo", "bar", "test"}
	testString := "test"
	testResult := stringInSlice(testString, testSlice)
	assert.True(t, testResult)
	testString = "giraffe"
	testResult = stringInSlice(testString, testSlice)
	assert.False(t, testResult)
}

func TestTrim(t *testing.T) {
	testString := "This string is 33 characters long"
	assert.Equal(t, trim(testString, 40), testString)
	assert.Equal(t, trim(testString, 4), "This")
}
