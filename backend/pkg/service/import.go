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
		utils.Mu.Lock()
		// Just pick the last task as the "active" one for this mock
		if len(utils.ImportTasks) == 0 {
			utils.Mu.Unlock()
			return
		}
		idx := len(utils.ImportTasks) - 1
		utils.ImportTasks[idx].Status = "Running"
		utils.ImportTasks[idx].TotalFiles = 100
		utils.Mu.Unlock()

		for i := 0; i < 100; i++ {
			time.Sleep(50 * time.Millisecond)
			utils.Mu.Lock()
			utils.ImportTasks[idx].Processed = i + 1
			utils.ImportTasks[idx].Progress = (i + 1)
			utils.Mu.Unlock()
		}

		utils.Mu.Lock()
		utils.ImportTasks[idx].Status = "Completed"
		utils.Mu.Unlock()
	}()
}
