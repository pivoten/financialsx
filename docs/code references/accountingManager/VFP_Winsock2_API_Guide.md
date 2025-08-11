# Visual FoxPro Winsock2 API Integration Guide

## Overview
This guide shows how to implement TCP socket communication in Visual FoxPro using native Windows Sockets 2 (Winsock2) API calls. This approach requires no external dependencies or registrations.

## Key Benefits of Winsock2 API
- **No Installation Required** - Built into every Windows version
- **No Registration Needed** - Direct Windows API calls
- **No Dependencies** - No external files or runtimes required
- **Always Available** - Part of Windows kernel
- **Better Performance** - Direct API calls

## Prerequisites
**NOTHING!** Winsock2 is part of Windows. The ws2_32.dll is in every Windows system32 folder.

## Implementation Guide

### Step 1: Create Winsock2 Class

Create a new PRG file called `winsock2listener.prg`:

```foxpro
*====================================================================
* Winsock2Listener Class
* TCP Socket Listener using Windows Sockets 2 API
* No external dependencies required
*====================================================================

DEFINE CLASS Winsock2Listener AS Custom
    
    * Properties
    nSocket = -1
    nPort = 23456
    lListening = .F.
    cLastError = ""
    nClientSocket = -1
    lDebugMode = .T.
    
    * API Constants
    #DEFINE AF_INET         2
    #DEFINE SOCK_STREAM     1
    #DEFINE IPPROTO_TCP     6
    #DEFINE INVALID_SOCKET  -1
    #DEFINE SOCKET_ERROR    -1
    #DEFINE SD_BOTH         2
    #DEFINE FIONBIO         0x8004667E
    
    * Initialize Winsock2 API
    PROCEDURE Init
        LPARAMETERS tnPort
        
        IF PCOUNT() > 0 AND TYPE("tnPort") = "N"
            This.nPort = tnPort
        ENDIF
        
        * Declare all Winsock2 API functions
        This.DeclareAPIs()
        
        * Initialize Winsock
        LOCAL lcWSAData, lnResult
        lcWSAData = SPACE(400)
        lnResult = WSAStartup(0x0202, @lcWSAData)
        
        IF lnResult != 0
            This.cLastError = "Failed to initialize Winsock2. Error: " + STR(lnResult)
            RETURN .F.
        ENDIF
        
        IF This.lDebugMode
            ? "Winsock2 initialized successfully"
        ENDIF
        
        RETURN .T.
    ENDPROC
    
    * Declare all needed API functions
    PROCEDURE DeclareAPIs
        * Core Winsock2 functions
        DECLARE INTEGER WSAStartup IN ws2_32 INTEGER wVersionRequested, STRING @lpWSAData
        DECLARE INTEGER WSACleanup IN ws2_32
        DECLARE INTEGER WSAGetLastError IN ws2_32
        
        * Socket functions
        DECLARE INTEGER socket IN ws2_32 INTEGER af, INTEGER type, INTEGER protocol
        DECLARE INTEGER closesocket IN ws2_32 INTEGER s
        DECLARE INTEGER shutdown IN ws2_32 INTEGER s, INTEGER how
        
        * Binding and listening
        DECLARE INTEGER bind IN ws2_32 INTEGER s, STRING @name, INTEGER namelen
        DECLARE INTEGER listen IN ws2_32 INTEGER s, INTEGER backlog
        DECLARE INTEGER accept IN ws2_32 INTEGER s, STRING @addr, INTEGER @addrlen
        
        * Data transfer
        DECLARE INTEGER send IN ws2_32 INTEGER s, STRING @buf, INTEGER len, INTEGER flags
        DECLARE INTEGER recv IN ws2_32 INTEGER s, STRING @buf, INTEGER len, INTEGER flags
        
        * Utility functions
        DECLARE INTEGER htons IN ws2_32 INTEGER hostshort
        DECLARE INTEGER inet_addr IN ws2_32 STRING cp
        DECLARE INTEGER ioctlsocket IN ws2_32 INTEGER s, INTEGER cmd, INTEGER @argp
        
        * Select for non-blocking operations
        DECLARE INTEGER select IN ws2_32 INTEGER nfds, STRING @readfds, ;
                STRING @writefds, STRING @exceptfds, STRING @timeout
    ENDPROC
    
    * Start listening on specified port
    PROCEDURE StartListening
        LOCAL lnResult
        
        * Create socket
        This.nSocket = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP)
        
        IF This.nSocket = INVALID_SOCKET
            This.cLastError = "Failed to create socket. Error: " + STR(WSAGetLastError())
            RETURN .F.
        ENDIF
        
        IF This.lDebugMode
            ? "Socket created: " + STR(This.nSocket)
        ENDIF
        
        * Set socket to non-blocking mode
        LOCAL lnNonBlocking
        lnNonBlocking = 1
        ioctlsocket(This.nSocket, FIONBIO, @lnNonBlocking)
        
        * Prepare socket address structure
        * Structure: sin_family (2 bytes) + sin_port (2 bytes) + sin_addr (4 bytes) + padding (8 bytes)
        LOCAL lcSockAddr, lnPortBE
        lnPortBE = htons(This.nPort)  && Convert port to big-endian
        
        lcSockAddr = ;
            CHR(BITAND(AF_INET, 0xFF)) + CHR(BITRSHIFT(AF_INET, 8)) + ;  && sin_family
            CHR(BITRSHIFT(lnPortBE, 8)) + CHR(BITAND(lnPortBE, 0xFF)) + ; && sin_port (big-endian)
            CHR(0) + CHR(0) + CHR(0) + CHR(0) + ;                         && sin_addr (INADDR_ANY)
            REPLICATE(CHR(0), 8)                                           && padding
        
        * Bind to port
        lnResult = bind(This.nSocket, @lcSockAddr, 16)
        
        IF lnResult = SOCKET_ERROR
            This.cLastError = "Failed to bind to port " + STR(This.nPort) + ;
                            ". Error: " + STR(WSAGetLastError()) + ;
                            ". Port may be in use."
            This.CloseSocket()
            RETURN .F.
        ENDIF
        
        IF This.lDebugMode
            ? "Bound to port " + STR(This.nPort)
        ENDIF
        
        * Start listening
        lnResult = listen(This.nSocket, 5)  && Allow up to 5 pending connections
        
        IF lnResult = SOCKET_ERROR
            This.cLastError = "Failed to listen. Error: " + STR(WSAGetLastError())
            This.CloseSocket()
            RETURN .F.
        ENDIF
        
        This.lListening = .T.
        
        IF This.lDebugMode
            ? "Listening on port " + STR(This.nPort) + "..."
        ENDIF
        
        RETURN .T.
    ENDPROC
    
    * Check for incoming connections (non-blocking)
    PROCEDURE CheckForConnection
        IF !This.lListening
            RETURN .F.
        ENDIF
        
        LOCAL lcClientAddr, lnAddrLen, lnClientSocket
        lcClientAddr = SPACE(16)
        lnAddrLen = 16
        
        * Try to accept a connection (non-blocking)
        lnClientSocket = accept(This.nSocket, @lcClientAddr, @lnAddrLen)
        
        IF lnClientSocket != INVALID_SOCKET
            This.nClientSocket = lnClientSocket
            IF This.lDebugMode
                ? "Client connected! Socket: " + STR(lnClientSocket)
            ENDIF
            RETURN .T.
        ENDIF
        
        RETURN .F.
    ENDPROC
    
    * Receive data from connected client
    PROCEDURE ReceiveData
        IF This.nClientSocket = INVALID_SOCKET
            RETURN ""
        ENDIF
        
        LOCAL lcBuffer, lnResult
        lcBuffer = SPACE(4096)
        
        lnResult = recv(This.nClientSocket, @lcBuffer, 4096, 0)
        
        IF lnResult > 0
            LOCAL lcData
            lcData = LEFT(lcBuffer, lnResult)
            IF This.lDebugMode
                ? "Received " + STR(lnResult) + " bytes: " + lcData
            ENDIF
            RETURN lcData
        ENDIF
        
        IF lnResult = 0
            IF This.lDebugMode
                ? "Client disconnected"
            ENDIF
            This.CloseClientSocket()
        ENDIF
        
        RETURN ""
    ENDPROC
    
    * Send data to connected client
    PROCEDURE SendData
        LPARAMETERS tcData
        
        IF This.nClientSocket = INVALID_SOCKET
            RETURN .F.
        ENDIF
        
        LOCAL lnResult
        lnResult = send(This.nClientSocket, @tcData, LEN(tcData), 0)
        
        IF lnResult = SOCKET_ERROR
            This.cLastError = "Send failed. Error: " + STR(WSAGetLastError())
            RETURN .F.
        ENDIF
        
        IF This.lDebugMode
            ? "Sent " + STR(lnResult) + " bytes"
        ENDIF
        
        RETURN .T.
    ENDPROC
    
    * Process incoming JSON command
    PROCEDURE ProcessCommand
        LPARAMETERS tcJsonData
        
        * Parse the JSON command
        * Expected formats:
        * {"action":"launchForm","formName":"form.scx","argument":"","company":"COMPANY01"}
        * {"action":"getCompany"}
        * {"action":"setCompany","company":"COMPANY01"}
        
        LOCAL lcResponse, lcAction
        
        * Extract action
        IF '"action":"' $ tcJsonData
            LOCAL lnStart, lnEnd
            lnStart = AT('"action":"', tcJsonData) + 10
            lnEnd = AT('"', tcJsonData, lnStart)
            lcAction = SUBSTR(tcJsonData, lnStart, lnEnd - lnStart)
        ELSE
            lcResponse = '{"success":false,"message":"No action specified"}'
            This.SendData(lcResponse)
            RETURN
        ENDIF
        
        DO CASE
            CASE lcAction == "getCompany"
                * Return current company
                LOCAL lcCurrentCompany
                lcCurrentCompany = This.GetCurrentCompany()
                lcResponse = '{"success":true,"company":"' + lcCurrentCompany + '"}'
                
            CASE lcAction == "setCompany"
                * Extract and set company
                LOCAL lcCompany
                lnStart = AT('"company":"', tcJsonData) + 11
                lnEnd = AT('"', tcJsonData, lnStart)
                lcCompany = SUBSTR(tcJsonData, lnStart, lnEnd - lnStart)
                
                IF This.SetCurrentCompany(lcCompany)
                    lcResponse = '{"success":true,"message":"Company changed to ' + lcCompany + '"}'
                ELSE
                    lcResponse = '{"success":false,"message":"Failed to change company"}'
                ENDIF
                
            CASE lcAction == "launchForm"
                * First check company match
                IF '"company":"' $ tcJsonData
                    LOCAL lcRequestedCompany
                    lnStart = AT('"company":"', tcJsonData) + 11
                    lnEnd = AT('"', tcJsonData, lnStart)
                    lcRequestedCompany = SUBSTR(tcJsonData, lnStart, lnEnd - lnStart)
                    
                    * Verify company match
                    IF UPPER(lcRequestedCompany) != UPPER(This.GetCurrentCompany())
                        * Company mismatch - ask to switch
                        lcResponse = '{"success":false,"needsCompanyChange":true,' + ;
                                   '"currentCompany":"' + This.GetCurrentCompany() + '",' + ;
                                   '"requestedCompany":"' + lcRequestedCompany + '",' + ;
                                   '"message":"Company mismatch"}'
                        This.SendData(lcResponse)
                        RETURN
                    ENDIF
                ENDIF
                
                * Extract form name
                LOCAL lcFormName
                lnStart = AT('"formName":"', tcJsonData) + 12
                lnEnd = AT('"', tcJsonData, lnStart)
                lcFormName = SUBSTR(tcJsonData, lnStart, lnEnd - lnStart)
                
                IF This.lDebugMode
                    ? "Launching form: " + lcFormName
                ENDIF
                
                * Try to launch the form
                TRY
                    DO FORM (lcFormName)
                    lcResponse = '{"success":true,"message":"Form launched successfully"}'
                CATCH TO loError
                    lcResponse = '{"success":false,"message":"' + ;
                               STRTRAN(loError.Message, '"', '\"') + '"}'
                ENDTRY
                
            OTHERWISE
                lcResponse = '{"success":false,"message":"Unknown action: ' + lcAction + '"}'
        ENDCASE
        
        * Send response back
        This.SendData(lcResponse)
        
        RETURN
    ENDPROC
    
    * Get current company from FoxPro application
    PROCEDURE GetCurrentCompany
        * This needs to be customized based on how your app stores current company
        * Examples:
        
        * Option 1: Global variable
        IF TYPE("gcCompany") = "C"
            RETURN gcCompany
        ENDIF
        
        * Option 2: Application object property
        IF TYPE("goApp.cCurrentCompany") = "C"
            RETURN goApp.cCurrentCompany
        ENDIF
        
        * Option 3: From a settings table
        IF USED("Settings")
            SELECT Settings
            LOCATE FOR SettingName = "CurrentCompany"
            IF FOUND()
                RETURN Settings.SettingValue
            ENDIF
        ENDIF
        
        * Default if not found
        RETURN "UNKNOWN"
    ENDPROC
    
    * Set current company in FoxPro application
    PROCEDURE SetCurrentCompany
        LPARAMETERS tcCompany
        
        * This needs to be customized based on how your app changes companies
        * Examples:
        
        * Option 1: Call your existing company change procedure
        IF FILE("changecompany.prg")
            DO changecompany WITH tcCompany
            RETURN .T.
        ENDIF
        
        * Option 2: Set global variable and reinitialize
        IF TYPE("gcCompany") = "C"
            gcCompany = tcCompany
            * You may need to reinitialize data paths, reopen tables, etc.
            IF TYPE("goApp.ChangeCompany") = "C"
                goApp.ChangeCompany(tcCompany)
            ENDIF
            RETURN .T.
        ENDIF
        
        * Option 3: Update settings and refresh
        IF USED("Settings")
            SELECT Settings
            LOCATE FOR SettingName = "CurrentCompany"
            IF FOUND()
                REPLACE Settings.SettingValue WITH tcCompany
                * Trigger refresh of data
                RETURN .T.
            ENDIF
        ENDIF
        
        RETURN .F.
    ENDPROC
    
    * Close client connection
    PROCEDURE CloseClientSocket
        IF This.nClientSocket != INVALID_SOCKET
            shutdown(This.nClientSocket, SD_BOTH)
            closesocket(This.nClientSocket)
            This.nClientSocket = INVALID_SOCKET
            IF This.lDebugMode
                ? "Client socket closed"
            ENDIF
        ENDIF
    ENDPROC
    
    * Close main socket
    PROCEDURE CloseSocket
        IF This.nSocket != INVALID_SOCKET
            shutdown(This.nSocket, SD_BOTH)
            closesocket(This.nSocket)
            This.nSocket = INVALID_SOCKET
            This.lListening = .F.
            IF This.lDebugMode
                ? "Main socket closed"
            ENDIF
        ENDIF
    ENDPROC
    
    * Stop listening and cleanup
    PROCEDURE StopListening
        This.CloseClientSocket()
        This.CloseSocket()
        This.lListening = .F.
    ENDPROC
    
    * Cleanup on destroy
    PROCEDURE Destroy
        This.StopListening()
        WSACleanup()
        IF This.lDebugMode
            ? "Winsock2 cleaned up"
        ENDIF
    ENDPROC
    
ENDDEFINE
```

