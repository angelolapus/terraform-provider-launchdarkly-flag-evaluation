package ldcomponents

import (
	"time"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	ldevents "gopkg.in/launchdarkly/go-sdk-events.v1"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces"
	"gopkg.in/launchdarkly/go-server-sdk.v5/internal"
	"gopkg.in/launchdarkly/go-server-sdk.v5/internal/endpoints"
)

const (
	// DefaultEventsBaseURI is the default value for EventProcessorBuilder.BaseURI.
	DefaultEventsBaseURI = "https://events.launchdarkly.com"
	// DefaultEventsCapacity is the default value for EventProcessorBuilder.Capacity.
	DefaultEventsCapacity = 10000
	// DefaultDiagnosticRecordingInterval is the default value for EventProcessorBuilder.DiagnosticRecordingInterval.
	DefaultDiagnosticRecordingInterval = 15 * time.Minute
	// DefaultFlushInterval is the default value for EventProcessorBuilder.FlushInterval.
	DefaultFlushInterval = 5 * time.Second
	// DefaultUserKeysCapacity is the default value for EventProcessorBuilder.UserKeysCapacity.
	DefaultUserKeysCapacity = 1000
	// DefaultUserKeysFlushInterval is the default value for EventProcessorBuilder.UserKeysFlushInterval.
	DefaultUserKeysFlushInterval = 5 * time.Minute
	// MinimumDiagnosticRecordingInterval is the minimum value for EventProcessorBuilder.DiagnosticRecordingInterval.
	MinimumDiagnosticRecordingInterval = 60 * time.Second
)

// EventProcessorBuilder provides methods for configuring analytics event behavior.
//
// See SendEvents for usage.
type EventProcessorBuilder struct {
	allAttributesPrivate        bool
	baseURI                     string
	capacity                    int
	diagnosticRecordingInterval time.Duration
	flushInterval               time.Duration
	inlineUsersInEvents         bool
	logUserKeyInErrors          bool
	privateAttributeNames       []lduser.UserAttribute
	userKeysCapacity            int
	userKeysFlushInterval       time.Duration
}

// SendEvents returns a configuration builder for analytics event delivery.
//
// The default configuration has events enabled with default settings. If you want to customize this
// behavior, call this method to obtain a builder, change its properties with the EventProcessorBuilder
// methods, and store it in Config.Events:
//
//     config := ld.Config{
//         Events: ldcomponents.SendEvents().Capacity(5000).FlushInterval(2 * time.Second),
//     }
//
// To disable analytics events, use NoEvents instead of SendEvents.
func SendEvents() *EventProcessorBuilder {
	return &EventProcessorBuilder{
		capacity:                    DefaultEventsCapacity,
		diagnosticRecordingInterval: DefaultDiagnosticRecordingInterval,
		flushInterval:               DefaultFlushInterval,
		userKeysCapacity:            DefaultUserKeysCapacity,
		userKeysFlushInterval:       DefaultUserKeysFlushInterval,
	}
}

// CreateEventProcessor is called by the SDK to create the event processor instance.
func (b *EventProcessorBuilder) CreateEventProcessor(
	context interfaces.ClientContext,
) (ldevents.EventProcessor, error) {
	loggers := context.GetLogging().GetLoggers()

	configuredBaseURI := endpoints.SelectBaseURI(
		context.GetBasic().ServiceEndpoints,
		endpoints.EventsService,
		b.baseURI,
		loggers,
	)

	eventSender := ldevents.NewServerSideEventSender(context.GetHTTP().CreateHTTPClient(),
		context.GetBasic().SDKKey, configuredBaseURI, context.GetHTTP().GetDefaultHeaders(), loggers)
	eventsConfig := ldevents.EventsConfiguration{
		AllAttributesPrivate:        b.allAttributesPrivate,
		Capacity:                    b.capacity,
		DiagnosticRecordingInterval: b.diagnosticRecordingInterval,
		EventSender:                 eventSender,
		FlushInterval:               b.flushInterval,
		InlineUsersInEvents:         b.inlineUsersInEvents,
		Loggers:                     loggers,
		LogUserKeyInErrors:          b.logUserKeyInErrors,
		PrivateAttributeNames:       b.privateAttributeNames,
		UserKeysCapacity:            b.userKeysCapacity,
		UserKeysFlushInterval:       b.userKeysFlushInterval,
	}
	if cci, ok := context.(*internal.ClientContextImpl); ok {
		eventsConfig.DiagnosticsManager = cci.DiagnosticsManager
	}
	return ldevents.NewDefaultEventProcessor(eventsConfig), nil
}

// AllAttributesPrivate sets whether or not all optional user attributes should be hidden from LaunchDarkly.
//
// If this is true, all user attribute values (other than the key) will be private, not just the attributes
// specified with PrivateAttributeNames or on a per-user basis with UserBuilder methods. By default, it is false.
func (b *EventProcessorBuilder) AllAttributesPrivate(value bool) *EventProcessorBuilder {
	b.allAttributesPrivate = value
	return b
}

