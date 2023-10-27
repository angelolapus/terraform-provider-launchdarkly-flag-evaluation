package launchdarkly_flag_eval

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceFlagEvaluationIntType struct{}

func (d dataSourceFlagEvaluationIntType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return getFlagEvaluationSchemaForType(types.Int64Type)
}

func (d dataSourceFlagEvaluationIntType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceFlagEvaluationInt{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceFlagEvaluationInt struct {
	p provider
}

func (d dataSourceFlagEvaluationInt) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var dataSourceState struct {
		FlagKey      types.String `tfsdk:"flag_key"`
		DefaultValue types.Int64  `tfsdk:"default_value"`
		Value        types.Int64  `tfsdk:"value"`
		UserContext  LDUser       `tfsdk:"context"`
	}

	diags := req.Config.Get(ctx, &dataSourceState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	userCtx, _ := convertUserContextToLDUserContext(ctx, dataSourceState.UserContext.Key.Value, dataSourceState.UserContext, resp.Diagnostics)
	evaluation, err := d.p.client.IntVariation(dataSourceState.FlagKey.Value, userCtx, int(dataSourceState.DefaultValue.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"Flag evaluation failed",
			"Could not evaluate flag: "+err.Error(),
		)
		return
	}
	d.p.client.Flush()

	dataSourceState.Value = types.Int64{
		Unknown: false,
		Null:    false,
		Value:   int64(evaluation),
	}

	// set state
	diags = resp.State.Set(ctx, &dataSourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
