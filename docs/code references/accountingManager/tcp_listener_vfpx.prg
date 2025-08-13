*====================================================================
* integrations_tcp_listener_vfpx.prg  (no PUBLIC memvars; _SCREEN props)
* Winsock2 TCP listener with non-blocking accept + blocking client I/O
* - Test handshake returns "OK"
* - Real commands return JSON (launchForm, getCompany, setCompany)
* - Tolerant JSON value extraction (handles whitespace around :)
* - Robust logging to tcp_listener.log
*====================================================================

#DEFINE AF_INET        2
#DEFINE SOCK_STREAM    1
#DEFINE IPPROTO_TCP    6
#DEFINE SOL_SOCKET     0xFFFF
#DEFINE SO_REUSEADDR   4
#DEFINE SO_SNDTIMEO    0x1005
#DEFINE SO_RCVTIMEO    0x1006
#DEFINE INVALID_SOCKET -1
#DEFINE SOCKET_ERROR   -1
#DEFINE SD_BOTH        2
#DEFINE FIONBIO        0x8004667E

PROCEDURE _WS2_Declare
    DECLARE INTEGER WSAStartup       IN ws2_32.dll INTEGER, STRING @
    DECLARE INTEGER WSACleanup       IN ws2_32.dll
    DECLARE INTEGER WSAGetLastError  IN ws2_32.dll

    DECLARE INTEGER socket           IN ws2_32.dll INTEGER, INTEGER, INTEGER
    DECLARE INTEGER closesocket      IN ws2_32.dll AS ws2_closesocket INTEGER
    DECLARE INTEGER bind             IN ws2_32.dll AS ws2_bind INTEGER, STRING @, INTEGER
    DECLARE INTEGER listen           IN ws2_32.dll AS ws2_listen INTEGER, INTEGER
    DECLARE INTEGER accept           IN ws2_32.dll INTEGER, STRING @, INTEGER @
    DECLARE INTEGER recv             IN ws2_32.dll INTEGER, STRING @, INTEGER, INTEGER
    DECLARE INTEGER send             IN ws2_32.dll INTEGER, STRING @, INTEGER, INTEGER
    DECLARE INTEGER shutdown         IN ws2_32.dll INTEGER, INTEGER
    DECLARE INTEGER setsockopt       IN ws2_32.dll AS ws2_setsockopt INTEGER, INTEGER, INTEGER, STRING @, INTEGER
    DECLARE INTEGER ioctlsocket      IN ws2_32.dll INTEGER, LONG, LONG @

    DECLARE INTEGER htons            IN ws2_32.dll INTEGER
    DECLARE LONG    htonl            IN ws2_32.dll LONG
ENDPROC

*---------------- Public API ----------------
FUNCTION StartTCPListener(tnPort)
    RETURN StartWinsock2Listener(tnPort)
ENDFUNC

FUNCTION StopTCPListener
    RETURN StopWinsock2Listener()
ENDFUNC

