package main

import "fmt"

type WorkerOperation int64

const (
	WOPQuitObsolete                       WorkerOperation = 0
	WOPIsValidPath                                        = 1
	WOPQuerySubstitutesObsolete                           = 2
	WOPQuerySubstitutes                                   = 3
	WOPQueryPathHashObsolete                              = 4
	WOPQueryReferencesObsolete                            = 5
	WOPQueryReferrers                                     = 6
	WOPAddToStore                                         = 7
	WOPAddTextToStore                                     = 8
	WOPBuildPaths                                         = 9
	WOPEnsurePath                                         = 10
	WOPAddTempRoot                                        = 11
	WOPAddIndirectRoot                                    = 12
	WOPSyncWithGC                                         = 13
	WOPFindRoots                                          = 14
	WOPCollectGarbageObsolete                             = 15
	WOPExportPathObsolete                                 = 16
	WOPImportPathObsolete                                 = 17
	WOPQueryDeriverObsolete                               = 18
	WOPSetOptions                                         = 19
	WOPCollectGarbage                                     = 20
	WOPQuerySubstitutablePathInfo                         = 21
	WOPQueryDerivationOutputsObsolete                     = 22
	WOPQueryAllValidPaths                                 = 23
	WOPQueryFailedPaths                                   = 24
	WOPClearFailedPaths                                   = 25
	WOPQueryPathInfo                                      = 26
	WOPImportPathsObsolete                                = 27
	WOPQueryDerivationOutputNamesObsolete                 = 28
	WOPQueryPathFromHashPart                              = 29
	WOPQuerySubstitutablePathInfos                        = 30
	WOPQueryValidPaths                                    = 31
	WOPQuerySubstitutablePaths                            = 32
	WOPQueryValidDerivers                                 = 33
	WOPOptimiseStore                                      = 34
	WOPVerifyStore                                        = 35
	WOPBuildDerivation                                    = 36
	WOPAddSignatures                                      = 37
	WOPNarFromPath                                        = 38
	WOPAddToStoreNar                                      = 39
	WOPQueryMissing                                       = 40
	WOPQueryDerivationOutputMap                           = 41
	WOPRegisterDrvOutput                                  = 42
	WOPQueryRealisation                                   = 43
	WOPAddMultipleToStore                                 = 44
	WOPAddBuildLog                                        = 45
	WOPBuildPathsWithResults                              = 46
)

func (w WorkerOperation) String() string {
	switch w {
	case WOPQuitObsolete:
		return "QuitObsolete"
	case WOPIsValidPath:
		return "IsValidPath"
	case WOPQuerySubstitutesObsolete:
		return "QuerySubstitutesObsolete"
	case WOPQuerySubstitutes:
		return "QuerySubstitutes"
	case WOPQueryPathHashObsolete:
		return "QueryPathHashObsolete"
	case WOPQueryReferencesObsolete:
		return "QueryReferencesObsolete"
	case WOPQueryReferrers:
		return "QueryReferrers"
	case WOPAddToStore:
		return "AddToStore"
	case WOPAddTextToStore:
		return "AddTextToStore"
	case WOPBuildPaths:
		return "BuildPaths"
	case WOPEnsurePath:
		return "EnsurePath"
	case WOPAddTempRoot:
		return "AddTempRoot"
	case WOPAddIndirectRoot:
		return "AddIndirectRoot"
	case WOPSyncWithGC:
		return "SyncWithGC"
	case WOPFindRoots:
		return "FindRoots"
	case WOPCollectGarbageObsolete:
		return "CollectGarbageObsolete"
	case WOPExportPathObsolete:
		return "ExportPathObsolete"
	case WOPImportPathObsolete:
		return "ImportPathObsolete"
	case WOPQueryDeriverObsolete:
		return "QueryDeriverObsolete"
	case WOPSetOptions:
		return "SetOptions"
	case WOPCollectGarbage:
		return "CollectGarbage"
	case WOPQuerySubstitutablePathInfo:
		return "QuerySubstitutablePathInfo"
	case WOPQueryDerivationOutputsObsolete:
		return "QueryDerivationOutputsObsolete"
	case WOPQueryAllValidPaths:
		return "QueryAllValidPaths"
	case WOPQueryFailedPaths:
		return "QueryFailedPaths"
	case WOPClearFailedPaths:
		return "ClearFailedPaths"
	case WOPQueryPathInfo:
		return "QueryPathInfo"
	case WOPImportPathsObsolete:
		return "ImportPathsObsolete"
	case WOPQueryDerivationOutputNamesObsolete:
		return "QueryDerivationOutputNamesObsolete"
	case WOPQueryPathFromHashPart:
		return "QueryPathFromHashPart"
	case WOPQuerySubstitutablePathInfos:
		return "QuerySubstitutablePathInfos"
	case WOPQueryValidPaths:
		return "QueryValidPaths"
	case WOPQuerySubstitutablePaths:
		return "QuerySubstitutablePaths"
	case WOPQueryValidDerivers:
		return "QueryValidDerivers"
	case WOPOptimiseStore:
		return "OptimiseStore"
	case WOPVerifyStore:
		return "VerifyStore"
	case WOPBuildDerivation:
		return "BuildDerivation"
	case WOPAddSignatures:
		return "AddSignatures"
	case WOPNarFromPath:
		return "NarFromPath"
	case WOPAddToStoreNar:
		return "AddToStoreNar"
	case WOPQueryMissing:
		return "QueryMissing"
	case WOPQueryDerivationOutputMap:
		return "QueryDerivationOutputMap"
	case WOPRegisterDrvOutput:
		return "RegisterDrvOutput"
	case WOPQueryRealisation:
		return "QueryRealisation"
	case WOPAddMultipleToStore:
		return "AddMultipleToStore"
	case WOPAddBuildLog:
		return "AddBuildLog"
	case WOPBuildPathsWithResults:
		return "BuildPathsWithResults"
	default:
		return fmt.Sprintf("Unknown WorkerOperation(%d)", w)
	}
}
