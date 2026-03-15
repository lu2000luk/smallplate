// Endpoints contained in this file:
// GET /{plateID}/info
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerInfo(mux *http.ServeMux, deps *plate.Dependencies) {
	mux.HandleFunc("GET /{plateID}/info", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		ctx := r.Context()
		var cursor uint64
		var count int64
		var sampleCount int64
		var sampleBytes int64
		for {
			keys, next, err := deps.Redis.Scan(ctx, cursor, plate.PrefixPattern(plateID, "*"), 500).Result()
			if err != nil {
				return err
			}
			count += int64(len(keys))
			for index, key := range keys {
				if index >= 10 {
					break
				}
				usage, err := deps.Redis.MemoryUsage(ctx, key, 0).Result()
				if err == nil && usage > 0 {
					sampleCount++
					sampleBytes += usage
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
		approxMemory := int64(0)
		if sampleCount > 0 && count > 0 {
			approxMemory = (sampleBytes / sampleCount) * count
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"keyCount": count, "memoryUsageBytes": approxMemory})
		return nil
	}))
}