*---------------- Start ----------------
FUNCTION StartWinsock2Listener(tnPort)
    LOCAL nPort, cWSA, nRet, nPortNet, nAddrAny, cSockAddr, cOpt, lnNonBlock

    nPort = IIF(VARTYPE(tnPort)='N', tnPort, 23456)
    _WS2_Declare()
    _WS2_EnsureScreenState()

    _WS2_Log("=== StartWinsock2Listener ===  Port: " + TRANSFORM(nPort))

    cWSA = REPLICATE(CHR(0), 400)
    nRet = WSAStartup(0x0202, @cWSA)
    IF nRet # 0
        _WS2_Log("WSAStartup failed: " + TRANSFORM(nRet))
        RETURN .F.
    ENDIF

    _SCREEN.WS2_Sock = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP)
    IF _SCREEN.WS2_Sock = INVALID_SOCKET
        _WS2_Log("socket() failed: " + TRANSFORM(WSAGetLastError()))
        =WSACleanup()
        _SCREEN.WS2_Sock = -1
        RETURN .F.
    ENDIF

    * Reuse address
    cOpt = BINTOC(1, "4RS")
    =ws2_setsockopt(_SCREEN.WS2_Sock, SOL_SOCKET, SO_REUSEADDR, @cOpt, 4)

    * Non-blocking listener
    lnNonBlock = 1
    =ioctlsocket(_SCREEN.WS2_Sock, FIONBIO, @lnNonBlock)

    * sockaddr_in
    nPortNet  = htons(nPort)
    nAddrAny  = htonl(0)
    cSockAddr = BINTOC(AF_INET,"2RS")+BINTOC(nPortNet,"2RS")+BINTOC(nAddrAny,"4RS")+REPLICATE(CHR(0),8)

    nRet = ws2_bind(_SCREEN.WS2_Sock, @cSockAddr, 16)
    IF nRet = SOCKET_ERROR
        _WS2_Log("bind() failed: " + TRANSFORM(WSAGetLastError()) + " port:" + TRANSFORM(nPort))
        _WS2_CloseListener()
        RETURN .F.
    ENDIF

    nRet = ws2_listen(_SCREEN.WS2_Sock, 8)
    IF nRet = SOCKET_ERROR
        _WS2_Log("listen() failed: " + TRANSFORM(WSAGetLastError()))
        _WS2_CloseListener()
        RETURN .F.
    ENDIF

    _SCREEN.WS2_Listening     = .T.
    _SCREEN.WS2_StopRequested = .F.
    _WS2_Log("Listening on port " + TRANSFORM(nPort) + ", sock " + TRANSFORM(_SCREEN.WS2_Sock))
    _WS2_Log("CWD=" + FULLPATH("") + " | PATH=" + SET("PATH"))

    _WS2_PrimeAcceptLoop(_SCREEN.WS2_Sock)

    IF VARTYPE(_SCREEN.WS2_Timer) = 'O'
        _SCREEN.WS2_Timer.Enabled = .F.
        _SCREEN.WS2_Timer = .NULL.
    ENDIF
    _SCREEN.WS2_Timer = CREATEOBJECT("WS2Timer", _SCREEN.WS2_Sock)
    _SCREEN.WS2_Timer.Interval = 50
    _SCREEN.WS2_Timer.Enabled  = .T.
    _WS2_Log("Timer started")

    RETURN .T.
ENDFUNC

*---------------- Stop ----------------
FUNCTION StopWinsock2Listener
    _WS2_EnsureScreenState()
    _SCREEN.WS2_StopRequested = .T.

    IF VARTYPE(_SCREEN.WS2_Timer)='O'
        _SCREEN.WS2_Timer.Enabled = .F.
        _SCREEN.WS2_Timer = .NULL.
    ENDIF

    _WS2_CloseListener()
    _SCREEN.WS2_Listening = .F.
    _WS2_Log("Listener stopped")
    RETURN .T.
ENDFUNC

PROCEDURE _WS2_CloseListener
    IF VARTYPE(_SCREEN.WS2_Sock)='N' AND _SCREEN.WS2_Sock > -1
        =ws2_closesocket(_SCREEN.WS2_Sock)
        _SCREEN.WS2_Sock = -1
    ENDIF
    =WSACleanup()
ENDPROC

PROCEDURE _WS2_EnsureScreenState
    IF TYPE("_SCREEN.WS2_Sock") # "N"
        _SCREEN.AddProperty("WS2_Sock", -1)
        _SCREEN.AddProperty("WS2_Listening", .F.)
        _SCREEN.AddProperty("WS2_StopRequested", .F.)
        _SCREEN.AddProperty("WS2_Timer", .NULL.)
    ENDIF
ENDPROC

*---------------- Timer ----------------
DEFINE CLASS WS2Timer AS Timer
    nSock = -1
    PROCEDURE Init
        LPARAMETERS tnSock
        This.nSock = IIF(VARTYPE(tnSock)='N', tnSock, -1)
    ENDPROC

    PROCEDURE Timer
        IF VARTYPE(_SCREEN.WS2_Listening)#'L' OR VARTYPE(_SCREEN.WS2_StopRequested)#'L'
            RETURN
        ENDIF
        IF !_SCREEN.WS2_Listening OR _SCREEN.WS2_StopRequested
            RETURN
        ENDIF
        IF This.nSock <= 0
            RETURN
        ENDIF

        LOCAL cAddr, nAddrLen, nClient, nErr
        cAddr    = REPLICATE(CHR(0), 16)
        nAddrLen = 16
        nClient  = accept(This.nSock, @cAddr, @nAddrLen)

        IF nClient = INVALID_SOCKET
            nErr = WSAGetLastError()
            IF nErr # 10035 && WSAEWOULDBLOCK
                _WS2_Log("accept() error: " + TRANSFORM(nErr))
            ENDIF
            RETURN
        ENDIF

        _WS2_ServeClient(nClient)
    ENDPROC
