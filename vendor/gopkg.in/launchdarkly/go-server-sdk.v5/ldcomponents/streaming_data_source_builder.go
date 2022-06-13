package ldcomponents

import (
	"time"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces"
	"gopkg.in/launchdarkly/go-server-sdk.v5/internal/datasource"
	"gopkg.in/launchdarkly/go-server-sdk.v5/internal/endpoints"
)

// DefaultStreamingBaseURI is the default value for StreamingDataSourceBuilder.BaseURI.
const DefaultStreamingBaseURI = endpoints.DefaultStreamingBaseURI

// DefaultInitialReconnectDelay is the default value for StreamingDataSourceBuilder.InitialReconnectDelay.
const DefaultInitialReconnectDelay = time.Second

// StreamingDataSourceBuilder provides methods for configuring the streaming data source.
//
// See StreamingDataSource for usage.
type StreamingDataSourceBuilder struct {
	baseURI               string
	initialReconnectDelay time.Duration
}

// StreamingDataSource returns a configurable factory for using streaming mode to get feature flag data.
//
// By default, the SDK uses a streaming connection to receive feature flag data from LaunchDarkly. To use the
// default behavior, you do not need to call this method. However, if you want to customize the behavior of
// the connection, call this method to obtain a builder, set its properties with the StreamingDataSourceBuilder
// methods, and then store it in the DataSource field of your SDK configuration:
//
//     config := ld.Config{
//         DataSource: ldcomponents.StreamingDataSource().InitialReconnectDelay(500 * time.Millisecond),
//     }
func StreamingDataSource() *StreamingDataSourceBuilder {
	return &StreamingDataSourceBuilder{
		initialReconnectDelay: DefaultInitialReconnectDelay,
	}
}

// BaseURI is a deprecated method for setting a custom base URI for the polling service.
//
// If you set this deprecated option to a non-empty value, it overrides any value that was set
// with ServiceEndpoints.
//
// Deprecated: Use config.ServiceEndpoints instead.
func (b *StreamingDataSourceBuilder) BaseURI(baseURI string) *StreamingDataSourceBuilder {
	b.baseURI = baseURI
	return b
}

// InitialReconnectDelay sets the initial reconnect delay for the streaming connection.
//
// The streaming service uses a backoff algorithm (with jitter) every time the connection needs to be
// reestablished. The delay for the first reconnection will start near this value, and then increase
// exponentially for any subsequent connection failures.
//
// The default value is DefaultInitialReconnectDelay.
func (b *StreamingDataSourceBuilder) InitialReconnectDelay(
	initialReconnectDelay time.Duration,
) *StreamingDataSourceBuilder {
	if initialReconnectDelay <= 0 {
		b.initialReconnectDelay = DefaultInitialReconnectDelay
	} else {
		b.initialReconnectDelay = initialReconnectDelay
	}
	return b
}

// CreateDataSource is called by the SDK to create the data source instance.
func (b *StreamingDataSourceBuilder) CreateDataSource(
	context interfaces.ClientContext,
	dataSourceUpdates interfaces.DataSourceUpdates,
) (interfaces.DataSource, error) {
	configuredBaseURI := endpoints.SelectBaseURI(
		context.GetBasic().ServiceEndpoints,
		endpoints.StreamingService,
		b.baseURI,
		context.GetLogging().GetLoggers(),
	)

	return datasource.NewStreamProcessor(
		context,
		dataSourceUpdates,
		configuredBaseURI,
		b.initialReconnectDelay,
	), nil
}

// DescribeConfiguration is obsolete and is not called by the SDK.
//
// Deprecated: This method will be removed in a future major version release.
func (b *StreamingDataSourceBuilder) DescribeConfiguration() ldvalue.Value {
	return ldvalue.Null()
}

// DescribeConfigurationContext is used internally by the SDK to inspect the configuration.
func (b *StreamingDataSourceBuilder) DescribeConfigurationContext(context interfaces.ClientContext) ldvalue.Value {
	return ldvalue.ObjectBuild().
		Set("streamingDisabled", ldvalue.Bool(false)).
		Set("customStreamURI", ldvalue.Bool(
			endpoints.IsCustom(context.GetBasic().ServiceEndpoints, endpoints.StreamingService, b.baseURI))).
		Set("reconnectTimeMillis", durationToMillisValue(b.initialReconnectDelay)).
		Set("usingRelayDaemon", ldvalue.Bool(false)).
		Build()
}