### Step 2: Create Timer-Based Listener Form

Create a form that uses a timer to check for connections:

```foxpro
*====================================================================
* Form: TCPListenerForm
* Uses Winsock2Listener class with timer for non-blocking operation
*====================================================================

PUBLIC goTCPListener

* Create and configure form
LOCAL loForm
loForm = CREATEOBJECT("TCPListenerForm")
loForm.Show()
READ EVENTS

DEFINE CLASS TCPListenerForm AS Form
    
    Caption = "TCP Listener (Winsock2 API)"
    Width = 400
    Height = 300
    
    * Add controls
    ADD OBJECT lblStatus AS Label WITH ;
        Caption = "Status: Not Started", ;
        Left = 10, Top = 10, Width = 380, Height = 20
    
    ADD OBJECT txtLog AS EditBox WITH ;
        Left = 10, Top = 40, Width = 380, Height = 200, ;
        ReadOnly = .T., ScrollBars = 2
    
    ADD OBJECT cmdStart AS CommandButton WITH ;
        Caption = "Start Listening", ;
        Left = 10, Top = 250, Width = 120, Height = 30
    
    ADD OBJECT cmdStop AS CommandButton WITH ;
        Caption = "Stop Listening", ;
        Left = 140, Top = 250, Width = 120, Height = 30, ;
        Enabled = .F.
    
    ADD OBJECT cmdClear AS CommandButton WITH ;
        Caption = "Clear Log", ;
        Left = 270, Top = 250, Width = 120, Height = 30
    
    ADD OBJECT tmrCheck AS Timer WITH ;
        Interval = 100, ;  && Check every 100ms
        Enabled = .F.
    
    * Form Init
    PROCEDURE Init
        * Create global listener object
        goTCPListener = CREATEOBJECT("Winsock2Listener", 23456)
        This.AddToLog("TCP Listener initialized")
    ENDPROC
    
    * Start button click
    PROCEDURE cmdStart.Click
        IF goTCPListener.StartListening()
            ThisForm.lblStatus.Caption = "Status: Listening on port " + ;
                                        STR(goTCPListener.nPort)
            ThisForm.AddToLog("Started listening on port " + ;
                            STR(goTCPListener.nPort))
            ThisForm.cmdStart.Enabled = .F.
            ThisForm.cmdStop.Enabled = .T.
            ThisForm.tmrCheck.Enabled = .T.
        ELSE
            ThisForm.AddToLog("Failed to start: " + goTCPListener.cLastError)
            MESSAGEBOX(goTCPListener.cLastError, 16, "Error")
        ENDIF
    ENDPROC
    
    * Stop button click
    PROCEDURE cmdStop.Click
        goTCPListener.StopListening()
        ThisForm.lblStatus.Caption = "Status: Stopped"
        ThisForm.AddToLog("Stopped listening")
        ThisForm.cmdStart.Enabled = .T.
        ThisForm.cmdStop.Enabled = .F.
        ThisForm.tmrCheck.Enabled = .F.
    ENDPROC
    
    * Clear log button
    PROCEDURE cmdClear.Click
        ThisForm.txtLog.Value = ""
    ENDPROC
    
    * Timer event - check for connections and data
    PROCEDURE tmrCheck.Timer
        * Check for new connection
        IF goTCPListener.CheckForConnection()
            ThisForm.AddToLog("Client connected!")
        ENDIF
        
        * Check for incoming data
        IF goTCPListener.nClientSocket != -1
            LOCAL lcData
            lcData = goTCPListener.ReceiveData()
            
            IF !EMPTY(lcData)
                ThisForm.AddToLog("Received: " + lcData)
                
                * Process the command
                goTCPListener.ProcessCommand(lcData)
            ENDIF
        ENDIF
    ENDPROC
    
    * Helper method to add to log
    PROCEDURE AddToLog
        LPARAMETERS tcMessage
        This.txtLog.Value = This.txtLog.Value + ;
                          DTOC(DATE()) + " " + TIME() + " - " + ;
                          tcMessage + CHR(13) + CHR(10)
    ENDPROC
    
    * Form cleanup
    PROCEDURE Destroy
        IF TYPE("goTCPListener") = "O"
            goTCPListener.StopListening()
            goTCPListener = NULL
        ENDIF
        CLEAR EVENTS
    ENDPROC
    
ENDDEFINE
```