ENDDEFINE

*---------------- Accept helper ----------------
PROCEDURE _WS2_PrimeAcceptLoop(tnSock)
    LOCAL i, cA, nL, nC, nE
    FOR i = 1 TO 40
        IF tnSock <= 0
            EXIT
        ENDIF
        cA = REPLICATE(CHR(0),16)
        nL = 16
        nC = accept(tnSock, @cA, @nL)
        IF nC = INVALID_SOCKET
            nE = WSAGetLastError()
            IF nE = 10035
                INKEY(0.01)
                LOOP
            ELSE
                _WS2_Log("prime accept error: " + TRANSFORM(nE))
                EXIT
            ENDIF
        ELSE
            _WS2_ServeClient(nC)
        ENDIF
    ENDFOR
ENDPROC

*---------------- Serve one connection ----------------
PROCEDURE _WS2_ServeClient(tnClient)
    LOCAL lnBlocking, cTimeout, cBuf, nRecv, lcResponse, nSend

    lnBlocking = 0
    =ioctlsocket(tnClient, FIONBIO, @lnBlocking)

    cTimeout = BINTOC(3000, "4RS")
    =ws2_setsockopt(tnClient, SOL_SOCKET, SO_RCVTIMEO, @cTimeout, 4)
    =ws2_setsockopt(tnClient, SOL_SOCKET, SO_SNDTIMEO, @cTimeout, 4)

    cBuf  = REPLICATE(CHR(0), 8192)
    nRecv = recv(tnClient, @cBuf, LEN(cBuf), 0)

    IF nRecv = -1
        _WS2_Log("recv() err: " + TRANSFORM(WSAGetLastError()))
        lcResponse = 'ERR Failed to receive data'
    ELSE
        IF nRecv > 0
            lcResponse = ProcessVFPCommand(LEFT(cBuf, nRecv))
        ELSE
            lcResponse = 'OK'
        ENDIF
    ENDIF

    lcResponse = lcResponse + CHR(13) + CHR(10)
    nSend = send(tnClient, @lcResponse, LEN(lcResponse), 0)
    IF nSend = -1
        _WS2_Log("send() err: " + TRANSFORM(WSAGetLastError()))
    ENDIF

    =shutdown(tnClient, SD_BOTH)
    =ws2_closesocket(tnClient)
ENDPROC

*---------------- Process + Launch ----------------
FUNCTION ProcessVFPCommand(tcData)
    LOCAL lcData, lcAction, lcFormName, lcCompany, lcResponse
    lcData = ALLTRIM(tcData)
    _WS2_Log("ProcessVFPCommand len=" + TRANSFORM(LEN(lcData)) + " data=" + LEFT(lcData,256))

    IF '"form":"TEST"' $ lcData OR '"TEST"' $ lcData OR "TEST" $ UPPER(lcData)
        RETURN "OK"
    ENDIF

    * Parse (tolerant to whitespace and casing)
    lcAction   = ExtractJSONValueFlexible(lcData, "action")
    lcFormName = ExtractJSONValueFlexible(lcData, "formName")
    IF EMPTY(lcAction) AND EMPTY(lcFormName)
        lcFormName = ExtractJSONValueFlexible(lcData, "form")
        IF !EMPTY(lcFormName)
            lcAction = "launchForm"
        ENDIF
    ENDIF
    _WS2_Log("Parsed action='" + IIF(EMPTY(lcAction),"(none)",lcAction) + "' formName='" + lcFormName + "'")

    DO CASE
    CASE lcAction == "getCompany"
        lcCompany  = GetCurrentCompany()
        lcResponse = '{"success":true,"company":"' + lcCompany + '"}'

    CASE lcAction == "setCompany"
        lcCompany = ExtractJSONValueFlexible(lcData, "company")
        SetCurrentCompany(lcCompany)
        lcResponse = '{"success":true,"message":"Company set to ' + lcCompany + '"}'

    CASE lcAction == "launchForm"
        IF EMPTY(lcFormName)
            lcResponse = '{"success":false,"message":"No form name provided"}'
        ELSE
            LOCAL lcMsg
            IF LaunchFormSafe(lcFormName, @lcMsg)
                lcResponse = '{"success":true,"message":"' + lcMsg + '"}'
            ELSE
                lcResponse = '{"success":false,"message":"' + ;
                    STRTRAN(IIF(EMPTY(lcMsg),'Unable to launch form',lcMsg), '"','\"') + '"}'
            ENDIF
        ENDIF

    OTHERWISE
        lcResponse = '{"success":false,"message":"Unknown action: ' + IIF(EMPTY(lcAction),'(none)',lcAction) + '"}'
    ENDCASE

    RETURN lcResponse
