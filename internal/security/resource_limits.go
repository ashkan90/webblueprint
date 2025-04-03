package security

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ResourceLimits defines constraints for blueprint execution
type ResourceLimits struct {
	MaxCPUTime        time.Duration // Maximum CPU time allowed
	MaxMemoryBytes    uint64        // Maximum memory allocation
	MaxExecutionTime  time.Duration // Maximum total execution time
	MaxDiskIOBytes    uint64        // Maximum disk I/O operations
	MaxNetworkIOBytes uint64        // Maximum network I/O operations
	MaxNodeExecutions uint64        // Maximum number of node executions
}

// DefaultResourceLimits provides sensible defaults for resource limits
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxCPUTime:        10 * time.Second,
		MaxMemoryBytes:    100 * 1024 * 1024, // 100MB
		MaxExecutionTime:  30 * time.Second,
		MaxDiskIOBytes:    10 * 1024 * 1024, // 10MB
		MaxNetworkIOBytes: 10 * 1024 * 1024, // 10MB
		MaxNodeExecutions: 1000,             // Maximum 1000 node executions per blueprint
	}
}

// ResourceProfile represents a named set of resource limits
type ResourceProfile string

const (
	ResourceProfileLow     ResourceProfile = "low"     // For simple blueprints
	ResourceProfileMedium  ResourceProfile = "medium"  // For normal blueprints
	ResourceProfileHigh    ResourceProfile = "high"    // For complex blueprints
	ResourceProfileUnlimit ResourceProfile = "unlimit" // For trusted, unrestricted blueprints
)

// GetResourceLimits returns appropriate resource limits for a profile
func GetResourceLimits(profile ResourceProfile) ResourceLimits {
	switch profile {
	case ResourceProfileLow:
		return ResourceLimits{
			MaxCPUTime:        5 * time.Second,
			MaxMemoryBytes:    50 * 1024 * 1024, // 50MB
			MaxExecutionTime:  15 * time.Second,
			MaxDiskIOBytes:    5 * 1024 * 1024, // 5MB
			MaxNetworkIOBytes: 5 * 1024 * 1024, // 5MB
			MaxNodeExecutions: 500,             // Maximum 500 node executions
		}
	case ResourceProfileMedium:
		return DefaultResourceLimits()
	case ResourceProfileHigh:
		return ResourceLimits{
			MaxCPUTime:        30 * time.Second,
			MaxMemoryBytes:    500 * 1024 * 1024, // 500MB
			MaxExecutionTime:  120 * time.Second,
			MaxDiskIOBytes:    50 * 1024 * 1024, // 50MB
			MaxNetworkIOBytes: 50 * 1024 * 1024, // 50MB
			MaxNodeExecutions: 3000,             // Maximum 3000 node executions
		}
	case ResourceProfileUnlimit:
		return ResourceLimits{
			MaxCPUTime:        1 * time.Hour,
			MaxMemoryBytes:    2 * 1024 * 1024 * 1024, // 2GB
			MaxExecutionTime:  1 * time.Hour,
			MaxDiskIOBytes:    1 * 1024 * 1024 * 1024, // 1GB
			MaxNetworkIOBytes: 1 * 1024 * 1024 * 1024, // 1GB
			MaxNodeExecutions: 10000,                  // Maximum 10000 node executions
		}
	default:
		return DefaultResourceLimits()
	}
}

// ResourceMonitor tracks resource usage for a blueprint execution
type ResourceMonitor struct {
	limits              ResourceLimits
	startTime           time.Time
	currentMem          uint64
	cpuTimeStart        time.Time
	cpuTime             time.Duration
	diskIO              uint64
	networkIO           uint64
	nodeExecutions      uint64
	ctx                 context.Context
	cancel              context.CancelFunc
	mutex               sync.RWMutex
	limitExceeded       bool
	limitExceededReason string
}

// NewResourceMonitor creates a new resource monitor with specified limits
func NewResourceMonitor(limits ResourceLimits) *ResourceMonitor {
	ctx, cancel := context.WithTimeout(context.Background(), limits.MaxExecutionTime)

	monitor := &ResourceMonitor{
		limits:         limits,
		startTime:      time.Now(),
		cpuTimeStart:   time.Now(),
		ctx:            ctx,
		cancel:         cancel,
		nodeExecutions: 0,
		limitExceeded:  false,
	}

	// Start monitoring goroutine
	go monitor.monitorResources()

	return monitor
}

// monitorResources continuously checks resource usage against limits
func (rm *ResourceMonitor) monitorResources() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var memStats runtime.MemStats

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.mutex.Lock()

			// Check memory usage
			runtime.ReadMemStats(&memStats)
			rm.currentMem = memStats.Alloc

			// Check if memory limit exceeded
			if rm.currentMem > rm.limits.MaxMemoryBytes {
				rm.limitExceeded = true
				rm.limitExceededReason = fmt.Sprintf("Memory limit exceeded: %d bytes used (limit: %d bytes)",
					rm.currentMem, rm.limits.MaxMemoryBytes)
				rm.cancel()
				rm.mutex.Unlock()
				return
			}

			// Check execution time
			execTime := time.Since(rm.startTime)
			if execTime > rm.limits.MaxExecutionTime {
				rm.limitExceeded = true
				rm.limitExceededReason = fmt.Sprintf("Execution time limit exceeded: %v (limit: %v)",
					execTime, rm.limits.MaxExecutionTime)
				rm.cancel()
				rm.mutex.Unlock()
				return
			}

			rm.mutex.Unlock()
		}
	}
}