### Step 3: Integration with Your Application

To integrate this into your FoxPro application:

1. **Add to your main application startup:**
```foxpro
* In your main app startup
SET PROCEDURE TO winsock2listener.prg ADDITIVE

* Create global listener
PUBLIC goTCPListener
goTCPListener = CREATEOBJECT("Winsock2Listener", 23456)

IF goTCPListener.StartListening()
    * Add a timer to your main form to check for connections
    * Or use a separate listener form
ELSE
    MESSAGEBOX("Failed to start TCP listener: " + goTCPListener.cLastError, 16)
ENDIF
```

2. **Add timer to your main form:**
```foxpro
* Add this to your main form's timer event (set to 100ms interval)
IF TYPE("goTCPListener") = "O" AND goTCPListener.lListening
    * Check for new connections
    IF goTCPListener.CheckForConnection()
        * Client connected
    ENDIF
    
    * Check for data
    IF goTCPListener.nClientSocket != -1
        LOCAL lcData
        lcData = goTCPListener.ReceiveData()
        IF !EMPTY(lcData)
            goTCPListener.ProcessCommand(lcData)
        ENDIF
    ENDIF
ENDIF
```

## Testing the Connection

### From FinancialsX (Go application)
The Go application doesn't need any changes. It will connect to port 23456 and send JSON commands just like before.

