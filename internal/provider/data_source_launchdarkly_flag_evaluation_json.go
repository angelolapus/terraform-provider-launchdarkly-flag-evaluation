package launchdarkly_flag_eval

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

type dataSourceFlagEvaluationJSONType struct{}

func (d dataSourceFlagEvaluationJSONType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return getFlagEvaluationSchemaForType(types.StringType)
}

func (d dataSourceFlagEvaluationJSONType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceFlagEvaluationJSON{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceFlagEvaluationJSON struct {
	p provider
}

func (d dataSourceFlagEvaluationJSON) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var dataSourceState struct {
		FlagKey      types.String `tfsdk:"flag_key"`
		DefaultValue types.String `tfsdk:"default_value"`
		Value        types.String `tfsdk:"value"`
		UserContext  LDUser       `tfsdk:"context"`
	}

	diags := req.Config.Get(ctx, &dataSourceState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	userCtx, _ := convertUserContextToLDUserContext(ctx, dataSourceState.UserContext.Key.Value, dataSourceState.UserContext, resp.Diagnostics)
	evaluation, err := d.p.client.JSONVariation(dataSourceState.FlagKey.Value, userCtx, ldvalue.Raw(json.RawMessage(dataSourceState.DefaultValue.Value)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Flag evaluation failed",
			"Could not evaluate flag: "+err.Error(),
		)
		return
	}
	d.p.client.Flush()

	dataSourceState.Value = types.String{
		Unknown: false,
		Null:    false,
		Value:   evaluation.JSONString(),
	}

	// set state
	diags = resp.State.Set(ctx, &dataSourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
