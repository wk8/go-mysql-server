// Code generated by "stringer -type=RuleId -linecomment"; DO NOT EDIT.

package analyzer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[validateOffsetAndLimitId-0]
	_ = x[validateCreateTableId-1]
	_ = x[validateExprSemId-2]
	_ = x[resolveVariablesId-3]
	_ = x[resolveNamedWindowsId-4]
	_ = x[resolveSetVariablesId-5]
	_ = x[resolveViewsId-6]
	_ = x[liftCtesId-7]
	_ = x[resolveCtesId-8]
	_ = x[liftRecursiveCtesId-9]
	_ = x[resolveDatabasesId-10]
	_ = x[resolveTablesId-11]
	_ = x[loadStoredProceduresId-12]
	_ = x[validateDropTablesId-13]
	_ = x[setTargetSchemasId-14]
	_ = x[resolveCreateLikeId-15]
	_ = x[parseColumnDefaultsId-16]
	_ = x[resolveDropConstraintId-17]
	_ = x[validateDropConstraintId-18]
	_ = x[loadCheckConstraintsId-19]
	_ = x[resolveCreateSelectId-20]
	_ = x[resolveSubqueriesId-21]
	_ = x[setViewTargetSchemaId-22]
	_ = x[resolveUnionsId-23]
	_ = x[resolveDescribeQueryId-24]
	_ = x[checkUniqueTableNamesId-25]
	_ = x[resolveTableFunctionsId-26]
	_ = x[resolveDeclarationsId-27]
	_ = x[validateCreateTriggerId-28]
	_ = x[validateCreateProcedureId-29]
	_ = x[loadInfoSchemaId-30]
	_ = x[validateReadOnlyDatabaseId-31]
	_ = x[validateReadOnlyTransactionId-32]
	_ = x[validateDatabaseSetId-33]
	_ = x[validatePrivilegesId-34]
	_ = x[reresolveTablesId-35]
	_ = x[validateJoinComplexityId-36]
	_ = x[resolveNaturalJoinsId-37]
	_ = x[resolveOrderbyLiteralsId-38]
	_ = x[resolveFunctionsId-39]
	_ = x[flattenTableAliasesId-40]
	_ = x[pushdownSortId-41]
	_ = x[pushdownGroupbyAliasesId-42]
	_ = x[pushdownSubqueryAliasFiltersId-43]
	_ = x[qualifyColumnsId-44]
	_ = x[resolveColumnsId-45]
	_ = x[resolveColumnDefaultsId-46]
	_ = x[validateCheckConstraintId-47]
	_ = x[resolveBarewordSetVariablesId-48]
	_ = x[expandStarsId-49]
	_ = x[resolveHavingId-50]
	_ = x[mergeUnionSchemasId-51]
	_ = x[flattenAggregationExprsId-52]
	_ = x[reorderProjectionId-53]
	_ = x[resolveSubqueryExprsId-54]
	_ = x[resolveJSONTableCrossJoinId-55]
	_ = x[replaceCrossJoinsId-56]
	_ = x[moveJoinCondsToFilterId-57]
	_ = x[evalFilterId-58]
	_ = x[optimizeDistinctId-59]
	_ = x[finalizeSubqueriesId-60]
	_ = x[finalizeUnionsId-61]
	_ = x[loadTriggersId-62]
	_ = x[processTruncateId-63]
	_ = x[resolveAlterColumnId-64]
	_ = x[resolveGeneratorsId-65]
	_ = x[removeUnnecessaryConvertsId-66]
	_ = x[assignCatalogId-67]
	_ = x[pruneColumnsId-68]
	_ = x[optimizeJoinsId-69]
	_ = x[pushdownFiltersId-70]
	_ = x[subqueryIndexesId-71]
	_ = x[inSubqueryIndexesId-72]
	_ = x[pruneTablesId-73]
	_ = x[setJoinScopeLenId-74]
	_ = x[eraseProjectionId-75]
	_ = x[replaceSortPkId-76]
	_ = x[insertTopNId-77]
	_ = x[cacheSubqueryResultsId-78]
	_ = x[cacheSubqueryAliasesInJoinsId-79]
	_ = x[applyHashLookupsId-80]
	_ = x[applyHashInId-81]
	_ = x[resolveInsertRowsId-82]
	_ = x[resolvePreparedInsertId-83]
	_ = x[applyTriggersId-84]
	_ = x[applyProceduresId-85]
	_ = x[assignRoutinesId-86]
	_ = x[modifyUpdateExprsForJoinId-87]
	_ = x[applyRowUpdateAccumulatorsId-88]
	_ = x[wrapWithRollbackId-89]
	_ = x[applyFKsId-90]
	_ = x[validateResolvedId-91]
	_ = x[validateOrderById-92]
	_ = x[validateGroupById-93]
	_ = x[validateSchemaSourceId-94]
	_ = x[validateIndexCreationId-95]
	_ = x[validateOperandsId-96]
	_ = x[validateCaseResultTypesId-97]
	_ = x[validateIntervalUsageId-98]
	_ = x[validateExplodeUsageId-99]
	_ = x[validateSubqueryColumnsId-100]
	_ = x[validateUnionSchemasMatchId-101]
	_ = x[validateAggregationsId-102]
	_ = x[AutocommitId-103]
	_ = x[TrackProcessId-104]
	_ = x[parallelizeId-105]
	_ = x[clearWarningsId-106]
}

