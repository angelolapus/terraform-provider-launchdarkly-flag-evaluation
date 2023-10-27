package launchdarkly_flag_eval

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ld "gopkg.in/launchdarkly/go-server-sdk.v5"
	"gopkg.in/launchdarkly/go-server-sdk.v5/ldcomponents"
)

const (
	LAUNCHDARKLY_SDK_KEY = "LAUNCHDARKLY_SDK_KEY"
	sdk_key              = "sdk_key"
	host                 = "host"
	flagKey              = "flag_key"
	userContext          = "context"
	variation            = "variation_type"
	value                = "value"
	defaultValue         = "default_value"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *ld.LDClient
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			sdk_key: {
				Type: types.StringType,
				// must be optional to allow setting via env variable
				Optional:    true,
				Sensitive:   true,
				Description: "The LaunchDarkly SDK key associated with the project and environment you would like to evaluate flags on",
			},
			host: {
				Type: types.StringType,
				// must be optional to allow setting via env variable
				Optional:    true,
				Sensitive:   true,
				Description: "The host to use for performing flag evaluations.",
			},
			// TODO add attributes:
			// polling/streaming
			// other things?
		},
	}, nil
}

type providerData struct {
	SDKKey types.String `tfsdk:"sdk_key"`
	Host   types.String `tfsdk:"host"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var sdkKey string
	if config.SDKKey.Unknown || config.SDKKey.Null {
		sdkKey = os.Getenv(LAUNCHDARKLY_SDK_KEY)
	} else {
		sdkKey = config.SDKKey.Value
	}

	if sdkKey == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"SDK key not found",
			"SDK Key must be provided",
		)
		return
	}

	var ldConfig ld.Config
	// TODO is there a shutdown call?
	ldConfig.Events = ldcomponents.SendEvents()
	// default poll interval is 30 seconds
	ldConfig.DataSource = ldcomponents.PollingDataSource()

	if config.Host.Value != "" {
		ldConfig.ServiceEndpoints.Polling = config.Host.Value
	}

	c, err := ld.MakeCustomClient(sdkKey, ldConfig, 5*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to initialize LaunchDarkly SDK client:\n\n"+err.Error(),
		)
		return
	}

	p.client = c
	p.configured = true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"ldflags_evaluation_boolean": dataSourceFlagEvaluationBooleanType{},
		"ldflags_evaluation_string":  dataSourceFlagEvaluationStringType{},
		"ldflags_evaluation_int":     dataSourceFlagEvaluationIntType{},
		"ldflags_evaluation_float":   dataSourceFlagEvaluationFloatType{},
		"ldflags_evaluation_json":    dataSourceFlagEvaluationJSONType{},
	}, nil
}