ENDFUNC

*----- Tolerant JSON extractor: "key" [spaces] : [spaces] "value" -----
FUNCTION ExtractJSONValueFlexible(tcJson, tcKey)
    LOCAL lcJson, lcKeyQuoted, nKeyPos, nFrom, cRest, nColonPos, nAfterColon, nOpenQ, nCloseQ, cVal

    lcJson = tcJson
    lcKeyQuoted = '"' + LOWER(ALLTRIM(tcKey)) + '"'

    * Find "key" ignoring case
    nKeyPos = ATC(lcKeyQuoted, lcJson)
    IF nKeyPos = 0
        RETURN ""
    ENDIF

    * Move to just after the key token
    nFrom = nKeyPos + LEN(lcKeyQuoted)

    * Take the remainder and skip whitespace to find the colon
    cRest = SUBSTR(lcJson, nFrom)
    cRest = LTRIM(STRTRAN(STRTRAN(STRTRAN(cRest, CHR(13), ""), CHR(10), ""), CHR(9), ""))

    IF EMPTY(cRest)
        RETURN ""
    ENDIF

    IF LEFT(cRest,1) # ":"
        nColonPos = AT(":", cRest)
        IF nColonPos = 0
            RETURN ""
        ENDIF
        cRest = SUBSTR(cRest, nColonPos)  && starts with ":..."
    ENDIF

    * Now skip ":" and whitespace, expect opening quote
    cRest = SUBSTR(cRest, 2)  && after ":"
    cRest = LTRIM(STRTRAN(STRTRAN(STRTRAN(cRest, CHR(13), ""), CHR(10), ""), CHR(9), ""))

    nOpenQ = AT('"', cRest)
    IF nOpenQ = 0
        RETURN ""
    ENDIF
    cRest  = SUBSTR(cRest, nOpenQ + 1)   && after first quote

    nCloseQ = AT('"', cRest)
    IF nCloseQ = 0
        RETURN ""
    ENDIF

    cVal = LEFT(cRest, nCloseQ - 1)
    RETURN cVal
ENDFUNC

