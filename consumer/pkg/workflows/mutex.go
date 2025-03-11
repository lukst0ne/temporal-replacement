package workflows

import (
	"fmt"
	"time"
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
	w.Logger.Warnf("Failed to acquire lock for %s", deviceId)
	return fmt.Errorf("failed to acquire lock for %s", deviceId)
}

func (w *WorkerEnv) ReleaseDeviceLock(deviceId string) error {
	_, err := w.Redis.Del(deviceId).Result()
	return err
}