const _RuleId_name = "validateOffsetAndLimitvalidateCreateTablevalidateExprSemresolveVariablesresolveNamedWindowsresolveSetVariablesresolveViewsliftCtesresolveCtesliftRecursiveCtesresolveDatabasesresolveTablesloadStoredProceduresvalidateDropTablessetTargetSchemasresolveCreateLikeparseColumnDefaultsresolveDropConstraintvalidateDropConstraintloadCheckConstraintsresolveCreateSelectresolveSubqueriessetViewTargetSchemaresolveUnionsresolveDescribeQuerycheckUniqueTableNamesresolveTableFunctionsresolveDeclarationsvalidateCreateTriggervalidateCreateProcedureloadInfoSchemavalidateReadOnlyDatabasevalidateReadOnlyTransactionvalidateDatabaseSetvalidatePrivilegesreresolveTablesvalidateJoinComplexityresolveNaturalJoinsresolveOrderbyLiteralsresolveFunctionsflattenTableAliasespushdownSortpushdownGroupbyAliasespushdownSubqueryAliasFiltersqualifyColumnsresolveColumnsresolveColumnDefaultsvalidateCheckConstraintresolveBarewordSetVariablesexpandStarsresolveHavingmergeUnionSchemasflattenAggregationExprsreorderProjectionresolveSubqueryExprsresolveJSONTableCrossJoinreplaceCrossJoinsmoveJoinCondsToFilterevalFilteroptimizeDistinctfinalizeSubqueriesfinalizeUnionsloadTriggersprocessTruncateresolveAlterColumnresolveGeneratorsremoveUnnecessaryConvertsassignCatalogpruneColumnsoptimizeJoinspushdownFilterssubqueryIndexesinSubqueryIndexespruneTablessetJoinScopeLeneraseProjectionreplaceSortPkinsertTopNcacheSubqueryResultscacheSubqueryAliasesInJoinsapplyHashLookupsapplyHashInresolveInsertRowsresolvePreparedInsertapplyTriggersapplyProceduresassignRoutinesmodifyUpdateExprsForJoinapplyRowUpdateAccumulatorsrollback triggersapplyFKsvalidateResolvedvalidateOrderByvalidateGroupByvalidateSchemaSourcevalidateIndexCreationvalidateOperandsvalidateCaseResultTypesvalidateIntervalUsagevalidateExplodeUsagevalidateSubqueryColumnsvalidateUnionSchemasMatchvalidateAggregationsaddAutocommitNodetrackProcessparallelizeclearWarnings"

var _RuleId_index = [...]uint16{0, 22, 41, 56, 72, 91, 110, 122, 130, 141, 158, 174, 187, 207, 225, 241, 258, 277, 298, 320, 340, 359, 376, 395, 408, 428, 449, 470, 489, 510, 533, 547, 571, 598, 617, 635, 650, 672, 691, 713, 729, 748, 760, 782, 810, 824, 838, 859, 882, 909, 920, 933, 950, 973, 990, 1010, 1035, 1052, 1073, 1083, 1099, 1117, 1131, 1143, 1158, 1176, 1193, 1218, 1231, 1243, 1256, 1271, 1286, 1303, 1314, 1329, 1344, 1357, 1367, 1387, 1414, 1430, 1441, 1458, 1479, 1492, 1507, 1521, 1545, 1571, 1588, 1596, 1612, 1627, 1642, 1662, 1683, 1699, 1722, 1743, 1763, 1786, 1811, 1831, 1848, 1860, 1871, 1884}

func (i RuleId) String() string {
	if i < 0 || i >= RuleId(len(_RuleId_index)-1) {
		return "RuleId(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _RuleId_name[_RuleId_index[i]:_RuleId_index[i+1]]
}