*----- Launch helpers -----
FUNCTION LaunchFormSafe(tcForm, tcOutMsg)
    LOCAL lcIn, lcName, lcFile, llOK, lcTried, lcErr
    lcIn    = ALLTRIM(tcForm)
    lcTried = ""
    llOK    = .F.
    lcErr   = ""

    _WS2_Log("LaunchFormSafe IN: " + lcIn + " | CWD=" + FULLPATH("") + " | PATH=" + SET("PATH"))

    lcName = lcIn
    IF UPPER(RIGHT(lcName,4)) == ".SCX"
        lcName = LEFT(lcName, LEN(lcName)-4)
    ENDIF
    lcFile = lcName + ".scx"

    IF (RAT("\", lcIn) > 0 OR RAT("/", lcIn) > 0) AND UPPER(JUSTEXT(lcIn)) == "SCX"
        lcTried = lcTried + " [DO FORM literal: " + lcIn + "]"
        _WS2_Log("DO FORM literal attempt: " + lcIn)
        TRY
            DO FORM (lcIn)
            llOK = .T.
        CATCH TO loE
            lcErr = "DO literal err: " + loE.Message
            _WS2_Log(lcErr)
        ENDTRY
    ENDIF

    IF !llOK
        lcTried = lcTried + " [DO FORM base: " + lcName + "]"
        _WS2_Log("DO FORM base attempt: " + lcName)
        TRY
            DO FORM (lcName)
            llOK = .T.
        CATCH TO loE
            lcErr = "DO base err: " + loE.Message
            _WS2_Log(lcErr)
        ENDTRY
    ENDIF

    IF !llOK
        LOCAL lcFound
        lcFound = _WS2_FindOnPath(lcFile)
        IF !EMPTY(lcFound)
            lcTried = lcTried + " [DO FORM found: " + lcFound + "]"
            _WS2_Log("DO FORM explicit path: " + lcFound)
            TRY
                DO FORM (lcFound)
                llOK = .T.
            CATCH TO loE
                lcErr = "DO explicit err: " + loE.Message
                _WS2_Log(lcErr)
            ENDTRY
        ELSE
            _WS2_Log("Not found on PATH: " + lcFile)
        ENDIF
    ENDIF

    IF !llOK
        lcTried = lcTried + " [CREATEOBJECT: " + lcName + "]"
        _WS2_Log("CREATEOBJECT attempt: " + lcName)
        TRY
            LOCAL loFrm
            loFrm = CREATEOBJECT(lcName)
            IF VARTYPE(loFrm)='O'
                loFrm.Show()
                llOK = .T.
            ENDIF
        CATCH TO loE
            lcErr = "CREATEOBJECT err: " + loE.Message
            _WS2_Log(lcErr)
        ENDTRY
    ENDIF

    IF llOK
        tcOutMsg = "Form launched: " + lcName
        _WS2_Log("SUCCESS " + tcOutMsg + " | tried:" + lcTried)
    ELSE
        IF EMPTY(lcErr)
            lcErr = "Form not found or failed to open: " + lcIn
        ENDIF
        tcOutMsg = lcErr + " | tried:" + lcTried
        _WS2_Log("FAIL " + tcOutMsg)
    ENDIF

    RETURN llOK
ENDFUNC

FUNCTION _WS2_FindOnPath(tcFile)
    LOCAL lcFile, lcFull, lcPath, lnDirs, aDirs[1], i
    lcFile = ALLTRIM(tcFile)

    lcFull = FULLPATH(lcFile)
    IF FILE(lcFull)
        RETURN lcFull
    ENDIF

    lcPath = SET("PATH")
    IF EMPTY(lcPath)
        RETURN ""
    ENDIF

    lnDirs = ALINES(aDirs, lcPath, .T., ";")
    FOR i = 1 TO lnDirs
        lcFull = FULLPATH(ADDBS(aDirs[i]) + lcFile)
        IF FILE(lcFull)
            RETURN lcFull
        ENDIF
    ENDFOR
    RETURN ""
ENDFUNC

*----- Company helpers -----
FUNCTION GetCurrentCompany
    IF TYPE("gcCompany")="C"
        RETURN gcCompany
    ENDIF
    IF TYPE("goApp")="O" AND TYPE("goApp.cCurrentCompany")="C"
        RETURN goApp.cCurrentCompany
    ENDIF
    RETURN "UNKNOWN"
ENDFUNC

FUNCTION SetCurrentCompany(tcCompany)
    IF TYPE("gcCompany")="C"
        gcCompany = tcCompany
    ENDIF
    IF TYPE("goApp")="O" AND TYPE("goApp.cCurrentCompany")="C"
        goApp.cCurrentCompany = tcCompany
    ENDIF
    RETURN .T.
ENDFUNC

*----- Logger -----
PROCEDURE _WS2_Log(tcMsg)
    LOCAL lcFile
    lcFile = "tcp_listener.log"
    STRTOFILE(TTOC(DATETIME(), 1) + " - " + tcMsg + CHR(13) + CHR(10), lcFile, 1)
ENDPROC