package scutils

import (
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
)

func toPointer[T any](v T) *T {
	return &v
}

// A holder type for higher level security hub properties
type SecurityHubConfig struct {
	AccountID   string
	ProductName string
	CompanyName string
	ProductARN  string
	Region      string
}

func NewSecurityHubConfig(accID, productName, companyName, productARN string) SecurityHubConfig {
	return SecurityHubConfig{
		AccountID:   accID,
		ProductName: productName,
		CompanyName: companyName,
		ProductARN:  productARN,
	}
}

// Turn the security hub finding to a searchable filter in SH
func ToResourceIDFilter(findings []types.AwsSecurityFinding) []types.StringFilter {
	list := map[string]bool{}
	for _, f := range findings {
		list[*f.Resources[0].Id] = true
	}

	filter := make([]types.StringFilter, 0, len(list))
	for id := range list {
		filter = append(filter, types.StringFilter{
			Comparison: types.StringFilterComparisonEquals,
			Value:      toPointer(id),
		})
	}

	return filter
}
