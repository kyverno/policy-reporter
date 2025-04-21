package securityhub

// func (c *client) some(polr v1alpha2.ReportInterface, results []payload.Payload) {
// 	/*
// 		- should send result
// 		- is result prexisted
// 		- to securityhub
// 	*/

// 	// get results that are failures
// 	results = filterResults(results)
// 	if len(results) == 0 {
// 		return
// 	}

// 	// get the list of findings for resource or report or individual results
// 	// an api call was made to aws here
// 	list, err := c.getFindingsByIDs(context.Background(), polr, toResourceIDFilter(polr, results), "")
// 	if err != nil {
// 		zap.L().Error(c.Name()+": failed to get findings", zap.Error(err))
// 		return
// 	}

// 	// why do i need to filter again fr ?
// 	list = filterFindings(list, results)

// 	// turn the findings to an array of findings identifier ? why ?
// 	findings := helper.Map(list, func(f types.AwsSecurityFinding) types.AwsSecurityFindingIdentifier {
// 		return types.AwsSecurityFindingIdentifier{
// 			Id:         f.Id,
// 			ProductArn: f.ProductArn,
// 		}
// 	})

// 	// update the existing findings and get the ones remaining that were not there before
// 	if len(findings) > 0 {
// 		updated, err := c.batchUpdate(context.Background(), findings, types.WorkflowStatusNew)
// 		if err != nil {
// 			zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err))
// 			return
// 		} else if updated > 0 {
// 			zap.L().Info(c.Name()+": PUSH OK", zap.Int("updated", updated))
// 		}

// 		// build a map of the existing findings
// 		mapping := make(map[string]bool, len(list))
// 		for _, f := range list {
// 			mapping[*f.Id] = true
// 		}

// 		// filter the original list of results by ones that are new
// 		results = helper.Filter(results, func(result v1alpha2.PolicyReportResult) bool {
// 			return !mapping[result.GetID()]
// 		})
// 	}

// 	if len(results) == 0 {
// 		return
// 	}

// 	res, err := c.hub.BatchImportFindings(context.Background(), &hub.BatchImportFindingsInput{
// 		Findings: c.mapFindings(polr, results),
// 	})
// 	if err != nil {
// 		zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err), zap.Any("response", res))
// 		return
// 	}

// 	zap.L().Info(c.Name()+": PUSH OK", zap.Int32("imported", *res.SuccessCount), zap.Int32("failed", *res.FailedCount), zap.String("report", polr.GetKey()))
// }
