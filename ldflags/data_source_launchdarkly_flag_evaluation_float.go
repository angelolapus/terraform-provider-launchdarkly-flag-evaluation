package launchdarkly_flag_eval

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceFlagEvaluationFloatType struct{}

func (d dataSourceFlagEvaluationFloatType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return getFlagEvaluationSchemaForType(types.Float64Type)
}

func (d dataSourceFlagEvaluationFloatType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceFlagEvaluationFloat{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceFlagEvaluationFloat struct {
	p provider
}

func (d dataSourceFlagEvaluationFloat) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var dataSourceState struct {
		FlagKey      types.String  `tfsdk:"flag_key"`
		DefaultValue types.Float64 `tfsdk:"default_value"`
		Value        types.Float64 `tfsdk:"value"`
		UserContext  LDUser        `tfsdk:"context"`
	}

	diags := req.Config.Get(ctx, &dataSourceState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	userCtx, _ := convertUserContextToLDUserContext(ctx, dataSourceState.UserContext.Key.Value, dataSourceState.UserContext, resp.Diagnostics)
	evaluation, err := d.p.client.Float64Variation(dataSourceState.FlagKey.Value, userCtx, dataSourceState.DefaultValue.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Flag evaluation failed",
			"Could not evaluate flag: "+err.Error(),
		)
		return
	}
	d.p.client.Flush()

	dataSourceState.Value = types.Float64{
		Unknown: false,
		Null:    false,
		Value:   evaluation,
	}

	// set state
	diags = resp.State.Set(ctx, &dataSourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