// BaseURI is a deprecated method for setting a custom base URI for the events service.
//
// If you set this deprecated option to a non-empty value, it overrides any value that was set
// with ServiceEndpoints.
//
// Deprecated: Use config.ServiceEndpoints instead.
func (b *EventProcessorBuilder) BaseURI(baseURI string) *EventProcessorBuilder {
	b.baseURI = baseURI
	return b
}

// Capacity sets the capacity of the events buffer.
//
// The client buffers up to this many events in memory before flushing. If the capacity is exceeded before
// the buffer is flushed (see FlushInterval), events will be discarded. Increasing the capacity means that
// events are less likely to be discarded, at the cost of consuming more memory.
//
// The default value is DefaultEventsCapacity.
func (b *EventProcessorBuilder) Capacity(capacity int) *EventProcessorBuilder {
	b.capacity = capacity
	return b
}

// DiagnosticRecordingInterval sets the interval at which periodic diagnostic data is sent.
//
// The default value is DefaultDiagnosticRecordingInterval; the minimum value is MinimumDiagnosticRecordingInterval.
// This property is ignored if Config.DiagnosticOptOut is set to true.
func (b *EventProcessorBuilder) DiagnosticRecordingInterval(interval time.Duration) *EventProcessorBuilder {
	if interval < MinimumDiagnosticRecordingInterval {
		b.diagnosticRecordingInterval = MinimumDiagnosticRecordingInterval
	} else {
		b.diagnosticRecordingInterval = interval
	}
	return b
}

// FlushInterval sets the interval between flushes of the event buffer.
//
// Decreasing the flush interval means that the event buffer is less likely to reach capacity (see Capacity).
//
// The default value is DefaultFlushInterval.
func (b *EventProcessorBuilder) FlushInterval(interval time.Duration) *EventProcessorBuilder {
	b.flushInterval = interval
	return b
}

// InlineUsersInEvents sets whether to include full user details in every analytics event.
//
// The default is false: events will only include the user key, except for one "index" event that provides
// the full details for the user.
func (b *EventProcessorBuilder) InlineUsersInEvents(value bool) *EventProcessorBuilder {
	b.inlineUsersInEvents = value
	return b
}

// PrivateAttributeNames marks a set of attribute names as always private.
//
// Any users sent to LaunchDarkly with this configuration active will have attributes with these
// names removed. This is in addition to any attributes that were marked as private for an
// individual user with UserBuilder methods. Setting AllAttributePrivate to true overrides this.
//
//     config := ld.Config{
//         Events: ldcomponents.SendEvents().
//             PrivateAttributeNames(lduser.EmailAttribute, lduser.UserAttribute("some-custom-attribute")),
//     }
func (b *EventProcessorBuilder) PrivateAttributeNames(attributes ...lduser.UserAttribute) *EventProcessorBuilder {
	b.privateAttributeNames = attributes
	return b
}

// UserKeysCapacity sets the number of user keys that the event processor can remember at any one time.
//
// To avoid sending duplicate user details in analytics events, the SDK maintains a cache of recently
// seen user keys, expiring at an interval set by UserKeysFlushInterval.
//
// The default value is DefaultUserKeysCapacity.
func (b *EventProcessorBuilder) UserKeysCapacity(userKeysCapacity int) *EventProcessorBuilder {
	b.userKeysCapacity = userKeysCapacity
	return b
}

// UserKeysFlushInterval sets the interval at which the event processor will reset its cache of known user keys.
//
// The default value is DefaultUserKeysFlushInterval.
func (b *EventProcessorBuilder) UserKeysFlushInterval(interval time.Duration) *EventProcessorBuilder {
	b.userKeysFlushInterval = interval
	return b
}

// DescribeConfiguration is obsolete and is not called by the SDK.
//
// Deprecated: This method will be removed in a future major version release.
func (b *EventProcessorBuilder) DescribeConfiguration() ldvalue.Value {
	return ldvalue.Null()
}

// DescribeConfigurationContext is used internally by the SDK to inspect the configuration.
func (b *EventProcessorBuilder) DescribeConfigurationContext(context interfaces.ClientContext) ldvalue.Value {
	return ldvalue.ObjectBuild().
		Set("allAttributesPrivate", ldvalue.Bool(b.allAttributesPrivate)).
		Set("customEventsURI", ldvalue.Bool(
			endpoints.IsCustom(context.GetBasic().ServiceEndpoints, endpoints.EventsService, b.baseURI))).
		Set("diagnosticRecordingIntervalMillis", durationToMillisValue(b.diagnosticRecordingInterval)).
		Set("eventsCapacity", ldvalue.Int(b.capacity)).
		Set("eventsFlushIntervalMillis", durationToMillisValue(b.flushInterval)).
		Set("inlineUsersInEvents", ldvalue.Bool(b.inlineUsersInEvents)).
		Set("userKeysCapacity", ldvalue.Int(b.userKeysCapacity)).
		Set("userKeysFlushIntervalMillis", durationToMillisValue(b.userKeysFlushInterval)).
		Build()
}

func durationToMillisValue(d time.Duration) ldvalue.Value {
	return ldvalue.Float64(float64(uint64(d / time.Millisecond)))
}