### Manual Test from FoxPro
```foxpro
* Test the listener
SET PROCEDURE TO winsock2listener.prg ADDITIVE
oListener = CREATEOBJECT("Winsock2Listener", 23456)
? oListener.StartListening()  && Should return .T.
? "Listening on port 23456..."
```

### Test with Telnet
```cmd
telnet localhost 23456
```
Then type:
```json
{"action":"launchForm","formName":"test.scx","argument":""}
```

## Troubleshooting

### Port Already in Use
If you get "Failed to bind to port", the port is already in use. Either:
1. Stop the other application using the port
2. Change to a different port number
3. Wait a minute for the port to be released

### Windows Firewall
Windows Firewall may prompt to allow the connection. Click "Allow" when prompted.

### Debugging
Set `lDebugMode = .T.` in the Winsock2Listener class to see detailed output.

## Error Codes Reference

Common Winsock2 error codes:
- **10048** (WSAEADDRINUSE): Port already in use
- **10049** (WSAEADDRNOTAVAIL): Cannot bind address
- **10061** (WSAECONNREFUSED): Connection refused
- **10060** (WSAETIMEDOUT): Connection timeout
- **10035** (WSAEWOULDBLOCK): Non-blocking operation (normal for non-blocking sockets)

## Complete Working Example

