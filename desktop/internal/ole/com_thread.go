package ole

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// comRequest represents a request to execute on the COM thread
type comRequest struct {
	companyPath string
	fn          func(*DbApiClient) error
	result      error
	done        chan bool
}

var (
	comClient   *DbApiClient
	comRequests = make(chan *comRequest, 100)
	comStarted  sync.Once
	comShutdown = make(chan bool)
	comPreload  = make(chan string, 1)  // Channel to request preloading
	comClose    = make(chan bool, 1)    // Channel to request closing
	lastActivity time.Time               // Track last activity for idle timeout
	idleTimeout = 5 * time.Minute       // Close after 5 minutes of inactivity
)

// initCOMThread starts the dedicated COM thread
func initCOMThread() {
	comStarted.Do(func() {
		go runCOMThread()
	})
}

// runCOMThread runs on a dedicated OS thread for COM operations
func runCOMThread() {
	// Lock this goroutine to its OS thread
	// This is CRITICAL for COM - all COM calls must happen on the same thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	writeLog("COM Thread: Starting dedicated COM thread")
	
	// Create the OLE client on this thread
	var err error
	comClient, err = NewDbApiClient()
	if err != nil {
		writeLog(fmt.Sprintf("COM Thread: Failed to create OLE client: %v", err))
		// Handle all pending requests with error
		for {
			select {
			case req := <-comRequests:
				req.result = fmt.Errorf("COM initialization failed: %w", err)
				req.done <- true
			case <-comShutdown:
				return
			default:
				return
			}
		}
	}
	
	writeLog("COM Thread: OLE client created successfully")
	currentPath := ""
	lastActivity = time.Now()
	
	// Create idle timer
	idleTimer := time.NewTimer(idleTimeout)
	defer idleTimer.Stop()
	
	// Process requests
	for {
		select {
		case companyPath := <-comPreload:
			// Preload request - open database but don't execute anything
			writeLog(fmt.Sprintf("COM Thread: Preloading database for company: %s", companyPath))
			
			if companyPath != "" && companyPath != currentPath {
				if currentPath != "" {
					comClient.CloseDbc()
				}
				
				err := comClient.OpenDbc(companyPath)
				if err != nil {
					writeLog(fmt.Sprintf("COM Thread: Failed to preload database: %v", err))
				} else {
					currentPath = companyPath
					writeLog(fmt.Sprintf("COM Thread: Database preloaded successfully for: %s", companyPath))
				}
			}
			
			lastActivity = time.Now()
			idleTimer.Reset(idleTimeout)
			
		case <-comClose:
			// Explicit close request
			writeLog("COM Thread: Explicit close requested")
			if comClient != nil && currentPath != "" {
				comClient.CloseDbc()
				currentPath = ""
				writeLog("COM Thread: Database closed")
			}
			
		case req := <-comRequests:
			writeLog(fmt.Sprintf("COM Thread: Processing request for company: %s", req.companyPath))
			
			// Switch database if needed
			if req.companyPath != "" && req.companyPath != currentPath {
				writeLog(fmt.Sprintf("COM Thread: Switching database from '%s' to '%s'", currentPath, req.companyPath))
				
				// Close current database if open
				if currentPath != "" {
					comClient.CloseDbc()
				}
				
				// Open new database
				err := comClient.OpenDbc(req.companyPath)
				if err != nil {
					writeLog(fmt.Sprintf("COM Thread: Failed to open database: %v", err))
					req.result = fmt.Errorf("failed to open database: %w", err)
					req.done <- true
					continue
				}
				currentPath = req.companyPath
			}
			
			// Execute the request
			req.result = req.fn(comClient)
			req.done <- true
			
			// Update activity time and reset idle timer
			lastActivity = time.Now()
			idleTimer.Reset(idleTimeout)
			
		case <-idleTimer.C:
			// Check if we've been idle too long
			if time.Since(lastActivity) >= idleTimeout {
				writeLog(fmt.Sprintf("COM Thread: Idle timeout reached (%v since last activity), closing database", time.Since(lastActivity)))
				if currentPath != "" {
					comClient.CloseDbc()
					currentPath = ""
					writeLog("COM Thread: Database closed due to inactivity")
				}
			}
			idleTimer.Reset(idleTimeout)
			
		case <-comShutdown:
			writeLog("COM Thread: Shutting down")
			if comClient != nil {
				comClient.Close()
				comClient = nil
			}
			return
		}
	}
}

// ExecuteOnCOMThread executes a function on the dedicated COM thread
func ExecuteOnCOMThread(companyPath string, fn func(*DbApiClient) error) error {
	// Ensure COM thread is started
	initCOMThread()
	
	// Create request
	req := &comRequest{
		companyPath: companyPath,
		fn:          fn,
		done:        make(chan bool, 1),
	}
	
	// Send request
	comRequests <- req
	
	// Wait for completion
	<-req.done
	
	return req.result
}

// PreloadOLEConnection preloads the OLE connection for a company
func PreloadOLEConnection(companyPath string) {
	initCOMThread()
	
	// Non-blocking send to preload channel
	select {
	case comPreload <- companyPath:
		writeLog(fmt.Sprintf("Preload request sent for company: %s", companyPath))
	default:
		writeLog("Preload request skipped - channel busy")
	}
}

// CloseOLEConnection explicitly closes the current OLE connection
func CloseOLEConnection() {
	// Non-blocking send to close channel
	select {
	case comClose <- true:
		writeLog("Close request sent")
	default:
		writeLog("Close request skipped - channel busy")
	}
}

// SetIdleTimeout sets the idle timeout duration
func SetIdleTimeout(duration time.Duration) {
	idleTimeout = duration
	writeLog(fmt.Sprintf("Idle timeout set to: %v", duration))
}

// ShutdownCOMThread shuts down the COM thread
func ShutdownCOMThread() {
	writeLog("Shutting down COM thread")
	close(comShutdown)
}