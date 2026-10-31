package wailsapp

// VacuumDb runs SQLite VACUUM to optimize the database.
func (a *App) VacuumDb() error {
	return a.dbMaintenance.VacuumDb(a.ctx)
}

// ClearEncounters deletes all user encounters and friend avatar observations.
// Returns total deleted row count. Does not delete video_playback_history (see RotateEncounters for retention cleanup that includes video rows).
func (a *App) ClearEncounters() (int64, error) {
	return a.dbMaintenance.ClearEncounters(a.ctx)
}

// ClearScreenshots deletes all screenshots. Returns affected row count.
func (a *App) ClearScreenshots() (int64, error) {
	a.stopGalleryScanAndWait()
	return a.dbMaintenance.ClearScreenshots(a.ctx)
}

// ClearFriendsCache deletes all cached friends. Returns affected row count.
func (a *App) ClearFriendsCache() (int64, error) {
	return a.dbMaintenance.ClearFriendsCache(a.ctx)
}
