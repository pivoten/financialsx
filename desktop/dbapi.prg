*----------------------------------------------------------
* dbapi.prg  (Project name: Pivoten)
* Build: Win32 EXE / COM server
* Register: dbapi.exe /regserver
* ProgID: Pivoten.DbApi
*----------------------------------------------------------

* Sensible defaults
SET EXCLUSIVE OFF
SET SAFETY OFF
SET DELETED ON
SET TALK OFF
SET NOTIFY OFF

DEFINE CLASS DbApi AS Custom OLEPUBLIC
    cVersion   = "1.0.1"
    cRoot      = ""         && optional base folder
    cDbcPath   = ""         && full path to opened appdata.dbc
    cLastError = ""
    lDebugMode = .F.        && Enable debug logging

    * Optional: set a base folder; defaults to this EXE's folder
    FUNCTION Initialize(tcRoot)
        IF VARTYPE(tcRoot) # "C" OR EMPTY(tcRoot)
            tcRoot = FULLPATH(JUSTPATH(SYS(16,1)) + "\")
        ENDIF
        THIS.cRoot = ADDBS(FULLPATH(tcRoot))
        RETURN .T.
    ENDFUNC

    * Open a DBC by folder (containing appdata.dbc) or full path to the DBC
    FUNCTION OpenDbc(tcPathOrDbc)
        LOCAL lcPath, lcDbc, lnError, lcError
        THIS.cLastError = ""

        IF VARTYPE(tcPathOrDbc) # "C" OR EMPTY(tcPathOrDbc)
            THIS.cLastError = "Path or DBC is required."
            RETURN .F.
        ENDIF

        lcPath = FULLPATH(tcPathOrDbc, IIF(EMPTY(THIS.cRoot), "", THIS.cRoot))

        IF UPPER(JUSTEXT(lcPath)) == "DBC"
            lcDbc = lcPath
        ELSE
            lcDbc = ADDBS(lcPath) + "appdata.dbc"
        ENDIF

        IF !FILE(lcDbc)
            THIS.cLastError = "DBC not found: " + lcDbc
            RETURN .F.
        ENDIF

        * Close current database if different
        IF NOT EMPTY(DBC())
            IF UPPER(FULLPATH(DBC())) # UPPER(FULLPATH(lcDbc))
                CLOSE DATABASES
            ENDIF
        ENDIF

        * Open the database if not already open
        IF EMPTY(DBC())
            lnError = 0
            lcError = ""
            ON ERROR lnError = ERROR()
            OPEN DATABASE (lcDbc) SHARED
            ON ERROR  && Reset error handler
            
            IF lnError > 0
                lcError = MESSAGE()
                THIS.cLastError = "Error opening DBC: " + TRANSFORM(lnError) + " - " + lcError
                RETURN .F.
            ENDIF
        ENDIF

        THIS.cDbcPath = FULLPATH(DBC())
        RETURN .T.
    ENDFUNC

    FUNCTION IsDbcOpen
        RETURN !EMPTY(DBC())
    ENDFUNC

    FUNCTION GetDbcPath
        RETURN IIF(EMPTY(DBC()), "", FULLPATH(DBC()))
    ENDFUNC

    FUNCTION CloseDbc
        IF !EMPTY(DBC())
            CLOSE DATABASES
        ENDIF
        THIS.cDbcPath = ""
        RETURN .T.
    ENDFUNC

    * SELECT into XML for Go to parse
    * tcSql must be a complete SELECT ... (no INTO)
    FUNCTION SelectToXml(tcSql)
        LOCAL lcCursor, lcXml, lcCmd, lnError, lcError, lcTempFile, lcOldDir
        THIS.cLastError = ""
        
        IF EMPTY(DBC())
            THIS.cLastError = "No DBC open."
            RETURN ""
        ENDIF
        
        IF VARTYPE(tcSql) # "C" OR EMPTY(tcSql)
            THIS.cLastError = "SQL text is required."
            RETURN ""
        ENDIF

        lcCursor = SYS(2015)
        lcXml    = ""
        lcCmd    = ALLTRIM(tcSql) + " INTO CURSOR " + lcCursor + " READWRITE"

        * Save current directory and switch to temp
        lcOldDir = FULLPATH(CURDIR())
        CD (SYS(2023))  && Change to Windows temp directory
        
        * Use traditional error handling instead of TRY/CATCH
        lnError = 0
        lcError = ""
        ON ERROR lnError = ERROR()
        
        * Execute the SQL command
        &lcCmd
        
        ON ERROR  && Reset error handler
        
        IF lnError > 0
            lcError = MESSAGE()
            THIS.cLastError = TRANSFORM(lnError) + ": " + lcError
            CD (lcOldDir)  && Restore directory
            RETURN ""
        ENDIF

        IF USED(lcCursor)
            * Check if cursor has records
            SELECT (lcCursor)
            LOCAL lnRecCount
            lnRecCount = RECCOUNT()
            
            IF lnRecCount > 0
                * Create temp file in temp directory
                lcTempFile = SYS(2023) + "\" + SYS(2015) + ".xml"
                
                * CURSORTOXML to file first, then read it
                * Format 1 = Element-centric, Flags 1 = with schema
                CURSORTOXML(lcCursor, lcTempFile, 1, 1)
                
                * Read the XML file into string
                IF FILE(lcTempFile)
                    lcXml = FILETOSTR(lcTempFile)
                    DELETE FILE (lcTempFile)
                ELSE
                    THIS.cLastError = "XML file was not created: " + lcTempFile
                ENDIF
            ELSE
                * No records, return empty result XML
                lcXml = "<?xml version='1.0' encoding='windows-1252'?><VFPData><" + lcCursor + "></" + lcCursor + "></VFPData>"
            ENDIF
            
            USE IN (lcCursor)
        ELSE
            THIS.cLastError = "Cursor was not created from query"
        ENDIF
        
        * Restore original directory
        CD (lcOldDir)
        
        RETURN lcXml
    ENDFUNC

    * Non-query (INSERT/UPDATE/DELETE)
    FUNCTION ExecNonQuery(tcSql)
        LOCAL lcCmd, lnError, lcError
        THIS.cLastError = ""
        
        IF EMPTY(DBC())
            THIS.cLastError = "No DBC open."
            RETURN .F.
        ENDIF
        
        IF VARTYPE(tcSql) # "C" OR EMPTY(tcSql)
            THIS.cLastError = "SQL text is required."
            RETURN .F.
        ENDIF

        lcCmd = tcSql
        
        * Use traditional error handling
        lnError = 0
        lcError = ""
        ON ERROR lnError = ERROR()
        
        * Execute the SQL command
        &lcCmd
        
        ON ERROR  && Reset error handler
        
        IF lnError > 0
            lcError = MESSAGE()
            THIS.cLastError = TRANSFORM(lnError) + ": " + lcError
            RETURN .F.
        ENDIF
        
        RETURN .T.
    ENDFUNC

    * Simple query test - returns cursor count
    FUNCTION TestQuery(tcTable)
        LOCAL lcCmd, lnCount, lnError, lcError, laCount[1]
        THIS.cLastError = ""
        
        IF VARTYPE(tcTable) # "C" OR EMPTY(tcTable)
            tcTable = "COA"
        ENDIF
        
        lnCount = -1
        lnError = 0
        lcError = ""
        laCount[1] = 0  && Initialize array
        
        ON ERROR lnError = ERROR()
        
        * Try to select from the table
        lcCmd = "SELECT COUNT(*) AS cnt FROM " + tcTable + " INTO ARRAY laCount"
        &lcCmd
        
        ON ERROR  && Reset error handler
        
        IF lnError > 0
            lcError = MESSAGE()
            THIS.cLastError = "Query failed: " + TRANSFORM(lnError) + " - " + lcError
            RETURN -1
        ENDIF
        
        IF TYPE("laCount[1]") = "N"
            lnCount = laCount[1]
        ENDIF
        
        RETURN lnCount
    ENDFUNC

    * List tables in the current database
    FUNCTION ListTables()
        LOCAL lcXml, lcCursor, lnError, lcError, lcTempFile, lcOldDir
        THIS.cLastError = ""
        
        IF EMPTY(DBC())
            THIS.cLastError = "No DBC open."
            RETURN ""
        ENDIF
        
        lcCursor = SYS(2015)
        lcXml = ""
        
        * Save current directory and switch to temp
        lcOldDir = FULLPATH(CURDIR())
        CD (SYS(2023))  && Change to Windows temp directory
        
        lnError = 0
        lcError = ""
        ON ERROR lnError = ERROR()
        
        * Get list of tables in the database
        SELECT objectname AS table_name, objectid AS table_id ;
            FROM DBC() ;
            WHERE objecttype = 'Table' ;
            INTO CURSOR (lcCursor) READWRITE
        
        ON ERROR  && Reset error handler
        
        IF lnError > 0
            lcError = MESSAGE()
            THIS.cLastError = "Failed to list tables: " + TRANSFORM(lnError) + " - " + lcError
            CD (lcOldDir)  && Restore directory
            RETURN ""
        ENDIF
        
        IF USED(lcCursor)
            * Check if cursor has records
            SELECT (lcCursor)
            LOCAL lnRecCount
            lnRecCount = RECCOUNT()
            
            IF lnRecCount > 0
                * Create temp file in temp directory
                lcTempFile = SYS(2023) + "\" + SYS(2015) + ".xml"
                
                * CURSORTOXML to file first, then read it
                * Format 1 = Element-centric, Flags 1 = with schema
                CURSORTOXML(lcCursor, lcTempFile, 1, 1)
                
                * Read the XML file into string
                IF FILE(lcTempFile)
                    lcXml = FILETOSTR(lcTempFile)
                    DELETE FILE (lcTempFile)
                ELSE
                    THIS.cLastError = "XML file was not created: " + lcTempFile
                ENDIF
            ELSE
                * No records, return empty result XML
                lcXml = "<?xml version='1.0' encoding='windows-1252'?><VFPData><" + lcCursor + "></" + lcCursor + "></VFPData>"
            ENDIF
            
            USE IN (lcCursor)
        ELSE
            THIS.cLastError = "Cursor was not created from query"
        ENDIF
        
        * Restore original directory
        CD (lcOldDir)
        
        RETURN lcXml
    ENDFUNC

    FUNCTION Ping()
        RETURN "OK:" + THIS.cVersion + " DBC:" + IIF(EMPTY(DBC()), "None", DBC())
    ENDFUNC

    FUNCTION GetLastError()
        RETURN THIS.cLastError
    ENDFUNC
    
    FUNCTION GetVersion()
        RETURN THIS.cVersion
    ENDFUNC
    
    * Enable/disable debug mode
    FUNCTION SetDebugMode(tlDebug)
        THIS.lDebugMode = tlDebug
        RETURN .T.
    ENDFUNC
    
    * Simple test to return table count
    FUNCTION GetTableCount()
        IF EMPTY(DBC())
            RETURN "No DBC open"
        ENDIF
        
        LOCAL lnCount
        SELECT COUNT(*) FROM DBC() WHERE objecttype = 'Table' INTO ARRAY laCount
        lnCount = laCount[1]
        RETURN "Found " + TRANSFORM(lnCount) + " tables in " + DBC()
    ENDFUNC
    
    * Return simple JSON-like string of tables
    FUNCTION GetTableListSimple()
        IF EMPTY(DBC())
            RETURN "[]"
        ENDIF
        
        LOCAL lcResult, lcSep
        lcResult = "["
        lcSep = ""
        
        SELECT objectname FROM DBC() WHERE objecttype = 'Table' ORDER BY objectname INTO CURSOR curTables
        IF USED("curTables")
            SELECT curTables
            SCAN
                lcResult = lcResult + lcSep + '"' + ALLTRIM(objectname) + '"'
                lcSep = ","
            ENDSCAN
            USE IN curTables
        ENDIF
        
        lcResult = lcResult + "]"
        RETURN lcResult
    ENDFUNC
    
    * Execute query and return simple JSON result
    FUNCTION QueryToJson(tcSql)
        LOCAL lcCursor, lcResult, lnError, lcError
        THIS.cLastError = ""
        
        IF EMPTY(DBC())
            THIS.cLastError = "No DBC open."
            RETURN '{"success":false,"error":"No DBC open"}'
        ENDIF
        
        IF VARTYPE(tcSql) # "C" OR EMPTY(tcSql)
            THIS.cLastError = "SQL text is required."
            RETURN '{"success":false,"error":"SQL required"}'
        ENDIF

        lcCursor = SYS(2015)
        lcResult = ""
        
        * Execute query
        lnError = 0
        lcError = ""
        ON ERROR lnError = ERROR()
        
        LOCAL lcCmd
        lcCmd = ALLTRIM(tcSql) + " INTO CURSOR " + lcCursor + " READWRITE"
        &lcCmd
        
        ON ERROR  && Reset error handler
        
        IF lnError > 0
            lcError = MESSAGE()
            THIS.cLastError = TRANSFORM(lnError) + ": " + lcError
            RETURN '{"success":false,"error":"' + STRTRAN(lcError,'"','') + '"}'
        ENDIF

        IF USED(lcCursor)
            SELECT (lcCursor)
            LOCAL lnRecCount, lnFields, i
            lnRecCount = RECCOUNT()
            lnFields = FCOUNT()
            
            * Build result
            lcResult = '{"success":true,"count":' + TRANSFORM(lnRecCount) + ',"data":['
            
            IF lnRecCount > 0
                LOCAL lcRow, lcField, lcValue, lnRow
                lnRow = 0
                
                SCAN
                    lnRow = lnRow + 1
                    IF lnRow > 1
                        lcResult = lcResult + ','
                    ENDIF
                    
                    lcRow = '{'
                    FOR i = 1 TO lnFields
                        IF i > 1
                            lcRow = lcRow + ','
                        ENDIF
                        
                        lcField = FIELD(i)
                        lcValue = ALLTRIM(TRANSFORM(EVALUATE(lcField)))
                        
                        * Handle special characters - order matters!
                        lcValue = STRTRAN(lcValue, '\', '\\')
                        lcValue = STRTRAN(lcValue, '"', '\"')
                        lcValue = STRTRAN(lcValue, CHR(13), '\r')
                        lcValue = STRTRAN(lcValue, CHR(10), '\n')
                        lcValue = STRTRAN(lcValue, CHR(9), '\t')
                        
                        lcRow = lcRow + '"' + lcField + '":"' + lcValue + '"'
                    ENDFOR
                    lcRow = lcRow + '}'
                    
                    lcResult = lcResult + lcRow
                ENDSCAN
            ENDIF
            
            lcResult = lcResult + ']}'
            
            USE IN (lcCursor)
        ELSE
            lcResult = '{"success":false,"error":"Query produced no cursor"}'
        ENDIF
        
        RETURN lcResult
    ENDFUNC
ENDDEFINE