Here's a minimal complete example that you can test immediately:

```foxpro
*====================================================================
* SimpleListener.prg - Minimal Winsock2 TCP Listener
* Save this as SimpleListener.prg and run it
*====================================================================

CLEAR
? "Starting Simple TCP Listener..."

* Declare APIs
DECLARE INTEGER WSAStartup IN ws2_32 INTEGER wVersionRequested, STRING @lpWSAData
DECLARE INTEGER WSACleanup IN ws2_32
DECLARE INTEGER socket IN ws2_32 INTEGER af, INTEGER type, INTEGER protocol
DECLARE INTEGER bind IN ws2_32 INTEGER s, STRING @name, INTEGER namelen
DECLARE INTEGER listen IN ws2_32 INTEGER s, INTEGER backlog
DECLARE INTEGER accept IN ws2_32 INTEGER s, STRING @addr, INTEGER @addrlen
DECLARE INTEGER recv IN ws2_32 INTEGER s, STRING @buf, INTEGER len, INTEGER flags
DECLARE INTEGER send IN ws2_32 INTEGER s, STRING @buf, INTEGER len, INTEGER flags
DECLARE INTEGER closesocket IN ws2_32 INTEGER s
DECLARE INTEGER htons IN ws2_32 INTEGER hostshort

* Initialize Winsock
lcWSAData = SPACE(400)
IF WSAStartup(0x0202, @lcWSAData) != 0
    ? "Failed to initialize Winsock"
    RETURN
ENDIF
? "Winsock initialized"

* Create socket
lnSocket = socket(2, 1, 6)  && AF_INET, SOCK_STREAM, IPPROTO_TCP
? "Socket created: " + STR(lnSocket)

* Bind to port 23456
lnPort = htons(23456)
lcAddr = CHR(2) + CHR(0) + ;  && AF_INET
         CHR(BITRSHIFT(lnPort,8)) + CHR(BITAND(lnPort,0xFF)) + ;  && Port
         REPLICATE(CHR(0), 12)  && Address + padding

IF bind(lnSocket, @lcAddr, 16) = -1
    ? "Bind failed"
    closesocket(lnSocket)
    WSACleanup()
    RETURN
ENDIF
? "Bound to port 23456"

* Listen
IF listen(lnSocket, 5) = -1
    ? "Listen failed"
    closesocket(lnSocket)
    WSACleanup()
    RETURN
ENDIF
? "Listening... Press ESC to stop"

* Accept loop
DO WHILE !INKEY() = 27  && ESC to exit
    lcClientAddr = SPACE(16)
    lnAddrLen = 16
    lnClient = accept(lnSocket, @lcClientAddr, @lnAddrLen)
    
    IF lnClient != -1
        ? "Client connected!"
        
        * Receive data
        lcBuffer = SPACE(1024)
        lnBytes = recv(lnClient, @lcBuffer, 1024, 0)
        IF lnBytes > 0
            ? "Received: " + LEFT(lcBuffer, lnBytes)
            
            * Send response
            lcResponse = '{"status":"ok","message":"Hello from FoxPro Winsock2"}'
            send(lnClient, @lcResponse, LEN(lcResponse), 0)
        ENDIF
        
        closesocket(lnClient)
    ENDIF
ENDDO

* Cleanup
closesocket(lnSocket)
WSACleanup()
? "Listener stopped"
```

## Summary

The Winsock2 API approach eliminates all dependency issues while providing reliable TCP communication. Your FoxPro developer can implement this without requiring any additional installations on client machines. The API calls work on all Windows versions from XP to Windows 11.