package workflows

import (
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
)

func (w *WorkerEnv) AcquireDeviceLock(deviceId string) error {
	lockDuration := time.Minute * 1 // Lock duration

	// Try to set the lock using SetNX (SET if Not eXists)
	result, err := w.Redis.SetNX(deviceId, "lock", lockDuration).Result()

	if err != nil {
		return fmt.Errorf("error while trying to acquire lock: %w", err)
	}
	if result {
		// Successfully acquired the lock
		w.Logger.Warnf("Successfully acquired lock for %s\n", deviceId)
		return nil
	}
	w.Logger.Errorf("Failed to acquire lock for %s", deviceId)
	return tasks.NewErrRetryTaskLater("Failed to acquire lock", 10 * time.Second)
}

func (w *WorkerEnv) ReleaseDeviceLock(deviceId string) error {
	w.Logger.Warnf("Releasing lock for %s\n", deviceId)
	_, err := w.Redis.Del(deviceId).Result()
	return err
}

func (w *WorkerEnv) ReleaseDeviceLockOnError(errorMsg string, deviceId string) error {
    w.Logger.WithField("error", errorMsg).Error("Releasing lock due to error")
    return w.ReleaseDeviceLock(deviceId)
}
