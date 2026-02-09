package service

import (
	"time"

	"anyadmin-backend/pkg/utils"
)

func StartImportTask(taskID uint) {
	// Need to find the index in slice since we don't have DB ID lookup easily in slice
	// For mock, we iterate by ID if we assigned them, but utils init doesn't assign IDs.
	// We will just update the last added task for simplicity or iterate by Name if unique.
	// Assuming the caller passed a valid ID (or we just pick the last one for demo).
	
	go func() {
		var idx int
		shouldRun := false

		utils.ExecuteWrite(func() {
			// Just pick the last task as the "active" one for this mock
			if len(utils.ImportTasks) == 0 {
				return
			}
			idx = len(utils.ImportTasks) - 1
			utils.ImportTasks[idx].Status = "Running"
			utils.ImportTasks[idx].TotalFiles = 100
			shouldRun = true
		}, true)

		if !shouldRun {
			return
		}

		for i := 0; i < 100; i++ {
			time.Sleep(50 * time.Millisecond)
			// Don't persist progress updates to avoid IO spam
			utils.ExecuteWrite(func() {
				utils.ImportTasks[idx].Processed = i + 1
				utils.ImportTasks[idx].Progress = (i + 1)
			}, false)
		}

		utils.ExecuteWrite(func() {
			utils.ImportTasks[idx].Status = "Completed"
		}, true)
	}()
}
