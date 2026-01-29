package service

import (
	"time"

	"anyadmin-backend/pkg/mockdata"
)

func StartImportTask(taskID uint) {
	// Need to find the index in slice since we don't have DB ID lookup easily in slice
	// For mock, we iterate by ID if we assigned them, but mockdata init doesn't assign IDs.
	// We will just update the last added task for simplicity or iterate by Name if unique.
	// Assuming the caller passed a valid ID (or we just pick the last one for demo).
	
	go func() {
		mockdata.Mu.Lock()
		// Just pick the last task as the "active" one for this mock
		if len(mockdata.ImportTasks) == 0 {
			mockdata.Mu.Unlock()
			return
		}
		idx := len(mockdata.ImportTasks) - 1
		mockdata.ImportTasks[idx].Status = "Running"
		mockdata.ImportTasks[idx].TotalFiles = 100
		mockdata.Mu.Unlock()

		for i := 0; i < 100; i++ {
			time.Sleep(50 * time.Millisecond)
			mockdata.Mu.Lock()
			mockdata.ImportTasks[idx].Processed = i + 1
			mockdata.ImportTasks[idx].Progress = (i + 1)
			mockdata.Mu.Unlock()
		}

		mockdata.Mu.Lock()
		mockdata.ImportTasks[idx].Status = "Completed"
		mockdata.Mu.Unlock()
	}()
}
