package message

// Shared i18n keys used by Go bindings and React locale files.
const (
	KeyErrorsGeneric = "errors.generic"

	KeyErrorsAuthMissingCredentials = "errors.auth.missing_credentials"
	KeyErrorsAuthLoginFailed        = "errors.auth.login_failed"
	KeyErrorsAuthSessionInvalid       = "errors.auth.session_invalid"

	KeyErrorsBalanceUnavailable = "errors.balance.unavailable"

	KeyErrorsPlannerLoadFallback            = "errors.planner.load_fallback"
	KeyErrorsPlannerJourneyExitBeforeEntry  = "errors.planner.journey_exit_before_entry"
	KeyErrorsPlannerJourneyEntryBeforeExit  = "errors.planner.journey_entry_before_exit"

	KeyUpdateBlockersInstallNotFound   = "update.blockers.install_not_found"
	KeyUpdateBlockersPMNotFound        = "update.blockers.pm_not_found"
	KeyUpdateBlockersGoNotFound        = "update.blockers.go_not_found"
	KeyUpdateBlockersNodeNotFound      = "update.blockers.node_not_found"
	KeyUpdateBlockersDirtyWorkingTree  = "update.blockers.dirty_working_tree"
	KeyUpdateBlockersFetchFailed       = "update.blockers.fetch_failed"
	KeyUpdateBlockersCompareFailed     = "update.blockers.compare_failed"
	KeyUpdateBlockersCheckFailed       = "update.blockers.check_failed"

	KeyUpdateResultSuccess        = "update.result.success"
	KeyUpdateResultSuccessCommit  = "update.result.success_commit"
	KeyUpdateResultFailed         = "update.result.failed"

	KeyRemindersSlotInMinutes      = "reminders.slot_in_minutes"
	KeyRemindersRecommendedTime    = "reminders.recommended_time"
	KeyRemindersRuntimeUnavailable = "reminders.runtime_unavailable"
	KeyRemindersNotAvailable       = "reminders.not_available"
)
