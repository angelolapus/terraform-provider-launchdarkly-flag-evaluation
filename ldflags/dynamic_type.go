package launchdarkly_flag_eval

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type DynamicType struct{}

func (d DynamicType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.DynamicPseudoType
}

func (d DynamicType) ValueFromTerraform(_ context.Context, val tftypes.Value) (attr.Value, error) {
	vals := map[string]tftypes.Value{}
	err := val.As(&vals)
	if err != nil {
		return nil, err
	}
	return Dynamic{
		Values: vals,
	}, err
}

func (d DynamicType) Equal(other attr.Type) bool {
	_, ok := other.(DynamicType)
	if !ok {
		return false
	}
	return true
}

func (d DynamicType) String() string {
	return "DynamicPseudoType"
}

func (d DynamicType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("can't step into dynamic pseudo type")
}

func (d DynamicType) Validate(_ context.Context, val tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	if val.IsNull() {
		return nil
	}

	if !val.Type().Is(tftypes.Object{}) {
		return diag.Diagnostics{
			diag.NewAttributeErrorDiagnostic(path, "Invalid type", "Can only be an object"),
		}
	}
	return nil
}

type Dynamic struct {
	Values map[string]tftypes.Value
}

func (d Dynamic) Type(_ context.Context) attr.Type {
	return DynamicType{}
}

func (d Dynamic) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	types := map[string]tftypes.Type{}
	for key, val := range d.Values {
		types[key] = val.Type()
	}
	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: types,
	}, d.Values), nil
}

func (d Dynamic) Equal(other attr.Value) bool {
	o, ok := other.(Dynamic)
	if !ok {
		return false
	}
	if len(d.Values) != len(o.Values) {
		return false
	}
	for key, val := range d.Values {
		oVal, ok := o.Values[key]
		if !ok {
			return false
		}
		if !val.Equal(oVal) {
			return false
		}
	}
	return true
}
