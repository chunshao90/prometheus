package ceresdb

import (
	"fmt"

	"github.com/CeresDB/ceresdbproto/go/ceresdbproto"
	"github.com/pkg/errors"

	"github.com/prometheus/prometheus/model/labels"
)

const (
	// FieldName is use tod select field(column in table model)
	FieldName = "__ceresdb_field__"
	// PushdownName is used to disable pushdown for debug
	PushdownName = "__ceresdb_pushdown__"
	// DefaultField is default field for Prometheus
	DefaultField = "value"
)

func parseLiteralOrRegularExpr(x string) ([]string, bool) {
	// TODO: regex match is expensive, try convert some regexp(eg: a|b) to literal_or
	return nil, false
}

type QueryParam struct {
	Metric  string
	Field   string
	Filters []*ceresdbproto.Filter
}

func QueryParamFrom(matchers []*labels.Matcher) (QueryParam, error) {
	param := QueryParam{
		Metric:  "",
		Field:   DefaultField,
		Filters: make([]*ceresdbproto.Filter, 0, len(matchers)-1),
	}

	for _, m := range matchers {
		switch m.Name {
		case labels.MetricName:
			if m.Type == labels.MatchEqual {
				param.Metric = m.Value
			} else {
				return param, errors.Errorf("%s label must use equal match, current: %s", labels.MetricName, m.Type)
			}
		case FieldName:
			param.Field = m.Value
		default:
			filterParam := []string{m.Value}
			var filterType ceresdbproto.FilterType

			switch m.Type {
			case labels.MatchEqual:
				filterType = ceresdbproto.FilterType_LITERAL_OR
			case labels.MatchNotEqual:
				filterType = ceresdbproto.FilterType_NOT_LITERAL_OR
			case labels.MatchRegexp:
				if literals, ok := parseLiteralOrRegularExpr(m.Value); ok {
					filterParam = literals
					filterType = ceresdbproto.FilterType_LITERAL_OR
				} else {
					filterParam = []string{"^(?:" + m.Value + ")$"}
					filterType = ceresdbproto.FilterType_REGEXP
				}
			case labels.MatchNotRegexp:
				if literals, ok := parseLiteralOrRegularExpr(m.Value); ok {
					filterParam = literals
					filterType = ceresdbproto.FilterType_NOT_LITERAL_OR
				} else {
					filterParam = []string{"^(?:" + m.Value + ")$"}
					filterType = ceresdbproto.FilterType_NOT_REGEXP_MATCH
				}
			default:
				return param, fmt.Errorf("unknown match type %s", m.Type)
			}

			param.Filters = append(param.Filters, &ceresdbproto.Filter{
				TagKey: m.Name,
				Operators: []*ceresdbproto.FilterOperator{
					{
						FilterType: filterType,
						Params:     filterParam,
					},
				},
			})
		}
	}

	return param, nil
}