// TrackNetworkIO records network I/O operations and checks limits
func (rm *ResourceMonitor) TrackNetworkIO(bytes uint64) bool {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Check if context is already canceled
	if rm.ctx.Err() != nil || rm.limitExceeded {
		return false
	}

	rm.networkIO += bytes

	// Check if network I/O limit exceeded
	if rm.networkIO > rm.limits.MaxNetworkIOBytes {
		rm.limitExceeded = true
		rm.limitExceededReason = fmt.Sprintf("Network I/O limit exceeded: %d bytes (limit: %d bytes)",
			rm.networkIO, rm.limits.MaxNetworkIOBytes)
		rm.cancel()
		return false
	}

	return true
}

// TrackDiskIO records disk I/O operations and checks limits
func (rm *ResourceMonitor) TrackDiskIO(bytes uint64) bool {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Check if context is already canceled
	if rm.ctx.Err() != nil || rm.limitExceeded {
		return false
	}

	rm.diskIO += bytes

	// Check if disk I/O limit exceeded
	if rm.diskIO > rm.limits.MaxDiskIOBytes {
		rm.limitExceeded = true
		rm.limitExceededReason = fmt.Sprintf("Disk I/O limit exceeded: %d bytes (limit: %d bytes)",
			rm.diskIO, rm.limits.MaxDiskIOBytes)
		rm.cancel()
		return false
	}

	return true
}

// TrackNodeExecution increments node execution count and checks limits
func (rm *ResourceMonitor) TrackNodeExecution() bool {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Check if context is already canceled
	if rm.ctx.Err() != nil || rm.limitExceeded {
		return false
	}

	rm.nodeExecutions++

	// Check if node execution limit exceeded
	if rm.nodeExecutions > rm.limits.MaxNodeExecutions {
		rm.limitExceeded = true
		rm.limitExceededReason = fmt.Sprintf("Node execution limit exceeded: %d executions (limit: %d)",
			rm.nodeExecutions, rm.limits.MaxNodeExecutions)
		rm.cancel()
		return false
	}

	return true
}

// EstimateMemorySize tries to estimate the memory size of a value
// This is a simplified implementation that only handles basic types
func (rm *ResourceMonitor) EstimateMemorySize(value interface{}) uint64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case string:
		return uint64(len(v))
	case []byte:
		return uint64(len(v))
	case []interface{}:
		var size uint64 = 0
		for _, item := range v {
			size += rm.EstimateMemorySize(item)
		}
		return size
	case map[string]interface{}:
		var size uint64 = 0
		for k, v := range v {
			size += uint64(len(k))
			size += rm.EstimateMemorySize(v)
		}
		return size
	default:
		// Default to a small fixed size for other types
		return 8
	}
}

// TrackMemory records memory allocation and checks limits
func (rm *ResourceMonitor) TrackMemory(bytes uint64) bool {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Check if context is already canceled
	if rm.ctx.Err() != nil || rm.limitExceeded {
		return false
	}

	// Use the current memory from ReadMemStats
	currentMemory := rm.currentMem + bytes

	// Check if memory limit exceeded
	if currentMemory > rm.limits.MaxMemoryBytes {
		rm.limitExceeded = true
		rm.limitExceededReason = fmt.Sprintf("Memory limit would be exceeded: %d bytes (limit: %d bytes)",
			currentMemory, rm.limits.MaxMemoryBytes)
		rm.cancel()
		return false
	}

	return true
}

// Context returns the monitor's context, which is canceled when limits are exceeded
func (rm *ResourceMonitor) Context() context.Context {
	return rm.ctx
}

// LimitExceeded returns true if any resource limit was exceeded
func (rm *ResourceMonitor) LimitExceeded() bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.limitExceeded || rm.ctx.Err() != nil
}

// GetLimitExceededReason returns the reason why a limit was exceeded
func (rm *ResourceMonitor) GetLimitExceededReason() string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if rm.limitExceededReason != "" {
		return rm.limitExceededReason
	}

	if rm.ctx.Err() != nil {
		return fmt.Sprintf("Execution timeout: exceeded %v", rm.limits.MaxExecutionTime)
	}

	return "No limits exceeded"
}

// GetResourceStats returns current resource usage statistics
func (rm *ResourceMonitor) GetResourceStats() map[string]interface{} {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return map[string]interface{}{
		"executionTime":  time.Since(rm.startTime).String(),
		"memoryUsed":     rm.currentMem,
		"networkIO":      rm.networkIO,
		"diskIO":         rm.diskIO,
		"nodeExecutions": rm.nodeExecutions,
		"limitExceeded":  rm.limitExceeded,
	}
}

// Cleanup releases resources used by the monitor
func (rm *ResourceMonitor) Cleanup() {
	rm.cancel()
}
