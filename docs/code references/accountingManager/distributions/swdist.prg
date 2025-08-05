**************************************************
*-- Class Library:  swdist.vcx
**************************************************

**************************************************
*-- Class:        distproc (swdist.vcx)
*-- ParentClass:  custom
*-- BaseClass:    custom
*-- Time Stamp:   08/20/08 04:18:02 PM
*
DEFINE CLASS distproc AS CUSTOM


   HEIGHT = 37
   WIDTH  = 128
*-- Beginning Owner ID
   cbegownerid = "''"
*-- Ending Owner ID
   cendownerid = "''"
*-- production year to process
   cyear = "''"
*-- The well group to process
   cgroup = "'00'"
*-- The accounting date associated with this production period
   dacctdate = .F.
*-- Beginning Well ID
   cbegwellid = " "
*-- Ending Well ID
   cendwellid = "''"
   oregistry  = .NULL.
   omessage   = .NULL.
*-- Disb Mgr Batch Number for the closing
   cdmbatch = "''"
*-- Revenue Clearing Account
   crevclear = "''"
*-- Expense Clearing Account
   cexpclear = "''"
*-- Run Year or Production Year
   crunyear   = "''"
   csysctlkey = "''"

*-- Next Run Number
   nnewrunno = .F.

*-- Next Run Year
   cnewrunyear = ""

   ldebug      = .F.
   oprogress   = .F.
   nprogress   = .F.
   lreport     = .F.
   ldmpro      = .F.
   ddirectdate = {}
   lrpterror   = .F.
   lclosed     = .F.
   nseconds    = 0

*-- File Handle for Optimization file
   nfilehandle  = 0
   ndeftransfer = 0
   nmintransfer = 0
   NAME         = "distproc"

*-- Period to process
   cperiod = .F.

   companypost = .F.

*-- Who called us.
   cprocess = .F.

*-- Don't display progress bar when .T.
   lquiet = .F.

*-- .T. if the minimums should be released.
   lrelmin = .F.

*-- Glmaint object
   ogl = .NULL.

*-- Partnership Object
   oPartnerShip = .NULL.

*-- .T. if the period is being closed.
   lclose = .F.

*-- Saves the deleted setting
   cdeleted = .F.

*-- .T. = an error occurred in one of the methods.
   lerrorflag = .F.
   cErrorMsg  = "Check the system log for more information. "

*-- The accounting year associated with the accounting date passed.
   cacctyear = .F.

*-- The accounting period associated with the accounting date passed.
   cacctprd = .F.

*-- Run No Parameter
   nrunno = .F.

*-- Quarterly Wells Released
   lrelqtr = .F.

*-- Close JIBs separately from Revenue
   lsepclose = .F.

*-- .T. = this is the QuickBooks version
   lqbversion = .F.

*-- QBFC Request Object
   orequest = .F.

*-- .T. if we're processing compressors
   lcompressor = .F.

*-- Flat Rate Processing Flag.
   lflatrates = .F.

*-- Posting date for G/L
   dpostdate = {}

*-- Date for checks
   dCheckDate = {}

*-- Company's Share Post Date
   dCompanyShare = {}

*-- Date for processing revenue
   drevdate = {}

*-- Date for processing expenses.
   dexpdate = {}

*-- Posting Flags
   lAdvPosting = .F.

   lrunclosed      = .F.
   ndirectdeptotal = .F.
   osuspense       = .F.

* Closing Reports
   lRptWellExcpt = .F.
   lRptSuspense  = .F.
   lRptRegister  = .F.
   lRptUnalloc   = .F.

* Posting a run to QB
   lQBPost       = .F.

* Run is a new run
   lNewRun       = .F.

* Don't net
   lDontNet      = .F.

* Cancel Processing Flag
   lCanceled     = .F.

* Options object
   oOptions      = .NULL.

* Datasession
   nDataSession  = 1

* Wellinv Biz Object
   oWellInv      = .NULL.

*********************************
   PROCEDURE Init
*********************************
      LPARA tcbegid      AS CHARACTER ;
         , tcendid      AS CHARACTER ;
         , tcPeriod   AS CHARACTER ;
         , tcYear      AS CHARACTER ;
         , tcGroup      AS CHARACTER ;
         , tcprocess   AS CHARACTER ;
         , tdacctdate   AS DATE ;
         , tlquiet      AS Logical ;
         , tnRunNo      AS INTEGER ;
         , tlClose      AS Logical ;
         , tlNewRun   AS Logical

      LOCAL oDist

*************************************************************
* Parameters:
*
* tcBegID, tcEndID - Range of owners or wells to process
* tcPeriod         - Production period
* tcYear           - Production Year
* tcProcess        - "O" - owners, "W" - wells
* tdAcctDate       - Accounting Date
* tlQuiet          - .T. don't run progress bars
* tnRunNo          - Run Number being processed
* tlClose          - The Run is being closed
* tlNewRun         - The original run number was "0'
************************************************************

* Check the parameters to see if they're valid
      IF TYPE('tcBegID') # 'C'
         swselect('investor')
         SET ORDER TO cownerid
         GO TOP
         tcbegid = cownerid
      ENDIF
      IF TYPE('tcEndID') # 'C'
         swselect('investor')
         SET ORDER TO cownerid
         GO BOTT
         tcbegid = cownerid
      ENDIF
      IF TYPE('tcPeriod') # 'C'
         tcPeriod = PADL(TRANSFORM(MONTH(DATE())), 2, '0')
      ENDIF
      IF TYPE('tcYear') # 'C'
         tcYear = TRANSFORM(YEAR(DATE()))
      ENDIF
      IF TYPE('tcGroup') # 'C'
         tcGroup = '00'
      ENDIF
      IF TYPE('tcProcess') # 'C'
         tcprocess = 'O'
      ENDIF
      IF TYPE('tdAcctDate') # 'D'
         tdacctdate = DATE()
      ENDIF
      IF TYPE('tnRunNo') # 'N'
         tnRunNo = 0
      ENDIF

* Create the oOptions object
      swselect('options')
      GO TOP
      SCATTER NAME THIS.oOptions

      THIS.cdeleted   = SET('DELETED')
      THIS.lerrorflag = .F.

      SET DELETED ON
      SET MULTILOCKS ON

      THIS.dpostdate = {01/01/1980}

*  Setup the registry object
      THIS.oregistry = IIF(ISNULL(THIS.oregistry), findglobalobject('cmRegistry'), .NULL.)

*  Setup the message object
      THIS.omessage = IIF(ISNULL(THIS.omessage), findglobalobject('cmMessage'), .NULL.)

* Create the wellinv biz object
      THIS.oWellInv = CREATEOBJECT('swbizobj_wellinv')

*  Setup the parameters as properties so all methods can use them
      THIS.cperiod   = PADL(ALLT(STR(MONTH(tdacctdate))), 2, '0')
      THIS.cyear     = STR(YEAR(tdacctdate), 4)
      THIS.cprocess  = tcprocess
      THIS.cgroup    = tcGroup
      THIS.dacctdate = tdacctdate
      THIS.drevdate  = tdacctdate
      THIS.dexpdate  = tdacctdate

*  Are we supposed to display the progressbars?
      THIS.lquiet = tlquiet

*  What runno is this processing associated with
      THIS.nrunno   = tnRunNo
      THIS.crunyear = tcYear
      THIS.lNewRun  = tlNewRun

* If the runno passed is 0, get the next valid run number
      IF tnRunNo = 0
         IF EMPTY(tcYear)
            tcYear = THIS.cyear
         ENDIF
         THIS.nnewrunno   = getrunno(tcYear, .T., 'R')
         THIS.cnewrunyear = tcYear
         THIS.lNewRun     = .T.
      ELSE
         THIS.nnewrunno   = tnRunNo
         THIS.cnewrunyear = tcYear
      ENDIF

*  Open the tables needed
      IF TXNLEVEL() = 0
         IF m.goApp.lAMVersion
            swselect('glmaster', .T.)
            swselect('apopt')
            swselect('glopt')
            swselect('coabal', .T.)
         ENDIF
         swselect('wells')
         SET ORDER TO cWellID
         swselect('investor')
         SET ORDER TO cownerid
         swselect('options')
         swselect('wellhist', .T.)
         swselect('disbhist', .T.)
         swselect('ownpcts', .T.)
         swselect('checks', .T.)
         swselect('susaudit', .T.)
         swselect('roundtmp', .T.)
         swselect('one_man_tax', .T.)
         swselect('sysctl', .T.)
         SET ORDER TO yrprdgrp
         swselect('income', .T.)
         swselect('expense', .T.)
         swselect('sevtax')
         SET ORDER TO ctable
         swselect('expcat')
         SET ORDER TO ccatcode
         swselect('groups')
         SET ORDER TO cgroup
         swselect('programs')
         SET ORDER TO cprogcode
         swselect('vendor')
         SET ORDER TO cvendorid
         swselect('wellinv')
         swselect('stmtnote')
         swselect('revcat')
         swselect('suspense', .T.)
         swselect('arpmtdet', .T.)
         swselect('arpmthdr', .T.)
         IF NOT USED('wellhist1')
            USE (m.goApp.cdatafilepath + 'wellhist') AGAIN IN 0 ALIAS wellhist1
            SELECT wellhist1
            SET ORDER TO wellprd
         ENDIF
         IF NOT USED('disbhist1')
            USE (m.goApp.cdatafilepath + 'disbhist') IN 0 AGAIN ALIAS disbhist1
            SELE disbhist1
            = CURSORSETPROP("Buffering", 5)
         ENDIF
      ENDIF

*  Get clearing accounts
      swselect('glopt')
      THIS.crevclear = crevclear
      THIS.cexpclear = cexpclear

*  Instantiate the glmaint object
      THIS.ogl = CREATEOBJECT('glmaint')

*  Get the accounting year and period
      IF m.goApp.ldmpro
         THIS.cacctyear = STR(YEAR(tdacctdate), 4)
         THIS.cacctprd  = PADL(ALLT(STR(MONTH(tdacctdate), 2)), 2, '0')
      ELSE
         THIS.cacctyear = THIS.ogl.GetPeriod(tdacctdate, .T.)
         THIS.cacctprd  = THIS.ogl.GetPeriod(tdacctdate, .F.)
      ENDIF

*  Fill in the beginning and ending owner and well ids
      DO CASE
         CASE tcprocess = 'O'
            THIS.cbegownerid = tcbegid
            THIS.cendownerid = tcendid

*  Get the list of wells that these owners are in
            IF tcGroup = '**'
               SELECT  cWellID ;
                   FROM wellinv ;
                   WHERE BETWEEN(cownerid, tcbegid, tcendid) ;
                   INTO CURSOR welltemp ;
                   ORDER BY cWellID ;
                   GROUP BY cWellID
            ELSE
               IF THIS.lNewRun OR tlClose
                  SELECT  cWellID ;
                      FROM wellinv ;
                      WHERE BETWEEN(cownerid, tcbegid, tcendid) ;
                          AND cWellID IN (SELECT  cWellID ;
                                              FROM wells ;
                                              WHERE cgroup = tcGroup) ;
                      INTO CURSOR welltemp ;
                      ORDER BY cWellID ;
                      GROUP BY cWellID
               ELSE
                  SELECT  cWellID ;
                      FROM wellinv ;
                      WHERE BETWEEN(cownerid, tcbegid, tcendid) ;
                      INTO CURSOR welltemp ;
                      ORDER BY cWellID ;
                      GROUP BY cWellID
               ENDIF
            ENDIF

            SELECT welltemp
            GO TOP
            THIS.cbegwellid = cWellID
            GO BOTT
            THIS.cendwellid = cWellID

*  Put the list of owners in the owner temp cursor
            SELECT  cownerid ;
                FROM investor ;
                WHERE BETWEEN(cownerid, tcbegid, tcendid) ;
                INTO CURSOR owntemp ;
                ORDER BY cownerid ;
                GROUP BY cownerid

         CASE tcprocess = 'W'

            THIS.cbegwellid = tcbegid
            THIS.cendwellid = tcendid

*  Put the list of wells in the welltemp cursorF
            IF tcGroup = '**'
               SELECT  cWellID ;
                   FROM wells ;
                   WHERE BETWEEN(cWellID, tcbegid, tcendid) ;
                   INTO CURSOR welltemp ;
                   ORDER BY cWellID ;
                   GROUP BY cWellID
            ELSE
               IF tnRunNo = 0 OR tlClose
                  SELECT  cWellID ;
                      FROM wells ;
                      WHERE BETWEEN(cWellID, tcbegid, tcendid) ;
                          AND cgroup = tcGroup ;
                      INTO CURSOR welltemp ;
                      ORDER BY cWellID ;
                      GROUP BY cWellID
               ELSE
                  SELECT  cWellID ;
                      FROM wells ;
                      WHERE BETWEEN(cWellID, tcbegid, tcendid) ;
                      INTO CURSOR welltemp ;
                      ORDER BY cWellID ;
                      GROUP BY cWellID
               ENDIF
            ENDIF

            SELECT  cownerid ;
                FROM wellinv ;
                WHERE BETWEEN(cWellID, tcbegid, tcendid) ;
                INTO CURSOR owntemp ;
                ORDER BY cownerid ;
                GROUP BY cownerid

            SELECT owntemp
            GO TOP
            THIS.cbegownerid = cownerid
            GO BOTT
            THIS.cendownerid = cownerid
      ENDCASE

      IF NOT tlquiet
         WAIT CLEAR
      ENDIF

      RETURN DODEFAULT()
   ENDPROC

*********************************
   PROCEDURE Destroy
*********************************
      LOCAL lcDeleted

      lcDeleted = THIS.cdeleted

      SET DELETED &lcDeleted

* Reset ESC key
      ON KEY LABEL ESC

      oOBJ = THIS.ogl
      RELEASE oOBJ
      THIS.ogl = .NULL.

      oOBJ = THIS.oregistry
      RELEASE oOBJ
      THIS.oregistry = .NULL.

      oOBJ = THIS.omessage
      RELEASE oOBJ
      THIS.omessage = .NULL.

      IF VARTYPE(THIS.oprogress) = 'O'
         oOBJ = THIS.oprogress
         RELEASE oOBJ
         THIS.oprogress.CloseProgress()
      ENDIF

      IF VARTYPE(THIS.osuspense) = 'O'
         oOBJ = THIS.osuspense
         RELEASE oOBJ
         THIS.osuspense = .NULL.
      ENDIF

      IF VARTYPE(THIS.oWellInv) = 'O'
         oOBJ = THIS.oWellInv
         RELEASE oOBJ
         THIS.oWellInv = .NULL.
      ENDIF

* Clear all external object refs
      lnProperties = AMEMBERS(laProps, THIS, 0, 'U')
      FOR x = 1 TO lnProperties
         lcProp = laProps[x]
         IF LEFT(laProps[x], 1) = 'O'
            lcProperty  = 'THIS.' + laProps[x]
            &lcProperty = .NULL.
         ENDIF
      ENDFOR

      DODEFAULT()
   ENDPROC

*********************************
   PROCEDURE Error
*********************************
      LPARAMETERS nError, cMethod, nLine
      LOCAL lnLevel, lnX

      THIS.lerrorflag = .T.

      = AERROR(gaerrors)
      THIS.cErrorMsg = 'Error: ' + TRANSFORM(nError) + CHR(10) + ;
         'Msg: ' + gaerrors[3] + CHR(10) + ;
         'Method: ' + cMethod + CHR(10) + ;
         'Line: ' + TRANSFORM(nLine)
      DO errorlog WITH m.cMethod, m.nLine, 'DISTPROC', m.nError, gaerrors[3]

      IF BETWEEN(m.nError, 1426, 1429) AND m.goApp.lqbversion
         m.goApp.oQB.QBError()
      ELSE
         lnLevel = TXNLEVEL()

         IF lnLevel > 0
            FOR lnX = 1 TO lnLevel
               ROLLBACK
            ENDFOR
         ENDIF

         m.goApp.ERROR(nError, cMethod, nLine)
      ENDIF
   ENDPROC

*********************************
   PROCEDURE Setup
*********************************
*
* Creates the Wellwork and Invtmp cursors for holding wellhist disbhist records
* for the run being processed.
*
      LOCAL lcRunYear, llReturn, loError
      LOCAL latemp[1], latempx[1], latempy[1], lnX, lnflatg, lnflato, lny, lprognet, lroysevtx
      LOCAL Y, cWellID, cgroup, cperiod, cprogcode, crectype, crunyear, cyear, hperiod, hyear, nprocess
      LOCAL nrunno, x

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


* Setup the runyear, runno variable for the select statements
         lcRunYear = THIS.crunyear + PADL(TRANSFORM(THIS.nrunno), 3, '0')

         SET DELETED ON

         IF TYPE('this.dRevDate') # 'D'
            THIS.drevdate = THIS.dacctdate
         ENDIF

         IF TYPE('this.dExpDate') # 'D'
            THIS.dexpdate = THIS.dacctdate
         ENDIF

         m.goApp.Rushmore(.T., 'Setup')

*  Check to see if the quarterly wells should be released
         THIS.CheckForQuarterlyWells()

         THIS.oprogress.SetProgressMessage('Checking Division of Interests....')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

* Check the DOI for owner with multiple interests of the same type in the same well
* Combine them if any are found.
         llReturn = CombineDOI()

* Prepare the wellwork cursor
         THIS.oprogress.SetProgressMessage('Preparing temporary well history...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

         IF THIS.lrelqtr
            SELE income.cWellID, income.cyear, income.cperiod, cdeck, ;
               wells.cgroup, ;
               THIS.nrunno, ;
               THIS.crunyear AS crunyear, ;
               wells.lroysevtx  AS lroysevtx, ;
               wells.nprocess   AS nprocess  ;
               FROM income, wells ;
               WHERE ((income.nrunno = 0 AND income.drevdate <= THIS.drevdate) OR (income.nrunno = THIS.nrunno AND income.crunyear = THIS.crunyear)) ;
               AND income.cWellID = wells.cWellID ;
               AND IIF(THIS.lNewRun OR THIS.lclose, wells.cgroup = THIS.cgroup, .T.) ;
               AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U') ;
               INTO CURSOR tempinc ;
               ORDER BY income.cdeck, income.cWellID, cyear, cperiod ;
               GROUP BY income.cdeck, income.cWellID, cyear, cperiod

            SELE expense.cWellID, expense.cyear, expense.cperiod, cdeck, ;
               wells.cgroup, ;
               THIS.nrunno, ;
               THIS.crunyear AS crunyear, ;
               wells.lroysevtx  AS lroysevtx, ;
               wells.nprocess   AS nprocess  ;
               FROM expense, wells ;
               WHERE ((expense.nRunNoRev = 0 AND expense.dexpdate <= THIS.dexpdate) OR (expense.nRunNoRev = THIS.nrunno AND expense.cRunYearRev = THIS.crunyear)) ;
               AND expense.cyear # 'FIXD' ;
               AND expense.cWellID = wells.cWellID ;
               AND IIF(THIS.lNewRun OR THIS.lclose, wells.cgroup = THIS.cgroup, .T.) ;
               AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U') ;
               INTO CURSOR tempexp ;
               ORDER BY expense.cdeck, expense.cWellID, cyear, cperiod ;
               GROUP BY expense.cdeck, expense.cWellID, cyear, cperiod
         ELSE
            SELE income.cWellID, income.cyear, income.cperiod, cdeck, ;
               wells.cgroup, ;
               THIS.nrunno, ;
               THIS.crunyear AS crunyear, ;
               wells.lroysevtx  AS lroysevtx, ;
               wells.nprocess   AS nprocess  ;
               FROM income, wells ;
               WHERE ((income.nrunno = 0 AND income.drevdate <= THIS.drevdate) OR (income.nrunno = THIS.nrunno AND income.crunyear = THIS.crunyear)) ;
               AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U') ;
               AND IIF(THIS.lNewRun OR THIS.lclose, wells.cgroup = THIS.cgroup, .T.) ;
               AND IIF(THIS.lNewRun OR THIS.lclose, wells.nprocess # 2, .T.) ;
               AND income.cWellID = wells.cWellID ;
               INTO CURSOR tempinc ;
               ORDER BY income.cdeck, income.cWellID, cyear, cperiod ;
               GROUP BY income.cdeck, income.cWellID, cyear, cperiod

            SELE expense.cWellID, expense.cyear, expense.cperiod, cdeck, ;
               wells.cgroup, ;
               THIS.nrunno, ;
               THIS.crunyear AS crunyear, ;
               wells.lroysevtx  AS lroysevtx, ;
               wells.nprocess   AS nprocess  ;
               FROM expense, wells ;
               WHERE ((expense.nRunNoRev = 0 AND expense.dexpdate <= THIS.dexpdate) OR (expense.nRunNoRev = THIS.nrunno AND expense.cRunYearRev = THIS.crunyear)) ;
               AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U') ;
               AND IIF(THIS.lNewRun OR THIS.lclose, wells.cgroup = THIS.cgroup, .T.) ;
               AND IIF(THIS.lNewRun OR THIS.lclose, wells.nprocess # 2, .T.) ;
               AND expense.cyear # 'FIXD' ;
               AND expense.cWellID = wells.cWellID ;
               INTO CURSOR tempexp ;
               ORDER BY expense.cdeck, expense.cWellID, cyear, cperiod ;
               GROUP BY expense.cdeck, expense.cWellID, cyear, cperiod
         ENDIF

         IF THIS.lflatrates  && Process Flat rates this run
            THIS.oprogress.SetProgressMessage('Processing Flat Rates...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
            INKEY(25)
* Get the flat rates to be processed this run
            CREATE CURSOR tempflat ;
               (cWellID     c(10), ;
                 cdeck       c(10), ;
                 cyear       c(4), ;
                 cperiod     c(2), ;
                 cgroup      c(2), ;
                 nrunno      i, ;
                 crunyear    c(4), ;
                 lroysevtx   L, ;
                 nprocess    i)

            SELE wells
            SCAN FOR BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) AND cgroup = THIS.cgroup
               m.cWellID   = cWellID
               m.lroysevtx = lroysevtx
               m.nprocess  = nprocess
               IF NOT THIS.lrelqtr
                  IF m.nprocess = 2
                     LOOP
                  ENDIF
               ENDIF
               m.nrunno   = 0
               m.crunyear = ALLT(STR(YEAR(THIS.dacctdate)))
               m.cyear    = m.crunyear
               m.cperiod  = PADL(ALLT(STR(MONTH(THIS.dacctdate))), 2, '0')
               m.cgroup   = cgroup
               m.cdeck    = THIS.oWellInv.DOIDeckNameLookup(m.cyear, m.cperiod, m.cWellID)
               lnflato    = THIS.getFlatAmt(m.cWellID, 'O', m.cdeck)
               lnflatg    = THIS.getFlatAmt(m.cWellID, 'G', m.cdeck)
               IF lnflato + lnflatg > 0
                  INSERT INTO tempflat FROM MEMVAR
               ENDIF
            ENDSCAN
         ENDIF

*   Create Temp Well Production History File
         swselect('wellhist')
         Make_Copy('wellhist', 'wellwork')

* Create wellwork records for wells that have revenue to be processed
         SELECT tempinc
         SCAN
            SCATTER MEMVAR

            IF EMPTY(m.cdeck)
* If the deck isn't specified look it up based on the prod year and period
               m.cdeck = THIS.oWellInv.DOIDeckNameLookup(m.cyear, m.cperiod, m.cWellID)
               SELECT wellwork
               LOCATE FOR hyear = m.cyear ;
                  AND hperiod = m.cperiod ;
                  AND cWellID = m.cWellID ;
                  AND cdeck = m.cdeck
               IF FOUND()
* This deck and period already exists in wellwork
                  LOOP
               ENDIF
            ENDIF
            m.nrunno   = THIS.nrunno
            m.hyear    = m.cyear
            m.hperiod  = m.cperiod
            m.crectype = 'R'
            INSERT INTO wellwork FROM MEMVAR
         ENDSCAN

         SELECT wellwork
         INDEX ON cWellID + hyear + hperiod + crectype TAG wellprd

* Create wellwork records for wells that have expenses to be processed
         SELECT wellwork
         SET ORDER TO wellprd
         SELECT tempexp
         SCAN
            SCATTER MEMVAR

            IF EMPTY(m.cdeck)
* If the deck isn't specified look it up based on the prod year and period
               m.cdeck = THIS.oWellInv.DOIDeckNameLookup(m.cyear, m.cperiod, m.cWellID)
            ENDIF
            m.nrunno   = THIS.nrunno
            m.hyear    = m.cyear
            m.hperiod  = m.cperiod
            m.crectype = 'R'
            SELE wellwork
            LOCATE FOR hyear = m.cyear ;
               AND hperiod = m.cperiod ;
               AND cWellID = m.cWellID ;
               AND cdeck == m.cdeck
            IF NOT FOUND()
               INSERT INTO wellwork FROM MEMVAR
            ENDIF
         ENDSCAN

         SELECT wellwork
         INDEX ON cWellID TAG cWellID
         SET ORDER TO cWellID

* Create wellwork records for wells that have flat rates to be processed
         IF THIS.lflatrates
            SELE tempflat
            SCAN
               SCATTER MEMVAR

               IF EMPTY(m.cdeck)
* If the deck isn't specified look it up based on the prod year and period
                  m.cdeck = THIS.oWellInv.DOIDeckNameLookup(m.cyear, m.cperiod, m.cWellID)
               ENDIF
               m.nrunno   = THIS.nrunno
               m.hyear    = m.cyear
               m.hperiod  = m.cperiod
               m.crectype = 'R'
               SELE wellwork
               LOCATE FOR cWellID = m.cWellID ;
                  AND cdeck == m.cdeck ;
                  AND hyear = m.hyear ;
                  AND hperiod = m.hperiod
               IF NOT FOUND()
                  INSERT INTO wellwork FROM MEMVAR
               ENDIF
            ENDSCAN
         ENDIF

         SELECT wellwork
         SCATTER MEMVAR BLANK
* Add in missing wells with no production so that frequency suspense can be released
         SELE wells
         SCAN FOR BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) AND cgroup = THIS.cgroup
            m.cWellID   = cWellID
            m.lroysevtx = lroysevtx
            m.nprocess  = nprocess
            m.nrunno    = THIS.nrunno
            m.crunyear  = ALLT(STR(YEAR(THIS.dacctdate)))
            m.hdate     = THIS.dacctdate
            m.hyear     = m.crunyear
            m.hperiod   = PADL(ALLT(STR(MONTH(THIS.dacctdate))), 2, '0')

            IF EMPTY(m.cdeck)
               m.cdeck = THIS.oWellInv.DOIDeckNameLookup(m.hyear, m.hperiod, m.cWellID)
            ENDIF
            m.cgroup   = cgroup
            m.crectype = 'R'
            SELECT wellwork
            LOCATE FOR cWellID == m.cWellID ;
               AND hyear == m.cyear ;
               AND hperiod == m.cperiod ;
               AND cdeck == m.cdeck
            IF NOT FOUND()
               INSERT INTO wellwork FROM MEMVAR
            ENDIF
         ENDSCAN

         SET SAFETY OFF

         SELECT wellwork
         INDEX ON hyear + hperiod TAG yearprd

*  Build invtmp table
         THIS.oprogress.SetProgressMessage('Preparing temporary owner history....')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

* Create Temp Investor Disbursement File
* Using suspense since it has all the interest fields in it
         swselect('suspense')
         llReturn = Make_Copy('suspense', 'invtmp')

         IF NOT llReturn
            EXIT
         ENDIF

         IF THIS.lrunclosed
* Create the cursor based on the history for the run
            SELECT  THIS.nrunno   AS nrunno, ;
                    THIS.crunyear AS crunyear, ;
                    {^1980-01-01} AS hdate, ;
                    disbhist.cownerid, ;
                    disbhist.cdeck, ;
                    disbhist.cWellID, ;
                    THIS.cgroup AS cgroup, ;
                    ownpcts.nworkint, ;
                    disbhist.ciddisb AS cidwinv, ;
                    ownpcts.nintclass1, ;
                    ownpcts.nintclass2, ;
                    ownpcts.nintclass3, ;
                    ownpcts.nintclass4, ;
                    ownpcts.nintclass5, ;
                    ownpcts.nacpint, ;
                    ownpcts.nbcpint, ;
                    ownpcts.nrevoil, ;
                    ownpcts.nrevgas, ;
                    ownpcts.nrevtrp, ;
                    ownpcts.nrevoth, ;
                    ownpcts.nrevtax1, ;
                    ownpcts.nrevtax2, ;
                    ownpcts.nrevtax3, ;
                    ownpcts.nrevtax4, ;
                    ownpcts.nrevtax5, ;
                    ownpcts.nrevtax6, ;
                    ownpcts.nrevtax7, ;
                    ownpcts.nrevtax8, ;
                    ownpcts.nrevtax9, ;
                    ownpcts.nrevtax10, ;
                    ownpcts.nrevtax11, ;
                    ownpcts.nrevtax12, ;
                    disbhist.ctypeinv, ;
                    disbhist.ctypeint, ;
                    disbhist.cdirect, ;
                    disbhist.lflat, ;
                    disbhist.nflatrate, ;
                    disbhist.cflatstart, ;
                    disbhist.nflatfreq, ;
                    disbhist.cprogcode, ;
                    ownpcts.nrevmisc1, ;
                    ownpcts.nrevmisc2, ;
                    disbhist.lJIB, ;
                    disbhist.lonhold, ;
                    disbhist.ciddisb, ;
                    disbhist.ndisbfreq, ;
                    disbhist.ntaxpct, ;
                    disbhist.nplugpct, ;
                    'R'  AS crectype, ;
                    .F.  AS lused, ;
                    .F.  AS lprognet, ;
                    000000.00 AS nbbltot, ;
                    000000.00 AS nmcftot, ;
                    000000.00 AS nothtot, ;
                    000000.00 AS nIncome, ;
                    000000.00 AS ngasrev, ;
                    000000.00 AS noilrev, ;
                    000000.00 AS ntrprev, ;
                    000000.00 AS nmiscrev1, ;
                    000000.00 AS nmiscrev2, ;
                    000000.00 AS nexpense, ;
                    000000.00 AS ntotale1, ;
                    000000.00 AS ntotale2, ;
                    000000.00 AS ntotale3, ;
                    000000.00 AS ntotale4, ;
                    000000.00 AS ntotale5, ;
                    000000.00 AS ntotalea, ;
                    000000.00 AS ntotaleb, ;
                    000000.00 AS nnetcheck, ;
                    000000.00 AS nsevtaxes, ;
                    disbhist.lhold, ;
                    disbhist.lprogram, ;
                    disbhist.creason, ;
                    disbhist.nrunno_in, ;
                    disbhist.crunyear_in ;
                FROM disbhist,;
                    ownpcts,;
                    wells ;
                WHERE disbhist.crunyear + PADL(TRANSFORM(disbhist.nrunno), 3, '0') = lcRunYear ;
                    AND disbhist.cownerid IN (SELECT  cownerid ;
                                                  FROM owntemp) ;
                    AND disbhist.cWellID IN (SELECT  cWellID ;
                                                 FROM welltemp) ;
                    AND disbhist.cWellID == wells.cWellID ;
                    AND NOT INLIST(wells.cwellstat, 'I', 'V', 'S', 'P', 'U') ;
                    AND disbhist.ciddisb == ownpcts.ciddisb ;
                INTO CURSOR invtmpx ;
                ORDER BY disbhist.cownerid,;
                    disbhist.cWellID


* Append records to the owner work cursor
            SELE wellwork
            SCAN
               m.cWellID = cWellID
               m.hyear   = hyear
               m.hperiod = hperiod
               SELE invtmpx
               SCAN FOR cWellID = m.cWellID
                  SCATTER MEMVAR
                  INSERT INTO invtmp FROM MEMVAR
               ENDSCAN
            ENDSCAN
         ELSE

* Build a list of wells and production periods
            SELECT  cWellID, ;
                    hyear, ;
                    hperiod, ;
                    cdeck ;
                FROM wellwork ;
                INTO CURSOR prodperiods ;
                ORDER BY cWellID,;
                    hyear,;
                    hperiod,;
                    cdeck ;
                GROUP BY cWellID,;
                    hyear,;
                    hperiod,;
                    cdeck

            SELECT investor
            SET ORDER TO cownerid
            SELECT wells
            SET ORDER TO cWellID

* Append records to the owner work cursor
            SELE wellwork
            SCAN
               m.cWellID = cWellID
               m.hyear   = hyear
               m.hperiod = hperiod
               m.cdeck   = cdeck
               STORE 0 TO m.nbbltot, m.nmcftot, m.nothtot, m.nIncome, m.ngasrev, m.noilrev, m.ntrprev, m.nmiscrev1, m.nmiscrev2, ;
                  m.nexpense, m.ntotale1, m.ntotale2, m.ntotale3, m.ntotale4, m.ntotale4, m.ntotale5, m.ntotalea, m.ntotaleb, ;
                  m.nnetcheck, m.nsevtaxes, m.nrunno_in
               m.crunyear_in = ''
               m.cgroup      = THIS.cgroup
               m.crectype    = 'R'
               m.hdate       = {1/1/1980}
               STORE .F. TO m.lused, m.lprognet

               SELECT welltemp
               LOCATE FOR cWellID = m.cWellID
               IF NOT FOUND()
                  LOOP
               ENDIF
               SELECT wells
               IF SEEK(m.cWellID)
                  IF INLIST(wells.cwellstat, 'I', 'V', 'S', 'P', 'U')
                     LOOP
                  ENDIF
               ENDIF

               IF USED('invtmp')
* Get the current deck for this well
                  IF EMPTY(m.cdeck)
                     m.cdeck = THIS.oWellInv.DOIDeckNameLookup(m.hyear, m.hperiod, m.cWellID)
                  ENDIF
                  SELECT wellinv
                  LOCATE FOR cWellID = m.cWellID ;
                     AND cdeck   == m.cdeck
                  IF FOUND()
                     SCAN FOR cWellID = m.cWellID AND cdeck == m.cdeck
                        SCATTER MEMVAR
                        m.ciddisb = m.cidwinv
                        SELECT owntemp
                        LOCATE FOR cownerid = m.cownerid
                        IF NOT FOUND()
                           LOOP
                        ENDIF
                        SELECT investor
                        IF SEEK(m.cownerid)
                           m.lhold      = lhold
                           m.ldirectdep = ldirectdep
                           m.ndisbfreq  = ndisbfreq
                           m.nrunno     = THIS.nrunno
                           m.crunyear   = THIS.crunyear
                           SELECT invtmp
                           LOCATE FOR cownerid = m.cownerid ;
                              AND ctypeinv = m.ctypeinv ;
                              AND cdeck    == m.cdeck ;
                              AND hyear    = m.hyear ;
                              AND hperiod  = m.hperiod ;
                              AND cWellID  = m.cWellID
                           IF NOT FOUND()
                              INSERT INTO invtmp FROM MEMVAR
                           ENDIF
                        ENDIF
                     ENDSCAN
                  ELSE
* No deck so don't include this well. It wouldn't be processed anyway
                     LOOP
                  ENDIF
               ENDIF
            ENDSCAN

         ENDIF

         SELECT invtmp
         INDEX ON cownerid + cprogcode + cWellID + ctypeinv + hyear + hperiod TAG invprog
         INDEX ON cprogcode TAG cprogcode
         INDEX ON hyear + hperiod TAG yearprd
         INDEX ON cownerid + cprogcode + cWellID + ctypeinv TAG ownertype
         INDEX ON cownerid + cWellID TAG invwell
         INDEX ON cWellID TAG cWellID
         INDEX ON csusptype TAG csusptype

*  Flag the programs as being netted outside the program or not
         swselect('programs')
         SET ORDER TO cprogcode

         SELECT invtmp
         SCAN FOR NOT EMPTY(cprogcode)
            m.cprogcode = cprogcode
            SELECT programs
            IF SEEK(m.cprogcode)
               m.lprognet = lprognet
            ELSE
               m.lprognet = .F.
            ENDIF
            SELECT invtmp
            REPLACE lprognet WITH m.lprognet
         ENDSCAN

* Remove any non quarterly wells if they ended up in the cursors
         IF NOT THIS.lrelqtr
            DELETE FROM invtmp WHERE invtmp.cWellID IN (SELECT cWellID FROM wells WHERE nprocess = 2)
            DELETE FROM wellwork WHERE wellwork.cWellID IN (SELECT cWellID FROM wells WHERE nprocess = 2)
         ENDIF

         WAIT CLEAR
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'Setup', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, '', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('Setup', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Main Processing routine
*********************************
   PROCEDURE Main
*********************************
      LPARA tlClose      AS CHARACTER ;
         , tlRpt      AS Logical

* tlClose  = Closing the run
* tlRpt    = Produce the rounding report

      LOCAL llNetDef, llClosed, llPeriodClosed
      LOCAL llReturn, lnwells, loError

      llReturn = .T.

      _VFP.AUTOYIELD = .F.
      SET ESCAPE ON

* Setup the ability to cancel processing
      ON ESCAPE m.goApp.lCanceled = .T.

      TRY
* Save the parameters
         THIS.lclose   = tlClose
         THIS.lreport  = tlRpt
         THIS.nseconds = DATETIME()

         llClosed       = .F.
         llPeriodClosed = .F.

         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


         IF tlClose
            THIS.lquiet    = .F.
            THIS.oprogress = THIS.omessage.progressbarex('Processing Revenue Closing for Run: ' + ;
                 PADL(TRANSFORM(THIS.nrunno), 3, '0') + '/' + THIS.crunyear, ' ')
         ELSE
            THIS.oprogress = THIS.omessage.progressbarex('Processing Request...', ' ')
         ENDIF

         THIS.nprogress = 1

* Check to see if this Run is closed
         THIS.lrunclosed = THIS.CheckHistRun()

* Set up the process bar based on which app is processing. They have a different # of steps

            IF THIS.lrunclosed
               THIS.oprogress.SetProgressRange(0, 8)
            ELSE
               THIS.oprogress.SetProgressRange(0, 53)
            ENDIF
          
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

         THIS.lsepclose = THIS.oOptions.lsepclose

         THIS.oprogress.SetProgressMessage('Creating work files...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

*
*  Create the work cursors (wellwork) and (invtmp)
*
         IF NOT THIS.lerrorflag AND NOT THIS.lrunclosed
            llReturn = THIS.SETUP()
            IF llReturn = .F.
               EXIT
            ENDIF
         ENDIF
         
         

****************************************************************************
*  Check to see if the run has already processed
*  If the run is closed, retrieve the information from the history
*  files.
****************************************************************************
         IF THIS.lrunclosed = .T.
            llClosed = .T.
            swselect('groups')
            SET ORDER TO cgroup
            IF SEEK(THIS.cgroup)
               llNetDef = lNetDef
            ELSE
               llNetDef = .T.
            ENDIF
            llReturn = THIS.GetHist()
            IF llReturn = .F.
               EXIT
            ENDIF
         ELSE
            llClosed = .F.
            swselect('groups')
            SET ORDER TO cgroup
            IF SEEK(THIS.cgroup)
               llNetDef = lNetDef
            ELSE
               llNetDef = .T.
            ENDIF

*
*  Allocate the revenue and expenses to the wells for this run
*
            IF NOT THIS.lerrorflag
               llReturn = THIS.WellProc()
               IF llReturn = .F.
                  EXIT
               ENDIF
            ENDIF

*
*  Allocate the revenue and expenses to the owners this run
*
            IF NOT THIS.lerrorflag
               llReturn = THIS.ownerproc()
               IF llReturn = .F.
                  EXIT
               ENDIF
            ENDIF

*
*  Calculate the rounding and allocate it to the rounding owners
*
            IF NOT THIS.lerrorflag AND tlRpt
               THIS.CalcRounding(tlRpt)
            ENDIF

*
* If we're closing the run call the closing processing
*
            IF tlClose AND NOT THIS.lerrorflag
               llReturn = THIS.closeproc()
               IF (llReturn AND NOT THIS.lerrorflag) OR THIS.lclosed
                  MESSAGEBOX('The revenue run processed successfully.', 64, 'Revenue Run Closing')
               ELSE
                  IF THIS.lCanceled
                     THIS.cErrorMsg = 'Processing canceled by user.'
                  ENDIF
                  MESSAGEBOX('The revenue run did not process successfully. All files have been reset.' + CHR(10) + ;
                       THIS.cErrorMsg, 16, 'Run Closing Error')
               ENDIF
            ELSE
*
*  Process suspense entries
*  Only process them for owner processes
*
               IF NOT THIS.lerrorflag
                  IF THIS.cprocess = 'O'
                     llReturn = THIS.suspense(llNetDef)
                     IF NOT llReturn
                        EXIT
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF
         ENDIF

         IF NOT tlClose
            THIS.oprogress.CloseProgress()
         ENDIF
         llReturn = .T.

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'Main', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('Main', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      IF VARTYPE(THIS.oprogress) = 'O'
         THIS.oprogress.CloseProgress()
         THIS.oprogress = .NULL.
      ENDIF

      ON ESCAPE

      RETURN llReturn
   ENDPROC

*-- Allocate the revenue and expenses to the wells
*********************************
   PROCEDURE WellProc
******************************
      LOCAL lnCount, lcSavePrd1, lcSavePrd2, lcYear, lnRunNo
      LOCAL lnCompress, lnGather, llSepClose, oprogress, lcType, lnTax
      LOCAL lACPInt, lBCPInt, lIntClass1, lIntClass2, lIntClass3, lIntClass4, lIntClass5, laArray[1]
      LOCAL laArrayg[1], laArrayo[1], lcProdPeriod, lcProdYear, lcWellFlat, ldAcctDate, ldPostDate
      LOCAL llReturn, lnAmount, lnDaysOn, lnGlobalCompress, lnGlobalGather, lnMarketing, lnMax, lnMktAmt
      LOCAL lnNETMCF, lnNetBBL, lnNetGasAmt, lnNetMisc1, lnNetMisc2, lnNetOilAmt, lnNetOthAmt, lnNetTax
      LOCAL lnNetTrpAmt, loError
      LOCAL  bbltot, cMethodBBL1, cMethodBBL2, cMethodBBL3, cMethodBBL4, cMethodMCF1, cMethodMCF2
      LOCAL  cMethodMCF3, cMethodMCF4, cMethodOTH1, cMethodOTH2, cMethodOTH3, cMethodOTH4, cownerid
      LOCAL  cWellID, cexpclass, crunyear, hperiod, hyear, mcftot, nBBLTax1, nBBLTax2, nBBLTax3
      LOCAL  nBBLTax4, nBBLTaxr, nBBLTaxw, nCompress, nExpgas, nExpoil, nGBBLTax1, nGBBLTax2, nGBBLTax3
      LOCAL  nGBBLTax4, nGMCFTax1, nGMCFTax2, nGMCFTax3, nGMCFTax4, nGOTHTax1, nGOTHTax2, nGOTHTax3
      LOCAL  nGOTHTax4, nGasInc, nGather, nMCFTax1, nMCFTax2, nMCFTax3, nMCFTax4, nMCFTaxr, nMCFTaxw
      LOCAL  nMiscinc1, nMiscinc2, nNetExp, nOthInc, nOthTax1, nOthTax2, nOthTax3, nOthTax4, nTaxBBL3
      LOCAL  nTotBBL, nTotMCF, nTotMKTG, nTotMisc1, nTotMisc2, nTotProd, nTotale, nTrpInc, nexpcl1
      LOCAL  nexpcl2, nexpcl3, nexpcl4, nexpcl5, nexpclA, nexpclB, nflatgas, nflatoil, ngProdTax3
      LOCAL  ngProdtax1, ngProdtax2, ngProdtax4, ngasint, ngaspct, ngastax1, ngastax1a, ngastax1b
      LOCAL  ngastax2, ngastax2a, ngastax2b, ngastax3, ngastax3a, ngastax3b, ngastax4, ngastax4a
      LOCAL  ngastax4b, nggastax1, nggastax2, nggastax3, nggastax4, ngoiltax1, ngoiltax2, ngoiltax3
      LOCAL  ngoiltax4, noilinc, noilint, noilpct, noiltax1, noiltax1a, noiltax1b, noiltax2, noiltax2a
      LOCAL  noiltax2b, noiltax3, noiltax3a, noiltax3b, noiltax4, noiltax4a, noiltax4b, nprocess
      LOCAL  nprodtax1, nprodtax1a, nprodtax1b, nprodtax2, nprodtax2a, nprodtax2b, nprodtax3, nprodtax3a
      LOCAL  nprodtax3b, nprodtax4, nprodtax4a, nprodtax4b, nprodwell, nroyintg, nroyinto, nrunno
      LOCAL  ntaxbbl1, ntaxbbl2, ntaxbbl4, ntaxmcf1, ntaxmcf2, ntaxmcf3, ntaxmcf4, ntaxoth1, ntaxoth2
      LOCAL  ntaxoth3, ntaxoth4, ntotsalt, nwrkintg, nwrkinto

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Allocating Revenue and Expenses to Wells...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnCount    = 1
         lnRunNo    = THIS.nrunno
         lcSavePrd1 = THIS.cperiod
         lcSavePrd2 = lcSavePrd1
         ldAcctDate = THIS.dacctdate
         ldPostDate = THIS.dpostdate

* Create cursors for calculated compression & gathering
         CREATE CURSOR compcalc ;
            (cWellID      c(10), ;
              cyear        c(4), ;
              cperiod      c(2), ;
              namount      N(12, 2))

         CREATE CURSOR gathcalc ;
            (cWellID      c(10), ;
              cyear        c(4), ;
              cperiod      c(2), ;
              namount      N(12, 2))


* Get the default compression and gathering charges
         lnGlobalCompress = THIS.oOptions.nCompress
         lnGlobalGather   = THIS.oOptions.nGather
         llSepClose       = .T.

         IF TYPE('ldAcctDate') # 'D'
            ldAcctDate = DATE()
         ENDIF

         CREATE CURSOR temptax ;
            (hyear      c(4), ;
              hperiod    c(2), ;
              cWellID    c(10), ;
              cdeck      c(10), ;
              noiltax1   N(9, 2), ;
              noiltax2   N(9, 2), ;
              noiltax3   N(9, 2), ;
              noiltax4   N(9, 2), ;
              ngastax1   N(9, 2), ;
              ngastax2   N(9, 2), ;
              ngastax3   N(9, 2), ;
              ngastax4   N(9, 2), ;
              nprodtax1  N(9, 2), ;
              nprodtax2  N(9, 2), ;
              nprodtax3  N(9, 2), ;
              nprodtax4  N(9, 2) )

         CREATE CURSOR temptax1 ;
            (hyear     c(4), ;
              hperiod    c(2), ;
              cWellID    c(10), ;
              cdeck      c(10), ;
              noiltax1a   N(9, 2), ;
              noiltax2a   N(9, 2), ;
              noiltax3a   N(9, 2), ;
              noiltax4a   N(9, 2), ;
              ngastax1a   N(9, 2), ;
              ngastax2a   N(9, 2), ;
              ngastax3a   N(9, 2), ;
              ngastax4a   N(9, 2), ;
              nprodtax1a  N(9, 2), ;
              nprodtax2a  N(9, 2), ;
              nprodtax3a  N(9, 2), ;
              nprodtax4a  N(9, 2) )

         CREATE CURSOR taxcalc ;
            (hyear     c(4), ;
              hperiod    c(2), ;
              cWellID    c(10), ;
              cdeck      c(10), ;
              ncoiltax1   N(9, 2), ;
              ncoiltax2   N(9, 2), ;
              ncoiltax3   N(9, 2), ;
              ncoiltax4   N(9, 2), ;
              ncgastax1   N(9, 2), ;
              ncgastax2   N(9, 2), ;
              ncgastax3   N(9, 2), ;
              ncgastax4   N(9, 2), ;
              ncprodtax1  N(9, 2), ;
              ncprodtax2  N(9, 2), ;
              ncprodtax3  N(9, 2), ;
              ncprodtax4  N(9, 2) )


         swselect('wellwork')
         COUNT FOR NOT DELETED() TO lnMax
         lcWellFlat = ' x '

         CREATE CURSOR royalty_temp  ;  &&  List of wells and whether there are any royalty owners with interests in other classes
            (cWellID     c(10),  ;      &&  Used to determine whether the lcType variable should be reset for a specific expense entry - BH 10/09/2007
              cdeck       c(10), ;
              lIntClass1   L,  ;
              lIntClass2   L,  ;
              lIntClass3   L,  ;
              lIntClass4   L,  ;
              lIntClass5   L,  ;
              lBCPInt      L,  ;
              lACPInt      L,  ;
              lPlugPct     L)
         INDEX ON cWellID TAG cWellID

         SELECT  cWellID,  ;  &&  Totals for non-WI owners, for the different expense classes
                 cdeck, ;
                 SUM(nintclass1) AS nintclass1,  ;
                 SUM(nintclass2) AS nintclass2,  ;
                 SUM(nintclass3) AS nintclass3,  ;
                 SUM(nintclass4) AS nintclass4,  ;
                 SUM(nintclass5) AS nintclass5,  ;
                 SUM(nbcpint) AS nbcpint,  ;
                 SUM(nacpint) AS nacpint, ;
                 SUM(nplugpct) AS nplugpct ;
             FROM wellinv  ;
             WHERE BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid)  ;
                 AND ctypeinv # 'W'  ;
             INTO CURSOR royalty_tempx  ;
             GROUP BY cWellID,;
                 cdeck

         SELECT royalty_tempx
         SCAN
            STORE .F. TO m.lIntClass1, m.lIntClass2, m.lIntClass3, m.lIntClass4, m.lIntClass5, m.lBCPInt, m.lACPInt, m.lPlugPct
            m.cWellID = cWellID
            IF nintclass1 # 0
               m.lIntClass1 = .T.
            ENDIF
            IF nintclass2 # 0
               m.lIntClass2 = .T.
            ENDIF
            IF nintclass3 # 0
               m.lIntClass3 = .T.
            ENDIF
            IF nintclass4 # 0
               m.lIntClass4 = .T.
            ENDIF
            IF nintclass5 # 0
               m.lIntClass5 = .T.
            ENDIF
            IF nbcpint # 0
               m.lBCPInt = .T.
            ENDIF
            IF nacpint # 0
               m.lACPInt = .T.
            ENDIF
            IF nplugpct # 0
               m.lPlugPct = .T.
            ENDIF

            IF m.lIntClass1 OR m.lIntClass2 OR m.lIntClass3 OR m.lIntClass4 OR m.lIntClass5 OR m.lBCPInt OR m.lACPInt OR m.lPlugPct
               INSERT INTO royalty_temp FROM MEMVAR
            ENDIF
         ENDSCAN

         SELECT royalty_temp
         SET ORDER TO cWellID

* Build a cursor to hold a subset of info for the wells being processed. - pws 5/10/08
         CREATE CURSOR wellamounts ;
            (cWellID     c(10), ;
              cdeck       c(10), ;
              ctable      c(2), ;
              cstate      c(2), ;
              lsev1o      L, ;
              lsev2o      L, ;
              lsev3o      L, ;
              lsev4o      L, ;
              lsev1g      L, ;
              lsev2g      L, ;
              lsev3g      L, ;
              lsev4g      L, ;
              lsev1p      L, ;
              lsev2p      L, ;
              lsev3p      L, ;
              lsev4p      L, ;
              ltaxexempt1 L, ;
              ltaxexempt2 L, ;
              ltaxexempt3 L, ;
              ltaxexempt4 L, ;
              lusesev     L, ;
              nprocess    i, ;
              nroyinto    N(11, 7), ;
              nroyintg    N(11, 7), ;
              nwrkinto    N(11, 7), ;
              nwrkintg    N(11, 7), ;
              nroysevo    N(11, 7), ;
              nroysevg    N(11, 7), ;
              nwrksevo    N(11, 7), ;
              nwrksevg    N(11, 7), ;
              lcompress   L, ;
              nCompress   N(12, 4), ;
              lGather     L, ;
              nGather     N(12, 4), ;
              nflatgas    N(12, 2), ;
              nflatoil    N(12, 2))

         swselect('wells')
         SCAN FOR BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) AND cgroup == THIS.cgroup

* Don't process inactive, sold or plugged wells unless the run is already closed.
            IF INLIST(cwellstat, 'I', 'S', 'P', 'U') AND NOT THIS.lrunclosed
               LOOP
            ENDIF

            SCATTER MEMVAR
            m.nflatgas = THIS.getFlatAmt(m.cWellID, 'G')
            m.nflatoil = THIS.getFlatAmt(m.cWellID, 'O')

            IF m.lusesev
* Get the total royalty pct to use in calculating the taxes when they're specified as pct by well (lusesev)
               SELECT SUM(nrevgas) FROM wellinv INTO ARRAY laArray WHERE INLIST(ctypeinv, 'O', 'L') AND cWellID = m.cWellID

               IF _TALLY > 0
                  m.nroyintg = laArray[1]
               ELSE
                  m.nroyintg = 0
               ENDIF

               SELECT SUM(nrevoil) FROM wellinv INTO ARRAY laArray WHERE INLIST(ctypeinv, 'O', 'L') AND cWellID = m.cWellID

               IF _TALLY > 0
                  m.nroyinto = laArray[1]
               ELSE
                  m.nroyinto = 0
               ENDIF
* Get the total working pct to use in calculating the taxes when they're specified as pct by well (lusesev)
               SELECT SUM(nrevoil) FROM wellinv INTO ARRAY laArrayo WHERE ctypeinv = 'W' AND cWellID = m.cWellID
               IF _TALLY > 0
                  m.nwrkinto = laArrayo[1]
               ELSE
                  m.nwrkinto = 0
               ENDIF

               SELECT SUM(nrevgas) FROM wellinv INTO ARRAY laArrayg WHERE ctypeinv = 'W' AND cWellID = m.cWellID
               IF _TALLY > 0
                  m.nwrkintg = laArrayg[1]
               ELSE
                  m.nwrkintg = 0
               ENDIF
* Check for exempt royalty owners and remove their pct
               SELE wellinv
               SCAN FOR cWellID == m.cWellID AND INLIST(ctypeinv, 'L', 'O')
                  m.cownerid = cownerid
                  m.ngaspct  = nrevgas
                  m.noilpct  = nrevoil
                  SELE investor
                  SET ORDER TO cownerid
                  IF SEEK(m.cownerid) AND lExempt
                     m.nroyintg = m.nroyintg - m.ngaspct
                     m.nroyinto = m.nroyinto - m.noilpct
                  ENDIF
               ENDSCAN

* Check for exempt Working Interest owners and remove their pct
               SELE wellinv
               SCAN FOR cWellID == m.cWellID AND ctypeinv = 'W'
                  m.cownerid = cownerid
                  m.ngaspct  = nrevgas
                  m.noilpct  = nrevoil
                  SELE investor
                  SET ORDER TO cownerid
                  IF SEEK(m.cownerid) AND lExempt
                     m.nwrkinto = m.nwrkinto - m.noilpct
                     m.nwrkintg = m.nwrkintg - m.ngaspct
                  ENDIF
               ENDSCAN
            ELSE
               STORE 0 TO m.nroyinto, m.nroyintg, m.nwrkinto, m.nwrkintg
            ENDIF

            INSERT INTO wellamounts FROM MEMVAR
         ENDSCAN

         SELECT wellamounts
         INDEX ON cWellID TAG cWellID
         SET ORDER TO cWellID

         SELECT wellamounts
         SCAN
            SCATTER MEMVAR
            IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
               llReturn          = .F.
               IF NOT m.goApp.CancelMsg()
                  THIS.lCanceled = .T.
                  EXIT
               ENDIF
            ENDIF

            IF EMPTY(m.nGather) AND m.lGather
               lnGather = lnGlobalGather
            ELSE
               lnGather = m.nGather
            ENDIF
            IF EMPTY(m.nCompress) AND m.lcompress
               lnCompress = lnGlobalCompress
            ELSE
               lnCompress = m.nCompress
            ENDIF
            SELECT wellwork
            SET ORDER TO wellprd
            SCAN FOR cWellID == wellamounts.cWellID
               m.nprocess   = nprocess
               m.ngasint    = ngasint
               m.noilint    = noilint
               lcProdPeriod = hperiod
               lcProdYear   = hyear
               m.hyear      = hyear
               m.hperiod    = hperiod
               lcDeck       = cdeck
               lcType       = THIS.JIB_or_Net(m.cWellID)

               THIS.oprogress.SetProgressMessage('Allocating Revenue and Expenses to Wells...' + m.cWellID)

               STORE 0 TO m.nTotale, m.nGasInc, m.noilinc, m.nTrpInc,  m.nExpgas, m.nExpoil, m.nTotProd
               STORE 0 TO m.nTotBBL, m.nTotMCF, m.bbltot, m.mcftot, m.nMiscinc1, m.nMiscinc2, m.nOthInc
               STORE 0 TO m.nGather, m.nCompress, m.nMCFTaxr, m.nMCFTaxw, m.nBBLTaxr, m.nBBLTaxw
               STORE 0 TO m.nBBLTax1, m.nBBLTax2, m.nBBLTax3, m.nBBLTax4, m.nTotMisc1, m.nTotMisc2
               STORE 0 TO m.nMCFTax1, m.nMCFTax2, m.nMCFTax3, m.nMCFTax4
               STORE 0 TO m.nOthTax1, m.nOthTax2, m.nOthTax3, m.nOthTax4
               STORE 0 TO m.ncBBLTax1, m.ncBBLTax2, m.ncBBLTax3, m.ncBBLTax4
               STORE 0 TO m.ncMCFTax1, m.ncMCFTax2, m.ncMCFTax3, m.ncMCFTax4
               STORE 0 TO m.ncprodtax1, m.ncprodtax2, m.ncprodtax3, m.ncprodtax4


* Get the tax rates for this well
               swselect('sevtax')
               SET ORDER TO ctable
               IF NOT SEEK (m.ctable)
**-
**-  No severance tax record was found for the current state, so
**-  zeros will be used for the tax rates for this well
**-
                  m.ntaxbbl1    = 0
                  m.ntaxmcf1    = 0
                  m.ntaxoth1    = 0
                  m.cMethodBBL1 = '*'
                  m.cMethodMCF1 = '*'
                  m.cMethodOTH1 = '*'
                  m.ntaxbbl2    = 0
                  m.ntaxmcf2    = 0
                  m.ntaxoth2    = 0
                  m.cMethodBBL2 = '*'
                  m.cMethodMCF2 = '*'
                  m.cMethodOTH2 = '*'
                  m.nTaxBBL3    = 0
                  m.ntaxmcf3    = 0
                  m.ntaxoth3    = 0
                  m.cMethodBBL3 = '*'
                  m.cMethodMCF3 = '*'
                  m.cMethodOTH3 = '*'
                  m.ntaxbbl4    = 0
                  m.ntaxmcf4    = 0
                  m.ntaxoth4    = 0
                  m.cMethodBBL4 = '*'
                  m.cMethodMCF4 = '*'
                  m.cMethodOTH4 = '*'
               ELSE
                  SCATTER MEMVAR
               ENDIF

               IF m.lusesev
                  STORE 0 TO m.ntaxbbl1, m.ntaxmcf1
                  STORE 'W' TO m.cMethodBBL1, m.cMethodMCF1
               ENDIF

*  06/24/2004 - pws
*  Get marketing expense totals so that we can subtract them from market value
*  of gas before calculating taxes
               lnMarketing = 0
               SELE expense
               SUM(namount) TO lnMarketing ;
                  FOR cWellID = m.cWellID AND ccatcode = 'MKTG' ;
                  AND cdeck == lcDeck ;
                  AND nRunNoRev = 0 ;
                  AND (cyear + cperiod = lcProdYear + lcProdPeriod) ;
                  AND dexpdate <= THIS.drevdate

*  04/24/2017 - pws
*  Set up variable for Gathering charges entered as an expense
               lnGathering = 0

*  04/24/2017 - pws
*  Set up variable for Compression charges entered as an expense
               lnCompression = 0

               m.nprodwell = 0
               lnDaysOn    = 0

               m.nprodwell = 0
               lnDaysOn    = 0

*
* Process statement notes and mark them as used if the run is being closed.
*
               IF THIS.lclose
                  swselect('stmtnote')
                  *!* * Removed 05/14/2025
                  *!* SCAN FOR nrunno = 0
                  *!*    SELECT wells
                  *!*    SET ORDER TO cWellID
                  *!*    IF SEEK(stmtnote.cWellID) AND wells.cgroup == THIS.cgroup
                  *!*       SELECT stmtnote
                  *!*       REPL nrunno    WITH lnRunNo, ;
                  *!*          crunyear  WITH THIS.crunyear
                  *!*    ENDIF
                  *!* ENDSCAN
                  LOCATE FOR cWellID = m.cWellID AND nrunno = 0
                  IF FOUND()
                     REPL nrunno    WITH lnRunNo, ;
                        crunyear  WITH THIS.crunyear
                  ENDIF
               ENDIF

               STORE 0 TO m.noiltax1, m.ngastax1, m.noiltax2, m.ngastax2, lnTax
               STORE 0 TO m.noiltax3, m.ngastax3, m.noiltax4, m.ngastax4
               STORE 0 TO m.nprodtax1, m.nprodtax2, m.nprodtax3, m.nprodtax4
               STORE 0 TO m.noiltax1a, m.ngastax1a, m.noiltax2a, m.ngastax2a
               STORE 0 TO m.noiltax3a, m.ngastax3a, m.noiltax4a, m.ngastax4a
               STORE 0 TO m.nprodtax1a, m.nprodtax2a, m.nprodtax3a, m.nprodtax4a
               STORE 0 TO m.noiltax1b, m.ngastax1b, m.noiltax2b, m.ngastax2b
               STORE 0 TO m.noiltax3b, m.ngastax3b, m.noiltax4b, m.ngastax4b
               STORE 0 TO m.nprodtax1b, m.nprodtax2b, m.nprodtax3b, m.nprodtax4b
               STORE 0 TO m.ngoiltax1, m.nggastax1, m.ngoiltax2, m.nggastax2
               STORE 0 TO m.ngoiltax3, m.nggastax3, m.ngoiltax4, m.nggastax4
               STORE 0 TO m.ngProdtax1, m.ngProdtax2, m.ngProdTax3, m.ngProdtax4
               STORE 0 TO m.nGBBLTax1, m.nGBBLTax2, m.nGBBLTax3, m.nGBBLTax4
               STORE 0 TO m.nGMCFTax1, m.nGMCFTax2, m.nGMCFTax3, m.nGMCFTax4
               STORE 0 TO m.nGOTHTax1, m.nGOTHTax2, m.nGOTHTax3, m.nGOTHTax4
               STORE 0 TO lnNetBBL, lnNETMCF, m.nTotComp, m.nTotGath

* Create cursors for tracking gathering and compression
               Make_Copy('income', 'comprev')
               Make_Copy('income', 'gathrev')

*  Process well income
*
*  Scan for all income records that have a runno of zero.  Also look to see if they match
*  one of the production periods from wellwork
*
* m.nGBBLTax? = gross oil taxes (includes: exempt, one-man and dummy owners)
* m.nBBLTax?  = net oil taxes (excludes: exempt, direct paid and dummy owners)

* Fill in the deck if missing
               llUpdateIncome = .F.
               swselect('income')
               SCAN FOR cWellID == m.cWellID AND nrunno = 0 AND cyear + cperiod = lcProdYear + lcProdPeriod AND EMPTY(cdeck)
                  m.cdeck = THIS.oWellInv.DOIDeckNameLookup(income.cyear, income.cperiod, income.cWellID)
                  REPLACE cdeck WITH m.cdeck
                  llUpdateIncome = .T.
               ENDSCAN
               IF llUpdateIncome
                  TABLEUPDATE(.T.,.T., 'income')
               ENDIF

               SELECT income
               SCAN FOR cWellID == m.cWellID ;
                     AND nrunno = 0 ;
                     AND cyear + cperiod = lcProdYear + lcProdPeriod ;
                     AND cdeck == lcDeck ;
                     AND drevdate <= THIS.drevdate

                  SCATTER MEMVAR

                  STORE 0 TO m.ncoiltax1, m.ncoiltax2, m.ncoiltax3, m.ncoiltax4
                  STORE 0 TO m.ncgastax1, m.ncgastax2, m.ncgastax3, m.ncgastax4
                  STORE 0 TO m.ncprodtax1, m.ncprodtax2, m.ncprodtax3, m.ncprodtax4

* If the run is being closed, replace the runno and year in the income record
                  IF THIS.lclose
                     REPL nrunno    WITH lnRunNo, ;
                        lclosed   WITH .T., ;
                        crunyear  WITH THIS.crunyear, ;
                        cacctyear WITH THIS.cacctyear, ;
                        cacctprd  WITH THIS.cacctprd

                     IF EMPTY(dacctdate)
                        REPL dacctdate WITH THIS.dacctdate
                     ENDIF
                     swselect('stmtnote')
                     LOCATE FOR cWellID = m.cWellID AND nrunno = 0
                     IF FOUND()
                        REPL nrunno    WITH lnRunNo, ;
                           crunyear  WITH THIS.crunyear
                     ENDIF
                  ENDIF

                  m.nprodwell = m.nprodwell + m.nTotalInc
* Put this in to avoid Numeric Overflow Errors - pws - 11/19/21
                  IF m.nDaysOn > 100000
                     m.nDaysOn = 31
                  ENDIF
                  lnDaysOn    = m.nDaysOn
                  DO CASE
                     CASE m.cSource = "BBL"
                        m.nTotBBL = m.nTotBBL + m.nUnits

                        IF EMPTY(m.cownerid)
* If the well is not severance tax exempt, calculate the sev taxes
                           IF NOT m.ltaxexempt1
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev1o
                                    m.nGBBLTax1 = m.nGBBLTax1
                                    lnTax       = 0
                                 CASE m.cMethodBBL1 = '*' OR m.cMethodBBL1 = 'A'   && Rate specified by amount per bbl
                                    m.nGBBLTax1 = m.nGBBLTax1 + swround((m.ntaxbbl1 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxbbl1 * m.nUnits), 2)
                                    m.nBBLTaxr  = 0
                                    m.nBBLTaxw  = 0
                                 CASE m.cMethodBBL1 = 'P'   && Rate specified by percentage
                                    lnAmount    = m.nTotalInc
                                    m.nGBBLTax1 = m.nGBBLTax1 + swround(((m.ntaxbbl1 / 100) * lnAmount), 2)
                                    lnTax       = swround(((m.ntaxbbl1 / 100) * lnAmount), 2)
                                    m.nBBLTaxr  = 0
                                    m.nBBLTaxw  = 0
                                 CASE m.cMethodBBL1 = '%'   && Rate specified by percentage per unit
                                    lnAmount    = m.nTotalInc
                                    m.nGBBLTax1 = m.nGBBLTax1 + swround(((m.ntaxbbl1 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxbbl1 / 100) * m.nUnits), 2)
                                    m.nBBLTaxr  = 0
                                    m.nBBLTaxw  = 0
                                 CASE m.cMethodBBL1 = 'W'       && Rate Specified by Well
                                    m.nBBLTaxr  = m.nBBLTaxr + (m.nTotalInc * (m.nroyinto / 100)) * (m.nroysevo / 100)
                                    m.nBBLTaxw  = m.nBBLTaxw + (m.nTotalInc * (m.nwrkinto / 100)) * (m.nwrksevo / 100)
                                    m.nBBLTax1  = 0
                                    lnTax       = 0
                                    m.nGBBLTax1 = 0
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.noiltax1a = m.noiltax1a + lnTax
                                 m.noiltax1b = m.noiltax1b + lnTax
                                 m.nBBLTax1  = m.nBBLTax1  + lnTax
                              ELSE
                                 m.nBBLTax1  = m.nBBLTax1 + swnetrev(m.cWellID, lnTax, 'O1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncoiltax1 = m.nGBBLTax1
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncoiltax1 WITH m.ncoiltax1
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nBBLTax1, m.nBBLTaxr, m.nBBLTaxw, m.noiltax1
                           ENDIF
* If the well is not exempt from tax2, calculate the taxes
                           IF NOT m.ltaxexempt2
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev2o
                                    m.nBBLTax2 = m.nBBLTax2
                                    lnTax      = 0
                                 CASE m.cMethodBBL2 = '*' OR m.cMethodBBL2 = 'A'
                                    m.nGBBLTax2 = m.nGBBLTax2 + swround((m.ntaxbbl2 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxbbl2 * m.nUnits), 2)
                                 CASE m.cMethodBBL2 = 'P'
                                    m.nGBBLTax2 = m.nGBBLTax2 + swround(((m.ntaxbbl2 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxbbl2 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodBBL2 = '%'
                                    m.nGBBLTax2 = m.nGBBLTax2 + swround(((m.ntaxbbl2 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxbbl2 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.noiltax2a = m.noiltax2a + lnTax
                                 m.noiltax2b = m.noiltax2b + lnTax
                                 m.nBBLTax2  = m.nBBLTax2  + lnTax
                              ELSE
                                 m.nBBLTax2  = m.nBBLTax2 + swnetrev(m.cWellID, lnTax, 'O2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncoiltax2 = m.nGBBLTax2
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncoiltax2 WITH m.ncoiltax2
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nBBLTax2, m.noiltax2
                           ENDIF
* If the well is not exempt from tax3, calculate the taxes
                           IF NOT m.ltaxexempt3
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev3o
                                    m.nBBLTax3 = m.nBBLTax3
                                    lnTax      = 0
                                 CASE m.cMethodBBL3 = '*' OR m.cMethodBBL3 = 'A'
                                    m.nGBBLTax3 = m.nGBBLTax3 + swround((m.nTaxBBL3 * m.nUnits), 2)
                                    lnTax       = swround((m.nTaxBBL3 * m.nUnits), 2)
                                 CASE m.cMethodBBL3 = 'P'
                                    m.nGBBLTax3 = m.nGBBLTax3 + swround(((m.nTaxBBL3 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.nTaxBBL3 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodBBL3 = '%'
                                    m.nGBBLTax3 = m.nGBBLTax3 + swround(((m.nTaxBBL3 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.nTaxBBL3 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.noiltax3a = m.noiltax3a + lnTax
                                 m.noiltax3b = m.noiltax3b + lnTax
                                 m.nBBLTax3  = m.nBBLTax3  + lnTax
                              ELSE
                                 m.nBBLTax3  = m.nBBLTax3 + swnetrev(m.cWellID, lnTax, 'O3', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncoiltax3 = m.nGBBLTax3
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncoiltax3 WITH m.ncoiltax3
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nBBLTax3, m.noiltax3
                           ENDIF
* If the well is not exempt from tax4, calculate the taxes
                           IF NOT m.ltaxexempt4
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev4o
                                    m.nBBLTax4 = m.nBBLTax4
                                    lnTax      = 0
                                 CASE m.cMethodBBL4 = '*' OR m.cMethodBBL4 = 'A'
                                    m.nGBBLTax4 = m.nGBBLTax4 + swround((m.ntaxbbl4 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxbbl4 * m.nUnits), 2)
                                 CASE m.cMethodBBL4 = 'P'
                                    m.nGBBLTax4 = m.nGBBLTax4 + swround(((m.ntaxbbl4 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxbbl4 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodBBL4 = '%'
                                    m.nGBBLTax4 = m.nGBBLTax4 + swround(((m.ntaxbbl4 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxbbl4 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.noiltax4a = m.noiltax4a + lnTax
                                 m.noiltax4b = m.noiltax4b + lnTax
                                 m.nBBLTax4  = m.nBBLTax4  + lnTax
                              ELSE
                                 m.nBBLTax4  = m.nBBLTax4 + swnetrev(m.cWellID, lnTax, 'O4', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncoiltax4 = m.nGBBLTax4
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncoiltax4 WITH m.ncoiltax4
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nBBLTax4, m.noiltax4
                           ENDIF
                        ENDIF
                        IF m.nTotalInc # 0
                           m.noilinc = m.noilinc + m.nTotalInc
                        ENDIF

                     CASE m.cSource = "MCF"
                        m.nTotMCF = m.nTotMCF + m.nUnits
                        IF EMPTY(m.cownerid)
* If the well is not severance tax exempt, calculate the sev taxes
                           IF NOT m.ltaxexempt1
                              DO CASE
                                 CASE m.lsev1g          && Purchaser withholds gas tax 1
                                    m.nMCFTax1 = m.nMCFTax1
                                    lnTax      = 0
                                 CASE m.cMethodMCF1 = '*' OR m.cMethodMCF1 = 'A'
                                    m.nGMCFTax1 = m.nGMCFTax1 + swround((m.ntaxmcf1 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxmcf1 * m.nUnits), 2)
                                    m.nMCFTaxr  = 0
                                    m.nMCFTaxw  = 0
                                 CASE m.cMethodMCF1 = 'P'
                                    lnAmount    = m.nTotalInc
                                    m.nGMCFTax1 = m.nGMCFTax1 + swround(((m.ntaxmcf1 / 100) * lnAmount), 2)
                                    lnTax       = swround(((m.ntaxmcf1 / 100) * lnAmount), 2)
                                    m.nMCFTaxr  = 0
                                    m.nMCFTaxw  = 0
                                 CASE m.cMethodMCF1 = '%'
                                    lnAmount    = m.nTotalInc
                                    m.nGMCFTax1 = m.nGMCFTax1 + swround(((m.ntaxmcf1 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxmcf1 / 100) * m.nUnits), 2)
                                    m.nMCFTaxr  = 0
                                    m.nMCFTaxw  = 0
                                 CASE m.cMethodMCF1 = 'W'
                                    m.nMCFTaxr  = m.nMCFTaxr + (m.nTotalInc * (m.nroyintg / 100)) * (m.nroysevg / 100)
                                    m.nMCFTaxw  = m.nMCFTaxw + (m.nTotalInc * (m.nwrkintg / 100)) * (m.nwrksevg / 100)
                                    m.nMCFTax1  = 0
                                    lnTax       = 0
                                    m.nGMCFTax1 = 0
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.ngastax1a = m.ngastax1a + lnTax
                                 m.ngastax1b = m.ngastax1b + lnTax
                                 m.nMCFTax1  = m.nMCFTax1  + lnTax
                              ELSE
                                 m.nMCFTax1  = m.nMCFTax1 + swnetrev(m.cWellID, lnTax, 'G1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
* Moved tax calc to after mktg taken out
                              ENDIF
                           ELSE
                              STORE 0 TO m.nMCFTax1, m.nMCFTaxr, m.nMCFTaxw, m.ngastax1
                           ENDIF
*  If the well is not exempt from tax 2, calculate the taxes
                           IF NOT m.ltaxexempt2
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev2g
                                    m.nMCFTax2 = m.nMCFTax2
                                    lnTax      = 0
                                 CASE m.cMethodMCF2 = '*' OR m.cMethodMCF2 = 'A'
                                    m.nGMCFTax2 = m.nGMCFTax2 + swround((m.ntaxmcf2 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxmcf2 * m.nUnits), 2)
                                 CASE m.cMethodMCF2 = 'P'
                                    m.nGMCFTax2 = m.nGMCFTax2 + swround(((m.ntaxmcf2 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxmcf2 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodMCF2 = '%'
                                    m.nGMCFTax2 = m.nGMCFTax2 + swround(((m.ntaxmcf2 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxmcf2 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.ngastax2a = m.ngastax2a + lnTax
                                 m.ngastax2b = m.ngastax2b + lnTax
                                 m.nMCFTax2  = m.nMCFTax2  + lnTax
                              ELSE
                                 m.nMCFTax2  = m.nMCFTax2 + swnetrev(m.cWellID, lnTax, 'G2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncgastax2 = m.nGMCFTax2
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncgastax2 WITH m.ncgastax2
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nMCFTax2, m.ngastax2
                           ENDIF
*  If the well is not exempt from tax 3, calculate the taxes
                           IF NOT m.ltaxexempt3
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev3g
                                    m.nMCFTax3 = m.nMCFTax3
                                    lnTax      = 0
                                 CASE m.cMethodMCF3 = '*' OR m.cMethodMCF3 = 'A'
                                    m.nGMCFTax3 = m.nGMCFTax3 + swround((m.ntaxmcf3 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxmcf3 * m.nUnits), 2)
                                 CASE m.cMethodMCF3 = 'P'
                                    m.nGMCFTax3 = m.nGMCFTax3 + swround(((m.ntaxmcf3 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxmcf3 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodMCF3 = '%'
                                    m.nGMCFTax3 = m.nGMCFTax3 + swround(((m.ntaxmcf3 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxmcf3 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.ngastax3a = m.ngastax3a + lnTax
                                 m.ngastax3b = m.ngastax3b + lnTax
                                 m.nMCFTax3  = m.nMCFTax3  + lnTax
                              ELSE
                                 m.nMCFTax3  = m.nMCFTax3 + swnetrev(m.cWellID, lnTax, 'G3', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncgastax3 = m.nGMCFTax3
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncgastax3 WITH m.ncgastax3
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nMCFTax3, m.ngastax3
                           ENDIF
*  If the well is not exempt from tax 4, calculate the taxes
                           IF NOT m.ltaxexempt4
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev4g
                                    m.nMCFTax4 = m.nMCFTax4
                                    lnTax      = 0
                                 CASE m.cMethodMCF4 = '*' OR m.cMethodMCF4 = 'A'
                                    m.nGMCFTax4 = m.nGMCFTax4 + swround((m.ntaxmcf4 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxmcf4 * m.nUnits), 2)
                                 CASE m.cMethodMCF4 = 'P'
                                    m.nGMCFTax4 = m.nGMCFTax4 + swround(((m.ntaxmcf4 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxmcf4 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodMCF3 = '%'
                                    m.nGMCFTax3 = m.nGMCFTax3 + swround(((m.ntaxmcf3 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxmcf3 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.ngastax4a = m.ngastax4a + lnTax
                                 m.ngastax4b = m.ngastax4b + lnTax
                                 m.nMCFTax4  = m.nMCFTax4  + lnTax
                              ELSE
                                 m.nMCFTax4  = m.nMCFTax4 + swnetrev(m.cWellID, lnTax, 'G4', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncgastax4 = m.nGMCFTax4
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncgastax4 WITH m.ncgastax4
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nMCFTax4, m.ngastax4
                           ENDIF
                        ENDIF
                        IF m.lcompress AND EMPTY(m.cownerid)  &&  No compression if it's one-man-item - BH 12/04/2008
                           lnComp      = 0
                           lnComp      = swround((lnCompress * m.nUnits), 2)
                           m.nCompress = m.nCompress + lnComp
                           SELECT compcalc
                           LOCATE FOR cWellID = m.cWellID AND cyear = m.cyear AND cperiod = m.cperiod
                           IF FOUND() AND lnComp # 0
                              REPLACE namount WITH namount + lnComp
                           ELSE
                              IF lnComp # 0
                                 m.namount = lnComp
                                 INSERT INTO compcalc FROM MEMVAR
                              ENDIF
                           ENDIF
                        ENDIF

                        IF m.lGather AND EMPTY(m.cownerid)  &&  No gathering if it's one-man-item - BH 12/04/2008
                           lnGath    = 0
                           lnGath    = swround((lnGather * m.nUnits), 2)
                           m.nGather = m.nGather + lnGath
                           SELECT gathcalc
                           LOCATE FOR cWellID = m.cWellID AND cyear = m.cyear AND cperiod = m.cperiod
                           IF FOUND() AND lnGath # 0
                              REPLACE namount WITH namount + lnGath
                           ELSE
                              IF lnGath # 0
                                 m.namount = lnGath
                                 INSERT INTO gathcalc FROM MEMVAR
                              ENDIF
                           ENDIF
                        ENDIF

                        IF m.nTotalInc # 0
                           m.nGasInc = m.nGasInc + m.nTotalInc
                        ENDIF

                     CASE m.cSource = "OTH"   && Other Product Taxes
                        lnTax      = 0
                        m.nTotProd = m.nTotProd + m.nUnits

                        IF EMPTY(m.cownerid)
* If the well is not severance tax exempt, calculate the taxes
                           IF NOT m.ltaxexempt1
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev1p
                                    lnTax = 0
                                 CASE m.cMethodOTH1 = '*' OR m.cMethodOTH1 = 'A'   && Rate specified by amount per unit
                                    m.nGOTHTax1 = m.nGOTHTax1 + swround((m.ntaxoth1 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxoth1 * m.nUnits), 2)
                                 CASE m.cMethodOTH1 = 'P'   && Rate specified by percentage
                                    m.nGOTHTax1 = m.nGOTHTax1 + swround(((m.ntaxoth1 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxoth1 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodOTH1 = '%'   && Rate specified by percentage
                                    m.nGOTHTax1 = m.nGOTHTax1 + swround(((m.ntaxoth1 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxoth1 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.nprodtax1a = m.nprodtax1a + lnTax
                                 m.nprodtax1b = m.nprodtax1b + lnTax
                                 m.nOthTax1   = m.nOthTax1   + lnTax
                              ELSE
                                 m.nOthTax1   = m.nOthTax1 + swnetrev(m.cWellID, lnTax, 'P1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncprodtax1 = m.nGOTHTax1
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncprodtax1 WITH m.ncprodtax1
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nOthTax1
                           ENDIF
* If the well is not exempt from tax2, calculate the taxes
                           IF NOT m.ltaxexempt2
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev2p
                                    m.nOthTax2 = m.nOthTax2
                                    lnTax      = 0
                                 CASE m.cMethodOTH2 = '*' OR m.cMethodOTH2 = 'A'
                                    m.nGOTHTax2 = m.nGOTHTax2 + swround((m.ntaxoth2 * m.nUnits), 2)
                                    lnTax       = + swround((m.ntaxoth2 * m.nUnits), 2)
                                 CASE m.cMethodOTH2 = 'P'
                                    m.nGOTHTax2 = m.nGOTHTax2 + swround(((m.ntaxoth2 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxoth2 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodOTH2 = '%'
                                    m.nGOTHTax2 = m.nGOTHTax2 + swround(((m.ntaxoth2 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxoth2 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.nprodtax2a = m.nprodtax2a + lnTax
                                 m.nprodtax2b = m.nprodtax2b + lnTax
                                 m.nOthTax2   = m.nOthTax2   + lnTax
                              ELSE
                                 m.nOthTax2   = m.nOthTax2 + swnetrev(m.cWellID, lnTax, 'P2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncprodtax2 = m.nGOTHTax2
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncprodtax2 WITH m.ncprodtax2
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nOthTax2
                           ENDIF
* If the well is not exempt from tax3, calculate the taxes
                           IF NOT m.ltaxexempt3
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev3p
                                    m.nOthTax3 = m.nOthTax3
                                 CASE m.cMethodOTH3 = '*' OR m.cMethodOTH3 = 'A'
                                    m.nGOTHTax3 = m.nGOTHTax3 + swround((m.ntaxoth3 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxoth3 * m.nUnits), 2)
                                 CASE m.cMethodOTH3 = 'P'
                                    m.nGOTHTax3 = m.nGOTHTax3 + swround(((m.ntaxoth3 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxoth3 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodOTH3 = '%'
                                    m.nGOTHTax3 = m.nGOTHTax3 + swround(((m.ntaxoth3 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxoth3 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.nprodtax3a = m.nprodtax3a + lnTax
                                 m.nprodtax3b = m.nprodtax3b + lnTax
                                 m.nOthTax3   = m.nOthTax3   + lnTax
                              ELSE
                                 m.nOthTax3   = m.nOthTax3 + swnetrev(m.cWellID, lnTax, 'P3', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncprodtax3 = m.nGOTHTax3
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncprodtax3 WITH m.ncprodtax3
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nOthTax3
                           ENDIF
* If the well is not exempt from tax4, calculate the taxes
                           IF NOT m.ltaxexempt4
                              lnTax = 0
                              DO CASE
                                 CASE m.lsev4p
                                    m.nOthTax4 = m.nOthTax4
                                    lnTax      = 0
                                 CASE m.cMethodOTH4 = '*' OR m.cMethodOTH4 = 'A'
                                    m.nGOTHTax4 = m.nGOTHTax4 + swround((m.ntaxoth4 * m.nUnits), 2)
                                    lnTax       = swround((m.ntaxoth4 * m.nUnits), 2)
                                 CASE m.cMethodOTH4 = 'P'
                                    m.nGOTHTax4 = m.nGOTHTax4 + swround(((m.ntaxoth4 / 100) * m.nTotalInc), 2)
                                    lnTax       = swround(((m.ntaxoth4 / 100) * m.nTotalInc), 2)
                                 CASE m.cMethodOTH4 = '%'
                                    m.nGOTHTax4 = m.nGOTHTax4 + swround(((m.ntaxoth4 / 100) * m.nUnits), 2)
                                    lnTax       = swround(((m.ntaxoth4 / 100) * m.nUnits), 2)
                              ENDCASE

* One man item - add tax to temptax1 variable
                              IF NOT EMPTY(m.cownerid)
                                 m.nprodtax4a = m.nprodtax4a + lnTax
                                 m.nprodtax4b = m.nprodtax4a + lnTax
                                 m.nOthTax4   = m.nOthTax4   + lnTax
                              ELSE
                                 m.nOthTax4   = m.nOthTax4 + swnetrev(m.cWellID, lnTax, 'P4', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                                 m.ncprodtax4 = m.nGOTHTax4
                                 SELECT taxcalc
                                 LOCATE FOR cWellID = m.cWellID AND ;
                                    hyear   = lcProdYear AND ;
                                    hperiod = lcProdPeriod
                                 IF FOUND()
                                    REPLACE ncprodtax4 WITH m.ncprodtax4
                                 ELSE
                                    INSERT INTO taxcalc FROM MEMVAR
                                 ENDIF
                              ENDIF
                           ELSE
                              STORE 0 TO m.nOthTax4
                           ENDIF
                        ENDIF

                        IF m.nTotalInc # 0
                           m.nOthInc = m.nOthInc + m.nTotalInc
                        ENDIF

                     CASE m.cSource = "TRANS"
                        m.nTrpInc = m.nTrpInc + m.nTotalInc

                     CASE m.cSource = "MISC1"
                        m.nTotMisc1 = m.nTotMisc1 + m.nUnits
                        m.nMiscinc1 = m.nMiscinc1 + m.nTotalInc

                     CASE m.cSource = "MISC2"
                        m.nTotMisc2 = m.nTotMisc2 + m.nUnits
                        m.nMiscinc2 = m.nMiscinc2 + m.nTotalInc

                     CASE m.cSource = "COMP"
                        lnCompression = lnCompression - m.nTotalInc

                     CASE m.cSource = "GATH"
                        lnGathering = lnGathering - m.nTotalInc

                     CASE m.cSource = 'OTAX1'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'O1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngoiltax1 = m.ngoiltax1 + m.nTotalInc * -1
                        m.noiltax1  = m.noiltax1 + lnNetTax * -1

                     CASE m.cSource = 'GTAX1'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'G1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.nggastax1 = m.nggastax1 + m.nTotalInc * -1
                        m.ngastax1  = m.ngastax1 + lnNetTax * -1

                     CASE m.cSource = 'PTAX1'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'P1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngProdtax1 = m.ngProdtax1 + (m.nTotalInc * -1)
                        m.nprodtax1  = m.nprodtax1 + lnNetTax * -1

                     CASE m.cSource = 'OTAX2'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'O2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngoiltax2 = m.ngoiltax2 + m.nTotalInc * -1
                        m.noiltax2  = m.noiltax2 + lnNetTax * -1

                     CASE m.cSource = 'GTAX2'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'G2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.nggastax2 = m.nggastax2 + m.nTotalInc * -1
                        m.ngastax2  = m.ngastax2 + lnNetTax * -1

                     CASE m.cSource = 'PTAX2'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'P2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngProdtax2 = m.ngProdtax2 + (m.nTotalInc * -1)
                        m.nprodtax2  = m.nprodtax2 + lnNetTax * -1

                     CASE m.cSource = 'OTAX3'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'O3', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngoiltax3 = m.ngoiltax3 + m.nTotalInc * -1
                        m.noiltax3  = m.noiltax3 + lnNetTax * -1

                     CASE m.cSource = 'GTAX3'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'G3', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.nggastax3 = m.nggastax3 + m.nTotalInc * -1
                        m.ngastax3  = m.ngastax3 + lnNetTax * -1

                     CASE m.cSource = 'PTAX3'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'P3', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngProdTax3 = m.ngProdTax3 + (m.nTotalInc * -1)
                        m.nprodtax3  = m.nprodtax3 + lnNetTax * -1

                     CASE m.cSource = 'OTAX4'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'O4', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngoiltax4 = m.ngoiltax4 + m.nTotalInc * -1
                        m.noiltax4  = m.noiltax4 + lnNetTax * -1

                     CASE m.cSource = 'GTAX4'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'G4', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.nggastax4 = m.nggastax4 + m.nTotalInc * -1
                        m.ngastax4  = m.ngastax4 + lnNetTax * -1

                     CASE m.cSource = 'PTAX4'
                        IF EMPTY(m.cownerid)
                           lnNetTax = swnetrev(m.cWellID, m.nTotalInc, 'P4', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnNetTax = m.nTotalInc
                        ENDIF
                        m.ngProdtax4 = m.ngProdtax4 + (m.nTotalInc * -1)
                        m.nprodtax4  = m.nprodtax4 + lnNetTax * -1
                  ENDCASE

                  IF NOT EMPTY(m.cownerid)
                     swselect('one_man_tax')
                     LOCATE FOR cWellID == m.cWellID AND cownerid == m.cownerid ;
                        AND hyear + hperiod == lcProdYear + m.hperiod ;
                        AND crunyear == THIS.crunyear AND nrunno == THIS.nrunno
                     IF FOUND()
                        REPL noiltax1b WITH noiltax1b + m.noiltax1b, ;
                           noiltax2b WITH noiltax2b + m.noiltax2b, ;
                           noiltax3b WITH noiltax3b + m.noiltax3b, ;
                           noiltax4b WITH noiltax4b + m.noiltax4b, ;
                           ngastax1b WITH ngastax1b + m.ngastax1b, ;
                           ngastax2b WITH ngastax2b + m.ngastax2b, ;
                           ngastax3b WITH ngastax3b + m.ngastax3b, ;
                           ngastax4b WITH ngastax4b + m.ngastax4b, ;
                           nprodtax1b WITH nprodtax1b + m.nprodtax1b, ;
                           nprodtax2b WITH nprodtax2b + m.nprodtax2b, ;
                           nprodtax3b WITH nprodtax3b + m.nprodtax3b, ;
                           nprodtax4b WITH nprodtax4b + m.nprodtax4b
                     ELSE
                        m.crunyear = THIS.crunyear
                        m.nrunno   = THIS.nrunno
*INSERT INTO one_man_tax FROM MEMVAR
                     ENDIF
                     STORE 0 TO m.noiltax1b, m.ngastax1b, m.noiltax2b, m.ngastax2b
                     STORE 0 TO m.noiltax3b, m.ngastax3b, m.noiltax4b, m.ngastax4b
                     STORE 0 TO m.nprodtax1b, m.nprodtax2b, m.nprodtax3b, m.nprodtax4b
                  ENDIF


               ENDSCAN        && income

               INSERT INTO temptax  FROM MEMVAR
               INSERT INTO temptax1 FROM MEMVAR

* Add the net calculated and entered taxes
               m.nBBLTax1 = m.nBBLTax1 + m.noiltax1
               m.nMCFTax1 = m.nMCFTax1 + m.ngastax1
               m.nOthTax1 = m.nOthTax1 + m.nprodtax1
               m.nBBLTax2 = m.nBBLTax2 + m.noiltax2
               m.nMCFTax2 = m.nMCFTax2 + m.ngastax2
               m.nOthTax2 = m.nOthTax2 + m.nprodtax2
               m.nBBLTax3 = m.nBBLTax3 + m.noiltax3
               m.nMCFTax3 = m.nMCFTax3 + m.ngastax3
               m.nOthTax3 = m.nOthTax3 + m.nprodtax3
               m.nBBLTax4 = m.nBBLTax4 + m.noiltax4
               m.nMCFTax4 = m.nMCFTax4 + m.ngastax4
               m.nOthTax4 = m.nOthTax4 + m.nprodtax4

* Add the gross calculated and entered taxes
               m.noiltax1  = m.ngoiltax1 + m.nGBBLTax1
               m.ngastax1  = m.nggastax1 + m.nGMCFTax1
               m.nprodtax1 = m.ngProdtax1 + m.nGOTHTax1
               m.noiltax2  = m.ngoiltax2 + m.nGBBLTax2
               m.ngastax2  = m.nggastax2 + m.nGMCFTax2
               m.nprodtax2 = m.ngProdtax2 + m.nGOTHTax2
               m.noiltax3  = m.ngoiltax3 + m.nGBBLTax3
               m.ngastax3  = m.nggastax3 + m.nGMCFTax3
               m.nprodtax3 = m.ngProdTax3 + m.nGOTHTax3
               m.noiltax4  = m.ngoiltax4 + m.nGBBLTax4
               m.ngastax4  = m.nggastax4 + m.nGMCFTax4
               m.nprodtax4 = m.ngProdtax4 + m.nGOTHTax4


* 06/24/2004 - pws
* Remove marketing expense from gas tax
*
               IF m.nMCFTax1 # 0
                  IF m.cMethodMCF1 = 'P'
                     lnMktAmt    = lnMarketing * (m.ntaxmcf1 / 100)
                     m.ncgastax1 = m.nGMCFTax1 - lnMktAmt
                     SELECT taxcalc
                     LOCATE FOR cWellID = m.cWellID AND ;
                        hyear   = lcProdYear AND ;
                        hperiod = lcProdPeriod
                     IF FOUND()
                        REPLACE ncgastax1 WITH m.ncgastax1
                     ELSE
                        INSERT INTO taxcalc FROM MEMVAR
                     ENDIF
                     IF NOT m.lsev1g AND NOT m.ltaxexempt1
* For the net tax, net down the marketing amount before subtracting it off - BH 12/11/07
                        m.nMCFTax1 = m.nMCFTax1 - ;
                           swnetrev(m.cWellID, lnMktAmt, 'G1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
                        m.ngastax1 = m.ngastax1 - lnMktAmt
                     ENDIF
                  ELSE
                     m.ncgastax1 = m.nGMCFTax1
                     SELECT taxcalc
                     LOCATE FOR cWellID = m.cWellID AND ;
                        hyear   = lcProdYear AND ;
                        hperiod = lcProdPeriod
                     IF FOUND()
                        REPLACE ncgastax1 WITH m.ncgastax1
                     ELSE
                        INSERT INTO taxcalc FROM MEMVAR
                     ENDIF
                  ENDIF

               ENDIF

               STORE 0 TO m.nTotale, m.nexpcl1, m.nexpcl2, m.nexpcl3, m.nexpcl4, m.nexpcl5, m.ntotsalt
               STORE 0 TO m.nexpclA, m.nexpclB, m.nPlugAmt

**-
**-  Process well expenses
**-
               swselect('expcat')
               SET ORDER TO ccatcode

* Fill in any missing deck codes
               llExpenseUpdate = .F.
               swselect('expense')
               SCAN FOR cWellID = m.cWellID AND nRunNoRev = 0 AND (cyear + cperiod = lcProdYear + lcProdPeriod) AND dexpdate <= THIS.dexpdate AND EMPTY(cdeck)
                  m.cdeck = THIS.oWellInv.DOIDeckNameLookup(expense.cyear, expense.cperiod, expense.cWellID)
                  SELECT expense
                  REPLACE cdeck WITH m.cdeck
                  llExpenseUpdate = .T.
               ENDSCAN
               IF llExpenseUpdate
                  TABLEUPDATE(.T.,.T., 'expense')
               ENDIF

               SELECT expense
               SCAN FOR cWellID = m.cWellID AND nRunNoRev = 0 AND (cyear + cperiod = lcProdYear + lcProdPeriod) AND dexpdate <= THIS.dexpdate AND cdeck == lcDeck
                  SCATTER MEMVAR

* Don't process marketing, compression or gathering expenses here
                  IF INLIST(m.ccatcode, 'MKTG', 'COMP', 'GATH')
                     LOOP
                  ENDIF

* Don't process plugging expenses if the plugging module isn't active
                  IF NOT m.goApp.lPluggingModule AND m.cexpclass = 'P'
                     LOOP
                  ENDIF

* Check to see if the expense should only be processed in a JIB run.
                  SELECT expcat
                  IF SEEK(m.ccatcode)
                     IF lJIBOnly
                        LOOP
                     ENDIF
                  ENDIF

* Check to see if the expense is a one man item assigned to a JIB owner
* If so, we don't process it here.
                  IF NOT EMPTY(m.cownerid)
                     swselect('wellinv')
                     LOCATE FOR cWellID = m.cWellID AND cownerid == m.cownerid AND lJIB
                     IF FOUND()
                        LOOP
                     ENDIF

*  If all WI owners are JIB owners, but it's a one-man item, check to see if the owner getting the expense is a royalty owner.
*  If they are, reset lcType back to R, so this expense gets processed during the revenue run closing.
                     IF lcType = 'J'
                        swselect('wellinv')
                        LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID AND ctypeinv # 'W'
                        IF FOUND()
                           lcType = 'R'
                        ENDIF
                     ENDIF
                  ENDIF

* Check for empty expense class, replace with '0'
                  IF EMPTY(cexpclass)
                     m.cexpclass = '0'
                     SELECT expense
                     REPLACE cexpclass WITH '0'
                  ENDIF

                  IF m.cexpclass # '0' AND lcType = 'J'  &&  Only check to see if the lcType flag should be re-set on non-WI expenses
                     SELECT royalty_temp
                     IF SEEK(m.cWellID)
                        DO CASE
                           CASE m.cexpclass = '1' AND lIntClass1
                              lcType = 'R'
                           CASE m.cexpclass = '2' AND lIntClass2
                              lcType = 'R'
                           CASE m.cexpclass = '3' AND lIntClass3
                              lcType = 'R'
                           CASE m.cexpclass = '4' AND lIntClass4
                              lcType = 'R'
                           CASE m.cexpclass = '5' AND lIntClass5
                              lcType = 'R'
                           CASE m.cexpclass = 'B' AND lBCPInt
                              lcType = 'R'
                           CASE m.cexpclass = 'A' AND lACPInt
                              lcType = 'R'
                           CASE m.cexpclass = 'P' AND lPlugPct
                              lcType = 'R'
                        ENDCASE
                     ENDIF
                  ENDIF


                  IF THIS.lclose
                     SELE expense
                     IF lcType = 'J'
* There are only JIB owners##
                        REPL nRunNoRev   WITH lnRunNo, ;
                           lclosed     WITH .T., ;
                           cRunYearRev WITH '1900', ;
                           cacctyear   WITH THIS.cacctyear, ;
                           cacctprd    WITH THIS.cacctprd
                     ELSE
* There are either NET owners or Both NET and JIB owners
                        REPL nRunNoRev   WITH lnRunNo, ;
                           lclosed     WITH .T., ;
                           cRunYearRev WITH THIS.crunyear, ;
                           cacctyear   WITH THIS.cacctyear, ;
                           cacctprd    WITH THIS.cacctprd
                     ENDIF

                     IF EMPTY(dacctdate)
                        REPL dacctdate WITH THIS.dacctdate
                     ENDIF

                  ENDIF

*  Total the expense classes
                  DO CASE
                     CASE m.cexpclass = '0'
                        m.nTotale = m.nTotale + m.namount
                     CASE m.cexpclass = '1'
                        m.nexpcl1 = m.nexpcl1 + m.namount
                     CASE m.cexpclass = '2'
                        m.nexpcl2 = m.nexpcl2 + m.namount
                     CASE m.cexpclass = '3'
                        m.nexpcl3 = m.nexpcl3 + m.namount
                     CASE m.cexpclass = '4'
                        m.nexpcl4 = m.nexpcl4 + m.namount
                     CASE m.cexpclass = '5'
                        m.nexpcl5 = m.nexpcl5 + m.namount
                     CASE m.cexpclass = 'A'
                        m.nexpclA = m.nexpclA + m.namount
                     CASE m.cexpclass = 'B'
                        m.nexpclB = m.nexpclB + m.namount
                     CASE m.cexpclass = 'P'
                        m.nPlugAmt = m.nPlugAmt + m.namount
                  ENDCASE
&&  Net down the salt BBL's, since the JIB run closing will be doing the same thing - BH 07/16/2008
                  m.ntotsalt = m.ntotsalt + swNetExp(m.nSaltWater, m.cWellID, .T., m.cexpclass, 'D', .F., .F., .F., m.cdeck)
               ENDSCAN

**-  Process marketing expenses
               STORE 0 TO m.nTotMKTG
               swselect('expense')
               SET ORDER TO 0
               SCAN FOR cWellID = m.cWellID ;
                     AND cdeck == lcDeck ;
                     AND nRunNoRev = 0  ;
                     AND (cyear + cperiod = lcProdYear + lcProdPeriod) ;
                     AND dexpdate <= THIS.drevdate AND ccatcode = 'MKTG'
                  SCATTER MEMVAR

                  IF THIS.lclose
                     SELE expense
                     REPL nRunNoRev   WITH lnRunNo, ;
                        lclosed     WITH .T., ;
                        cRunYearRev WITH THIS.crunyear, ;
                        cacctyear   WITH THIS.cacctyear, ;
                        cacctprd    WITH THIS.cacctprd
                     IF EMPTY(dacctdate)
                        REPL dacctdate WITH THIS.dacctdate
                     ENDIF
                  ENDIF

                  m.nTotMKTG = m.nTotMKTG + m.namount
               ENDSCAN

**-  Process compression expenses
               swselect('expense')
               SET ORDER TO 0
               SCAN FOR cWellID = m.cWellID ;
                     AND cdeck == lcDeck ;
                     AND nRunNoRev = 0  ;
                     AND (cyear + cperiod = lcProdYear + lcProdPeriod) ;
                     AND dexpdate <= THIS.drevdate AND ccatcode = 'COMP'
                  SCATTER MEMVAR


                  IF THIS.lclose
                     SELE expense
                     REPL nRunNoRev   WITH lnRunNo, ;
                        lclosed     WITH .T., ;
                        cRunYearRev WITH THIS.crunyear, ;
                        cacctyear   WITH THIS.cacctyear, ;
                        cacctprd    WITH THIS.cacctprd
                     IF EMPTY(dacctdate)
                        REPL dacctdate WITH THIS.dacctdate
                     ENDIF
                  ENDIF

                  m.nCompress = m.nCompress + m.namount
               ENDSCAN

**-  Process gathering expenses
               swselect('expense')
               SET ORDER TO 0
               SCAN FOR cWellID = m.cWellID ;
                     AND cdeck == lcDeck ;
                     AND nRunNoRev = 0  ;
                     AND (cyear + cperiod = lcProdYear + lcProdPeriod) ;
                     AND dexpdate <= THIS.drevdate AND ccatcode = 'GATH'
                  SCATTER MEMVAR


                  IF THIS.lclose
                     SELE expense
                     REPL nRunNoRev   WITH lnRunNo, ;
                        lclosed     WITH .T., ;
                        cRunYearRev WITH THIS.crunyear, ;
                        cacctyear   WITH THIS.cacctyear, ;
                        cacctprd    WITH THIS.cacctprd
                     IF EMPTY(dacctdate)
                        REPL dacctdate WITH THIS.dacctdate
                     ENDIF
                  ENDIF

                  IF m.cdeck # lcDeck
                     LOOP
                  ENDIF

                  m.nGather = m.nGather + m.namount
               ENDSCAN

**-
**-  Get the net amounts of oil and gas based upon the percentage of
**-  non-directly paid owners.  03/06/96 pws
**-
               IF m.noilinc # 0
                  lnNetOilAmt = swnetrev(m.cWellID, m.noilinc, 'O', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetOilAmt = 0
               ENDIF
               IF m.nGasInc # 0
                  lnNetGasAmt = swnetrev(m.cWellID, m.nGasInc, 'G', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetGasAmt = 0
               ENDIF

               IF m.nTotBBL # 0
                  lnNetBBL = swnetrev(m.cWellID, m.nTotBBL, 'O', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetBBL = 0
               ENDIF

               IF m.nTotMCF # 0
                  lnNETMCF = swnetrev(m.cWellID, m.nTotMCF, 'G', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNETMCF = 0
               ENDIF

               IF m.nTotMisc1 # 0
                  lnNetMisc1 = swnetrev(m.cWellID, m.nTotMisc1, '1', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetMisc1 = 0
               ENDIF

               IF m.nTotMisc2 # 0
                  lnNetMisc2 = swnetrev(m.cWellID, m.nTotMisc2, '2', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetMisc2 = 0
               ENDIF

               IF m.nOthInc # 0
                  lnNetOthAmt = swnetrev(m.cWellID, m.nOthInc, 'P', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetOthAmt = 0
               ENDIF
               IF m.nTrpInc # 0
                  lnNetTrpAmt = swnetrev(m.cWellID, m.nTrpInc, 'T', .F., .F., .F., .F., .F., .F., .F., m.cdeck)
               ELSE
                  lnNetTrpAmt = 0
               ENDIF

               IF m.nTotale # 0
*  Net out JIB and Dummy
                  m.nNetExp = swNetExp(m.nTotale, m.cWellID, .T., m.cexpclass, 'N', .F., .F., .F., m.cdeck)
*  Only net out JIB amounts
                  m.nTotale = swNetExp(m.nTotale, m.cWellID, .T., m.cexpclass, 'D', .F., .F., .F., m.cdeck)
               ELSE
                  m.nNetExp = 0
               ENDIF

               IF m.nexpcl1 # 0
                  m.nexpcl1 = swNetExp(m.nexpcl1, m.cWellID, .T., '1', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nexpcl2 # 0
                  m.nexpcl2 = swNetExp(m.nexpcl2, m.cWellID, .T., '2', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nexpcl3 # 0
                  m.nexpcl3 = swNetExp(m.nexpcl3, m.cWellID, .T., '3', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nexpcl4 # 0
                  m.nexpcl4 = swNetExp(m.nexpcl4, m.cWellID, .T., '4', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nexpcl5 # 0
                  m.nexpcl5 = swNetExp(m.nexpcl5, m.cWellID, .T., '5', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nexpclA # 0
                  m.nexpclA = swNetExp(m.nexpclA, m.cWellID, .T., 'A', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nexpclB # 0
                  m.nexpclB = swNetExp(m.nexpclB, m.cWellID, .T., 'B', 'D', .F., .F., .F., m.cdeck)
               ENDIF
               IF m.nPlugAmt # 0
                  m.nPlugAmt = swNetExp(m.nPlugAmt, m.cWellID, .T., 'P', 'D', .F., .F., .F., m.cdeck)
               ENDIF

               IF TYPE('ldPostDate') # 'D'
                  ldPostDate = DATE()
               ENDIF

               SELECT wellwork
               REPLACE nGasInc WITH  lnNetGasAmt, ;
                       ngrossgas  WITH  m.nGasInc, ;
                       noilinc    WITH  lnNetOilAmt, ;
                       ngrossoil  WITH  m.noilinc, ;
                       nTrpInc    WITH  m.nTrpInc, ;
                       nOthInc    WITH  m.nOthInc, ;
                       nMiscinc1  WITH  m.nMiscinc1, ;
                       nMiscinc2  WITH  m.nMiscinc2, ;
                       nTotMKTG   WITH  m.nTotMKTG, ;
                       nNetExp    WITH  m.nNetExp, ;
                       nTotale    WITH  m.nTotale, ;
                       nexpcl1    WITH  m.nexpcl1, ;
                       nexpcl2    WITH  m.nexpcl2, ;
                       nexpcl3    WITH  m.nexpcl3, ;
                       nexpcl4    WITH  m.nexpcl4, ;
                       nexpcl5    WITH  m.nexpcl5, ;
                       nexpclA    WITH  m.nexpclA, ;
                       nexpclB    WITH  m.nexpclB, ;
                       nPlugAmt   WITH  m.nPlugAmt, ;
                       ntotbbltx1 WITH  m.nBBLTax1, ;
                       ntotmcftx1 WITH  m.nMCFTax1, ;
                       ntotothtx1 WITH  m.nOthTax1, ;
                       ntotbbltx2 WITH  m.nBBLTax2, ;
                       ntotmcftx2 WITH  m.nMCFTax2, ;
                       ntotothtx2 WITH  m.nOthTax2, ;
                       ntotbbltx3 WITH  m.nBBLTax3, ;
                       ntotmcftx3 WITH  m.nMCFTax3, ;
                       ntotothtx3 WITH  m.nOthTax3, ;
                       ntotbbltx4 WITH  m.nBBLTax4, ;
                       ntotmcftx4 WITH  m.nMCFTax4, ;
                       ntotothtx4 WITH  m.nOthTax4, ;
                       ntotbbltxR WITH  m.nBBLTaxr, ;
                       ntotmcftxR WITH  m.nMCFTaxr, ;
                       ntotbbltxW WITH  m.nBBLTaxw, ;
                       ntotmcftxW WITH  m.nMCFTaxw, ;
                       nGather    WITH  m.nGather   + lnGathering, ;
                       nCompress  WITH  m.nCompress + lnCompression, ;
                       nPlugAmt   WITH  m.nPlugAmt, ;
                       nTotBBL    WITH  m.nTotBBL, ;
                       nTotMCF    WITH  m.nTotMCF, ;
                       nTotProd   WITH  m.nTotProd, ;
                       nTotMisc1  WITH  m.nTotMisc1, ;
                       nTotMisc2  WITH  m.nTotMisc2, ;
                       nGrossBBL  WITH  m.nTotBBL, ;
                       nGrossMCF  WITH  m.nTotMCF, ;
                       noilint    WITH  m.noilint, ;
                       ngasint    WITH  m.ngasint, ;
                       nDaysOn    WITH  lnDaysOn, ;
                       ntotsalt   WITH  m.ntotsalt, ;
                       nflatgas   WITH  m.nflatgas, ;
                       nflatoil   WITH  m.nflatoil, ;
                       nGBBLTax1  WITH  m.noiltax1, ;
                       nGMCFTax1  WITH  m.ngastax1, ;
                       nGBBLTax2  WITH  m.noiltax2, ;
                       nGMCFTax2  WITH  m.ngastax2, ;
                       nGBBLTax3  WITH  m.noiltax3, ;
                       nGMCFTax3  WITH  m.ngastax3, ;
                       nGBBLTax4  WITH  m.noiltax4, ;
                       nGMCFTax4  WITH  m.ngastax4, ;
                       nGOTHTax1  WITH  m.nprodtax1, ;
                       nGOTHTax2  WITH  m.nprodtax2, ;
                       nGOTHTax3  WITH  m.nprodtax3, ;
                       nGOTHTax4  WITH  m.nprodtax4, ;
                       hdate      WITH  THIS.dpostdate
            ENDSCAN  && Wellwork
         ENDSCAN  && Wellamounts


         THIS.oprogress.SetProgressMessage('Allocating Revenue and Expenses to Wells')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'WellProc', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('WellProc', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Process Owners
*********************************
   PROCEDURE OwnerProc
*********************************
      LOCAL lnflatoilamt, lnflatgasamt, lcSavePrd1, lcSavePrd2, oprogress, lnRunNo, llrelqtr, llFlatReleased, llFlatAllocated
      LOCAL llExpSum, llRevSum, llRoyComp, llOperator, ldHistDate
      LOCAL lExempt, latemp[1], lbackwith, lcdirect, lcowner, lcperiod1, lcperiod2, lcyear1, lcyear2
      LOCAL llReturn, llincome, llroyaltyowner, llroysevtx, lnCompress, lnCount, lnGather, lnMKTGExp
      LOCAL lnMax, lnOilTax, lnPct, lnX, lngasexp, lngasinc, lngasmantax1, lngasmantax2, lngasmantax3
      LOCAL lngasmantax4, lngasrevenue, lngastax1, lngastax2, lngastax3, lngastax4, lnmi1inc, lnmi2inc
      LOCAL lnoilexp, lnoilinc, lnoilmantax1, lnoilmantax2, lnoilmantax3, lnoilmantax4, lnoilrevenue
      LOCAL lnoiltax1, lnoiltax2, lnoiltax3, lnoiltax4, lnothinc, lnothmantax1, lnothmantax2, lnothmantax3
      LOCAL lnothmantax4, lnothtax1, lnothtax2, lnothtax3, lnothtax4, lntaxes, lntaxgross, lntotexp
      LOCAL lntotinc, lntrpinc, lnwelltot, loError, loneman, ltaxgross, ltaxwith
      LOCAL  cownerid, cSource, category, cdescrip, cdescript, cdirect, cmiscmemo, ctype, gncompress
      LOCAL  gngather, jcownerid, jcwellid, jflatcnt, jgross, jinvcnt, jnclass1, jnclass2, jnclass3
      LOCAL  jnclass4, jnclass5, jnclassa, jnclassb, jncompress, jnflatcount, jngather, jnrevgas, jnPlugPct
      LOCAL  jnrevm1, jnrevm2, jnrevoil, jnrevoth, jnrevtax1, jnrevtax10, jnrevtax11, jnrevtax12
      LOCAL  jnrevtax2, jnrevtax3, jnrevtax4, jnrevtax5, jnrevtax6, jnrevtax7, jnrevtax8, jnrevtax9
      LOCAL  jnrevtrans, jnworkint, junits, jwrk, nUnits, nacpint, nbackpct, nbackwith, nbcpint, nexpcl1
      LOCAL  nexpcl2, nexpcl3, nexpcl4, nexpcl5, nexpclA, nexpclB, nflatrate, ngasrev, ngastax1, nPlugAmt
      LOCAL  ngastax1a, ngastax2, ngastax2a, ngastax3, ngastax3a, ngastax4, ngastax4a, nintclass1
      LOCAL  nintclass2, nintclass3, nintclass4, nintclass5, ninvamt, nmiscrev1, nmiscrev2, noilrev
      LOCAL  noiltax1, noiltax1a, noiltax2, noiltax2a, noiltax3, noiltax3a, noiltax4, noiltax4a, nothrev
      LOCAL  nprice, nprocess, nprodtax1, nprodtax1a, nprodtax2, nprodtax2a, nprodtax3, nprodtax3a
      LOCAL  nprodtax4, nprodtax4a, nprodwell, nrevgas, nrevgtax, nrevint, nrevmisc1, nrevmisc2, nrevoil
      LOCAL  nrevotax, nrevoth, nrevtax1, nrevtax10, nrevtax11, nrevtax12, nrevtax2, nrevtax3, nrevtax4
      LOCAL  nrevtax5, nrevtax6, nrevtax7, nrevtax8, nrevtax9, nrevtrp, nroyalty, nroyint, ntaxpct, nplugpct
      LOCAL  ntaxwith, ntotal, ntotale1, ntotale2, ntotale3, ntotale4, ntotale5, ntotalea, ntotaleb
      LOCAL  ntotbbltx1, ntotbbltx2, ntotbbltx3, ntotbbltx4, ntotbbltxR, ntotbbltxW, ntotmcftx1
      LOCAL  ntotmcftx2, ntotmcftx3, ntotmcftx4, ntotmcftxR, ntotmcftxW, ntotothtx1, ntotothtx2
      LOCAL  ntotothtx3, ntotothtx4, ntrprev, nworkint, nworktot, lcDeck

      llReturn = .T.
      
      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         THIS.oprogress.SetProgressMessage('Allocating Revenue and Expenses to the Owners...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

         lcSavePrd1 = THIS.cperiod
         lcSavePrd2 = lcSavePrd1
         lcyear1    = THIS.crunyear
         lcyear2    = THIS.crunyear
         lnRunNo    = THIS.nrunno

*  Get options
         gncompress = THIS.oOptions.nCompress
         gngather   = THIS.oOptions.nGather
         llRoyComp  = THIS.oOptions.lroycomp
         llExpSum   = THIS.oOptions.lexpsum
         llRevSum   = THIS.oOptions.lRevSum

         Make_Copy('expense', 'exptemp')
         Make_Copy('expense', 'mktgtemp')
         Make_Copy('expense', 'comptemp')
         Make_Copy('expense', 'gathtemp')

         swselect('expcat')
         SET ORDER TO ccatcode

****************************************************************************
*  Setup the expense file. Either create a summary of expenses or a temp
*  work cursor. We do this because if the expenses are being summarized
*  we have to process them as summarized so the rounding matches what shows
*  on their statement.
****************************************************************************
         IF llExpSum  AND (NOT THIS.lclose OR m.goApp.lPartnershipMod)       && Summarize expenses
            swselect('expense')
            SCAN FOR BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear) ;
                  AND cyear # 'FIXD'
               SCATTER MEMVAR
               IF INLIST(m.ccatcode, 'MKTG', 'COMP', 'GATH')
                  LOOP
               ENDIF
               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip

* Don't process JIB only expenses during a revenue run (v6089)
                  IF expcat.lJIBOnly
                     LOOP
                  ENDIF

               ELSE
                  LOOP
               ENDIF
               m.loneman = .F.

* Check to make sure the deck matches the record we're working with
               IF EMPTY(m.cdeck)
                  m.cdeck = THIS.oWellInv.DOIDeckNameLookup(m.cyear, m.cperiod, m.cWellID)
               ENDIF

               SELE exptemp
               LOCATE FOR cWellID = m.cWellID AND cdeck == m.cdeck AND ;
                  ccatcode == m.ccatcode AND cownerid = m.cownerid AND cexpclass = m.cexpclass AND cperiod = m.cperiod AND cyear = m.cyear
               IF FOUND()
                  REPL namount WITH namount + m.namount
               ELSE
                  INSERT INTO exptemp FROM MEMVAR
               ENDIF
            ENDSCAN

* Create a summary cursor of marketing expenses
            swselect('expense')
            SCAN FOR ccatcode = 'MKTG' ;
                  AND BETWEEN(expense.cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear)
               SCATTER MEMVAR

               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip
               ELSE
                  LOOP
               ENDIF
               m.loneman = .F.
               SELE mktgtemp
               LOCATE FOR cWellID = m.cWellID AND cyear + cperiod = m.cyear + m.cperiod AND cdeck == m.cdeck AND ;
                  ccatcode == m.ccatcode AND cownerid = m.cownerid AND cexpclass = m.cexpclass AND cperiod = m.cperiod AND cyear = m.cyear
               IF FOUND()
                  REPL namount WITH namount + m.namount
               ELSE
                  INSERT INTO mktgtemp FROM MEMVAR
               ENDIF
            ENDSCAN

* Create a summary cursor of compression expenses
            swselect('expense')
            SCAN FOR ccatcode = 'COMP' ;
                  AND BETWEEN(expense.cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear)
               SCATTER MEMVAR

               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip
               ELSE
                  LOOP
               ENDIF
               m.loneman = .F.
               SELE comptemp
               LOCATE FOR cWellID = m.cWellID AND cyear + cperiod = m.cyear + m.cperiod AND cdeck == m.cdeck AND ;
                  ccatcode == m.ccatcode AND cownerid = m.cownerid AND cexpclass = m.cexpclass AND cperiod = m.cperiod AND cyear = m.cyear
               IF FOUND()
                  REPL namount WITH namount + m.namount
               ELSE
                  INSERT INTO comptemp FROM MEMVAR
               ENDIF
            ENDSCAN

* Create a summary cursor of gathering expenses
            swselect('expense')
            SCAN FOR ccatcode = 'GATH' ;
                  AND BETWEEN(expense.cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear)
               SCATTER MEMVAR

               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip
               ELSE
                  LOOP
               ENDIF
               m.loneman = .F.
               SELE gathtemp
               LOCATE FOR cWellID = m.cWellID AND cyear + cperiod = m.cyear + m.cperiod AND cdeck == m.cdeck AND ;
                  ccatcode == m.ccatcode AND cownerid = m.cownerid AND cexpclass = m.cexpclass AND cperiod = m.cperiod AND cyear = m.cyear
               IF FOUND()
                  REPL namount WITH namount + m.namount
               ELSE
                  INSERT INTO gathtemp FROM MEMVAR
               ENDIF
            ENDSCAN

         ELSE

* Create a temp cursor of expenses
            SELECT  expense.*,;
                    expcat.cdescrip;
                FROM expense WITH (BUFFERING = .T.);
                JOIN expcat;
                    ON expcat.ccatcode = expense.ccatcode ;
                INTO CURSOR exptemp READWRITE ;
                WHERE BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                    AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear) ;
                    AND cyear # 'FIXD' ;
                    AND expcat.lJIBOnly = .F. ;
                ORDER BY cWellID


* Create a temp cursor of marketing expenses
            swselect('expense')
            SCAN FOR ccatcode = 'MKTG' AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear)
               SCATTER MEMVAR
               m.loneman = .F.
               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip
               ELSE
                  LOOP
               ENDIF

               INSERT INTO mktgtemp FROM MEMVAR
            ENDSCAN

* Create a temp cursor of compression expenses
            swselect('expense')
            SCAN FOR ccatcode = 'COMP' AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear)
               SCATTER MEMVAR
               m.loneman = .F.
               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip
               ELSE
                  LOOP
               ENDIF

               INSERT INTO comptemp FROM MEMVAR
            ENDSCAN

* Create a temp cursor of gathering expenses
            swselect('expense')
            SCAN FOR ccatcode = 'GATH' AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                  AND IIF(THIS.lNewRun, nRunNoRev = 0 AND dexpdate <= THIS.dexpdate, nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear)
               SCATTER MEMVAR
               m.loneman = .F.
               SELE expcat
               IF SEEK(m.ccatcode)
                  m.cdescrip  = cdescrip
               ELSE
                  LOOP
               ENDIF

               INSERT INTO gathtemp FROM MEMVAR
            ENDSCAN
         ENDIF

* Add compression expenses that were entered as revenue - for backward compatibility
         swselect('income')
         SCAN FOR cSource = 'COMP' ;
               AND BETWEEN(income.cWellID, THIS.cbegwellid, THIS.cendwellid) ;
               AND IIF(THIS.lNewRun, nrunno = 0 AND drevdate <= THIS.drevdate, nrunno = THIS.nrunno AND crunyear = THIS.crunyear)
            SCATTER MEMVAR

            m.loneman = .F.
            SELE comprev
            LOCATE FOR cWellID = m.cWellID AND cyear + cperiod = m.cyear + m.cperiod AND cdeck == m.cdeck AND ;
               cSource == m.cSource AND cownerid = m.cownerid AND cperiod = m.cperiod AND cyear = m.cyear
            IF FOUND()
               REPL nTotalInc WITH nTotalInc + m.nTotalInc
            ELSE
               INSERT INTO comprev FROM MEMVAR
            ENDIF
         ENDSCAN

* Add gathering charges that were entered as revenue - for backward compatibility
         swselect('income')
         SCAN FOR cSource = 'GATH' ;
               AND BETWEEN(income.cWellID, THIS.cbegwellid, THIS.cendwellid) ;
               AND IIF(THIS.lNewRun, nrunno = 0 AND drevdate <= THIS.drevdate, nrunno = THIS.nrunno AND crunyear = THIS.crunyear)
            SCATTER MEMVAR

            m.loneman = .F.
            SELE gathrev
            LOCATE FOR cWellID = m.cWellID AND cyear + cperiod = m.cyear + m.cperiod AND cdeck == m.cdeck AND ;
               cSource == m.cSource AND cownerid = m.cownerid AND cperiod = m.cperiod AND cyear = m.cyear
            IF FOUND()
               REPL nTotalInc WITH nTotalInc + m.nTotalInc
            ELSE
               INSERT INTO gathrev FROM MEMVAR
            ENDIF
         ENDSCAN

****************************************************************************
*  Calculate individual working and royalty interest shares by owner
****************************************************************************

         lnCount = 1
         swselect('wells')
         SET ORDER TO cWellID

*  Process revenue into temp cursor

         swselect('income')
         Make_Copy('income', 'inctemp')

         IF llRevSum
* Create a summary of revenue cursor
            SELE * FROM income  with (Buffering = .T.) ;
               WHERE IIF(THIS.lNewRun, nrunno = 0 AND drevdate <= THIS.drevdate, nrunno = THIS.nrunno AND crunyear = THIS.crunyear) ;
               AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
               INTO CURSOR inctmp READWRITE ;
               ORDER BY cWellID, cownerid
            IF _TALLY > 0
               SELE inctmp
               SCAN
                  SCATTER MEMVAR
                  m.loneman = .F.

* Check to make sure the deck matches the record we're working with
                  IF EMPTY(inctmp.cdeck)
                     m.cdeck = THIS.oWellInv.DOIDeckNameLookup(inctmp.cyear, inctmp.cperiod, inctmp.cWellID)
                  ENDIF

                  IF EMPTY(m.cownerid)
                     SELECT inctemp
                     LOCATE FOR cWellID == m.cWellID AND cSource == m.cSource AND cyear == m.cyear AND cperiod == m.cperiod AND cdeck == m.cdeck
                     IF FOUND()
                        REPLACE nUnits WITH nUnits + m.nUnits, ;
                                nTotalInc WITH nTotalInc + m.nTotalInc
                     ELSE
                        INSERT INTO inctemp FROM MEMVAR
                     ENDIF
                  ELSE
                     INSERT INTO inctemp FROM MEMVAR
                  ENDIF
               ENDSCAN
               SELE inctemp
               INDEX ON cWellID TAG cWellID
               INDEX ON cyear + cperiod TAG yearprd
            ENDIF
         ELSE
* Create a temp cursor of non-summarized revenue
            SELE * FROM income with (buffering = .T. );
               WHERE IIF(THIS.lNewRun, nrunno = 0 AND drevdate <= THIS.drevdate, nrunno = THIS.nrunno AND crunyear = THIS.crunyear) ;
               AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
               INTO CURSOR inctemp READWRITE ;
               ORDER BY cWellID
            SELE inctemp
            REPLACE loneman WITH .F. ALL
            INDEX ON cWellID TAG cWellID
            INDEX ON cyear + cperiod TAG yearprd
         ENDIF
         STORE .F. TO m.lExempt, m.ltaxwith, m.lbackwith, m.ltaxgross
         STORE 0 TO m.ntaxpct, m.nbackpct, m.nplugpct

         jcownerid = '  '
         swselect('invtmp')
         COUNT FOR NOT DELETED() TO lnMax

         jcwellid        = '  '
         llFlatReleased  = .F.
         llFlatAllocated = .F.
         llOperator      = .F.

* Make sure investor table is open
         swselect('investor')
         SET ORDER TO cownerid

********************************************************************
*  Start the scan through the invtmp file to process each interest.
********************************************************************
         SELECT invtmp
         SET ORDER TO invprog
         SCAN FOR BETWEEN(cownerid, THIS.cbegownerid, THIS.cendownerid)
            SCATTER MEMVAR
            lcDeck = m.cdeck
            THIS.oprogress.SetProgressMessage('Allocating Revenue and Expenses to the Owners...' + m.cownerid)

            IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
               llReturn          = .F.
               IF NOT m.goApp.CancelMsg()
                  THIS.lCanceled = .T.
                  EXIT
               ENDIF
            ENDIF

            IF m.cownerid # jcownerid
* Look up the owner info when the owner changes.
               jcownerid = m.cownerid
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(jcownerid)
                  m.lExempt       = lExempt
                  m.lbackwith     = lbackwith
                  m.nbackpct      = nbackpct
                  m.ltaxgross     = ltaxgross
                  llFlatReleased  = .F.
                  llFlatAllocated = .F.
                  llOperator      = lIntegGL
* Can't set a posting owner for backup wh - pws 2/26/13
                  IF llOperator
                     m.lbackwith = .F.
                  ENDIF
               ELSE
*  Shouldn't ever get here
                  LOOP
               ENDIF
            ENDIF

            IF m.cWellID # jcwellid
               jcwellid = m.cWellID
* Reset Flat Income flags when the well changes.
               llFlatReleased  = .F.
               llFlatAllocated = .F.
            ENDIF

            STORE 0 TO lnwelltot, lntotinc, lntotexp, m.ninvamt, jinvcnt, lnoilinc, ;
               lngasinc, lnothinc, lntrpinc, lntaxes, jnflatcount, lnmi1inc, lnmi2inc, ;
               lnCompress, lnGather, lnoilexp, lngasexp, lnCompression, lnGathering
            STORE 0 TO lngastax1, lnoiltax1, lnothtax1, lnoiltax2, lngastax2, lnothtax2, ;
               lngastax3, lnoiltax3, lnothtax3, lnoiltax4, lngastax4, lnothtax4
            STORE 0 TO lngasmantax1, lnoilmantax1, lnothmantax1, lnoilmantax2, lngasmantax2, lnothmantax2, ;
               lngasmantax3, lnoilmantax3, lnothmantax3, lnoilmantax4, lngasmantax4, lnothmantax4

            SELECT invtmp
            lcdirect  = m.cdirect
            jnworkint = m.nworkint
            STORE 0 TO lnflatoilamt, lnflatgasamt

*  Check to see if this owner is paid a flat rate
            IF m.lflat AND NOT llFlatReleased
               m.nflatrate = THIS.getownflat(m.cWellID, jcownerid, m.ctypeint, m.ciddisb)
               SELECT invtmp
               REPL nrevoil WITH 0, nrevgas WITH 0, nflatrate WITH m.nflatrate, nIncome WITH nIncome + m.nflatrate

               lntotinc = lntotinc + m.nflatrate  &&  Add the flat rate to the running income total

* Set the flat released flag so that a flat rate is only released
* once per owner/well combination.
               llFlatReleased = .T.
            ELSE
               m.nflatrate = 0
            ENDIF

*  Determine if the owner is a royalty owner
            IF m.ctypeinv = 'L' OR m.ctypeinv = 'O'
               llroyaltyowner = .T.
            ELSE
               llroyaltyowner = .F.
            ENDIF

            STORE 0 TO jflatcnt
            STORE 'N' TO m.cdirect

* Get the well information for the well being processed
            SELE wells
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               llroysevtx = lroysevtx
               lcowner    = m.cownerid
               SCATTER MEMVAR
               m.cownerid = lcowner
               m.nroyint  = m.nlandpct + m.noverpct
            ELSE
* Should never get here...
               LOOP
            ENDIF

*  Get the taxes calculated for the well
            SELECT wellwork
            LOCATE FOR cWellID = m.cWellID AND hyear = m.hyear AND hperiod = m.hperiod AND cdeck == lcDeck
            IF FOUND()
               m.nprodwell  = nGasInc + noilinc + nTrpInc
               m.ntotbbltx1 = nGBBLTax1
               m.ntotmcftx1 = nGMCFTax1
               m.ntotbbltx2 = nGBBLTax2
               m.ntotmcftx2 = nGMCFTax2
               m.ntotbbltx3 = nGBBLTax3
               m.ntotmcftx3 = nGMCFTax3
               m.ntotbbltx4 = nGBBLTax4
               m.ntotmcftx4 = nGMCFTax4
               m.ntotothtx1 = nGOTHTax1
               m.ntotothtx2 = nGOTHTax2
               m.ntotothtx3 = nGOTHTax3
               m.ntotothtx4 = nGOTHTax4
               m.ntotbbltxR = ntotbbltxR
               m.ntotmcftxR = ntotmcftxR
               m.ntotbbltxW = ntotbbltxW
               m.ntotmcftxW = ntotmcftxW
               lngasrevenue = ngrossgas
               lnoilrevenue = ngrossoil
               lnflatoilamt = nflatoil
               lnflatgasamt = nflatgas
               m.nroyalty   = nroyint
               m.nprocess   = nprocess
               jncompress   = nCompress
               jngather     = nGather
               m.nexpcl1    = nexpcl1
               m.nexpcl2    = nexpcl2
               m.nexpcl3    = nexpcl3
               m.nexpcl4    = nexpcl4
               m.nexpcl5    = nexpcl5
               m.nexpclA    = nexpclA
               m.nexpclB    = nexpclB
               m.nPlugAmt   = nPlugAmt
            ELSE
* Should never really get here...
               m.nprodwell  = 0
               STORE 0 TO m.ntotbbltx1, m.ntotbbltx2, m.ntotbbltx3, m.ntotbbltx4
               STORE 0 TO m.ntotmcftx1, m.ntotmcftx2, m.ntotmcftx3, m.ntotmcftx4
               STORE 0 TO m.ntotothtx1, m.ntotothtx2, m.ntotothtx3, m.ntotothtx4
               STORE 0 TO m.nexpcl1, m.nexpcl2, m.nexpcl3, m.nexpcl4, m.nexpcl5
               STORE 0 TO m.nexpclA, m.nexpclB, m.ntotbbltxR, m.ntotmcftxR
               STORE 0 TO m.ntotbbltxW, m.ntotmcftxW, m.ntotal, m.nroyalty, m.nPlugAmt
               STORE 0 TO jncompress, jngather, lngasrevenue, lnoilrevenue, lnflatgasamt, lnflatoilamt
               m.nprocess   = 1
            ENDIF

**********************************************************
*  Process income
**********************************************************

            SELECT inctemp
            SET ORDER TO 0
            jflatcnt = 0
            llincome = .F.
            STORE 0 TO m.ngasrev, m.noilrev, m.ntrprev, m.nmiscrev1, m.nmiscrev2, m.nothrev
            SCAN FOR cWellID = m.cWellID AND cyear + cperiod = m.hyear + m.hperiod AND cdeck == lcDeck
               m.ctype = cSource

* Check to make sure this owner has the interest type required to process this
* revenue entry. If he doesn't, loop out.
               DO CASE
                  CASE m.ctype = 'BBL'
                     IF m.ctypeint = 'G'
                        LOOP
                     ENDIF
                  CASE m.ctype = 'MCF'
                     IF m.ctypeint = 'O'
                        LOOP
                     ENDIF
               ENDCASE

* Save the original interests so we can manipulate the interests
* for one-man items.
               jnrevoil   = m.nrevoil
               jnrevgas   = m.nrevgas
               jnrevoth   = m.nrevoth
               jnrevtax1  = m.nrevtax1
               jnrevtax2  = m.nrevtax2
               jnrevtax3  = m.nrevtax3
               jnrevtax4  = m.nrevtax4
               jnrevtax5  = m.nrevtax5
               jnrevtax6  = m.nrevtax6
               jnrevtax7  = m.nrevtax7
               jnrevtax8  = m.nrevtax8
               jnrevtax9  = m.nrevtax9
               jnrevtax10 = m.nrevtax10
               jnrevtax11 = m.nrevtax11
               jnrevtax12 = m.nrevtax12
               jnrevm1    = m.nrevmisc1
               jnrevm2    = m.nrevmisc2
               jnrevtrans = m.nrevtrp
*
*  Get the revenue interest for this owner
*
               DO CASE
                  CASE cownerid = jcownerid AND NOT loneman
* The owner on this income record matches the owner
* we're processing and we haven't processed this entry
* for the owner yet.

* Check to see if the owner has multiple interests
* If so, use the working interest
                     IF m.ctypeinv # 'W'
                        swselect('wellinv')
                        LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID AND ctypeinv = 'W'
                        IF FOUND()
                           LOOP
                        ENDIF
                     ENDIF
                     SELECT inctemp
                     REPLACE loneman WITH .T.
                     DO CASE
                        CASE m.ctype = 'BBL'
                           m.nrevoil = 100
                        CASE m.ctype = 'MCF'
                           m.nrevgas = 100
                        CASE m.ctype = 'OTH'
                           m.nrevoth = 100
                        CASE m.ctype = 'EXO'
                           m.nrevoil = 100
                        CASE m.ctype = 'EXG'
                           m.nrevgas = 100
                        CASE m.ctype = 'TRANS'
                           m.nrevtrp = 100
                        CASE m.ctype = 'OTAX1'
                           m.nrevtax1 = 100
                        CASE m.ctype = 'GTAX1'
                           m.nrevtax2 = 100
                        CASE m.ctype = 'PTAX1'
                           m.nrevtax3 = 100
                        CASE m.ctype = 'OTAX2'
                           m.nrevtax4 = 100
                        CASE m.ctype = 'GTAX2'
                           m.nrevtax5 = 100
                        CASE m.ctype = 'PTAX2'
                           m.nrevtax6 = 100
                        CASE m.ctype = 'OTAX3'
                           m.nrevtax7 = 100
                        CASE m.ctype = 'GTAX3'
                           m.nrevtax8 = 100
                        CASE m.ctype = 'PTAX3'
                           m.nrevtax9 = 100
                        CASE m.ctype = 'OTAX4'
                           m.nrevtax10 = 100
                        CASE m.ctype = 'GTAX4'
                           m.nrevtax11 = 100
                        CASE m.ctype = 'PTAX4'
                           m.nrevtax12 = 100
                        CASE m.ctype = 'MISC1'
                           m.nrevmisc1 = 100
                        CASE m.ctype = 'MISC2'
                           m.nrevmisc2 = 100
                     ENDCASE

                  CASE cownerid = ' '
* There is no ownerid on the income entry.
* It's not a one-man item.
                  CASE cownerid # jcownerid
* The owner on the record doesn't match the
* owner we're processing....loop out.
                     LOOP
                  CASE cownerid = jcownerid AND loneman
* We've already processed this entry for this owner.
                     LOOP
               ENDCASE

               llincome = .T.

               DO CASE
                  CASE m.ctype = 'OTAX1'
                     m.noiltax1 = m.noiltax1 + inctemp.nTotalInc

                  CASE m.ctype = 'GTAX1'
                     m.ngastax1 = m.ngastax1 + inctemp.nTotalInc

                  CASE m.ctype = 'OTAX2'
                     m.noiltax2 = m.noiltax2 + inctemp.nTotalInc

                  CASE m.ctype = 'GTAX2'
                     m.ngastax2 = m.ngastax2 + inctemp.nTotalInc

                  CASE m.ctype = 'OTAX3'
                     m.noiltax3 = m.noiltax3 + inctemp.nTotalInc

                  CASE m.ctype = 'GTAX3'
                     m.ngastax3 = m.ngastax3 + inctemp.nTotalInc

                  CASE m.ctype = 'OTAX4'
                     m.noiltax4 = m.noiltax4 + inctemp.nTotalInc

                  CASE m.ctype = 'GTAX4'
                     m.ngastax4 = m.ngastax4 + inctemp.nTotalInc

               ENDCASE

***********************************************************
*  Plug in the correct interest for the revenue
*  being processed.
***********************************************************
               DO CASE
                  CASE m.ctype = 'BBL'
                     m.nrevint = m.nrevoil
                     llincome  = .T.
                  CASE m.ctype = 'MCF'
                     m.nrevint = m.nrevgas
                     llincome  = .T.
                  CASE m.ctype = 'OTH'
                     m.nrevint = m.nrevoth
                     llincome  = .T.
                  CASE m.ctype = 'TRANS'
                     m.nrevint = m.nrevtrp
                     llincome  = .T.
                  CASE m.ctype = 'NOT'
                     m.nrevint = 1
                  CASE m.ctype = 'MISC1'
                     m.nrevint = m.nrevmisc1
                     llincome  = .T.
                  CASE m.ctype = 'MISC2'
                     m.nrevint = m.nrevmisc2
                     llincome  = .T.
                  CASE m.ctype = 'OTAX1'
                     m.noiltax1 = m.noiltax1 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax1
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax1
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'GTAX1'
                     m.ngastax1 = m.ngastax1 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax2
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax2
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'PTAX1'
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax3
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax3
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'OTAX2'
                     m.noiltax2 = m.noiltax2 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax4
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax4
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'GTAX2'
                     m.ngastax2 = m.ngastax2 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax5
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax5
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'PTAX2'
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax6
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax6
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'OTAX3'
                     m.noiltax3 = m.noiltax3 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax7
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax7
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'GTAX3'
                     m.ngastax3 = m.ngastax3 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax8
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax8
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'PTAX3'
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax9
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax9
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'OTAX4'
                     m.noiltax4 = m.noiltax4 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax10
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax10
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'GTAX4'
                     m.ngastax4 = m.ngastax4 + nTotalInc
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax11
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax11
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'PTAX4'
                     IF NOT llroyaltyowner
                        IF llroysevtx
                           m.nrevint = m.nworkint
                        ELSE
                           m.nrevint = m.nrevtax12
                        ENDIF
                     ELSE
                        IF llroysevtx
                           m.nrevint = 0
                        ELSE
                           m.nrevint = m.nrevtax12
                        ENDIF
                     ENDIF
                     llincome = .T.
                  CASE m.ctype = 'EXG'
                     m.nrevint = m.nworkint
                  CASE m.ctype = 'EXO'
                     m.nrevint = m.nworkint
                  OTHERWISE
                     m.nrevint = m.nrevgas
               ENDCASE

*  If the interest type is oil and gas is being processed,
*  change the interests to zero.
               IF m.ctypeint = 'O' AND INLIST(m.ctype, 'MCF', 'GTAX1', 'GTAX2', 'GTAX3', 'GTAX4', 'EXG')
                  m.nrevint = 0
                  m.nrevgas = 0
               ENDIF

*  If the interest type is gas and oil is being processed,
*  change the interests to zero.
               IF m.ctypeint = 'G' AND INLIST(m.ctype, 'BBL', 'OTAX1', 'OTAX2', 'OTAX3', 'OTAX4', 'EXO')
                  m.nrevint = 0
                  m.nrevoil = 0
               ENDIF

* Setup the revenue variables used in processing
               SELE inctemp
               jgross     = nTotalInc
               junits     = nUnits
               m.nworktot = 0

*  Check if this investor is a direct pay
               DO CASE
                  CASE INLIST(lcdirect, 'O', 'B') AND ;
                        INLIST(m.ctype, 'BBL', 'OTAX1', 'OTAX2', 'OTAX3', 'OTAX4', 'EXO') AND NOT m.lflat
*
*  If the owner is directly paid oil revenue and oil is being processed, zero out his totals on revenue
*
                     m.ninvamt  = 0
                     m.ntotal   = 0
                     m.nworktot = 0
                     DO CASE
                        CASE m.ctype = 'OTAX1'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot = 0
                              lnoiltax1  = 0
                              m.nrevtax1 = 0
                           ELSE
                              IF NOT m.lsev1o
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser pays revenue directly, so mark as directly paid
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid revenue
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax1 = m.nrevint
                           ENDIF
                        CASE m.ctype = 'OTAX2'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot = 0
                              lnoiltax2  = 0
                              m.nrevtax4 = m.nrevint
                           ELSE
                              IF NOT m.lsev2o
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser pays directly, so mark as directly paid
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid revenue
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax4 = m.nrevint
                           ENDIF
                        CASE m.ctype = 'OTAX3'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot = 0
                              lnoiltax3  = 0
                              m.nrevtax7 = 0
                           ELSE
                              IF NOT m.lsev3o
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser pay, so mark as directly paid
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid revenue
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax7 = m.nrevint
                           ENDIF
                        CASE m.ctype = 'OTAX4'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot  = 0
                              lnoiltax4   = 0
                              m.nrevtax10 = 0
                           ELSE
                              IF NOT m.lsev4o
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser pays directly, so mark as directly paid
                                 IF NOT m.lDirOilPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid revenue
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax10 = m.nrevint
                           ENDIF
                        OTHERWISE
* Oil Revenue, not taxes
                           m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                     ENDCASE

                  CASE INLIST(lcdirect, 'G', 'B') AND ;
                        INLIST(m.ctype, 'MCF', 'GTAX1', 'GTAX2', 'GTAX3', 'GTAX4', 'EXG') AND NOT m.lflat
*  If the owner is directly paid gas revenue and gas is being processed, zero out his totals on income
                     m.ninvamt = 0
                     m.ntotal  = 0
                     DO CASE
                        CASE m.ctype = 'GTAX1'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot   = 0
                              lngasmantax1 = 0
                              m.nrevtax2   = 0
                           ELSE
                              IF NOT m.lsev1g
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser withholds, so mark as directly paid
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid tax
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax2 = m.nrevint
                           ENDIF
                        CASE m.ctype = 'GTAX2'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot = 0
                              lngastax2  = 0
                              m.nrevtax5 = 0
                           ELSE
                              IF NOT m.lsev2g
* Purchaser doesn't withhold, so subtract it
                                 m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                              ELSE
* Purchaser withholds, so mark as directly paid
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid tax
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax5 = m.nrevint
                           ENDIF
                        CASE m.ctype = 'GTAX3'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot = 0
                              lngastax3  = 0
                              m.nrevtax8 = 0
                           ELSE
                              IF NOT m.lsev3g
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser withholds, so mark as directly paid
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid tax
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax8 = m.nrevint
                           ENDIF
                        CASE m.ctype = 'GTAX4'
                           IF m.lExempt
* Owner is exempt from tax. Zero out taxes
                              m.nworktot  = 0
                              lngastax4   = 0
                              m.nrevtax11 = 0
                           ELSE
                              IF NOT m.lsev4g
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold, so subtract it
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ENDIF
                              ELSE
* Purchaser withholds, so mark as directly paid
                                 IF NOT m.lDirGasPurch
* Purchaser doesn't withhold tax on directly paid owners
                                    m.ninvamt  = swround((jgross * (m.nrevint / 100)), 2)
                                 ELSE
* Purchaser withholds tax on direct paid tax
                                    m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                                 ENDIF
                              ENDIF
                              m.nrevtax11 = m.nrevint
                           ENDIF
                        OTHERWISE
                           m.nworktot = swround(jgross * (m.nrevint / 100), 2)
                     ENDCASE

                  CASE INLIST(lcdirect, 'G', 'B') AND ;
                        m.ctype = 'MCF' AND m.lflat AND m.ctypeint = 'G'
*  This owner is a gas owner and paid a flat rate
                     IF jflatcnt = 0
                        m.ninvamt  = swround(m.nflatrate * (m.nworkint / 100), 2)
                        m.ntotal   = 0
                        m.nworktot = m.nflatrate
                        jflatcnt   = 1
                     ELSE
                        m.ninvamt  = 0
                        m.nworktot = 0
                     ENDIF
                  CASE INLIST(lcdirect, 'O', 'B') AND m.ctype = 'BBL' AND m.lflat AND m.ctypeint = 'O'
*  This owner is an oil owner and paid a flat rate
                     IF jflatcnt = 0
                        m.ninvamt  = swround(m.nflatrate * (m.nworkint / 100), 2)
                        m.ntotal   = 0
                        m.nworktot = m.nflatrate
                        jflatcnt   = 1
                     ELSE
                        m.ninvamt  = 0
                        m.nworktot = 0
                     ENDIF

                  OTHERWISE
*  The owner is not directly paid,
*  So the revenue is processed normally.
                     m.ntotal   = jgross
                     m.ninvamt  = 0
                     m.nworktot = 0
                     IF m.lflat AND jnflatcount = 0
                        m.ninvamt   = m.nflatrate
                        jnflatcount = 1
                     ELSE
                        IF 'TAX' $ m.ctype AND m.lExempt
                           m.ninvamt = 0
                        ELSE
                           m.ninvamt =  swround((jgross * (m.nrevint / 100)), 2)
                        ENDIF
                     ENDIF
               ENDCASE

               m.ntotal   = jgross             && Re-establish total

*  Check to see if this owner gets gas or oil interest
*  If not, don't show that income on his statement
               DO CASE
                  CASE m.ctype = 'BBL' AND m.nrevoil # 0
                     IF m.nworktot # 0
                        lntotinc = lntotinc + m.nworktot
                        lnoilinc = lnoilinc + m.nworktot
                     ELSE
                        lntotinc = lntotinc + m.ninvamt
                        lnoilinc = lnoilinc + m.ninvamt
                     ENDIF
                     lnwelltot = lnwelltot + m.ninvamt
                  CASE m.ctype = 'MCF' AND m.nrevgas # 0
                     IF m.nworktot # 0
                        lntotinc = lntotinc + m.nworktot
                        lngasinc = lngasinc + m.nworktot
                     ELSE
                        IF m.lflat
                           IF NOT llFlatReleased
                              lntotinc = lntotinc + m.ninvamt
                              lngasinc = lngasinc + m.ninvamt
                           ENDIF
                        ELSE
                           lntotinc = lntotinc + m.ninvamt
                           lngasinc = lngasinc + m.ninvamt
                        ENDIF
                     ENDIF

                     lnwelltot = lnwelltot + m.ninvamt
                  CASE m.ctype = 'OTH' AND m.nrevoth # 0
                     IF m.nworktot # 0
                        lntotinc = lntotinc + m.nworktot
                        lnothinc = lnothinc + m.nworktot
                     ELSE
                        lntotinc = lntotinc + m.ninvamt
                        lnothinc = lnothinc + m.ninvamt
                     ENDIF
                     lnwelltot = lnwelltot + m.ninvamt
                  CASE m.ctype = 'OTAX1' AND m.nrevtax1 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + m.ninvamt * -1
                     ENDIF
                     lnoilmantax1 = lnoilmantax1 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        IF llroyaltyowner
                           IF lcdirect = 'B' OR lcdirect = 'O'
                              lnwelltot = lnwelltot + m.ninvamt
                           ELSE
                              lnwelltot = lnwelltot + m.ninvamt
                           ENDIF
                        ELSE
                           IF lcdirect = 'B' OR lcdirect = 'O'
                              lnwelltot = lnwelltot + m.ninvamt
                           ELSE
                              lnwelltot = lnwelltot + m.ninvamt
                           ENDIF
                        ENDIF
                     ENDIF

                  CASE m.ctype = 'OTAX2' AND m.nrevtax4 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnoilmantax2 = lnoilmantax2 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'OTAX3' AND m.nrevtax7 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnoilmantax3 = lnoilmantax3 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'OTAX4' AND m.nrevtax10 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnoilmantax4 = lnoilmantax4 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'GTAX1' AND m.nrevtax2 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lngasmantax1 = lngasmantax1 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        IF llroyaltyowner
                           IF lcdirect = 'B' OR lcdirect = 'G'
                              lnwelltot = lnwelltot + m.ninvamt
                           ELSE
                              lnwelltot = lnwelltot + m.ninvamt
                           ENDIF
                        ELSE
                           IF lcdirect = 'B' OR lcdirect = 'G'
                              lnwelltot = lnwelltot + m.ninvamt
                           ELSE
                              lnwelltot = lnwelltot + m.ninvamt
                           ENDIF
                        ENDIF
                     ENDIF

                  CASE m.ctype = 'GTAX2' AND m.nrevtax5 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lngasmantax2 = lngasmantax2 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'GTAX3' AND m.nrevtax8 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lngasmantax3 = lngasmantax3 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'GTAX4' AND m.nrevtax11 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lngasmantax4 = lngasmantax4 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'PTAX1' AND m.nrevtax3 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnothmantax1 = lnothmantax1 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF
                  CASE m.ctype = 'PTAX2' AND m.nrevtax6 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnothmantax2 = lnothmantax2 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF
                  CASE m.ctype = 'PTAX3' AND m.nrevtax9 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnothmantax3 = lnothmantax3 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'PTAX4' AND m.nrevtax12 # 0
                     IF m.nworktot = 0
                        lntaxes = lntaxes + (m.ninvamt * -1)
                     ENDIF
                     lnothmantax4 = lnothmantax4 + (m.ninvamt * -1)
                     IF NOT m.lExempt
                        lnwelltot = lnwelltot + m.ninvamt
                     ENDIF

                  CASE m.ctype = 'TRANS' AND m.nrevtrp # 0
                     lntrpinc  = lntrpinc + m.ninvamt
                     lntotinc  = lntotinc + m.ninvamt
                     lnwelltot = lnwelltot + m.ninvamt

                  CASE m.ctype = 'MISC1' AND m.nrevmisc1 # 0
                     lnmi1inc  = lnmi1inc + m.ninvamt
                     lntotinc  = lntotinc + m.ninvamt
                     lnwelltot = lnwelltot + m.ninvamt

                  CASE m.ctype = 'MISC2' AND m.nrevmisc2 # 0
                     lnmi2inc  = lnmi2inc + m.ninvamt
                     lntotinc  = lntotinc + m.ninvamt
                     lnwelltot = lnwelltot + m.ninvamt

                  CASE m.lflat
                     lnwelltot = lnwelltot + m.ninvamt
               ENDCASE
               STORE 0 TO m.ninvamt, m.ntotal, m.nprice, m.nUnits, m.nrevint, jgross, ;
                  m.ngasrev, m.noilrev, m.nothrev, m.ntrprev, m.nworktot, m.nmiscrev1, m.nmiscrev2
               STORE ' ' TO m.ctype, m.cdescript, m.cmiscmemo
               m.nrevoil   = jnrevoil
               m.nrevgas   = jnrevgas
               m.nrevoth   = jnrevoth
               m.nrevtax1  = jnrevtax1
               m.nrevtax2  = jnrevtax2
               m.nrevtax3  = jnrevtax3
               m.nrevtax4  = jnrevtax4
               m.nrevtax5  = jnrevtax5
               m.nrevtax6  = jnrevtax6
               m.nrevtax7  = jnrevtax7
               m.nrevtax8  = jnrevtax8
               m.nrevtax9  = jnrevtax9
               m.nrevtax10 = jnrevtax10
               m.nrevtax11 = jnrevtax11
               m.nrevtax12 = jnrevtax12
               m.nrevmisc1 = jnrevm1
               m.nrevmisc2 = jnrevm2
               m.nrevtrp   = jnrevtrans
            ENDSCAN && Inctemp

            STORE ' ' TO m.ctype, m.cSource
            IF NOT llincome AND m.lflat
               lnwelltot = lnwelltot + m.nflatrate
               llincome  = .F.
            ENDIF

**********************************************************
*  Process flat-rate royalties
**********************************************************
            IF (lnflatgasamt # 0 OR lnflatoilamt # 0) AND NOT llroyaltyowner AND NOT llFlatAllocated
* Process Flat Gas
               m.nrevint = m.nworkint
               jwrk      = lnflatgasamt * (m.nworkint / 100)
               m.ninvamt = jwrk * -1
               m.cdirect = lcdirect
               lngasinc  = lngasinc + m.ninvamt
               lntotinc  = lntotinc + m.ninvamt
               lnwelltot = lnwelltot + m.ninvamt

* Process Flat Oil
               jwrk            = lnflatoilamt * (m.nworkint / 100)
               m.ninvamt       = jwrk * -1
               m.cdirect       = lcdirect
               lnoilinc        = lnoilinc + m.ninvamt
               lntotinc        = lntotinc + m.ninvamt
               lnwelltot       = lnwelltot + m.ninvamt
               llFlatAllocated = .T.
            ENDIF
            STORE 0 TO lnflatoilamt, lnflatgasamt
            STORE ' ' TO m.ctype
**********************************************************
*  Process taxes
**********************************************************
*
*  We check the lnTax variables to see if they are zero.
*  If they are, then no taxes were entered manually for
*  that product and we can calculate them here.  If the
*  variable is not zero, then it means the tax was entered
*  manually and we don't want to calculate it here for that
*  product.
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK (m.cWellID)
               SCATTER MEMVAR
               llRoyComp = lExclRoyComp
            ELSE
               LOOP
            ENDIF
            IF m.nroyint # 0
               m.nrevotax = (m.nrevoil / m.nroyint) * 100
               m.nrevgtax = (m.nrevgas / m.nroyint) * 100
            ELSE
               m.nrevotax = 0
               m.nrevgtax = 0
            ENDIF

            SELE temptax
            LOCATE FOR cWellID == m.cWellID AND hyear + hperiod = invtmp.hyear + invtmp.hperiod AND cdeck = invtmp.cdeck
            IF NOT FOUND()
               STORE 0 TO m.noiltax1, m.ngastax1, m.noiltax2, m.ngastax2, m.noiltax3, m.ngastax3, m.noiltax4, m.ngastax4, m.nprodtax1, m.nprodtax2, m.nprodtax3, m.nprodtax4
            ENDIF

            SELE temptax1
            LOCATE FOR cWellID == m.cWellID AND hyear + hperiod = invtmp.hyear + invtmp.hperiod AND cdeck = invtmp.cdeck
            IF NOT FOUND()
               STORE 0 TO m.noiltax1a, m.ngastax1a, m.noiltax2a, m.ngastax2a, m.noiltax3a, m.ngastax3a, m.noiltax4a, m.ngastax4a
               STORE 0 TO m.nprodtax1a, m.nprodtax2a, m.nprodtax3a, m.nprodtax4a
            ENDIF

            STORE 0 TO m.ncoiltax1, m.ncoiltax2, m.ncoiltax3, m.ncoiltax4
            STORE 0 TO m.ncgastax1, m.ncgastax2, m.ncgastax3, m.ncgastax4
            STORE 0 TO m.ncprodtax1, m.ncprodtax2, m.ncprodtax3, m.ncprodtax4

            SELECT taxcalc
            LOCATE FOR hyear + hperiod + cWellID = invtmp.hyear + invtmp.hperiod + m.cWellID AND cdeck = invtmp.cdeck
            IF FOUND()
               SCATTER MEMVAR
            ENDIF
            swselect('sevtax')
            SET ORDER TO ctable
            IF (NOT EMPTY(m.ctable) AND SEEK(m.ctable))
               SCATTER MEMVAR
               IF m.lExempt             && Owner is Tax Exempt
                  m.ninvamt = 0
                  STORE 0 TO lnoiltax1, lnoiltax2, lnoiltax3, lnoiltax4
                  STORE 0 TO lngastax1, lngastax2, lngastax3, lngastax4
                  STORE 0 TO lnothtax1, lnothtax2, lnothtax3, lnothtax4
               ELSE
                  IF INLIST(m.ctypeint, 'B', 'O')     && Oil Interest
                     IF lnoiltax1 = 0     && Oil Tax 1 has not been entered by user
                        IF m.lusesev  && Use well tax rates
                           IF llroyaltyowner    && Owner is a royalty owner
                              lnoiltax1 = swround(((lnoilrevenue * (m.nroysevo / 100)) * (m.nrevoil / 100)), 2)
                              lntaxes   = lntaxes + lnoiltax1
                           ELSE
                              lnoiltax1 = swround(((lnoilrevenue * (m.nwrksevo / 100)) * (m.nrevoil / 100)), 2)
                              lntaxes   = lntaxes + lnoiltax1
                           ENDIF
                        ELSE
                           DO CASE
                              CASE m.ltaxexempt1
* Well is exempt from tax 1
                                 lnOilTax = 0
                              CASE m.lsev1o OR (m.lDirOilPurch AND INLIST(lcdirect, 'O', 'B'))
*  If the purchaser pays the tax, we don't calculate it here
                                 lnoiltax1 = 0
                              OTHERWISE
                                 IF NOT llroyaltyowner AND llroysevtx
                                    lnPct = m.nworkint
* The royalty owner is excluded from severance taxes
                                    lnoiltax1   = (swround((m.ntotbbltx1 - temptax.noiltax1 - temptax1.noiltax1a) * (lnPct / 100), 2))
                                 ELSE
                                    IF llroyaltyowner AND llroysevtx
                                       lnoiltax1 = 0
                                    ELSE
                                       lnPct     = m.nrevtax1
*                                                lnoiltax1 = (swround((m.ntotbbltx1 - temptax.noiltax1 - temptax1.noiltax1a) * (lnPct / 100), 2))
                                       lnoiltax1 = lnoiltax1 + (swround(m.ncoiltax1 * (lnPct / 100), 2))
                                    ENDIF
                                 ENDIF
                           ENDCASE
                           lntaxes = lntaxes + lnoiltax1
                        ENDIF
                        IF lnoiltax1 # 0 AND NOT m.lsev1o
                           lnwelltot = lnwelltot - lnoiltax1
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                     IF lnoiltax2 = 0     && Oil Tax 2 has not been entered by user
                        IF m.lsev2o OR (m.lDirOilPurch AND INLIST(lcdirect, 'O', 'B')) OR m.ltaxexempt2
                           lnoiltax2 = 0
                        ELSE
                           IF NOT llroyaltyowner AND llroysevtx
                              lnoiltax2   = (swround((m.ntotbbltx2 - temptax.noiltax2 - temptax1.noiltax2a) * (m.nworkint / 100), 2))
                           ELSE
                              IF llroyaltyowner AND llroysevtx
                                 lnoiltax2 = 0
                              ELSE
*                                        lnoiltax2   = (swround((m.ntotbbltx2 - temptax.noiltax2 - temptax1.noiltax2a) * (m.nrevtax4 / 100), 2))
                                 lnoiltax2 = lnoiltax2 + (swround(m.ncoiltax2 * (m.nrevtax4 / 100), 2))
                              ENDIF
                           ENDIF
                           lntaxes = lntaxes + lnoiltax2
                        ENDIF
                        IF lnoiltax2  # 0 AND NOT m.lsev2o
                           lnwelltot = lnwelltot - lnoiltax2
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                     IF lnoiltax3 = 0     && Oil Tax 3 has not been entered by user
                        IF m.lsev3o OR (m.lDirOilPurch AND INLIST(lcdirect, 'O', 'B')) OR m.ltaxexempt3
                           lnoiltax3 = 0
                        ELSE
                           IF NOT llroyaltyowner AND llroysevtx
                              lnoiltax3   = (swround((m.ntotbbltx3 - temptax.noiltax3 - temptax1.noiltax3a) * (m.nworkint / 100), 2))
                           ELSE
                              IF llroyaltyowner AND llroysevtx
                                 lnoiltax3 = 0
                              ELSE
*                                        lnoiltax3   = (swround((m.ntotbbltx3 - temptax.noiltax3 - temptax1.noiltax3a) * (m.nrevtax7 / 100), 2))
                                 lnoiltax3 = lnoiltax3 + (swround(m.ncoiltax3 * (m.nrevtax7 / 100), 2))
                              ENDIF
                           ENDIF
                           lntaxes = lntaxes + lnoiltax3
                        ENDIF
                        IF lnoiltax3  # 0 AND NOT m.lsev3o
                           lnwelltot = lnwelltot - lnoiltax3
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                     IF lnoiltax4 = 0     && Oil Tax 4 has not been entered by user
                        IF m.lsev4o OR (m.lDirOilPurch AND INLIST(lcdirect, 'O', 'B')) OR m.ltaxexempt4
                           lnoiltax4 = 0
                        ELSE
                           IF NOT llroyaltyowner AND llroysevtx
                              lnoiltax4   = (swround((m.ntotbbltx4 - temptax.noiltax4 - temptax1.noiltax4a) * (m.nworkint / 100), 2))
                           ELSE
                              IF llroyaltyowner AND llroysevtx
                                 lnoiltax4 = 0
                              ELSE
*                                        lnoiltax4   = (swround((m.ntotbbltx4 - temptax.noiltax4 - temptax1.noiltax4a) * (m.nrevtax10 / 100), 2))
                                 lnoiltax4 = lnoiltax4 + (swround(m.ncoiltax4 * (m.nrevtax10 / 100), 2))
                              ENDIF
                           ENDIF
                           lntaxes = lntaxes + lnoiltax4
                        ENDIF
                        IF lnoiltax4  # 0 AND NOT m.lsev4o
                           lnwelltot = lnwelltot - lnoiltax4
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                  ENDIF
*****************************************************************************
*  Calculate gas taxes
*****************************************************************************

                  IF INLIST(m.ctypeint, 'B', 'G')
                     IF lngastax1 = 0      && User did not enter gas tax 1
                        IF m.lusesev  && Use well tax rates
                           IF llroyaltyowner    && Owner is a royalty owner
                              lngastax1 = swround(((lngasrevenue * (m.nroysevg / 100)) * (m.nrevgas / 100)), 2)
                              lntaxes   = lntaxes + lngastax1
                           ELSE
                              lngastax1 = swround(((lngasrevenue * (m.nwrksevg / 100)) * (m.nrevgas / 100)), 2)
                              lntaxes   = lntaxes + lngastax1
                           ENDIF
                        ELSE
*  If the purchaser pays the tax, we don't calculate it here
                           IF m.lsev1g OR (m.lDirGasPurch AND INLIST(lcdirect, 'G', 'B')) OR m.ltaxexempt1
                              lngastax1 = 0
                           ELSE
                              IF NOT llroyaltyowner AND llroysevtx
                                 lnPct     = m.nworkint
                                 lngastax1 = (swround((m.ntotmcftx1 - temptax.ngastax1 - temptax1.ngastax1a) * (m.nworkint / 100), 2))
                              ELSE
                                 IF llroyaltyowner AND llroysevtx
                                    lngastax1 = 0
                                 ELSE
                                    lnPct     = m.nrevtax2
*                                            lngastax1 = lngastax1 + (swround((m.ntotmcftx1 - temptax.ngastax1 - temptax1.ngastax1a) * (lnPct / 100), 2))
                                    lngastax1 = lngastax1 + (swround(m.ncgastax1 * (lnPct / 100), 2))
                                 ENDIF
                              ENDIF
                              lntaxes = lntaxes + lngastax1
                           ENDIF
                        ENDIF
                        IF lngastax1 # 0 AND NOT m.lsev1g
                           lnwelltot = lnwelltot - lngastax1
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                     IF lngastax2 = 0      && User did not enter gas tax 2
                        IF m.lsev2g OR (m.lDirGasPurch AND INLIST(lcdirect, 'G', 'B')) OR m.ltaxexempt2
                           lngastax2 = 0
                        ELSE
                           IF NOT llroyaltyowner AND llroysevtx
                              lnPct     = m.nworkint
                              lngastax2 = (swround((m.ntotmcftx2 - temptax.ngastax2 - temptax1.ngastax2a) * (m.nworkint / 100), 2))
                           ELSE
                              IF llroyaltyowner AND llroysevtx
                                 lngastax2 = 0
                              ELSE
*                                        lngastax2   = (swround((m.ntotmcftx2 - temptax.ngastax2 - temptax1.ngastax2a) * (m.nrevtax5 / 100), 2))
                                 lngastax2 = lngastax2 + (swround(m.ncgastax2 * (m.nrevtax5 / 100), 2))
                              ENDIF
                           ENDIF
                           lntaxes = lntaxes + lngastax2
                        ENDIF
                        IF lngastax2 # 0 AND NOT m.lsev2g
                           lnwelltot = lnwelltot - lngastax2
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                     IF lngastax3 = 0      && User did not enter gas tax 3
                        IF m.lsev3g OR (m.lDirGasPurch AND INLIST(lcdirect, 'G', 'B')) OR m.ltaxexempt3
                           lngastax3 = 0
                        ELSE
                           IF NOT llroyaltyowner AND llroysevtx
                              lngastax3   = (swround((m.ntotmcftx3 - temptax.ngastax3 - temptax1.ngastax3a) * (m.nworkint / 100), 2))
                           ELSE
                              IF llroyaltyowner AND llroysevtx
                                 lngastax3 = 0
                              ELSE
*                                        lngastax3   = (swround((m.ntotmcftx3 - temptax.ngastax3 - temptax1.ngastax3a) * (m.nrevtax8 / 100), 2))
                                 lngastax3 = lngastax3 + (swround(m.ncgastax3 * (m.nrevtax8 / 100), 2))
                              ENDIF
                           ENDIF
                           lntaxes = lntaxes + lngastax3
                        ENDIF
                        IF lngastax3 # 0 AND NOT m.lsev3g
                           lnwelltot = lnwelltot - lngastax3
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                     IF lngastax4 = 0      && User did not enter gas tax 4
                        IF m.lsev4g OR (m.lDirGasPurch AND INLIST(lcdirect, 'G', 'B')) OR m.ltaxexempt4
                           lngastax4 = 0
                        ELSE
                           IF NOT llroyaltyowner AND llroysevtx
                              lngastax4   = (swround((m.ntotmcftx4 - temptax.ngastax4 - temptax1.ngastax4a) * (m.nworkint / 100), 2))
                           ELSE
                              IF llroyaltyowner AND llroysevtx
                                 lngastax4 = 0
                              ELSE
*                                        lngastax4   = (swround((m.ntotmcftx4 - temptax.ngastax4 - temptax1.ngastax4a) * (m.nrevtax11 / 100), 2))
                                 lngastax4 = lngastax4 + (swround(m.ncgastax4 * (m.nrevtax11 / 100), 2))
                              ENDIF
                           ENDIF
                           lntaxes = lntaxes + lngastax4
                        ENDIF
                        IF lngastax4 # 0 AND NOT m.lsev4g
                           lnwelltot = lnwelltot - lngastax4
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                  ENDIF
*****************************************************************************
*  Calculate other product taxes
*****************************************************************************
                  IF lnothtax1 = 0
                     IF m.lsev1p OR m.ltaxexempt1
                        lnothtax1 = 0
                     ELSE
                        IF NOT llroyaltyowner AND llroysevtx
                           lnothtax1   = (swround((m.ntotothtx1 - temptax.nprodtax1 - temptax1.nprodtax1a) * (m.nworkint / 100), 2))
                        ELSE
                           IF llroyaltyowner AND llroysevtx
                              lnothtax1 = 0
                           ELSE
                              lnothtax1 = lnothtax1 + (swround(m.ncprodtax1 * (m.nrevtax3 / 100), 2))
                           ENDIF
                        ENDIF
                        lntaxes = lntaxes + lnothtax1
                     ENDIF
                     IF lnothtax1 # 0 AND NOT m.lsev1p
                        lnwelltot = lnwelltot - lnothtax1
                        m.ninvamt = 0
                     ENDIF
                  ENDIF
                  IF lnothtax2 = 0
                     IF m.lsev2p OR m.ltaxexempt2
                        lnothtax2 = 0
                     ELSE
                        IF NOT llroyaltyowner AND llroysevtx
                           lnothtax2   = (swround((m.ntotothtx2 - temptax.nprodtax2 - temptax1.nprodtax2a) * (m.nworkint / 100), 2))
                        ELSE
                           IF llroyaltyowner AND llroysevtx
                              lnothtax2 = 0
                           ELSE
*                                    lnothtax2   = (swround((m.ntotothtx2 - temptax.nprodtax2 - temptax1.nprodtax2a) * (m.nrevtax6 / 100), 2))
                              lnothtax2 = lnothtax2 + (swround(m.ncprodtax2 * (m.nrevtax6 / 100), 2))
                           ENDIF
                        ENDIF
                        lntaxes = lntaxes + lnothtax2
                     ENDIF
                     IF lnothtax2 # 0 AND NOT m.lsev2p
                        lnwelltot = lnwelltot - lnothtax2
                        m.ninvamt = 0
                     ENDIF
                  ENDIF
                  IF lnothtax3 = 0
                     IF m.lsev3p OR m.ltaxexempt3
                        lnothtax3 = 0
                     ELSE
                        IF NOT llroyaltyowner AND llroysevtx
                           lnothtax3   = (swround((m.ntotothtx3 - temptax.nprodtax3 - temptax1.nprodtax3a) * (m.nworkint / 100), 2))
                        ELSE
                           IF llroyaltyowner AND llroysevtx
                              lnothtax3 = 0
                           ELSE
*                                    lnothtax3   = (swround((m.ntotothtx3 - temptax.nprodtax3 - temptax1.nprodtax3a) * (m.nrevtax9 / 100), 2))
                              lnothtax3 = lnothtax3 + (swround(m.ncprodtax3 * (m.nrevtax9 / 100), 2))
                           ENDIF
                        ENDIF
                        lntaxes = lntaxes + lnothtax3
                     ENDIF
                     IF lnothtax3 # 0 AND NOT m.lsev3p
                        lnwelltot = lnwelltot - lnothtax3
                        m.ninvamt = 0
                     ENDIF
                  ENDIF
                  IF lnothtax4 = 0
                     IF m.lsev4p OR m.ltaxexempt4
                        lnothtax4 = 0
                     ELSE
                        IF NOT llroyaltyowner AND llroysevtx
                           lnothtax4   = (swround((m.ntotothtx4 - temptax.nprodtax4 - temptax1.nprodtax4a) * (m.nworkint / 100), 2))
                        ELSE
                           IF llroyaltyowner AND llroysevtx
                              lnothtax4 = 0
                           ELSE
*                                    lnothtax4   = (swround((m.ntotothtx4 - temptax.nprodtax4 - temptax1.nprodtax4a) * (m.nrevtax12 / 100), 2))
                              lnothtax4 = lnothtax4 + (swround(m.ncprodtax4 * (m.nrevtax12 / 100), 2))
                           ENDIF
                        ENDIF
                        lntaxes = lntaxes + lnothtax4
                     ENDIF
                     IF lnothtax4 # 0 AND NOT m.lsev4p
                        lnwelltot = lnwelltot - lnothtax4
                        m.ninvamt = 0
                     ENDIF
                  ENDIF
               ENDIF
*
* Look for one man item taxes and apply to this owner
*
               IF NOT m.lExempt
                  swselect('one_man_tax')
                  LOCATE FOR cWellID == m.cWellID AND cownerid == jcownerid ;
                     AND hyear + hperiod = invtmp.hyear + invtmp.hperiod ;
                     AND crunyear == THIS.crunyear AND nrunno == THIS.nrunno
                  IF FOUND()
                     lnoiltax1 = lnoiltax1 + one_man_tax.noiltax1b
                     lnoiltax2 = lnoiltax2 + one_man_tax.noiltax2b
                     lnoiltax3 = lnoiltax3 + one_man_tax.noiltax3b
                     lnoiltax4 = lnoiltax4 + one_man_tax.noiltax4b
                     lngastax1 = lngastax1 + one_man_tax.ngastax1b
                     lngastax2 = lngastax2 + one_man_tax.ngastax2b
                     lngastax3 = lngastax3 + one_man_tax.ngastax3b
                     lngastax4 = lngastax4 + one_man_tax.ngastax4b
                     lnothtax1 = lnothtax1 + one_man_tax.nprodtax1b
                     lnothtax2 = lnothtax2 + one_man_tax.nprodtax2b
                     lnothtax3 = lnothtax3 + one_man_tax.nprodtax3b
                     lnothtax4 = lnothtax4 + one_man_tax.nprodtax4b
                     lntaxes   = lntaxes   + one_man_tax.noiltax1b + ;
                        one_man_tax.noiltax2b + ;
                        one_man_tax.noiltax3b + ;
                        one_man_tax.noiltax4b + ;
                        one_man_tax.ngastax1b + ;
                        one_man_tax.ngastax2b + ;
                        one_man_tax.ngastax3b + ;
                        one_man_tax.ngastax4b + ;
                        one_man_tax.nprodtax1b + ;
                        one_man_tax.nprodtax2b + ;
                        one_man_tax.nprodtax3b + ;
                        one_man_tax.nprodtax4b
                     lnwelltot = lnwelltot - one_man_tax.noiltax1b - ;
                        one_man_tax.noiltax2b - ;
                        one_man_tax.noiltax3b - ;
                        one_man_tax.noiltax4b - ;
                        one_man_tax.ngastax1b - ;
                        one_man_tax.ngastax2b - ;
                        one_man_tax.ngastax3b - ;
                        one_man_tax.ngastax4b - ;
                        one_man_tax.nprodtax1b - ;
                        one_man_tax.nprodtax2b - ;
                        one_man_tax.nprodtax3b - ;
                        one_man_tax.nprodtax4b
                  ENDIF
               ENDIF
            ELSE
               IF m.lExempt             && Owner is Tax Exempt
                  m.ninvamt = 0
                  STORE 0 TO lnoiltax1, lnoiltax2, lnoiltax3, lnoiltax4
                  STORE 0 TO lngastax1, lngastax2, lngastax3, lngastax4
                  STORE 0 TO lnothtax1, lnothtax2, lnothtax3, lnothtax4
               ELSE
                  IF INLIST(m.ctypeint, 'B', 'O')     && Oil Interest
                     IF lnoiltax1 = 0     && Oil Tax 1 has not been entered by user
                        IF m.lusesev  && Use well tax rates
                           IF NOT INLIST(lcdirect, 'O', 'B')
                              IF llroyaltyowner    && Owner is a royalty owner
                                 lnoiltax1 = swround(((lnoilrevenue * (m.nroysevo / 100)) * (m.nrevoil / 100)), 2)
                                 lntaxes   = lntaxes + lnoiltax1
                              ELSE
                                 lnoiltax1 = swround(((lnoilrevenue * (m.nwrksevo / 100)) * (m.nrevoil / 100)), 2)
                                 lntaxes   = lntaxes + lnoiltax1
                              ENDIF
                           ENDIF
                        ENDIF
                        IF lnoiltax1 # 0 AND NOT m.lsev1o
                           lnwelltot = lnwelltot - lnoiltax1
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                  ENDIF
*****************************************************************************
*  Calculate gas taxes
*****************************************************************************
                  IF INLIST(m.ctypeint, 'B', 'G')
                     IF lngastax1 = 0      && User did not enter gas tax 1
                        IF m.lusesev  && Use well tax rates
                           IF llroyaltyowner    && Owner is a royalty owner
                              lngastax1 = swround(((lngasrevenue * (m.nroysevg / 100)) * (m.nrevgas / 100)), 2)
                              lntaxes   = lntaxes + lngastax1
                           ELSE
                              lngastax1 = swround(((lngasrevenue * (m.nwrksevg / 100)) * (m.nrevgas / 100)), 2)
                              lntaxes   = lntaxes + lngastax1
                           ENDIF
                        ENDIF
                        IF lngastax1 # 0 AND NOT m.lsev1g
                           lnwelltot = lnwelltot - lngastax1
                           m.ninvamt = 0
                        ENDIF
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF

            STORE 0 TO m.ntotal, m.ninvamt, m.nrevint
            STORE ' ' TO m.cSource, m.ctype

**********************************************************
*  Process compression and gathering charges
**********************************************************
            swselect('expcat')
            SET ORDER TO ccatcode

            SELECT wells
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               llGather   = lGather
               llCompress = lcompress
               STORE 0 TO lnExpGather, lnExpGatherOneMan
               m.ninvamt = 0
               IF INLIST(m.ctypeint, 'B', 'G')
                  lnExpGather = 0
                  IF USED('gathtemp')
                     SELECT gathtemp
                     SCAN FOR cWellID == m.cWellID AND EMPTY(cownerid) AND cyear == m.hyear AND cperiod == m.hperiod AND cdeck = invtmp.cdeck
                        lnExpGather = lnExpGather + namount
                     ENDSCAN
                     SCAN FOR cWellID == m.cWellID AND cownerid == jcownerid AND cyear == m.hyear AND cperiod == m.hperiod AND cdeck = invtmp.cdeck
                        lnExpGatherOneMan = lnExpGatherOneMan + namount
                     ENDSCAN
                  ENDIF

                  SELECT expcat
                  IF SEEK('GATH')
                     DO CASE
                        CASE expcat.cexpclass = '0'
                           lnInterest = 'm.nWorkint'
                        CASE expcat.cexpclass = '1'
                           lnInterest = 'm.nIntClass1'
                        CASE expcat.cexpclass = '2'
                           lnInterest = 'm.nIntClass2'
                        CASE expcat.cexpclass = '3'
                           lnInterest = 'm.nIntClass3'
                        CASE expcat.cexpclass = '4'
                           lnInterest = 'm.nIntClass4'
                        CASE expcat.cexpclass = '5'
                           lnInterest = 'm.nIntClass5'
                        CASE expcat.cexpclass = 'A'
                           lnInterest = 'm.nIntClassa'
                        CASE expcat.cexpclass = 'B'
                           lnInterest = 'm.nIntClassB'
                        CASE expcat.cexpclass = 'G'
                           lnInterest = 'm.nRevGas'
                        OTHERWISE
                           lnInterest = 'm.nWorkint'
                     ENDCASE
                  ELSE
                     lnInterest = 'm.nRevGas'
                  ENDIF


                  m.ninvamt = m.ninvamt + swround(lnExpGather * (&lnInterest / 100), 2)
                  m.ninvamt = m.ninvamt + lnExpGatherOneMan

                  lnRevGather = 0
                  IF USED('gathrev')
                     SELECT gathrev
                     SCAN FOR cWellID == m.cWellID
                        lnRevGather = lnRevGather + nTotalInc
                     ENDSCAN
                  ENDIF

                  IF (llroyaltyowner AND NOT llRoyComp) OR NOT llroyaltyowner
                     IF NOT llroyaltyowner AND llRoyComp
                        lnInterest = 'm.nworkint'
                     ELSE
                        lnInterest = 'm.nrevgas'
                     ENDIF
                  ELSE
                     lcInterest = 0
                     lnInterest = 'lcInterest'
                  ENDIF

                  m.ninvamt = m.ninvamt - swround(lnRevGather * (&lnInterest / 100), 2)

                  lnCalcGather = 0
                  SELECT gathcalc
                  SCAN FOR cWellID == m.cWellID AND cyear = m.hyear AND cperiod = m.hperiod AND cdeck = invtmp.cdeck
                     lnCalcGather = lnCalcGather + namount
                  ENDSCAN

                  m.ninvamt = m.ninvamt + swround(lnCalcGather * (&lnInterest / 100), 2)

                  lnGather  = lnGather + m.ninvamt

                  m.nrevint   = m.nrevgas
                  IF m.ninvamt # 0
                     lnwelltot = lnwelltot - m.ninvamt
                  ENDIF

                  STORE 0 TO lnExpCompress, lnExpCompressOneMan
                  IF USED('comptemp')
                     SELECT comptemp
                     SCAN FOR cWellID == m.cWellID AND EMPTY(cownerid) AND cyear == m.hyear AND cperiod == m.hperiod AND cdeck = invtmp.cdeck
                        lnExpCompress = lnExpCompress + namount
                     ENDSCAN
                     SCAN FOR cWellID == m.cWellID AND cownerid == jcownerid AND cyear == m.hyear AND cperiod == m.hperiod AND cdeck = invtmp.cdeck
                        lnExpCompressOneMan = lnExpCompressOneMan + namount
                     ENDSCAN
                  ENDIF

                  SELECT expcat
                  IF SEEK('COMP')
                     DO CASE
                        CASE expcat.cexpclass = '0'
                           lnInterest = 'm.nWorkint'
                        CASE expcat.cexpclass = '1'
                           lnInterest = 'm.nIntClass1'
                        CASE expcat.cexpclass = '2'
                           lnInterest = 'm.nIntClass2'
                        CASE expcat.cexpclass = '3'
                           lnInterest = 'm.nIntClass3'
                        CASE expcat.cexpclass = '4'
                           lnInterest = 'm.nIntClass4'
                        CASE expcat.cexpclass = '5'
                           lnInterest = 'm.nIntClass5'
                        CASE expcat.cexpclass = 'A'
                           lnInterest = 'm.nIntClassa'
                        CASE expcat.cexpclass = 'B'
                           lnInterest = 'm.nIntClassB'
                        CASE expcat.cexpclass = 'G'
                           lnInterest = 'm.nRevGas'
                        OTHERWISE
                           lnInterest = 'm.nWorkint'
                     ENDCASE
                  ELSE
                     lnInterest = 'm.nrevGas'
                  ENDIF

                  m.ninvamt = swround(lnExpCompress * (&lnInterest / 100), 2)
                  m.ninvamt = m.ninvamt + lnExpCompressOneMan

                  lnRevCompress = 0
                  SELECT comprev
                  SCAN FOR cWellID == m.cWellID
                     lnRevCompress = lnRevCompress + nTotalInc
                  ENDSCAN

                  IF (llroyaltyowner AND NOT llRoyComp) OR NOT llroyaltyowner
                     IF NOT llroyaltyowner AND llRoyComp
                        lnInterest = 'm.nworkint'
                     ELSE
                        lnInterest = 'm.nrevgas'
                     ENDIF
                  ELSE
                     lcInterest = 0
                     lnInterest = 'lcInterest'
                  ENDIF

                  m.ninvamt = m.ninvamt - swround(lnRevCompress * (&lnInterest / 100), 2)

                  lnCalcCompress = 0
                  SELECT compcalc
                  SCAN FOR cWellID == m.cWellID AND cyear = m.hyear AND cperiod = m.hperiod
                     lnCalcCompress = lnCalcCompress + namount
                  ENDSCAN

                  m.ninvamt = m.ninvamt + swround(lnCalcCompress * (&lnInterest / 100), 2)

                  lnCompress  = lnCompress + m.ninvamt

                  m.nrevint   = m.nrevgas
                  IF m.ninvamt # 0
                     lnwelltot = lnwelltot - m.ninvamt
                  ENDIF
               ENDIF
            ENDIF
            STORE 0 TO m.ntotal, m.ninvamt, m.nrevint
            STORE ' ' TO m.cSource, m.ctype

**********************************************************
*  Process expenses
**********************************************************
            STORE 0 TO m.ntotale1, m.ntotale2, m.ntotale3, m.ntotale4, m.ntotale5, m.ntotalea, m.ntotaleb, m.nPlugExp

            IF NOT m.lJIB
               SELECT exptemp
               SCAN FOR cWellID = m.cWellID AND cyear + cperiod = m.hyear + m.hperiod AND cdeck == lcDeck
                  SCATTER MEMVAR

*  Don't process marketing expenses here.
                  IF INLIST(m.ccatcode, 'MKTG', 'COMP', 'GATH')
                     LOOP
                  ENDIF

* Don't process plugging expenses if the plugging module isn't active
                  IF NOT m.goApp.lPluggingModule AND m.cexpclass = 'P'
                     LOOP
                  ENDIF

                  SELECT exptemp
                  IF llroyaltyowner AND m.cexpclass = '0' AND EMPTY(cownerid)
* Royalty owners don't get working interest expenses (unless it's a one-man item)....
                     LOOP
                  ENDIF
                  m.category = m.ccateg
*  Store original interests so we can change for one-man-items.  pws 3/11/97
                  jnworkint = m.nworkint
                  jnclass1  = m.nintclass1
                  jnclass2  = m.nintclass2
                  jnclass3  = m.nintclass3
                  jnclass4  = m.nintclass4
                  jnclass5  = m.nintclass5
                  jnclassa  = m.nacpint
                  jnclassb  = m.nbcpint
                  jnPlugPct = m.nplugpct

                  DO CASE
                     CASE m.cownerid = jcownerid AND NOT m.loneman
* If this is a royalty owner and they also have a working interest,
* don't process one man item for royalty interest.  pws 7/27/05
                        IF m.ctypeinv # 'W'
                           SELE wellinv
                           LOCATE FOR cWellID == m.cWellID AND cownerid == jcownerid AND ctypeinv = 'W' AND cdeck = invtmp.cdeck
                           IF FOUND()
                              LOOP
                           ENDIF
                        ENDIF
                        SELECT exptemp
                        REPLACE loneman WITH .T.
                        DO CASE
                           CASE m.cexpclass = '0'
                              m.nworkint = 100
                           CASE m.cexpclass = '1'
                              m.nintclass1 = 100
                           CASE m.cexpclass = '2'
                              m.nintclass2 = 100
                           CASE m.cexpclass = '3'
                              m.nintclass3 = 100
                           CASE m.cexpclass = '4'
                              m.nintclass4 = 100
                           CASE m.cexpclass = '5'
                              m.nintclass5 = 100
                           CASE m.cexpclass = 'A'
                              m.nacpint = 100
                           CASE m.cexpclass = 'B'
                              m.nbcpint = 100
                           CASE m.cexpclass = 'P'
                              m.nplugpct = 100
                        ENDCASE
                     CASE m.cownerid = ' '
* Don't do anything if the cOwnerID is blank
                     CASE m.cownerid # jcownerid
                        LOOP
                     CASE m.cownerid = jcownerid AND loneman
                        LOOP
                  ENDCASE
                  DO CASE
                     CASE m.cexpclass = '0'
                        m.ninvamt = swround(m.namount * (m.nworkint / 100), 2)
                        lntotexp  = lntotexp + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = '1' AND m.nexpcl1 # 0
                        m.ninvamt  = swround((m.namount * (m.nintclass1 / 100)), 2)
                        m.ntotale1 = m.ntotale1 + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = '2' AND m.nexpcl2 # 0
                        m.ninvamt  = swround((m.namount * (m.nintclass2 / 100)), 2)
                        m.ntotale2 = m.ntotale2 + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = '3' AND m.nexpcl3 # 0
                        m.ninvamt  = swround((m.namount * (m.nintclass3 / 100)), 2)
                        m.ntotale3 = m.ntotale3 + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = '4' AND m.nexpcl4 # 0
                        m.ninvamt  = swround((m.namount * (m.nintclass4 / 100)), 2)
                        m.ntotale4 = m.ntotale4 + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = '5' AND m.nexpcl5 # 0
                        m.ninvamt  = swround((m.namount * (m.nintclass5 / 100)), 2)
                        m.ntotale5 = m.ntotale5 + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = 'A' AND m.nexpclA # 0
                        m.ninvamt  = swround((m.namount * (m.nacpint / 100)), 2)
                        m.ntotalea = m.ntotalea + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = 'B' AND m.nexpclB # 0
                        m.ninvamt  = swround((m.namount * (m.nbcpint / 100)), 2)
                        m.ntotaleb = m.ntotaleb + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt

                     CASE m.cexpclass = 'P' AND m.nPlugAmt # 0
                        m.ninvamt  = swround((m.namount * (m.nplugpct / 100)), 2)
                        m.nPlugExp = m.nPlugExp + m.ninvamt
                        IF NOT m.lJIB
                           lnwelltot  = lnwelltot - m.ninvamt
                        ENDIF
                        STORE 0 TO m.ninvamt
                  ENDCASE
*  Restore original interests in case one-man-item logic changed them. pws 3/11/97
                  m.nworkint   = jnworkint
                  m.nintclass5 = jnclass5
                  m.nintclass4 = jnclass4
                  m.nintclass3 = jnclass3
                  m.nintclass2 = jnclass2
                  m.nintclass1 = jnclass1
                  m.nacpint    = jnclassa
                  m.nbcpint    = jnclassb
                  m.nplugpct   = jnPlugPct
               ENDSCAN  && exptemp
            ENDIF
            WAIT CLEAR

*************************************************************
*  Process marketing expenses
*************************************************************
            STORE 0 TO lnMKTGExp
            SELECT mktgtemp
            SCAN FOR cWellID = m.cWellID AND cyear + cperiod = m.hyear + m.hperiod AND cdeck = invtmp.cdeck
               SCATTER MEMVAR
               DO CASE
                  CASE m.cexpclass = '0'
                     m.ninvamt   = swround(m.namount * (m.nrevgas / 100), 2)
                  CASE m.cexpclass = '1'
                     m.ninvamt   = swround(m.namount * (m.nintclass1 / 100), 2)
                  CASE m.cexpclass = '2'
                     m.ninvamt   = swround(m.namount * (m.nintclass2 / 100), 2)
                  CASE m.cexpclass = '3'
                     m.ninvamt   = swround(m.namount * (m.nintclass3 / 100), 2)
                  CASE m.cexpclass = '4'
                     m.ninvamt   = swround(m.namount * (m.nintclass4 / 100), 2)
                  CASE m.cexpclass = '5'
                     m.ninvamt   = swround(m.namount * (m.nintclass5 / 100), 2)
                  CASE m.cexpclass = 'A'
                     m.ninvamt   = swround(m.namount * (m.nacpint / 100), 2)
                  CASE m.cexpclass = 'B'
                     m.ninvamt   = swround(m.namount * (m.nbcpint / 100), 2)
                  OTHERWISE
                     m.ninvamt   = swround(m.namount * (m.nrevgas / 100), 2)
               ENDCASE
               lnMKTGExp = lnMKTGExp + m.ninvamt
               lnwelltot = lnwelltot - m.ninvamt
            ENDSCAN  && mktgtemp


*************************************************************
*  Process well net total
*************************************************************
            m.nworkint = jnworkint
*
* Calculate tax withholding
*
            IF m.ntaxpct # 0
               IF lcdirect = 'O'
                  lntaxgross = lntotinc - lnoilinc
               ELSE
                  IF lcdirect = 'G'
                     lntaxgross = lntotinc - lngasinc
                  ELSE
                     IF lcdirect = 'B'
                        lntaxgross = lntotinc - lngasinc - lnoilinc
                     ELSE
                        lntaxgross = lntotinc
                     ENDIF
                  ENDIF
               ENDIF
               IF THIS.oOptions.lTaxNet
                  m.ntaxwith = swround(lnwelltot * (m.ntaxpct / 100), 2)
               ELSE
                  m.ntaxwith = swround(lntaxgross * (m.ntaxpct / 100), 2)
               ENDIF
            ELSE
               m.ntaxwith = 0
            ENDIF

* Calculate backup withholding
            IF m.lbackwith
               m.nbackwith = swround((lnwelltot - m.ntaxwith) * (m.nbackpct / 100), 2)
            ELSE
               m.nbackwith = 0
            ENDIF

*  Subtract the tax withholding and backup withholding from net total
            lnwelltot = lnwelltot - m.ntaxwith - m.nbackwith

* Initialize ldHistDate
            ldHistDate = THIS.dCheckDate

*  Use Check Date unless we're doing advanced posting and this is the operator's record
            IF THIS.lAdvPosting AND llOperator
               ldHistDate = THIS.dCompanyShare
            ENDIF

            SELECT invtmp
            REPLACE hperiod   WITH m.hperiod, ;
                    hyear      WITH m.hyear, ;
                    hdate      WITH ldHistDate, ;
                    noilrev    WITH lnoilinc, ;
                    ngasrev    WITH lngasinc, ;
                    ntrprev    WITH lntrpinc, ;
                    nothrev    WITH lnothinc, ;
                    nmiscrev1  WITH lnmi1inc, ;
                    nmiscrev2  WITH lnmi2inc, ;
                    nMKTGExp   WITH lnMKTGExp, ;
                    nIncome    WITH lntotinc, ;
                    nexpense   WITH lntotexp, ;
                    nflatrate  WITH m.nflatrate, ;
                    ntotale1   WITH m.ntotale1, ;
                    ntotale2   WITH m.ntotale2, ;
                    ntotale3   WITH m.ntotale3, ;
                    ntotale4   WITH m.ntotale4, ;
                    ntotale5   WITH m.ntotale5, ;
                    ntotalea   WITH m.ntotalea, ;
                    ntotaleb   WITH m.ntotaleb, ;
                    nPlugExp   WITH m.nPlugExp, ;
                    nrevtax1   WITH m.nrevtax1, ;
                    nrevtax2   WITH m.nrevtax2, ;
                    nrevtax3   WITH m.nrevtax3, ;
                    nrevtax4   WITH m.nrevtax4, ;
                    nrevtax5   WITH m.nrevtax5, ;
                    nrevtax6   WITH m.nrevtax6, ;
                    nrevtax7   WITH m.nrevtax7, ;
                    nrevtax8   WITH m.nrevtax8, ;
                    nrevtax9   WITH m.nrevtax9, ;
                    nrevtax10  WITH m.nrevtax10, ;
                    nrevtax11  WITH m.nrevtax11, ;
                    nrevtax12  WITH m.nrevtax12, ;
                    nnetcheck  WITH lnwelltot, ;
                    nsevtaxes  WITH lntaxes, ;
                    noiltax1   WITH lnoiltax1 + lnoilmantax1, ;
                    ngastax1   WITH lngastax1 + lngasmantax1, ;
                    nOthTax1   WITH lnothtax1 + lnothmantax1, ;
                    noiltax2   WITH lnoiltax2 + lnoilmantax2, ;
                    ngastax2   WITH lngastax2 + lngasmantax2, ;
                    nOthTax2   WITH lnothtax2 + lnothmantax2, ;
                    noiltax3   WITH lnoiltax3 + lnoilmantax3, ;
                    ngastax3   WITH lngastax3 + lngasmantax3, ;
                    nOthTax3   WITH lnothtax3 + lnothmantax3, ;
                    noiltax4   WITH lnoiltax4 + lnoilmantax4, ;
                    ngastax4   WITH lngastax4 + lngasmantax4, ;
                    nOthTax4   WITH lnothtax4 + lnothmantax4, ;
                    nGather    WITH lnGather, ;
                    nCompress  WITH lnCompress, ;
                    nbackwith  WITH m.nbackwith, ;
                    ntaxwith   WITH m.ntaxwith

            STORE 0 TO m.ntotale1, m.ntotale2, m.ntotale3, m.ntotale4, m.ntotale5, lnOilTax, lngastax1, ;
               lnmi1inc, lnmi2inc, lnoilinc, lngasinc, lntrpinc, lntotinc, lntotexp, lnGATHExp, lnCOMPExp, ;
               lnoilexp, lngasexp, m.nbackwith, m.ntaxwith, m.ntotalea, m.ntotaleb, lnGather, lnCompress, ;
               lnCalcCompress, lnCalcGather, m.nPlugExp
            lcperiod1 = lcSavePrd1
            lcperiod2 = lcSavePrd2
         ENDSCAN

* Check for negative backup and/or tax withholding amounts for the run and remove them.
         THIS.checkbacktax()

* Remove one man item tax from wellwork if the owner is exempt
         swselect('one_man_tax')
         SCAN
            SCATTER MEMVAR
            SELE investor
            LOCATE FOR cownerid == m.cownerid
            IF FOUND()
               IF lExempt  && The owner is tax exempt
                  SELE wellwork
                  LOCATE FOR cWellID == m.cWellID AND hyear + hperiod = m.hyear + m.hperiod ;
                     AND crunyear == m.crunyear AND nrunno == m.nrunno
                  IF FOUND()  && Remove the owner's share of tax from wellwork
                     REPL ntotbbltx1 WITH ntotbbltx1 - one_man_tax.noiltax1b, ;
                        ntotbbltx2 WITH ntotbbltx2 - one_man_tax.noiltax2b, ;
                        ntotbbltx3 WITH ntotbbltx3 - one_man_tax.noiltax3b, ;
                        ntotbbltx4 WITH ntotbbltx4 - one_man_tax.noiltax4b, ;
                        ntotmcftx1 WITH ntotmcftx1 - one_man_tax.ngastax1b, ;
                        ntotmcftx2 WITH ntotmcftx2 - one_man_tax.ngastax2b, ;
                        ntotmcftx3 WITH ntotmcftx3 - one_man_tax.ngastax3b, ;
                        ntotmcftx4 WITH ntotmcftx4 - one_man_tax.ngastax4b, ;
                        ntotothtx1 WITH ntotothtx1 - one_man_tax.nprodtax1b, ;
                        ntotothtx2 WITH ntotothtx2 - one_man_tax.nprodtax2b, ;
                        ntotothtx3 WITH ntotothtx3 - one_man_tax.nprodtax3b, ;
                        ntotothtx4 WITH ntotothtx4 - one_man_tax.nprodtax4b, ;
                        nGBBLTax1  WITH nGBBLTax1  - one_man_tax.noiltax1b, ;
                        nGBBLTax2  WITH nGBBLTax2  - one_man_tax.noiltax2b, ;
                        nGBBLTax3  WITH nGBBLTax3  - one_man_tax.noiltax3b, ;
                        nGBBLTax4  WITH nGBBLTax4  - one_man_tax.noiltax4b, ;
                        nGMCFTax1  WITH nGMCFTax1  - one_man_tax.ngastax1b, ;
                        nGMCFTax2  WITH nGMCFTax2  - one_man_tax.ngastax2b, ;
                        nGMCFTax3  WITH nGMCFTax3  - one_man_tax.ngastax3b, ;
                        nGMCFTax4  WITH nGMCFTax4  - one_man_tax.ngastax4b, ;
                        nGOTHTax1  WITH nGOTHTax1  - one_man_tax.nprodtax1b, ;
                        nGOTHTax2  WITH nGOTHTax2  - one_man_tax.nprodtax2b, ;
                        nGOTHTax3  WITH nGOTHTax3  - one_man_tax.nprodtax3b, ;
                        nGOTHTax4  WITH nGOTHTax4  - one_man_tax.nprodtax4b
                  ENDIF
               ENDIF
            ENDIF
         ENDSCAN

         SET SAFETY OFF
         SELECT invtmp
         INDEX ON cWellID + cownerid TAG wellinv
         INDEX ON cownerid + cWellID TAG invwell
         INDEX ON lhold TAG lhold
         INDEX ON lonhold TAG lonhold
         INDEX ON ndisbfreq TAG ndisbfreq
         INDEX ON cWellID TAG cWellID
         INDEX ON cownerid + cWellID + ctypeinv + ctypeint + cprogcode TAG invtype
         INDEX ON nrunno TAG nrunno
         INDEX ON crunyear TAG crunyear
         INDEX ON cownerid TAG cownerid

         THIS.oprogress.SetProgressMessage('Allocating Revenue and Expenses to the Owners...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'OwnerProc', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('OwnerProc', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE CheckHist
*********************************
*-- Checks to see if the period is closed.
      LOCAL llHist, llSepClose, lcDeleted
*
*  Checks to see if the given period is closed
*  Returns .T. if the period is closed
*
      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF TYPE('xcDebug') = 'C' AND xcDebug = 'Y'
            WAIT WIND 'In Distproc Checkhist...'
         ENDIF

         lcDeleted = SET('DELETED')
         SET DELETED ON

         llHist = .F.

         IF THIS.cgroup = '**'
            swselect('sysctl')
            LOCATE FOR cyear + cperiod = THIS.crunyear + THIS.cperiod AND lDisbMan AND cTypeClose = 'R'
            IF FOUND()
               llHist        = .T.
               THIS.lrelmin  = sysctl.lrelmin
               THIS.cdmbatch = sysctl.cdmbatch
            ENDIF
         ELSE
            swselect('sysctl')
            SET ORDER TO yrprdgrp
            IF SEEK(THIS.crunyear + THIS.cperiod + THIS.cgroup + 'YR')
               llHist        = .T.
               THIS.lrelmin  = sysctl.lrelmin
               THIS.cdmbatch = sysctl.cdmbatch
            ENDIF
         ENDIF

         SET DELETED &lcDeleted
         llReturn = llHist

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CheckHist', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('CheckHist', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Returns the given flat rate royalty amount.
*********************************
   PROCEDURE GetFlatAmt
*********************************
      LPARA tcwellid, tcType, tcDeck
      LOCAL lcCurrent, lnCount, lnBBL, lnMCF, lnAmount, lcAcctMonth, ldAcctDate
      STORE 0 TO lnBBL, lnMCF, lnAmount

      lnReturn = 0

      IF EMPTY(tcDeck)
         tcDeck = 'DEFAULT'
      ENDIF

      TRY
         IF THIS.lerrorflag
            lnReturn = 0
            EXIT
         ENDIF

         IF NOT THIS.lflatrates
            lnReturn = 0
            EXIT
         ENDIF

         ldAcctDate = THIS.dacctdate
         IF TYPE('ldAcctDate') # 'D'
            ldAcctDate = DATE()
         ENDIF

* Check for flat rates in this well, if none exist return zero.
         swselect('wellinv')
         LOCATE FOR cWellID = tcwellid AND lflat = .T.
         IF NOT FOUND()
            lnReturn = 0
            EXIT
         ENDIF

         lcAcctMonth = PADL(ALLTRIM(STR(MONTH(ldAcctDate), 2)), 2, '0')
         lcCurrent   = SELECT()

         SELECT  ctypeint, ;
                 cflatstart, ;
                 nflatfreq, ;
                 cdeck, ;
                 SUM(nflatrate) AS nflatrate, ;
                 SUM(nflatMCF)  AS nflatMCF, ;
                 SUM(nflatBBL)  AS nflatBBL ;
             FROM wellinv ;
             WHERE cWellID = tcwellid ;
                 AND cdeck == tcDeck ;
                 AND lflat ;
             ORDER BY cdeck,;
                 cflatstart,;
                 nflatfreq ;
             GROUP BY cdeck,;
                 cflatstart,;
                 nflatfreq ;
             INTO CURSOR temp

         IF _TALLY > 0
            SELECT temp
            SCAN
               IF ctypeint # 'B' AND ctypeint # tcType
                  LOOP
               ENDIF
               DO CASE
                  CASE temp.ctypeint = 'B' AND tcType = 'O'
                     lnAmount = 0
                  CASE temp.nflatfreq = 1 OR temp.nflatfreq = 0
                     lnAmount = lnAmount + temp.nflatrate
                  CASE temp.nflatfreq = 2
                     DO CASE
                        CASE LEFT(temp.cflatstart, 2) = '01'
                           lcQtr1 = '01'
                           lcQtr2 = '04'
                           lcQtr3 = '07'
                           lcQtr4 = '10'
                        CASE LEFT(temp.cflatstart, 2) = '02'
                           lcQtr1 = '02'
                           lcQtr2 = '05'
                           lcQtr3 = '08'
                           lcQtr4 = '11'
                        CASE LEFT(temp.cflatstart, 2) = '03'
                           lcQtr1 = '03'
                           lcQtr2 = '06'
                           lcQtr3 = '09'
                           lcQtr4 = '12'
                        CASE LEFT(temp.cflatstart, 2) = '04'
                           lcQtr1 = '04'
                           lcQtr2 = '07'
                           lcQtr3 = '10'
                           lcQtr4 = '01'
                        CASE LEFT(temp.cflatstart, 2) = '05'
                           lcQtr1 = '05'
                           lcQtr2 = '08'
                           lcQtr3 = '11'
                           lcQtr4 = '02'
                        CASE LEFT(temp.cflatstart, 2) = '06'
                           lcQtr1 = '06'
                           lcQtr2 = '09'
                           lcQtr3 = '12'
                           lcQtr4 = '03'
                        CASE LEFT(temp.cflatstart, 2) = '07'
                           lcQtr1 = '07'
                           lcQtr2 = '10'
                           lcQtr3 = '01'
                           lcQtr4 = '04'
                        CASE LEFT(temp.cflatstart, 2) = '08'
                           lcQtr1 = '08'
                           lcQtr2 = '11'
                           lcQtr3 = '02'
                           lcQtr4 = '05'
                        CASE LEFT(temp.cflatstart, 2) = '09'
                           lcQtr1 = '09'
                           lcQtr2 = '12'
                           lcQtr3 = '03'
                           lcQtr4 = '06'
                        CASE LEFT(temp.cflatstart, 2) = '10'
                           lcQtr1 = '10'
                           lcQtr2 = '01'
                           lcQtr3 = '04'
                           lcQtr4 = '07'
                        CASE LEFT(temp.cflatstart, 2) = '11'
                           lcQtr1 = '11'
                           lcQtr2 = '02'
                           lcQtr3 = '05'
                           lcQtr4 = '08'
                        CASE LEFT(temp.cflatstart, 2) = '12'
                           lcQtr1 = '12'
                           lcQtr2 = '03'
                           lcQtr3 = '06'
                           lcQtr4 = '09'
                        OTHERWISE
                           lcQtr1 = '01'
                           lcQtr2 = '04'
                           lcQtr3 = '07'
                           lcQtr4 = '10'
                     ENDCASE
                     IF INLIST(lcAcctMonth, lcQtr1, lcQtr2, lcQtr3, lcQtr4)
                        lnAmount = lnAmount + temp.nflatrate
                     ENDIF
                  CASE temp.nflatfreq = 3
                     DO CASE
                        CASE LEFT(temp.cflatstart, 2) = '01'
                           lcSem1 = '01'
                           lcSem2 = '07'
                        CASE LEFT(temp.cflatstart, 2) = '02'
                           lcSem1 = '02'
                           lcSem2 = '08'
                        CASE LEFT(temp.cflatstart, 2) = '03'
                           lcSem1 = '03'
                           lcSem2 = '09'
                        CASE LEFT(temp.cflatstart, 2) = '04'
                           lcSem1 = '04'
                           lcSem2 = '10'
                        CASE LEFT(temp.cflatstart, 2) = '05'
                           lcSem1 = '05'
                           lcSem2 = '11'
                        CASE LEFT(temp.cflatstart, 2) = '06'
                           lcSem1 = '06'
                           lcSem2 = '12'
                        CASE LEFT(temp.cflatstart, 2) = '07'
                           lcSem1 = '07'
                           lcSem2 = '01'
                        CASE LEFT(temp.cflatstart, 2) = '08'
                           lcSem1 = '08'
                           lcSem2 = '02'
                        CASE LEFT(temp.cflatstart, 2) = '09'
                           lcSem1 = '09'
                           lcSem2 = '03'
                        CASE LEFT(temp.cflatstart, 2) = '10'
                           lcSem1 = '10'
                           lcSem2 = '04'
                        CASE LEFT(temp.cflatstart, 2) = '11'
                           lcSem1 = '11'
                           lcSem2 = '05'
                        CASE LEFT(temp.cflatstart, 2) = '12'
                           lcSem1 = '12'
                           lcSem2 = '06'
                        OTHERWISE
                           lcSem1 = '01'
                           lcSem2 = '07'
                     ENDCASE
                     IF INLIST(lcAcctMonth, lcSem1, lcSem2)
                        lnAmount = lnAmount + temp.nflatrate
                     ENDIF
                  CASE temp.nflatfreq = 4
                     IF LEFT(temp.cflatstart, 2) = lcAcctMonth
                        lnAmount = lnAmount + temp.nflatrate
                     ENDIF
                  OTHERWISE
                     lnAmount = lnAmount + temp.nflatrate
               ENDCASE
            ENDSCAN
         ELSE
            lnAmount = 0
         ENDIF

         STORE 0 TO lnMCF, lnBBL
         STORE 0 TO lnFlatMCF, lnFlatBBL

*  Get total mcf for this run if we have flat rate per mcf owners
*  Calculate the total flat rate per mcf total for the well
         IF tcType = 'G'
            lnRevInt = 0
            swselect('wellinv')
            SCAN FOR cWellID = tcwellid AND cdeck == tcDeck AND lflat AND nflatMCF # 0

               SELE SUM(nUnits) AS nTotMCF, AVG(nprice) AS nprice FROM income ;
                  WHERE cSource = 'MCF' AND (nrunno = 0 AND drevdate <= THIS.drevdate OR (nrunno = THIS.nrunno AND crunyear = THIS.crunyear)) ;
                  AND cWellID == tcwellid ;
                  INTO CURSOR inctmp
               SELE inctmp
               lnMCF     = nTotMCF
               jnFlatMCF = (swround(lnMCF * nprice * (wellinv.nrevgas / 100), 2) - swround(lnMCF * wellinv.nflatMCF * (wellinv.nrevgas / 100), 2)) * -1
               lnFlatMCF = lnFlatMCF + jnFlatMCF
            ENDSCAN
         ENDIF

* Get total bbl for this run if we have flat rate per bbl owners
*  Calculate the total flat rate per bbl total for the well
         IF tcType = 'O'
            lnRevInt = 0
            swselect('wellinv')
            SCAN FOR cWellID = tcwellid AND cdeck == tcDeck AND lflat AND nflatBBL # 0

               SELE SUM(nUnits) AS nTotBBL, AVG(nprice) AS nprice FROM income ;
                  WHERE cSource = 'BBL' AND (nrunno = 0 AND drevdate <= THIS.drevdate  OR (nrunno = THIS.nrunno AND crunyear = THIS.crunyear)) ;
                  AND cWellID == tcwellid ;
                  INTO CURSOR inctmp
               SELE inctmp
               lnBBL     = nTotBBL
               jnFlatBBL = (swround(lnBBL * nprice * (wellinv.nrevoil / 100), 2) - swround(lnBBL * wellinv.nflatBBL * (wellinv.nrevoil / 100), 2)) * -1
               lnFlatBBL = lnFlatBBL + jnFlatBBL
            ENDSCAN
         ENDIF

* Add the flat rate, flat rate per mcf and flat rate per bbl totals
         lnAmount = lnAmount + lnFlatMCF + lnFlatBBL

         SELECT (lcCurrent)

         lnReturn = lnAmount

      CATCH TO loError
         lnReturn = 0
         DO errorlog WITH 'GetFlatAmt', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('GetFlatAmt', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN (lnReturn)
   ENDPROC

*********************************
   PROCEDURE GetOwnFlat
*********************************
      LPARA tcwellid, tcOwner, tcType, tciddisb

      LOCAL lcCurrent, lnCount, lnBBL, lnMCF, lnAmount, lcAcctMonth, ldAcctDate, lnFlatRate
      LOCAL lcFlatStart, lcQtr1, lcQtr2, lcQtr3, lcQtr4, lcSem1, lcSem2, lcWellID, llReturn, lnFlatFreq
      LOCAL lnFlatRateBBL, lnFlatRateMCF, lnReturn, lnRevInt, loError

      lnReturn = 0

      TRY
         STORE 0 TO lnBBL, lnMCF, lnAmount, lnFlatRate

         IF THIS.lerrorflag
            lnReturn = 0
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF NOT THIS.lflatrates
            lnReturn = 0
            EXIT
         ENDIF

         ldAcctDate = THIS.dacctdate
         IF TYPE('ldAcctDate') # 'D'
            ldAcctDate = DATE()
         ENDIF

         lcAcctMonth = PADL(ALLTRIM(STR(MONTH(ldAcctDate), 2)), 2, '0')
         lcCurrent   = SELECT()

         swselect('wellinv')
         SET ORDER TO cidwinv
         IF SEEK(tciddisb)
            lcWellID = cWellID
            IF INLIST(ctypeinv, 'L', 'O')
               lnFlatRate  = nflatrate
               lnFlatFreq  = nflatfreq
               lcFlatStart = cflatstart

               IF tcType = 'B'
                  IF nflatBBL = 0
                     tcType = 'G'
                  ENDIF
                  IF nflatMCF = 0
                     tcType = 'O'
                  ENDIF
               ENDIF

               DO CASE
                  CASE tcType = 'G'
                     lnFlatRateMCF = nflatMCF
                     lnRevInt      = nrevgas
                     lnFlatRateBBL = 0
                  CASE tcType = 'O'
                     lnFlatRateBBL = nflatBBL
                     lnRevInt      = nrevoil
                     lnFlatRateMCF = 0
                  OTHERWISE
                     STORE 0 TO lnFlatRateMCF, lnFlatRateBBL, lnRevInt
               ENDCASE
            ELSE
               lnFlatRate    = 0
               lnFlatFreq    = 1
               lcFlatStart   = '01'
               lnFlatRateMCF = 0
               lnFlatRateBBL = 0
            ENDIF
         ELSE
            lcWellID      = '@@@@***'
            lnFlatRate    = 0
            lnFlatFreq    = 1
            lcFlatStart   = '01'
            lnFlatRateMCF = 0
            lnFlatRateBBL = 0
         ENDIF

         DO CASE
            CASE lnFlatRate > 0
               DO CASE
                  CASE lnFlatFreq = 1 OR lnFlatFreq = 0

                  CASE lnFlatFreq = 2
                     DO CASE
                        CASE LEFT(lcFlatStart, 2) = '01'
                           lcQtr1 = '01'
                           lcQtr2 = '04'
                           lcQtr3 = '07'
                           lcQtr4 = '10'
                        CASE LEFT(lcFlatStart, 2) = '02'
                           lcQtr1 = '02'
                           lcQtr2 = '05'
                           lcQtr3 = '08'
                           lcQtr4 = '11'
                        CASE LEFT(lcFlatStart, 2) = '03'
                           lcQtr1 = '03'
                           lcQtr2 = '06'
                           lcQtr3 = '09'
                           lcQtr4 = '12'
                        CASE LEFT(lcFlatStart, 2) = '04'
                           lcQtr1 = '04'
                           lcQtr2 = '07'
                           lcQtr3 = '10'
                           lcQtr4 = '01'
                        CASE LEFT(lcFlatStart, 2) = '05'
                           lcQtr1 = '05'
                           lcQtr2 = '08'
                           lcQtr3 = '11'
                           lcQtr4 = '02'
                        CASE LEFT(lcFlatStart, 2) = '06'
                           lcQtr1 = '06'
                           lcQtr2 = '09'
                           lcQtr3 = '12'
                           lcQtr4 = '03'
                        CASE LEFT(lcFlatStart, 2) = '07'
                           lcQtr1 = '07'
                           lcQtr2 = '10'
                           lcQtr3 = '01'
                           lcQtr4 = '04'
                        CASE LEFT(lcFlatStart, 2) = '08'
                           lcQtr1 = '08'
                           lcQtr2 = '11'
                           lcQtr3 = '02'
                           lcQtr4 = '05'
                        CASE LEFT(lcFlatStart, 2) = '09'
                           lcQtr1 = '09'
                           lcQtr2 = '12'
                           lcQtr3 = '03'
                           lcQtr4 = '06'
                        CASE LEFT(lcFlatStart, 2) = '10'
                           lcQtr1 = '10'
                           lcQtr2 = '01'
                           lcQtr3 = '04'
                           lcQtr4 = '07'
                        CASE LEFT(lcFlatStart, 2) = '11'
                           lcQtr1 = '11'
                           lcQtr2 = '02'
                           lcQtr3 = '05'
                           lcQtr4 = '08'
                        CASE LEFT(lcFlatStart, 2) = '12'
                           lcQtr1 = '12'
                           lcQtr2 = '03'
                           lcQtr3 = '06'
                           lcQtr4 = '09'
                        OTHERWISE
                           lcQtr1 = '01'
                           lcQtr2 = '04'
                           lcQtr3 = '07'
                           lcQtr4 = '10'
                     ENDCASE
                     IF NOT INLIST(lcAcctMonth, lcQtr1, lcQtr2, lcQtr3, lcQtr4)
                        lnFlatRate = 0
                     ENDIF
                  CASE lnFlatFreq = 3
                     DO CASE
                        CASE LEFT(lcFlatStart, 2) = '01'
                           lcSem1 = '01'
                           lcSem2 = '07'
                        CASE LEFT(lcFlatStart, 2) = '02'
                           lcSem1 = '02'
                           lcSem2 = '08'
                        CASE LEFT(lcFlatStart, 2) = '03'
                           lcSem1 = '03'
                           lcSem2 = '09'
                        CASE LEFT(lcFlatStart, 2) = '04'
                           lcSem1 = '04'
                           lcSem2 = '10'
                        CASE LEFT(lcFlatStart, 2) = '05'
                           lcSem1 = '05'
                           lcSem2 = '11'
                        CASE LEFT(lcFlatStart, 2) = '06'
                           lcSem1 = '06'
                           lcSem2 = '12'
                        CASE LEFT(lcFlatStart, 2) = '07'
                           lcSem1 = '07'
                           lcSem2 = '01'
                        CASE LEFT(lcFlatStart, 2) = '08'
                           lcSem1 = '08'
                           lcSem2 = '02'
                        CASE LEFT(lcFlatStart, 2) = '09'
                           lcSem1 = '09'
                           lcSem2 = '03'
                        CASE LEFT(lcFlatStart, 2) = '10'
                           lcSem1 = '10'
                           lcSem2 = '04'
                        CASE LEFT(lcFlatStart, 2) = '11'
                           lcSem1 = '11'
                           lcSem2 = '05'
                        CASE LEFT(lcFlatStart, 2) = '12'
                           lcSem1 = '12'
                           lcSem2 = '06'
                        OTHERWISE
                           lcSem1 = '01'
                           lcSem2 = '07'
                     ENDCASE
                     IF NOT INLIST(lcAcctMonth, lcSem1, lcSem2)
                        lnFlatRate = 0
                     ENDIF
                  CASE lnFlatFreq = 4
                     IF LEFT(lcFlatStart, 2) # lcAcctMonth
                        lnFlatRate = 0
                     ENDIF
                  OTHERWISE
                     lnFlatRate = 0
               ENDCASE
            CASE lnFlatRateMCF # 0 AND tcType = 'G'
               SELE SUM(nUnits) AS nTotMCF FROM income ;
                  WHERE cSource = 'MCF' AND (nrunno = 0 OR (nrunno = THIS.nrunno AND crunyear = THIS.crunyear)) AND drevdate <= THIS.drevdate ;
                  AND cWellID == lcWellID ;
                  INTO CURSOR inctmp
               SELE inctmp
               GO TOP
               lnMCF      = nTotMCF * (lnRevInt / 100)
               lnFlatRate = swround(lnMCF * lnFlatRateMCF, 2)
            CASE lnFlatRateBBL # 0 AND tcType = 'O'
               SELE SUM(nUnits) AS nTotBBL FROM income ;
                  WHERE cSource = 'BBL' AND (nrunno = 0 OR (nrunno = THIS.nrunno AND crunyear = THIS.crunyear)) AND drevdate <= THIS.drevdate ;
                  AND cWellID == lcWellID ;
                  INTO CURSOR inctmp
               SELE inctmp
               GO TOP
               lnBBL      = nTotBBL
               lnFlatRate = swround(lnBBL * lnFlatRateBBL, 2)
         ENDCASE

         SELECT (lcCurrent)

         lnReturn = lnFlatRate
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'GetOwnFlat', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('GetOwnFlat', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN lnReturn
   ENDPROC

*-- Adds the current processing to the history files.
*********************************
   PROCEDURE AddHist
*********************************
      LOCAL lnMax, lnCount, lnX, oprogress
      LOCAL llReturn, loError
      LOCAL cAcctYr, cacctprd, ciddisb, cidwhst

      llReturn = .T.

      TRY
         STORE 0 TO lnMax, lnCount


         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         SET DELETED ON
         THIS.oprogress.SetProgressMessage('Adding The Current Run Data to the History Files...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

         SELECT invtmp
         COUNT FOR NOT DELETED() TO lnX
         lnMax = lnMax + lnX
         SELECT wellwork
         COUNT FOR NOT DELETED() TO lnX
         lnMax = lnMax + lnX

* Set the order outside of the loop
         SELECT disbhist
         SET ORDER TO ciddisb
         SELECT ownpcts
         SET ORDER TO ciddisb

         lciddisb  = GetNextPK('DISBHIST')
         THIS.oprogress.SetProgressMessage('Adding The Current Run Data to the History Files...Owner History')
         SELECT invtmp
         SET ORDER TO 0
         SCAN
            lnCount = lnCount + 1
            SCATTER MEMVAR

            IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
               llReturn          = .F.
               IF NOT m.goApp.CancelMsg()
                  THIS.lCanceled = .T.
                  EXIT
               ENDIF
            ENDIF

            m.ciddisb = lciddisb

            IF EMPTY(m.csusptype) AND m.nrunno_in # 0
               REPLACE nrunno_in WITH 0, crunyear_in WITH ''
            ENDIF

            m.cAcctYr  = THIS.cacctyear
            m.cacctprd = THIS.cacctprd

            INSERT INTO disbhist FROM MEMVAR
            INSERT INTO ownpcts FROM MEMVAR
            lciddisb = GetNextPK('DISBHIST')
            SET DELETED ON
         ENDSCAN

         THIS.oprogress.SetProgressMessage('Adding The Current Run Data to the History Files...Well History')
         SELECT wellwork
         SET ORDER TO 0
         SCAN
            lnCount = lnCount + 1
            SCATTER MEMVAR

            IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
               llReturn          = .F.
               IF NOT m.goApp.CancelMsg()
                  THIS.lCanceled = .T.
                  EXIT
               ENDIF
            ENDIF

            m.cAcctYr  = THIS.cacctyear
            m.cacctprd = THIS.cacctprd
            m.cidwhst  = GetNextPK('WELLHIST')
            SELECT wellhist1
            INSERT INTO wellhist FROM MEMVAR
         ENDSCAN
         THIS.oprogress.SetProgressMessage('Adding The Current Run Data to the History Files...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'AddHist', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('AddHist', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Retrieves owner and well history records from closed periods.
*********************************
   PROCEDURE GetHist
*********************************
      LOCAL lnCount, lnMax, lnX, lcBegID, lcEndID, lcRunYear, llReturn

      llReturn = .T.

      TRY
         STORE 0 TO lnCount, lnMax, lnX

         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


*   Create Temp Investor Disbursement File
         swselect('disbhist')
         lnX = AFIELDS(latempx)
         swselect('ownpcts')
         lny = AFIELDS(latempy)
         DIMENSION latemp[lnx + lny - 1, 18]
         FOR x = 1 TO lnX
            FOR Y = 1 TO 18
               latemp[x, y] = latempx[x, y]
            ENDFOR
         ENDFOR
         FOR x = 1 TO lny - 1
            FOR Y = 1 TO 18
               latemp[x + lnx, y] = latempy[x + 1, y]
            ENDFOR
         ENDFOR
         FOR x = 1 TO lnX + lny - 1
            latemp[X, 7]  = ''
            latemp[X, 8]  = ''
            latemp[X, 9]  = ''
            latemp[X, 10] = ''
            latemp[X, 11] = ''
            latemp[X, 12] = ''
            latemp[X, 13] = ''
            latemp[X, 14] = ''
            latemp[X, 15] = ''
            latemp[X, 18] = ''
         ENDFOR
         CREATE CURSOR invtmp FROM ARRAY latemp

         lcRunYear = THIS.crunyear + PADL(TRANSFORM(THIS.nrunno), 3, '0')


         THIS.oprogress.SetProgressMessage('Retrieving Owner History Records...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

* Check for archived year and set up disbhist and ownpcts for that year if needed
         TRY
            llArchived = THIS.CheckArchived(THIS.crunyear)
            lcYear     = THIS.crunyear

            IF llArchived
               lcFileHist = 'ownhist' + lcYear
               lcFilePcts = 'ownpcts' + lcYear
               swselect(lcFileHist)
               swselect(lcFilePcts)
               swclose('disbhist')

               WAIT WINDOW NOWAIT 'Retrieving Archived History for ' + lcYear
               SELECT *, crunyear_i AS crunyear_in FROM (lcFileHist) INTO CURSOR disbhist READWRITE

* Created needed indicies
               INDEX ON ciddisb TAG ciddisb
               INDEX ON crunyear TAG crunyear
               INDEX ON nrunno  TAG nrunno

               swclose('ownpcts')
               SELECT * FROM (lcFilePcts) INTO CURSOR ownpcts  READWRITE

* Create needed indicies
               INDEX ON ciddisb TAG ciddisb
               WAIT CLEAR
            ENDIF
         CATCH TO loError
            MESSAGEBOX(loError.MESSAGE, 0, 'Archived Hist')
         ENDTRY

         swselect('disbhist')
         IF THIS.cgroup = '**'          && All Groups
            SELECT  disbhist.*, ;
                    ownpcts.* ;
                FROM disbhist,;
                    ownpcts ;
                WHERE  ((disbhist.crunyear + PADL(TRANSFORM(disbhist.nrunno), 3, '0') = lcRunYear ;
                        AND INLIST(disbhist.crectype, 'R', 'P')) OR ;
                      (disbhist.crunyear_in + PADL(TRANSFORM(disbhist.nrunno_in), 3, '0') = lcRunYear ;
                        AND disbhist.crectype = 'R')) ;
                    AND BETWEEN(cownerid, THIS.cbegownerid, THIS.cendownerid) ;
                    AND disbhist.ciddisb == ownpcts.ciddisb ;
                INTO CURSOR temp
            IF _TALLY > 0
               SELECT temp
               SCAN
                  SCATTER MEMVAR
                  INSERT INTO invtmp FROM MEMVAR
               ENDSCAN
               THIS.dacctdate = m.hdate
            ENDIF
            IF THIS.lreport
               SELECT  disbhist.*, ;
                       ownpcts.* ;
                   FROM disbhist,;
                       ownpcts ;
                   WHERE disbhist.crunyear_in + PADL(TRANSFORM(disbhist.nrunno_in), 3, '0') = lcRunYear ;
                       AND disbhist.crectype = 'R'  ;
                       AND BETWEEN(disbhist.cownerid, THIS.cbegownerid, THIS.cendownerid) ;
                       AND disbhist.ciddisb == ownpcts.ciddisb ;
                   INTO CURSOR temp
               IF _TALLY > 0
                  SELECT temp
                  SCAN
                     SCATTER MEMVAR
                     m.nrunno   = m.nrunno_in
                     m.crunyear = m.crunyear_in
                     INSERT INTO invtmp FROM MEMVAR
                  ENDSCAN
                  THIS.dacctdate = m.hdate
               ENDIF
            ENDIF
         ELSE
            SELECT  disbhist.*, ;
                    ownpcts.* ;
                FROM disbhist,;
                    ownpcts ;
                WHERE  ((disbhist.crunyear + PADL(TRANSFORM(disbhist.nrunno), 3, '0') = lcRunYear ;
                        AND INLIST(disbhist.crectype, 'R', 'P')) OR ;
                      (disbhist.crunyear_in + PADL(TRANSFORM(disbhist.nrunno_in), 3, '0') = lcRunYear ;
                        AND disbhist.crectype = 'R')) ;
                    AND BETWEEN(disbhist.cownerid, THIS.cbegownerid, THIS.cendownerid) ;
                    AND disbhist.cgroup = THIS.cgroup ;
                    AND disbhist.ciddisb == ownpcts.ciddisb ;
                INTO CURSOR temp
            IF _TALLY > 0
               SELECT temp
               SCAN
                  SCATTER MEMVAR
                  INSERT INTO invtmp FROM MEMVAR
               ENDSCAN
               THIS.dacctdate = m.hdate
            ENDIF
            IF THIS.lreport AND NOT THIS.lQBPost
               SELECT  suspense.* ;
                   FROM suspense ;
                   WHERE suspense.crunyear_in + PADL(TRANSFORM(suspense.nrunno_in), 3, '0') = lcRunYear ;
                       AND BETWEEN(suspense.cownerid, THIS.cbegownerid, THIS.cendownerid) ;
                       AND suspense.crectype = 'R' ;
                       AND suspense.cgroup = THIS.cgroup ;
                   INTO CURSOR temp
               IF _TALLY > 0
                  SELECT temp
                  SCAN
                     SCATTER MEMVAR
                     m.nrunno   = m.nrunno_in
                     m.crunyear = m.crunyear_in
                     INSERT INTO invtmp FROM MEMVAR
                  ENDSCAN
                  THIS.dacctdate = m.hdate
               ENDIF
            ENDIF
         ENDIF

         SET SAFETY OFF
         SELECT invtmp
         INDEX ON cWellID + cownerid TAG wellinv
         INDEX ON cownerid + cWellID TAG invwell
         INDEX ON cownerid + cWellID + ctypeinv + ctypeint + cprogcode TAG invtype
         INDEX ON cWellID TAG cWellID
         INDEX ON cownerid + cprogcode + cWellID + ctypeinv + hyear + hperiod TAG invprog
         INDEX ON cprogcode TAG cprogcode
         INDEX ON hyear + hperiod TAG yearprd
         INDEX ON cownerid + cprogcode + cWellID + ctypeinv TAG ownertype
         INDEX ON DELETED() TAG _deleted BINARY

*
****************************************************************
*   Create Temp Well Production History File
****************************************************************
*

         THIS.oprogress.SetProgressMessage('Retrieving Well History Records...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1
         swselect('wellhist')
         Make_Copy('wellhist', 'wellwork')

         SELECT wellwork
         INDEX ON cWellID TAG cWellID
         INDEX ON hyear + hperiod TAG yearprd
         INDEX ON DELETED() TAG _deleted BINARY
         INDEX ON cWellID + hyear + hperiod + crectype + crunyear + PADL(ALLTRIM(STR(nrunno)), 3, '0') TAG wellprdrun
         INDEX ON cWellID + hyear + hperiod + crectype TAG wellprd

         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

         IF THIS.cgroup = '**'
            swselect('wellhist')
            SELECT  * ;
                FROM wellhist ;
                WHERE  nrunno = THIS.nrunno ;
                    AND  crunyear = THIS.crunyear ;
                    AND crectype = 'R' ;
                    AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                INTO CURSOR tempwhist READWRITE

            SELECT tempwhist
            SCAN
               SCATTER MEMVAR
               INSERT INTO wellwork FROM MEMVAR
            ENDSCAN
         ELSE
            SELECT  * ;
                FROM wellhist ;
                WHERE  nrunno = THIS.nrunno ;
                    AND  crunyear = THIS.crunyear ;
                    AND BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
                    AND crectype = 'R' ;
                    AND cgroup = THIS.cgroup ;
                INTO CURSOR tempwhist READWRITE
            IF _TALLY > 0
               SELECT wellwork
               APPEND FROM DBF('tempwhist')
            ENDIF
         ENDIF


* Make sure all the correpsonding wellwork records are there
         SELECT invtmp
         SCAN
            m.hyear   = hyear
            m.hperiod = hperiod
            m.cWellID = cWellID
            m.nrunno  = nrunno
            IF nrunno_in # 0
               m.nrunno = nrunno_in
            ENDIF
            m.crunyear = crunyear
            IF NOT EMPTY(crunyear_in)
               m.crunyear = crunyear_in
            ENDIF
            IF THIS.cprocess # 'W'
               THIS.osuspense.getwellhist(m.cWellID, m.hyear, m.hperiod, m.crunyear, m.nrunno)
            ENDIF
         ENDSCAN

* Open the suspense table as tsuspense so reports don't bomb
         IF NOT USED('tsuspense')
            USE suspense AGAIN IN 0 ALIAS tsuspense
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'GetHist', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('GetHist', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC


*-- Closes the revenue period.
*********************************
   PROCEDURE CloseProc
*********************************
      LOCAL oprogress, llReturn1, llReturn2
      LOCAL lCloseYear, lCompanyPost, lDisbMan, lSumExp, lSumRev, llNetDef, llOK, llReturn, llSummaryPost
      LOCAL llSummaryWell, lnBuff, loError, lrelmin, lrelqtr
      LOCAL cTimeClose, cTypeClose, cdmbatch, cgroup, cidsysctl, cperiod, crunyear, cversion, cyear
      LOCAL dDateClose, dacctdate, dexpdate, dpostdate, drevdate, nrunno

      TRY

         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               THIS.cErrorMsg = 'Processing canceled by user.'
               EXIT
            ENDIF
         ENDIF

         STORE .T. TO llReturn1, llReturn2


*  Mark the Revenue period as being closed
         m.cdmbatch      = GetNextPK('BATCH')
         m.cidsysctl     = GetNextPK('SYSCTL')
         m.lSumRev       = THIS.oOptions.lRevSum
         m.lSumExp       = THIS.oOptions.lexpsum
         THIS.csysctlkey = m.cidsysctl
         THIS.cdmbatch   = m.cdmbatch
         m.cperiod       = THIS.cperiod
         m.cyear         = THIS.crunyear
         m.crunyear      = THIS.crunyear
         m.dacctdate     = THIS.dacctdate
         m.dpostdate     = THIS.dpostdate
         m.drevdate      = THIS.drevdate
         m.dexpdate      = THIS.dexpdate
         m.dCompanyShare = THIS.dCompanyShare
         m.cgroup        = THIS.cgroup
         m.dDateClose    = DATE()
         m.cTimeClose    = TIME()
         m.cTypeClose    = 'R'
         m.lrelmin       = THIS.lrelmin
         m.lrelqtr       = THIS.lrelqtr
         m.lCloseYear    = .F.
         m.lDisbMan      = .T.
         m.nrunno        = THIS.nrunno
         m.lCompanyPost  = THIS.lAdvPosting
         m.dAdvPosting   = THIS.dCompanyShare
         m.cversion      = STRTRAN(m.goApp.cfileversion, '.', '')

* Set the posted flag if the following file exists.
* This is for companies that are using the QB version but
* Don't post to QB. (Northfield Enterprises)
         IF FILE(m.goApp.cCommonFolder + 'markposted.txt')
            m.lposted = .T.
         ENDIF
         INSERT INTO sysctl FROM MEMVAR

         llReturn = THIS.CalcRounding(.T.)
         IF NOT llReturn
            EXIT
         ENDIF

*  Check for non-jib wells, replace jib runno with rev runno if well is all net
         llReturn = THIS.allnetcheck()
         IF NOT llReturn
            EXIT
         ENDIF

*  Perform suspense processing
         swselect('groups')
         SET ORDER TO cgroup
         IF SEEK(THIS.cgroup)
            llNetDef = lNetDef
         ELSE
            llNetDef = .T.
         ENDIF

* Processing suspense
         llReturn = THIS.suspense(llNetDef, .T.)
         IF NOT llReturn
            EXIT
         ENDIF

*  Add the owner history and well history records.
         llReturn = THIS.AddHist()
         IF llReturn = .F.
            EXIT
         ENDIF

* Create the checks for owners and vendors
* in the non-AM versions
         IF NOT m.goApp.lAMVersion
            llReturn = THIS.ownerchks()
            IF NOT llReturn
               EXIT
            ENDIF

            llReturn = THIS.vendorchks()
            IF NOT llReturn
               EXIT
            ENDIF

* Only call directdeposit if the module is available
            IF m.goApp.lDirDMDep
               llReturn = THIS.directdeposit()
               IF NOT llReturn
                  EXIT
               ENDIF
            ENDIF
         ENDIF

*  Mark the expense entries as being tied to this DM batch
         THIS.oprogress.SetProgressMessage('Marking Expenses as Processed...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1
         swselect('expense')
         SCAN FOR nRunNoRev = THIS.nrunno ;
               AND cRunYearRev = THIS.crunyear ;
               AND EMPTY(expense.cBatch)
            swselect('expense')
            REPL cBatch WITH THIS.cdmbatch
         ENDSCAN

*  Create the journal entries (AM Version)
         IF m.goApp.lAMVersion
* Get the summary posting options
            llSummaryPost = THIS.oOptions.lSummaryPost
            llSummaryWell = THIS.oOptions.lSummaryWell
            IF llSummaryPost
               IF llSummaryWell
* Post journal entries summarized by well
                  llReturn = THIS.PostSummaryWell()
                  IF NOT llReturn
                     EXIT
                  ENDIF
               ELSE
* Post summarized journal entries
                  llReturn = THIS.postsummary()
                  IF NOT llReturn
                     EXIT
                  ENDIF
               ENDIF
            ELSE
               llReturn = THIS.postjourn()
               IF llReturn = .F.
                  EXIT
               ENDIF
            ENDIF
         ENDIF

         IF THIS.lCanceled
            EXIT
         ENDIF

         IF m.goApp.lPluggingModule
            llReturn = THIS.PluggingFund()
         ENDIF

         IF THIS.lCanceled
            EXIT
         ENDIF

         IF NOT llReturn OR THIS.lerrorflag
*  Rollback the entries
            SELE sysctl
            DELETE FROM sysctl WHERE  cidsysctl = THIS.csysctlkey
            llReturn = .F.
            EXIT
         ELSE


            IF THIS.lCanceled
               EXIT
            ENDIF

*  Save the entries
            THIS.oprogress.SetProgressMessage('Finalizing Run Closing...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1

            IF m.goApp.lAMVersion
               IF THIS.ldebug
                  lnBuff = CURSORGETPROP('Buffering', 'disbhist')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table disbhist is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'ownpcts')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table ownpcts is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'glmaster')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table glmaster is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'wellhist')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table sysctl is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'coabal')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table coabal is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'checks')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table checks is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'expense')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table expense is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'income')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table income is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'suspense')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table suspense is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'roundtmp')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table roundtmp is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
                  lnBuff = CURSORGETPROP('Buffering', 'one_man_tax')
                  IF lnBuff # 5
                     MESSAGEBOX('Buffering for table one_man_tax is: ' + TRANSFORM(lnBuff), 0, 'Bad Buffering')
                  ENDIF
               ENDIF
               llOK = .T.
               BEGIN TRANSACTION
               SELE disbhist
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating owner history.'
                  EXIT
               ENDIF
               SELE ownpcts
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating owner interest history.'
                  EXIT
               ENDIF
               SELE glmaster
               llOK = TABLEUPDATE(2, .T., 'glmaster', laErrorRecords)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating G/L journal.'
                  EXIT
               ENDIF
               SELE wellhist
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating well history.'
                  EXIT
               ENDIF
               SELE sysctl
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating close control.'
                  EXIT
               ENDIF
               SELE coabal
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating account balances.'
                  EXIT
               ENDIF
               SELE checks
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating check register.'
                  EXIT
               ENDIF
               SELE expense
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating well expenses.'
                  EXIT
               ENDIF
               SELE income
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating well revenue.'
                  EXIT
               ENDIF
               SELE suspense
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  AERROR(updateerror)
                  MESSAGEBOX('Error Updating Suspense' + CHR(10) + ;
                       'Error: ' + TRANSFORM(updateerror[1]) + CHR(10) + ;
                       'Message: ' + updateerror[2] + CHR(10) + ;
                       'Extra: ' + TRANSFORM(updateerror[5]), 0, 'Suspense Update Error')
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating suspense.'
                  EXIT
               ENDIF
               SELE roundtmp
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating rounding.'
                  EXIT
               ENDIF
               SELE one_man_tax
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating one man tax.'
                  EXIT
               ENDIF
               swselect('plugwellbal',.T.)
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating plugging fund.'
                  EXIT
               ENDIF
               SELECT arpmthdr
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating A/R Pmt Hdr.'
                  EXIT
               ENDIF
               SELECT arpmtdet
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating A/R Pmt Det.'
                  EXIT
               ENDIF
               SELECT stmtnote
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating Statement Note.'
                  EXIT
               ENDIF
               END TRANSACTION
            ELSE
               llOK = .T.
               BEGIN TRANSACTION
               SELE disbhist
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  AERROR(updateerror)
                  MESSAGEBOX('Error Updating Owner History' + CHR(10) + ;
                       'Error: ' + TRANSFORM(updateerror[1]) + CHR(10) + ;
                       'Message: ' + updateerror[2], 0, 'Disbhist Update Error')
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating owner history.'
                  EXIT
               ENDIF
               SELE ownpcts
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating owner interest history.'
                  EXIT
               ENDIF
               SELE wellhist
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating well history.'
                  EXIT
               ENDIF
               SELE sysctl
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn = .F.
                  EXIT
               ENDIF
               SELE checks
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating close control.'
                  EXIT
               ENDIF
               SELE expense
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating well expenses.'
                  EXIT
               ENDIF
               SELE income
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating well revenue.'
                  EXIT
               ENDIF
               SELE suspense
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
*!*                     AERROR(updateerror)
*!*                     MESSAGEBOX('Error Updating Suspense' + CHR(10) + ;
*!*                        'Error: ' + TRANSFORM(updateerror[1]) + CHR(10) + ;
*!*                        'Message: ' + updateerror[2], 0, 'Suspense Update Error')
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating suspense.'
                  EXIT
               ENDIF
               SELE roundtmp
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating rounding.'
                  EXIT
               ENDIF
               SELE one_man_tax
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating one man tax.'
                  EXIT
               ENDIF
               swselect('plugwellbal',.T.)
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating plugging fund.'
                  EXIT
               ENDIF
               SELECT arpmthdr
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating A/R Pmt Hdr.'
                  EXIT
               ENDIF
               SELECT arpmtdet
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating A/R Pmt Det.'
                  EXIT
               ENDIF
               SELECT stmtnote
               llOK = TABLEUPDATE(.T., .T.)
               IF NOT llOK
                  ROLLBACK
                  llReturn       = .F.
                  THIS.cErrorMsg = 'Problem updating Statement Note.'
                  EXIT
               ENDIF
               END TRANSACTION
            ENDIF

*  Only do this if it's the AM, since these are AM-only options - BH 11/05/10
            IF m.goApp.lAMVersion
               IF llSummaryPost AND NOT llSummaryWell
                  THIS.PostTransferSummary()
               ENDIF
            ENDIF

* We have the partnership module so create the partner posting records
            IF m.goApp.lPartnershipMod

               THIS.oPartnerShip.lSelected = .T.
               THIS.oPartnerShip.nrunno    = THIS.nrunno
               THIS.oPartnerShip.crunyear  = THIS.crunyear
               THIS.oPartnerShip.cgroup    = THIS.cgroup
               THIS.oPartnerShip.dacctdate = THIS.dacctdate
               THIS.oPartnerShip.cdmbatch  = THIS.cdmbatch

               llReturn = THIS.oPartnerShip.CreatePartnerPost()

* Mark any settlement statement notes has having been used this run
               llReturn = THIS.oPartnerShip.MarkSettlementNotes(THIS.crunyear, THIS.nrunno)

            ENDIF

            THIS.lclosed       = .T.
            THIS.lRptWellExcpt = .T.
* Print the closing summary reports
            THIS.print_closing_summary()

            WAIT CLEAR
            llReturn = .T.
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CloseProc', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('CloseProc', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()
      RETURN llReturn
   ENDPROC

* Opens a given revenue run
*********************************
   PROCEDURE ReOpen_Rev_Run
*********************************
      LPARAMETERS lcRunYear, lnRunNo, lcGroup, lcDMBatch, lnDataSession, tlquiet, tlForce
      LOCAL lcsyskey, lcDMBatchJ, lcYear, lcPeriod, lcIDChec
      LOCAL llSepClose, oGLMaint, lcRunNo, lnSession, llVoidCheck, llDirDep
      LOCAL lcMessage, lcOldRunYear, lcVersion, lnReturn, lnOldRunNo, lnPrtCount
      LOCAL loError, lnMax, oprogress, lnProgress

*:Global cBatch, ciddisb, cidsysctl, crunyear, dpostdate, nrunno

      lnReturn = 0

      _VFP.AUTOYIELD = .T.
      SET ESCAPE ON
* Setup the ability to cancel processing
      ON ESCAPE m.goApp.lCanceled = .T.

      TRY
         STORE .F. TO llSepClose, llVoidCheck, llDirDep
         STORE ''  TO lcsyskey, lcDMBatchJ, lcYear, lcPeriod
         STORE ''  TO lcIDChec, lcRunNo, lcMessage, lcOldRunYear, lcVersion
         STORE 0   TO lnSession, lnOldRunNo, lnPrtCount, lnProgress, lnMax
         STORE .NULL. TO loError, oGLMaint


* Make sure we're in the right datasession before opening the run - pws 10/12/09
         lnSession = SET('datasession')
         IF lnSession # lnDataSession
            SET DATASESSION TO (lnDataSession)
         ENDIF

         llSepClose  = .T.
         llVoidCheck = .F.
         lcRunNo     = lcRunYear + PADL(TRANSFORM(lnRunNo), 3, '0')

* Create the GLMaint object
         oGLMaint = CREATEOBJECT('glmaint')

         swselect('sysctl', .T.)
*  Find any later closings, and don't let them proceed - BH 03/25/2008
         LOCATE FOR ALLT(crunyear) + PADL(ALLT(STR(nrunno)), 3, '0') > ALLT(lcRunYear) + PADL(ALLT(STR(lnRunNo)), 3, '0') ;
            AND cgroup = lcGroup AND cTypeClose = 'R'  AND nrunno # 9999
         IF FOUND()
            MESSAGEBOX('Only the most recently closed run can be re-opened for a group.', 16, 'Unable To Reopen Run')
            lnReturn = 0
            EXIT
         ENDIF

* Look to see if the run has already been posted in another company (Partnership Mod)
         IF m.goApp.lPartnershipMod
            llReturn = THIS.oPartnerShip.CheckPosted(lcDMBatch)
            IF llReturn
               lnReturn = 0
               EXIT
            ENDIF
         ENDIF

         swselect('sysctl', .T.)
         LOCATE FOR crunyear = lcRunYear AND lDisbMan AND cTypeClose = 'R' AND nrunno = lnRunNo
         IF FOUND()

            lcVersion = '2024'

            IF EMPTY(cversion) OR cversion < lcVersion  &&  Don't let them re-open a run closed at a previous version - BH 11/29/06
               MESSAGEBOX('You cannot re-open this run, because it was closed with a previous version of the software.', 16, 'Unable To Reopen Run')
               lnReturn = 0
               EXIT
            ENDIF

            m.cidsysctl = cidsysctl
            m.dpostdate = dpostdate

            IF NOT oGLMaint.checkperiod(m.dpostdate, .T.)
               MESSAGEBOX('This revenue run cannot be reopened. ' + ;
                    'The fiscal year/period this run was posted to has been closed.', 16, 'Unable To Reopen Run')
               lnReturn = 0
               EXIT
            ENDIF
         ELSE
            MESSAGEBOX('Could not find the system control record for this closed run.', 16, 'Unable To Reopen Run')
            lnReturn = 0
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn          = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*  Look for pmts that have been received for this run
         swselect('suspense')
         SELECT  cownerid, ;
                 DTOS(hdate) AS mydate, ;
                 hdate, ;
                 SUM(nnetcheck) AS ntotal ;
             FROM suspense ;
             WHERE crunyear_in = lcRunYear ;
                 AND nrunno_in = lnRunNo ;
                 AND cgroup = lcGroup ;
                 AND crectype = 'P';
             INTO CURSOR tempsusp ;
             ORDER BY cownerid,;
                 mydate ;
             GROUP BY cownerid,;
                 mydate

         IF _TALLY > 0
            lcMessage = 'You cannot re-open this run, because there are owner net payments' + CHR(10) + ;
               'that have been received for deficits created in this run.' + CHR(10) + CHR(10) + ;
               'Delete the payments first and then try to open the run again.' + CHR(10) + CHR(10) + ;
               'Payments Found: ' + CHR(10) + CHR(10) + ;
               PADR('Owner ID  ', 15, ' ') + ' Date        ' + '        Amount' + CHR(10)
            SCAN
               lcMessage = lcMessage + tempsusp.cownerid + SPACE(5) + DTOC(tempsusp.hdate) + SPACE(.5) + TRANSFORM(tempsusp.ntotal, '$$$,$$$,$$$.99') + CHR(10)
            ENDSCAN
            MESSAGEBOX(lcMessage, 16, 'Unable To Reopen Run')
            lnReturn = 0
            EXIT
         ELSE
            swselect('disbhist')
            SELECT  cownerid, ;
                    DTOS(hdate) AS mydate, ;
                    hdate, ;
                    SUM(nnetcheck) AS ntotal ;
                FROM disbhist ;
                WHERE crunyear_in = lcRunYear ;
                    AND nrunno_in = lnRunNo ;
                    AND cgroup = lcGroup ;
                    AND crectype = 'P';
                INTO CURSOR tempsusp ;
                ORDER BY cownerid,;
                    mydate ;
                GROUP BY cownerid,;
                    mydate

            IF _TALLY > 0
               lcMessage = 'You cannot re-open this run, because there are owner net payments' + CHR(10) + ;
                  'that have been received for deficits created in this run.' + CHR(10) + CHR(10) + ;
                  'Delete the payments first and then try to open the run again.' + CHR(10) + CHR(10) + ;
                  'Payments Found: ' + CHR(10) + CHR(10) + ;
                  PADR('Owner ID  ', 15, ' ') + ' Date        ' + '        Amount' + CHR(10)
               SCAN
                  lcMessage = lcMessage + tempsusp.cownerid + SPACE(5) + DTOC(tempsusp.hdate) + SPACE(5) + TRANSFORM(tempsusp.ntotal, '$$$,$$$,$$$.99') + CHR(10)
               ENDSCAN
               MESSAGEBOX(lcMessage, 16, 'Unable To Reopen Run')
               lnReturn = 0
               EXIT
            ENDIF
         ENDIF

*  Make sure the period or year isn't closed
         IF NOT oGLMaint.checkperiod(m.dpostdate)
            MESSAGEBOX('Unable to reopen this revenue run. Either the fiscal year or period ' + ;
                 'represented by this date has been closed.', 16, 'Unable To Reopen Run')
            lnReturn = 0
            EXIT
         ENDIF

*  Make sure this is what the user wants to do.
         IF MESSAGEBOX('Are you sure run: ' + ALLT(STR(lnRunNo)) + '/' + lcRunYear + ' should be reopened?', 36, 'Confirmation Required') = 7
            MESSAGEBOX('Reopen of revenue run cancelled.', 48, 'Reopen of Run Cancelled')
            lnReturn = 0
            EXIT
         ENDIF

         IF m.goApp.lAMVersion
            lnMax = 17
         ELSE
            lnMax = 17
         ENDIF
         THIS.oprogress = THIS.omessage.progressbarex('Opening Revenue Run: ' + lcRunYear + '/' + PADL(TRANSFORM(lnRunNo), 3, '0') + ' GROUP: ' + lcGroup)
         THIS.oprogress.SetProgressRange(0, lnMax)

* Get the last run number for the group so we can put suspense back in this run
         lnOldRunNo   = getrunno(lcRunYear, .F., 'R', .F., lcGroup)
         lcOldRunYear = getrunno(lcRunYear, .F., 'R', .T., lcGroup)

*  Look to see if the checks have been printed.
*  If they have, ask if they should be voided.
*  If they haven't, delete them.
*
* STEP 1

         THIS.oprogress.SetProgressMessage('Removing checks from the check register...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         lnPrtCount = 0
         swselect('checks', .T.)
         LOCATE FOR cBatch == lcDMBatch
         IF FOUND()
            COUNT FOR lPrinted AND cBatch == lcDMBatch AND NOT "DIRDEP" $ ccheckno AND NOT LEFT(ALLTRIM(ccheckno), 1) = 'E' TO lnPrtCount
            IF lnPrtCount > 0
               IF MESSAGEBOX('There are checks from this run that have already been printed! ' + CHR(10) + CHR(10) + ;
                       'NOTE: ' + CHR(10) + ;
                       'If the checks from the run closing have been mailed out, DO NOT open this run!' + CHR(10) + CHR(10) + ;
                       'If any time has lapsed since they were mailed out there is a very good chance that the check amounts will NOT match the original '  + ;
                       'checks created when the run is closed again.' + CHR(10) + CHR(10) + ;
                       'Do you want to void them and continue?', 20, 'Confirmation Required') = 7
                  THIS.oprogress.CloseProgress()
                  lnReturn = 0
                  EXIT
               ELSE
                  THIS.oprogress.SetProgressMessage('Voiding checks from the check register...')
                  llVoidCheck = .T.
                  swselect('checks', .T.)
                  SCAN FOR cBatch = lcDMBatch
                     lcIDChec = cidchec
                     IF cidtype = 'V'
                        swselect('expense', .T.)
                        REPL cpaidbyck WITH '', cprdpaid WITH '', lclosed WITH .F. FOR cpaidbyck = lcIDChec
                     ENDIF
                     swselect('checks', .T.)
                     IF lPrinted
                        REPL nvoidamt WITH namount, ;
                           namount  WITH 0, ;
                           lVoid    WITH .T.,  ;
                           lCleared WITH .T.  &&  Mark as cleared, so they don't show on the recon screen - BH 02/15/2008
                     ELSE
                        DELETE NEXT 1
                     ENDIF
                  ENDSCAN
               ENDIF
            ELSE
               swselect('checks', .T.)
               SCAN FOR cBatch = lcDMBatch
                  lcIDChec = cidchec
                  IF cidtype = 'V'
                     swselect('expense', .T.)
                     REPL cpaidbyck WITH '', cprdpaid WITH '', lclosed WITH .F. FOR cpaidbyck = lcIDChec
                  ENDIF
                  swselect('checks', .T.)
                  DELETE NEXT 1
               ENDSCAN
            ENDIF
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


*
*  Remove the sysctl record
*
*  STEP 2

         THIS.oprogress.SetProgressMessage('Removing closing control record...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('sysctl', .T.)
         DELETE FROM sysctl WHERE cdmbatch == lcDMBatch

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*
*  Remove rounding records created this run
*
*  STEP 3

         THIS.oprogress.SetProgressMessage('Removing rounding information...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1

         swselect('roundtmp', .T.)
         DELETE FROM roundtmp WHERE cdmbatch == lcDMBatch

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


*
*  Remove records from the well history table
*
*  STEP 4

         THIS.oprogress.SetProgressMessage('Removing well history records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('wellhist', .T.)
         DELETE FROM wellhist WHERE nrunno == lnRunNo AND crunyear == lcRunYear AND crectype == 'R'

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*
*  Remove records from the one_man_tax table
*
*  STEP 5

         THIS.oprogress.SetProgressMessage('Removing one-man tax records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('one_man_tax', .T.)
         SET ORDER TO 0
         SCAN FOR nrunno == lnRunNo AND crunyear == lcRunYear AND NOT lManual
            DELE NEXT 1
         ENDSCAN

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*
*  Remove suspense entries
*
*  STEP 6

         THIS.oprogress.SetProgressMessage('Removing suspense records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('suspense', .T.)
         DELETE FROM suspense WHERE crunyear_in + PADL(TRANSFORM(nrunno_in), 3, '0') = lcRunNo AND NOT lManual

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*
* Place suspense entries brought into this run back
* into suspense.
*
*  STEP 7

         THIS.oprogress.SetProgressMessage('Moving released suspense back to suspense file...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('ownpcts', .T.)  &&  Get it ready for the seek below
         SET ORDER TO ciddisb
* Check for both rectypes of R and P so that payments get put back into suspense
         swselect('disbhist', .T.)
         SCAN FOR crunyear + PADL(TRANSFORM(nrunno), 3, '0') == lcRunNo AND INLIST(crectype, 'R', 'P') AND crunyear_in + PADL(TRANSFORM(nrunno_in), 3, '0') # lcRunNo AND nrunno_in # 0
            SCATTER MEMVAR

            m.nrunno   = disbhist.nrunno_in
            m.crunyear = disbhist.crunyear_in

            swselect('ownpcts', .T.)  &&  Get the interests to go along with this history entry - BH 06/19/2008
            IF SEEK(disbhist.ciddisb)
               SCATTER MEMVAR
               DELETE NEXT 1
            ELSE
               SCATTER MEMVAR BLANK  &&  Should never hit this, but if it does, it's best to insert zero interests, instead of the last record's interests
            ENDIF
            m.ciddisb = GetNextPK('DISBHIST')
            INSERT INTO suspense FROM MEMVAR
            SELE disbhist
            DELE NEXT 1
         ENDSCAN

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

* Reset the nrunno field. Not really necessary but to keep things straight...
         swselect('suspense')
         SCAN FOR crunyear + PADL(TRANSFORM(nrunno), 3, '0') = lcRunNo
            REPLACE crunyear WITH crunyear_in, ;
                    nrunno   WITH nrunno_in
         ENDSCAN

*
*  Remove records from the owner history table
*
*  STEP 8
         THIS.oprogress.SetProgressMessage('Removing owner history records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('disbhist', .T.)
         SELECT  ciddisb, ;
                 cownerid, ;
                 cWellID ;
             FROM disbhist ;
             WHERE crunyear + PADL(TRANSFORM(nrunno), 3, '0') = lcRunNo ;
                 AND crectype == 'R' ;
                 AND NOT lManual ;
             INTO CURSOR tempdisb NOFILTER
         swselect('disbhist')
         DELETE FROM disbhist WHERE crunyear + PADL(TRANSFORM(nrunno), 3, '0') = lcRunNo AND crectype == 'R' AND NOT lManual

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


* Remove owncpcts records from this run
*  STEP 9

         THIS.oprogress.SetProgressMessage('Removing Owner DOI history records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('ownpcts', .T.)
         DELETE FROM ownpcts WHERE ciddisb IN (SELECT ciddisb FROM tempdisb)
         USE IN tempdisb

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


*
*  Remove journal entries from the glmaster table
*
*  STEP 10

         IF m.goApp.lAMVersion
            THIS.oprogress.SetProgressMessage('Removing G/L journal records...')
            THIS.oprogress.UpdateProgress(lnProgress)
            lnProgress = lnProgress + 1
            swselect('glbatches', .T.)
            SELECT cBatch FROM glmaster WHERE cdmbatch = lcDMBatch INTO CURSOR tempbatch ORDER BY cBatch GROUP BY cBatch
            SELECT tempbatch
            DELETE FROM glbatches WHERE cBatch IN (SELECT cBatch FROM tempbatch)
            swselect('glmaster', .T.)
            DELETE FROM glmaster WHERE cBatch IN (SELECT cBatch FROM tempbatch)
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


*
*  Reset the flag in the income table
*
*  STEP 11

         THIS.oprogress.SetProgressMessage('Resetting closed flag in the well income records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('income', .T.)
         UPDATE  income ;
             SET nrunno   = 0, ;
                 crunyear = '', ;
                 cacctprd = '', ;
                 cacctyear = '', ;
                 lclosed   = .F. ;
             WHERE nrunno == lnRunNo ;
                 AND crunyear == lcRunYear

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


*
*  Reset the flag in the expense table
*
*  STEP 12

         THIS.oprogress.SetProgressMessage('Resetting closed flag in the well expense records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('expense', .T.)
         SCAN FOR nRunNoRev == lnRunNo AND cRunYearRev == lcRunYear
            REPL lclosed     WITH .F., ;
               cacctyear   WITH '', ;
               cacctprd    WITH '', ;
               cRunYearRev WITH '', ;
               nRunNoRev   WITH 0
            IF crunyearjib == '1900'
               REPL nrunnojib   WITH 0, ;
                  crunyearjib WITH ''
            ENDIF
            IF lAPTran = .F.
* Blank out check key if it's not an AP transaction -
* avoids bogus entries on expenses if the run closing hangs up
* Also check for the check being created by this closing. If not, don't clear cpaidbyck
               swselect('checks')
               LOCATE FOR cidchec = expense.cpaidbyck AND cBatch = lcDMBatch
               IF FOUND()
                  swselect('expense')
                  REPLACE cpaidbyck WITH ''
               ENDIF
            ENDIF
         ENDSCAN

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


         swselect('wells')
         SET ORDER TO cWellID

* Look for expenses that have 1901 for the run year, but no JIB stuff on them -
* these are entries that are entirely JIB, but the revenue run got closed first,
* so they need re-set as well - BH 07/10/08
         swselect('expense', .T.)
         SCAN FOR nRunNoRev == lnRunNo AND cRunYearRev == '1901' AND nrunnojib = 0 AND crunyearjib = ''
            SELECT wells
            IF SEEK(expense.cWellID)
               IF wells.cgroup = lcGroup  &&  Make sure this is the same group we're dealing with
                  SELECT expense
                  REPLACE nRunNoRev WITH 0, cRunYearRev WITH ''
               ENDIF
            ENDIF
         ENDSCAN

*  STEP 13

         THIS.oprogress.SetProgressMessage('Resetting statement note records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('stmtnote', .T.)  &&  Re-set the statement notes
         SCAN FOR crunyear == lcRunYear AND nrunno == lnRunNo
            REPLACE nrunno WITH 0, crunyear WITH ''
         ENDSCAN

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


* Delete the run closing summary totals record
*  STEP 14

         THIS.oprogress.SetProgressMessage('Removing the run closing total record...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('runclose', .T.)
         DELETE FROM runclose WHERE cdmbatch == lcDMBatch

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*  Delete any direct-deposit entries from dirdep
*  STEP 15

         THIS.oprogress.SetProgressMessage('Removing direct deposit records...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         IF FILE(m.goApp.cdatafilepath + 'dirdep.dbf')
            llDirDep = .T.
            swselect('dirdep', .T.)
            DELETE FROM dirdep WHERE nrunno = lnRunNo AND crunyear = lcRunYear
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*  Delete plugging charges from the plugging fund
*  STEP 16

         THIS.oprogress.SetProgressMessage('Removing plugging fund charges...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         swselect('plugwellbal', .T.)
         DELETE FROM plugwellbal WHERE cdmbatch == lcDMBatch

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            lnReturn = 0
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

*  Delete entries from the partnership posting table
*  STEP 17

         THIS.oprogress.SetProgressMessage('Removing partnership postings...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1

         IF m.goApp.lPartnershipMod
            llReturn = THIS.oPartnerShip.DeleteEntries(lcDMBatch)

* Unmark settlement statement notes has having been used this run
            llReturn = THIS.oPartnerShip.UnMarkSettlementNotes(THIS.crunyear, THIS.nrunno)

            IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
               lnReturn = 0
               IF NOT m.goApp.CancelMsg()
                  THIS.lCanceled = .T.
                  EXIT
               ENDIF
            ENDIF

            IF NOT llReturn
               lnReturn = 0
               EXIT
            ENDIF
         ENDIF
* Step 18
* Remove any A/R Pmt Header and Detail records created for Suspense being applied against house gas
         lcRunNo = lcRunYear + '/' + PADL(TRANSFORM(lnRunNo), 3, '0')
         swselect('arpmtdet',.T.)
         swselect('arpmthdr',.T.)
         SCAN FOR cidchec = lcRunNo
            m.cBatch = cBatch
            SELECT arpmtdet
            DELETE FOR cBatch = m.cBatch
            SELECT arpmthdr
            DELETE NEXT 1
         ENDSCAN
*
*  Start a transaction to encompass all changes we're
*  about to make.  This is so they can be rolled back
*  if there's any problems
*
*  STEP 19

         THIS.oprogress.SetProgressMessage('Finalizing the files after opening the run...')
         THIS.oprogress.UpdateProgress(lnProgress)
         lnProgress = lnProgress + 1
         BEGIN TRANSACTION
         SELE checks
         llReturn = TABLEUPDATE(.T., .T.)
         IF llReturn AND m.goApp.lAMVersion
            SELE glmaster
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn AND m.goApp.lAMVersion
            SELE glbatches
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELE expense
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELE income
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELE wellhist
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELE sysctl
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELE disbhist
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT ownpcts
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELE suspense
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT runclose
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT stmtnote
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT one_man_tax
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn AND llDirDep AND FILE(m.goApp.cdatafilepath + 'dirdep.dbf')
            SELECT dirdep
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT plugwellbal
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT arpmthdr
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn
            SELECT arpmtdet
            llReturn = TABLEUPDATE(.T., .T.)
         ENDIF
         IF llReturn AND m.goApp.lPartnershipMod
            SELECT partnerpost
            llReturn = TABLEUPDATE(.T., .T.)
            IF llReturn
               SELECT partnerinterest
               llReturn = TABLEUPDATE(.T., .T.)
            ENDIF
         ENDIF
         IF llReturn = .F.
            ROLLBACK
         ELSE
            END TRANSACTION
         ENDIF
         THIS.oprogress.CloseProgress()
         THIS.oprogress = .NULL.
         IF NOT tlquiet
            IF llReturn
               MESSAGEBOX('Revenue run: ' + lcRunYear + '/' + ALLT(STR(lnRunNo)) + '/' + lcGroup + ' is now open.', 64, 'Revenue Run Reopened')
            ELSE
               MESSAGEBOX('There was a problem reopening Revenue run: ' + lcRunYear + '/' + ALLT(STR(lnRunNo)) + '/' + lcGroup + ' is not opened.', 64, 'Revenue Run Open Problem')
            ENDIF
         ENDIF

         IF llVoidCheck
            lnReturn = 2
         ELSE
            lnReturn = 1
         ENDIF

      CATCH TO loError
         lnReturn = 0
         IF VARTYPE(oprogress) = 'O'
            THIS.oprogress.CloseProgress()
            THIS.oprogress = .NULL.
         ENDIF
         DO errorlog WITH 'Reopen_Rev_Run', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         ErrorMessageText()
      ENDTRY


      RETURN lnReturn

   ENDPROC

*-- Posts the disbursement manager journal entries to the G/L Master file.
*********************************
   PROCEDURE PostJourn
******************************
      LOCAL tcYear, tcPeriod, tdCheckDate, tcGroup, tdPostDate, oprogress
      LOCAL lnMax, lnCount, lnTotal, lcName, lnJIBInv, m.cCustName, lcDMBatch, llSepClose
      LOCAL lcRevClear, lcSuspense, m.cDisbAcct, m.cVendComp, m.cGathAcct, m.cBackWith
      LOCAL llIntegComp, lcAPAcct, lcAcctYear, lcAcctMonth, lcIDChec, llExpSum, lcSuspType
      LOCAL llRound, llRoundIt, m.cPlugAcct
      LOCAL lDirGasPurch, lDirOilPurch, lcDMExp, lcDeptNo, lcExpClear, lcOwnerID, llJibNet, llNoPostDM
      LOCAL llReturn, lnAmount, lnDefSwitch, lnExpense, lnIncome, lnMinSwitch, lnOwner, lnOwns, lnVendor
      LOCAL cBackWith, cBatch, cCRAcctV, cCustName, cDRAcctV, cDefAcct, cDisbAcct, cGathAcct, cID
      LOCAL cMinAcct, cownerid, cTaxAcct1, cTaxAcct2, cTaxAcct3, cTaxAcct4, cUnitNo, cVendComp
      LOCAL cVendName, cWellID, ccatcode, ccateg, cexpclass, cidchec, cownname, csusptype, nCompress
      LOCAL nGather, namount, ntotal, tdCompanyPost, loError

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


         lnMax       = 0
         lnCount     = 0
         lnTotal     = 0
         lcName      = 'Owner'
         m.cCustName = ' '
         lcIDChec    = ''
         lcSuspType  = ''

         lcAcctYear  = STR(YEAR(THIS.dpostdate), 4)
         lcAcctMonth = PADL(ALLTRIM(STR(MONTH(THIS.dpostdate), 2)), 2, '0')

*  Set the posting dates
         IF THIS.lAdvPosting = .T.
            tdCompanyPost = THIS.dCompanyShare
            tdPostDate    = THIS.dCheckDate
         ELSE
            tdCompanyPost = THIS.dCheckDate
            tdPostDate    = THIS.dCheckDate
         ENDIF

*  Plug the DM batch number into glmaint so that each
*  batch created can be traced to this closing
         THIS.ogl.DMBatch  = THIS.cdmbatch
         THIS.ogl.cSource  = 'DM'
         THIS.ogl.nDebits  = 0
         THIS.ogl.nCredits = 0
         THIS.ogl.dGLDate  = THIS.dpostdate

*  Get the suspense account from glopt

            swselect('glopt')
            lcSuspense = cSuspense
            IF EMPTY(lcSuspense)
               lcSuspense = '999999'
            ENDIF
            lcRevClear = crevclear
            lcExpClear = cexpclear
          
         llNoPostDM = lDMNoPost

*  Get the A/P account
         IF NOT USED('apopt')
            USE apopt IN 0
         ENDIF
         swselect('apopt')
         lcAPAcct = capacct

* Set the order for wells
         swselect('wells')
         SET ORDER TO cWellID

*  Set up the parameters used by processing in this method
         tcYear   = THIS.crunyear
         tcPeriod = THIS.cperiod
         tcGroup  = THIS.cgroup


*   Get Disbursement Checking Acct Number


            m.cDisbAcct = THIS.oOptions.cDisbAcct
            IF EMPTY(ALLT(m.cDisbAcct))
               m.cDisbAcct = lcSuspense
            ENDIF

            m.cVendComp = THIS.oOptions.cVendComp

            m.cGathAcct = THIS.oOptions.cGathAcct
            IF EMPTY(ALLT(m.cGathAcct))
               m.cGathAcct = lcSuspense
            ENDIF
            m.cBackWith = THIS.oOptions.cBackAcct
            IF EMPTY(ALLT(m.cBackWith))
               m.cBackWith = lcSuspense
            ENDIF
            m.cTaxAcct1  = THIS.oOptions.cTaxAcct1
            IF EMPTY(ALLT(m.cTaxAcct1))
               m.cTaxAcct1 = lcSuspense
            ENDIF
            m.cTaxAcct2  = THIS.oOptions.cTaxAcct2
            IF EMPTY(ALLT(m.cTaxAcct2))
               m.cTaxAcct2 = lcSuspense
            ENDIF
            m.cTaxAcct3  = THIS.oOptions.cTaxAcct3
            IF EMPTY(ALLT(m.cTaxAcct3))
               m.cTaxAcct3 = lcSuspense
            ENDIF
            m.cTaxAcct4 = THIS.oOptions.cTaxAcct4
            IF EMPTY(ALLT(m.cTaxAcct4))
               m.cTaxAcct4 = lcSuspense
            ENDIF
            m.cDefAcct  = THIS.oOptions.cDefAcct
            IF EMPTY(ALLT(m.cDefAcct))
               m.cDefAcct = lcSuspense
            ENDIF
            m.cMinAcct  = THIS.oOptions.cMinAcct
            IF EMPTY((m.cMinAcct))
               m.cMinAcct = lcSuspense
            ENDIF
            lcDMExp = THIS.oOptions.cFixedAcct
            IF EMPTY(lcDMExp)
               lcDMExp = lcAPAcct
            ENDIF
            m.cPlugAcct = THIS.oOptions.cPlugAcct
            IF EMPTY(m.cPlugAcct)
               m.cPlugAcct = lcSuspense
            ENDIF

          
 
            lcDeptNo         = THIS.oOptions.cdeptno

         llExpSum    = THIS.oOptions.lexpsum

         llJibNet    = .T.
         IF TYPE('m.goApp') = 'O'
* Turn off net jib processing for disb mgr
            IF m.goApp.ldmpro
               llJibNet = .F.
* Don't create journal entries for stand-alone disb mgr
               llNoPostDM = .T.
            ENDIF
         ENDIF
         llSepClose  = .T.

*   Check to see if vendor compression & gathering is to be posted
         llIntegComp = .F.

         IF NOT EMPTY(ALLT(m.cVendComp))
            swselect('vendor')
            SET ORDER TO cvendorid
            IF SEEK(m.cVendComp)
               IF lIntegGL
                  llIntegComp = .T.
               ENDIF
            ENDIF
         ENDIF

*   Post compression and gathering
         THIS.ogl.cBatch = GetNextPK('BATCH')

         IF m.goApp.lAMVersion
            SELE wellwork
            SCAN FOR nCompress # 0 OR nGather # 0
               SELECT wells
               IF SEEK(wellwork.cWellID)
                  IF NOT wells.lcompress AND NOT wells.lGather
                     LOOP
                  ENDIF
               ELSE
                  LOOP
               ENDIF
               SELECT wellwork
               m.cWellID           = cWellID
               m.nCompress         = nCompress
               m.nGather           = nGather
               THIS.ogl.cReference = 'Period: ' + THIS.cyear + '/' + THIS.cperiod + '/' + THIS.cgroup
               THIS.ogl.cyear      = THIS.cyear
               THIS.ogl.cperiod    = THIS.cperiod
               THIS.ogl.dCheckDate = THIS.dacctdate
               IF llIntegComp
                  THIS.ogl.dGLDate  = tdCompanyPost
               ELSE
                  THIS.ogl.dGLDate  = tdPostDate
               ENDIF
               THIS.ogl.cDesc   = 'Compression/Gathering'
               THIS.ogl.cID     = ''
               THIS.ogl.cidtype = ''
               THIS.ogl.cSource = 'DM'

* Get Expense Class for Comp/Gath
               swselect('expcat')
               LOCATE FOR ccatcode = 'COMP'
               IF FOUND()
                  lcExpClass = cexpclass
               ELSE
                  LOCATE FOR ccatcode = 'GATH'
                  IF FOUND()
                     lcExpClass = cexpclass
                  ELSE
                     lcExpClass = 'G'
                  ENDIF
               ENDIF


               lnCompGathAmount = swNetExp(m.nCompress + m.nGather, m.cWellID, .T., lcExpClass, 'B')

               THIS.ogl.cAcctNo    = m.cGathAcct
               THIS.ogl.cgroup     = THIS.cgroup
               THIS.ogl.cEntryType = 'C'
               THIS.ogl.cUnitNo    = m.cWellID
               THIS.ogl.namount    = lnCompGathAmount * -1
               THIS.ogl.UpdateBatch()

               THIS.ogl.cAcctNo = THIS.cexpclear
               THIS.ogl.namount = lnCompGathAmount
               THIS.ogl.UpdateBatch()
            ENDSCAN
         ENDIF


*   Create Investor Checks
         llReturn = THIS.ownerchks()
         IF NOT llReturn
            EXIT
         ENDIF

*   Create Vendor Checks
         llReturn = THIS.vendorchks()
         IF NOT llReturn
            EXIT
         ENDIF

* Only call directdeposit if the module is available
         IF m.goApp.lDirDMDep
            llReturn = THIS.directdeposit()
            IF NOT llReturn
               EXIT
            ENDIF
         ENDIF


*   Post owner amounts to G/L
         swselect('wells')
         SET ORDER TO cWellID

*  Clear out balances
         THIS.ogl.nDebits  = 0
         THIS.ogl.nCredits = 0

         IF NOT llNoPostDM
* Get a cursor of owners to be posted from invtmp
            SELECT  cownerid, ;
                    SUM(nnetcheck) AS ntotal ;
                FROM invtmp WITH (BUFFERING = .T.) ;  && This is a cursor. Don't think we need the buffering command
                ORDER BY cownerid ;
                GROUP BY cownerid ;
                INTO CURSOR tmpowners READWRITE
            lnMax = _TALLY
            IF lnMax > 0
               SELE tmpowners
               INDEX ON cownerid TAG owner
            ENDIF
            lnCount = 1

* Get the suspense types before this run so we know how to post the owners
            THIS.osuspense.GetLastType(.F., .T., THIS.cgroup, .T.)

            swselect('investor')
            SET ORDER TO cownerid
            STORE .F. TO llFedWire, llDirectDep
            SELECT tmpowners
            SCAN
               m.cownerid = cownerid
               m.ntotal   = ntotal
               THIS.oprogress.SetProgressMessage('Posting Owner Checks to General Ledger...' + m.cownerid)
               THIS.ogl.cBatch = GetNextPK('BATCH')
               swselect('investor')
               IF SEEK(m.cownerid)
                  m.cownname     = cownname
                  THIS.ogl.cID   = m.cownerid
                  THIS.ogl.cDesc = m.cownname

                  llDirectDep = ldirectdep
                  llFedWire   = lFedwire

                  IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
                     llReturn          = .F.
                     IF NOT m.goApp.CancelMsg()
                        THIS.lCanceled = .T.
                        EXIT
                     ENDIF
                  ENDIF

* Don't post "Dummy" owner amounts
                  IF investor.ldummy
                     LOOP
                  ENDIF
* Don't post owners that are transfered to G/L here.
                  IF investor.lIntegGL
                     LOOP
                  ENDIF
               ELSE
                  LOOP
               ENDIF

               m.cidchec          = ' '
               THIS.ogl.dpostdate = tdPostDate

               SELECT invtmp
               SCAN FOR cownerid == m.cownerid AND (nIncome # 0 OR nexpense # 0 OR nsevtaxes # 0 OR nnetcheck # 0) AND csusptype = ' '
                  SCATTER MEMVAR
                  lcIDChec = m.cidchec

                  swselect('wells')
                  IF SEEK(m.cWellID)
                     SCATTER FIELDS LIKE lSev* MEMVAR
                     m.lDirOilPurch = lDirOilPurch
                     m.lDirGasPurch = lDirGasPurch
                  ELSE
                     m.lDirOilPurch = .F.
                     m.lDirGasPurch = .F.
                  ENDIF

                  THIS.ogl.cUnitNo    = m.cWellID

* Don't post prior suspense in this section
                  IF NOT INLIST(m.csusptype, 'D', 'M', 'I', 'H')
                     lnIncome   = m.nIncome
*  Remove direct paid amounts
                     DO CASE
                        CASE m.cdirect = 'O'
                           lnIncome = lnIncome - m.noilrev
                        CASE m.cdirect = 'G'
                           lnIncome = lnIncome - m.ngasrev
                        CASE m.cdirect = 'B'
                           lnIncome = lnIncome - m.noilrev - m.ngasrev
                     ENDCASE

* Post the Revenue
                     IF NOT m.lflat OR m.nflatrate = 0
                        THIS.ogl.cAcctNo    = THIS.crevclear
                        THIS.ogl.namount    = lnIncome
                        THIS.ogl.cID        = m.cownerid
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cDesc      = m.cownname
                        THIS.ogl.cidchec    = m.cidchec
                        THIS.ogl.cReference = 'Revenue'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        THIS.ogl.cAcctNo    = THIS.crevclear
                        THIS.ogl.namount    = lnIncome
                        THIS.ogl.cID        = m.cownerid
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cDesc      = m.cownname
                        THIS.ogl.cidchec    = m.cidchec
                        THIS.ogl.cReference = 'Flat Revenue'
                        THIS.ogl.UpdateBatch()
                     ENDIF

* Post the sev taxes
                     IF m.noiltax1 # 0
                        IF NOT m.lsev1o
                           IF NOT m.lDirOilPurch
                              THIS.ogl.cAcctNo    = m.cTaxAcct1
                              THIS.ogl.namount    = m.noiltax1 * -1
                              THIS.ogl.cReference = 'Oil Tax 1'
                              THIS.ogl.UpdateBatch()
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'O')
                                 THIS.ogl.cAcctNo    = m.cTaxAcct1
                                 THIS.ogl.namount    = m.noiltax1 * -1
                                 THIS.ogl.cReference = 'Oil Tax 1'
                                 THIS.ogl.UpdateBatch()
                              ENDIF
                           ENDIF
                        ELSE
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.noiltax1 * -1
                           THIS.ogl.cReference = 'Oil Tax 1'
                           THIS.ogl.UpdateBatch()
                        ENDIF
                     ENDIF

                     IF m.ngastax1 # 0
                        IF NOT m.lsev1g
                           IF NOT m.lDirGasPurch
                              THIS.ogl.cAcctNo    = m.cTaxAcct1
                              THIS.ogl.namount    = m.ngastax1 * -1
                              THIS.ogl.cReference = 'Gas Tax 1'
                              THIS.ogl.UpdateBatch()
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'G')
                                 THIS.ogl.cAcctNo    = m.cTaxAcct1
                                 THIS.ogl.namount    = m.ngastax1 * -1
                                 THIS.ogl.cReference = 'Gas Tax 1'
                                 THIS.ogl.UpdateBatch()
                              ENDIF
                           ENDIF
                        ELSE
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.ngastax1 * -1
                           THIS.ogl.cReference = 'Gas Tax 1'
                           THIS.ogl.UpdateBatch()
                        ENDIF
                     ENDIF

                     IF m.nOthTax1 # 0
                        IF NOT m.lsev1p
                           THIS.ogl.cAcctNo = m.cTaxAcct1
                        ELSE
                           THIS.ogl.cAcctNo = THIS.crevclear
                        ENDIF
                        THIS.ogl.namount    = m.nOthTax1 * -1
                        THIS.ogl.cReference = 'Other Tax 1'
                        THIS.ogl.UpdateBatch()
                     ENDIF

                     IF m.noiltax2 # 0
                        IF NOT m.lsev2o
                           THIS.ogl.cAcctNo    = m.cTaxAcct2
                           THIS.ogl.namount    = m.noiltax2 * -1
                           THIS.ogl.cReference = 'Oil Tax 2'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.noiltax2 * -1
                              THIS.ogl.cReference = 'Oil Tax 2'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ENDIF

                     IF m.ngastax2 # 0
                        IF NOT m.lsev2g
                           THIS.ogl.cAcctNo    = m.cTaxAcct2
                           THIS.ogl.namount    = m.ngastax2 * -1
                           THIS.ogl.cReference = 'Gas Tax 2'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.ngastax2 * -1
                              THIS.ogl.cReference = 'Gas Tax 2'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ENDIF

                     IF m.nOthTax2 # 0
                        IF NOT m.lsev2p
                           THIS.ogl.cAcctNo = m.cTaxAcct2
                        ELSE
                           THIS.ogl.cAcctNo = THIS.crevclear
                        ENDIF
                        THIS.ogl.namount    = m.nOthTax2 * -1
                        THIS.ogl.cReference = 'Other Tax 2'
                        THIS.ogl.UpdateBatch()
                     ENDIF

                     IF m.noiltax3 # 0
                        IF NOT m.lsev3o
                           THIS.ogl.cAcctNo    = m.cTaxAcct3
                           THIS.ogl.namount    = m.noiltax3 * -1
                           THIS.ogl.cReference = 'Oil Tax 3'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.noiltax3 * -1
                              THIS.ogl.cReference = 'Oil Tax 3'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ENDIF

                     IF m.ngastax3 # 0
                        IF NOT m.lsev3g
                           THIS.ogl.cAcctNo    = m.cTaxAcct3
                           THIS.ogl.namount    = m.ngastax3 * -1
                           THIS.ogl.cReference = 'Gas Tax 3'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.ngastax3 * -1
                              THIS.ogl.cReference = 'Gas Tax 3'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ENDIF

                     IF m.nOthTax3 # 0
                        IF NOT m.lsev3p
                           THIS.ogl.cAcctNo = m.cTaxAcct3
                        ELSE
                           THIS.ogl.cAcctNo = THIS.crevclear
                        ENDIF
                        THIS.ogl.namount    = m.nOthTax3 * -1
                        THIS.ogl.cReference = 'Other Tax 3'
                        THIS.ogl.UpdateBatch()
                     ENDIF

                     IF m.noiltax4 # 0
                        IF NOT m.lsev4o
                           IF NOT m.lDirOilPurch
                              THIS.ogl.cAcctNo    = m.cTaxAcct4
                              THIS.ogl.namount    = m.noiltax4 * -1
                              THIS.ogl.cReference = 'Oil Tax 4'
                              THIS.ogl.UpdateBatch()
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'O')
                                 THIS.ogl.cAcctNo    = m.cTaxAcct4
                                 THIS.ogl.namount    = m.noiltax4 * -1
                                 THIS.ogl.cReference = 'Oil Tax 4'
                                 THIS.ogl.UpdateBatch()
                              ENDIF
                           ENDIF
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.noiltax4 * -1
                              THIS.ogl.cReference = 'Oil Tax 4'
                              THIS.ogl.UpdateBatch()
                           ELSE
                              IF NOT m.lDirOilPurch
                                 THIS.ogl.cAcctNo    = THIS.crevclear
                                 THIS.ogl.namount    = m.noiltax4 * -1
                                 THIS.ogl.cReference = 'Oil Tax 4'
                                 THIS.ogl.UpdateBatch()
                              ENDIF
                           ENDIF
                        ENDIF
                     ENDIF

                     IF m.ngastax4 # 0
                        IF NOT m.lsev4g
                           THIS.ogl.cAcctNo    = m.cTaxAcct4
                           THIS.ogl.namount    = m.ngastax4 * -1
                           THIS.ogl.cReference = 'Gas Tax 4'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.ngastax4 * -1
                              THIS.ogl.cReference = 'Gas Tax 4'
                              THIS.ogl.UpdateBatch()
                           ELSE
                              IF NOT m.lDirGasPurch
                                 THIS.ogl.cAcctNo    = THIS.crevclear
                                 THIS.ogl.namount    = m.ngastax4 * -1
                                 THIS.ogl.cReference = 'Gas Tax 4'
                                 THIS.ogl.UpdateBatch()
                              ENDIF
                           ENDIF
                        ENDIF
                     ENDIF

                     IF m.nOthTax4 # 0
                        IF NOT m.lsev4p
                           THIS.ogl.cAcctNo = m.cTaxAcct4
                        ELSE
                           THIS.ogl.cAcctNo = THIS.crevclear
                        ENDIF
                        THIS.ogl.namount    = m.nOthTax4 * -1
                        THIS.ogl.cReference = 'Other Tax 4'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*
*  Post compression and gathering
*
                     IF m.nCompress # 0
                        THIS.ogl.cAcctNo    = THIS.cexpclear
                        THIS.ogl.namount    = m.nCompress * -1
                        THIS.ogl.cReference = 'Compression'
                        THIS.ogl.UpdateBatch()
                     ENDIF

                     IF m.nGather # 0
                        THIS.ogl.cAcctNo    = THIS.cexpclear
                        THIS.ogl.namount    = m.nGather * -1
                        THIS.ogl.cReference = 'Gathering'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*
*  Post marketing expenses
*
                     IF m.nMKTGExp # 0
                        THIS.ogl.cAcctNo    = THIS.cexpclear
                        THIS.ogl.namount    = m.nMKTGExp * -1
                        THIS.ogl.cReference = 'Marketing Exp'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post the Expenses
                     lnExpense = m.nexpense + ;
                        m.ntotale1 + ;
                        m.ntotale2 + ;
                        m.ntotale3 + ;
                        m.ntotale4 + ;
                        m.ntotale5 + ;
                        m.ntotalea + ;
                        m.ntotaleb + ;
                        m.nPlugExp

                     IF lnExpense # 0
                        THIS.ogl.cAcctNo    = THIS.cexpclear
                        THIS.ogl.namount    = lnExpense * -1
                        THIS.ogl.cReference = 'Expenses'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Backup Withholding
                     IF m.nbackwith # 0
                        THIS.ogl.cAcctNo    = m.cBackWith
                        THIS.ogl.namount    = m.nbackwith * -1
                        THIS.ogl.cReference = 'Backup W/H'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Tax Withholding
                     IF m.ntaxwith # 0
                        THIS.ogl.cAcctNo    = m.cBackWith
                        THIS.ogl.namount    = m.ntaxwith * -1
                        THIS.ogl.cReference = 'Tax W/H'
                        THIS.ogl.UpdateBatch()
                     ENDIF
                  ENDIF
               ENDSCAN

* Post prior suspense

               SELECT invtmp
               SCAN FOR cownerid = m.cownerid AND (nIncome # 0 OR nexpense # 0 OR nsevtaxes # 0 OR nnetcheck # 0) AND csusptype <> ' '
                  SCATTER MEMVAR
                  THIS.ogl.cUnitNo   = m.cWellID

                  SELECT curLastSuspType
                  LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID AND ctypeinv == m.ctypeinv
                  IF FOUND()
                     m.csusptype = csusptype
*  Post Prior Period Deficits
                     IF m.csusptype = 'D'
                        THIS.ogl.cAcctNo    = m.cDefAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Prior Def'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Prior Period Minimums
                     IF m.csusptype = 'M'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Prior Min'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Interest on Hold being released
                     IF m.csusptype = 'I'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Int On Hold Rel'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Owner on Hold being released
                     IF m.csusptype = 'H'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Owner On Hold Rel'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Quarterly Owner being released
                     IF m.csusptype = 'Q'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Quarterly Rel'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Semi-Annual Owner being released
                     IF m.csusptype = 'S'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Semi-Annual Rel'
                        THIS.ogl.UpdateBatch()
                     ENDIF

*  Post Annual Owner being released
                     IF m.csusptype = 'A'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.nnetcheck
                        THIS.ogl.cReference = 'Annual Rel'
                        THIS.ogl.UpdateBatch()
                     ENDIF
                  ELSE
*  Not found in curLastSuspType.
*  They still have a suspense entry that needs posted, so treat it as a deficit.
*  Most commonly, this results from receiving a payment for an owner with no balance,
*  so nothing returns for them in curLastSuspType. - BH 11/28/12
                     THIS.ogl.cAcctNo    = m.cDefAcct
                     THIS.ogl.namount    = m.nnetcheck
                     THIS.ogl.cReference = 'Prior Def'
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDSCAN

* Post Check Amount To Cash or the Fedwire clearing account
               IF llFedWire = .T.
                  IF NOT EMPTY(THIS.oOptions.cFedwire)
                     THIS.ogl.cAcctNo    = THIS.oOptions.cFedwire
                  ELSE
                     THIS.ogl.cAcctNo    = m.cDisbAcct
                  ENDIF
               ELSE
                  THIS.ogl.cAcctNo    = m.cDisbAcct
               ENDIF
               THIS.ogl.namount    = m.ntotal * -1
               THIS.ogl.cUnitNo    = ''
               THIS.ogl.cReference = 'Check'
               THIS.ogl.UpdateBatch()

               lcIDChec = ''
            ENDSCAN

            THIS.ogl.cReference = 'Run: R' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + '/' + THIS.cgroup


*  Post owners going into suspense
            SET SAFETY OFF

* Get a cursor of owners to be posted from suspense
            SELECT  cownerid, ;
                    SUM(nnetcheck) AS ntotal ;
                FROM tsuspense ;
                WHERE nrunno_in = THIS.nrunno ;
                    AND crunyear_in = THIS.crunyear ;
                ORDER BY cownerid ;
                GROUP BY cownerid ;
                INTO CURSOR tmpowners READWRITE

            lnMax = _TALLY

            IF lnMax > 0
               SELECT tmpowners
               INDEX ON cownerid TAG owner
            ENDIF
            lnCount = 1

            SELECT tmpowners
            SCAN
               m.cownerid      = cownerid
               m.ntotal        = ntotal
               THIS.ogl.cBatch = GetNextPK('BATCH')
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(m.cownerid)
                  m.cownname     = cownname
                  THIS.ogl.cID   = m.cownerid
                  THIS.ogl.cDesc = m.cownname
* Don't post "Dummy" owner amounts
                  IF investor.ldummy
                     LOOP
                  ENDIF
* Don't post owners that are transfered to G/L here.
                  IF investor.lIntegGL
                     LOOP
                  ENDIF
               ELSE
                  LOOP
               ENDIF

               m.cidchec          = ' '
               THIS.ogl.dpostdate = tdPostDate

               SELECT tsuspense
               SCAN FOR cownerid = m.cownerid ;
                     AND (nIncome # 0 OR nexpense # 0 OR nsevtaxes # 0 OR nnetcheck # 0) ;
                     AND nrunno_in = THIS.nrunno AND crunyear_in == THIS.crunyear ;
                     AND crectype = 'R'
                  SCATTER MEMVAR
                  lcIDChec = m.cidchec
                  THIS.oprogress.SetProgressMessage('Posting Owner Suspense to General Ledger...' + m.cownerid + ' Well: ' + m.cWellID)
                  swselect('wells')
                  IF SEEK(m.cWellID)
                     SCATTER FIELDS LIKE lSev* MEMVAR
                     m.lDirOilPurch = lDirOilPurch
                     m.lDirGasPurch = lDirGasPurch
                  ELSE
                     m.lDirOilPurch = .F.
                     m.lDirGasPurch = .F.
                  ENDIF

                  THIS.ogl.cUnitNo    = m.cWellID

                  lnIncome   = m.nIncome
*  Remove direct paid amounts
                  DO CASE
                     CASE m.cdirect = 'O'
                        lnIncome = lnIncome - m.noilrev
                     CASE m.cdirect = 'G'
                        lnIncome = lnIncome - m.ngasrev
                     CASE m.cdirect = 'B'
                        lnIncome = lnIncome - m.noilrev - m.ngasrev
                  ENDCASE


* Post the Revenue
                  IF NOT m.lflat AND lnIncome # 0
                     THIS.ogl.cAcctNo    = THIS.crevclear
                     THIS.ogl.namount    = lnIncome
                     THIS.ogl.cID        = m.cownerid
                     THIS.ogl.cUnitNo    = m.cWellID
                     THIS.ogl.cDesc      = m.cownname
                     THIS.ogl.cidchec    = m.cidchec
                     THIS.ogl.cReference = 'Revenue'
                     THIS.ogl.UpdateBatch()
                  ELSE
                     IF m.nflatrate # 0
                        THIS.ogl.cAcctNo    = THIS.crevclear
                        THIS.ogl.namount    = m.nflatrate
                        THIS.ogl.cID        = m.cownerid
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cDesc      = m.cownname
                        THIS.ogl.cidchec    = m.cidchec
                        THIS.ogl.cReference = 'Flat Revenue'
                        THIS.ogl.UpdateBatch()
                     ENDIF
                  ENDIF

* Post the sev taxes
                  IF m.noiltax1 # 0
                     IF NOT m.lsev1o
                        IF NOT m.lDirOilPurch
                           THIS.ogl.cAcctNo    = m.cTaxAcct1
                           THIS.ogl.namount    = m.noiltax1 * -1
                           THIS.ogl.cReference = 'Oil Tax 1'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              THIS.ogl.cAcctNo    = m.cTaxAcct1
                              THIS.ogl.namount    = m.noiltax1 * -1
                              THIS.ogl.cReference = 'Oil Tax 1'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ELSE
                        THIS.ogl.cAcctNo    = THIS.crevclear
                        THIS.ogl.namount    = m.noiltax1 * -1
                        THIS.ogl.cReference = 'Oil Tax 1'
                        THIS.ogl.UpdateBatch()
                     ENDIF
                  ENDIF

                  IF m.ngastax1 # 0
                     IF NOT m.lsev1g
                        IF NOT m.lDirGasPurch
                           THIS.ogl.cAcctNo    = m.cTaxAcct1
                           THIS.ogl.namount    = m.ngastax1 * -1
                           THIS.ogl.cReference = 'Gas Tax 1'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              THIS.ogl.cAcctNo    = m.cTaxAcct1
                              THIS.ogl.namount    = m.ngastax1 * -1
                              THIS.ogl.cReference = 'Gas Tax 1'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ELSE
                        THIS.ogl.cAcctNo    = THIS.crevclear
                        THIS.ogl.namount    = m.ngastax1 * -1
                        THIS.ogl.cReference = 'Gas Tax 1'

                        THIS.ogl.UpdateBatch()
                     ENDIF
                  ENDIF

                  IF m.nOthTax1 # 0
                     IF NOT m.lsev1p
                        THIS.ogl.cAcctNo = m.cTaxAcct1
                     ELSE
                        THIS.ogl.cAcctNo = THIS.crevclear
                     ENDIF
                     THIS.ogl.namount    = m.nOthTax1 * -1
                     THIS.ogl.cReference = 'Other Tax 1'
                     THIS.ogl.UpdateBatch()
                  ENDIF

                  IF m.noiltax2 # 0
                     IF NOT m.lsev2o
                        THIS.ogl.cAcctNo    = m.cTaxAcct2
                        THIS.ogl.namount    = m.noiltax2 * -1
                        THIS.ogl.cReference = 'Oil Tax 2'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        IF NOT INLIST(m.cdirect, 'B', 'O')
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.noiltax2 * -1
                           THIS.ogl.cReference = 'Oil Tax 2'
                           THIS.ogl.UpdateBatch()
                        ENDIF
                     ENDIF
                  ENDIF

                  IF m.ngastax2 # 0
                     IF NOT m.lsev2g
                        THIS.ogl.cAcctNo    = m.cTaxAcct2
                        THIS.ogl.namount    = m.ngastax2 * -1
                        THIS.ogl.cReference = 'Gas Tax 2'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        IF NOT INLIST(m.cdirect, 'B', 'G')
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.ngastax2 * -1
                           THIS.ogl.cReference = 'Gas Tax 2'
                           THIS.ogl.UpdateBatch()
                        ENDIF
                     ENDIF
                  ENDIF

                  IF m.nOthTax2 # 0
                     IF NOT m.lsev2p
                        THIS.ogl.cAcctNo = m.cTaxAcct2
                     ELSE
                        THIS.ogl.cAcctNo = THIS.crevclear
                     ENDIF
                     THIS.ogl.namount    = m.nOthTax2 * -1
                     THIS.ogl.cReference = 'Other Tax 2'
                     THIS.ogl.UpdateBatch()
                  ENDIF

                  IF m.noiltax3 # 0
                     IF NOT m.lsev3o
                        THIS.ogl.cAcctNo    = m.cTaxAcct3
                        THIS.ogl.namount    = m.noiltax3 * -1
                        THIS.ogl.cReference = 'Oil Tax 3'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        IF NOT INLIST(m.cdirect, 'B', 'O')
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.noiltax3 * -1
                           THIS.ogl.cReference = 'Oil Tax 3'
                           THIS.ogl.UpdateBatch()
                        ENDIF
                     ENDIF
                  ENDIF

                  IF m.ngastax3 # 0
                     IF NOT m.lsev3g
                        THIS.ogl.cAcctNo    = m.cTaxAcct3
                        THIS.ogl.namount    = m.ngastax3 * -1
                        THIS.ogl.cReference = 'Gas Tax 3'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        IF NOT INLIST(m.cdirect, 'B', 'G')
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.ngastax3 * -1
                           THIS.ogl.cReference = 'Gas Tax 3'
                           THIS.ogl.UpdateBatch()
                        ENDIF
                     ENDIF
                  ENDIF

                  IF m.nOthTax3 # 0
                     IF NOT m.lsev3p
                        THIS.ogl.cAcctNo = m.cTaxAcct3
                     ELSE
                        THIS.ogl.cAcctNo = THIS.crevclear
                     ENDIF
                     THIS.ogl.namount    = m.nOthTax3 * -1
                     THIS.ogl.cReference = 'Other Tax 3'
                     THIS.ogl.UpdateBatch()
                  ENDIF

                  IF m.noiltax4 # 0
                     IF NOT m.lsev4o
                        THIS.ogl.cAcctNo    = m.cTaxAcct4
                        THIS.ogl.namount    = m.noiltax4 * -1
                        THIS.ogl.cReference = 'Oil Tax 4'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        IF NOT INLIST(m.cdirect, 'B', 'O')
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.noiltax4 * -1
                           THIS.ogl.cReference = 'Oil Tax 4'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT m.lDirOilPurch
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.noiltax4 * -1
                              THIS.ogl.cReference = 'Oil Tax 4'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ENDIF
                  ENDIF

                  IF m.ngastax4 # 0
                     IF NOT m.lsev4g
                        THIS.ogl.cAcctNo    = m.cTaxAcct4
                        THIS.ogl.namount    = m.ngastax4 * -1
                        THIS.ogl.cReference = 'Gas Tax 4'
                        THIS.ogl.UpdateBatch()
                     ELSE
                        IF NOT INLIST(m.cdirect, 'B', 'G')
                           THIS.ogl.cAcctNo    = THIS.crevclear
                           THIS.ogl.namount    = m.ngastax4 * -1
                           THIS.ogl.cReference = 'Gas Tax 4'
                           THIS.ogl.UpdateBatch()
                        ELSE
                           IF NOT m.lDirGasPurch
                              THIS.ogl.cAcctNo    = THIS.crevclear
                              THIS.ogl.namount    = m.ngastax4 * -1
                              THIS.ogl.cReference = 'Gas Tax 4'
                              THIS.ogl.UpdateBatch()
                           ENDIF
                        ENDIF
                     ENDIF
                  ENDIF

                  IF m.nOthTax4 # 0
                     IF NOT m.lsev4p
                        THIS.ogl.cAcctNo = m.cTaxAcct4
                     ELSE
                        THIS.ogl.cAcctNo = THIS.crevclear
                     ENDIF
                     THIS.ogl.namount    = m.nOthTax4 * -1
                     THIS.ogl.cReference = 'Other Tax 4'
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post compression and gathering
                  IF m.nCompress # 0
                     THIS.ogl.cAcctNo    = THIS.cexpclear
                     THIS.ogl.namount    = m.nCompress * -1
                     THIS.ogl.cReference = 'Compression'
                     THIS.ogl.UpdateBatch()
                  ENDIF

                  IF m.nGather # 0
                     THIS.ogl.cAcctNo    = THIS.cexpclear
                     THIS.ogl.namount    = m.nGather * -1
                     THIS.ogl.cReference = 'Gathering'
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post marketing expenses
                  IF m.nMKTGExp # 0
                     THIS.ogl.cAcctNo    = THIS.cexpclear
                     THIS.ogl.namount    = m.nMKTGExp * -1
                     THIS.ogl.cReference = 'Marketing Exp'
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post the Expenses
                  lnExpense = m.nexpense + ;
                     m.ntotale1 + ;
                     m.ntotale2 + ;
                     m.ntotale3 + ;
                     m.ntotale4 + ;
                     m.ntotale5 + ;
                     m.ntotalea + ;
                     m.ntotaleb + ;
                     m.nPlugExp

                  IF lnExpense # 0
                     THIS.ogl.cAcctNo    = THIS.cexpclear
                     THIS.ogl.namount    = lnExpense * -1
                     THIS.ogl.cReference = 'Expenses'
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post Backup Withholding
                  IF m.nbackwith # 0
                     THIS.ogl.cAcctNo    = m.cBackWith
                     THIS.ogl.namount    = m.nbackwith * -1
                     THIS.ogl.cReference = 'Backup W/H'
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post Tax Withholding
                  IF m.ntaxwith # 0
                     THIS.ogl.cAcctNo    = m.cBackWith
                     THIS.ogl.namount    = m.ntaxwith * -1
                     THIS.ogl.cReference = 'Tax W/H'
                     THIS.ogl.UpdateBatch()
                  ENDIF

                  lcSuspType = m.csusptype
                  m.namount  = m.nnetcheck

*  Check for minimum amount checks and post
                  DO CASE
                     CASE INLIST(lcSuspType, 'M', 'Q', 'S', 'A')
                        THIS.ogl.cAcctNo = m.cMinAcct
                        THIS.ogl.namount = m.namount * -1
                        THIS.ogl.cUnitNo = m.cWellID
                        DO CASE
                           CASE lcSuspType  = 'M'
                              THIS.ogl.cReference = 'Under Min'
                           CASE lcSuspType  = 'Q'
                              THIS.ogl.cReference = 'Qtrly Freq'
                           CASE lcSuspType  = 'S'
                              THIS.ogl.cReference = 'Semi-Annual Freq'
                           CASE lcSuspType  = 'A'
                              THIS.ogl.cReference = 'Annual Freq'
                        ENDCASE
                        THIS.ogl.UpdateBatch()

                     CASE lcSuspType = 'D'
                        THIS.ogl.cAcctNo    = m.cDefAcct
                        THIS.ogl.namount    = m.namount * -1
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cReference = 'Cur Deficit'
                        THIS.ogl.UpdateBatch()

                     CASE lcSuspType = 'I'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.namount * -1
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cReference = 'Int on Hold'
                        THIS.ogl.UpdateBatch()

                     CASE lcSuspType = 'H'
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.namount * -1
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cReference = 'Owner on Hold'
                        THIS.ogl.UpdateBatch()
                  ENDCASE

               ENDSCAN

               lcIDChec = ''
            ENDSCAN

* Post suspense amounts that are moving between deficit
* and minimum suspense.
            STORE 0 TO lnDefSwitch, lnMinSwitch

            IF NOT FILE('datafiles\noswitch.txt')
               IF NOT USED('baltransfer')
* Post amounts transfering between deficit and minimum
* Get the amount of deficits that transferred
                  THIS.oprogress.SetProgressMessage('Calculating amounts of suspense that switched between deficit and minimum')
                  THIS.oprogress.UpdateProgress(THIS.nprogress)
                  THIS.nprogress = THIS.nprogress + 1
                  lnDefSwitch    = THIS.osuspense.GetBalTransfer('D', .T.)

                  IF USED('baltransfer')
                     THIS.ogl.cDesc = 'Deficit Transfer'
                     SELECT baltransfer
                     SCAN FOR ctype = 'D'
                        SCATTER MEMVAR
                        THIS.ogl.cBatch     = GetNextPK('BATCH')
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.namount * -1
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cID        = m.cownerid
                        THIS.ogl.cReference = 'Deficit to Minimum'
                        THIS.ogl.UpdateBatch()

                        THIS.ogl.cAcctNo    = m.cDefAcct
                        THIS.ogl.namount    = m.namount
                        THIS.ogl.cReference = 'Deficit to Minimum'
                        THIS.ogl.UpdateBatch()
                        lnDefSwitch = lnDefSwitch + m.namount
                     ENDSCAN

* Get the amount of minimums that transferred
                     THIS.oprogress.SetProgressMessage('Calculating amounts of suspense that switched between minimum and deficit')
                     THIS.oprogress.UpdateProgress(THIS.nprogress)
                     THIS.nprogress = THIS.nprogress + 1

                     lnMinSwitch = THIS.osuspense.GetBalTransfer('M', .T.) * -1

                     SELECT baltransfer
                     SCAN FOR ctype = 'M'
                        SCATTER MEMVAR
                        THIS.ogl.cDesc      = 'Minimum Transfer'
                        THIS.ogl.cBatch     = GetNextPK('BATCH')
                        THIS.ogl.cAcctNo    = m.cMinAcct
                        THIS.ogl.namount    = m.namount
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cID        = m.cownerid
                        THIS.ogl.cReference = 'Minimum to Deficit'
                        THIS.ogl.UpdateBatch()

                        THIS.ogl.cAcctNo    = m.cDefAcct
                        THIS.ogl.namount    = m.namount * -1
                        THIS.ogl.cUnitNo    = m.cWellID
                        THIS.ogl.cReference = 'Minimum to Deficit'
                        THIS.ogl.UpdateBatch()
                        lnMinSwitch = lnMinSwitch - m.namount
                     ENDSCAN
                     swclose('baltransfer')
                  ENDIF
               ENDIF

               THIS.ndeftransfer = lnDefSwitch
               THIS.nmintransfer = lnMinSwitch
            ELSE
               THIS.ndeftransfer = 0
               THIS.nmintransfer = 0
            ENDIF

            THIS.ogl.cReference = 'Run: R' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + '/' + THIS.cgroup

*  Mark the expense entries as being tied to this DM batch
            swselect('expense')
            SCAN FOR nRunNoRev = THIS.nrunno ;
                  AND EMPTY(expense.cBatch)
               m.cWellID = cWellID
               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  IF cgroup = tcGroup
                     swselect('expense')
                     REPL cBatch WITH THIS.cdmbatch
                  ENDIF
               ENDIF
            ENDSCAN

*   Post the Vendor amounts that are designated to be posted.
            lnVendor = 0
            SELECT cvendorid, cVendName FROM vendor WHERE lIntegGL = .T. INTO CURSOR curVends
            IF NOT llNoPostDM AND _TALLY > 0
               lnMax            = _TALLY
               m.cID            = ''
               THIS.ogl.dGLDate = tdCompanyPost
               THIS.ogl.cBatch  = GetNextPK('BATCH')
               SELECT curVends
               SCAN
                  SCATTER MEMVAR

                  THIS.oprogress.SetProgressMessage('Posting Vendor Checks to General Ledger...' + m.cVendName)

                  lnAmount     = 0
                  THIS.ogl.cID = m.cvendorid
                  swselect('expense')
                  lnCount = 1
                  swselect('expense')
                  SCAN FOR cvendorid == m.cvendorid AND ;
                        nRunNoRev = THIS.nrunno  AND ;
                        cRunYearRev == THIS.crunyear AND ;
                        namount # 0 AND NOT lAPTran AND ;
                        NOT INLIST(ccatcode, 'COMP', 'GATH', 'PLUG')

                     m.cWellID   = cWellID
                     m.cexpclass = cexpclass
                     m.ccateg    = ccateg
                     m.namount   = namount
                     m.cUnitNo   = cWellID
                     m.cID       = cvendorid
                     m.cVendName = m.cVendName
                     m.ccatcode  = ccatcode
                     lcOwnerID   = cownerid
                     m.cdeck     = cdeck

*  Check to make sure the well is in the right group
                     swselect('wells')
                     SET ORDER TO cWellID
                     IF SEEK(m.cWellID)
                        IF wells.cgroup # tcGroup
                           LOOP
                        ENDIF
                     ENDIF

*  Get the account numbers to be posted for this expense category
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        SCATTER MEMVAR
                        m.ccateg   = ccateg
                        m.cDRAcctV = THIS.cexpclear
                        IF EMPTY(m.cCRAcctV)
                           m.cCRAcctV = lcSuspense
                        ENDIF
                     ELSE
                        m.cCRAcctV = lcSuspense
                     ENDIF

*  Net out any JIB interest shares from the expense
                     m.namount   = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., lcOwnerID, .F., m.cdeck)

*  Add amount of this invoice to the total the vendor is to be paid
                     lnAmount = lnAmount + m.namount

                     THIS.ogl.cUnitNo = m.cUnitNo
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.cAcctNo = m.cCRAcctV
                     THIS.ogl.namount = m.namount * -1
                     THIS.ogl.UpdateBatch()

                     THIS.ogl.cAcctNo = lcDMExp
                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.UpdateBatch()
                  ENDSCAN && Expense

                  llReturn = THIS.ogl.ChkBalance()

                  IF NOT llReturn
                     TRY
                        IF NOT FILE('datafiles\outbal.dbf')
                           CREATE TABLE datafiles\outbal FREE (cBatch  c(8), cownerid  c(10))
                        ENDIF
                        IF NOT USED('outbal')
                           USE datafiles\outbal IN 0
                        ENDIF
                        m.cBatch   = THIS.ogl.cBatch
                        m.cownerid = m.cID
                        INSERT INTO outbal FROM MEMVAR
                        llReturn = .T.
                     CATCH
                     ENDTRY
                  ENDIF


                  lnAmount     = 0
                  THIS.ogl.cID = m.cvendorid
                  swselect('expense')
                  lnCount = 1
                  swselect('expense')
                  SCAN FOR cvendorid == m.cvendorid AND ;
                        nRunNoRev = THIS.nrunno  AND ;
                        cRunYearRev == THIS.crunyear AND ;
                        namount # 0 AND NOT lAPTran AND ;
                        ccatcode = 'PLUG'

                     m.cWellID   = cWellID
                     m.cexpclass = cexpclass
                     m.ccateg    = ccateg
                     m.namount   = namount
                     m.cUnitNo   = cWellID
                     m.cID       = cvendorid
                     m.cVendName = m.cVendName
                     m.ccatcode  = ccatcode
                     lcOwnerID   = cownerid
                     m.cdeck     = cdeck

*  Check to make sure the well is in the right group
                     swselect('wells')
                     SET ORDER TO cWellID
                     IF SEEK(m.cWellID)
                        IF wells.cgroup # tcGroup
                           LOOP
                        ENDIF
                     ENDIF

*  Get the account numbers to be posted for this expense category
                     swselect('plugwell')
                     LOCATE FOR cWellID == m.cWellID
                     IF FOUND()
                        m.cCRAcctV = cAcctNo
                     ELSE
                        m.cCRAcctV = lcSuspense
                     ENDIF

*  Net out any JIB interest shares from the expense
                     m.namount   = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., lcOwnerID, .F., m.cdeck)

*  Add amount of this invoice to the total the vendor is to be paid
                     lnAmount = lnAmount + m.namount

                     THIS.ogl.cUnitNo = m.cUnitNo
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.cAcctNo = m.cCRAcctV
                     THIS.ogl.namount = m.namount * -1
                     THIS.ogl.UpdateBatch()

                     THIS.ogl.cAcctNo = lcDMExp
                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.UpdateBatch()
                  ENDSCAN && Expense

                  llReturn = THIS.ogl.ChkBalance()

                  IF NOT llReturn
                     TRY
                        IF NOT FILE('datafiles\outbal.dbf')
                           CREATE TABLE datafiles\outbal FREE (cBatch  c(8), cownerid  c(10))
                        ENDIF
                        IF NOT USED('outbal')
                           USE datafiles\outbal IN 0
                        ENDIF
                        m.cBatch   = THIS.ogl.cBatch
                        m.cownerid = m.cID
                        INSERT INTO outbal FROM MEMVAR
                        llReturn = .T.
                     CATCH
                     ENDTRY
                  ENDIF

               ENDSCAN && curVends
            ENDIF

*  Post the owners that are designated to be posted
            swselect('wells')
            SET ORDER TO cWellID

            lnOwner = 0

*  Get the owners to be posted.
            SELECT cownerid, cownname FROM investor WHERE lIntegGL = .T. INTO CURSOR curPostOwns
            lnOwns  = _TALLY
            lnCount = 1

            IF NOT llNoPostDM AND lnOwns > 0
               lnAmount         = 0
               THIS.ogl.dGLDate = tdCompanyPost

               SELECT curPostOwns
               SCAN
                  SCATTER MEMVAR

                  THIS.oprogress.SetProgressMessage('Posting Operator Owner Amounts to General Ledger...' + m.cownname)

* Post operator amounts
                  llReturn = THIS.postoperator('Invtmp', m.cownerid, m.cownname)
                  IF NOT llReturn
                     EXIT
                  ENDIF
* Post operator suspense amounts
                  llReturn = THIS.postoperator('tSuspense', m.cownerid, m.cownname)
                  IF NOT llReturn
                     EXIT
                  ENDIF
               ENDSCAN && curPostOwns

            ENDIF
         ENDIF


      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'PostJourn', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
         ENDIF
         THIS.ERRORMESSAGE('PostJourn', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()


      RETURN llReturn
   ENDPROC

*-- Create owner checks and places them in the check register to be printed.
*********************************
   PROCEDURE OwnerChks
*********************************
      LOCAL  tcYear, tcPeriod, tdCheckDate, tcGroup, tcBatch, tdPostDate
      LOCAL  m.nTotalChk, oprogress
      LOCAL lcAcctPrd, lcDeptNo, lcDisbAcct, lcExpClear, lcRevClear, lcIDChec, llReturn, lnChkAmt, lnCount
      LOCAL lnMax, lnMinCheck, lnTotal, loError
      LOCAL cyear, jnMinCheck, nTotalChk, tlRelMin

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Creating Owner Checks...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         tcYear   = THIS.crunyear
         tcPeriod = THIS.cperiod
         tcGroup  = THIS.cgroup
         tcBatch  = THIS.cdmbatch
         tlRelMin = THIS.lrelmin

         m.nTotalChk = 0

*  Build the owner check totals
         SELECT  invtmp.cownerid, ;
                 investor.cownname AS cPayee, ;
                 investor.lhold, ;
                 investor.ndisbfreq, ;
                 investor.ninvmin, ;
                 investor.lIntegGL, ;
                 SUM(invtmp.nnetcheck) AS nTotalCK ;
             FROM invtmp,;
                 investor ;
             WHERE investor.cownerid = invtmp.cownerid ;
                 AND IIF(m.goApp.lDirDMDep, (investor.ldirectdep =.F. AND investor.lFedwire =.F.), .T.) ;
                 AND EMPTY(cidcomp) ;
                 AND investor.ldummy = .F. ;
                 AND investor.lIntegGL = .F. ;
             INTO CURSOR invtotal ;
             ORDER BY invtmp.cownerid ;
             GROUP BY invtmp.cownerid

*  Get the Accounting Month
         lcAcctPrd = PADL(TRANSFORM(MONTH(THIS.dCheckDate)), 2, '0')

*  Get the Disbursements checking account
         lcDisbAcct = THIS.oOptions.cDisbAcct
         lnMinCheck = THIS.oOptions.nMinCheck
         IF m.goApp.lAMVersion
            lcDeptNo   = THIS.oOptions.cdeptno
         ELSE
            lcDeptNo   = ''
         ENDIF

* Get the clearing accounts
         swselect('glopt')
         GO TOP
         lcRevClear = crevclear
         lcExpClear = cexpclear

         m.cyear = tcYear
         lnTotal = 0

         swselect('investor')
         SET ORDER TO cownerid

         SELECT invtotal
         COUNT FOR nTotalCK > 0 TO lnMax
         lnCount = 1
         lnTotal = 0

         THIS.ogl.cReference = 'Period: ' + tcYear + '/' + tcPeriod + '/' + tcGroup

         THIS.ogl.cyear   = THIS.ogl.GetPeriod(THIS.dCheckDate, .T.)
         THIS.ogl.cperiod = THIS.ogl.GetPeriod(THIS.dCheckDate, .F.)

         THIS.ogl.dCheckDate = THIS.dCheckDate
         THIS.ogl.dGLDate    = THIS.dCheckDate
         THIS.ogl.dpostdate  = THIS.dCheckDate
         THIS.ogl.cgroup     = tcGroup
         THIS.ogl.cAcctNo    = lcDisbAcct
         THIS.ogl.cidtype    = 'I'
         THIS.ogl.cSource    = 'DM'
         THIS.ogl.cUnitNo    = ''
         THIS.ogl.cdeptno    = lcDeptNo
         THIS.ogl.cEntryType = 'C'

         lnChkAmt = 0

         SELECT invtotal
         SCAN FOR nTotalCK > 0
            SCATTER MEMVAR

            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Creating Owner Checks...' + m.cownerid)
            ENDIF

*  Setup the minimum check amount
            IF m.ninvmin = 0
               jnMinCheck = lnMinCheck
            ELSE
               jnMinCheck = m.ninvmin
            ENDIF

*  Reset the minimum amount to hold this owner's check
*  if it's not to be disbursed monthly or he's on hold.
            DO CASE
               CASE m.lhold                   && Owner on hold
                  jnMinCheck = 99999999
               CASE m.ndisbfreq = 2          && Quarterly
                  IF NOT INLIST(lcAcctPrd, '03', '06', '09', '12')
                     jnMinCheck = 99999999
                  ELSE
                     IF tlRelMin
                        jnMinCheck = 0
                     ENDIF
                  ENDIF

               CASE m.ndisbfreq = 3          && SemiAnnually
                  IF NOT INLIST(lcAcctPrd, '06', '12')
                     jnMinCheck = 99999999
                  ELSE
                     IF tlRelMin
                        jnMinCheck = 0
                     ENDIF
                  ENDIF
               CASE m.ndisbfreq = 4          && Annually
                  IF lcAcctPrd # '12'
                     jnMinCheck = 99999999
                  ELSE
                     IF tlRelMin
                        jnMinCheck = 0
                     ENDIF
                  ENDIF
               CASE tlRelMin                 && Release minimums
                  jnMinCheck = 0
            ENDCASE

            IF m.nTotalCK # 0 AND m.nTotalCK >= jnMinCheck

* Add the check to the check register
               THIS.ogl.cID     = m.cownerid
               THIS.ogl.namount = m.nTotalCK
               THIS.ogl.cPayee  = m.cPayee
               THIS.ogl.cBatch  = tcBatch
               THIS.ogl.cAcctNo = lcDisbAcct
               THIS.ogl.addcheck(.T.)
               lcIDChec = THIS.ogl.cidchec

               SELE invtmp
               SCAN FOR cownerid = m.cownerid
                  REPL cidchec WITH lcIDChec
               ENDSCAN
            ENDIF
         ENDSCAN

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Creating Owner Checks...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1

         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'OwnerChks', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('OwnerChks', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC


*-- Creates vendor checks and places them in the check register to be printed.
*********************************
   PROCEDURE VendorChks
*********************************
      LOCAL tcYear, tcPeriod, tdCheckDate, tcGroup, tcBatch, tdPostDate, oprogress
      LOCAL jUnique, lcBatch, llSepClose, lcVendComp, lcDisbAcct, lnFixedAmt, lnAPAmt, lnCompAmt
      LOCAL lIntegGL, laGathComp[1], lcAPAcct, lcCompGath, lcDMExp, lcNetType, lcVendListID, lcVendName
      LOCAL lcIDChec, lhold, llNoPostDM, llReturn, lnChkAmt, lnCount, lnMax, lnMax1, lnMax2, lnTotal
      LOCAL loError, cPayee, nMinCheck, namount

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Creating Vendor Checks...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         tcYear      = THIS.crunyear
         tcPeriod    = THIS.cperiod
         tdCheckDate = THIS.dCheckDate
         tdPostDate  = THIS.dCheckDate
         tcBatch     = THIS.cdmbatch
         tcGroup     = THIS.cgroup

         swselect('glopt')
         llNoPostDM = lDMNoPost

         swselect('apopt')
         lcAPAcct = capacct

*  Get the Disbursements checking account
         lcDisbAcct = THIS.oOptions.cDisbAcct
         lcVendComp = THIS.oOptions.cVendComp
         llSepClose = .T.
         lcDMExp    = THIS.oOptions.cFixedAcct
         lcCompGath = THIS.oOptions.cGathAcct

         IF EMPTY(lcDMExp)
            lcDMExp = lcAPAcct
         ENDIF

         lcNetType = 'N'

         IF m.goApp.ldmpro
* Don't create journal entries for stand-alone disb mgr
            llNoPostDM = .T.
         ENDIF

         STORE 0 TO lnMax1, lnMax2, lnFixedAmt, lnAPAmt, lnCompAmt

**************************************
*  Net down expenses for dummy owners
**************************************
         swselect('expense')
         Make_Copy('Expense', 'Exptemp')
         swselect('wells')
         SET ORDER TO cWellID
         swselect('expense')
         SCAN FOR nRunNoRev = THIS.nrunno AND ;
               (cRunYearRev = THIS.crunyear OR INLIST(cRunYearRev, '1900', '1901') AND cyear > '2008') ;
               AND EMPTY(cpaidbyck) AND cyear # 'FIXD' AND ccatcode # 'PLUG'
            SCATTER MEMVAR MEMO
            SELECT wells
            IF SEEK(m.cWellID) AND cgroup == THIS.cgroup
               m.namount = swNetExp(m.namount, m.cWellID, .T., expense.cexpclass, 'B', .F., m.cownerid, .F., m.cdeck)
               INSERT INTO exptemp FROM MEMVAR
            ENDIF
         ENDSCAN

         SELECT  exptemp.cvendorid, ;
                 cyear, ;
                 cperiod, ;
                 wells.nprocess, ;
                 SUM(exptemp.namount) AS namount, ;
                 vendor.cVendName AS cPayee   ;
             FROM exptemp,;
                 vendor,;
                 wells ;
             WHERE vendor.lIntegGL = .F. ;
                 AND exptemp.cvendorid = vendor.cvendorid           ;
                 AND exptemp.cWellID = wells.cWellID                ;
                 AND EMPTY(exptemp.cpaidbyck)                       ;
                 AND exptemp.lAPTran # .T.                         ;
                 AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U')        ;
                 AND exptemp.cWellID IN (SELECT  DISTINCT cWellID ;
                                             FROM wellinv) ;
             INTO CURSOR vendc1 ;
             GROUP BY exptemp.cvendorid ;
             ORDER BY exptemp.cvendorid

         lnMax   = _TALLY
         lnCount = 1

         SELECT  exptemp.cvendorid, ;
                 cyear, ;
                 cperiod, ;
                 wells.nprocess, ;
                 SUM(exptemp.namount) AS namount, ;
                 vendor.cVendName AS cPayee, ;
                 exptemp.lfixed ;
             FROM exptemp,;
                 vendor,;
                 wells ;
             WHERE  vendor.lIntegGL = .F. ;
                 AND exptemp.cvendorid = vendor.cvendorid           ;
                 AND exptemp.cWellID = wells.cWellID                ;
                 AND EMPTY(exptemp.cpaidbyck)                       ;
                 AND exptemp.lAPTran # .T.                         ;
                 AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U')        ;
                 AND exptemp.cWellID IN (SELECT  DISTINCT cWellID ;
                                             FROM wellinv) ;
             INTO CURSOR vendtype ;
             GROUP BY exptemp.cvendorid,;
                 exptemp.lfixed ;
             ORDER BY exptemp.cvendorid,;
                 exptemp.lfixed

         CREATE CURSOR vendchks ;
            (cvendorid    c(10), ;
              cPayee      c(40), ;
              cyear        c(4), ;
              cperiod      c(2), ;
              nprocess     N(1), ;
              namount      N(12, 2))

         SELECT vendchks
         APPEND FROM DBF('vendc1')
         swclose('vendc1')

*  Calculate the outstanding minimums for the vendors
         lnMax   = lnMax * 2
         lnTotal = 0

*   DO suspense processing for vendor checks
         SELECT vendchks
         SCAN FOR namount > 0
            SCATTER MEMVAR

*  Get the minimum check amount for this vendor
            swselect('Vendor')
            SET ORDER TO cvendorid
            IF SEEK(m.cvendorid)
               m.nMinCheck = nMinCheck
               IF m.goApp.ldmpro
                  m.lhold     = .F.
               ELSE
                  m.lhold     = lhold
               ENDIF
               m.lIntegGL  = lIntegGL
               IF NOT llNoPostDM
* Don't create a check for this vendor
                  IF lSkipCheck
                     LOOP
                  ENDIF
               ENDIF
            ELSE
               m.nMinCheck = 0
               m.lhold     = .F.
               m.lIntegGL  = .F.
            ENDIF

*  If this vendor's income should be posted directly to the G/L
*  don't create a check.
            IF m.lIntegGL
               LOOP
            ENDIF

         ENDSCAN

         lnChkAmt = 0

         THIS.ogl.DMBatch    = THIS.cdmbatch

         IF NOT llNoPostDM
            lcBatch             = GetNextPK('BATCH')
         ENDIF

         THIS.ogl.cDesc      = 'DM Vendor Check'
         THIS.ogl.cReference = 'Period: ' + tcYear + '/' + tcPeriod + '/' + tcGroup
         THIS.ogl.cyear      = tcYear
         THIS.ogl.cperiod    = tcPeriod
         THIS.ogl.dCheckDate = tdCheckDate
         THIS.ogl.dpostdate  = tdPostDate
         THIS.ogl.cidtype    = 'V'
         THIS.ogl.cSource    = 'DM'
         THIS.ogl.cAcctNo    = lcDisbAcct
         THIS.ogl.cgroup     = tcGroup
         THIS.ogl.cEntryType = 'C'

*  Create checks in check register
         SELECT vendchks
         SCAN FOR namount > 0
            SCATTER MEMVAR

            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Creating Vendor Checks...' + m.cvendorid)
            ENDIF

            swselect('Vendor')
            SET ORDER TO cvendorid
            IF SEEK(m.cvendorid)
               lcVendName = cVendName
               
               IF vendor.lIntegGL
                  LOOP
               ENDIF
               IF vendor.lSkipCheck
                  LOOP
               ENDIF
            ELSE
               LOOP
            ENDIF

*  Create a check to pay the vendor's expenses
            THIS.ogl.cBatch  = tcBatch
            THIS.ogl.cID     = m.cvendorid
            THIS.ogl.cPayee  = lcVendName
            THIS.ogl.namount = m.namount
            THIS.ogl.cAcctNo = lcDisbAcct
            THIS.ogl.cidtype = 'V'
            THIS.ogl.addcheck()
            lcIDChec         = THIS.ogl.GETKEY()
            THIS.ogl.namount = m.namount * -1
            THIS.ogl.cidchec = lcIDChec
            THIS.ogl.UpdateBatch()

            SELECT vendtype
            LOCATE FOR cvendorid == m.cvendorid AND NOT lfixed
            IF FOUND()
               m.namount        = vendtype.namount
               THIS.ogl.cAcctNo = lcAPAcct
               THIS.ogl.cDesc   = 'A/P Expenses'
               THIS.ogl.namount = m.namount
               THIS.ogl.UpdateBatch()
            ENDIF
            SELECT vendtype
            LOCATE FOR cvendorid == m.cvendorid AND lfixed
            IF FOUND()
               m.namount        = vendtype.namount
               THIS.ogl.cAcctNo = lcDMExp
               THIS.ogl.cDesc   = 'Fixed Expenses'
               THIS.ogl.namount = m.namount
               THIS.ogl.UpdateBatch()
            ENDIF

*   Update the expense records with the check they were paid with
            THIS.expenseupd(m.cvendorid, lcIDChec)

*  Add the check to the total so we can post one
*  entry to the expense clearing account
            lnChkAmt = lnChkAmt + m.namount

         ENDSCAN

         swclose('vendchks')
         swclose('vendtype')

*   Process compression & gathering charges if a vendor is to be paid
*!*    Commented out 6/1/2022 by pws - Deprecating this feature
*!*            IF NOT EMPTY(lcVendComp)
*!*               swselect('Vendor')
*!*               SET ORDER TO cvendorid
*!*               IF SEEK(lcVendComp) AND NOT lIntegGL
*!*                  m.cPayee = cVendName
*!*                  SELECT  SUM(nCompress + nGather) AS nCompGath ;
*!*                     FROM wellwork WITH (BUFFERING = .T.) ;
*!*                     WHERE nrunno = THIS.nrunno ;
*!*                     AND crunyear = THIS.crunyear ;
*!*                     AND cgroup = tcGroup ;
*!*                     INTO ARRAY laGathComp
*!*                  IF _TALLY > 0 AND laGathComp[1] # 0
*!*                     THIS.ogl.cID        = lcVendComp
*!*                     THIS.ogl.cPayee     = m.cPayee
*!*                     THIS.ogl.cidtype    = 'V'
*!*                     THIS.ogl.cperiod    = tcPeriod
*!*                     THIS.ogl.cyear      = tcYear
*!*                     THIS.ogl.cMemo      = 'Compression/Gathering '
*!*                     THIS.ogl.cgroup     = tcGroup
*!*                     THIS.ogl.cAcctNo    = lcDisbAcct
*!*                     THIS.ogl.dCheckDate = tdCheckDate
*!*                     THIS.ogl.dpostdate  = tdPostDate
*!*                     THIS.ogl.namount    = laGathComp[1]
*!*                     THIS.ogl.cBatch     = tcBatch
*!*                     THIS.ogl.cSource    = 'DM'
*!*                     THIS.ogl.addcheck()
*!*                     lcIDChec = THIS.ogl.GETKEY()

*!*                     *  Build the G/L
*!*                     THIS.ogl.cBatch  = tcBatch
*!*                     THIS.ogl.cAcctNo = lcDisbAcct
*!*                     THIS.ogl.namount = laGathComp[1] * -1
*!*                     THIS.ogl.cDesc   = m.cPayee
*!*                     THIS.ogl.cidchec = lcIDChec
*!*                     THIS.ogl.UpdateBatch()

*!*                     THIS.ogl.cAcctNo = lcCompGath
*!*                     THIS.ogl.namount = laGathComp[1]
*!*                     THIS.ogl.UpdateBatch()
*!*                  ENDIF
*!*               ENDIF
*!*            ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Creating Vendor Checks...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'VendorChks', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('VendorChks', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Marks the expenses as having been paid
*********************************
   PROCEDURE ExpenseUpd
*********************************
      LPARA tcVendor, tcidChec
      LOCAL tcYear, tcPeriod, tcGroup

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         tcYear   = THIS.crunyear
         tcPeriod = THIS.cperiod
         tcGroup  = THIS.cgroup

         SELE expense
         SCAN FOR nRunNoRev = THIS.nrunno ;
               AND (cRunYearRev = tcYear OR INLIST(cRunYearRev, '1900', '1901')) ;
               AND cvendorid = tcVendor ;
               AND EMPTY(cpaidbyck) ;
               AND lAPTran = .F. ;
               AND cyear # "FIXD"

            m.cWellID = cWellID
            SELE wells
            LOCATE FOR cWellID = m.cWellID
            IF FOUND() AND nprocess = 2
               IF NOT THIS.lrelqtr
                  LOOP
               ENDIF
            ENDIF
            SELE expense
            REPLACE cpaidbyck WITH tcidChec
         ENDSCAN

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'ExpenseUpd', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('ExpenseUpd', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC


*-- Calculates the closing summary
*********************************
   PROCEDURE CalcSummary
*********************************
      LPARAMETERS tlReport
      LOCAL lcYear, lcPeriod, lcGroup, lnRevEnt, lnExpEnt
      LOCAL lnRevAll, lnExpAll, lcTitle1, lcDMBatch, llSepClose, lnFlatRate, llFoundRev
      LOCAL lnJIBAll, lnJIBInv, lnJibInvAmt, lnRunNo, m.lDirGasPurch, m.lDirOilPurch
      LOCAL lnDirectDep, lnDirectDe2
      LOCAL lDirGasPurch, lDirOilPurch, lExempt, lIntegGL, lRoyExempt, laTotal[1], lcExpProdPrds
      LOCAL lcRevProdPrds, lcScanFor, ldirectdep, llReturn, lnBackupWH, lnCompression, lnDefSwitch
      LOCAL lnGTax1, lnGTax2, lnGTax3, lnGTax4, lnGathering, lnIncome, lnMinSwitch, lnMinimum, lnOTax1
      LOCAL lnOTax2, lnOTax3, lnOTax4, lnOwnerPost, lnPTax1, lnPTax2, lnPTax3, lnPTax4, lnSevTaxOwn
      LOCAL lnSevTaxWell, lnTaxWH, lnTaxWith, lnTime, lnTransfer, lnVendorPost, lnpctcnt, loError
      LOCAL cExpProdPrds, cGrpName, cownerid, cProcessor, cProducer, cRevProdPrds, cRunTime, cWellID
      LOCAL cdmbatch, cexpclass, cperiod, cvendorid, cyear, dexpense, dpostdate, drevenue, glGrpName
      LOCAL nBack, nCompression, nDefAfter, nDefBefore, nDefCurr, nDefPrior, nDirectDe2, nDirectDep
      LOCAL nExpAllocated, nExpEntered, nGathering, nHoldCurr, nHoldPrior, nJExpAllocated
      LOCAL nJIBInvAmount, nJIBInvCount, nMinAfter, nMinBefore, nMinCurr, nMinPrior, nOwnChkAmt
      LOCAL nOwnChkCount, nOwnerPost, nRevAllocated, nRevEntered, nSevTaxOwn, nSevTaxWell, nTax
      LOCAL nVendChkAmt, nVendChkCount, nVendorPost, namount, nbackwith, ndeftransfer, nflatrate
      LOCAL nmintransfer, nnetcheck, ntaxwith, lnPlugging


      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF tlReport
            THIS.lclose = .F.
         ENDIF

         STORE 0 TO lnDirectDep, lnDirectDe2

*  Builds the closing summary information
         CREATE CURSOR closetemp ;
            (nRevEntered    N(12, 2), ;
              nRevAllocated  N(12, 2), ;
              nExpEntered    N(12, 2), ;
              nExpAllocated  N(12, 2), ;
              nJExpAllocated N(12, 2), ;
              nJIBInvCount   i, ;
              nJIBInvAmount  N(12, 2), ;
              nOwnChkCount   i, ;
              nOwnChkAmt     N(12, 2), ;
              nDirectDep     N(12, 2), ;
              nDirectDe2     i, ;
              nOwnerPost     N(12, 2), ;
              nVendorPost    N(12, 2), ;
              nVendChkCount  i, ;
              nVendChkAmt    N(12, 2), ;
              nSevTaxWell    N(9, 2), ;
              nSevTaxOwn     N(9, 2), ;
              nGathering     N(12, 2), ;
              nCompression   N(12, 2), ;
              nDefPrior      N(12, 2), ;
              nDefCurr       N(12, 2), ;
              nMinPrior      N(12, 2), ;
              nMinCurr       N(12, 2), ;
              nHoldPrior     N(12, 2), ;
              nHoldCurr      N(12, 2), ;
              nbackwith      N(12, 2), ;
              ntaxwith       N(12, 2), ;
              nDefBefore     N(12, 2), ;
              nDefAfter      N(12, 2), ;
              ndeftransfer   N(12, 2), ;
              nMinBefore     N(12, 2), ;
              nMinAfter      N(12, 2), ;
              nmintransfer   N(12, 2), ;
              nHoldsBefore   N(12, 2), ;
              nHoldsAfter    N(12, 2), ;
              nflatrate      N(12, 2),  ;
              nPluggingFund  N(12, 2), ;
              nPartnerPost   N(12, 2), ;
              cRevProdPrds   c(80), ;
              cExpProdPrds   c(80), ;
              drevenue       D, ;
              dexpense       D, ;
              dpostdate      D, ;
              cRunTime       c(40))

         STORE '' TO cRevProdPrds, cExpProdPrds, cRunTime
         STORE 0  TO nRevEntered, nRevAllocated, nExpEntered, nExpAllocated, nJExpAllocated
         STORE 0  TO nJIBInvCount, nJIBInvAmount, nOwnChkCount, nOwnChkAmt, nDirectDep, nDirectDe2
         STORE 0  TO nOwnerPost, nVendorPost, nVendChkCount, nVendChkAmt, nSevTaxWell, nSevTaxOwn
         STORE 0  TO nGathering, nCompression, nDefPrior, nDefCurr, nMinPrior, nMinCurr, nHoldPrior
         STORE 0  TO nHoldCurr, nbackwith, ntaxwith, nDefBefore, nDefAfter, ndeftransfer, nMinBefore
         STORE 0  TO nMinAfter, nmintransfer, nHoldsBefore, nHoldsAfter, nflatrate
         STORE {} TO drevenue, dexpense, dpostdate

         glGrpName  = THIS.oOptions.lGrpName
         llSepClose = .T.
         lnMinimum  = THIS.oOptions.nMinCheck
         lcVendComp = THIS.oOptions.cVendComp

         lcYear     = THIS.crunyear
         lcGroup    = THIS.cgroup
         lnRunNo    = THIS.nrunno
         lnFlatRate = 0

         IF glGrpName
            swselect('groups')
            SET ORDER TO cgroup
            IF SEEK(lcGroup)
               m.cGrpName = cDesc
            ELSE
               IF lcGroup = '**'
                  m.cGrpName = 'All Companies'
               ELSE
                  m.cGrpName = ''
               ENDIF
            ENDIF
         ELSE
            m.cGrpName = ''
         ENDIF

         IF TYPE('m.goApp') = 'O'
            m.cProducer = m.goApp.cCompanyName
         ELSE
            m.cProducer = 'Development Company'
         ENDIF

         m.cProcessor = ''

         STORE .F. TO m.lDirGasPurch, m.lDirOilPurch
         STORE 0 TO lnJIBAll, lnJIBInv, lnJibInvAmt

         IF THIS.lclose
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Creating the Closing Summary Page...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            swselect('wells')
            SET ORDER TO cWellID

            m.drevenue  = THIS.drevdate
            m.dexpense  = THIS.dexpdate
            m.dpostdate = THIS.dpostdate

            lnRevEnt = 0

            CREATE CURSOR tempprd ;
               (cyear   c(4), ;
                 cperiod c(2))

            SELE income
            SCAN FOR nrunno = lnRunNo AND crunyear = lcYear
               m.cyear   = cyear
               m.cperiod = cperiod
               INSERT INTO tempprd FROM MEMVAR
            ENDSCAN

            lcRevProdPrds = ''
            SELE cyear, cperiod FROM tempprd ORDER BY cyear, cperiod GROUP BY cyear, cperiod INTO CURSOR temp
            SELE temp
            SCAN
               lcRevProdPrds = lcRevProdPrds + cyear + '/' + cperiod + ' '
            ENDSCAN

            swclose('temp')

            SELE tempprd
            DELE ALL

            SELE expense
            SCAN FOR nRunNoRev = lnRunNo AND cRunYearRev = lcYear
               m.cyear   = cyear
               m.cperiod = cperiod
               INSERT INTO tempprd FROM MEMVAR
            ENDSCAN
            lcExpProdPrds = ''
            SELE cyear, cperiod FROM tempprd ORDER BY cyear, cperiod GROUP BY cyear, cperiod INTO CURSOR temp
            SELE temp
            SCAN
               lcExpProdPrds = lcExpProdPrds + cyear + '/' + cperiod + ' '
            ENDSCAN

            swclose('temp')

            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Revenue Entered...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

*  Get the revenue entered for monthly wells
            swselect('income')
            SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND NOT 'TAX' $ cSource
               SCATTER MEMVAR
               swselect('wells')
               IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess # 2
                  DO CASE
                     CASE m.cSource = 'BBL'
                        lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'O', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'MCF'
                        lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'G', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'MISC1'
                        lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, '1', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'MISC2'
                        lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, '2', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'TRANS'
                        lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'T', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'OTH'
                        lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'P', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE INLIST(m.cSource, 'COMP', 'GATH')
* Ignore
                     OTHERWISE
                        IF NOT 'TAX' $ m.cSource
                           lnRevEnt = lnRevEnt + m.nTotalInc
                        ELSE
                           LOOP
                        ENDIF
                  ENDCASE
               ENDIF
            ENDSCAN

            IF THIS.lrelqtr
               swselect('income')
               SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND NOT 'TAX' $ cSource
                  SCATTER MEMVAR
                  swselect('wells')
                  IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess = 2
                     DO CASE
                        CASE m.cSource = 'BBL'
                           lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'O', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                        CASE m.cSource = 'MCF'
                           lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'G', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                        CASE m.cSource = 'MISC1'
                           lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, '1', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                        CASE m.cSource = 'MISC2'
                           lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, '2', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                        CASE m.cSource = 'TRANS'
                           lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'T', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                        CASE m.cSource = 'OTH'
                           lnRevEnt = lnRevEnt + swnetrev(m.cWellID, m.nTotalInc, 'P', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                        CASE INLIST(m.cSource, 'COMP', 'GATH')
* Ignore
                        OTHERWISE
                           IF NOT 'TAX' $ m.cSource
                              lnRevEnt = lnRevEnt + m.nTotalInc
                           ELSE
                              LOOP
                           ENDIF
                     ENDCASE
                  ENDIF
               ENDSCAN
            ENDIF

            IF THIS.lflatrates  &&  Check disbhist for flat rates allocated - BH 06/24/08
               SELECT invtmp
               SCAN FOR ((nrunno = lnRunNo AND crunyear = lcYear))
                  lnFlatRate = lnFlatRate + nflatrate
               ENDSCAN

               SELECT suspense
               SCAN FOR nrunno_in = lnRunNo AND crunyear_in = lcYear
                  lnFlatRate = lnFlatRate + nflatrate
               ENDSCAN
            ENDIF

            lnExpEnt = 0

            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Expenses Entered...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            IF llSepClose
               lcScanFor = "nRunNoRev = lnRunNo AND cRunYearRev = lcYear and not INLIST(ccatcode,'GATH','COMP')"
            ELSE
               lcScanFor = "(nRunNoRev = lnRunNo AND cRunYearRev = lcYear) OR (nRunnoJIB = lnRunNo AND cRunYearJIB = lcYear) and not INLIST(ccatcode,'GATH','COMP')"
            ENDIF
            swselect('expense')
            SCAN FOR &lcScanFor
               SCATTER MEMVAR
               IF NOT m.goApp.lPluggingModule AND m.cexpclass = 'P'
                  LOOP
               ENDIF
               swselect('wells')
               IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess # 2 AND cwellstat # 'V'
                  IF NOT EMPTY(m.cownerid)
                     swselect('investor')
                     LOCATE FOR cownerid == m.cownerid  &&  If it's been allocated to the dummy, ignore it
                     IF FOUND() AND NOT investor.ldummy
                        SELE wellinv
                        LOCATE FOR cownerid = m.cownerid AND cWellID = m.cWellID AND ctypeinv = 'W'
                        IF FOUND() AND lJIB
                           IF llSepClose
                              lnJIBAll = lnJIBAll + m.namount
                           ENDIF
                        ELSE
                           lnExpEnt = lnExpEnt + m.namount
                        ENDIF
                     ENDIF
                  ELSE
                     IF m.ccatcode = 'MKTG'
                        lnExpEnt = lnExpEnt + swnetrev(m.cWellID, m.namount, 'G', .F., .T., .T., .F., .F., .F., .F., m.cdeck)
                     ELSE
                        lnExpEnt = lnExpEnt + swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, m.ccatcode, m.cdeck)
                        lnJIBAll = lnJIBAll + swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'J', .F., m.cownerid, m.ccatcode, m.cdeck)
                     ENDIF
                  ENDIF
               ENDIF
               swselect('expense')
            ENDSCAN

            IF THIS.lrelqtr
               swselect('expense')
               SCAN FOR (nRunNoRev = lnRunNo AND cRunYearRev = lcYear) OR ;
                     (nrunnojib = lnRunNo AND crunyearjib = lcYear AND INLIST(cRunYearRev, '1900', '1901')) AND NOT ;
                     INLIST(ccatcode, 'GATH', 'COMP')
                  SCATTER MEMVAR
                  IF NOT m.goApp.lPluggingModule AND m.cexpclass = 'P'
                     LOOP
                  ENDIF
                  swselect('wells')
                  IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess = 2 AND cwellstat # 'V'
                     IF NOT EMPTY(m.cownerid)
                        swselect('investor')
                        LOCATE FOR cownerid == m.cownerid  &&  If it's been allocated to the dummy, ignore it
                        IF FOUND() AND NOT investor.ldummy
                           SELE wellinv
                           LOCATE FOR cownerid = m.cownerid AND cWellID = m.cWellID AND ctypeinv = 'W'
                           IF FOUND() AND lJIB
                              IF llSepClose
                                 lnJIBAll = lnJIBAll + m.namount
                              ENDIF
                           ELSE
                              lnExpEnt = lnExpEnt + m.namount
                           ENDIF
                        ENDIF
                     ELSE
                        IF m.ccatcode = 'MKTG'
                           lnExpEnt = lnExpEnt + swnetrev(m.cWellID, m.namount, 'G', .F., .T., .T., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           lnExpEnt = lnExpEnt + swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, m.ccatcode, m.cdeck)
                           lnJIBAll = lnJIBAll + swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'J', .F., m.cownerid, m.ccatcode, m.cdeck)
                        ENDIF
                     ENDIF
                  ENDIF
                  swselect('expense')
               ENDSCAN
            ENDIF

            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Direct Deposits...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1

            ENDIF

* Get the direct deposit totals and count
            lnDirectDep    = 0
            lnDirectDepCnt = 0
            IF FILE(m.goApp.cdatafilepath + 'dirdep.dbf')
               IF NOT USED('dirdep')
                  USE (m.goApp.cdatafilepath + 'dirdep') IN 0
               ENDIF
               SELECT dirdep
               LOCATE FOR nrunno = THIS.nrunno AND crunyear = THIS.crunyear
               IF FOUND()
                  SCATTER MEMVAR
                  lnDirectDep    = m.namount
                  lnDirectDepCnt = m.nCount
               ENDIF
               swclose('dirdep')
            ELSE
               lnDirectDep    = 0
               lnDirectDepCnt = 0
            ENDIF

*  Get the revenue and expenses allocated
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Revenue/Expenses Allocated...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1

            ENDIF

            lnRevAll    = 0
            lnExpAll    = 0
            lnSevTaxOwn = 0
            lnOwnerPost = 0

            swselect('disbhist')
            SELECT  * ;
                FROM disbhist WITH (BUFFERING = .T.) ;
                WHERE ((nrunno = lnRunNo ;
                        AND crunyear = lcYear ;
                        AND csusptype = ' ') ;
                      OR (nrunno_in = lnRunNo ;
                        AND crunyear_in = lcYear ;
                        AND NOT nrunno = 9999)) ;
                    AND crectype = 'R' ;
                INTO CURSOR temphist READWRITE

            SELECT temphist
            SCAN FOR NOT lManual
               SCATTER MEMVAR
               m.lExempt = .F.
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(m.cownerid)
                  m.lExempt    = lExempt
                  m.ldirectdep = ldirectdep
                  m.lIntegGL   = lIntegGL
                  IF investor.ldummy
                     LOOP
                  ENDIF
               ENDIF
               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  m.lDirOilPurch = lDirOilPurch
                  m.lDirGasPurch = lDirGasPurch
                  SCATTER FIELDS LIKE lSev* MEMVAR
                  m.lRoyExempt = lroysevtx
               ELSE
                  LOOP
               ENDIF

               lnIncome = m.nIncome
               lnIncome = lnIncome - IIF(m.cdirect = 'O', m.noilrev, 0)
               lnIncome = lnIncome - IIF(m.cdirect = 'G', m.ngasrev, 0)
               lnIncome = lnIncome - IIF(m.cdirect = 'B', m.noilrev + m.ngasrev, 0)

               lnRevAll = lnRevAll + lnIncome
               lnExpAll = lnExpAll + ;
                  m.nexpense + ;
                  m.ntotale1 + ;
                  m.ntotale2 + ;
                  m.ntotale3 + ;
                  m.ntotale4 + ;
                  m.ntotale5 + ;
                  m.ntotalea + ;
                  m.ntotaleb + ;
                  m.nMKTGExp + ;
                  m.nPlugExp

               IF m.lIntegGL
                  lnOwnerPost = lnOwnerPost + m.nnetcheck
               ENDIF
               IF m.lExempt OR (m.lRoyExempt AND m.ctypeinv # 'W')  && If royalty owners are exempt don't include them
                  LOOP
               ELSE
                  DO CASE
                     CASE m.cdirect = 'O' OR m.cdirect = 'B'
                        IF m.lDirOilPurch
                           lnSevTaxOwn = lnSevTaxOwn +  m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ELSE
                           lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ENDIF
                     CASE m.cdirect = 'G' OR m.cdirect = 'B'
                        IF m.lDirGasPurch
                           lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ELSE
                           lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ENDIF
                     CASE m.cdirect = 'N'
                        lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                  ENDCASE
               ENDIF
            ENDSCAN

            IF USED('tsuspense')
               USE IN tsuspense
            ENDIF
            SELECT  * ;
                FROM suspense WITH (BUFFERING = .T.) ;
                INTO CURSOR tsuspense NOFILTER  ;
                WHERE nrunno_in = lnRunNo ;
                    AND crectype = 'R' ;
                    AND crunyear_in = lcYear ;
                    AND NOT lManual
            SELECT tsuspense
            SCAN
               SCATTER MEMVAR
               m.cownerid = cownerid
               m.lExempt  = .F.
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(m.cownerid)
                  m.lExempt    = lExempt
                  m.ldirectdep = ldirectdep
                  m.lIntegGL   = lIntegGL
                  IF investor.ldummy
                     LOOP
                  ENDIF
               ENDIF
               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  m.lDirOilPurch = lDirOilPurch
                  m.lDirGasPurch = lDirGasPurch
                  SCATTER FIELDS LIKE lSev* MEMVAR
                  m.lRoyExempt = lroysevtx
               ELSE
                  LOOP
               ENDIF

               lnIncome = m.nIncome
               lnIncome = lnIncome - IIF(m.cdirect = 'O', m.noilrev, 0)
               lnIncome = lnIncome - IIF(m.cdirect = 'G', m.ngasrev, 0)
               lnIncome = lnIncome - IIF(m.cdirect = 'B', m.noilrev + m.ngasrev, 0)

               lnRevAll = lnRevAll + lnIncome
               lnExpAll = lnExpAll + ;
                  m.nexpense + ;
                  m.ntotale1 + ;
                  m.ntotale2 + ;
                  m.ntotale3 + ;
                  m.ntotale4 + ;
                  m.ntotale5 + ;
                  m.ntotalea + ;
                  m.ntotaleb + ;
                  m.nMKTGExp + ;
                  m.nPlugExp

               IF m.lIntegGL
                  lnOwnerPost = lnOwnerPost + m.nnetcheck
               ENDIF
               IF m.lExempt OR (m.lRoyExempt AND m.ctypeinv # 'W') && If royalty owners are exempt don't include them
                  LOOP
               ELSE
                  DO CASE
                     CASE m.cdirect = 'O'
                        IF m.lDirOilPurch
                           lnSevTaxOwn = lnSevTaxOwn +  m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ELSE
                           lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ENDIF
                     CASE m.cdirect = 'G' OR m.cdirect = 'B'
                        IF m.lDirGasPurch
                           lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ELSE
                           lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        ENDIF
                     CASE m.cdirect = 'B'
                        lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                        IF m.lDirOilPurch
                           lnSevTaxOwn = lnSevTaxOwn - m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4
                        ENDIF
                        IF m.lDirGasPurch
                           lnSevTaxOwn = lnSevTaxOwn - m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4
                        ENDIF

                     CASE m.cdirect = 'N'
                        lnSevTaxOwn = lnSevTaxOwn +  m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                  ENDCASE
               ENDIF
            ENDSCAN

            IF NOT THIS.lclose
               WAIT WIND NOWAIT 'Summing Expenses Allocated to JIB Owners...'
            ENDIF

* Get the count of JIB Invoices
            swselect('invhdr')
            COUNT FOR nrunno = lnRunNo AND cgroup = lcGroup AND cInvType = 'J' AND NOT lNetJIB TO lnJIBInv

* Get the amount of JIB Invoices
            SELECT  SUM(nInvTot) AS nInvTot ;
                FROM invhdr,;
                    investor ;
                WHERE nrunno = lnRunNo ;
                    AND crunyear = lcYear ;
                    AND cgroup = lcGroup ;
                    AND cInvType = 'J' ;
                    AND cCustID = investor.cownerid ;
                    AND NOT lNetJIB ;
                    AND NOT investor.ldummy ;
                INTO ARRAY laTotal
            IF _TALLY > 0
               lnJibInvAmt = laTotal[1]
            ELSE
               lnJibInvAmt = 0
            ENDIF

            lnGathering   = 0
            lnCompression = 0

*  Get the severance taxes allocated to the well
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Well Taxes...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            swselect('expcat')
            SET ORDER TO ccatcode   && CCATCODE
            IF SEEK('COMP')
               lcCompClass = cexpclass
            ELSE
               lcCompClass = 'G'
            ENDIF
            IF SEEK('GATH')
               lcGathClass = cexpclass
            ELSE
               lcGathClass = 'G'
            ENDIF

            lnSevTaxWell = 0
            SELECT wellwork
            SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND cgroup = lcGroup AND crectype = 'R'
               SCATTER MEMVAR

               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  m.lDirOilPurch = lDirOilPurch
                  m.lDirGasPurch = lDirGasPurch
                  SCATTER FIELDS LIKE lSev* MEMVAR
                  m.lRoyExempt = lroysevtx
               ELSE
                  LOOP
               ENDIF


               lnGathering   = lnGathering + swnetrev(m.cWellID, m.nGather, lcGathClass, .F., .T., .T., .F., .F., .F., .F., m.cdeck)
               lnCompression = lnCompression + swnetrev(m.cWellID, m.nCompress, lcCompClass, .F., .T., .T., .F., .F., .F., .F., m.cdeck)

               STORE 0 TO lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnGTax1, lnGTax2, lnGTax3, lnGTax4,  ;
                  lnPTax1, lnPTax2, lnPTax3, lnPTax4
               swselect('income')  &&  Total up the one-man tax entries, so they can be subtracted off before doing the netrev
               SCAN FOR cWellID == m.cWellID AND nrunno = m.nrunno AND crunyear = m.crunyear  ;
                     AND cyear + cperiod = m.hyear + m.hperiod AND 'TAX' $ cSource AND NOT EMPTY(cownerid)
                  DO CASE
                     CASE cSource = 'OTAX1'
                        lnOTax1 = lnOTax1 + nTotalInc
                     CASE cSource = 'OTAX2'
                        lnOTax2 = lnOTax2 + nTotalInc
                     CASE cSource = 'OTAX3'
                        lnOTax3 = lnOTax3 + nTotalInc
                     CASE cSource = 'OTAX4'
                        lnOTax4 = lnOTax4 + nTotalInc
                     CASE cSource = 'GTAX1'
                        lnGTax1 = lnGTax1 + nTotalInc
                     CASE cSource = 'GTAX2'
                        lnGTax2 = lnGTax2 + nTotalInc
                     CASE cSource = 'GTAX3'
                        lnGTax3 = lnGTax3 + nTotalInc
                     CASE cSource = 'GTAX4'
                        lnGTax4 = lnGTax4 + nTotalInc
                     CASE cSource = 'PTAX1'
                        lnPTax1 = lnPTax1 + nTotalInc
                     CASE cSource = 'PTAX2'
                        lnPTax2 = lnPTax2 + nTotalInc
                     CASE cSource = 'PTAX3'
                        lnPTax3 = lnPTax3 + nTotalInc
                     CASE cSource = 'PTAX4'
                        lnPTax4 = lnPTax4 + nTotalInc
                  ENDCASE
               ENDSCAN

               lnOTax1      = lnOTax1 * -1  &&  Since the numbers in the income table for taxes are negative, switch the sign before the netrev method
               lnOTax2      = lnOTax2 * -1
               lnOTax3      = lnOTax3 * -1
               lnOTax4      = lnOTax4 * -1
               lnGTax1      = lnGTax1 * -1
               lnGTax2      = lnGTax2 * -1
               lnGTax3      = lnGTax3 * -1
               lnGTax4      = lnGTax4 * -1
               lnPTax1      = lnPTax1 * -1
               lnPTax2      = lnPTax2 * -1
               lnPTax3      = lnPTax3 * -1
               lnPTax4      = lnPTax4 * -1
               m.cownerid   = ''
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax1 - lnOTax1, 'O1', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax1
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax2 - lnOTax2, 'O2', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax2
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax3 - lnOTax3, 'O3', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax3
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax4 - lnOTax4, 'O4', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax4
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax1 - lnGTax1, 'G1', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax1
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax2 - lnGTax2, 'G2', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax2
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax3 - lnGTax3, 'G3', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax3
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax4 - lnGTax4, 'G4', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax4
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax1 - lnPTax1, 'P1', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnPTax1
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax2 - lnPTax2, 'P2', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnPTax2
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax3 - lnPTax3, 'P3', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnPTax3
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax4 - lnPTax4, 'P4', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnPTax4
               lnSevTaxWell = lnSevTaxWell + m.ntotbbltxR + m.ntotmcftxR + m.ntotbbltxW + m.ntotmcftxW
            ENDSCAN

            STORE 0 TO m.nOwnChkCount, m.nVendChkCount
            STORE 0 TO m.nOwnChkAmt, m.nVendChkAmt
            STORE 0 TO m.nDefPrior, m.nDefCurr, m.nMinPrior, m.nMinCurr
            STORE 0 TO m.nHoldPrior, m.nHoldCurr, m.nbackwith, m.ntaxwith

*  Get the amount of vendor posting
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Vendor Amounts Posted...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            lnVendorPost = 0
            SELE vendor
            SELE cvendorid INTO CURSOR tempvend FROM vendor WHERE lIntegGL ORDER BY cvendorid
            IF _TALLY > 0
               SELE tempvend
               SCAN
                  m.cvendorid = cvendorid
                  SELE expense
                  SCAN FOR cRunYearRev = lcYear ;
                        AND nRunNoRev = lnRunNo ;
                        AND NOT lAPTran ;
                        AND cvendorid = m.cvendorid ;
                        AND cexpclass # 'P' ;
                        AND ccatcode # 'PLUG'
                     m.namount    = namount
                     m.cWellID    = cWellID
                     m.cexpclass  = cexpclass
                     m.cownerid   = cownerid
                     m.cdeck      = cdeck
                     lnVendorPost = lnVendorPost + swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'B', .F., m.cownerid, '', m.cdeck)
                  ENDSCAN
               ENDSCAN
            ENDIF
            swclose('tempvend')

            IF NOT EMPTY(lcVendComp)
               swselect('vendor')
               LOCATE FOR cvendorid = lcVendComp
               IF FOUND()
                  IF vendor.lIntegGL
                     SELECT wellwork
                     SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND cgroup = lcGroup AND crectype = 'R'
                        lnVendorPost = lnVendorPost + nCompress + nGather
                     ENDSCAN
                  ENDIF
               ENDIF
            ENDIF

*  Get the check counts and amounts
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Check Counts and Totals...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            swselect('sysctl')
            LOCATE FOR cyear = lcYear AND nrunno = lnRunNo AND cTypeClose = 'R'
            IF FOUND()
               lcDMBatch = cdmbatch
               swselect('checks')
               SCAN FOR cBatch = lcDMBatch AND cidtype = 'I' AND NOT 'DIRDEP' $ ccheckno AND NOT LEFT(ALLTRIM(ccheckno), 1) = 'E'
                  m.nOwnChkCount = m.nOwnChkCount + 1
                  m.nOwnChkAmt   = m.nOwnChkAmt + namount
               ENDSCAN
               SCAN FOR cBatch = lcDMBatch AND cidtype = 'V'
                  m.nVendChkCount = m.nVendChkCount + 1
                  m.nVendChkAmt   = m.nVendChkAmt + namount
               ENDSCAN
            ENDIF

*  Get the suspense Amounts
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Closing Summary Page: Summing Suspense Amounts...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

* Get the suspense types per owner before this run
            THIS.osuspense.GetLastType(.F., .T., THIS.cgroup, .T.)

* Make sure the order is set for the investor table
            swselect('investor')
            SET ORDER TO cownerid

* Get deficits not covered this run
            m.nDefCurr = 0

            SELECT tsuspense
            SCAN FOR crunyear_in == THIS.crunyear ;
                  AND nrunno_in == THIS.nrunno ;
                  AND csusptype = 'D' ;
                  AND (NOT lManual OR (lManual AND crectype = 'P'))
               m.cownerid  = cownerid
               m.nnetcheck = nnetcheck
               swselect('investor')
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               SELECT tsuspense
               m.nDefCurr = m.nDefCurr + m.nnetcheck
            ENDSCAN

* Get prior deficits covered this run

            m.nDefPrior = 0
            SELECT disbhist
            SCAN FOR crunyear == THIS.crunyear ;
                  AND nrunno == THIS.nrunno ;
                  AND csusptype <> ' ' ;
                  AND csusptype = 'D'
               m.cownerid  = cownerid
               m.cWellID   = cWellID
               m.nnetcheck = nnetcheck
               swselect('investor')
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               SELECT curLastSuspType
               LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID
               IF FOUND()
                  IF csusptype # 'D'
                     LOOP
                  ENDIF
               ENDIF
               m.nDefPrior = m.nDefPrior + m.nnetcheck
            ENDSCAN

* Get current minimum amounts held

            m.nMinCurr = 0
            SELECT tsuspense
            SCAN FOR (nrunno_in = lnRunNo AND crunyear_in = lcYear) ;
                  AND csusptype = 'M' ;
                  AND (NOT lManual OR (lManual AND crectype = 'P'))
               m.cownerid  = cownerid
               m.nnetcheck = nnetcheck
               swselect('investor')
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               m.nMinCurr = m.nMinCurr + m.nnetcheck
            ENDSCAN

* Get prior minimum amounts released

            m.nMinPrior = 0
            SELECT disbhist
            SCAN FOR crunyear == THIS.crunyear ;
                  AND nrunno == THIS.nrunno ;
                  AND csusptype <> ' ' ;
                  AND csusptype = 'M'
               m.cownerid  = cownerid
               m.cWellID   = cWellID
               m.nnetcheck = nnetcheck
               swselect('investor')
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               SELECT curLastSuspType
               LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID
               IF FOUND()
                  IF csusptype = 'M'
                     m.nMinPrior = m.nMinPrior + m.nnetcheck
                  ENDIF
               ENDIF
            ENDSCAN

* Get current owners on hold

            m.nHoldCurr = 0
            SELECT tsuspense
            SCAN FOR crunyear_in == THIS.crunyear ;
                  AND nrunno_in == THIS.nrunno ;
                  AND INLIST(csusptype, 'H', 'Q', 'S', 'A', 'I') ;
                  AND (NOT lManual OR (lManual AND crectype = 'P'))
               m.cownerid = cownerid
               swselect('investor')
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               SELECT tsuspense
               m.nHoldCurr = m.nHoldCurr + nnetcheck
            ENDSCAN

* Get prior owners on hold released

            m.nHoldPrior = 0
            SELECT disbhist
            SCAN FOR crunyear == THIS.crunyear AND nrunno == THIS.nrunno ;
                  AND csusptype <> ' '
               m.cownerid  = cownerid
               m.cWellID   = cWellID
               m.nnetcheck = nnetcheck
               swselect('investor')
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               SELECT curLastSuspType
               LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID
               IF FOUND()
                  IF INLIST(csusptype, 'H', 'Q', 'S', 'A', 'I')
                     m.nHoldPrior = m.nHoldPrior + m.nnetcheck
                  ENDIF
               ENDIF
            ENDSCAN

            lcLastRun     = GetLastRun(THIS.cnewrunyear, THIS.nnewrunno, THIS.cgroup, 'R')
            lcLastRunYear = LEFT(lcLastRun, 4)
            lnLastRunNo   = INT(VAL(RIGHT(lcLastRun, 3)))

            m.nDefBefore = THIS.osuspense.SuspBalByRun('ODEF', THIS.cbegownerid,;
                 THIS.cendownerid,;
                 lcLastRunYear,;
                 lnLastRunNo,;
                 THIS.cgroup,;
                 THIS.dacctdate, '')
* Get prior deficit balance
*            m.nDefBefore = THIS.osuspense.Suspense_Balance('D', .F., .T., .T.)

* Get prior minimum balance
            m.nMinBefore = THIS.osuspense.SuspBalByRun('OSUSP', THIS.cbegownerid,;
                 THIS.cendownerid,;
                 lcLastRunYear,;
                 lnLastRunNo,;
                 THIS.cgroup,;
                 THIS.dacctdate, '')
*            m.nMinBefore = THIS.osuspense.Suspense_Balance('M', .F., .T., .T.)

* Get current deficit balance after this run
            m.nDefAfter = THIS.osuspense.SuspBalByRun('ODEF', THIS.cbegownerid,;
                 THIS.cendownerid,;
                 THIS.cnewrunyear,;
                 THIS.nnewrunno,;
                 THIS.cgroup,;
                 THIS.dacctdate, '')
*            m.nDefAfter = THIS.osuspense.Suspense_Balance('D', .T., .T., .T.)

* Get current minimum balance after this run
            m.nMinAfter = THIS.osuspense.SuspBalByRun('OSUSP', THIS.cbegownerid,;
                 THIS.cendownerid,;
                 THIS.cnewrunyear,;
                 THIS.nnewrunno,;
                 THIS.cgroup,;
                 THIS.dacctdate, '')
*            m.nMinAfter = THIS.osuspense.Suspense_Balance('M', .T., .T., .T.)

            IF NOT m.goApp.lAMVersion
* Get the deficits transfering to minimums
               lnDefSwitch = THIS.osuspense.GetBalTransfer('D')

* Get the amount of minimums that transferred
               lnMinSwitch = THIS.osuspense.GetBalTransfer('M') * -1

               m.ndeftransfer = lnDefSwitch
               m.nmintransfer = lnMinSwitch
            ELSE
               m.ndeftransfer = THIS.ndeftransfer
               m.nmintransfer = THIS.nmintransfer
            ENDIF

            lnTransfer     = m.ndeftransfer + m.nmintransfer
            m.ndeftransfer = lnTransfer * -1
            m.nmintransfer = lnTransfer

* add in suspense to the temphist cursor
            swselect('suspense')
            SELECT  * ;
                FROM suspense WITH (BUFFERING = .T.) ;
                WHERE (nrunno_in = lnRunNo ;
                      AND crunyear_in = lcYear) ;
                    AND crectype = 'R' ;
                INTO CURSOR temphist1
            SELECT temphist
            APPEND FROM DBF('temphist1')
            swclose('temphist1')

* Get backup and tax withholding amounts
            SELE investor
            COUNT FOR lbackwith TO lnTaxWith
            SELE wellinv
            COUNT FOR ntaxpct # 0 TO lnpctcnt
            lnTaxWith = lnTaxWith + lnpctcnt

            STORE 0 TO lnBackupWH, lnTaxWH, lnPlugging

            IF lnTaxWith > 0
               IF NOT THIS.lclose
                  WAIT WIND NOWAIT 'Summing Tax Withholding Totals...'
               ENDIF
               SELECT temphist
               SCAN FOR (nrunno = lnRunNo AND crunyear = lcYear) AND crectype = 'R' AND (ntaxwith # 0 OR nbackwith # 0)
                  m.cownerid = cownerid
                  m.nTax     = ntaxwith
                  m.nBack    = nbackwith
                  swselect('investor')
                  SET ORDER TO cownerid
                  IF SEEK(m.cownerid) AND investor.ldummy
                     LOOP
                  ENDIF
                  IF NOT EMPTY(temphist.csusptype)
* Only include suspense entries if their run equals this run
                     IF temphist.nrunno_in = lnRunNo AND temphist.crunyear_in = lcYear
                        lnTaxWH    = lnTaxWH + m.nTax
                        lnBackupWH = lnBackupWH + m.nBack
                     ENDIF
                  ELSE
                     lnTaxWH    = lnTaxWH + m.nTax
                     lnBackupWH = lnBackupWH + m.nBack
                  ENDIF
               ENDSCAN
            ENDIF

            IF NOT THIS.lclose
               WAIT WIND NOWAIT 'Summing Plugging Charges...'
            ENDIF
            SELECT temphist
            SCAN FOR (nrunno = lnRunNo AND crunyear = lcYear) AND crectype = 'R' AND nPlugExp # 0
               m.cownerid = cownerid
               m.nPlugExp = nPlugExp
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               IF NOT EMPTY(temphist.csusptype)
* Only include suspense entries if their run equals this run
                  IF temphist.nrunno_in = lnRunNo AND temphist.crunyear_in = lcYear
                     lnPlugging = lnPlugging + m.nPlugExp
                  ENDIF
               ELSE
                  lnPlugging = lnPlugging + m.nPlugExp
               ENDIF
            ENDSCAN

* Get the amount of partnership Posting
            lnPartnerPost = 0
            IF m.goApp.lPartnershipMod
               lnPartnerPost = THIS.oPartnerShip.GetPostAmount(THIS.cdmbatch)
            ENDIF

            m.nRevEntered    = lnRevEnt
            m.nRevAllocated  = lnRevAll
            m.nExpEntered    = lnExpEnt
            m.nExpAllocated  = lnExpAll
            m.nSevTaxWell    = lnSevTaxWell
            m.nSevTaxOwn     = lnSevTaxOwn
            m.nJExpAllocated = lnJIBAll
            m.nJIBInvCount   = lnJIBInv
            m.nJIBInvAmount  = lnJibInvAmt
            m.cRevProdPrds   = lcRevProdPrds
            m.cExpProdPrds   = lcExpProdPrds
            m.nDirectDep     = lnDirectDep
            m.nDirectDepCnt  = lnDirectDepCnt
            m.nOwnerPost     = lnOwnerPost
            m.nVendorPost    = lnVendorPost
            m.nGathering     = lnGathering
            m.nCompression   = lnCompression
            m.nflatrate      = lnFlatRate
            m.nbackwith      = lnBackupWH
            m.ntaxwith       = lnTaxWH
            m.nPluggingFund  = lnPlugging
            m.nPartnerPost   = lnPartnerPost

            lnTime           = ROUND((DATETIME() - THIS.nseconds) / 60, 2)
            IF lnTime > 60
               lnHours   = INT(lnTime / 60)
               lnMinutes = INT(MOD(lnTime, 60))
               IF lnHours > 1
                  lcHours = ' hours and '
               ELSE
                  lcHours = ' hour and '
               ENDIF
               m.cRunTime = 'Run Closed In: ' + TRANSFORM(lnHours) + lcHours + TRANSFORM(lnMinutes) + ' minutes'
            ELSE
               IF lnTime > 1
                  lnTime     = INT(lnTime)
                  m.cRunTime = 'Run Closed In: ' + TRANSFORM(lnTime, '99') + ' minutes...'
               ELSE
                  IF lnTime < 1
                     lnTime     = INT(MOD(lnTime, 60) * 60)
                     m.cRunTime = 'Run Closed In: ' + TRANSFORM(lnTime, '99') + ' seconds...'
                  ELSE
                     lnTime = INT(lnTime)
                     IF lnTime = 1
                        m.cRunTime       = 'Run Closed In: ' + TRANSFORM(lnTime, '99') + ' minute...'
                     ELSE
                        m.cRunTime       = 'Run Closed In: ' + TRANSFORM(lnTime, '99') + ' minutes...'
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF
            WAIT CLEAR

* Insert into the runclose table so we can report on it later
            IF THIS.lclose
               swselect('runclose')
               m.cdmbatch = THIS.cdmbatch
               INSERT INTO runclose FROM MEMVAR
            ENDIF

            SET SAFETY OFF
            SELE closetemp
            ZAP
            m.nDirectDe2 = m.nDirectDepCnt
            INSERT INTO closetemp FROM MEMVAR
            SELECT closetemp

            IF THIS.lclose
               THIS.oprogress.CloseProgress()
               THIS.oprogress = .NULL.
            ENDIF
         ELSE
            SET SAFETY OFF
            SELECT closetemp
            ZAP

            swselect('runclose')
            SCAN FOR cdmbatch == THIS.cdmbatch
               SCATTER MEMVAR
 
                  m.nDirectDe2 = m.nDirectDepCnt
  
               INSERT INTO closetemp FROM MEMVAR
            ENDSCAN
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CalcSummary', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CalcSummary', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Calculates a closing summary by well.
*********************************
   PROCEDURE CalcSumByWell
*********************************
      LPARA tlExceptions

      LOCAL lcYear, lcPeriod, lcGroup, lnRevEnt, lnExpEnt
      LOCAL lnRevAll, lnExpAll, lcTitle1, lcDMBatch, llSepClose
      LOCAL lnJIBAll, lnJIBInv, lnJibInvAmt, llExceptions
      LOCAL lDirGasPurch, lDirOilPurch, lExempt, lcTitle2, llReturn, lnGTax1, lnGTax2, lnGTax3, lnGTax4
      LOCAL lnIncome, lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnPTax1, lnPTax2, lnPTax3, lnPTax4, lnRunNo
      LOCAL lnSevTaxOwn, lnSevTaxWell, loError
      LOCAL cGrpName, cownerid, cProcessor, cProducer, cwellname, glGrpName, nExpAllocated, nExpEntered
      LOCAL nJExpAllocated, nRevAllocated, nRevEntered, nSevTaxOwn, nSevTaxWell

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF



         lcYear  = THIS.crunyear
         lcGroup = THIS.cgroup
         lnRunNo = THIS.nrunno

         STORE 0 TO lnJIBAll, lnJIBInv, lnJibInvAmt

         SELECT cWellID FROM wells WHERE cwellstat # 'A' INTO CURSOR inactwell

         glGrpName  = THIS.oOptions.lGrpName
         llSepClose = .T.
         IF glGrpName
            swselect('groups')
            SET ORDER TO cgroup
            IF SEEK(lcGroup)
               m.cGrpName = cDesc
            ELSE
               IF lcGroup = '**'
                  m.cGrpName = 'All Companies'
               ELSE
                  m.cGrpName = ''
               ENDIF
            ENDIF
         ELSE
            m.cGrpName = ''
         ENDIF

         lcTitle1 = 'Revenue Closing Summary By Well'

         lcTitle2 = 'For Run No ' + ALLT(STR(lnRunNo)) + '/' + lcYear + ' Group ' + lcGroup

         IF TYPE('m.goApp') = 'O'
            m.cProducer = m.goApp.cCompanyName
         ELSE
            m.cProducer = 'Development Company'
         ENDIF

         m.cProcessor = ''

*  Builds the closing summary information
         CREATE CURSOR tempclose ;
            (cWellID        c(10), ;
              cwellname      c(30), ;
              nRevEntered    N(12, 2), ;
              nRevAllocated  N(12, 2), ;
              nExpEntered    N(12, 2), ;
              nExpAllocated  N(12, 2), ;
              nJExpAllocated N(12, 2), ;
              nSevTaxWell    N(9, 2), ;
              nSevTaxOwn     N(9, 2))
         INDEX ON cWellID TAG cWellID

         CREATE CURSOR closetmp ;
            (cWellID        c(10), ;
              cwellname      c(30), ;
              nRevEntered    N(12, 2), ;
              nRevAllocated  N(12, 2), ;
              nExpEntered    N(12, 2), ;
              nExpAllocated  N(12, 2), ;
              nJExpAllocated N(12, 2), ;
              nSevTaxWell    N(9, 2), ;
              nSevTaxOwn     N(9, 2))
         INDEX ON cWellID TAG cWellID

         swselect('wells')
         SET ORDER TO cWellID

         m.nRevEntered    = 0
         m.nRevAllocated  = 0
         m.nExpEntered    = 0
         m.nExpAllocated  = 0
         m.nSevTaxWell    = 0
         m.nSevTaxOwn     = 0
         m.nJExpAllocated = 0

         lnRevEnt = 0


*  Get the revenue and expenses entered
         swselect('income')
         SET ORDER TO 0
         SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND NOT 'TAX' $ cSource
            SCATTER MEMVAR

            swselect('wells')
            IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess # 2
               m.cwellname = cwellname
               DO CASE
                  CASE m.cSource = 'BBL'
                     m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'O', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                  CASE m.cSource = 'MCF'
                     m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'G', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                  CASE m.cSource = 'TRANS'
                     m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'T', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                  CASE m.cSource = 'MISC1'
                     m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, '1', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                  CASE m.cSource = 'MISC2'
                     m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, '2', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                  CASE m.cSource = 'OTH'
                     m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'P', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                  CASE INLIST(m.cSource, 'COMP', 'GATH')
* Ignore
                  OTHERWISE
                     IF NOT 'TAX' $ m.cSource
                        m.nRevEntered = m.nTotalInc
                     ELSE
                        LOOP
                     ENDIF
               ENDCASE

               IF m.nRevEntered # 0
                  INSERT INTO closetmp FROM MEMVAR
               ENDIF
            ENDIF
         ENDSCAN

         STORE 0 TO m.nRevEntered

*  Get the revenue and expenses entered for quarterly wells
         IF THIS.lrelqtr
            swselect('income')
            SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND NOT 'TAX' $ cSource
               SCATTER MEMVAR
               swselect('wells')
               IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess = 2
                  m.cwellname = cwellname
                  DO CASE
                     CASE m.cSource = 'BBL'
                        m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'O', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'MCF'
                        m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'G', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'MISC1'
                        m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, '1', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'MISC2'
                        m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, '2', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'TRANS'
                        m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'T', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE m.cSource = 'OTH'
                        m.nRevEntered = swnetrev(m.cWellID, m.nTotalInc, 'P', .F., .T., .F., m.cownerid, .F., .F., .F., m.cdeck)
                     CASE INLIST(m.cSource, 'COMP', 'GATH')
* Ignore
                     OTHERWISE
                        IF NOT 'TAX' $ m.cSource
                           m.nRevEntered = m.nTotalInc
                        ELSE
                           LOOP
                        ENDIF
                  ENDCASE

                  IF m.nRevEntered # 0
                     INSERT INTO closetmp FROM MEMVAR
                  ENDIF
               ENDIF
            ENDSCAN
         ENDIF

         lnExpEnt         = 0
         m.nRevEntered    = 0
         m.nRevAllocated  = 0
         m.nExpEntered    = 0
         m.nExpAllocated  = 0
         m.nSevTaxWell    = 0
         m.nSevTaxOwn     = 0
         m.nJExpAllocated = 0

         swselect('expense')
         SCAN FOR ((nRunNoRev = lnRunNo AND cRunYearRev = lcYear) OR ;
                 (nrunnojib = lnRunNo AND crunyearjib = lcYear AND INLIST(cRunYearRev, '1900', '1901'))) AND ;
               NOT INLIST(ccatcode, 'GATH', 'COMP')
            SCATTER MEMVAR


            lnExpEnt         = 0
            m.nRevEntered    = 0
            m.nRevAllocated  = 0
            m.nExpEntered    = 0
            m.nExpAllocated  = 0
            m.nSevTaxWell    = 0
            m.nSevTaxOwn     = 0
            m.nJExpAllocated = 0

            swselect('wells')
            IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess # 2
               m.cwellname = cwellname
               IF NOT EMPTY(m.cownerid)
                  swselect('investor')
                  LOCATE FOR cownerid == m.cownerid  &&  If it's been allocated to the dummy, ignore it
                  IF FOUND() AND NOT investor.ldummy
                     SELE wellinv
                     LOCATE FOR cownerid = m.cownerid AND cWellID = m.cWellID AND ctypeinv = 'W'
                     IF FOUND() AND lJIB
                        IF llSepClose
                           m.nJExpAllocated = m.namount
                           m.nExpEntered    = m.namount
                        ENDIF
                     ELSE
                        m.nExpEntered = m.namount
                     ENDIF
                  ENDIF
               ELSE
                  IF m.ccatcode = 'MKTG'
                     m.nExpEntered = swnetrev(m.cWellID, m.namount, 'G', .F., .T., .T., .F., .F., .F., .F., m.cdeck)
                  ELSE
                     m.nExpEntered    = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, m.ccatcode, m.cdeck)
                     m.nJExpAllocated = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'J', .F., m.cownerid, m.ccatcode, m.cdeck)
                  ENDIF
               ENDIF
               IF m.nJExpAllocated # 0 OR m.nExpEntered # 0
                  INSERT INTO closetmp FROM MEMVAR
               ENDIF
            ENDIF
            swselect('expense')
         ENDSCAN

         IF THIS.lrelqtr
            swselect('expense')
            SCAN FOR ((nRunNoRev = lnRunNo AND cRunYearRev = lcYear) OR ;
                    (nrunnojib = lnRunNo AND crunyearjib = lcYear AND INLIST(cRunYearRev, '1900', '1901'))) AND NOT ;
                  INLIST(ccatcode, 'GATH', 'COMP')
               SCATTER MEMVAR
               swselect('wells')
               IF SEEK(m.cWellID) AND cgroup = lcGroup AND nprocess = 2
                  IF NOT EMPTY(m.cownerid)
                     swselect('investor')
                     LOCATE FOR cownerid == m.cownerid  &&  If it's been allocated to the dummy, ignore it
                     IF FOUND() AND NOT investor.ldummy
                        SELE wellinv
                        LOCATE FOR cownerid = m.cownerid AND cWellID = m.cWellID AND ctypeinv = 'W'
                        IF FOUND() AND lJIB
                           IF llSepClose
                              m.nJExpAllocated = m.namount
                              m.nExpEntered    = m.nExpEntered + m.namount
                           ENDIF
                        ELSE
                           m.nExpEntered = m.namount
                        ENDIF
                     ENDIF
                  ELSE
                     m.cwellname = cwellname
                     IF llSepClose
                        m.nExpEntered = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, m.ccatcode, m.cdeck,)
                        IF NOT INLIST(m.ccatcode, 'MKTG', 'COMP', 'GATH')
                           m.nJExpAllocated = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'J', .F., m.cownerid, m.ccatcode, m.cdeck)
                        ENDIF
                     ELSE
                        IF m.ccatcode = 'MKTG'
                           m.nExpEntered = swnetrev(m.cWellID, m.namount, 'G', .F., .T., .T., .F., .F., .F., .F., m.cdeck)
                        ELSE
                           m.nExpEntered    = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, m.ccatcode, m.cdeck)
                           m.nJExpAllocated = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'J', .F., m.cownerid, m.ccatcode, m.cdeck)
                        ENDIF
                     ENDIF
                  ENDIF
                  IF m.nJExpAllocated # 0 OR m.nExpEntered # 0
                     INSERT INTO closetmp FROM MEMVAR
                  ENDIF
               ENDIF
            ENDSCAN
         ENDIF

*  Get the revenue and expenses allocated
         lnRevAll         = 0
         lnExpAll         = 0
         lnSevTaxOwn      = 0
         m.nRevEntered    = 0
         m.nRevAllocated  = 0
         m.nExpEntered    = 0
         m.nExpAllocated  = 0
         m.nSevTaxWell    = 0
         m.nSevTaxOwn     = 0
         m.nJExpAllocated = 0

         SELECT  * ;
             FROM disbhist WITH (BUFFERING = .T.) ;
             WHERE ((nrunno = lnRunNo ;
                     AND crunyear = lcYear ;
                     AND csusptype = ' ') ;
                   OR (nrunno_in = lnRunNo ;
                     AND crunyear_in = lcYear ;
                     AND nrunno # 9999)) ;
                 AND crectype = 'R' ;
             INTO CURSOR temphist READWRITE

         SELECT temphist
         SET ORDER TO 0
         SCAN FOR NOT lManual
            SCATTER MEMVAR
            m.lExempt = .F.
            swselect('investor')
            SET ORDER TO cownerid
            IF SEEK(m.cownerid)
               m.lExempt = lExempt
               IF investor.ldummy
                  LOOP
               ENDIF
            ENDIF
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               m.lDirOilPurch = lDirOilPurch
               m.lDirGasPurch = lDirGasPurch
               SCATTER FIELDS LIKE lSev* MEMVAR
               m.cwellname = cwellname
            ELSE
               LOOP
            ENDIF

            lnIncome        = m.nIncome
            lnIncome        = lnIncome - IIF(m.cdirect = 'O', m.noilrev, 0)
            lnIncome        = lnIncome - IIF(m.cdirect = 'G', m.ngasrev, 0)
            lnIncome        = lnIncome - IIF(m.cdirect = 'B', m.noilrev + m.ngasrev, 0)
            m.nRevAllocated = lnIncome
            m.nExpAllocated = m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + ;
               m.ntotalea + m.ntotaleb + m.nMKTGExp

            IF m.lExempt  && m.lexempt
               m.nSevTaxOwn = 0
            ELSE
               DO CASE
                  CASE m.cdirect = 'O' OR m.cdirect = 'B'
                     IF m.lDirOilPurch
                        m.nSevTaxOwn = m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ELSE
                        m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ENDIF
                  CASE m.cdirect = 'G' OR m.cdirect = 'B'
                     IF m.lDirGasPurch
                        m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ELSE
                        m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ENDIF
                  CASE m.cdirect = 'N'
                     m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
               ENDCASE
            ENDIF

            IF m.nRevAllocated # 0 OR m.nExpAllocated # 0 OR m.nSevTaxOwn # 0
               INSERT INTO closetmp FROM MEMVAR
            ENDIF
         ENDSCAN

         SELECT  * ;
             FROM suspense WITH (BUFFERING = .T.) ;
             WHERE nrunno_in = lnRunNo ;
                 AND crunyear_in = lcYear ;
                 AND NOT lManual ;
             INTO CURSOR temphist1
         SELECT temphist1
         SCAN
            SCATTER MEMVAR

            m.cownerid = cownerid
            m.lExempt  = .F.
            swselect('investor')
            SET ORDER TO cownerid
            IF SEEK(m.cownerid)
               m.lExempt = lExempt
               IF investor.ldummy
                  LOOP
               ENDIF
            ENDIF
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               m.lDirOilPurch = lDirOilPurch
               m.lDirGasPurch = lDirGasPurch
               SCATTER FIELDS LIKE lSev* MEMVAR
               m.cwellname = cwellname
            ELSE
               LOOP
            ENDIF
            lnIncome        = m.nIncome
            lnIncome        = lnIncome - IIF(m.cdirect = 'O', m.noilrev, 0)
            lnIncome        = lnIncome - IIF(m.cdirect = 'G', m.ngasrev, 0)
            lnIncome        = lnIncome - IIF(m.cdirect = 'B', m.noilrev + m.ngasrev, 0)
            m.nRevAllocated = lnIncome
            m.nExpAllocated = m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + ;
               m.ntotalea + m.ntotaleb + m.nMKTGExp

            IF m.lExempt  && m.lexempt
               m.nSevTaxOwn = 0
            ELSE
               DO CASE
                  CASE m.cdirect = 'O' OR m.cdirect = 'B'
                     IF m.lDirOilPurch
                        m.nSevTaxOwn = m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ELSE
                        m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ENDIF
                  CASE m.cdirect = 'G' OR m.cdirect = 'B'
                     IF m.lDirGasPurch
                        m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ELSE
                        m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
                     ENDIF
                  CASE m.cdirect = 'N'
                     m.nSevTaxOwn = m.noiltax1 + m.noiltax2 + m.noiltax3 + m.noiltax4 + m.ngastax1 + m.ngastax2 + m.ngastax3 + m.ngastax4 + m.nOthTax1 + m.nOthTax2 + m.nOthTax3 + m.nOthTax4
               ENDCASE
            ENDIF

            IF m.nRevAllocated # 0 OR m.nExpAllocated # 0 OR m.nSevTaxOwn # 0
               SELECT closetmp
               LOCATE FOR cWellID == m.cWellID
               IF FOUND()
                  REPLACE nRevAllocated WITH nRevAllocated + m.nRevAllocated, ;
                          nExpAllocated WITH nExpAllocated + m.nExpAllocated, ;
                          nSevTaxOwn    WITH nSevTaxOwn    + m.nSevTaxOwn
               ELSE
                  INSERT INTO closetmp FROM MEMVAR
               ENDIF
            ENDIF

         ENDSCAN


         m.nRevEntered    = 0
         m.nRevAllocated  = 0
         m.nExpEntered    = 0
         m.nExpAllocated  = 0
         m.nSevTaxWell    = 0
         m.nSevTaxOwn     = 0
         m.nJExpAllocated = 0

         IF NOT llSepClose
* Get the JIB expenses allocated
            swselect('disbhist')
            SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND crectype = 'J'
               SCATTER MEMVAR
               swselect('wells')
               IF SEEK(m.cWellID)
                  m.cwellname = cwellname
               ENDIF
*  Net out "Dummy" owners
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(m.cownerid) AND investor.ldummy
                  LOOP
               ENDIF
               swselect('wells')
               IF SEEK(m.cWellID)
                  m.cwellname = cwellname
               ENDIF
               m.nJExpAllocated = m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5
               INSERT INTO closetmp FROM MEMVAR
            ENDSCAN
         ENDIF

         m.nRevEntered    = 0
         m.nRevAllocated  = 0
         m.nExpEntered    = 0
         m.nExpAllocated  = 0
         m.nSevTaxWell    = 0
         m.nSevTaxOwn     = 0
         m.nJExpAllocated = 0

*  Get the severance taxes allocated to the well
         lnSevTaxWell = 0
         SELECT wellhist
         SCAN FOR nrunno = lnRunNo AND crunyear = lcYear AND cgroup = lcGroup AND crectype = 'R'
            SCATTER MEMVAR
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               m.lDirOilPurch = lDirOilPurch
               m.lDirGasPurch = lDirGasPurch
               SCATTER FIELDS LIKE lSev* MEMVAR
               STORE 0 TO lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnGTax1, lnGTax2, lnGTax3, lnGTax4,  ;
                  lnPTax1, lnPTax2, lnPTax3, lnPTax4

               swselect('income')  &&  Total up the one-man tax entries, so they can be subtracted off before doing the netrev
               SCAN FOR cWellID == m.cWellID AND nrunno = m.nrunno AND crunyear = m.crunyear  ;
                     AND cyear + cperiod = m.hyear + m.hperiod AND 'TAX' $ cSource AND NOT EMPTY(cownerid)
                  DO CASE
                     CASE cSource = 'OTAX1'
                        lnOTax1 = lnOTax1 + nTotalInc
                     CASE cSource = 'OTAX2'
                        lnOTax2 = lnOTax2 + nTotalInc
                     CASE cSource = 'OTAX3'
                        lnOTax3 = lnOTax3 + nTotalInc
                     CASE cSource = 'OTAX4'
                        lnOTax4 = lnOTax4 + nTotalInc
                     CASE cSource = 'GTAX1'
                        lnGTax1 = lnGTax1 + nTotalInc
                     CASE cSource = 'GTAX2'
                        lnGTax2 = lnGTax2 + nTotalInc
                     CASE cSource = 'GTAX3'
                        lnGTax3 = lnGTax3 + nTotalInc
                     CASE cSource = 'GTAX4'
                        lnGTax4 = lnGTax4 + nTotalInc
                     CASE cSource = 'PTAX1'
                        lnPTax1 = lnPTax1 + nTotalInc
                     CASE cSource = 'PTAX2'
                        lnPTax2 = lnPTax2 + nTotalInc
                     CASE cSource = 'PTAX3'
                        lnPTax3 = lnPTax3 + nTotalInc
                     CASE cSource = 'PTAX4'
                        lnPTax4 = lnPTax4 + nTotalInc
                  ENDCASE
               ENDSCAN

               lnOTax1 = lnOTax1 * -1  &&  Since the numbers in the income table for taxes are negative, switch the sign before the netrev method
               lnOTax2 = lnOTax2 * -1
               lnOTax3 = lnOTax3 * -1
               lnOTax4 = lnOTax4 * -1
               lnGTax1 = lnGTax1 * -1
               lnGTax2 = lnGTax2 * -1
               lnGTax3 = lnGTax3 * -1
               lnGTax4 = lnGTax4 * -1
               lnPTax1 = lnPTax1 * -1
               lnPTax2 = lnPTax2 * -1
               lnPTax3 = lnPTax3 * -1
               lnPTax4 = lnPTax4 * -1

               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax1 - lnOTax1, 'O1', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax1
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax2 - lnOTax2, 'O2', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax2
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax3 - lnOTax3, 'O3', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax3
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGBBLTax4 - lnOTax4, 'O4', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnOTax4
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax1 - lnGTax1, 'G1', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax1
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax2 - lnGTax2, 'G2', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax2
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax3 - lnGTax3, 'G3', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax3
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGMCFTax4 - lnGTax4, 'G4', .F., .T., .F., .F., .T., .F., .F., m.cdeck) + lnGTax4
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax1 - lnPTax1, 'P1', .F., .T., .F., .F., .T., .F., .F., m.cdeck)  + lnPTax1
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax2 - lnPTax2, 'P1', .F., .T., .F., .F., .T., .F., .F., m.cdeck)  + lnPTax2
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax3 - lnPTax3, 'P1', .F., .T., .F., .F., .T., .F., .F., m.cdeck)  + lnPTax3
               lnSevTaxWell = lnSevTaxWell + swnetrev(m.cWellID, m.nGOTHTax4 - lnPTax4, 'P1', .F., .T., .F., .F., .T., .F., .F., m.cdeck)  + lnPTax4
               lnSevTaxWell = lnSevTaxWell + m.ntotbbltxR + m.ntotmcftxR + m.ntotbbltxW + m.ntotmcftxW
            ENDIF
            m.nSevTaxWell = lnSevTaxWell
            INSERT INTO closetmp FROM MEMVAR
            lnSevTaxWell = 0
         ENDSCAN

         SELECT  closetmp.cWellID, ;
                 wells.cwellname, ;
                 SUM(nRevEntered) AS nRevEntered, ;
                 SUM(nRevAllocated) AS nRevAllocated, ;
                 SUM(nExpEntered) AS nExpEntered, ;
                 SUM(nExpAllocated) AS nExpAllocated, ;
                 SUM(nSevTaxWell) AS nSevTaxWell, ;
                 SUM(nSevTaxOwn)  AS nSevTaxOwn, ;
                 SUM(nJExpAllocated) AS nJExpAllocated ;
             FROM closetmp,;
                 wells ;
             WHERE closetmp.cWellID = wells.cWellID ;
             INTO CURSOR temp ;
             ORDER BY closetmp.cWellID ;
             GROUP BY closetmp.cWellID

         IF tlExceptions
            SELECT temp
            SCAN
               SCATTER MEMVAR
               IF m.nRevEntered - m.nRevAllocated <= .50 AND m.nRevEntered - m.nRevAllocated >= - .50
                  IF m.nExpEntered - m.nExpAllocated - m.nJExpAllocated <= .50 AND m.nExpEntered - m.nExpAllocated -  m.nJExpAllocated >= - .50
                     IF m.nSevTaxWell - m.nSevTaxOwn <= .50 AND m.nSevTaxWell - m.nSevTaxOwn >= - .50
                        LOOP
                     ENDIF
                  ENDIF
               ENDIF
               INSERT INTO tempclose FROM MEMVAR
            ENDSCAN
         ELSE
            SELECT temp
            SCAN
               SCATTER MEMVAR
               IF m.nRevEntered # 0 OR m.nExpEntered # 0
                  INSERT INTO tempclose FROM MEMVAR
               ENDIF
            ENDSCAN
         ENDIF
         WAIT CLEAR

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CalcSumByWell', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CalcSumByWell', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()



      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE TaxExempt
*********************************
      LPARA tladdback
      LOCAL lnoiltax1, lnoiltax2, lnoiltax3, lnoiltax4
      LOCAL lngastax1, lngastax2, lngastax3, lngastax4
      LOCAL lnothtax1, lnothtax2, lnothtax3, lnothtax4
      LOCAL llReturn, llroyaltyowner, lnrevtax1, lnrevtax10, lnrevtax11, lnrevtax12, lnrevtax2, lnrevtax3
      LOCAL lnrevtax4, lnrevtax5, lnrevtax6, lnrevtax7, lnrevtax8, lnrevtax9, loError, lusesev
      LOCAL nrevgtax, nrevotax, nroyint

*  If tlAddBack is True, add the exempt owner taxes back into wellwork
*  otherwise, remove them.

      llReturn = .T.
      TRY

         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

* Look to see if there are any exempt owners. If not we'll bail out
         swselect('investor')
         SELECT  * ;
             FROM investor ;
             INTO CURSOR exemptowner READWRITE ;
             WHERE lExempt = .T. ;
                 AND cownerid IN (SELECT  cownerid ;
                                      FROM wellinv)

         IF _TALLY = 0
            llReturn = .T.
            EXIT
         ENDIF

         STORE 0 TO lnoiltax1, lnoiltax2, lnoiltax3, lnoiltax4
         STORE 0 TO lngastax1, lngastax2, lngastax3, lngastax4
         STORE 0 TO lnothtax1, lnothtax2, lnothtax3, lnothtax4

         IF tladdback = .F.
            SELECT wellwork
            SCAN
               SCATTER MEMVAR
               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  m.lusesev = lusesev
                  m.nroyint = nlandpct + noverpct
               ELSE
                  LOOP
               ENDIF
               swselect('wellinv')
               SCAN FOR cWellID = m.cWellID
                  SCATTER MEMVAR
                  IF INLIST(m.ctypeinv, 'L', 'O')
                     llroyaltyowner = .T.
                  ELSE
                     llroyaltyowner = .F.
                  ENDIF

                  m.nrevotax = (m.nrevoil / m.nroyint) * 100
                  m.nrevgtax = (m.nrevgas / m.nroyint) * 100

                  swselect('investor')
                  SET ORDER TO cownerid
                  IF SEEK(m.cownerid) AND lExempt
                     IF INLIST(m.ctypeint, 'B', 'O')
                        IF m.lusesev  && Use state severance tax table rates
                           IF llroyaltyowner
                              lnoiltax1  = (ROUND(m.ntotbbltxR * (m.nrevotax / 100), 2))
                           ELSE
                              lnoiltax1  = (ROUND(m.ntotbbltxW * (m.nworkint / 100), 2))
                           ENDIF
                        ELSE
                           IF m.nrevtax1 # 0
                              lnoiltax1   = (ROUND(m.ntotbbltx1 * (m.nrevtax1 / 100), 2))
                           ENDIF
                           IF m.nrevtax4 # 0
                              lnoiltax2   = (ROUND(m.ntotbbltx2 * (m.nrevtax4 / 100), 2))
                           ENDIF
                           IF m.nrevtax7 # 0
                              lnoiltax3   = (ROUND(m.ntotbbltx3 * (m.nrevtax7 / 100), 2))
                           ENDIF
                           IF m.nrevtax10 # 0
                              lnoiltax4   = (ROUND(m.ntotbbltx4 * (m.nrevtax10 / 100), 2))
                           ENDIF
                        ENDIF
                        IF lnoiltax1 + lnoiltax2 + lnoiltax3 + lnoiltax4 # 0
                           SELECT wellwork
                           REPL ntotbbltx1 WITH ntotbbltx1 - lnoiltax1, ;
                              ntotbbltx2 WITH ntotbbltx2 - lnoiltax2, ;
                              ntotbbltx3 WITH ntotbbltx3 - lnoiltax3, ;
                              ntotbbltx4 WITH ntotbbltx4 - lnoiltax4
                        ENDIF
                     ENDIF
                     IF INLIST(m.ctypeint, 'B', 'G')
                        IF m.lusesev  && Use state severance tax table rates
                           IF llroyaltyowner
                              lngastax1  = (ROUND(m.ntotmcftxR * (m.nrevgtax / 100), 2))
                           ELSE
                              lngastax1  = (ROUND(m.ntotmcftxW * (m.nworkint / 100), 2))
                           ENDIF
                        ELSE
                           IF m.nrevtax2 # 0
                              lngastax1   = (ROUND(m.ntotmcftx1 * (m.nrevtax2 / 100), 2))
                           ENDIF
                           IF m.nrevtax5 # 0
                              lngastax2   = (ROUND(m.ntotmcftx2 * (m.nrevtax5 / 100), 2))
                           ENDIF
                           IF m.nrevtax8 # 0
                              lngastax3   = (ROUND(m.ntotmcftx3 * (m.nrevtax8 / 100), 2))
                           ENDIF
                           IF m.nrevtax11 # 0
                              lngastax4   = (ROUND(m.ntotmcftx4 * (m.nrevtax11 / 100), 2))
                           ENDIF
                        ENDIF
                        IF lngastax1 + lngastax2 + lngastax3 + lngastax4 # 0
                           SELECT wellwork
                           REPL ntotmcftx1 WITH ntotmcftx1 - lngastax1, ;
                              ntotmcftx2 WITH ntotmcftx2 - lngastax2, ;
                              ntotmcftx3 WITH ntotmcftx3 - lngastax3, ;
                              ntotmcftx4 WITH ntotmcftx4 - lngastax4
                        ENDIF
                     ENDIF
*
*  Calculate other product taxes
*
                     IF m.nrevtax3 # 0
                        lnothtax1 = (ROUND(m.ntotothtx1 * (m.nrevtax3 / 100), 2))
                     ENDIF
                     IF m.nrevtax6 # 0
                        lnothtax2 = (ROUND(m.ntotothtx2 * (m.nrevtax6 / 100), 2))
                     ENDIF
                     IF m.nrevtax9 # 0
                        lnothtax3 = (ROUND(m.ntotothtx3 * (m.nrevtax9 / 100), 2))
                     ENDIF
                     IF m.nrevtax12 # 0
                        lnothtax4 = (ROUND(m.ntotothtx4 * (m.nrevtax12 / 100), 2))
                     ENDIF
                     IF lnothtax1 + lnothtax2 + lnothtax3 + lnothtax4 # 0
                        SELECT wellwork
                        REPL ntotothtx1 WITH ntotothtx1 - lnothtax1, ;
                           ntotothtx2 WITH ntotothtx2 - lnothtax2, ;
                           ntotothtx3 WITH ntotothtx3 - lnothtax3, ;
                           ntotothtx4 WITH ntotothtx4 - lnothtax4
                     ENDIF
                  ENDIF
               ENDSCAN
            ENDSCAN
         ELSE
            SELECT wellwork
            SCAN
               SCATTER MEMVAR
               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  m.lusesev = lusesev
                  m.nroyint = nlandpct + noverpct
               ELSE
                  LOOP
               ENDIF
               SELECT  SUM(nrevtax1) AS nrevtax1, ;
                       SUM(nrevtax2) AS nrevtax2, ;
                       SUM(nrevtax3) AS nrevtax3, ;
                       SUM(nrevtax4) AS nrevtax4, ;
                       SUM(nrevtax5) AS nrevtax5, ;
                       SUM(nrevtax6) AS nrevtax6, ;
                       SUM(nrevtax7) AS nrevtax7, ;
                       SUM(nrevtax8) AS nrevtax8, ;
                       SUM(nrevtax9) AS nrevtax9, ;
                       SUM(nrevtax10) AS nrevtax10, ;
                       SUM(nrevtax11) AS nrevtax11, ;
                       SUM(nrevtax12) AS nrevtax12 ;
                   FROM wellinv,;
                       investor ;
                   WHERE cWellID = m.cWellID ;
                       AND wellinv.cownerid = investor.cownerid ;
                       AND investor.lExempt = .T. ;
                   INTO CURSOR exemptowns ;
                   GROUP BY cWellID

               SELE exemptowns
               SCAN
                  SCATTER MEMVAR

                  lnrevtax1  = 1 - (m.nrevtax1 / 100)
                  lnrevtax2  = 1 - (m.nrevtax2 / 100)
                  lnrevtax3  = 1 - (m.nrevtax3 / 100)
                  lnrevtax4  = 1 - (m.nrevtax4 / 100)
                  lnrevtax5  = 1 - (m.nrevtax5 / 100)
                  lnrevtax6  = 1 - (m.nrevtax6 / 100)
                  lnrevtax7  = 1 - (m.nrevtax7 / 100)
                  lnrevtax8  = 1 - (m.nrevtax8 / 100)
                  lnrevtax9  = 1 - (m.nrevtax9 / 100)
                  lnrevtax10 = 1 - (m.nrevtax10 / 100)
                  lnrevtax11 = 1 - (m.nrevtax11 / 100)
                  lnrevtax12 = 1 - (m.nrevtax12 / 100)

                  IF lnrevtax1 # 0
                     lnoiltax1   = (ROUND(m.ntotbbltx1 / lnrevtax1, 2))
                  ENDIF
                  IF lnrevtax4 # 0
                     lnoiltax2   = (ROUND(m.ntotbbltx2 / lnrevtax4, 2))
                  ENDIF
                  IF lnrevtax7 # 0
                     lnoiltax3   = (ROUND(m.ntotbbltx3 / lnrevtax7, 2))
                  ENDIF
                  IF lnrevtax10 # 0
                     lnoiltax4   = (ROUND(m.ntotbbltx4 / lnrevtax10, 2))
                  ENDIF

                  IF lnoiltax1 + lnoiltax2 + lnoiltax3 + lnoiltax4 # 0
                     SELECT wellwork
                     REPL ntotbbltx1 WITH lnoiltax1, ;
                        ntotbbltx2 WITH lnoiltax2, ;
                        ntotbbltx3 WITH lnoiltax3, ;
                        ntotbbltx4 WITH lnoiltax4
                  ENDIF

                  IF lnrevtax2 # 0
                     lngastax1   = (ROUND(m.ntotmcftx1 / lnrevtax2, 2))
                  ENDIF
                  IF lnrevtax5 # 0
                     lngastax2   = (ROUND(m.ntotmcftx2 / lnrevtax5, 2))
                  ENDIF
                  IF lnrevtax8 # 0
                     lngastax3   = (ROUND(m.ntotmcftx3 / lnrevtax8, 2))
                  ENDIF
                  IF lnrevtax11 # 0
                     lngastax4   = (ROUND(m.ntotmcftx4 / lnrevtax11, 2))
                  ENDIF

                  IF lngastax1 + lngastax2 + lngastax3 + lngastax4 # 0
                     SELECT wellwork
                     REPL ntotmcftx1 WITH lngastax1, ;
                        ntotmcftx2 WITH lngastax2, ;
                        ntotmcftx3 WITH lngastax3, ;
                        ntotmcftx4 WITH lngastax4
                  ENDIF

*
*  Calculate other product taxes
*
                  IF lnrevtax3 # 0
                     lnothtax1 = (ROUND(m.ntotothtx1 / lnrevtax3, 2))
                  ENDIF
                  IF lnrevtax6 # 0
                     lnothtax2 = (ROUND(m.ntotothtx2 / lnrevtax6, 2))
                  ENDIF
                  IF lnrevtax9 # 0
                     lnothtax3 = (ROUND(m.ntotothtx3 / lnrevtax9, 2))
                  ENDIF
                  IF lnrevtax12 # 0
                     lnothtax4 = (ROUND(m.ntotothtx4 / lnrevtax12, 2))
                  ENDIF
                  IF lnothtax1 + lnothtax2 + lnothtax3 + lnothtax4 # 0
                     SELECT wellwork
                     REPL ntotothtx1 WITH lnothtax1, ;
                        ntotothtx2 WITH lnothtax2, ;
                        ntotothtx3 WITH lnothtax3, ;
                        ntotothtx4 WITH lnothtax4
                  ENDIF
               ENDSCAN
            ENDSCAN
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'TaxExempt', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('TaxExempt', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN llReturn
   ENDPROC

*-- Prints suspense reports after closing summary
*********************************
   PROCEDURE PrintSuspense
*********************************

      LOCAL lcTitle1, lcTitle2, llReturn, loError
      LOCAL cGrpName, cProducer, csuspdesc, csusptype, glGrpName, jcAction, jkey, jnAmount, jname
      LOCAL jwell, tcGroup, tcPeriod, tcYear, tnRunNo

      llReturn = .T.

      TRY

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

* Create the suspense object
         THIS.osuspense.GetLastType(.T., .T., THIS.cgroup, .T.)

         tcYear   = THIS.crunyear
         tcPeriod = THIS.cperiod
         tcGroup  = THIS.cgroup
         tnRunNo  = THIS.nrunno

         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         glGrpName = THIS.oOptions.lGrpName

         IF glGrpName
            swselect('groups')
            SET ORDER TO cgroup
            IF SEEK(tcGroup)
               m.cGrpName = cDesc
            ELSE
               IF tcGroup = '**'
                  m.cGrpName = 'All Companies'
               ELSE
                  m.cGrpName = ''
               ENDIF
            ENDIF
         ELSE
            m.cGrpName = ''
         ENDIF

         IF TYPE('m.cProducer') # 'C'
            m.cProducer = 'Development Company'
         ENDIF

*  Create cursor for suspense report thats tacked on to closing summary
         CREATE CURSOR audclose1 ;
            (crptgroup    c(2),    ;
              cownerid     c(10),   ;
              cWellID      c(10),   ;
              csuspdesc    c(45),   ;
              cprogcode    c(10),   ;
              cperiod      c(7), ;
              namount      N(12, 2), ;
              cAction      c(1),    ;
              csusptype    c(1),    ;
              cownname     c(30),   ;
              cwellname    c(30))

*  Into suspense
         SELECT  tsuspense.cownerid, ;
                 tsuspense.cWellID,   ;
                 SUM(tsuspense.nIncome) AS nIncome, ;
                 SUM(tsuspense.nexpense) AS nexpense, ;
                 SUM(tsuspense.nsevtaxes) AS nsevtaxes, ;
                 SUM(tsuspense.nnetcheck) AS namount,   ;
                 'I' AS cAction,   ;
                 tsuspense.cprogcode, ;
                 tsuspense.csusptype, ;
                 tsuspense.hyear + '/' + tsuspense.hperiod AS cperiod, ;
                 NVL(investor.cownname, '**UNKNOWN OWNER**') AS cownname,     ;
                 NVL(wells.cwellname, '**UNKNOWN WELL**') AS cwellname, ;
                 'IA' AS crptgroup   ;
             FROM tsuspense ;
             LEFT OUTER JOIN investor;
                 ON investor.cownerid = tsuspense.cownerid  ;
             LEFT OUTER JOIN wells;
                 ON wells.cWellID = tsuspense.cWellID ;
             WHERE tsuspense.crunyear_in  = tcYear  ;
                 AND tsuspense.nrunno_in    = tnRunNo ;
                 AND tsuspense.cgroup       = tcGroup ;
                 AND NOT investor.ldummy ;
             INTO CURSOR audclosx ;
             ORDER BY tsuspense.csusptype,;
                 tsuspense.cownerid,;
                 tsuspense.cWellID ;
             GROUP BY tsuspense.csusptype,;
                 tsuspense.cownerid,;
                 tsuspense.cWellID

         SELECT audclose1
         APPEND FROM DBF('audclosx')
         swclose('audclosx')
         SELECT  disbhist.cownerid, ;
                 disbhist.cWellID,   ;
                 SUM(disbhist.nIncome) AS nIncome, ;
                 SUM(disbhist.nexpense) AS nexpense, ;
                 SUM(disbhist.nsevtaxes) AS nsevtaxes, ;
                 SUM(disbhist.nnetcheck) AS namount,   ;
                 'I' AS cAction,   ;
                 disbhist.cprogcode, ;
                 disbhist.csusptype, ;
                 disbhist.hyear + '/' + disbhist.hperiod AS cperiod, ;
                 NVL(investor.cownname, '**UNKNOWN OWNER**') AS cownname,     ;
                 NVL(wells.cwellname, '**UNKNOWN WELL**') AS cwellname, ;
                 'IA' AS crptgroup   ;
             FROM disbhist WITH (BUFFERING = .T.)  ;
             LEFT OUTER JOIN investor;
                 ON investor.cownerid = disbhist.cownerid  ;
             LEFT OUTER JOIN wells;
                 ON wells.cWellID = disbhist.cWellID ;
             WHERE disbhist.crunyear_in = tcYear  ;
                 AND disbhist.nrunno_in     = tnRunNo ;
                 AND disbhist.cgroup        = tcGroup ;
                 AND NOT EMPTY(disbhist.csusptype) ;
                 AND NOT investor.ldummy ;
             INTO CURSOR audclosx ;
             ORDER BY disbhist.csusptype,;
                 disbhist.cownerid,;
                 disbhist.cWellID ;
             GROUP BY disbhist.csusptype,;
                 disbhist.cownerid,;
                 disbhist.cWellID

         SELECT audclose1
         APPEND FROM DBF('audclosx')
         swclose('audclosx')

         SELECT audclose1
         SCAN FOR cAction = 'I'
            jnAmount = namount
            jcAction = cAction
            SELECT curLastSuspType
            LOCATE FOR cownerid = audclose1.cownerid AND cWellID = audclose1.cWellID AND cprogcode = audclose1.cprogcode
            IF FOUND()
               m.csusptype = csusptype
               SELECT audclose1
               REPLACE csusptype WITH m.csusptype
            ELSE
               m.csusptype = audclose1.csusptype
            ENDIF
            jkey     = m.csusptype + jcAction
            DO CASE
               CASE jkey = 'DI'
                  m.csuspdesc = 'Deficits Not Covered This Run'
               CASE jkey = 'DO'
                  m.csuspdesc = 'Deficits Covered This Run'
               CASE jkey = 'MI'
                  m.csuspdesc = 'Minimum Check Amounts This Run'
               CASE jkey = 'MO'
                  m.csuspdesc = 'Minimum Check Amounts Released This Run'
               CASE jkey = 'HI'
                  m.csuspdesc = 'Owner Amounts Held This Period'
               CASE jkey = 'HO'
                  m.csuspdesc = 'Owner Amounts Released From Hold This Run'
               CASE jkey = 'QI'
                  m.csuspdesc = 'Quarterly Pays Held'
               CASE jkey = 'QO'
                  m.csuspdesc = 'Quarterly Pays Released'
               CASE jkey = 'SI'
                  m.csuspdesc = 'Semi-Annual Pays Held'
               CASE jkey = 'SO'
                  m.csuspdesc = 'Semi-Annual Pays Released'
               CASE jkey = 'AI'
                  m.csuspdesc = 'Annual Pays Held'
               CASE jkey = 'AO'
                  m.csuspdesc = 'Annual Pays Released'
               CASE jkey = 'II'
                  m.csuspdesc = 'Interests Held This Run'
               CASE jkey = 'IO'
                  m.csuspdesc = 'Held Interests Released This Run'
               OTHERWISE
                  m.csuspdesc = 'Unknown Suspense Type'
            ENDCASE
            SELECT audclose1
            REPLACE crptgroup WITH jkey, ;
                    csuspdesc WITH m.csuspdesc
         ENDSCAN

* Out of suspense
         SELECT  disbhist.cownerid, ;
                 disbhist.cWellID,   ;
                 SUM(disbhist.nIncome) AS nIncome, ;
                 SUM(disbhist.nexpense) AS nexpense, ;
                 SUM(disbhist.nsevtaxes) AS nsevtaxes, ;
                 SUM(disbhist.nnetcheck) AS namount,   ;
                 'O' AS cAction,   ;
                 disbhist.cprogcode, ;
                 disbhist.csusptype, ;
                 disbhist.hyear + '/' + disbhist.hperiod AS cperiod, ;
                 NVL(investor.cownname, '**UNKNOWN OWNER**') AS cownname,     ;
                 NVL(wells.cwellname, '**UNKNOWN WELL**') AS cwellname, ;
                 'OB' AS crptgroup   ;
             FROM disbhist WITH (BUFFERING = .T.) ;
             LEFT OUTER JOIN investor;
                 ON investor.cownerid = disbhist.cownerid  ;
             LEFT OUTER JOIN wells;
                 ON wells.cWellID = disbhist.cWellID ;
             WHERE disbhist.crunyear  = tcYear  ;
                 AND disbhist.nrunno    = tnRunNo ;
                 AND disbhist.cgroup    = tcGroup ;
                 AND (disbhist.nnetcheck # 0 ;
                   OR disbhist.nIncome # 0 ;
                   OR disbhist.nexpense # 0 ;
                   OR disbhist.nsevtaxes # 0) ;
                 AND NOT EMPTY(disbhist.csusptype) ;
                 AND NOT investor.ldummy ;
             INTO CURSOR audclosx ;
             ORDER BY disbhist.csusptype,;
                 disbhist.cownerid,;
                 disbhist.cWellID ;
             GROUP BY disbhist.csusptype,;
                 disbhist.cownerid,;
                 disbhist.cWellID

         SELECT audclose1
         APPEND FROM DBF('audclosx')
         swclose('audclosx')

         THIS.osuspense.GetLastType(.F., .T., THIS.cgroup)

         SELECT audclose1
         SCAN FOR cAction = 'O'
            jnAmount    = namount
            jcAction    = cAction
            m.csusptype = csusptype
            jkey        = m.csusptype + jcAction
            DO CASE
               CASE jkey = 'DI'
                  m.csuspdesc = 'Deficits Not Covered This Run'
               CASE jkey = 'DO'
                  m.csuspdesc = 'Deficits Covered This Run'
               CASE jkey = 'MI'
                  m.csuspdesc = 'Minimum Check Amounts This Run'
               CASE jkey = 'MO'
                  m.csuspdesc = 'Minimum Check Amounts Released This Run'
               CASE jkey = 'HI'
                  m.csuspdesc = 'Owner Amounts Held This Period'
               CASE jkey = 'HO'
                  m.csuspdesc = 'Owner Amounts Released From Hold This Run'
               CASE jkey = 'QI'
                  m.csuspdesc = 'Quarterly Pays Held'
               CASE jkey = 'QO'
                  m.csuspdesc = 'Quarterly Pays Released'
               CASE jkey = 'SI'
                  m.csuspdesc = 'Semi-Annual Pays Held'
               CASE jkey = 'SO'
                  m.csuspdesc = 'Semi-Annual Pays Released'
               CASE jkey = 'AI'
                  m.csuspdesc = 'Annual Pays Held'
               CASE jkey = 'AO'
                  m.csuspdesc = 'Annual Pays Released'
               CASE jkey = 'II'
                  m.csuspdesc = 'Interests Held This Run'
               CASE jkey = 'IO'
                  m.csuspdesc = 'Held Interests Released This Run'
               OTHERWISE
                  m.csuspdesc = 'Unknown Suspense Type'
            ENDCASE
            SELECT audclose1
            REPLACE crptgroup WITH jkey, ;
                    csuspdesc WITH m.csuspdesc
         ENDSCAN

         SELECT  crptgroup, ;
                 cownerid, ;
                 cWellID, ;
                 csuspdesc, ;
                 cprogcode, ;
                 cperiod, ;
                 SUM(namount) AS namount, ;
                 cAction, ;
                 csusptype, ;
                 cownname, ;
                 cwellname ;
             FROM audclose1 ;
             INTO CURSOR audclose READWRITE ;
             ORDER BY crptgroup,;
                 cownerid,;
                 csusptype,;
                 cWellID ;
             GROUP BY crptgroup,;
                 cownerid,;
                 csusptype,;
                 cWellID

         lcTitle1 = 'Run Suspense Activity'
         lcTitle2 = 'For Run No ' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + ' Group ' + THIS.cgroup

         SELECT audclose
         INDEX ON crptgroup + csusptype + cownerid + cWellID + cperiod TAG audkey
         DELETE FOR namount = 0

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'PrintSuspense', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('PrintSuspense', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()


      RETURN llReturn
   ENDPROC

*-- Calculates Well Payout
*********************************
   PROCEDURE Payout
*********************************
*
*  Calculates the well payout based upon the original investments and net cash paid to date
*
      LPARA tcWellID1, tcWellID2, tdDate, tlSelected

      LOCAL llReturn, loError

      llReturn = .T.


      IF NOT tlSelected
         SELECT cWellID AS cID FROM wells WHERE BETWEEN(cWellID, tcWellID1, tcWellID2) INTO CURSOR SELECTED
      ENDIF

      TRY
         IF TYPE('tcWellID1') # 'C'
            tcWellID2 = '*'
         ENDIF
         IF TYPE('tcWellID2') # 'C'
            tcWellID2 = '*'
         ENDIF
         IF TYPE('tdDate') # 'D'
            tdDate = {}
         ENDIF

         CREATE CURSOR Payout ;
            (cWellID     c(10), ;
              cwellname   c(30), ;
              nInvestment N(12, 2), ;
              dDate       D, ;
              nRevenue    N(12, 2), ;
              nTaxes      N(12, 2), ;
              nExpenses   N(12, 2), ;
              nPayout     N(12, 2))

         SELECT  wellinv.cWellID, ;
                 wells.cwellname, ;
                 SUM(wellinv.nInvAmount) AS nInvestment ;
             FROM wellinv,;
                 wells ;
             WHERE wellinv.cWellID IN (SELECT  cID;
                                           FROM SELECTED) ;
                 AND wells.cWellID = wellinv.cWellID ;
             INTO CURSOR tempinv ;
             ORDER BY wellinv.cWellID ;
             GROUP BY wellinv.cWellID

         IF _TALLY > 0
            SELECT Payout
            APPEND FROM DBF('tempinv')
            swclose('tempinv')
         ENDIF

         SELE cWellID, ;
            cdeck, ;
            SUM(noilinc)   AS noilinc, ;
            SUM(nGasInc)   AS nGasInc, ;
            SUM(nOthInc)   AS nOthInc, ;
            SUM(nMiscinc1) AS nMiscinc1, ;
            SUM(nMiscinc2) AS nMiscinc2, ;
            SUM(nTrpInc) AS nTrpInc, ;
            SUM(nTotale) AS nTotale, ;
            SUM(nexpcl1) AS nexpcl1, ;
            SUM(nexpcl2) AS nexpcl2, ;
            SUM(nexpcl3) AS nexpcl3, ;
            SUM(nexpcl4) AS nexpcl4, ;
            SUM(nexpcl5) AS nexpcl5, ;
            SUM(nexpclA) AS nexpclA, ;
            SUM(nexpclB) AS nexpclB, ;
            SUM(ntotbbltx1) AS ntotbbltx1,  ;
            SUM(ntotbbltx2) AS ntotbbltx2,  ;
            SUM(ntotbbltx3) AS ntotbbltx3,  ;
            SUM(ntotbbltx4) AS ntotbbltx4,  ;
            SUM(ntotmcftx1) AS ntotmcftx1,  ;
            SUM(ntotmcftx2) AS ntotmcftx2,  ;
            SUM(ntotmcftx3) AS ntotmcftx3,  ;
            SUM(ntotmcftx4) AS ntotmcftx4,  ;
            SUM(ntotothtx1) AS ntotothtx1,  ;
            SUM(ntotothtx2) AS ntotothtx2,  ;
            SUM(ntotothtx3) AS ntotothtx3,  ;
            SUM(ntotothtx4) AS ntotothtx4  ;
            FROM wellhist ;
            WHERE BETWEEN(cWellID, tcWellID1, tcWellID2) AND crectype = 'R' AND hdate <= tdDate ;
            INTO CURSOR temp ;
            ORDER BY cWellID GROUP BY cWellID


         SELECT  cWellID,;
                 cdeck, ;
                 SUM(noilinc + nGasInc + nTrpInc + nOthInc + nMiscinc1 + nMiscinc2) AS nRevenue, ;
                 SUM(nTotale + nexpcl1 + nexpcl2 + nexpcl3 + nexpcl4 + nexpcl5 + nexpclA + nexpclB) AS nExpenses, ;
                 SUM(ntotbbltx1 + ntotbbltx2 + ntotbbltx3 + ntotbbltx4 + ntotmcftx1 + ntotmcftx2 + ntotmcftx3 + ntotmcftx4 +  ;
                     ntotothtx1 + ntotothtx2 + ntotothtx3 + ntotothtx4) AS nTaxes ;
             FROM temp ;
             INTO CURSOR temphist ;
             ORDER BY cWellID ;
             GROUP BY cWellID

         IF _TALLY > 0
            SELECT temphist
            SCAN
               SCATTER MEMVAR
               SELECT Payout
               LOCATE FOR cWellID = m.cWellID
               IF FOUND()
                  m.nRevenue  = swnetrev(m.cWellID, m.nRevenue, 'G', .F., .T., .T., .F., .F., .F., .F., m.cdeck)
                  m.nExpenses = swNetExp(m.nExpenses, m.cWellID, .T., '0', 'B', '', '', .F., '' )
                  m.nTaxes    = swnetrev(m.cWellID, m.nTaxes, 'G', .F., .T., .T., .F., .F., .F., .F., m.cdeck)
                  SELECT Payout
                  REPL nRevenue WITH m.nRevenue, ;
                     nTaxes   WITH m.nTaxes, ;
                     nExpenses WITH m.nExpenses, ;
                     nPayout   WITH nInvestment - (m.nRevenue - m.nTaxes - m.nExpenses)
               ENDIF
            ENDSCAN
            swclose('temphist')
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'Payout', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('Payout', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Calculates the split between royalty and working owners for wellhist.
*********************************
   PROCEDURE RoyWork
*********************************

      LOCAL lckey, llReturn, loError

      llReturn = .T.

      TRY
* Get the working interest totals
         SELECT  cWellID, ;
                 hyear, ;
                 hperiod, ;
                 0.00 AS nroyalty, ;
                 SUM(nnetcheck) AS nworking ;
             FROM invtmp WITH (BUFFERING = .T.);
             WHERE ctypeinv = 'W' ;
                 AND lprogram = .F. ;
             INTO CURSOR temphistw ;
             ORDER BY cWellID,;
                 hyear,;
                 hperiod ;
             GROUP BY cWellID,;
                 hyear,;
                 hperiod

* Get the royalty totals
         SELECT  cWellID, ;
                 hyear, ;
                 hperiod, ;
                 0.00 AS nworking, ;
                 SUM(nnetcheck) AS nroyalty ;
             FROM invtmp WITH (BUFFERING = .T.);
             WHERE (ctypeinv = 'L' ;
                   OR ctypeinv = 'O') ;
                 AND lprogram = .F. ;
             INTO CURSOR temphistr ;
             ORDER BY cWellID,;
                 hyear,;
                 hperiod ;
             GROUP BY cWellID,;
                 hyear,;
                 hperiod

         CREATE CURSOR fixhist ;
            (cWellID    c(10), ;
              hyear     c(4), ;
              hperiod   c(2), ;
              nroyalty   N(12, 2), ;
              nworking   N(12, 2))
         INDEX ON cWellID + hyear + hperiod TAG wellprd

         SELECT fixhist
         APPEND FROM DBF('temphistw')

         SELECT temphistr
         GO TOP
         SCAN
            SCATTER MEMVAR
            lckey = m.cWellID + m.hyear + m.hperiod
            SELECT fixhist
            SET ORDER TO wellprd
            SEEK lckey
            IF FOUND()
               REPLACE nroyalty WITH m.nroyalty
            ELSE
               m.nworking = 0
               INSERT INTO fixhist FROM MEMVAR
            ENDIF
            SELECT temphistr
         ENDSCAN

         swclose('temphistr')
         swclose('temphistw')

         swselect('wellhist')
         SET ORDER TO wellprd

         SELECT fixhist
         GO TOP
         SCAN
            SCATTER MEMVAR
            lckey = m.cWellID + m.hyear + m.hperiod + 'R'
            swselect('wellhist')
            IF SEEK(lckey)
               REPLACE nroyint WITH m.nroyalty, ;
                       nwrkint WITH m.nworking
            ENDIF
         ENDSCAN

         IF USED('fixhist')
            SELECT fixhist
            USE
         ENDIF

         WAIT CLEAR

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'RoyWork', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('RoyWork', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE CalcRounding
*********************************
      LPARA tlAll
      LOCAL llRoundHigh, lnMaxRound, lcOwnerID

      llReturn = .T.

      TRY
* If we're not closing, invtmp won't have all owners in it so
* there's no use in trying to calculate rounding amounts
         IF NOT tlAll
            llReturn = .T.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               THIS.cErrorMsg = 'Processing canceled by user.'
               EXIT
            ENDIF
         ENDIF


         THIS.oprogress.SetProgressMessage('Adjusting for Rounding...')
         THIS.oprogress.UpdateProgress(THIS.nprogress)
         THIS.nprogress = THIS.nprogress + 1

         llRoundHigh = THIS.oOptions.lRoundHigh
         lnMaxRound  = THIS.oOptions.nMaxRound
         lcOwnerID   = ''
         m.cdmbatch  = THIS.cdmbatch

         IF lnMaxRound = 0
            lnMaxRound = 1.99
         ENDIF

         swselect('wells')
         SET ORDER TO cWellID

         swselect('wellinv')
         SET ORDER TO wellinvid

* Check to make sure the owners specified to get the rounding really
* have an interest in the well
         SELECT wells
         SCAN FOR NOT EMPTY(cownerid)
            m.cWellID  = cWellID
            m.cownerid = cownerid
* Look in the doi for that owner and well
            SELECT wellinv
            IF NOT SEEK(m.cWellID + m.cownerid + 'W')
               SELECT wells
* Blank out the ownerid
               REPLACE cownerid WITH ''
            ENDIF
         ENDSCAN

         swselect('roundtmp')

         CREATE CURSOR roundtmpx  ;
            (cWellID    c(10),    ;
              cdmbatch    c(8),     ;
              lused       L,        ;
              cwellname   c(40),    ;
              cownerid    c(10),    ;
              cownname    c(40),    ;
              ngasrev     N(12, 2),  ;
              noilrev     N(12, 2),  ;
              ntrprev     N(12, 2),  ;
              nothrev     N(12, 2),  ;
              nCompress   N(12, 2),  ;
              nGather     N(12, 2),  ;
              nMKTGExp    N(12, 2),  ;
              nmiscrev1   N(12, 2),  ;
              nmiscrev2   N(12, 2),  ;
              nexpense    N(12, 2),  ;
              nPlugExp    N(12, 2),  ;
              ntotale1    N(12, 2),  ;
              ntotale2    N(12, 2),  ;
              ntotale3    N(12, 2),  ;
              ntotale4    N(12, 2),  ;
              ntotale5    N(12, 2),  ;
              ntotalea    N(12, 2),  ;
              ntotaleb    N(12, 2),  ;
              noiltax1    N(12, 2),  ;
              noiltax2    N(12, 2),  ;
              noiltax3    N(12, 2),  ;
              noiltax4    N(12, 2),  ;
              ngastax1    N(12, 2),  ;
              ngastax2    N(12, 2),  ;
              ngastax3    N(12, 2),  ;
              ngastax4    N(12, 2),  ;
              nOthTax1    N(12, 2),  ;
              nOthTax2    N(12, 2),  ;
              nOthTax3    N(12, 2),  ;
              nOthTax4    N(12, 2))

         SELECT wellwork
         COUNT FOR NOT DELETED() TO lnMax
         lnCount = 1

* Total amounts from well history
         SELECT  cWellID, ;
                 SUM(ngrossgas + nflatgas) AS ngrossgas, ;
                 SUM(ngrossoil + nflatoil) AS ngrossoil, ;
                 SUM(nTrpInc)   AS nTrpInc, ;
                 SUM(nCompress) AS nCompress, ;
                 SUM(nGather)   AS nGather, ;
                 SUM(nTotMKTG)  AS nTotMKTG, ;
                 SUM(nMiscinc1) AS nMiscinc1, ;
                 SUM(nMiscinc2) AS nMiscinc2, ;
                 SUM(nOthInc)   AS nOthInc, ;
                 SUM(nTotale)   AS nNetExp, ;
                 SUM(nexpcl1)   AS nexpcl1, ;
                 SUM(nexpcl2)   AS nexpcl2, ;
                 SUM(nexpcl3)   AS nexpcl3, ;
                 SUM(nexpcl4)   AS nexpcl4, ;
                 SUM(nexpcl5)   AS nexpcl5, ;
                 SUM(nexpclA)   AS nexpclA, ;
                 SUM(nexpclB)   AS nexpclB, ;
                 SUM(ntotbbltx1 + ntotbbltxR + ntotbbltxW) AS ntotbbltx1, ;
                 SUM(ntotbbltx2) AS ntotbbltx2, ;
                 SUM(ntotbbltx3) AS ntotbbltx3, ;
                 SUM(ntotbbltx4) AS ntotbbltx4, ;
                 SUM(ntotmcftx1 + ntotmcftxR + ntotmcftxW) AS ntotmcftx1, ;
                 SUM(ntotmcftx2) AS ntotmcftx2, ;
                 SUM(ntotmcftx3) AS ntotmcftx3, ;
                 SUM(ntotmcftx4) AS ntotmcftx4, ;
                 SUM(ntotothtx1) AS ntotothtx1, ;
                 SUM(ntotothtx2) AS ntotothtx2, ;
                 SUM(ntotothtx3) AS ntotothtx3, ;
                 SUM(ntotothtx4) AS ntotothtx4, ;
                 SUM(nPlugAmt)   AS nplug ;
             FROM wellwork ;
             WHERE BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid) ;
             INTO CURSOR wellwrk ;
             ORDER BY cWellID ;
             GROUP BY cWellID

         SELECT wellwrk
         SCAN
            SCATTER MEMVAR
* Total amounts from owner history
            SELECT  cWellID,  ;
                    SUM(ngasrev) AS ngasrev,  ;
                    SUM(noilrev) AS noilrev,  ;
                    SUM(ntrprev) AS ntrprev,  ;
                    SUM(nothrev) AS nothrev,  ;
                    SUM(nCompress) AS nCompress, ;
                    SUM(nGather)   AS nGather, ;
                    SUM(nMKTGExp)  AS nMKTGExp, ;
                    SUM(nmiscrev1) AS nmiscrev1,  ;
                    SUM(nmiscrev2) AS nmiscrev2,  ;
                    SUM(nexpense) AS nexpense,  ;
                    SUM(ntotale1) AS ntotale1,  ;
                    SUM(ntotale2) AS ntotale2,  ;
                    SUM(ntotale3) AS ntotale3,  ;
                    SUM(ntotale4) AS ntotale4,  ;
                    SUM(ntotale5) AS ntotale5,  ;
                    SUM(ntotalea) AS ntotalea,  ;
                    SUM(ntotaleb) AS ntotaleb,  ;
                    SUM(noiltax1) AS noiltax1,  ;
                    SUM(noiltax2) AS noiltax2,  ;
                    SUM(noiltax3) AS noiltax3,  ;
                    SUM(noiltax4) AS noiltax4,  ;
                    SUM(ngastax1) AS ngastax1,  ;
                    SUM(ngastax2) AS ngastax2,  ;
                    SUM(ngastax3) AS ngastax3,  ;
                    SUM(ngastax4) AS ngastax4,  ;
                    SUM(nOthTax1) AS nOthTax1,  ;
                    SUM(nOthTax2) AS nOthTax2,  ;
                    SUM(nOthTax3) AS nOthTax3,  ;
                    SUM(nOthTax4) AS nOthTax4,  ;
                    SUM(nPlugExp) AS nPlugExp   ;
                FROM invtmp ;
                WHERE cWellID = m.cWellID  ;
                INTO CURSOR difftmp ;
                ORDER BY cWellID ;
                GROUP BY cWellID

* Get the difference between amounts in well history and owner history
            IF _TALLY > 0
               SELECT difftmp
               SCAN
                  m.ngasrev   = m.ngrossgas - ngasrev
                  m.noilrev   = m.ngrossoil - noilrev
                  m.ntrprev   = m.nTrpInc - ntrprev
                  m.nCompress = m.nCompress - nCompress
                  m.nGather   = m.nGather - nGather
                  m.nMKTGExp  = m.nTotMKTG - nMKTGExp
                  m.nmiscrev1 = m.nMiscinc1 - nmiscrev1
                  m.nmiscrev2 = m.nMiscinc2 - nmiscrev2
                  m.nothrev   = m.nOthInc - nothrev
                  m.nexpense  = m.nNetExp - nexpense
                  m.ntotale1  = m.nexpcl1 - ntotale1
                  m.ntotale2  = m.nexpcl2 - ntotale2
                  m.ntotale3  = m.nexpcl3 - ntotale3
                  m.ntotale4  = m.nexpcl4 - ntotale4
                  m.ntotale5  = m.nexpcl5 - ntotale5
                  m.ntotalea  = m.nexpclA - ntotalea
                  m.ntotaleb  = m.nexpclB - ntotaleb
                  m.noiltax1  = m.ntotbbltx1 - noiltax1
                  m.noiltax2  = m.ntotbbltx2 - noiltax2
                  m.noiltax3  = m.ntotbbltx3 - noiltax3
                  m.noiltax4  = m.ntotbbltx4 - noiltax4
                  m.ngastax1  = m.ntotmcftx1 - ngastax1
                  m.ngastax2  = m.ntotmcftx2 - ngastax2
                  m.ngastax3  = m.ntotmcftx3 - ngastax3
                  m.ngastax4  = m.ntotmcftx4 - ngastax4
                  m.nOthTax1  = m.ntotothtx1 - nOthTax1
                  m.nOthTax2  = m.ntotothtx2 - nOthTax2
                  m.nOthTax3  = m.ntotothtx3 - nOthTax3
                  m.nOthTax4  = m.ntotothtx4 - nOthTax4
                  m.nPlugExp  = m.nplug - nPlugExp
               ENDSCAN
               STORE '' TO m.cwellname, m.cownname, m.cownerid
               m.lused = .F.
               INSERT INTO roundtmpx FROM MEMVAR
            ENDIF
         ENDSCAN

* Get a cursor with only wells that have rounding
         SELECT  * ;
             FROM roundtmpx ;
             INTO CURSOR roundtmp1 ;
             WHERE (ngasrev # 0)  ;
                 OR (noilrev # 0)  ;
                 OR (ntrprev # 0)  ;
                 OR (nmiscrev1 # 0);
                 OR (nmiscrev2 # 0)  ;
                 OR (nexpense # 0) ;
                 OR (nothrev # 0)  ;
                 OR (nCompress # 0);
                 OR (nGather # 0)  ;
                 OR (ntotale1 # 0) ;
                 OR (ntotale2 # 0)  ;
                 OR (ntotale3 # 0) ;
                 OR (ntotale4 # 0)  ;
                 OR (ntotale5 # 0) ;
                 OR (noiltax1 # 0)  ;
                 OR (noiltax2 # 0) ;
                 OR (noiltax3 # 0)  ;
                 OR (noiltax4 # 0) ;
                 OR (ngastax1 # 0)  ;
                 OR (ngastax2 # 0) ;
                 OR (ngastax3 # 0)  ;
                 OR (ngastax4 # 0) ;
                 OR (nOthTax1 # 0)  ;
                 OR (nOthTax2 # 0) ;
                 OR (nOthTax3 # 0)  ;
                 OR (nOthTax4 # 0) ;
                 OR (nMKTGExp # 0)  ;
                 OR (ntotalea # 0) ;
                 OR (ntotaleb # 0) ;
                 OR (nPlugExp # 0)


* Pare down the records by comparing the rounding to the max rounding allowed
* If the rounding is within the range then it can be adjusted.
         SELECT  * ;
             FROM roundtmp1 ;
             INTO CURSOR roundtmp2 ;
             WHERE (ABS(ngasrev) <= lnMaxRound)  ;
                 AND  (ABS(noilrev)  <= lnMaxRound);
                 AND  (ABS(ntrprev)  <= lnMaxRound)  ;
                 AND  (ABS(nmiscrev1) <= lnMaxRound);
                 AND  (ABS(nmiscrev2) <= lnMaxRound)  ;
                 AND  (ABS(nCompress) <= lnMaxRound);
                 AND  (ABS(nGather)  <= lnMaxRound)  ;
                 AND  (ABS(nexpense) <= lnMaxRound);
                 AND  (ABS(nothrev)  <= lnMaxRound)  ;
                 AND  (ABS(ntotale1) <= lnMaxRound);
                 AND  (ABS(ntotale2) <= lnMaxRound)  ;
                 AND  (ABS(ntotale3) <= lnMaxRound);
                 AND  (ABS(ntotale4) <= lnMaxRound)  ;
                 AND  (ABS(ntotale5) <= lnMaxRound);
                 AND  (ABS(noiltax1) <= lnMaxRound)  ;
                 AND  (ABS(noiltax2) <= lnMaxRound);
                 AND  (ABS(noiltax3) <= lnMaxRound)  ;
                 AND  (ABS(noiltax4) <= lnMaxRound);
                 AND  (ABS(ngastax1) <= lnMaxRound)  ;
                 AND  (ABS(ngastax2) <= lnMaxRound);
                 AND  (ABS(ngastax3) <= lnMaxRound)  ;
                 AND  (ABS(ngastax4) <= lnMaxRound);
                 AND  (ABS(nOthTax1) <= lnMaxRound)  ;
                 AND  (ABS(nOthTax2) <= lnMaxRound);
                 AND  (ABS(nOthTax3) <= lnMaxRound)  ;
                 AND  (ABS(nOthTax4) <= lnMaxRound);
                 AND  (ABS(nMKTGExp) <= lnMaxRound)  ;
                 AND  (ABS(ntotalea) <= lnMaxRound);
                 AND  (ABS(ntotaleb) <= lnMaxRound) ;
                 AND  (ABS(nPlugExp) <= lnMaxRound)

         IF _TALLY > 0
            SELECT roundtmp2
            USE DBF('roundtmp2') AGAIN IN 0 ALIAS roundtmp3  && Why are we doing this?  pws-2/8/21
            SELECT roundtmp3
            SCAN
               SCATTER MEMVAR
               THIS.oprogress.SetProgressMessage('Adjusting for Rounding...' + roundtmp3.cWellID)
               swselect('wells')
               IF SEEK(roundtmp3.cWellID)
                  SCATTER FIELDS LIKE lSev* MEMVAR
                  IF EMPTY(wells.cownerid)  &&  No owner specified to adjust rounding to
                     IF llRoundHigh  &&  If rounding to highest owner
                        swselect('wellinv')
                        SET ORDER TO cownerid DESC
                        LOCATE FOR cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nworkint # 0 AND NOT lJIB
                        IF FOUND('wellinv')
                           lcOwnerID = wellinv.cownerid
                        ENDIF
                     ELSE  &&  Rounding to lowest owner
                        swselect('wellinv')
                        SET ORDER TO cownerid
                        LOCATE FOR cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nworkint # 0 AND NOT lJIB
                        IF FOUND('wellinv')
                           lcOwnerID = wellinv.cownerid
                        ENDIF
                     ENDIF
                  ELSE  &&  An owner is specified
                     lcOwnerID = wells.cownerid
                  ENDIF
                  REPLACE roundtmp3.cownerid WITH lcOwnerID

                  SELECT invtmp
                  IF roundtmp3.ngasrev # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ngasrev # 0 AND nrevgas # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (noilrev # 0 OR ntrprev # 0 OR nothrev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevgas # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevgas # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ngasrev WITH invtmp.ngasrev + roundtmp3.ngasrev
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.noilrev # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND noilrev # 0 AND nrevoil # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR ntrprev # 0 OR nothrev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevoil # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevoil # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.noilrev WITH invtmp.noilrev + roundtmp3.noilrev
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntrprev # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntrprev # 0 AND nrevtrp # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR nothrev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevtrp # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevtrp # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntrprev WITH invtmp.ntrprev + roundtmp3.ntrprev
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nCompress # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nCompress # 0 AND nrevgas # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR nothrev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevgas # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevgas # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nCompress WITH invtmp.nCompress + roundtmp3.nCompress
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nGather # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nGather # 0 AND nrevgas # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR nothrev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevgas # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevgas # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nGather WITH invtmp.nGather + roundtmp3.nGather
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nMKTGExp # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nMKTGExp # 0 AND nrevgas # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR nothrev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevgas # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevgas # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nMKTGExp WITH invtmp.nMKTGExp + roundtmp3.nMKTGExp
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nmiscrev1 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nmiscrev1 # 0 AND nrevmisc1 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR nothrev # 0 OR ntrprev # 0 OR nmiscrev2 # 0) AND nrevmisc1 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevmisc1 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nmiscrev1 WITH invtmp.nmiscrev1 + roundtmp3.nmiscrev1
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nmiscrev2 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nmiscrev2 # 0 AND nrevmisc2 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR nothrev # 0 OR ntrprev # 0 OR nmiscrev1 # 0) AND nrevmisc2 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevmisc2 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nmiscrev2 WITH invtmp.nmiscrev2 + roundtmp3.nmiscrev2
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nothrev # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nothrev # 0 AND nrevoth # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ngasrev # 0 OR noilrev # 0 OR ntrprev # 0 OR nmiscrev1 # 0 OR nmiscrev2 # 0) AND nrevoth # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevoth # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nothrev WITH invtmp.nothrev + roundtmp3.nothrev
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nexpense # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nexpense # 0 AND nworkint # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (ntotale1 # 0 OR ntotale2 # 0 OR ntotale3 # 0 OR ntotale4 # 0 OR ntotale5 # 0  ;
                             OR ntotaleb # 0 OR ntotalea # 0) AND nworkint # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nworkint # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nexpense WITH invtmp.nexpense + roundtmp3.nexpense
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotale1 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotale1 # 0 AND nintclass1 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale2 # 0 OR ntotale3 # 0 OR ntotale4 # 0 OR ntotale5 # 0  ;
                             OR ntotaleb # 0 OR ntotalea # 0) AND nintclass1 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nintclass1 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotale1 WITH invtmp.ntotale1 + roundtmp3.ntotale1
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotale2 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotale2 # 0 AND nintclass2 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale1 # 0 OR ntotale3 # 0 OR ntotale4 # 0 OR ntotale5 # 0  ;
                             OR ntotaleb # 0 OR ntotalea # 0) AND nintclass2 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nintclass2 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotale2 WITH invtmp.ntotale2 + roundtmp3.ntotale2
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotale3 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotale3 # 0 AND nintclass3 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale1 # 0 OR ntotale2 # 0 OR ntotale4 # 0 OR ntotale5 # 0  ;
                             OR ntotaleb # 0 OR ntotalea # 0) AND nintclass3 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nintclass3 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotale3 WITH invtmp.ntotale3 + roundtmp3.ntotale3
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotale4 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotale4 # 0 AND nintclass4 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale1 # 0 OR ntotale2 # 0 OR ntotale3 # 0 OR ntotale5 # 0  ;
                             OR ntotaleb # 0 OR ntotalea # 0) AND nintclass4 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nintclass1 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotale4 WITH invtmp.ntotale4 + roundtmp3.ntotale4
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotale5 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotale5 # 0 AND nintclass5 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale1 # 0 OR ntotale2 # 0 OR ntotale3 # 0 OR ntotale4 # 0  ;
                             OR ntotaleb # 0 OR ntotalea # 0) AND nintclass5 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotale5 WITH invtmp.ntotale5 + roundtmp3.ntotale5
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotaleb # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotaleb # 0 AND nbcpint # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale1 # 0 OR ntotale2 # 0 OR ntotale3 # 0 OR ntotale4 # 0  ;
                             OR ntotale5 # 0 OR ntotalea # 0) AND nbcpint # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nbcpint # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotaleb WITH invtmp.ntotaleb + roundtmp3.ntotaleb
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ntotalea # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ntotalea # 0 AND nacpint # 0
                     IF NOT FOUND()  &&  Not found, so find a record with some type of expense
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'  ;
                           AND (nexpense # 0 OR ntotale1 # 0 OR ntotale2 # 0 OR ntotale3 # 0 OR ntotale4 # 0  ;
                             OR ntotale5 # 0 OR ntotaleb # 0) AND nacpint # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nacpint # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ntotalea WITH invtmp.ntotalea + roundtmp3.ntotalea
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.noiltax1 # 0 OR roundtmp3.noiltax2 # 0 OR roundtmp3.noiltax3 # 0 OR roundtmp3.noiltax4 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND (noiltax1 # 0  ;
                          OR noiltax2 # 0 OR noiltax3 # 0 OR noiltax4 # 0) AND nrevtax1 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with the right kind of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND noilrev # 0 AND nrevtax1 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevtax1 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.noiltax1 WITH invtmp.noiltax1 + roundtmp3.noiltax1,  ;
                                invtmp.noiltax2 WITH invtmp.noiltax2 + roundtmp3.noiltax2,  ;
                                invtmp.noiltax3 WITH invtmp.noiltax3 + roundtmp3.noiltax3,  ;
                                invtmp.noiltax4 WITH invtmp.noiltax4 + roundtmp3.noiltax4
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.ngastax1 # 0 OR roundtmp3.ngastax2 # 0 OR roundtmp3.ngastax3 # 0 OR roundtmp3.ngastax4 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND (ngastax1 # 0  ;
                          OR ngastax2 # 0 OR ngastax3 # 0 OR ngastax4 # 0) AND nrevtax2 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with the right kind of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND ngasrev # 0 AND nrevtax2 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevtax2 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.ngastax1 WITH invtmp.ngastax1 + roundtmp3.ngastax1,  ;
                                invtmp.ngastax2 WITH invtmp.ngastax2 + roundtmp3.ngastax2,  ;
                                invtmp.ngastax3 WITH invtmp.ngastax3 + roundtmp3.ngastax3,  ;
                                invtmp.ngastax4 WITH invtmp.ngastax4 + roundtmp3.ngastax4
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nOthTax1 # 0 OR roundtmp3.nOthTax2 # 0 OR roundtmp3.nOthTax3 # 0 OR roundtmp3.nOthTax4 # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND (nOthTax1 # 0  ;
                          OR nOthTax2 # 0 OR nOthTax3 # 0 OR nOthTax4 # 0) AND nrevtax3 # 0
                     IF NOT FOUND()  &&  Not found, so find a record with the right kind of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nothrev # 0 AND nrevtax3 # 0
                        IF NOT FOUND()
                           LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nrevtax3 # 0
                        ENDIF
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nOthTax1 WITH invtmp.nOthTax1 + roundtmp3.nOthTax1,  ;
                                invtmp.nOthTax2 WITH invtmp.nOthTax2 + roundtmp3.nOthTax2,  ;
                                invtmp.nOthTax3 WITH invtmp.nOthTax3 + roundtmp3.nOthTax3,  ;
                                invtmp.nOthTax4 WITH invtmp.nOthTax4 + roundtmp3.nOthTax4
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF

                  IF roundtmp3.nPlugExp # 0  &&  Only do a replace if there is rounding present
                     LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W' AND nplugpct # 0
                     IF NOT FOUND()  &&  Not found, so find a record with the right kind of revenue
                        LOCATE FOR cownerid = lcOwnerID AND cWellID = roundtmp3.cWellID AND ctypeinv = 'W'
                     ENDIF
                     IF NOT EOF('invtmp')  &&  If not eof(), it found a record
                        REPLACE invtmp.nPlugExp WITH invtmp.nPlugExp + roundtmp3.nPlugExp
                        THIS.CalcRoundingTotals()  &&  Update Totals with the changes
                     ENDIF
                  ENDIF
               ENDIF
            ENDSCAN

            SELECT roundtmp3
            USE DBF('roundtmp3') AGAIN IN 0 ALIAS roundtmp4
            SELECT roundtmp4
            SCAN
               SCATTER MEMVAR
               swselect('wells')
               LOCATE FOR cWellID = roundtmp4.cWellID
               IF FOUND('wells')
                  IF EMPTY(wells.cownerid)  &&  No owner specified to adjust rounding to
                     IF llRoundHigh  &&  If rounding to highest owner
                        swselect('wellinv')
                        SET ORDER TO cownerid DESC
                        LOCATE FOR cWellID = roundtmp4.cWellID AND ctypeinv = 'W' AND nworkint # 0 AND NOT lJIB
                        IF FOUND('wellinv')
                           lcOwnerID = wellinv.cownerid
                        ENDIF
                     ELSE  &&  Rounding to lowest owner
                        swselect('wellinv')
                        SET ORDER TO cownerid
                        LOCATE FOR cWellID = roundtmp4.cWellID AND ctypeinv = 'W' AND nworkint # 0 AND NOT lJIB
                        IF FOUND('wellinv')
                           lcOwnerID = wellinv.cownerid
                        ENDIF
                     ENDIF
                  ELSE  &&  An owner is specified
                     lcOwnerID = wells.cownerid
                  ENDIF
                  REPLACE roundtmp4.cownerid WITH lcOwnerID
                  REPLACE roundtmp4.cwellname WITH wells.cwellname

                  swselect('investor')
                  LOCATE FOR cownerid = lcOwnerID
                  IF FOUND('Investor')
                     REPLACE roundtmp4.cownname WITH investor.cownname
                  ENDIF
               ENDIF
            ENDSCAN

            IF THIS.lclose
               SELECT roundtmp
               APPEND FROM DBF('roundtmp4')
            ENDIF

         ENDIF

         swclose('roundtmp1')
         swclose('roundtmp2')
         swclose('roundtmp3')
         swclose('roundtmp4')

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Adjusting for Rounding...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CalcRounding', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CalcRounding', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Prints the precheck register.
*********************************
   PROCEDURE CheckListing
*********************************
      LOCAL tcAccount, llDMPRO
      LOCAL lcSelect, lcSortOrder, lcTitle1, lcTitle2, llReturn, loError

      llReturn = .T.

      TRY
         tcAccount = THIS.oOptions.cDisbAcct

         SET DELETED ON

         m.cProducer  = m.goApp.cCompanyName
         m.cProcessor = ''

         IF THIS.lclose
            lcTitle1     = 'Checks Created for Run: R' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + '/' + THIS.cgroup
         ELSE
            lcTitle1     = 'Checks Created for Period: ' + THIS.crunyear + '/' + THIS.cperiod + ' Group: ' + THIS.cgroup
         ENDIF
         lcTitle2    = 'For Account ' + ALLTRIM(tcAccount)
         lcSelect    = ''
         lcSortOrder = ''

         CREATE CURSOR tempchk ;
            (cidtype    c(1), ;
              cID        c(10), ;
              cPayee     c(40), ;
              dCheckDate D, ;
              cyear      c(4), ;
              cperiod    c(2), ;
              namount    N(12, 2))

         llDMPRO = .F.

         IF TYPE('m.goApp') = 'O'
            IF m.goApp.ldmpro
               llDMPRO = .T.
            ENDIF
         ENDIF

         SELECT  'R' AS crptgroup, ;
                 cidtype, ;
                 cID, ;
                 cPayee, ;
                 dCheckDate, ;
                 cyear, ;
                 cperiod, ;
                 namount ;
             FROM checks ;
             WHERE cBatch = THIS.cdmbatch ;
                 AND EMPTY(ccheckno) ;
             INTO CURSOR tempchk READWRITE ;
             ORDER BY crptgroup,;
                 cidtype,;
                 cID

         SELECT  'D' AS crptgroup, ;
                 cidtype, ;
                 cID, ;
                 cPayee, ;
                 dCheckDate, ;
                 cyear, ;
                 cperiod, ;
                 namount ;
             FROM checks ;
             WHERE cBatch = THIS.cdmbatch ;
                 AND NOT EMPTY(ccheckno) ;
                 AND ('DIRDEP' $ ccheckno OR LEFT(ALLTRIM(ccheckno), 1) = 'E') OR ('FEDWIRE' $ ccheckno) ;
             INTO CURSOR temp ;
             ORDER BY crptgroup,;
                 cidtype,;
                 cID

         IF _TALLY > 0
            SELECT tempchk
            APPEND FROM DBF('temp')
            swclose('temp')
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CheckListing', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CheckListing', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY


      RETURN llReturn

   ENDPROC

*-- Unallocated Revenue and Expense Report
*********************************
   PROCEDURE UnAllRpt
*********************************
      LOCAL tcWell1, tcWellname1, tcWell2, tcWellname2, tcGroup, tnSort, tnReport
      LOCAL lBetween, lOrderby, lcSortOrder, lcTitle1, lcTitle2, llReturn, loError
      LOCAL tdDate1, tdDate2, tlNewRun

      llReturn = .T.


      TRY
         tcWell1  = THIS.cbegwellid
         tcWell2  = THIS.cendwellid
         tdDate1  = {01/01/1900}
         tdDate2  = {12/31/2999}
         tnSort   = 3
         tnReport = 3
         tcGroup  = THIS.cgroup
         tlNewRun = .T.
         STORE '' TO tcWellname1, tcWellname2

         IF swUnallRpt(tcWell1, tcWellname1, tcWell2, tcWellname2, tnSort, tnReport, tdDate1, tdDate2, tlNewRun, tcGroup)

            lOrderby    = 'wells.cwellid'
            lcSortOrder = 'Well ID'
            lBetween    = 'BETWEEN(wells.cwellid,tcWell1,tcWell2)'

            lcTitle1    = 'Unallocated/Unprocessed Revenue and Expenses'
            lcTitle2    = ''
            glGrpName   = .F.
            m.cProducer = m.goApp.cCompanyName
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'UnAllRpt', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('UnAllRpt', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN llReturn
   ENDPROC

*-- Checks to see if a run no has been closed
*********************************
   PROCEDURE CheckHistRun
*********************************
      LOCAL llHist, llSepClose, lcDeleted
      LOCAL llReturn, loError
*
*  Checks to see if the given period is closed
*  Returns .T. if the period is closed
*
      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF TYPE('xcDebug') = 'C' AND xcDebug = 'Y'
            WAIT WIND 'In Distproc Checkhist...'
         ENDIF

         lcDeleted = SET('DELETED')
         SET DELETED ON

         llHist = .F.

         IF THIS.cgroup = '**'
            swselect('sysctl')
            LOCATE FOR cyear == THIS.crunyear AND nrunno = THIS.nrunno AND lDisbMan AND cTypeClose = 'R'
            IF FOUND()
               llHist       = .T.
               THIS.lrelmin = sysctl.lrelmin
               THIS.lrelqtr = sysctl.lrelqtr
            ENDIF
         ELSE
            swselect('sysctl')
            LOCATE FOR cyear == THIS.crunyear AND nrunno = THIS.nrunno AND cgroup = THIS.cgroup AND cTypeClose = 'R'
            IF FOUND()
               llHist       = .T.
               THIS.lrelmin = sysctl.lrelmin
               THIS.lrelqtr = sysctl.lrelqtr
            ENDIF
         ENDIF

         SET DELETED &lcDeleted

         llReturn = llHist
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CheckHistRun', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CheckHistRun', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn

   ENDPROC

*********************************
   PROCEDURE CheckActivity
*********************************
*  Check for Any Activity During Period Range

      LOCAL llFound, llReturn, lnCount, lnflatg, lnflato, loError
      LOCAL cWellID, counte, counti, cperiod, cyear

      llReturn = .T.

      TRY

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         lnCount = 0
         STORE 0 TO m.counti, m.counte
         SELECT welltemp
         SCAN
            m.cWellID = cWellID
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               SELE wellwork
               llFound = .F.  &&  Variable for whether it found this well in wellwork, so we know whether to look at disbhist for any activity
               SCAN FOR cWellID = m.cWellID
                  m.cyear   = hyear
                  m.cperiod = hperiod
                  llFound   = .T.
                  swselect('income')
                  SCAN FOR cWellID = m.cWellID AND (nrunno = 0 OR (nrunno = THIS.nrunno AND crunyear = THIS.crunyear))
                     m.counti = m.counti + 1
                  ENDSCAN
                  SELE expense
                  SCAN FOR cWellID = m.cWellID AND (nRunNoRev = 0 OR (nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear))
                     m.counte = m.counte + 1
                  ENDSCAN
                  lnCount = lnCount + m.counti + m.counte
               ENDSCAN
               IF NOT llFound
                  SELECT disbhist
                  LOCATE FOR cWellID = m.cWellID AND nrunno = THIS.nrunno AND crunyear = THIS.crunyear
                  IF FOUND()
                     lnCount = 1
                  ENDIF
               ENDIF
            ENDIF
            lnflato = THIS.getFlatAmt(m.cWellID, 'O')
            lnflatg = THIS.getFlatAmt(m.cWellID, 'G')
            IF lnflato + lnflatg > 0
               lnCount = 1
            ENDIF
         ENDSCAN

         IF lnCount = 0
            llReturn = .F.
         ELSE
            llReturn = .T.
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CheckActivity', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CheckActivity', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*-- Check for quarterly wells.
*********************************
   PROCEDURE CheckForQuarterlyWells
*********************************
      LOCAL llReturn, loError
*
*  Checks for quarterly wells in the wells table
*
      llReturn = .T.

      TRY

         IF NOT THIS.lclose  &&  Only check this if it's a new run, and you're not closing
            IF NOT THIS.CheckHistRun()
               swselect('wells')
               LOCATE FOR nprocess = 2 AND cgroup == THIS.cgroup
               IF FOUND() AND NOT THIS.lrelqtr AND THIS.CheckQtrActivity()
                  THIS.lrelqtr = THIS.omessage.CONFIRM('Should the quarterly wells be released during this run?')
               ENDIF
            ENDIF
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CheckForQuarterlyWells', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CheckForQuarterlyWells', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN llReturn
   ENDPROC

*-- Checks for activity on quarterly wells
*********************************
   PROCEDURE CheckQtrActivity
*********************************


      SELE income.* FROM income, wells ;
         WHERE nrunno = 0 ;
         AND income.drevdate <= THIS.drevdate  ;
         AND income.cWellID = wells.cWellID ;
         AND wells.cgroup = THIS.cgroup ;
         AND wells.nprocess = 2 ;
         AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U') ;
         INTO CURSOR tempinc

      lnCount = _TALLY

      SELE expense.* FROM expense, wells ;
         WHERE nRunNoRev = 0 ;
         AND cyear <> "FIXD" ;
         AND expense.dexpdate <= THIS.dexpdate  ;
         AND expense.cWellID = wells.cWellID ;
         AND wells.cgroup = THIS.cgroup ;
         AND wells.nprocess = 2 ;
         AND NOT INLIST(wells.cwellstat, 'I', 'S', 'P', 'U') ;
         INTO CURSOR tempexp

      lnCount = lnCount + _TALLY

      SELE cWellID, .F. AS junk FROM wells INTO CURSOR tmpwell WHERE cgroup == THIS.cgroup
      STORE 0 TO lnflato, lnflatg
      llFound = .F.
      SELE tmpwell
      SCAN
         IF NOT llFound
            m.cWellID = cWellID
            lnflato   = THIS.getFlatAmt(m.cWellID, 'O')
            lnflatg   = THIS.getFlatAmt(m.cWellID, 'G')
            IF lnflato + lnflatg > 0
               llFound = .T.
            ENDIF
         ENDIF
      ENDSCAN

      lnCount = lnCount + lnflato + lnflatg

      IF lnCount # 0
         llReturn = .T.
      ELSE
         llReturn = .F.
      ENDIF

      RETURN llReturn

   ENDPROC


*-- Checks for all net wells and replaces jib runno with rev runno
*********************************
   PROCEDURE AllNetCheck
*********************************
      LOCAL llReturn, loError, cWellID

      llReturn = .T.


      TRY
*
*  Check to see if there are any JIB owners in a well.
*  If not, assume all owners are net and the expenses that fall into
*  this run will never be processed in a JIB.  Change cRunYearJIB
*  to '1900' and nRunNoJIB to match the revenue runno.
*
         SELE wells
         SCAN FOR BETWEEN(cWellID, THIS.cbegwellid, THIS.cendwellid)
            m.cWellID = cWellID
            SELE wellinv
            LOCATE FOR cWellID == m.cWellID AND ctypeinv = 'W' AND lJIB = .T.
            IF NOT FOUND()
               SELE expense
               SCAN FOR cWellID == m.cWellID AND nRunNoRev == THIS.nrunno AND cRunYearRev == THIS.crunyear AND nrunnojib = 0
                  REPL nrunnojib WITH THIS.nrunno, crunyearjib WITH '1900'
               ENDSCAN
            ENDIF
*  Next, scan for one-person expense items, and fill in the '1900' year on them, too.  Those will never be looked at for a revenue closing, so they may as well be marked here
            SELECT expense
            SCAN FOR cWellID == m.cWellID AND nRunNoRev == THIS.nrunno AND cRunYearRev == THIS.crunyear AND NOT EMPTY(cownerid)
               REPL nrunnojib WITH THIS.nrunno, crunyearjib WITH '1900'
            ENDSCAN
         ENDSCAN
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'AllNetCheck', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('AllNetCheck', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN llReturn

   ENDPROC


*-- Checks for negative backup and tax withholding and makes corrections if any are found.
*********************************
   PROCEDURE CheckBackTax
*********************************
*
*  Build a cursor of backup and tax withholding totals for the run by owner
*

      LOCAL llReturn, loError

      llReturn = .T.

      TRY
         SELE cownerid, SUM(nbackwith) AS nbackwith, SUM(ntaxwith) AS ntaxwith ;
            FROM invtmp WITH (BUFFERING = .T.);
            INTO CURSOR backtax ;
            ORDER BY cownerid ;
            GROUP BY cownerid

         IF _TALLY > 0

*  Scan through the totals looking for negative backup and/or tax withholding amounts
            SELE backtax
            SCAN
               SCATTER MEMVAR

* If there are negative backup withholding amounts, add them back into the netcheck and zero them out
               IF m.nbackwith < 0
                  SELE invtmp
                  SCAN FOR cownerid = m.cownerid
                     REPL nnetcheck WITH nnetcheck + nbackwith, ;
                        nbackwith WITH 0
                  ENDSCAN
               ENDIF

* If there are negative tax withholding amounts, add them back into the netcheck and zero them out
               IF m.ntaxwith < 0
                  SELE invtmp
                  SCAN FOR cownerid = m.cownerid
                     REPL nnetcheck WITH nnetcheck + ntaxwith, ;
                        ntaxwith  WITH 0
                  ENDSCAN
               ENDIF
            ENDSCAN
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CheckBackTax', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CheckBackTax', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN llReturn

   ENDPROC

*-- Creates the direct deposit file for owners who are direct deposited.
*********************************
   PROCEDURE DirectDeposit
*********************************
      LPARAMETERS tlTestFile, tlCCD, tlSelected

*  Creates the direct deposit file

      LOCAL m.ntotrecs, m.maxrecs, m.nreccount, m.lccrlf, m.nDebits, m.nCredits, lnCompIDLen
      LOCAL fh, oprogress, lnACount, lnBCount, lnCCount, lnDCount, lnECount, lnGlobalMin
      LOCAL m.cCompID
      LOCAL lcAcctPrd, lcAcctdate, lcBankAccount, lcBankName, lcBankTransit, lcDate, lcDisbAcct, lcMagFile
      LOCAL lcIDChec, lhold, llACHError, llReturn, lnBlocks, lnCredits, lnDebits, lnFile, lnHash
      LOCAL lnMinimum, lnMod, lnPTaxLen, loError
      LOCAL ABankName, ABlocking, ACompany, ADate, ADest, AFormat, AModifier, AOrigin, APriority
      LOCAL ARecSize, ARecType, ARefCode, ATime, BBankAcct, BBatchNo, BClass, BCompany, BDate1, BDate2
      LOCAL BDesc, BFiller, BOrigin, BRecType, BServCode, BSpace, BTaxID, CAmount, CBankABA, CBlank
      LOCAL CRecInd, CRecordType, CTrace1, CTrace2, CTransCode, Cowner, Cownid, DBatch, DBlanks
      LOCAL DCompany, DCount, DCredits, DDebits, DHash, DOrigin, DRecType, DServClass, EBatchCount
      LOCAL EBlank, EBlockCount, ECredits, EDebits, EEntries, EHash, ERecType, cBankAcctNo, cCompID
      LOCAL cPayee, cProducer, cdddest, cdddestname, lccrlf, crunyear, nCount, namount, ndisbfreq
      LOCAL nreccount, nrunno, ntotrecs, paddr1, paddr2, paddr3, pcity, pcontact, pphone, pstate, ptax
      LOCAL pzip, tlRelMin, lnDDChecks

      llReturn   = .T.
      lnDDChecks = 0

      TRY
* Don't process if the direct deposit module is not active
         IF NOT m.goApp.lDirDMDep
            llReturn = .T.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

* Check for owners that are marked for direct deposit
         SELECT investor
         LOCATE FOR ldirectdep
         IF NOT FOUND()
            llReturn = .T.
            EXIT
         ENDIF

* Check on the existence of the application object
* if it doesn't exist, we're running in development
* mode and need to initialize the company address info.
         IF TYPE('m.goApp') = 'O'
            m.cProducer = UPPER(m.goApp.cCompanyName)
            m.paddr1    = m.goApp.cAddress1
            m.paddr2    = m.goApp.cAddress2
            m.paddr3    = m.goApp.cAddress3
            m.ptax      = m.goApp.cTaxid
            m.pcity     = m.goApp.ccity
            m.pzip      = m.goApp.czip
            m.pstate    = m.goApp.cstate
            m.pcontact  = m.goApp.cContact
            m.pphone    = m.goApp.cPhoneno
         ELSE
            m.cProducer = 'Pivoten'
            m.paddr1    = '370 17th Street, Suite 3025'
            m.paddr2    = 'Denver, CO  80202'
            m.paddr3    = ''
            m.ptax      = '99-9999999'
            m.pcontact  = 'Pivoten Team'
            m.pphone    = '877-748-6836'
            m.pcity     = 'Denver'
            m.pstate    = 'CO'
            m.pzip      = '80202'
         ENDIF


         lcBankName    = UPPER(THIS.oOptions.cBankName)
         lcBankTransit = cmEncrypt(ALLTRIM(THIS.oOptions.cBankTransit), m.goApp.cEncryptionKey)
         lcBankAccount = cmEncrypt(ALLTRIM(THIS.oOptions.cBankAcct), m.goApp.cEncryptionKey)
         lcDest        = THIS.oOptions.cdddest
         lcDestname    = UPPER(THIS.oOptions.cdddestname)
         lcCompID      = THIS.oOptions.cCompID
         lcOrigDFI     = THIS.oOptions.cddorigdfi
         lcOrigin      = THIS.oOptions.cddOrigin
         lcOrigName    = UPPER(THIS.oOptions.cddOrigName)
         lcCompName    = UPPER(THIS.oOptions.cDDCompName)
         lcCompName    = swstrtran(lcCompName)
         lcCompName    = STRTRAN(lcCompName, '.', '')
         lcServeClass  = THIS.oOptions.cDDServClass
         lnGlobalMin   = THIS.oOptions.nMinCheck

            lcDisbAcct    = THIS.oOptions.cDisbAcct

         m.ptax = m.goApp.cTaxid
         m.ptax = cmEncrypt(ALLTRIM(m.ptax), m.goApp.cEncryptionKey)
* Strip out invalid characters from the tax id.
         m.ptax = STRTRAN(m.ptax, '-', '')
         m.ptax = STRTRAN(m.ptax, '/', '')
         m.ptax = STRTRAN(m.ptax, '\', '')
         m.ptax = STRTRAN(m.ptax, ' ', '')
         m.ptax = STRTRAN(m.ptax, ',', '')

         lnPTaxLen = LEN(ALLTRIM(m.ptax))

         swselect('options')
         GO TOP

         IF EMPTY(lcDestname)
            lcDestname = lcBankName
            SELECT options
            REPLACE cdddestname WITH lcDestname
         ENDIF

         IF EMPTY(lcCompName)
            lcCompName = swstrtran(UPPER(m.goApp.cCompanyName))
            lcCompName = STRTRAN(lcCompName, '.', '')
            SELECT options
            REPLACE cDDCompName WITH lcCompName
         ENDIF

         IF EMPTY(lcOrigName)
            lcOrigName = lcCompName
            SELECT options
            REPLACE cddOrigName WITH lcOrigName
         ENDIF

         IF EMPTY(lcServeClass)
            IF FILE('datafiles\no-dd-offset.txt')
               lcServeClass = '220'
            ELSE
               lcServeClass = '200'
            ENDIF
            SELECT options
            REPLACE cDDServClass WITH lcServeClass
         ENDIF

         IF EMPTY(lcDest)
            lcDest = ' ' + LEFT(ALLTRIM(lcBankTransit), 9)
            SELECT options
            REPLACE cdddest WITH lcDest
         ENDIF

         IF EMPTY(lcDestname)
            lcDestname = lcBankName
            SELECT options
            REPLACE cdddestname WITH lcDestname
         ENDIF

         IF 'CHASE' $ UPPER(lcBankName)
            IF EMPTY(lcOrigin)
               lcOrigin = '0000000000'
               SELECT options
               REPLACE cddOrigin WITH lcOrigin
            ENDIF
         ELSE
            IF FILE(m.goApp.cCommonFolder + 'BOK.txt')
               IF EMPTY(lcOrigin)
                  lcOrigin = ' ' + PADR(ALLT(m.ptax), 9, ' ')
                  SELECT options
                  REPLACE cddOrigin WITH lcOrigin
               ENDIF
            ELSE
               IF EMPTY(lcOrigin)
                  lcOrigin = '1' + PADR(ALLT(m.ptax), 9, ' ')
                  SELECT options
                  REPLACE cddOrigin WITH lcOrigin
               ENDIF
            ENDIF
         ENDIF

         IF FILE('datafiles\no-dd-offset.txt')
            IF EMPTY(lcServeClass)
               lcServeClass = '220'          && ACH Credits           02-04
               SELECT options
               REPLACE cDDServClass WITH lcServeClass
            ENDIF
         ELSE
            IF EMPTY(lcServeClass)
               lcServeClass = '200'          && ACH Credits           02-04
               SELECT options
               REPLACE cDDServClass WITH lcServeClass
            ENDIF
         ENDIF

         TABLEUPDATE(.T.,.T., 'Options')
         SELECT options
         GO TOP
         SCATTER NAME THIS.oOptions

* Fix length of lcDestname and lcOrigname to truncate to 23 chars
         lcDestname = PADR(ALLTRIM(lcDestname), 23, ' ')
         lcOrigName = PADR(ALLTRIM(lcOrigName), 23, ' ')

* Get the length of the company id field
* If the length is greater than 1 we'll use it instead
* of the company's tax id. If it is only 1 character
* we'll prepend it to the company's taxid
         lnCompIDLen    = LEN(ALLTRIM(lcCompID))


* Were minimums released this run?
         tlRelMin    = THIS.lrelmin

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Creating Direct Deposit ACH Entries...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         m.cProducer = UPPER(m.goApp.oOptions.cDDCompName)

         IF USED('tempdep')
            swclose('tempdep')
         ENDIF

         IF NOT tlSelected
            IF NOT tlTestFile
               SELE invtmp.cownerid, investor.cownname, investor.cBankTransit, investor.cBankAcct, ;
                  SUM(invtmp.nnetcheck) AS ntotal ;
                  FROM invtmp, investor ;
                  WHERE investor.ldirectdep = .T. ;
                  AND investor.lFedwire = .F. ;
                  AND invtmp.nnetcheck # 0 ;
                  AND invtmp.cownerid == investor.cownerid ;
                  INTO CURSOR tempdep READWRITE ;
                  ORDER BY invtmp.cownerid GROUP BY invtmp.cownerid
            ELSE
               SELE investor.cownerid, investor.cownname, investor.cBankTransit, investor.cBankAcct, ;
                  000000.00 AS ntotal ;
                  FROM investor ;
                  WHERE investor.ldirectdep = .T. ;
                  AND investor.lFedwire = .F. ;
                  INTO CURSOR tempdep READWRITE ;
                  ORDER BY investor.cownerid GROUP BY investor.cownerid
            ENDIF
         ELSE
            IF NOT tlTestFile
               SELE invtmp.cownerid, investor.cownname, investor.cBankTransit, investor.cBankAcct, ;
                  SUM(invtmp.nnetcheck) AS ntotal ;
                  FROM invtmp, investor ;
                  WHERE investor.ldirectdep = .T. ;
                  AND investor.lFedwire = .F. ;
                  AND invtmp.nnetcheck # 0 ;
                  AND invtmp.cownerid IN (SELECT cID FROM SELECTED) ;
                  AND invtmp.cownerid == investor.cownerid ;
                  INTO CURSOR tempdep READWRITE ;
                  ORDER BY invtmp.cownerid GROUP BY invtmp.cownerid
            ELSE
               SELE investor.cownerid, investor.cownname, investor.cBankTransit, investor.cBankAcct, ;
                  000000.00 AS ntotal ;
                  FROM investor ;
                  WHERE investor.ldirectdep = .T. ;
                  AND investor.lFedwire = .F. ;
                  AND investor.cownerid IN (SELECT cID FROM SELECTED) ;
                  INTO CURSOR tempdep READWRITE ;
                  ORDER BY investor.cownerid GROUP BY investor.cownerid
            ENDIF
         ENDIF


         IF _TALLY = 0
* No direct deposits to process
            llReturn = .T.
            EXIT
         ELSE
            lnDDChecks = _TALLY
         ENDIF

* Check to see if there is any activity for direct deposit owners.
         IF NOT tlTestFile
            SELECT tempdep
            LOCATE FOR ntotal # 0
            IF NOT FOUND()
               EXIT
            ENDIF
         ENDIF

         STORE 0 TO lnACount, lnBCount, lnCCount, lnDCount, lnECount, lnDebits, lnCredits, lnHash, m.ntotrecs

         IF NOT FILE(m.goApp.cCommonFolder + 'ddnocrlf.txt')
            m.lccrlf = CHR(13) + CHR(10)
         ELSE
            m.lccrlf = ''
         ENDIF

         fh         = ' '
         lcDate     = DTOC(DATE())
         lcAcctdate = DTOC(THIS.ddirectdate)

*
*  Get the Accounting Month
*
         lcAcctPrd = PADL(ALLTRIM(STR(MONTH(THIS.dacctdate), 2)), 2, '0')

*
*  File Header Record
*

         ARecType  = '1'           && Record Type Code        01
         APriority = '01'          && Priority Code          02-03
         ADest     = lcDest       && Bank ABA             04-13
         AOrigin   = lcOrigin
         ADate     = SUBSTR(lcDate, 9, 2) + SUBSTR(lcDate, 1, 2) + SUBSTR(lcDate, 4, 2)
         ATime     = SUBSTR(TIME(), 1, 2) + SUBSTR(TIME(), 4, 2)
         AModifier = 'A'           && File Modifier          34
         ARecSize  = '094'         && Record Size            35-37
         ABlocking = '10'          && Blocking Factor        38-39
         AFormat   = '1'           && Format Code            40
         ADestName = lcDestname    &&       41-63
         AOrigName = lcOrigName    &&       64-86

         IF 'CHASE' $ UPPER(lcBankName)
            ARefCode = '        '
         ELSE
            ARefCode    = 'SherWare'    && Reference Code         87-94
         ENDIF

*
*  Company/Batch Header Record
*
         BRecType  = '5'            && Record Type           01
         BServCode = lcServeClass          && ACH Credits           02-04
         BCompany  = PADR(ALLT(lcCompName), 16, ' ') &&       05-20

         IF 'CHASE' $ UPPER(lcBankName)
            BFiller = PADL(ALLTRIM(lcBankAccount), 20, '0')
         ELSE
            BFiller   = SPACE(20)      && Filler                21-40
         ENDIF

         DO CASE
            CASE lnCompIDLen < 1
               BTaxID      = '1' + PADR(ALLT(m.ptax), 9, ' ') && 41-50
            CASE lnCompIDLen = 1
               BTaxID      = ALLTRIM(lcCompID) + PADR(ALLTRIM(m.ptax), 9, ' ') && 41-50
            OTHERWISE
               BTaxID      = PADR(ALLTRIM(lcCompID), 10, ' ')    && 41-50
         ENDCASE

         IF tlCCD
            BClass    = 'CCD'          && Standard Entry Class  51-53
         ELSE
            BClass    = 'PPD'          && Standard Entry Class  51-53
         ENDIF
         BDesc     = 'REV DIST  '   && Description           54-63
         BDate1    = SUBSTR(lcAcctdate, 9, 2) + SUBSTR(lcAcctdate, 1, 2) + SUBSTR(lcAcctdate, 4, 2)
         BDate2    = BDate1
         BSpace    = '   '          && Settlement Date       76-78
         BOrigin   = '1'            && Originator Status     79
         BBankAcct = SUBSTR(lcOrigDFI, 1, 8) &&           80-87
         BBatchNo  = '0000001'      && Batch Number          88-94

*
* Entry Detail Record
*
         CRecordType = '6'            && Record Type           01
         CTransCode  = '22'           && Trans Code            02-03
         CBankABA    = SPACE(9)       && Bank Transit          04-12
         cBankAcctNo = SPACE(17)      && Bank Account          13-29
         CAmount     = SPACE(10)      && Amount                30-39
         Cownid      = SPACE(15)      && Owner ID              40-54
         Cowner      = SPACE(22)      && Owner Name            55-76
         CBlank      = SPACE(2)       && Blank                 77-78
         CRecInd     = '0'            && Addenda Indicator      79
         CTrace1     = SUBSTR(ALLTRIM(lcBankTransit), 1, 8)
         CTrace2     = SPACE(7)

*
* Addenda record
*
         AdRecordType = '7'
         AdType       = '05'
         AddendaInfo  = SPACE(80)
         AdSeq        = '0001'
         AdEntrySeq   = SPACE(7)

*
*  Company Record Control
*
         DRecType   = '8'
         DServClass = lcServeClass
         DCount     = SPACE(6)
         DHash      = SPACE(10)
         DDebits    = '000000000000'
         DCredits   = '000000000000'
         DO CASE
            CASE lnCompIDLen < 1
               DCompany    = '1' + PADR(ALLTRIM(m.ptax), 9, ' ')
            CASE lnCompIDLen = 1
               DCompany    = ALLTRIM(lcCompID) + PADR(ALLT(m.ptax), 9, ' ')
            OTHERWISE
               DCompany    = PADR(ALLTRIM(lcCompID), 10, ' ')
         ENDCASE
         DBlanks = SPACE(25)
         DOrigin = PADR(LEFT(lcOrigDFI, 8), 8, ' ')
         DBatch  = '0000001'

*
* File Control Record
*
         ERecType    = '9'
         EBatchCount = '000001'
         EBlockCount = '000000'
         EEntries    = '00000000'
         EHash       = '0000000000'
         EDebits     = '000000000000'
         ECredits    = '000000000000'
         EBlank      = SPACE(39)

         m.nreccount = 9999999

* Create the ACH files in an ACH directory
         IF NOT m.goApp.lCloudServer
            IF NOT DIRECTORY(m.goApp.cdatafilepath + 'ACH')
               MD (m.goApp.cdatafilepath + 'ACH')
            ENDIF
         ELSE
            IF NOT DIRECTORY('S:\ACH')
               MD ('S:\ACH')
            ENDIF
         ENDIF

         IF NOT tlTestFile
            IF NOT m.goApp.lCloudServer
               lcMagFile = m.goApp.cdatafilepath + 'ACH\DP_' + THIS.crunyear + PADL(ALLT(STR(THIS.nrunno)), 3, '0') + '.txt'
            ELSE
               lcMagFile = 'S:\ACH\DP_' + swstrtran(ALLTRIM(m.cProducer)) + '_' + THIS.crunyear + PADL(ALLT(STR(THIS.nrunno)), 3, '0') + '.txt'
            ENDIF
         ELSE
            IF NOT m.goApp.lCloudServer
               lcMagFile = m.goApp.cdatafilepath + 'ACH\DP_TEST_FILE.TXT'
            ELSE
               lcMagFile = 'S:\ACH\DP_' + swstrtran(ALLTRIM(m.cProducer)) + '_TEST_FILE.TXT'
            ENDIF
         ENDIF

         IF NOT tlTestFile
            lnFile    = 1
            DO WHILE FILE(lcMagFile)
               lcMagFile = JUSTSTEM(lcMagFile)
               IF ATC('_', lcMagFile) > 0
                  lcMagFile = SUBSTR(lcMagFile, 1, LEN(lcMagFile) - 2)
               ENDIF
               IF NOT m.goApp.lCloudServer
                  lcMagFile = m.goApp.cdatafilepath + 'ach\' + JUSTSTEM(lcMagFile) + '_' + TRANSFORM(lnFile) + '.txt'
               ELSE
                  lcMagFile = 'S:\ach\' + JUSTSTEM(lcMagFile) + '_' + TRANSFORM(lnFile) + '.txt'
               ENDIF
               lnFile    = lnFile + 1
            ENDDO
         ENDIF

         llACHError = .F.

         lcString = ARecType + APriority + ADest + AOrigin + ;
            ADate + ATime + AModifier + ARecSize + ;
            ABlocking + AFormat + ADestName + AOrigName + ARefCode + m.lccrlf

         llReturn = STRTOFILE(lcString, lcMagFile, 0)

         m.ntotrecs = m.ntotrecs + 1

*  Write Batch Header

         lcString = BRecType + BServCode + BCompany + BFiller + BTaxID + BClass + BDesc + ;
            BDate1 + BDate2 + BSpace + BOrigin + BBankAcct + BBatchNo + m.lccrlf
         llReturn = STRTOFILE(lcString, lcMagFile, 1)

         m.ntotrecs = m.ntotrecs + 1

         lnACount    = lnACount + 1
         m.nreccount = 1

* Get the next ach transaction #
         IF m.goApp.lPartnershipMod
            IF NOT FILE(m.goApp.cdatafilepath + 'achnumber.txt')
               lcACHNumber = '1001'
               STRTOFILE(lcACHNumber, m.goApp.cdatafilepath + 'achnumber.txt')
            ELSE
               lcACHNumber = FILETOSTR(m.goApp.cdatafilepath + 'achnumber.txt')
            ENDIF
            lnACHNumber = INT(VAL(lcACHNumber))
         ENDIF


         SELECT tempdep
         SCAN
            SCATTER MEMVAR

            IF m.ntotal = 0 AND NOT tlTestFile  && There shouldn't be a zero amount ach if it's not a test file
               lnDDChecks = lnDDChecks - 1
               LOOP
            ENDIF

            SELE investor
            LOCATE FOR cownerid == m.cownerid
            IF FOUND()
               m.lhold     = lhold
               m.ndisbfreq = ndisbfreq
               IF ninvmin # 0
                  lnMinimum = ninvmin
               ELSE
                  lnMinimum = lnGlobalMin
               ENDIF
               m.cPayee = cownname
               IF NOT EMPTY(investor.caddenda)
                  m.caddenda = caddenda
                  CRecInd    = '1'
               ELSE
                  m.caddenda = ''
                  CRecInd    = '0'
               ENDIF
               IF NOT tlTestFile
                  IF investor.caccttype = 'S'
                     CTransCode = '32'
                  ELSE
                     CTransCode = '22'
                  ENDIF
               ELSE
                  IF investor.caccttype = 'S'
                     CTransCode = '33'
                  ELSE
                     CTransCode = '23'
                  ENDIF
               ENDIF

            ELSE
* Shouldn't get here...
               LOOP
            ENDIF

*
*  Reset the minimum amount to hold this owner's check
*  if it's not to be disbursed monthly or he's on hold.
*
            DO CASE
               CASE m.lhold                   && Owner on hold
                  lnMinimum = 99999999
               CASE m.ndisbfreq = 2          && Quarterly
                  IF NOT INLIST(lcAcctPrd, '03', '06', '09', '12')
                     lnMinimum = 99999999
                  ELSE
                     IF tlRelMin
                        lnMinimum = 0
                     ENDIF
                  ENDIF

               CASE m.ndisbfreq = 3          && SemiAnnually
                  IF NOT INLIST(lcAcctPrd, '06', '12')
                     lnMinimum = 99999999
                  ELSE
                     IF tlRelMin
                        lnMinimum = 0
                     ENDIF
                  ENDIF
               CASE m.ndisbfreq = 4          && Annually
                  IF lcAcctPrd # '12'
                     lnMinimum = 99999999
                  ELSE
                     IF tlRelMin
                        lnMinimum = 0
                     ENDIF
                  ENDIF
               CASE tlRelMin                 && Release minimums
                  lnMinimum = 0
            ENDCASE

            IF tlTestFile
               lnMinimum = 0
            ENDIF

            IF m.ntotal < lnMinimum
               SELE tempdep
               DELE NEXT 1
               LOOP
            ENDIF

            THIS.ogl.cidtype    = 'I'
            THIS.ogl.cSource    = 'DM'
            THIS.ogl.cUnitNo    = ''
            THIS.ogl.cdeptno    = ''
            THIS.ogl.cEntryType = 'C'
            THIS.ogl.cID        = m.cownerid
            THIS.ogl.namount    = m.ntotal
            THIS.ogl.cPayee     = m.cPayee
            THIS.ogl.cBatch     = THIS.cdmbatch
            THIS.ogl.cAcctNo    = lcDisbAcct
            THIS.ogl.lPrinted   = .T.
            THIS.ogl.lCleared   = .F.
            THIS.ogl.drecdate   = {}

            IF NOT m.goApp.lPartnershipMod
               THIS.ogl.ccheckno   = '    DIRDEP'
            ELSE
               THIS.ogl.ccheckno = 'E' + PADL(TRANSFORM(lnACHNumber), 6, '0')
               lnACHNumber       = lnACHNumber + 1
            ENDIF
            THIS.ogl.dpostdate  = THIS.ddirectdate
            THIS.ogl.dCheckDate = THIS.ddirectdate
            IF NOT tlTestFile
               IF m.ntotal = 0
                  lcIDChec = ''
               ELSE
                  THIS.ogl.addcheck(.T.)
                  lcIDChec = THIS.ogl.cidchec
               ENDIF
               SELE invtmp
               SCAN FOR cownerid == m.cownerid
                  REPL cidchec WITH lcIDChec
               ENDSCAN
            ELSE
               lcIDChec = '*****'
            ENDIF

            lnCCount = lnCCount + 1

            m.cBankTransit = cmEncrypt(ALLTRIM(m.cBankTransit), m.goApp.cEncryptionKey)
            m.cBankAcct    = cmEncrypt(ALLTRIM(m.cBankAcct), m.goApp.cEncryptionKey)

            CBankABA    = PADR(ALLT(m.cBankTransit), 9, ' ')
            cBankAcctNo = PADR(ALLT(m.cBankAcct), 17, ' ')
            CAmount     = PADL(STRTRAN(ALLT(STR(m.ntotal, 12, 2)), '.', ''), 10, '0')
            Cownid      = PADR(STRTRAN(m.cownerid, '-', ''), 15, ' ')
            m.cownname  = swstrtran(m.cownname)
            Cowner      = PADR(ALLT(m.cownname), 22, ' ')
            CTrace2     = PADL(ALLTRIM(STR(lnCCount)), 7, '0')

            lcString = CRecordType + CTransCode + CBankABA + ;
               cBankAcctNo + CAmount + Cownid + Cowner + CBlank + ;
               CRecInd + CTrace1 + CTrace2 + m.lccrlf
            llReturn = STRTOFILE(lcString, lcMagFile, 1)

            IF NOT EMPTY(m.caddenda)
               AddendaInfo = PADR(ALLTRIM(m.caddenda), 80, ' ')
               AdEntrySeq  = CTrace2
               lcString    = AdRecordType + AdType + AddendaInfo + AdSeq + AdEntrySeq + m.lccrlf
               lnCCount    = lnCCount + 1
               m.ntotrecs  = m.ntotrecs + 1
               llReturn    = STRTOFILE(lcString, lcMagFile, 1)
            ENDIF

            m.ntotrecs = m.ntotrecs + 1
            lnHash     = lnHash + VAL(LEFT(ALLTRIM(m.cBankTransit), 8))
            lnCredits  = lnCredits + m.ntotal

         ENDSCAN

* Store the next ach transaction number
         IF m.goApp.lPartnershipMod
            STRTOFILE(TRANSFORM(lnACHNumber + 1), m.goApp.cdatafilepath + 'achnumber.txt')
         ENDIF

* Add DDA Debit for company
         IF THIS.oOptions.lddalternate
            CRecordType = '6'
            CTransCode  = '27'
            CBankABA    = PADR(ALLT(lcBankTransit), 9, ' ')
            cBankAcctNo = PADR(ALLT(lcBankAccount), 17, ' ')
            CAmount     = PADL(STRTRAN(ALLT(STR(lnCredits, 12, 2)), '.', ''), 10, '0')
            m.ntotrecs  = m.ntotrecs + 1
            lnCCount    = lnCCount + 1
            Cownid      = PADR('OPER', 15, ' ')
            Cowner      = PADR(ALLT(m.cProducer), 22, ' ')
            CTrace2     = PADL(ALLTRIM(STR(lnCCount)), 7, '0')

            lcString = CRecordType + CTransCode + CBankABA + ;
               cBankAcctNo + CAmount + Cownid + Cowner + CBlank + ;
               CRecInd + CTrace1 + CTrace2 + m.lccrlf

            llReturn = STRTOFILE(lcString, lcMagFile, 1)

            lnHash  = lnHash + VAL(LEFT(lcBankTransit, 8))
            DDebits = PADL(STRTRAN(ALLT(STR(lnCredits, 12, 2)), '.', ''), 12, '0')
         ENDIF

* Company Record
         DCount   = PADL(ALLT(STR(lnCCount)), 6, '0')
         DHash    = PADL(ALLT(STR(lnHash)), 10, '0')
         DCredits = PADL(STRTRAN(ALLT(STR(lnCredits, 12, 2)), '.', ''), 12, '0')

         lcString = DRecType + DServClass + DCount + DHash + ;
            DDebits + DCredits + DCompany + DBlanks + DOrigin + DBatch + m.lccrlf

         llReturn = STRTOFILE(lcString, lcMagFile, 1)

         m.ntotrecs = m.ntotrecs + 1

* File Control Record

         m.ntotrecs = m.ntotrecs + 1
         lnBlocks   = INT((m.ntotrecs * 94) / 940)
         lnMod      = MOD((m.ntotrecs * 94) / 940, 1)
         lnExtra    = 0
         IF lnMod # 0
            lnExtra  = 10 - (lnMod * 10)
            lnBlocks = lnBlocks + 1
         ENDIF
         EBlockCount = PADL(ALLT(STR(lnBlocks)), 6, '0')
         EEntries    = PADL(ALLTRIM(STR(lnCCount)), 8, '0')
         EHash       = DHash
         ECredits    = DCredits
         EDebits     = DDebits

         lcString = ERecType + EBatchCount + EBlockCount + EEntries + ;
            EHash + EDebits + ECredits + EBlank + m.lccrlf

         llReturn = STRTOFILE(lcString, lcMagFile, 1)

         IF lnExtra > 0
            lcString = REPLICATE('9', 94) + m.lccrlf
            FOR lnX = 1 TO lnExtra
               llReturn = STRTOFILE(lcString, lcMagFile, 1)
            ENDFOR
         ENDIF

         WAIT CLEAR


         THIS.ndirectdeptotal = lnCredits
* THIS.omessage.DISPLAY('Created ' + ALLTRIM(STR(lnCCount)) + ' Direct Deposit Records Totaling: ' + ALLT(STR(lnCredits,12,2)))
         IF NOT FILE(m.goApp.cdatafilepath + 'dirdep.dbf')
            CREATE TABLE m.goApp.cdatafilepath + 'dirdep' FREE ;
               (nrunno        i, ;
                 crunyear      c(4), ;
                 nCount        i, ;
                 namount       N(12, 2))
         ENDIF

         IF NOT USED('dirdep')
            USE (m.goApp.cdatafilepath + 'dirdep') IN 0
         ENDIF

         SELE dirdep
         LOCATE FOR nrunno = THIS.nrunno AND crunyear = THIS.crunyear
         IF FOUND()
            REPL nCount  WITH lnDDChecks, ;
               namount WITH lnCredits
         ELSE
            m.crunyear = THIS.crunyear
            m.nrunno   = THIS.nrunno
            IF THIS.oOptions.lddalternate
               m.nCount   = lnCCount - 1  && Don't count the offset entry
            ELSE
               m.nCount = lnCCount
            ENDIF
            m.namount  = lnCredits
            INSERT INTO dirdep FROM MEMVAR
         ENDIF

         swclose('tempdep1')
         swclose('tempdep')

         llReturn = .T.

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'DirectDeposit', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('DirectDeposit', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      THIS.fedwire()

      RETURN llReturn

   ENDPROC

*-- Creates the direct deposit ccd file for owners who are federal.
*********************************
   PROCEDURE FedWire
*********************************
      LPARAMETERS tlTestFile

*  Creates the direct deposit file

      LOCAL m.ntotrecs, m.maxrecs, m.nreccount, m.lccrlf, m.nDebits, m.nCredits, lnCompIDLen
      LOCAL fh, oprogress, lnACount, lnBCount, lnCCount, lnDCount, lnECount, lnGlobalMin
      LOCAL m.cCompID
      LOCAL lcAcctPrd, lcAcctdate, lcBankAccount, lcBankName, lcBankTransit, lcDate, lcDisbAcct, lcMagFile
      LOCAL lcIDChec, lhold, llACHError, llReturn, lnBlocks, lnCredits, lnDebits, lnFile, lnHash
      LOCAL lnMinimum, lnMod, lnPTaxLen, loError
      LOCAL ABankName, ABlocking, ACompany, ADate, ADest, AFormat, AModifier, AOrigin, APriority
      LOCAL ARecSize, ARecType, ARefCode, ATime, BBankAcct, BBatchNo, BClass, BCompany, BDate1, BDate2
      LOCAL BDesc, BFiller, BOrigin, BRecType, BServCode, BSpace, BTaxID, CAmount, CBankABA, CBlank
      LOCAL CRecInd, CRecordType, CTrace1, CTrace2, CTransCode, Cowner, Cownid, DBatch, DBlanks
      LOCAL DCompany, DCount, DCredits, DDebits, DHash, DOrigin, DRecType, DServClass, EBatchCount
      LOCAL EBlank, EBlockCount, ECredits, EDebits, EEntries, EHash, ERecType, cBankAcctNo, cCompID
      LOCAL cPayee, cProducer, cdddest, cdddestname, lccrlf, crunyear, nCount, namount, ndisbfreq
      LOCAL nreccount, nrunno, ntotrecs, paddr1, paddr2, paddr3, pcity, pcontact, pphone, pstate, ptax
      LOCAL pzip, tlRelMin, lnDDChecks

      llReturn   = .T.
      lnDDChecks = 0

      TRY
* Don't process if the direct deposit module is not active
         IF NOT m.goApp.lDirDMDep
            llReturn = .T.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

* Check for owners that are marked for direct deposit
         SELECT investor
         LOCATE FOR lFedwire
         IF NOT FOUND()
            llReturn = .T.
            EXIT
         ENDIF

         lcBankName    = THIS.oOptions.cBankName
         lcBankTransit = cmEncrypt(ALLTRIM(THIS.oOptions.cBankTransit), m.goApp.cEncryptionKey)
         lcBankAccount = cmEncrypt(ALLTRIM(THIS.oOptions.cBankAcct), m.goApp.cEncryptionKey)
         m.cdddest     = THIS.oOptions.cdddest
         m.cdddestname = THIS.oOptions.cdddestname
         m.cCompID     = THIS.oOptions.cCompID
         lnGlobalMin   = THIS.oOptions.nMinCheck

         IF THIS.oOptions.lFedwireClr = .T.
            IF NOT EMPTY(THIS.oOptions.cFedwire)
               lcDisbAcct = THIS.oOptions.cFedwire
            ELSE
               lcDisbAcct = THIS.oOptions.cDisbAcct
            ENDIF
         ELSE
            lcDisbAcct    = THIS.oOptions.cDisbAcct
         ENDIF

* Get the length of the company id field
* If the length is greater than 1 we'll use it instead
* of the company's tax id. If it is only 1 character
* we'll prepend it to the company's taxid
         lnCompIDLen    = LEN(ALLTRIM(m.cCompID))

* Were minimums released this run?
         tlRelMin    = THIS.lrelmin

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Creating Direct Deposit ACH Entries...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

* Check on the existence of the application object
* if it doesn't exist, we're running in development
* mode and need to initialize the company address info.
         IF TYPE('m.goApp') = 'O'
            m.cProducer = m.goApp.cCompanyName
            m.paddr1    = m.goApp.cAddress1
            m.paddr2    = m.goApp.cAddress2
            m.paddr3    = m.goApp.cAddress3
            m.ptax      = m.goApp.cTaxid
            m.pcity     = m.goApp.ccity
            m.pzip      = m.goApp.czip
            m.pstate    = m.goApp.cstate
            m.pcontact  = m.goApp.cContact
            m.pphone    = m.goApp.cPhoneno
         ELSE
            m.cProducer = 'Pivoten'
            m.paddr1    = '370 17th Street, Suite 3025'
            m.paddr2    = 'Denver, CO  80202'
            m.paddr3    = ''
            m.ptax      = '99-9999999'
            m.pcontact  = 'Pivoten Team'
            m.pphone    = '877-748-6836'
            m.pcity     = 'Denver'
            m.pstate    = 'CO'
            m.pzip      = '80202'
         ENDIF

         IF USED('tempdep')
            swclose('tempdep')
         ENDIF

         IF NOT tlTestFile
            SELE invtmp.cownerid, ;
               investor.cownname, ;
               investor.cBankTransit, ;
               investor.cBankAcct, ;
               investor.cpayor, ;
               investor.cpad, ;
               investor.cfedindian, ;
               investor.cfundcode, ;
               SUM(invtmp.nnetcheck) AS ntotal ;
               FROM invtmp, investor ;
               WHERE investor.lFedwire = .T. ;
               AND invtmp.nnetcheck # 0 ;
               AND invtmp.cownerid == investor.cownerid ;
               INTO CURSOR tempdep READWRITE ;
               ORDER BY invtmp.cownerid GROUP BY invtmp.cownerid
         ELSE
            SELE investor.cownerid, ;
               investor.cownname, ;
               investor.cBankTransit, ;
               investor.cBankAcct, ;
               investor.cpayor, ;
               investor.cpad, ;
               investor.cfedindian, ;
               investor.cfundcode, ;
               000000.00 AS ntotal ;
               FROM investor ;
               WHERE investor.lFedwire = .T. ;
               INTO CURSOR tempdep READWRITE ;
               ORDER BY investor.cownerid GROUP BY investor.cownerid
         ENDIF

         IF _TALLY = 0
* No direct deposits to process
            llReturn = .T.
            EXIT
         ELSE
            lnDDChecks = _TALLY
         ENDIF

* Check to see if there is any activity for direct deposit owners.
         IF NOT tlTestFile
            SELECT tempdep
            LOCATE FOR ntotal # 0
            IF NOT FOUND()
               EXIT
            ENDIF
         ENDIF

         STORE 0 TO lnACount, lnBCount, lnCCount, lnDCount, lnECount, lnDebits, lnCredits, lnHash, m.ntotrecs

         IF NOT FILE(m.goApp.cCommonFolder + 'ddnocrlf.txt')
            m.lccrlf = CHR(13) + CHR(10)
         ELSE
            m.lccrlf = ''
         ENDIF

         fh         = ' '
         lcDate     = DTOC(DATE())
         lcAcctdate = DTOC(THIS.ddirectdate)

         m.ptax    = cmEncrypt(ALLTRIM(m.ptax), m.goApp.cEncryptionKey)
* Strip out invalid characters from the tax id.
         m.ptax = STRTRAN(m.ptax, '-', '')
         m.ptax = STRTRAN(m.ptax, '/', '')
         m.ptax = STRTRAN(m.ptax, '\', '')
         m.ptax = STRTRAN(m.ptax, ' ', '')
         m.ptax = STRTRAN(m.ptax, ',', '')

         lnPTaxLen = LEN(ALLTRIM(m.ptax))
*
*  Get the Accounting Month
*
         lcAcctPrd = PADL(ALLTRIM(STR(MONTH(THIS.dacctdate), 2)), 2, '0')

*
*  File Header Record
*

         ARecType  = '1'           && Record Type Code        01
         APriority = '01'          && Priority Code          02-03
         IF NOT THIS.oOptions.lddalternate
            ADest   = ' ' + PADR(ALLT(lcBankTransit), 9, ' ')        && Bank ABA             04-13
            AOrigin = '1' + PADR(ALLT(m.ptax), 9, ' ')
         ELSE
            ADest   = ' ' + PADR(ALLT(m.cdddest), 9, ' ')        && Bank ABA             04-13
            AOrigin = '1' + PADR(ALLT(lcBankTransit), 9, ' ')
         ENDIF
         ADate     = SUBSTR(lcDate, 9, 2) + SUBSTR(lcDate, 1, 2) + SUBSTR(lcDate, 4, 2)
         ATime     = SUBSTR(TIME(), 1, 2) + SUBSTR(TIME(), 4, 2)
         AModifier = 'A'           && File Modifier          34
         ARecSize  = '094'         && Record Size            35-37
         ABlocking = '10'          && Blocking Factor        38-39
         AFormat   = '1'           && Format Code            40

         IF NOT THIS.oOptions.lddalternate
            ABankName = PADR(ALLT(lcBankName), 23, ' ') &&       41-63
            ACompany  = PADR(ALLT(m.cProducer), 23, ' ') &&       64-86
         ELSE
            IF FILE('datafiles\ddalt.txt')
               ABankName = PADR(ALLT(lcBankName), 23, ' ') &&       41-63
               ACompany  = PADR(ALLT(m.cProducer), 23, ' ') &&       64-86
            ELSE
               ABankName = PADR(ALLT(m.cdddestname), 23, ' ') &&       41-63
               ACompany  = PADR(ALLT(lcBankName), 23, ' ') &&       64-86
            ENDIF
         ENDIF
         ARefCode    = 'SherWare'    && Reference Code         87-94

*
*  Company/Batch Header Record
*
         BRecType  = '5'            && Record Type           01
         BServCode = '200'          && ACH Credits           02-04
         BCompany  = PADR(ALLT(m.cProducer), 16, ' ') &&       05-20
         BFiller   = SPACE(20)      && Filler                21-40

         DO CASE
            CASE lnCompIDLen < 1
               BTaxID      = '1' + PADR(ALLT(m.ptax), 9, ' ') && 41-50
            CASE lnCompIDLen = 1
               BTaxID      = ALLTRIM(m.cCompID) + PADR(ALLT(m.ptax), 9, ' ') && 41-50
            OTHERWISE
               BTaxID      = PADR(ALLTRIM(m.cCompID), 10, ' ')    && 41-50
         ENDCASE

         BClass    = 'CCD'          && Standard Entry Class  51-53
         BDesc     = 'Rev Dist  '   && Description           54-63
         BDate1    = SUBSTR(lcAcctdate, 9, 2) + SUBSTR(lcAcctdate, 1, 2) + SUBSTR(lcAcctdate, 4, 2)
         BDate2    = BDate1
         BSpace    = '   '          && Settlement Date       76-78
         BOrigin   = '1'            && Originator Status     79
         BBankAcct = SUBSTR(lcBankTransit, 1, 8) &&           80-87
         BBatchNo  = '0000001'      && Batch Number          88-94

*
* Entry Detail Record
*
         CRecordType = '6'            && Record Type           01
         CTransCode  = '22'           && Trans Code            02-03
         CBankABA    = SPACE(9)       && Bank Transit          04-12
         cBankAcctNo = SPACE(17)      && Bank Account          13-29
         CAmount     = SPACE(10)      && Amount                30-39
         Cownid      = SPACE(15)      && Owner ID              40-54
         Cowner      = SPACE(22)      && Owner Name            55-76
         CBlank      = SPACE(2)       && Blank                 77-78
         CRecInd     = '1'            && Record Indicator      79
         CTrace1     = SUBSTR(lcBankTransit, 1, 8)
         CTrace2     = SPACE(7)


*
*  Addenda Record
*
         FRecType  = '7'
         FAddenda  = '05'
         FFreeForm = SPACE(80)
         FSequence = PADR(LEFT(lcBankTransit, 8), 8, ' ')
         FBatch    = '0001'
         FTrace    = SPACE(7)

*
*  Company Record Control
*
         DRecType   = '8'
         DServClass = '200'
         DCount     = SPACE(6)
         DHash      = SPACE(10)
         DDebits    = '000000000000'
         DCredits   = '000000000000'
         DO CASE
            CASE lnCompIDLen < 1
               DCompany    = '1' + PADR(ALLT(m.ptax), 9, ' ')
            CASE lnCompIDLen = 1
               DCompany    = ALLTRIM(m.cCompID) + PADR(ALLT(m.ptax), 9, ' ')
            OTHERWISE
               DCompany    = PADR(ALLTRIM(m.cCompID), 10, ' ')
         ENDCASE
         DBlanks = SPACE(25)
         DOrigin = PADR(LEFT(lcBankTransit, 8), 8, ' ')
         DBatch  = '0000001'

*
* File Control Record
*
         ERecType    = '9'
         EBatchCount = '000001'
         EBlockCount = '000000'
         EEntries    = '00000000'
         EHash       = '0000000000'
         EDebits     = '000000000000'
         ECredits    = '000000000000'
         EBlank      = SPACE(39)

         m.nreccount = 9999999

* Create the ACH files in an ACH directory
         IF NOT DIRECTORY(m.goApp.cdatafilepath + 'ACH')
            MD (m.goApp.cdatafilepath + 'ACH')
         ENDIF

         IF NOT tlTestFile
            IF NOT m.goApp.lCloudServer
               lcMagFile = m.goApp.cdatafilepath + 'ACH\FW' + THIS.crunyear + PADL(ALLT(STR(THIS.nrunno)), 3, '0') + '.txt'
            ELSE
               lcMagFile = 'S:\ACH\FW' + THIS.crunyear + PADL(ALLT(STR(THIS.nrunno)), 3, '0') + '.txt'
            ENDIF
         ELSE
            IF NOT m.goApp.lCloudServer
               lcMagFile = m.goApp.cdatafilepath + 'ACH\FW_TEST_FILE.TXT'
            ELSE
               lcMagFile = 'S:\ACH\FW_TEST_FILE.TXT'
            ENDIF
         ENDIF

         IF NOT tlTestFile
            lnFile    = 1
            DO WHILE FILE(lcMagFile)
               lcMagFile = JUSTSTEM(lcMagFile)
               IF ATC('_', lcMagFile) > 0
                  lcMagFile = SUBSTR(lcMagFile, 1, LEN(lcMagFile) - 2)
               ENDIF
               IF NOT m.goApp.lCloudServer
                  lcMagFile = m.goApp.cdatafilepath + 'ach\' + JUSTSTEM(lcMagFile) + '_' + TRANSFORM(lnFile) + '.txt'
               ELSE
                  lcMagFile = 'S:\ach\' + JUSTSTEM(lcMagFile) + '_' + TRANSFORM(lnFile) + '.txt'
               ENDIF
               lnFile    = lnFile + 1
            ENDDO
         ENDIF

         llACHError = .F.

         lcString = ARecType + APriority + ADest + AOrigin + ;
            ADate + ATime + AModifier + ARecSize + ;
            ABlocking + AFormat + ABankName + ACompany + ARefCode + m.lccrlf

         llReturn   = STRTOFILE(lcString, lcMagFile, 0)
         m.ntotrecs = m.ntotrecs + 1

*  Write Batch Header

         lcString = BRecType + BServCode + BCompany + BFiller + BTaxID + BClass + BDesc + ;
            BDate1 + BDate2 + BSpace + BOrigin + BBankAcct + BBatchNo + m.lccrlf
         llReturn = STRTOFILE(lcString, lcMagFile, 1)

         m.ntotrecs = m.ntotrecs + 1

         lnACount    = lnACount + 1
         m.nreccount = 1

         SELECT tempdep
         SCAN
            SCATTER MEMVAR
            SELE investor
            LOCATE FOR cownerid == m.cownerid
            IF FOUND()
               m.lhold     = lhold
               m.ndisbfreq = ndisbfreq
               IF ninvmin # 0
                  lnMinimum = ninvmin
               ELSE
                  lnMinimum = lnGlobalMin
               ENDIF
               m.cPayee = cownname
               IF NOT tlTestFile
                  IF investor.caccttype = 'S'
                     CTransCode = '32'
                  ELSE
                     CTransCode = '22'
                  ENDIF
               ELSE
                  IF investor.caccttype = 'S'
                     CTransCode = '33'
                  ELSE
                     CTransCode = '23'
                  ENDIF
               ENDIF

            ELSE
* Shouldn't get here...
               LOOP
            ENDIF

*
*  Reset the minimum amount to hold this owner's check
*  if it's not to be disbursed monthly or he's on hold.
*
            DO CASE
               CASE m.lhold                   && Owner on hold
                  lnMinimum = 99999999
               CASE m.ndisbfreq = 2          && Quarterly
                  IF NOT INLIST(lcAcctPrd, '03', '06', '09', '12')
                     lnMinimum = 99999999
                  ELSE
                     IF tlRelMin
                        lnMinimum = 0
                     ENDIF
                  ENDIF

               CASE m.ndisbfreq = 3          && SemiAnnually
                  IF NOT INLIST(lcAcctPrd, '06', '12')
                     lnMinimum = 99999999
                  ELSE
                     IF tlRelMin
                        lnMinimum = 0
                     ENDIF
                  ENDIF
               CASE m.ndisbfreq = 4          && Annually
                  IF lcAcctPrd # '12'
                     lnMinimum = 99999999
                  ELSE
                     IF tlRelMin
                        lnMinimum = 0
                     ENDIF
                  ENDIF
               CASE tlRelMin                 && Release minimums
                  lnMinimum = 0
            ENDCASE

            IF tlTestFile
               lnMinimum = 0
            ENDIF

            IF m.ntotal < lnMinimum
               SELE tempdep
               DELE NEXT 1
               LOOP
            ENDIF

            THIS.ogl.cidtype    = 'I'
            THIS.ogl.cSource    = 'DM'
            THIS.ogl.cUnitNo    = ''
            THIS.ogl.cdeptno    = ''
            THIS.ogl.cEntryType = 'C'
            THIS.ogl.cID        = m.cownerid
            THIS.ogl.namount    = m.ntotal
            THIS.ogl.cPayee     = m.cPayee
            THIS.ogl.cBatch     = THIS.cdmbatch
            THIS.ogl.cAcctNo    = lcDisbAcct
            THIS.ogl.lPrinted   = .T.
            THIS.ogl.ccheckno   = '   FEDWIRE'
            THIS.ogl.dpostdate  = THIS.ddirectdate
            THIS.ogl.dCheckDate = THIS.ddirectdate
            IF NOT tlTestFile
               THIS.ogl.addcheck(.T.)
               lcIDChec = THIS.ogl.cidchec
               SELE invtmp
               SCAN FOR cownerid == m.cownerid
                  REPL cidchec WITH lcIDChec
               ENDSCAN
            ELSE
               lcIDChec = '*****'
            ENDIF

            lnCCount = lnCCount + 1

            m.cBankTransit = cmEncrypt(ALLTRIM(m.cBankTransit), m.goApp.cEncryptionKey)
            m.cBankAcct    = cmEncrypt(ALLTRIM(m.cBankAcct), m.goApp.cEncryptionKey)

            CBankABA    = PADR(ALLT(m.cBankTransit), 9, ' ')
            cBankAcctNo = PADR(ALLT(m.cBankAcct), 17, ' ')
            CAmount     = PADL(STRTRAN(ALLT(STR(m.ntotal, 12, 2)), '.', ''), 10, '0')
            Cownid      = PADR(m.cownerid, 15, ' ')
            Cowner      = PADR(ALLT(m.cownname), 22, ' ')
            CTrace2     = PADL(ALLTRIM(STR(lnCCount)), 7, '0')

            lcString = CRecordType + CTransCode + CBankABA + ;
               cBankAcctNo + CAmount + Cownid + Cowner + CBlank + ;
               CRecInd + CTrace1 + CTrace2 + m.lccrlf
            llReturn = STRTOFILE(lcString, lcMagFile, 1)

            m.ntotrecs = m.ntotrecs + 1
            lnHash     = lnHash + VAL(LEFT(m.cBankTransit, 8))
            lnCredits  = lnCredits + m.ntotal

            FFreeForm  = PADR(m.cpayor + '*' + ALLTRIM(UPPER(LEFT(m.cpad, 4))) + ;
                 PADL(TRANSFORM(MONTH(THIS.dacctdate)), 2, '0') + ;
                 RIGHT(TRANSFORM(YEAR(THIS.dacctdate)), 2) + ;
                 '*' + m.cfedindian + '*' + ;
                 IIF(EMPTY(m.cfundcode), '*', m.cfundcode + '*'), 80, ' ')
            FTrace     = CTrace2
            lnCCount   = lnCCount + 1
            lcString   = FRecType + FAddenda + FFreeForm + FBatch + FTrace + m.lccrlf
            llReturn   = STRTOFILE(lcString, lcMagFile, 1)
            m.ntotrecs = m.ntotrecs + 1

         ENDSCAN

* Add DDA Debit for company
         IF THIS.oOptions.lddalternate
            CRecordType = '6'
            CTransCode  = '27'
            CBankABA    = PADR(ALLT(lcBankTransit), 9, ' ')
            cBankAcctNo = PADR(ALLT(lcBankAccount), 17, ' ')
            CAmount     = PADL(STRTRAN(ALLT(STR(lnCredits, 12, 2)), '.', ''), 10, '0')
            m.ntotrecs  = m.ntotrecs + 1
            lnCCount    = lnCCount + 1
            Cownid      = PADR('OPER', 15, ' ')
            Cowner      = PADR(ALLT(m.cProducer), 22, ' ')
            CTrace2     = PADL(ALLTRIM(STR(lnCCount)), 7, '0')

            lcString = CRecordType + CTransCode + CBankABA + ;
               cBankAcctNo + CAmount + Cownid + Cowner + CBlank + ;
               CRecInd + CTrace1 + CTrace2 + m.lccrlf

            llReturn = STRTOFILE(lcString, lcMagFile, 1)

            lnHash  = lnHash + VAL(LEFT(lcBankTransit, 8))
            DDebits = PADL(STRTRAN(ALLT(STR(lnCredits, 12, 2)), '.', ''), 12, '0')
         ENDIF

* Company Record
         DCount   = PADL(ALLT(STR(lnCCount)), 6, '0')
         DHash    = PADL(ALLT(STR(lnHash)), 10, '0')
         DCredits = PADL(STRTRAN(ALLT(STR(lnCredits, 12, 2)), '.', ''), 12, '0')

         lcString = DRecType + DServClass + DCount + DHash + ;
            DDebits + DCredits + DCompany + DBlanks + DOrigin + DBatch + m.lccrlf

         llReturn = STRTOFILE(lcString, lcMagFile, 1)

         m.ntotrecs = m.ntotrecs + 1

* File Control Record

         m.ntotrecs = m.ntotrecs + 1
         lnBlocks   = INT((m.ntotrecs * 94) / 940)
         lnMod      = MOD((m.ntotrecs * 94) / 940, 1)
         lnExtra    = 0
         IF lnMod # 0
            lnExtra  = 10 - (lnMod * 10)
            lnBlocks = lnBlocks + 1
         ENDIF
         EBlockCount = PADL(ALLT(STR(lnBlocks)), 6, '0')
         EEntries    = PADL(ALLTRIM(STR(lnCCount)), 8, '0')
         EHash       = DHash
         ECredits    = DCredits
         EDebits     = DDebits

         lcString = ERecType + EBatchCount + EBlockCount + EEntries + ;
            EHash + EDebits + ECredits + EBlank + m.lccrlf

         llReturn = STRTOFILE(lcString, lcMagFile, 1)

         IF lnExtra > 0
            lcString = REPLICATE('9', 94) + m.lccrlf
            FOR lnX = 1 TO lnExtra
               llReturn = STRTOFILE(lcString, lcMagFile, 1)
            ENDFOR
         ENDIF

         WAIT CLEAR


         THIS.ndirectdeptotal = lnCredits
* THIS.omessage.DISPLAY('Created ' + ALLTRIM(STR(lnCCount)) + ' Direct Deposit Records Totaling: ' + ALLT(STR(lnCredits,12,2)))
         IF NOT FILE(m.goApp.cdatafilepath + 'dirdep.dbf')
            CREATE TABLE m.goApp.cdatafilepath + 'dirdep' FREE ;
               (nrunno        i, ;
                 crunyear      c(4), ;
                 nCount        i, ;
                 namount       N(12, 2))
         ENDIF

         IF NOT USED('dirdep')
            USE (m.goApp.cdatafilepath + 'dirdep') IN 0
         ENDIF

         SELE dirdep
         LOCATE FOR nrunno = THIS.nrunno AND crunyear = THIS.crunyear
         IF FOUND()
            REPL nCount  WITH nCount  + lnCCount, ;
               namount WITH namount + lnCredits
         ELSE
            m.crunyear = THIS.crunyear
            m.nrunno   = THIS.nrunno
            m.nCount   = lnCCount
            m.namount  = lnCredits
            INSERT INTO dirdep FROM MEMVAR
         ENDIF

         swclose('tempdep1')
         swclose('tempdep')

         llReturn = .T.

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'FedWire', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('FedWire', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn

   ENDPROC

*-- Calculates check digit from bank aba
*********************************
   PROCEDURE DP_CheckDigit
*********************************
      LPARA tcABA
      LOCAL lndig1, lndig2, lndig3, lndig4, lndig5, lndig6, lndig7, lndig8
      LOCAL lndigits, lnfactor, lncheckdigit, lccheckdigit
      LOCAL lcReturn, loError

      lcReturn = '*'

      TRY
         IF NOT TYPE('tcaba') = 'C'
            WAIT WIND 'The Bank ABA number must be passed to dp_checkdigit'
            lcReturn = '*'
            EXIT
         ENDIF


         lndig1 = VAL(SUBST(tcABA, 1, 1)) * 3
         lndig2 = VAL(SUBST(tcABA, 2, 1)) * 7
         lndig3 = VAL(SUBST(tcABA, 3, 1)) * 1
         lndig4 = VAL(SUBST(tcABA, 4, 1)) * 3
         lndig5 = VAL(SUBST(tcABA, 5, 1)) * 7
         lndig6 = VAL(SUBST(tcABA, 6, 1)) * 1
         lndig7 = VAL(SUBST(tcABA, 7, 1)) * 3
         lndig8 = VAL(SUBST(tcABA, 8, 1)) * 7

         lndigits = lndig1 + lndig2 + lndig3 + lndig4 + lndig5 + lndig6 + lndig7 + lndig8

         lnfactor = lndigits / 10

         lnfactor = ROUND(lnfactor, 0)

         lnfactor = lnfactor * 10

         lncheckdigit = lnfactor - lndigits

         lccheckdigit = STR(lncheckdigit, 1)

         lcReturn    = lccheckdigit

      CATCH TO loError
         lcReturn = '*'
         DO errorlog WITH 'DP_CheckDigit', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('DP_CheckDigit', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN lcReturn
   ENDPROC

*********************************
   PROCEDURE RemoveTaxExempt
*********************************
      LPARA tcwellid, tcType, tnAmount
      LOCAL lnShare, m.cownerid, m.nrevoil, m.nrevgas

*
* NOT USED ANYMORE
*
      RETURN

*
* Removes the tax exempt owners share from the gross amount before taxes
* are calculated on the gross.  - For New Mexico Wells
*

      lnShare = 0

      swselect('wellinv')
      SCAN FOR cWellID == tcwellid
         m.cownerid = cownerid
         m.nrevoil  = nrevoil
         m.nrevgas  = nrevgas
         swselect('investor')
         SET ORDER TO cownerid
         IF SEEK(m.cownerid) AND lExempt
            DO CASE
               CASE tcType = 'BBL'
                  lnShare = lnShare + swround((tnAmount * m.nrevoil / 100), 2)
               CASE tcType = 'MCF'
                  lnShare = lnShare + swround((tnAmount * m.nrevgas / 100), 2)
            ENDCASE
         ENDIF
      ENDSCAN

      lnAmount = tnAmount - lnShare

      RETURN (lnAmount)

   ENDPROC

*********************************
   PROCEDURE GrossUpTaxPct
*********************************
      LPARA tcwellid, tnPct, tcType
      LOCAL lnPct, lnNewPct
      LOCAL lnReturn, loError
      LOCAL m.cownerid
*
*  Gross up the given percentage to account for tax exempt owners
*

      lnReturn = tnPct

      TRY
         lnPct    = 0
         lnNewPct = 0

         SELE investor
         SET ORDER TO cownerid

         SELE wellinv
         SCAN FOR cWellID == tcwellid
            m.cownerid = cownerid

            SELE investor
            IF SEEK(m.cownerid) AND lExempt
               IF tcType = 'G'
                  lnPct = lnPct + wellinv.nrevtax2
               ELSE
                  lnPct = lnPct + wellinv.nrevtax1
               ENDIF
            ENDIF
         ENDSCAN

         IF lnPct # 0
            lnNewPct = tnPct / (100 - lnPct) * 100
         ELSE
            lnNewPct = tnPct
         ENDIF

         lnReturn = lnNewPct

      CATCH TO loError
         lnReturn = .F.
         DO errorlog WITH 'GrossUpTaxPct', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('GrossUpTaxPct', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)

      ENDTRY

      RETURN lnReturn

   ENDPROC

*-- Suspense processing
*********************************
   PROCEDURE Suspense
*********************************
      LPARAMETERS tlNetDef, tlClose
* Don't process suspense for well reports
      LOCAL llReturn, loError

      llReturn = .T.

      TRY

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               THIS.cErrorMsg = 'Processing canceled by user.'
               EXIT
            ENDIF
         ENDIF

         IF THIS.cprocess = 'W'
            EXIT
         ENDIF

         IF THIS.lrunclosed
            EXIT
         ENDIF

         swclose('tsuspense')
         swclose('suspense1')
         swclose('disbhist1')
         Make_Copy('suspense', 'TSuspense')

* Close the suspense table because it might
* be a temporary table that has stuff in it we
* don't want.
         IF tlClose
            swclose('suspense')

* Open the live suspense table
            swselect('suspense', .T.)
         ENDIF

*  If we're not closing the run, don't change the real suspense file.
*  Instead, get a copy of it to work with.
         IF NOT THIS.lclose
            SELECT suspense.* FROM suspense NOFILTER INTO CURSOR tempsuspx
            swclose('suspense')
            USE DBF('tempsuspx') AGAIN IN 0 ALIAS suspense
            SELE suspense
            INDEX ON ciddisb TAG ciddisb
            INDEX ON crectype TAG crectype
            INDEX ON cownerid + crectype TAG owner_type
            INDEX ON cgroup TAG cgroup
            INDEX ON cownerid TAG cownerid
            INDEX ON cWellID TAG cWellID
            INDEX ON cownerid + cWellID TAG invwell
            INDEX ON crunyear_in + PADL(TRANSFORM(nrunno_in), 3, '0') TAG run_in
            INDEX ON nrunno_in TAG nrunno_in
            INDEX ON crunyear_in TAG crunyear
            INDEX ON DELETED() TAG _deleted

         ENDIF


         THIS.osuspense.crunyear  = THIS.cnewrunyear
         THIS.osuspense.nrunno    = THIS.nnewrunno
         THIS.osuspense.cgroup    = THIS.cgroup
         THIS.osuspense.lClosing  = THIS.lclose
         THIS.osuspense.dacctdate = THIS.dacctdate
         THIS.osuspense.lrelmin   = THIS.lrelmin
         THIS.osuspense.cBegOwner = THIS.cbegownerid
         THIS.osuspense.cEndOwner = THIS.cendownerid
         THIS.osuspense.lNewRun   = THIS.lNewRun
         THIS.osuspense.lrelqtr   = THIS.lrelqtr

         IF tlNetDef
            THIS.oprogress.SetProgressMessage('Processing Suspense by Owner...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1

            lnProgressCnt  = THIS.nprogress
            llReturn       = THIS.osuspense.owner_suspense(THIS.oprogress, @lnProgressCnt)
            THIS.nprogress = lnProgressCnt

            IF NOT llReturn
               THIS.cErrorMsg = 'Processing failed in Owner_Suspense'
               EXIT
            ENDIF
         ELSE
            THIS.oprogress.SetProgressMessage('Processing Suspense by Well...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
            lnProgressCnt  = THIS.nprogress
            THIS.osuspense.well_suspense(THIS.oprogress, @lnProgressCnt)
            THIS.nprogress = lnProgressCnt
         ENDIF

* Create a temp suspense file with the buffered data
* for speed purposes
         THIS.oprogress.SetProgressMessage('Creating temporary suspense file...')
         swselect('suspense')
         SELECT tsuspense
         INDEX ON cownerid TAG cownerid
         INDEX ON cownerid + cWellID TAG invwell
         INDEX ON csusptype TAG csusptype
         INDEX ON cWellID  TAG cWellID
         INDEX ON crunyear_in TAG crunyearin
         INDEX ON nrunno_in TAG nrunnoin
         INDEX ON crunyear_in + PADL(TRANSFORM(nrunno_in), 3, '0') TAG runyear
         WAIT CLEAR
         THIS.osuspense = .NULL.

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'Suspense', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('Suspense', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE postsummary
*********************************
      LOCAL tcYear, tcPeriod, tdCheckDate, tcGroup, tdPostDate
      LOCAL lnMax, lnCount, lnTotal, lcName, lnJIBInv, m.cCustName, lcDMBatch, llSepClose
      LOCAL lcRevClear, lcSuspense, m.cDisbAcct, m.cVendComp, m.cGathAcct, m.cBackWith
      LOCAL llIntegComp, llSepClose, lcAPAcct, lcAcctYear, lcAcctMonth, lcIDChec, llExpSum
      LOCAL llRound, llRoundIt, lnOwner
      LOCAL lDirGasPurch, lDirOilPurch, lcAcctC, lcAcctD, lcDMExp, lcDeptNo, lcDesc, lcExpClear, llJIB
      LOCAL llJibNet, llNoPostDM, llReturn, lnAmount, lnBackWith, lnCheck, lnCompress, lnCredits, lnDebits
      LOCAL lnDeficit, lnExpenses, lnFreqs, lnGTax1, lnGTax1p, lnGTax2, lnGTax2p, lnGTax3, lnGTax3p
      LOCAL lnGTax4, lnGTax4p, lnGasRev, lnGasTax, lnGathering, lnHold, lnHolds, lnIntHold, lnMarketing
      LOCAL lnMi1Rev, lnMi2Rev, lnMin, lnMinimum, lnOTax1, lnOTax1p, lnOTax2, lnOTax2p, lnOTax3, lnOTax3p
      LOCAL lnOTax4, lnOTax4p, lnOilRev, lnOilTax, lnOthRev, lnPTax1, lnPTax1p, lnPTax2, lnPTax2p, lnPTax3
      LOCAL lnPTax3p, lnPTax4, lnPTax4p, lnRevenue, lnTaxWith, lnTrpRev, lnVendor, loError
      LOCAL cBackWith, cBatch, cCRAcctV, cCustName, cDRAcctV, cDefAcct, cDisbAcct, cGathAcct, cID
      LOCAL cMinAcct, cownerid, cRevSource, cTaxAcct1, cTaxAcct2, cTaxAcct3, cTaxAcct4, cUnitNo
      LOCAL cVendComp, cWellID, ccateg, cidchec, cownname, csusptype, nCompress, nGather, namount
      LOCAL nprocess, ntotal, tdCompanyPost

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting to General Ledger...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnMax       = 0
         lnCount     = 0
         lnTotal     = 0
         lcName      = 'Owner'
         m.cCustName = ' '
         lcIDChec    = ''

         lcAcctYear  = STR(YEAR(THIS.dpostdate), 4)
         lcAcctMonth = PADL(ALLTRIM(STR(MONTH(THIS.dpostdate), 2)), 2, '0')

*  Set the posting dates
         IF THIS.lAdvPosting = .T.
            tdCompanyPost = THIS.dCompanyShare
            tdPostDate    = THIS.dCheckDate
         ELSE
            tdCompanyPost = THIS.dCheckDate
            tdPostDate    = THIS.dCheckDate
         ENDIF

*  Plug the DM batch number into glmaint so that each
*  batch created can be traced to this closing
         THIS.ogl.DMBatch  = THIS.cdmbatch
         THIS.ogl.cSource  = 'DM'
         THIS.ogl.nDebits  = 0
         THIS.ogl.nCredits = 0
         THIS.ogl.dGLDate  = THIS.dCheckDate

*  Get the suspense account from glopt
         swselect('glopt')
         lcSuspense = cSuspense
         IF EMPTY(lcSuspense)
            lcSuspense = '999999'
         ENDIF
         lcRevClear = crevclear
         lcExpClear = cexpclear
         llNoPostDM = lDMNoPost

*  Get the A/P account
         swselect('apopt')
         lcAPAcct = capacct

*  Set up the parameters used by processing in this method
         tcYear   = THIS.crunyear
         tcPeriod = THIS.cperiod
         tcGroup  = THIS.cgroup

*   Get Disbursement Checking Acct Number
         m.cDisbAcct = THIS.oOptions.cDisbAcct
         IF EMPTY(ALLT(m.cDisbAcct))
            m.cDisbAcct = lcSuspense
         ENDIF

         m.cVendComp = THIS.oOptions.cVendComp

         m.cGathAcct = THIS.oOptions.cGathAcct
         IF EMPTY(ALLT(m.cGathAcct))
            m.cGathAcct = lcSuspense
         ENDIF
         m.cBackWith = THIS.oOptions.cBackAcct
         IF EMPTY(ALLT(m.cBackWith))
            m.cBackWith = lcSuspense
         ENDIF
         m.cTaxAcct1  = THIS.oOptions.cTaxAcct1
         IF EMPTY(ALLT(m.cTaxAcct1))
            m.cTaxAcct1 = lcSuspense
         ENDIF
         m.cTaxAcct2  = THIS.oOptions.cTaxAcct2
         IF EMPTY(ALLT(m.cTaxAcct2))
            m.cTaxAcct2 = lcSuspense
         ENDIF
         m.cTaxAcct3  = THIS.oOptions.cTaxAcct3
         IF EMPTY(ALLT(m.cTaxAcct3))
            m.cTaxAcct3 = lcSuspense
         ENDIF
         m.cTaxAcct4 = THIS.oOptions.cTaxAcct4
         IF EMPTY(ALLT(m.cTaxAcct4))
            m.cTaxAcct4 = lcSuspense
         ENDIF
         m.cDefAcct  = THIS.oOptions.cDefAcct
         IF EMPTY(ALLT(m.cDefAcct))
            m.cDefAcct = lcSuspense
         ENDIF
         m.cMinAcct  = THIS.oOptions.cMinAcct
         IF EMPTY((m.cMinAcct))
            m.cMinAcct = lcSuspense
         ENDIF
         lcDMExp = THIS.oOptions.cFixedAcct
         IF EMPTY(lcDMExp)
            lcDMExp = lcAPAcct
         ENDIF
         IF m.goApp.lAMVersion
            lcDeptNo         = THIS.oOptions.cdeptno
            THIS.ogl.cdeptno = lcDeptNo
         ELSE
            lcDeptNo = ''
         ENDIF

         llExpSum    = THIS.oOptions.lexpsum

         llJibNet    = .T.
         IF TYPE('m.goApp') = 'O'
* Turn off net jib processing for disb mgr
            IF m.goApp.ldmpro
               llJibNet = .F.
* Don't create journal entries for stand-alone disb mgr
               llNoPostDM = .T.
            ENDIF
         ENDIF
         llSepClose  = .T.

* Get the suspense types before this run so we know how to post the owners
         THIS.osuspense.GetLastType(.F., .T., THIS.cgroup, .T.)

*   Check to see if vendor compression & gathering is to be posted
         llIntegComp = .F.

         IF NOT EMPTY(ALLT(m.cVendComp))
            swselect('vendor')
            SET ORDER TO cvendorid
            IF SEEK(m.cVendComp)
               IF lIntegGL
                  llIntegComp = .T.
               ENDIF
            ENDIF
         ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Compression and Gathering to General Ledger...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

*   Post compression and gathering
         THIS.ogl.cBatch = GetNextPK('BATCH')

         IF NOT m.goApp.ldmpro
            SELE SUM(wellwork.nCompress) AS nCompress, ;
               SUM(wellwork.nGather)   AS nGather ;
               FROM wellwork ;
               JOIN wells ON wells.cWellID = wellwork.cWellID ;
               WHERE (wells.lcompress OR wells.lGather) ;
               INTO CURSOR tempcomp

            IF _TALLY > 0
               SELE tempcomp
               m.nCompress         = nCompress
               m.nGather           = nGather
               THIS.ogl.cReference = 'Period: ' + THIS.cyear + '/' + THIS.cperiod + '/' + THIS.cgroup
               THIS.ogl.cyear      = THIS.cyear
               THIS.ogl.cperiod    = THIS.cperiod
               THIS.ogl.dCheckDate = THIS.dacctdate
               IF llIntegComp
                  THIS.ogl.dGLDate  = tdCompanyPost
               ELSE
                  THIS.ogl.dGLDate  = tdPostDate
               ENDIF
               THIS.ogl.cDesc   = 'Compression/Gathering'
               THIS.ogl.cID     = ''
               THIS.ogl.cidtype = ''
               THIS.ogl.cSource = 'DM'
               IF NOT EMPTY(ALLT(m.cVendComp))
                  THIS.ogl.cAcctNo = lcDMExp
               ELSE
                  THIS.ogl.cAcctNo = m.cGathAcct
               ENDIF
               THIS.ogl.cgroup     = THIS.cgroup
               THIS.ogl.cEntryType = 'C'
               THIS.ogl.cUnitNo    = ''
               THIS.ogl.namount    = (m.nCompress + m.nGather) * -1
               THIS.ogl.UpdateBatch()

               THIS.ogl.cAcctNo = THIS.cexpclear
               THIS.ogl.namount = (m.nCompress + m.nGather)
               THIS.ogl.UpdateBatch()
            ENDIF
         ENDIF

*   Create Investor and Vendor Checks
         llReturn = THIS.ownerchks()
         IF NOT llReturn
            EXIT
         ENDIF
         llReturn = THIS.vendorchks()
         IF NOT llReturn
            EXIT
         ENDIF
         llReturn = THIS.directdeposit()
         IF NOT llReturn
            EXIT
         ENDIF


         STORE 0 TO lnRevenue, lnExpenses, lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnCheck, lnMinimum, lnDeficit, lnHolds, lnFreqs
         STORE 0 TO lnCompress, lnGathering, lnMarketing, lnBackWith, lnTaxWith, lnIntHold, lnHold, lnCheck
         STORE 0 TO lnGTax1, lnGTax2, lnGTax3, lnGTax4, lnPTax1, lnPTax2, lnPTax3, lnPTax4
         STORE 0 TO lnOTax1p, lnOTax2p, lnOTax3p, lnOTax4p, lnGTax1p, lnGTax2p, lnGTax3p, lnGTax4p, lnPTax1p, lnPTax2p, lnPTax3p, lnPTax4p

         swselect('wells')
         SET ORDER TO cWellID

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Owner Checks to General Ledger (Summary)...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         IF NOT llNoPostDM

* Get a cursor of owners to be posted from invtmp
            SELECT  cownerid,;
                    SUM(nnetcheck) AS ntotal ;
                FROM invtmp WITH (BUFFERING = .T.) ;
                ORDER BY cownerid;
                GROUP BY cownerid ;
                INTO CURSOR tmpowners READWRITE
            lnMax = _TALLY
            INDEX ON cownerid TAG owner

            lnCount = 1

            SELECT tmpowners
            SCAN
               m.cownerid = cownerid

               IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
                  llReturn          = .F.
                  IF NOT m.goApp.CancelMsg()
                     THIS.lCanceled = .T.
                     EXIT
                  ENDIF
               ENDIF

               IF THIS.lclose
* Show progressbar only when closing the run.
                  THIS.oprogress.SetProgressMessage('Posting Owner Checks - (Summary)...' + m.cownerid)
               ENDIF

               m.ntotal        = ntotal
               THIS.ogl.cBatch = GetNextPK('BATCH')
               swselect('investor')
               SET ORDER TO cownerid
               IF SEEK(m.cownerid)
                  m.cownname     = cownname
* Don't post "Dummy" owner amounts
                  IF investor.ldummy
                     LOOP
                  ENDIF
* Don't post owners that are transfered to G/L here.
                  IF investor.lIntegGL
                     LOOP
                  ENDIF
               ELSE
                  LOOP
               ENDIF

               m.cidchec = ' '
               SELECT invtmp
               SCAN FOR cownerid = m.cownerid AND csusptype = ' '
                  IF (nIncome = 0 AND nexpense = 0 AND nsevtaxes = 0 AND nnetcheck = 0)
                     LOOP
                  ENDIF
                  SCATTER MEMVAR
                  lcIDChec = m.cidchec

                  swselect('wells')
                  IF SEEK(m.cWellID)
                     SCATTER FIELDS LIKE lSev* MEMVAR
                     m.lDirOilPurch = lDirOilPurch
                     m.lDirGasPurch = lDirGasPurch
                  ELSE
                     m.lDirOilPurch = .F.
                     m.lDirGasPurch = .F.
                  ENDIF

                  lnRevenue   = lnRevenue + m.nIncome
*  Remove direct paid amounts
                  DO CASE
                     CASE m.cdirect = 'O'
                        lnRevenue = lnRevenue - m.noilrev
                     CASE m.cdirect = 'G'
                        lnRevenue = lnRevenue - m.ngasrev
                     CASE m.cdirect = 'B'
                        lnRevenue = lnRevenue - m.noilrev - m.ngasrev
                  ENDCASE

* Post the sev taxes
                  IF m.noiltax1 # 0
                     IF NOT m.lsev1o
                        IF NOT m.lDirOilPurch
                           lnOTax1 = lnOTax1 - m.noiltax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax1 = lnOTax1 - m.noiltax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax1p = lnOTax1p - m.noiltax1
                     ENDIF
                  ENDIF

                  IF m.ngastax1 # 0
                     IF NOT m.lsev1g
                        IF NOT m.lDirGasPurch
                           lnGTax1 = lnGTax1 - m.ngastax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax1 = lnGTax1 - m.ngastax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax1p = lnGTax1p - m.ngastax1
                     ENDIF
                  ENDIF

                  IF m.nOthTax1 # 0
                     IF NOT m.lsev1p
                        lnPTax1 = lnPTax1 - m.nOthTax1
                     ELSE
                        lnPTax1p = lnPTax1p - m.nOthTax1
                     ENDIF
                  ENDIF

                  IF m.noiltax2 # 0
                     IF NOT m.lsev2o
                        IF NOT m.lDirOilPurch
                           lnOTax2 = lnOTax2 - m.noiltax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax2 = lnOTax2 - m.noiltax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax2p = lnOTax2p - m.noiltax2
                     ENDIF
                  ENDIF

                  IF m.ngastax2 # 0
                     IF NOT m.lsev2g
                        IF NOT m.lDirGasPurch
                           lnGTax2 = lnGTax2 - m.ngastax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax2 = lnGTax2 - m.ngastax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax2p = lnGTax2p - m.ngastax2
                     ENDIF
                  ENDIF

                  IF m.nOthTax2 # 0
                     IF NOT m.lsev2p
                        lnPTax2 = lnPTax2 - m.nOthTax2
                     ELSE
                        lnPTax2p = lnPTax2p - m.nOthTax2
                     ENDIF
                  ENDIF

                  IF m.noiltax3 # 0
                     IF NOT m.lsev3o
                        IF NOT m.lDirOilPurch
                           lnOTax3 = lnOTax3 - m.noiltax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax3 = lnOTax3 - m.noiltax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax3p = lnOTax3p - m.noiltax3
                     ENDIF
                  ENDIF

                  IF m.ngastax3 # 0
                     IF NOT m.lsev3g
                        IF NOT m.lDirGasPurch
                           lnGTax3 = lnGTax3 - m.ngastax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax3 = lnGTax3 - m.ngastax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax3p = lnGTax3p - m.ngastax3
                     ENDIF
                  ENDIF

                  IF m.nOthTax3 # 0
                     IF NOT m.lsev3p
                        lnPTax3 = lnPTax3 - m.nOthTax3
                     ELSE
                        lnPTax3p = lnPTax3p - m.nOthTax3
                     ENDIF
                  ENDIF

                  IF m.noiltax4 # 0
                     IF NOT m.lsev4o
                        IF NOT m.lDirOilPurch
                           lnOTax4 = lnOTax4 - m.noiltax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax4 = lnOTax4 - m.noiltax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax4p = lnOTax4p - m.noiltax4
                     ENDIF
                  ENDIF

                  IF m.ngastax4 # 0
                     IF NOT m.lsev4g
                        IF NOT m.lDirGasPurch
                           lnGTax4 = lnGTax4 - m.ngastax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax4 = lnGTax4 - m.ngastax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax4p = lnGTax4p - m.ngastax4
                     ENDIF
                  ENDIF

                  IF m.nOthTax4 # 0
                     IF NOT m.lsev4p
                        lnPTax4 = lnPTax4 - m.nOthTax4
                     ELSE
                        lnPTax4p = lnPTax4p - m.nOthTax4
                     ENDIF
                  ENDIF

*  Post compression and gathering
                  IF m.nCompress # 0
                     lnCompress = lnCompress - m.nCompress
                  ENDIF

                  IF m.nGather # 0
                     lnGathering = lnGathering - m.nGather
                  ENDIF

*  Post marketing expenses
                  IF m.nMKTGExp # 0
                     lnMarketing = lnMarketing - m.nMKTGExp
                  ENDIF

*  Post the Expenses 
                  lnExpenses = lnExpenses + m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + m.ntotalea + m.ntotaleb

*  Post Backup Withholding
                  IF m.nbackwith # 0
                     lnBackWith = lnBackWith - m.nbackwith
                  ENDIF
*  Post Tax Withholding
                  IF m.ntaxwith # 0
                     lnTaxWith = lnTaxWith - m.ntaxwith
                  ENDIF
               ENDSCAN

* Post prior suspense
               SELECT invtmp
               SCAN FOR cownerid = m.cownerid AND csusptype <> ' '
                  IF nIncome = 0 AND nexpense = 0 AND nsevtaxes = 0 AND nnetcheck = 0
                     LOOP
                  ENDIF

                  SCATTER MEMVAR

*  Post Prior Period Deficits
                  IF m.csusptype = 'D'
                     lnDeficit = lnDeficit - m.nnetcheck
                  ENDIF

*  Post Prior Period Minimums
                  IF m.csusptype = 'M'
                     lnMinimum = lnMinimum + m.nnetcheck
                  ENDIF

*  Post Interest on Hold being released
                  IF m.csusptype = 'I'
                     lnIntHold = lnIntHold + m.nnetcheck
                  ENDIF

*  Post Owner on Hold being released
                  IF m.csusptype = 'H'
                     lnHolds = lnHolds + m.nnetcheck
                  ENDIF

*  Post Quarterly Owner being released
                  IF INLIST(m.csusptype, 'Q', 'S', 'A')
                     lnFreqs = lnFreqs + m.nnetcheck
                  ENDIF
*!*                       ENDIF
               ENDSCAN

* Post Check Amount To Cash
               lnCheck = lnCheck + m.ntotal

            ENDSCAN
            lcIDChec = ''

*  Post amounts going into suspense this run
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Posting Owner Suspense to General Ledger (Summary)...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            SELECT  cownerid, ;
                    csusptype, ;
                    SUM(nbackwith) AS nbackwith, ;
                    SUM(ntaxwith) AS ntaxwith, ;
                    SUM(nnetcheck) AS ntotal ;
                FROM tsuspense  ;
                WHERE nrunno_in = THIS.nrunno ;
                    AND crunyear_in = THIS.crunyear ;
                ORDER BY cownerid,;
                    csusptype ;
                GROUP BY cownerid,;
                    csusptype ;
                INTO CURSOR tmpowners READWRITE

            SELECT tmpowners
            SCAN
               SCATTER MEMVAR
               IF THIS.lclose
* Show progressbar only when closing the run.
                  THIS.oprogress.SetProgressMessage('Posting Owner Suspense - (Summary)...' + m.cownerid)
               ENDIF
               SELECT tsuspense
               SCAN FOR cownerid == m.cownerid ;
                     AND csusptype = m.csusptype ;
                     AND nrunno_in = THIS.nrunno ;
                     AND crunyear_in = THIS.crunyear
                  SCATTER MEMVAR

                  swselect('wells')
                  IF SEEK(m.cWellID)
                     SCATTER FIELDS LIKE lSev* MEMVAR
                     m.lDirOilPurch = lDirOilPurch
                     m.lDirGasPurch = lDirGasPurch
                  ELSE
                     m.lDirOilPurch = .F.
                     m.lDirGasPurch = .F.
                  ENDIF

                  lnRevenue   = lnRevenue + m.nIncome
*  Remove direct paid amounts
                  DO CASE
                     CASE m.cdirect = 'O'
                        lnRevenue = lnRevenue - m.noilrev
                     CASE m.cdirect = 'G'
                        lnRevenue = lnRevenue - m.ngasrev
                     CASE m.cdirect = 'B'
                        lnRevenue = lnRevenue - m.noilrev - m.ngasrev
                  ENDCASE

*!*                       IF m.nflatrate <> 0
*!*                           lnRevenue = lnRevenue + m.nflatrate
*!*                       ENDIF

* Post the sev taxes
                  IF m.noiltax1 # 0
                     IF NOT m.lsev1o
                        IF NOT m.lDirOilPurch
                           lnOTax1 = lnOTax1 - m.noiltax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax1 = lnOTax1 - m.noiltax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax1p = lnOTax1p - m.noiltax1
                     ENDIF
                  ENDIF

                  IF m.ngastax1 # 0
                     IF NOT m.lsev1g
                        IF NOT m.lDirGasPurch
                           lnGTax1 = lnGTax1 - m.ngastax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax1 = lnGTax1 - m.ngastax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax1p = lnGTax1p - m.ngastax1
                     ENDIF
                  ENDIF

                  IF m.nOthTax1 # 0
                     IF NOT m.lsev1p
                        lnPTax1 = lnPTax1 - m.nOthTax1
                     ELSE
                        lnPTax1p = lnPTax1p - m.nOthTax1
                     ENDIF
                  ENDIF

                  IF m.noiltax2 # 0
                     IF NOT m.lsev2o
                        IF NOT m.lDirOilPurch
                           lnOTax2 = lnOTax2 - m.noiltax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax2 = lnOTax2 - m.noiltax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax2p = lnOTax2p - m.noiltax2
                     ENDIF
                  ENDIF

                  IF m.ngastax2 # 0
                     IF NOT m.lsev2g
                        IF NOT m.lDirGasPurch
                           lnGTax2 = lnGTax2 - m.ngastax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax2 = lnGTax2 - m.ngastax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax2p = lnGTax2p - m.ngastax2
                     ENDIF
                  ENDIF

                  IF m.nOthTax2 # 0
                     IF NOT m.lsev2p
                        lnPTax2 = lnPTax2 - m.nOthTax2
                     ELSE
                        lnPTax2 = lnPTax2 - m.nOthTax2
                     ENDIF
                  ENDIF

                  IF m.noiltax3 # 0
                     IF NOT m.lsev3o
                        IF NOT m.lDirOilPurch
                           lnOTax3 = lnOTax3 - m.noiltax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax3 = lnOTax3 - m.noiltax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax3p = lnOTax3p - m.noiltax3
                     ENDIF
                  ENDIF

                  IF m.ngastax3 # 0
                     IF NOT m.lsev3g
                        IF NOT m.lDirGasPurch
                           lnGTax3 = lnGTax3 - m.ngastax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax3 = lnGTax3 - m.ngastax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax3p = lnGTax3p - m.ngastax3
                     ENDIF
                  ENDIF

                  IF m.nOthTax3 # 0
                     IF NOT m.lsev3p
                        lnPTax3 = lnPTax3 - m.nOthTax3
                     ELSE
                        lnPTax3p = lnPTax3p - m.nOthTax3
                     ENDIF
                  ENDIF

                  IF m.noiltax4 # 0
                     IF NOT m.lsev4o
                        IF NOT m.lDirOilPurch
                           lnOTax4 = lnOTax4 - m.noiltax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax4 = lnOTax4 - m.noiltax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax4p = lnOTax4p - m.noiltax4
                     ENDIF
                  ENDIF

                  IF m.ngastax4 # 0
                     IF NOT m.lsev4g
                        IF NOT m.lDirGasPurch
                           lnGTax4 = lnGTax4 - m.ngastax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax4 = lnGTax4 - m.ngastax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax4p = lnGTax4p - m.ngastax4
                     ENDIF
                  ENDIF

                  IF m.nOthTax4 # 0
                     IF NOT m.lsev4p
                        lnPTax4 = lnPTax4 - m.nOthTax4
                     ELSE
                        lnPTax4p = lnPTax4p - m.nOthTax4
                     ENDIF
                  ENDIF

*  Post compression and gathering
                  IF m.nCompress # 0
                     lnCompress = lnCompress - m.nCompress
                  ENDIF

                  IF m.nGather # 0
                     lnGathering = lnGathering - m.nGather
                  ENDIF

*  Post marketing expenses
                  IF m.nMKTGExp # 0
                     lnMarketing = lnMarketing - m.nMKTGExp
                  ENDIF

*  Post the Expenses
                  lnExpenses = lnExpenses + m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + m.ntotalea + m.ntotaleb

*  Post Backup Withholding
                  IF m.nbackwith # 0
                     lnBackWith = lnBackWith - m.nbackwith
                  ENDIF

*  Post Tax Withholding
                  IF m.ntaxwith # 0
                     lnTaxWith = lnTaxWith - m.ntaxwith
                  ENDIF

*  Post net
                  DO CASE
                     CASE m.csusptype = 'D'
                        lnDeficit = lnDeficit + m.nnetcheck
                     CASE m.csusptype = 'I'
                        lnIntHold = lnIntHold - m.nnetcheck
                     CASE m.csusptype = 'M'
                        lnMinimum = lnMinimum - m.nnetcheck
                     CASE m.csusptype = 'H'
                        lnHolds  = lnHolds - m.nnetcheck
                     CASE INLIST(m.csusptype, 'Q', 'S', 'A')
                        lnFreqs = lnFreqs - m.nnetcheck
                  ENDCASE
               ENDSCAN
            ENDSCAN


            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Posting Owner Suspense - Finishing')
            ENDIF

            THIS.ogl.cidchec    = ''
            THIS.ogl.cReference = 'Run: R' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + '/' + THIS.cgroup
            THIS.ogl.cID        = ''
            THIS.ogl.dGLDate    = tdPostDate
            THIS.ogl.cBatch     = GetNextPK('BATCH')
            THIS.ogl.cUnitNo    = ''
            THIS.ogl.cdeptno    = ''

* Post to Revenue Clearing
            IF lnRevenue # 0
               THIS.ogl.cDesc   = 'Revenue'
               THIS.ogl.cAcctNo = lcRevClear
               THIS.ogl.namount = lnRevenue
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Expense Clearing
            IF lnExpenses # 0
               THIS.ogl.cDesc   = 'Expenses'
               THIS.ogl.cAcctNo = lcExpClear
               THIS.ogl.namount = lnExpenses * -1
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Taxes
            IF lnOTax1 # 0 OR lnOTax1p # 0
               THIS.ogl.cDesc     = 'Oil Tax 1'
               IF lnOTax1p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnOTax1p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnOTax1 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct1
                  THIS.ogl.namount = lnOTax1
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnOTax2 # 0 OR lnOTax2p # 0
               THIS.ogl.cDesc     = 'Oil Tax 2'
               IF lnOTax2p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnOTax2p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnOTax2 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct2
                  THIS.ogl.namount = lnOTax2
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnOTax3 # 0 OR lnOTax3p # 0
               THIS.ogl.cDesc     = 'Oil Tax 3'
               IF lnOTax3p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnOTax3p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnOTax3 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct3
                  THIS.ogl.namount = lnOTax3
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnOTax4 # 0 OR lnOTax4p # 0
               THIS.ogl.cDesc     = 'Oil Tax 4'
               IF lnOTax4p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnOTax4p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnOTax4 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct4
                  THIS.ogl.namount = lnOTax4
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnGTax1 # 0 OR lnGTax1p # 0
               THIS.ogl.cDesc     = 'Gas Tax 1'
               IF lnGTax1p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnGTax1p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnGTax1 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct1
                  THIS.ogl.namount = lnGTax1
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnGTax2 # 0 OR lnGTax2p # 0
               THIS.ogl.cDesc     = 'Gas Tax 2'
               IF lnGTax2p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnGTax2p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnGTax2 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct2
                  THIS.ogl.namount = lnGTax2
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnGTax3 # 0 OR lnGTax3p # 0
               THIS.ogl.cDesc     = 'Gas Tax 3'
               IF lnGTax3p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnGTax3p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnGTax3 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct3
                  THIS.ogl.namount = lnGTax3
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnGTax4 # 0 OR lnGTax4p # 0
               THIS.ogl.cDesc     = 'Gas Tax 4'
               IF lnGTax4p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnGTax4p
                  THIS.ogl.UpdateBatch()
               ENDIF
               IF lnGTax4 # 0
                  THIS.ogl.cAcctNo = m.cTaxAcct4
                  THIS.ogl.namount = lnGTax4
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF lnPTax1 # 0 OR lnPTax1p # 0
               THIS.ogl.cDesc     = 'Other Tax 1'
               IF lnPTax1p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnPTax1p
               ELSE
                  THIS.ogl.cAcctNo = m.cTaxAcct1
                  THIS.ogl.namount = lnPTax1
               ENDIF
               THIS.ogl.UpdateBatch()
            ENDIF

            IF lnPTax2 # 0 OR lnPTax2p # 0
               THIS.ogl.cDesc     = 'Other Tax 2'
               IF lnPTax2p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnPTax2p
               ELSE
                  THIS.ogl.cAcctNo = m.cTaxAcct2
                  THIS.ogl.namount = lnPTax2
               ENDIF
               THIS.ogl.UpdateBatch()
            ENDIF

            IF lnPTax3 # 0 OR lnPTax3p # 0
               THIS.ogl.cDesc     = 'Other Tax 3'
               IF lnPTax3p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnPTax3p
               ELSE
                  THIS.ogl.cAcctNo = m.cTaxAcct3
                  THIS.ogl.namount = lnPTax3
               ENDIF
               THIS.ogl.UpdateBatch()
            ENDIF

            IF lnPTax4 # 0 OR lnPTax4p # 0
               THIS.ogl.cDesc     = 'Other Tax 4'
               IF lnPTax4p # 0
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.namount = lnPTax4p
               ELSE
                  THIS.ogl.cAcctNo = m.cTaxAcct4
                  THIS.ogl.namount = lnPTax4
               ENDIF
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Compression & Gathering
            IF lnCompress # 0
               THIS.ogl.cDesc   = 'Compression'
               THIS.ogl.cAcctNo = lcExpClear
               THIS.ogl.namount = lnCompress
               THIS.ogl.UpdateBatch()
            ENDIF

            IF lnGathering # 0
               THIS.ogl.cDesc   = 'Gathering'
               THIS.ogl.cAcctNo = lcExpClear
               THIS.ogl.namount = lnGathering
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Marketing
            IF lnMarketing # 0
               THIS.ogl.cDesc   = 'Marketing'
               THIS.ogl.cAcctNo = lcExpClear
               THIS.ogl.namount = lnCompress
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Backup Withholding
            IF lnBackWith # 0
               THIS.ogl.cDesc   = 'Backup Withholding'
               THIS.ogl.cAcctNo = m.cBackWith
               THIS.ogl.namount = lnBackWith
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Tax Withholding
            IF lnTaxWith # 0
               THIS.ogl.cDesc   = 'Tax Withholding'
               THIS.ogl.cAcctNo = m.cBackWith
               THIS.ogl.namount = lnTaxWith
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Interest On Hold
            IF lnIntHold # 0
               THIS.ogl.cDesc   = 'Interest On Hold'
               THIS.ogl.cAcctNo = m.cMinAcct
               THIS.ogl.namount = lnIntHold
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Minimums
            IF lnMinimum # 0
               THIS.ogl.cDesc   = 'Minimum Checks'
               THIS.ogl.cAcctNo = m.cMinAcct
               THIS.ogl.namount = lnMinimum
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Deficits
            IF lnDeficit # 0
               THIS.ogl.cDesc   = 'Deficits'
               THIS.ogl.cAcctNo = m.cDefAcct
               THIS.ogl.namount = lnDeficit * -1
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Owner Holds
            IF lnHolds # 0
               THIS.ogl.cDesc   = 'Owner Holds'
               THIS.ogl.cAcctNo = m.cMinAcct
               THIS.ogl.namount = lnHolds
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Owner Frequency Holds
            IF lnFreqs # 0
               THIS.ogl.cDesc   = 'Owner Freq Holds'
               THIS.ogl.cAcctNo = m.cMinAcct
               THIS.ogl.namount = lnFreqs
               THIS.ogl.UpdateBatch()
            ENDIF

* Post Checks
            IF lnCheck # 0
               THIS.ogl.cDesc   = 'Check Amounts'
               THIS.ogl.cAcctNo = m.cDisbAcct
               THIS.ogl.namount = lnCheck * -1
               THIS.ogl.UpdateBatch()
            ENDIF

            llReturnx = THIS.ogl.ChkBalance()

         ENDIF

*  Mark the expense entries as being tied to this DM batch
         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Marking Expenses For This Run')
         ENDIF
         swselect('expense')
         SCAN FOR nRunNoRev = THIS.nrunno AND cRunYearRev = THIS.crunyear ;
               AND expense.cBatch = ' '
            m.cWellID = cWellID
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               IF cgroup = tcGroup
                  swselect('expense')
                  REPL cBatch WITH THIS.cdmbatch
               ENDIF
            ENDIF
         ENDSCAN

*   Post the Vendor amounts that are designated to be posted.
         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Vendor Checks to General Ledger...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnVendor = 0

         swselect('wells')
         SET ORDER TO cWellID

*  Get the vendors to be posted.
         SELECT cvendorid, cVendName FROM vendor WHERE lIntegGL = .T. INTO CURSOR curVends
         IF NOT llNoPostDM AND _TALLY > 0
            THIS.ogl.dGLDate = tdCompanyPost
            THIS.ogl.cBatch  = GetNextPK('BATCH')
            SELECT curVends
            SCAN
               SCATTER MEMVAR

               lnAmount     = 0
               THIS.ogl.cID = m.cvendorid
               swselect('expense')
               lnCount = 1

* Summarize by ccateg to cut down on journal entries
               SELECT  expense.cWellID, ;
                       ccateg, ;
                       cexpclass, ;
                       SUM(namount) AS namount, ;
                       cvendorid, ;
                       ccatcode, ;
                       expense.cownerid ;
                   FROM expense WITH (BUFFERING = .T.) ;
                   WHERE cvendorid = m.cvendorid ;
                       AND nRunNoRev = THIS.nrunno ;
                       AND cRunYearRev = THIS.crunyear ;
                       AND expense.cWellID IN (SELECT  cWellID ;
                                                   FROM wellwork) ;
                   INTO CURSOR tempexp ;
                   ORDER BY cvendorid,;
                       cWellID,;
                       expense.cownerid,;
                       ccatcode ;
                   GROUP BY cvendorid,;
                       cWellID,;
                       expense.cownerid,;
                       ccatcode
               SELECT tempexp
               COUNT FOR cvendorid = m.cvendorid TO lnMax

               SELECT tempexp
               SCAN FOR namount # 0
                  SCATTER MEMVAR
                  m.cUnitNo = cWellID
                  m.cID     = cvendorid

*  Check to make sure the well is in the right group
                  swselect('wells')
                  IF SEEK(m.cWellID)
                     IF wells.cgroup # tcGroup
                        LOOP
                     ENDIF
                  ENDIF

*  Get the account numbers to be posted for this expense category
                  swselect('expcat')
                  SET ORDER TO ccatcode
                  IF SEEK(m.ccatcode)
                     SCATTER MEMVAR
                     m.ccateg   = ccateg
                     m.cDRAcctV = THIS.cexpclear
                     IF EMPTY(m.cCRAcctV)
                        m.cCRAcctV = lcSuspense
                     ENDIF
                  ELSE
                     m.cCRAcctV = lcSuspense
                  ENDIF

*  Net out any JIB interest shares from the expense
                  m.namount   = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, m.ccatcode, m.cdeck)

*  Add amount of this invoice to the total the vendor is to be paid
                  lnAmount = lnAmount + m.namount

                  THIS.ogl.cReference = 'Vendor Amts'
                  THIS.ogl.cUnitNo    = m.cUnitNo
                  THIS.ogl.cDesc      = m.ccateg
                  THIS.ogl.cdeptno    = lcDeptNo
                  THIS.ogl.cAcctNo    = m.cCRAcctV
                  THIS.ogl.namount    = m.namount * -1
                  THIS.ogl.UpdateBatch()

                  THIS.ogl.cAcctNo = lcDMExp
                  THIS.ogl.namount = m.namount
                  THIS.ogl.cDesc   = m.ccateg
                  THIS.ogl.UpdateBatch()
               ENDSCAN && tempexp

               llReturn = THIS.ogl.ChkBalance()

               IF NOT llReturn
                  IF NOT FILE('outbal.dbf')
                     CREATE TABLE outbal FREE (cBatch  c(8), cownerid  c(10))
                  ENDIF
                  IF NOT USED('outbal')
                     USE outbal IN 0
                  ENDIF
                  m.cBatch   = THIS.ogl.cBatch
                  m.cownerid = m.cID
                  INSERT INTO outbal FROM MEMVAR
               ENDIF

            ENDSCAN && curVends
         ENDIF

*  Post the owners that are designated to be posted
         swselect('wells')
         SET ORDER TO cWellID

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Operator Owner Amounts to General Ledger (Summary)...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnOwner = 0

*  Get the owners to be posted.
         SELECT cownerid, cownname FROM investor WHERE lIntegGL = .T. INTO CURSOR curPostOwn

         IF NOT llNoPostDM AND _TALLY > 0
            lnAmount         = 0
            THIS.ogl.dGLDate = tdCompanyPost
            SELECT curPostOwn
            SCAN
               SCATTER MEMVAR

               STORE 0 TO lnWIOilRev, lnWIGasRev, lnWIOthRev, lnExpenses, lnOilTax, lnGasTax
               STORE 0 TO lnRYOilRev, lnRYGasRev, lnRYOthRev, lnOVOilRev, lnOVGasRev, lnOVOthRev
               STORE 0 TO lnRYMisc1Rev, lnRYMisc2Rev, lnWIMisc1Rev, lnWIMisc2Rev
               STORE 0 TO lnOVMIsc1Rev, lnOVMisc2Rev
               STORE 0 TO lnWIOTax1, lnWIOTax2, lnWIOTax3, lnWIOTax4, lnCheck, lnMinimum, lnDeficit
               STORE 0 TO lnWIGTax1, lnWIGTax2, lnWIGTax3, lnWIGTax4
               STORE 0 TO lnWIPTax1, lnWIPTax2, lnWIPTax3, lnWIPTax4
               STORE 0 TO lnRYOTax1, lnRYOTax2, lnRYOTax3, lnRYOTax4
               STORE 0 TO lnRYGTax1, lnRYGTax2, lnRYGTax3, lnRYGTax4
               STORE 0 TO lnRYPTax1, lnRYPTax2, lnRYPTax3, lnRYPTax4
               STORE 0 TO lnOVOTax1, lnOVOTax2, lnOVOTax3, lnOVOTax4
               STORE 0 TO lnOVGTax1, lnOVGTax2, lnOVGTax3, lnOVGTax4
               STORE 0 TO lnOVPTax1, lnOVPTax2, lnOVPTax3, lnOVPTax4
               STORE 0 TO lnWIOTax1p, lnWIOTax2p, lnWIOTax3p, lnWIOTax4p
               STORE 0 TO lnWIGTax1p, lnWIGTax2p, lnWIGTax3p, lnWIGTax4p
               STORE 0 TO lnWIPTax1p, lnWIPTax2p, lnWIPTax3p, lnWIPTax4p
               STORE 0 TO lnRYOTax1p, lnRYOTax2p, lnRYOTax3p, lnRYOTax4p
               STORE 0 TO lnRYGTax1p, lnRYGTax2p, lnRYGTax3p, lnRYGTax4p
               STORE 0 TO lnRYPTax1p, lnRYPTax2p, lnRYPTax3p, lnRYPTax4p
               STORE 0 TO lnOVOTax1p, lnOVOTax2p, lnOVOTax3p, lnOVOTax4p
               STORE 0 TO lnOVGTax1p, lnOVGTax2p, lnOVGTax3p, lnOVGTax4p
               STORE 0 TO lnOVPTax1p, lnOVPTax2p, lnOVPTax3p, lnOVPTax4p
               STORE 0 TO lnWICompress, lnWIGathering, lnMarketing, lnBackWith, lnTaxWith, lnIntHold, lnHold, lnCheck
               STORE 0 TO lnRYCompress, lnRYGathering, lnRYFlatRate
               STORE 0 TO lnOVCompress, lnOVGathering, lnOVFlatRate
               STORE 0 TO lnWITrpRev, lnWIMi1Rev, lnWIMi2Rev, lnDebits, lnCredits, lnMin, lnDeficit
               STORE 0 TO lnRYTrpRev, lnRYMi1Rev, lnRYMi2Rev
               STORE 0 TO lnOVTrpRev, lnOVMi1Rev, lnOVMi2Rev

               THIS.ogl.cBatch = GetNextPK('BATCH')

               SELECT invtmp
               COUNT FOR cownerid = m.cownerid AND nrunno = THIS.nrunno AND cgroup = tcGroup TO lnMax
               lnCount = 1

               SELECT invtmp
               SET ORDER TO invprog
               SCAN FOR cownerid = m.cownerid AND nrunno = THIS.nrunno AND cgroup = tcGroup AND crectype = 'R'
                  SCATTER MEMVAR

                  llJIB  = lJIB
                  lcName = m.cownname

                  swselect('wells')
                  IF SEEK(m.cWellID)
                     SCATTER FIELDS LIKE lSev* MEMVAR
                     m.nprocess     = nprocess
                     m.lDirOilPurch = lDirOilPurch
                     m.lDirGasPurch = lDirGasPurch
                  ELSE
                     m.nprocess = 1
                     STORE .F. TO m.lDirOilPurch, m.lDirGasPurch
                  ENDIF

*  Post Oil Income
                  IF m.noilrev # 0 AND NOT INLIST(m.cdirect, 'O', 'B')
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOilRev = lnWIOilRev - m.noilrev
                        CASE m.ctypeinv = 'L'
                           lnRYOilRev = lnRYOilRev - m.noilrev
                        CASE m.ctypeinv = 'O'
                           lnOVOilRev = lnOVOilRev - m.noilrev
                        OTHERWISE
                           lnWIOilRev = lnWIOilRev - m.noilrev
                     ENDCASE
                  ENDIF
*  Post Gas Income
                  IF m.ngasrev # 0 AND NOT INLIST(m.cdirect, 'G', 'B')
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIGasRev = lnWIGasRev - m.ngasrev
                        CASE m.ctypeinv = 'L'
                           lnRYGasRev = lnRYGasRev - m.ngasrev
                        CASE m.ctypeinv = 'O'
                           lnOVGasRev = lnOVGasRev - m.ngasrev
                        OTHERWISE
                           lnWIGasRev = lnWIGasRev - m.ngasrev
                     ENDCASE
                  ENDIF
*  Post Other Income
                  IF m.nothrev # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOthRev = lnWIOthRev - m.nothrev
                        CASE m.ctypeinv = 'L'
                           lnRYOthRev = lnRYOthRev - m.nothrev
                        CASE m.ctypeinv = 'O'
                           lnOVOthRev = lnOVOthRev - m.nothrev
                        OTHERWISE
                           lnWIOthRev = lnWIOthRev - m.nothrev
                     ENDCASE
                  ENDIF
*  Post Trp Income
                  IF m.ntrprev # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOthRev = lnWIOthRev - m.ntrprev
                        CASE m.ctypeinv = 'L'
                           lnRYOthRev = lnRYOthRev - m.ntrprev
                        CASE m.ctypeinv = 'O'
                           lnOVOthRev = lnOVOthRev - m.ntrprev
                        OTHERWISE
                           lnWIOthRev = lnWIOthRev - m.ntrprev
                     ENDCASE
                  ENDIF
*  Post Misc 1 Income
                  IF m.nmiscrev1 # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIMisc1Rev = lnWIMisc1Rev - m.nmiscrev1
                        CASE m.ctypeinv = 'L'
                           lnRYMisc1Rev = lnRYMisc1Rev - m.nmiscrev1
                        CASE m.ctypeinv = 'O'
                           lnOVMIsc1Rev = lnOVMIsc1Rev - m.nmiscrev1
                        OTHERWISE
                           lnWIMisc1Rev = lnWIMisc1Rev - m.nmiscrev1
                     ENDCASE
                  ENDIF
*  Post Misc 2 Income
                  IF m.nmiscrev2 # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIMisc2Rev = lnWIMisc2Rev - m.nmiscrev2
                        CASE m.ctypeinv = 'L'
                           lnRYMisc2Rev = lnRYMisc2Rev - m.nmiscrev2
                        CASE m.ctypeinv = 'O'
                           lnOVMisc2Rev = lnOVMisc2Rev - m.nmiscrev2
                        OTHERWISE
                           lnWIMisc2Rev = lnWIMisc2Rev - m.nmiscrev2
                     ENDCASE
                  ENDIF
*  Post Flat Rates
                  IF m.nflatrate # 0
                     DO CASE
                        CASE m.ctypeinv = 'L'
                           lnRYGasRev = lnRYGasRev - m.nflatrate
                        CASE m.ctypeinv = 'O'
                           lnOVGasRev = lnOVGasRev - m.nflatrate
                        OTHERWISE
                           lnRYGasRev = lnRYGasRev - m.nflatrate
                     ENDCASE
                  ENDIF
*  Post Oil Taxes
                  IF m.noiltax1 # 0
                     lnOTax1  = 0
                     lnOTax1p = 0
                     IF NOT m.lsev1o
                        IF NOT m.lDirOilPurch
                           lnOTax1 = lnOTax1 + m.noiltax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax1 = lnOTax1 + m.noiltax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax1p = lnOTax1p + m.noiltax1
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOTax1  = lnWIOTax1 + lnOTax1
                           lnWIOTax1p = lnWIOTax1p + lnOTax1p
                        CASE m.ctypeinv = 'L'
                           lnRYOTax1  = lnRYOTax1 + lnOTax1
                           lnRYOTax1p = lnRYOTax1p + lnOTax1p
                        CASE m.ctypeinv = 'O'
                           lnOVOTax1  = lnOVOTax1 + lnOTax1
                           lnOVOTax1p = lnOVOTax1p + lnOTax1p
                        OTHERWISE
                           lnWIOTax1  = lnWIOTax1 + lnOTax1
                           lnWIOTax1p = lnWIOTax1p + lnOTax1p
                     ENDCASE
                  ENDIF

                  IF m.noiltax2 # 0
                     lnOTax2  = 0
                     lnOTax2p = 0
                     IF NOT m.lsev2o
                        IF NOT m.lDirOilPurch
                           lnOTax2 = lnOTax2 + m.noiltax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax2 = lnOTax2 + m.noiltax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax2p = lnOTax2p + m.noiltax2
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOTax2  = lnWIOTax2 + lnOTax2
                           lnWIOTax2p = lnWIOTax2p + lnOTax2p
                        CASE m.ctypeinv = 'L'
                           lnRYOTax2  = lnRYOTax2 + lnOTax2
                           lnRYOTax2p = lnRYOTax2p + lnOTax2p
                        CASE m.ctypeinv = 'O'
                           lnOVOTax2  = lnOVOTax2 + lnOTax2
                           lnOVOTax2p = lnOVOTax2p + lnOTax2p
                        OTHERWISE
                           lnWIOTax2  = lnWIOTax2 + lnOTax2
                           lnWIOTax2p = lnWIOTax2p + lnOTax2p
                     ENDCASE
                  ENDIF

                  IF m.noiltax3 # 0
                     lnOTax3  = 0
                     lnOTax3p = 0
                     IF NOT m.lsev3o
                        IF NOT m.lDirOilPurch
                           lnOTax3 = lnOTax3 + m.noiltax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax3 = lnOTax3 + m.noiltax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax3p = lnOTax3p + m.noiltax3
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOTax3  = lnWIOTax3 + lnOTax3
                           lnWIOTax3p = lnWIOTax3p + lnOTax3p
                        CASE m.ctypeinv = 'L'
                           lnRYOTax3  = lnRYOTax3 + lnOTax3
                           lnRYOTax3p = lnRYOTax3p + lnOTax3p
                        CASE m.ctypeinv = 'O'
                           lnOVOTax3  = lnOVOTax3 + lnOTax3
                           lnOVOTax4p = lnOVOTax3p + lnOTax3p
                        OTHERWISE
                           lnWIOTax3  = lnWIOTax3 + lnOTax3
                           lnWIOTax3p = lnWIOTax3p + lnOTax3p
                     ENDCASE
                  ENDIF

                  IF m.noiltax4 # 0
                     lnOTax4  = 0
                     lnOTax4p = 0
                     IF NOT m.lsev4o
                        IF NOT m.lDirOilPurch
                           lnOTax4 = lnOTax4 + m.noiltax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax4 = lnOTax4 + m.noiltax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax4p = lnOTax4p + m.noiltax4
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIOTax4  = lnWIOTax4 + lnOTax4
                           lnWIOTax4p = lnWIOTax4p + lnOTax4p
                        CASE m.ctypeinv = 'L'
                           lnRYOTax4  = lnRYOTax4 + lnOTax4
                           lnRYOTax4p = lnRYOTax4p + lnOTax4p
                        CASE m.ctypeinv = 'O'
                           lnOVOTax4  = lnOVOTax4 + lnOTax4
                           lnOVOTax4p = lnOVOTax4p + lnOTax4p
                        OTHERWISE
                           lnWIOTax4  = lnWIOTax4 + lnOTax4
                           lnWIOTax4p = lnWIOTax4p + lnOTax4p
                     ENDCASE
                  ENDIF

*  Post Gas Taxes
                  IF m.ngastax1 # 0
                     lnGTax1  = 0
                     lnGTax1p = 0
                     IF NOT m.lsev1g
                        IF NOT m.lDirOilPurch
                           lnGTax1 = lnGTax1 + m.ngastax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnGTax1 = lnGTax1 + m.ngastax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax1p = lnGTax1p + m.ngastax1
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIGTax1  = lnWIGTax1 + lnGTax1
                           lnWIGTax1p = lnWIGTax1p + lnGTax1p
                        CASE m.ctypeinv = 'L'
                           lnRYGTax1  = lnRYGTax1 + lnGTax1
                           lnRYGTax1p = lnRYGTax1p + lnGTax1p
                        CASE m.ctypeinv = 'O'
                           lnOVGTax1  = lnOVGTax1 + lnGTax1
                           lnOVGTax1p = lnOVGTax1p + lnGTax1p
                        OTHERWISE
                           lnWIOTax1  = lnWIOTax1 + lnGTax1
                           lnWIOTax1p = lnWIOTax1p + lnGTax1p
                     ENDCASE
                  ENDIF

                  IF m.ngastax2 # 0
                     lnGTax2  = 0
                     lnGTax2p = 0
                     IF NOT m.lsev2g
                        IF NOT m.lDirOilPurch
                           lnGTax2 = lnGTax2 + m.ngastax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnGTax2 = lnGTax2 + m.ngastax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax2p = lnGTax2p + m.ngastax2
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIGTax2  = lnWIGTax2 + lnGTax2
                           lnWIGTax2p = lnWIGTax2p + lnGTax2p
                        CASE m.ctypeinv = 'L'
                           lnRYGTax2  = lnRYGTax2 + lnGTax2
                           lnRYGTax2p = lnRYGTax2p + lnGTax2p
                        CASE m.ctypeinv = 'O'
                           lnOVGTax2  = lnOVGTax2 + lnGTax2
                           lnOVGTax2p = lnOVGTax2p + lnGTax2p
                        OTHERWISE
                           lnWIGTax2  = lnWIGTax2 + lnGTax2
                           lnWIGTax2p = lnWIGTax2p + lnGTax2p
                     ENDCASE
                  ENDIF

                  IF m.ngastax3 # 0
                     lnGTax3  = 0
                     lnGTax3p = 0
                     IF NOT m.lsev3g
                        IF NOT m.lDirOilPurch
                           lnGTax3 = lnGTax3 + m.ngastax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnGTax3 = lnGTax3 + m.ngastax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax3p = lnGTax3p + m.ngastax3
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIGTax3  = lnWIGTax3 + lnGTax3
                           lnWIGTax3p = lnWIGTax3p + lnGTax3p
                        CASE m.ctypeinv = 'L'
                           lnRYGTax3  = lnRYGTax3 + lnGTax3
                           lnRYGTax3p = lnRYGTax3p + lnGTax3p
                        CASE m.ctypeinv = 'O'
                           lnOVGTax3  = lnOVGTax3 + lnGTax3
                           lnOVGTax4p = lnOVGTax3p + lnGTax3p
                        OTHERWISE
                           lnWIGTax3  = lnWIGTax3 + lnGTax3
                           lnWIGTax3p = lnWIGTax3p + lnGTax3p
                     ENDCASE
                  ENDIF

                  IF m.ngastax4 # 0
                     lnGTax4  = 0
                     lnGTax4p = 0
                     IF NOT m.lsev4g
                        IF NOT m.lDirOilPurch
                           lnGTax4 = lnGTax4 + m.ngastax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnGTax4 = lnGTax4 + m.ngastax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax4p = lnGTax4p + m.ngastax4
                     ENDIF
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIGTax4  = lnWIGTax4 + lnGTax4
                           lnWIGTax4p = lnWIGTax4p + lnGTax4p
                        CASE m.ctypeinv = 'L'
                           lnRYGTax4  = lnRYGTax4 + lnGTax4
                           lnRYGTax4p = lnRYGTax4p + lnGTax4p
                        CASE m.ctypeinv = 'O'
                           lnOVGTax4  = lnOVGTax4 + lnGTax4
                           lnOVGTax4p = lnOVGTax4p + lnGTax4p
                        OTHERWISE
                           lnWIGTax4  = lnWIGTax4 + lnGTax4
                           lnWIGTax4p = lnWIGTax4p + lnGTax4p
                     ENDCASE
                  ENDIF
*  Post Other Product Taxes
                  IF m.nOthTax1 # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIPTax1 = lnWIPTax1 + m.nOthTax1
                        CASE m.ctypeinv = 'L'
                           lnRYPTax1 = lnRYPTax1 + m.nOthTax1
                        CASE m.ctypeinv = 'O'
                           lnOVPTax1 = lnOVPTax1 + m.nOthTax1
                        OTHERWISE
                           lnWIPTax1 = lnWIPTax1 + m.nOthTax1
                     ENDCASE
                  ENDIF

                  IF m.nOthTax2 # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIPTax2 = lnWIPTax2 + m.nOthTax2
                        CASE m.ctypeinv = 'L'
                           lnRYPTax2 = lnRYPTax2 + m.nOthTax2
                        CASE m.ctypeinv = 'O'
                           lnOVPTax2 = lnOVPTax2 + m.nOthTax2
                        OTHERWISE
                           lnWIPTax2 = lnWIPTax2 + m.nOthTax2
                     ENDCASE
                  ENDIF

                  IF m.nOthTax3 # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIPTax3 = lnWIPTax3 + m.nOthTax3
                        CASE m.ctypeinv = 'L'
                           lnRYPTax3 = lnRYPTax3 + m.nOthTax3
                        CASE m.ctypeinv = 'O'
                           lnOVPTax3 = lnOVPTax3 + m.nOthTax3
                        OTHERWISE
                           lnWIPTax3 = lnWIPTax3 + m.nOthTax3
                     ENDCASE
                  ENDIF

                  IF m.nOthTax4 # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIPTax4 = lnWIPTax4 + m.nOthTax4
                        CASE m.ctypeinv = 'L'
                           lnRYPTax4 = lnRYPTax4 + m.nOthTax4
                        CASE m.ctypeinv = 'O'
                           lnOVPTax4 = lnOVPTax4 + m.nOthTax4
                        OTHERWISE
                           lnWIPTax4 = lnWIPTax4 + m.nOthTax4
                     ENDCASE
                  ENDIF

*  Post compression and gathering
                  IF m.nCompress # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWICompress = lnWICompress + m.nCompress
                        CASE m.ctypeinv = 'L'
                           lnRYCompress = lnRYCompress + m.nCompress
                        CASE m.ctypeinv = 'O'
                           lnOVCompress = lnOVCompress + m.nCompress
                        OTHERWISE
                           lnWICompress = lnWICompress + m.nCompress
                     ENDCASE
                  ENDIF

                  IF m.nGather # 0
                     DO CASE
                        CASE m.ctypeinv = 'W'
                           lnWIGathering = lnWIGathering + m.nGather
                        CASE m.ctypeinv = 'L'
                           lnRYGathering = lnRYGathering + m.nGather
                        CASE m.ctypeinv = 'O'
                           lnOVGathering = lnOVGathering + m.nGather
                        OTHERWISE
                           lnWIGathering = lnWIGathering + m.nGather
                     ENDCASE
                  ENDIF

*  Post marketing expenses
                  IF m.nMKTGExp # 0
                     lnMarketing = lnMarketing + m.nMKTGExp
                  ENDIF

*  Process default class expenses
                  SELE roundtmp
                  LOCATE FOR cownerid == m.cownerid AND cdmbatch = THIS.cdmbatch
                  IF FOUND()
                     llRound = .T.
                  ELSE
                     llRound = .F.
                  ENDIF

                  llRoundIt = .F.
                  IF llSepClose AND llJIB
*  Do Nothing
                  ELSE

                     lnExpenses = lnExpenses + m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + m.ntotalea + m.ntotaleb

                  ENDIF

*  Post Prior Period Deficits
                  IF m.ctypeinv = 'X' AND m.nnetcheck # 0
                     lnDeficit = lnDeficit + m.nnetcheck
                  ENDIF

*  Post Prior Period Minimums
                  IF m.ctypeinv = 'M' AND m.nnetcheck # 0
                     lnMinimum = lnMinimum - m.nnetcheck
                  ENDIF

                  lnCheck = lnCheck + m.nnetcheck

               ENDSCAN  && Invtmp

               THIS.ogl.cUnitNo    = ''
               THIS.ogl.cdeptno    = lcDeptNo
               THIS.ogl.cReference = 'Owner Post'

               IF lnWIOilRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('BBL')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Working Oil Revenue'

                  THIS.ogl.namount = lnWIOilRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = lnWIOilRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYOilRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('BBL')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Royalty Oil Revenue'

                  THIS.ogl.namount = lnRYOilRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYOilRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnOVOilRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('BBL')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Oil Revenue'

                  THIS.ogl.namount = lnOVOilRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVOilRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnWIGasRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MCF')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource = 'Working Gas Revenue'

                  THIS.ogl.namount = lnWIGasRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnWIGasRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYGasRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MCF')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource = 'Royalty Gas Revenue'

                  THIS.ogl.namount = lnRYGasRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYGasRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnOVGasRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MCF')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource = 'Working Gas Revenue'

                  THIS.ogl.namount = lnOVGasRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVGasRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnWIMisc1Rev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MISC1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Working Misc1 Revenue'
                  THIS.ogl.namount = lnWIMisc1Rev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnWIMisc1Rev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnWIMisc2Rev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MISC2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Working Misc2 Revenue'
                  THIS.ogl.namount = lnWIMisc2Rev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnWIMisc2Rev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYMisc1Rev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MISC1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Royalty Misc1 Revenue'
                  THIS.ogl.namount = lnRYMisc1Rev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYMisc1Rev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYMisc2Rev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MISC2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Royalty Misc2 Revenue'
                  THIS.ogl.namount = lnRYMisc2Rev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYMisc2Rev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnOVMIsc1Rev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MISC1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'ORRI Misc1 Revenue'
                  THIS.ogl.namount = lnOVMIsc1Rev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVMIsc1Rev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()

               ENDIF

               IF lnOVMisc2Rev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MISC2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'ORRI Misc2 Revenue'
                  THIS.ogl.namount = lnOVMisc2Rev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVMisc2Rev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()

               ENDIF

               IF lnWIOthRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Working Other Revenue'
                  THIS.ogl.namount = lnWIOthRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnWIOthRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYOthRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Royalty Other Revenue'
                  THIS.ogl.namount = lnRYOthRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYOthRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnOVOthRev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'ORRI Other Revenue'
                  THIS.ogl.namount = lnOVOthRev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVOthRev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()

               ENDIF


               IF lnWIOTax1 # 0 OR lnWIOTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Oil Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIOTax1 + lnWIOTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnWIOTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIOTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIOTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIOTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYOTax1 # 0 OR lnRYOTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Oil Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYOTax1 + lnRYOTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnRYOTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYOTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYOTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYOTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVOTax1 # 0 OR lnOVOTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Oil Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVOTax1 + lnOVOTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnOVOTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVOTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVOTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVOTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIOTax2 # 0 OR lnWIOTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Oil Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIOTax2 + lnWIOTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnWIOTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIOTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIOTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIOTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYOTax2 # 0 OR lnRYOTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Oil Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYOTax2 + lnRYOTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnRYOTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYOTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYOTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYOTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVOTax2 # 0 OR lnOVOTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Oil Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVOTax2 + lnOVOTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnOVOTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVOTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVOTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVOTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIOTax3 # 0 OR lnWIOTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Oil Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIOTax3 + lnWIOTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnWIOTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIOTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIOTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIOTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYOTax3 # 0 OR lnRYOTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Oil Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYOTax3 + lnRYOTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnRYOTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYOTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYOTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYOTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVOTax3 # 0 OR lnOVOTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Oil Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVOTax3 + lnOVOTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnOVOTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVOTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVOTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVOTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIOTax4 # 0 OR lnWIOTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Oil Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIOTax4 + lnWIOTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnWIOTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIOTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIOTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIOTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYOTax4 # 0 OR lnRYOTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Oil Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYOTax4 + lnRYOTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnRYOTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYOTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYOTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYOTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVOTax4 # 0 OR lnOVOTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Oil Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVOTax4 + lnOVOTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnOVOTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVOTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVOTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVOTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF


               IF lnWIGTax1 # 0 OR lnWIGTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Gas Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIGTax1 + lnWIGTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnWIGTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIGTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIGTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIGTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYGTax1 # 0 OR lnRYGTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Gas Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYGTax1 + lnRYGTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnRYGTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYGTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYGTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYGTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVGTax1 # 0 OR lnOVGTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Gas Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVGTax1 + lnOVGTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnOVGTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVGTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVGTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVGTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIGTax2 # 0 OR lnWIGTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Gas Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIGTax2 + lnWIGTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnWIGTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIGTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIGTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIGTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYGTax2 # 0 OR lnRYGTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Gas Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYGTax2 + lnRYGTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnRYGTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYGTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYGTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYGTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVGTax2 # 0 OR lnOVGTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Gas Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVGTax2 + lnOVGTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnOVGTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVGTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVGTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVGTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIGTax3 # 0 OR lnWIGTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Gas Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIGTax3 + lnWIGTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnWIGTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIGTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIGTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIGTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYGTax3 # 0 OR lnRYGTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Gas Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYGTax3 + lnRYGTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnRYGTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYGTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYGTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYGTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVGTax3 # 0 OR lnOVGTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Gas Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVGTax3 + lnOVGTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnOVGTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVGTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVGTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVGTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIGTax4 # 0 OR lnWIGTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Gas Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIGTax4 + lnWIGTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnWIGTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIGTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIGTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIGTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYGTax4 # 0 OR lnRYGTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Gas Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYGTax4 + lnRYGTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnRYGTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYGTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYGTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYGTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVGTax4 # 0 OR lnOVGTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Gas Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVGTax4 + lnOVGTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnOVGTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVGTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcc4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVGTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVGTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIPTax1 # 0 OR lnWIPTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Other Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIPTax1 + lnWIPTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnWIPTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIPTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIPTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIPTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYPTax1 # 0 OR lnRYPTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Other Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYPTax1 + lnRYPTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnRYPTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYPTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYPTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYPTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVPTax1 # 0 OR lnOVPTax1p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Other Tax 1'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVPTax1 + lnOVPTax1p
                  THIS.ogl.UpdateBatch()

                  IF lnOVPTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVPTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVPTax1p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVPTax1p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIPTax2 # 0 OR lnWIPTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Other Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIPTax2 + lnWIPTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnWIPTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIPTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIPTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIPTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYPTax2 # 0 OR lnRYPTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Other Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYPTax2 + lnRYPTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnRYPTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYPTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYPTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYPTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVPTax2 # 0 OR lnOVPTax2p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Other Tax 2'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVPTax2 + lnOVPTax2p
                  THIS.ogl.UpdateBatch()

                  IF lnOVPTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVPTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVPTax2p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVPTax2p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIPTax3 # 0 OR lnWIPTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Other Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIPTax3 + lnWIPTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnWIPTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIPTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIPTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIPTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYPTax3 # 0 OR lnRYPTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Other Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYPTax3 + lnRYPTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnRYPTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYPTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYPTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYPTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVPTax3 # 0 OR lnOVPTax3p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Other Tax 3'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVPTax3 + lnOVPTax3p
                  THIS.ogl.UpdateBatch()

                  IF lnOVPTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVPTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVPTax3p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVPTax3p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWIPTax4 # 0 OR lnWIPTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'WI Other Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnWIPTax4 + lnWIPTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnWIPTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnWIPTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnWIPTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnWIPTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnRYPTax4 # 0 OR lnRYPTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'RI Other Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnRYPTax4 + lnRYPTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnRYPTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnRYPTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnRYPTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnRYPTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnOVPTax4 # 0 OR lnOVPTax4p # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'ORRI Other Tax 4'

                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.namount = lnOVPTax4 + lnOVPTax4p
                  THIS.ogl.UpdateBatch()

                  IF lnOVPTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = lnOVPTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF lnOVPTax4p # 0
* Post to revenue clearing
                     THIS.ogl.namount = lnOVPTax4p * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF lnWICompress # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('COMP')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF

                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('COMP')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Compression'
                     ENDIF
                  ENDIF

                  m.cRevSource = 'Compression'

                  THIS.ogl.namount = lnWICompress
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnWICompress * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnWIGathering # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GATH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('GATH')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Gathering'
                     ENDIF
                  ENDIF


                  m.cRevSource = 'Gathering'

                  THIS.ogl.namount = lnWIGathering
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnWIGathering * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYCompress # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('COMP')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('COMP')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Compression'
                     ENDIF
                  ENDIF


                  m.cRevSource = 'Compression'

                  THIS.ogl.namount = lnRYCompress
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYCompress * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnRYGathering # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GATH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnol
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('GATH')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Gathering'
                     ENDIF
                  ENDIF


                  m.cRevSource = 'Gathering'

                  THIS.ogl.namount = lnRYGathering
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnRYGathering * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnOVCompress # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('COMP')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('COMP')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Compression'
                     ENDIF
                  ENDIF


                  m.cRevSource = 'Compression'

                  THIS.ogl.namount = lnOVCompress
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVCompress * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnOVGathering # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GATH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnoo
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('GATH')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Compression'
                     ENDIF
                  ENDIF


                  m.cRevSource = 'Gathering'

                  THIS.ogl.namount = lnOVGathering
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = lnOVGathering * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnMarketing # 0
                  STORE '' TO lcAcctD, lcAcctC
                  swselect('expcat')
                  SET ORDER TO ccatcode
                  IF SEEK('MKTG')
                     lcAcctD = cdraccto
                     lcDesc  = ccateg
                     IF EMPTY(lcAcctD)
                        lcAcctD = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctD = lcSuspense
                     lcDesc  = 'Unknown'
                  ENDIF
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcAcctD
                  THIS.ogl.namount = lnMarketing
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.namount = lnMarketing * -1
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF lnExpenses # 0
                  STORE '' TO lcAcctD, lcAcctC
                  swselect('expcat')
                  SET ORDER TO ccatcode
                  IF SEEK('EXPS')
                     lcAcctD = cdraccto
                     lcDesc  = ccateg
                     IF EMPTY(lcAcctD)
                        lcAcctD = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctD = lcSuspense
                     lcDesc  = 'Unknown - EXPS'
                  ENDIF

                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcAcctD
                  THIS.ogl.namount = lnExpenses
                  THIS.ogl.cID     = m.cownerid
                  THIS.ogl.UpdateBatch()
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.namount = lnExpenses * -1
                  THIS.ogl.UpdateBatch()
               ENDIF

               llReturn = THIS.ogl.ChkBalance()

            ENDSCAN && curPostOwn
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'PostSummary', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('PostSummary', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn

   ENDPROC

*********************************
   PROCEDURE PostSummaryWell
*********************************
      LOCAL tcYear, tcPeriod, tdCheckDate, tcGroup, tdPostDate
      LOCAL lnMax, lnCount, lnTotal, lcName, lnJIBInv, m.cCustName, lcDMBatch, llSepClose
      LOCAL lcRevClear, lcSuspense, m.cDisbAcct, m.cVendComp, m.cGathAcct, m.cBackWith
      LOCAL llIntegComp, llSepClose, lcAPAcct, lcAcctYear, lcAcctMonth, lcIDChec, llExpSum
      LOCAL llRound, llRoundIt, lnOwner
      LOCAL lDirGasPurch, lDirOilPurch, lcAcctC, lcAcctD, lcDMExp, lcDeptNo, lcDesc, lcExpClear, llJIB
      LOCAL llJibNet, llNoPostDM, llReturn, lnAmount, lnBackWith, lnCheck, lnCompress, lnCredits, lnDebits
      LOCAL lnDefSwitch, lnDeficit, lnExpenses, lnFreqs, lnGTax1, lnGTax1p, lnGTax2, lnGTax2p, lnGTax3
      LOCAL lnGTax3p, lnGTax4, lnGTax4p, lnGasRev, lnGasTax, lnGathering, lnHold, lnHolds, lnIntHold
      LOCAL lnMarketing, lnMi1Rev, lnMi2Rev, lnMin, lnMinSwitch, lnMinimum, lnOTax1, lnOTax1p, lnOTax2
      LOCAL lnOTax2p, lnOTax3, lnOTax3p, lnOTax4, lnOTax4p, lnOilRev, lnOilTax, lnOthRev, lnPTax1
      LOCAL lnPTax1p, lnPTax2, lnPTax2p, lnPTax3, lnPTax3p, lnPTax4, lnPTax4p, lnRevenue, lnTaxWith
      LOCAL lnTrpRev, lnVendor, loError
      LOCAL cBackWith, cBatch, cCRAcctV, cCustName, cDRAcctV, cDefAcct, cDisbAcct, cGathAcct, cID
      LOCAL cMinAcct, cownerid, cRevSource, cTaxAcct1, cTaxAcct2, cTaxAcct3, cTaxAcct4, cUnitNo
      LOCAL cVendComp, cWellID, ccateg, cidchec, cownname, csusptype, nCheck, nCompress, nDeficit
      LOCAL nExpenses, nFreqs, nGTax1, nGTax1P, nGTax2, nGTax2P, nGTax3, nGTax3P, nGTax4, nGTax4P
      LOCAL nGather, nGathering, nHolds, nIntHold, nMarketing, nMinimum, nOTax1, nOTax1P, nOTax2
      LOCAL nOTax2P, nOTax3, nOTax3P, nOTax4, nOTax4P, nPTax1, nPTax1P, nPTax2, nPTax2P, nPTax3
      LOCAL nPTax3P, nPTax4, nPTax4P, nRevenue, namount, nbackwith, ngasrev, ngastax, noilrev, noiltax
      LOCAL nothrev, nprocess, ntaxwith, ntotal, tdCompanyPost

      llReturn = .T.

      TRY
         IF THIS.lerrorflag
            llReturn = .F.
            EXIT
         ENDIF

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting to General Ledger...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnMax       = 0
         lnCount     = 0
         lnTotal     = 0
         lcName      = 'Owner'
         m.cCustName = ' '
         lcIDChec    = ''

         lcAcctYear  = STR(YEAR(THIS.dpostdate), 4)
         lcAcctMonth = PADL(ALLTRIM(STR(MONTH(THIS.dpostdate), 2)), 2, '0')

*  Set the posting dates
         IF THIS.lAdvPosting = .T.
            tdCompanyPost = THIS.dCompanyShare
            tdPostDate    = THIS.dCheckDate
         ELSE
            tdCompanyPost = THIS.dCheckDate
            tdPostDate    = THIS.dCheckDate
         ENDIF

*  Plug the DM batch number into glmaint so that each
*  batch created can be traced to this closing
         THIS.ogl.DMBatch  = THIS.cdmbatch
         THIS.ogl.cSource  = 'DM'
         THIS.ogl.nDebits  = 0
         THIS.ogl.nCredits = 0
         THIS.ogl.dGLDate  = tdPostDate

*  Get the suspense account from glopt
         swselect('glopt')
         lcSuspense = cSuspense
         IF EMPTY(lcSuspense)
            lcSuspense = '999999'
         ENDIF
         lcRevClear = crevclear
         lcExpClear = cexpclear
         llNoPostDM = lDMNoPost

*  Get the A/P account
         swselect('apopt')
         lcAPAcct = capacct

* Get the suspense types before this run so we know how to post the owners
         THIS.osuspense.GetLastType(.F., .T., THIS.cgroup, .T.)

*  Set up the parameters used by processing in this method
         tcYear   = THIS.crunyear
         tcPeriod = THIS.cperiod
         tcGroup  = THIS.cgroup

*   Get Disbursement Checking Acct Number
         m.cDisbAcct = THIS.oOptions.cDisbAcct
         IF EMPTY(ALLT(m.cDisbAcct))
            m.cDisbAcct = lcSuspense
         ENDIF

         m.cVendComp = THIS.oOptions.cVendComp

         m.cGathAcct = THIS.oOptions.cGathAcct
         IF EMPTY(ALLT(m.cGathAcct))
            m.cGathAcct = lcSuspense
         ENDIF
         m.cBackWith = THIS.oOptions.cBackAcct
         IF EMPTY(ALLT(m.cBackWith))
            m.cBackWith = lcSuspense
         ENDIF
         m.cTaxAcct1  = THIS.oOptions.cTaxAcct1
         IF EMPTY(ALLT(m.cTaxAcct1))
            m.cTaxAcct1 = lcSuspense
         ENDIF
         m.cTaxAcct2  = THIS.oOptions.cTaxAcct2
         IF EMPTY(ALLT(m.cTaxAcct2))
            m.cTaxAcct2 = lcSuspense
         ENDIF
         m.cTaxAcct3  = THIS.oOptions.cTaxAcct3
         IF EMPTY(ALLT(m.cTaxAcct3))
            m.cTaxAcct3 = lcSuspense
         ENDIF
         m.cTaxAcct4 = THIS.oOptions.cTaxAcct4
         IF EMPTY(ALLT(m.cTaxAcct4))
            m.cTaxAcct4 = lcSuspense
         ENDIF
         m.cDefAcct  = THIS.oOptions.cDefAcct
         IF EMPTY(ALLT(m.cDefAcct))
            m.cDefAcct = lcSuspense
         ENDIF
         m.cMinAcct  = THIS.oOptions.cMinAcct
         IF EMPTY((m.cMinAcct))
            m.cMinAcct = lcSuspense
         ENDIF
         lcDMExp = THIS.oOptions.cFixedAcct
         IF EMPTY(lcDMExp)
            lcDMExp = lcAPAcct
         ENDIF
         IF m.goApp.lAMVersion
            lcDeptNo         = THIS.oOptions.cdeptno
            THIS.ogl.cdeptno = lcDeptNo
         ELSE
            lcDeptNo = ''
         ENDIF

         llExpSum    = THIS.oOptions.lexpsum

         llJibNet    = .T.
         IF TYPE('m.goApp') = 'O'
* Turn off net jib processing for disb mgr
            IF m.goApp.ldmpro
               llJibNet = .F.
* Don't create journal entries for stand-alone disb mgr
               llNoPostDM = .T.
            ENDIF
         ENDIF
         llSepClose  = .T.

*   Check to see if vendor compression & gathering is to be posted
         llIntegComp = .F.

         IF NOT EMPTY(ALLT(m.cVendComp))
            swselect('vendor')
            SET ORDER TO cvendorid
            IF SEEK(m.cVendComp)
               IF lIntegGL
                  llIntegComp = .T.
               ENDIF
            ENDIF
         ENDIF

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Compression and Gathering to General Ledger...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

*   Post compression and gathering
         THIS.ogl.cBatch = GetNextPK('BATCH')

         IF NOT m.goApp.ldmpro
            SELE SUM(wellwork.nCompress) AS nCompress, ;
               SUM(wellwork.nGather)   AS nGather ;
               FROM wellwork ;
               JOIN wells ON wells.cWellID = wellwork.cWellID ;
               WHERE (wells.lcompress OR wells.lGather) ;
               INTO CURSOR tempcomp ;
               ORDER BY wellwork.cWellID GROUP BY wellwork.cWellID

            IF _TALLY > 0
               SELE tempcomp
               SCAN FOR nCompress # 0 OR nGather # 0
                  SCATTER MEMVAR
                  m.nCompress         = nCompress
                  m.nGather           = nGather
                  THIS.ogl.cReference = 'Period: ' + THIS.cyear + '/' + THIS.cperiod + '/' + THIS.cgroup
                  THIS.ogl.cyear      = THIS.cyear
                  THIS.ogl.cperiod    = THIS.cperiod
                  THIS.ogl.dCheckDate = THIS.dacctdate
                  IF llIntegComp
                     THIS.ogl.dGLDate  = tdCompanyPost
                  ELSE
                     THIS.ogl.dGLDate  = tdPostDate
                  ENDIF
                  THIS.ogl.cDesc      = 'Compression/Gathering'
                  THIS.ogl.cID        = ''
                  THIS.ogl.cidtype    = ''
                  THIS.ogl.cSource    = 'DM'
                  THIS.ogl.cAcctNo    = m.cGathAcct
                  THIS.ogl.cgroup     = THIS.cgroup
                  THIS.ogl.cEntryType = 'C'
                  THIS.ogl.cUnitNo    = m.cWellID
                  THIS.ogl.namount    = (m.nCompress + m.nGather) * -1
                  THIS.ogl.UpdateBatch()

                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.namount = (m.nCompress + m.nGather)
                  THIS.ogl.UpdateBatch()
               ENDSCAN
            ENDIF
         ENDIF

*   Create Investor and Vendor Checks
         llReturn = THIS.ownerchks()
         IF NOT llReturn
            EXIT
         ENDIF

         llReturn = THIS.vendorchks()
         IF NOT llReturn
            EXIT
         ENDIF

         llReturn = THIS.directdeposit()
         IF NOT llReturn
            EXIT
         ENDIF

*   Post owner amounts to G/L
         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Owner Checks to General Ledger (Summary)...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         swselect('wells')
         SET ORDER TO cWellID

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Owner Checks to General Ledger (Summary)...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         CREATE CURSOR temppost ;
            (cWellID      c(10), ;
              nRevenue     N(12, 2), ;
              nExpenses    N(12, 2), ;
              nOTax1       N(12, 2), ;
              nOTax2       N(12, 2), ;
              nOTax3       N(12, 2), ;
              nOTax4       N(12, 2), ;
              nOTax1P      N(12, 2), ;
              nOTax2P      N(12, 2), ;
              nOTax3P      N(12, 2), ;
              nOTax4P      N(12, 2), ;
              nGTax1       N(12, 2), ;
              nGTax2       N(12, 2), ;
              nGTax3       N(12, 2), ;
              nGTax4       N(12, 2), ;
              nGTax1P      N(12, 2), ;
              nGTax2P      N(12, 2), ;
              nGTax3P      N(12, 2), ;
              nGTax4P      N(12, 2), ;
              nPTax1       N(12, 2), ;
              nPTax2       N(12, 2), ;
              nPTax3       N(12, 2), ;
              nPTax4       N(12, 2), ;
              nPTax1P      N(12, 2), ;
              nPTax2P      N(12, 2), ;
              nPTax3P      N(12, 2), ;
              nPTax4P      N(12, 2), ;
              nCompress    N(12, 2), ;
              nGathering   N(12, 2), ;
              nMarketing   N(12, 2), ;
              nbackwith    N(12, 2), ;
              ntaxwith     N(12, 2), ;
              nIntHold     N(12, 2), ;
              nMinimum     N(12, 2), ;
              nDeficit     N(12, 2), ;
              nHolds       N(12, 2), ;
              nFreqs       N(12, 2), ;
              nDefSwitch   N(12, 2), ;
              nMinSwitch   N(12, 2), ;
              nCheck       N(12, 2))

         IF NOT llNoPostDM
* Get a cursor of owners to be posted from invtmp
            SELECT cownerid, SUM(nnetcheck) AS ntotal FROM invtmp WITH (BUFFERING = .T.) ORDER BY cownerid GROUP BY cownerid INTO CURSOR tmpowners
            IF _TALLY > 0
               INDEX ON cownerid TAG owner
            ENDIF

            lnCount = 1

            swselect('wells')
            COUNT FOR NOT DELETED() TO lnMax

            swselect('investor')
            SET ORDER TO cownerid

            swselect('wells')
            SCAN
               SCATTER FIELDS LIKE lSev* MEMVAR
               m.cWellID      = cWellID
               m.lDirOilPurch = lDirOilPurch
               m.lDirGasPurch = lDirGasPurch


               THIS.oprogress.SetProgressMessage('Posting Owner Checks Amts by Well- (Summary)...' + m.cWellID)

               STORE 0 TO lnRevenue, lnExpenses, lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnCheck, lnMinimum, lnDeficit, lnHolds, lnFreqs
               STORE 0 TO lnCompress, lnGathering, lnMarketing, lnBackWith, lnTaxWith, lnIntHold, lnHold, lnCheck
               STORE 0 TO lnGTax1, lnGTax2, lnGTax3, lnGTax4, lnPTax1, lnPTax2, lnPTax3, lnPTax4
               STORE 0 TO lnOTax1p, lnOTax2p, lnOTax3p, lnOTax4p, lnGTax1p, lnGTax2p, lnGTax3p, lnGTax4p, lnPTax1p, lnPTax2p, lnPTax3p, lnPTax4p

               THIS.ogl.cBatch = GetNextPK('BATCH')

               m.cidchec = ' '
               SELECT invtmp
               SCAN FOR cWellID = m.cWellID AND csusptype = ' '
                  SCATTER MEMVAR

                  IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
                     llReturn          = .F.
                     IF NOT m.goApp.CancelMsg()
                        THIS.lCanceled = .T.
                        EXIT
                     ENDIF
                  ENDIF

* Make sure this owner is one that needs to be posted
                  SELECT tmpowners
                  IF NOT SEEK(m.cownerid)
                     LOOP
                  ENDIF

                  SELECT investor
                  IF SEEK(m.cownerid)
                     m.cownname     = cownname
* Don't post "Dummy" owner amounts
                     IF investor.ldummy
                        LOOP
                     ENDIF
* Don't post owners that are transfered to G/L here.
                     IF investor.lIntegGL
                        LOOP
                     ENDIF
                  ELSE
                     LOOP
                  ENDIF

* Don't post zero amount records
                  IF (m.nIncome = 0 AND m.nexpense = 0 AND m.nsevtaxes = 0 AND m.nnetcheck = 0)
                     LOOP
                  ENDIF
                  lcIDChec = m.cidchec

                  lnRevenue   = lnRevenue + m.nIncome
*  Remove direct paid amounts
                  DO CASE
                     CASE m.cdirect = 'O'
                        lnRevenue = lnRevenue - m.noilrev
                     CASE m.cdirect = 'G'
                        lnRevenue = lnRevenue - m.ngasrev
                     CASE m.cdirect = 'B'
                        lnRevenue = lnRevenue - m.noilrev - m.ngasrev
                  ENDCASE

* Post the sev taxes
                  IF m.noiltax1 # 0
                     IF NOT m.lsev1o
                        IF NOT m.lDirOilPurch
                           lnOTax1 = lnOTax1 - m.noiltax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax1 = lnOTax1 - m.noiltax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax1p = lnOTax1p - m.noiltax1
                     ENDIF
                  ENDIF

                  IF m.ngastax1 # 0
                     IF NOT m.lsev1g
                        IF NOT m.lDirGasPurch
                           lnGTax1 = lnGTax1 - m.ngastax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax1 = lnGTax1 - m.ngastax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax1p = lnGTax1p - m.ngastax1
                     ENDIF
                  ENDIF

                  IF m.nOthTax1 # 0
                     IF NOT m.lsev1p
                        lnPTax1 = lnPTax1 - m.nOthTax1
                     ELSE
                        lnPTax1p = lnPTax1p - m.nOthTax1
                     ENDIF
                  ENDIF

                  IF m.noiltax2 # 0
                     IF NOT m.lsev2o
                        IF NOT m.lDirOilPurch
                           lnOTax2 = lnOTax2 - m.noiltax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax2 = lnOTax2 - m.noiltax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax2p = lnOTax2p - m.noiltax2
                     ENDIF
                  ENDIF

                  IF m.ngastax2 # 0
                     IF NOT m.lsev2g
                        IF NOT m.lDirGasPurch
                           lnGTax2 = lnGTax2 - m.ngastax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax2 = lnGTax2 - m.ngastax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax2p = lnGTax2p - m.ngastax2
                     ENDIF
                  ENDIF

                  IF m.nOthTax2 # 0
                     IF NOT m.lsev2p
                        lnPTax2 = lnPTax2 - m.nOthTax2
                     ELSE
                        lnPTax2p = lnPTax2p - m.nOthTax2
                     ENDIF
                  ENDIF

                  IF m.noiltax3 # 0
                     IF NOT m.lsev3o
                        IF NOT m.lDirOilPurch
                           lnOTax3 = lnOTax3 - m.noiltax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax3 = lnOTax3 - m.noiltax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax3p = lnOTax3p - m.noiltax3
                     ENDIF
                  ENDIF

                  IF m.ngastax3 # 0
                     IF NOT m.lsev3g
                        IF NOT m.lDirGasPurch
                           lnGTax3 = lnGTax3 - m.ngastax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax3 = lnGTax3 - m.ngastax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax3p = lnGTax3p - m.ngastax3
                     ENDIF
                  ENDIF

                  IF m.nOthTax3 # 0
                     IF NOT m.lsev3p
                        lnPTax3 = lnPTax3 - m.nOthTax3
                     ELSE
                        lnPTax3p = lnPTax3p - m.nOthTax3
                     ENDIF
                  ENDIF

                  IF m.noiltax4 # 0
                     IF NOT m.lsev4o
                        IF NOT m.lDirOilPurch
                           lnOTax4 = lnOTax4 - m.noiltax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax4 = lnOTax4 - m.noiltax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax4p = lnOTax4p - m.noiltax4
                     ENDIF
                  ENDIF

                  IF m.ngastax4 # 0
                     IF NOT m.lsev4g
                        IF NOT m.lDirGasPurch
                           lnGTax4 = lnGTax4 - m.ngastax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax4 = lnGTax4 - m.ngastax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax4p = lnGTax4p - m.ngastax4
                     ENDIF
                  ENDIF

                  IF m.nOthTax4 # 0
                     IF NOT m.lsev4p
                        lnPTax4 = lnPTax4 - m.nOthTax4
                     ELSE
                        lnPTax4p = lnPTax4p - m.nOthTax4
                     ENDIF
                  ENDIF

*  Post compression and gathering
                  IF m.nCompress # 0
                     lnCompress = lnCompress - m.nCompress
                  ENDIF

                  IF m.nGather # 0
                     lnGathering = lnGathering - m.nGather
                  ENDIF

*  Post marketing expenses
                  IF m.nMKTGExp # 0
                     lnMarketing = lnMarketing - m.nMKTGExp
                  ENDIF

*  Post the Expenses
                  lnExpenses = lnExpenses + m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + m.ntotalea + m.ntotaleb

*  Post Backup Withholding
                  IF m.nbackwith # 0
                     lnBackWith = lnBackWith - m.nbackwith
                  ENDIF

*  Post Tax Withholding
                  IF m.ntaxwith # 0
                     lnTaxWith = lnTaxWith - m.ntaxwith
                  ENDIF

                  lnCheck = lnCheck + m.nnetcheck

               ENDSCAN && Invtmp

               SELECT invtmp
               SCAN FOR cWellID = m.cWellID AND csusptype <> ' '
                  SCATTER MEMVAR

                  IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
                     llReturn          = .F.
                     IF NOT m.goApp.CancelMsg()
                        THIS.lCanceled = .T.
                        EXIT
                     ENDIF
                  ENDIF

* Post prior suspense
                  SELECT curLastSuspType
                  LOCATE FOR cownerid == m.cownerid AND cWellID == m.cWellID AND ctypeinv == m.ctypeinv
                  IF FOUND()
                     m.csusptype = csusptype

*  Post Prior Period Deficits
                     IF m.csusptype = 'D'
                        lnDeficit = lnDeficit - m.nnetcheck
                     ENDIF

*  Post Prior Period Minimums
                     IF m.csusptype = 'M'
                        lnMinimum = lnMinimum + m.nnetcheck
                     ENDIF

*  Post Interest on Hold being released
                     IF m.csusptype = 'I'
                        lnIntHold = lnIntHold + m.nnetcheck
                     ENDIF

*  Post Owner on Hold being released
                     IF m.csusptype = 'H'
                        lnHolds = lnHolds + m.nnetcheck
                     ENDIF

*  Post Quarterly Owner being released
                     IF INLIST(m.csusptype, 'Q', 'S', 'A')
                        lnFreqs = lnFreqs + m.nnetcheck
                     ENDIF

                     lnCheck = lnCheck + m.nnetcheck
                  ENDIF
               ENDSCAN  && Invtmp

* Add the totals to the temppost cursor
               m.nRevenue   = lnRevenue
               m.nExpenses  = lnExpenses
               m.nOTax1     = lnOTax1
               m.nOTax2     = lnOTax2
               m.nOTax3     = lnOTax3
               m.nOTax4     = lnOTax4
               m.nOTax1P    = lnOTax1p
               m.nOTax2P    = lnOTax2p
               m.nOTax3P    = lnOTax3p
               m.nOTax4P    = lnOTax4p
               m.nGTax1     = lnGTax1
               m.nGTax2     = lnGTax2
               m.nGTax3     = lnGTax3
               m.nGTax4     = lnGTax4
               m.nGTax1P    = lnGTax1p
               m.nGTax2P    = lnGTax2p
               m.nGTax3P    = lnGTax3p
               m.nGTax4P    = lnGTax4p
               m.nPTax1     = lnPTax1
               m.nPTax2     = lnPTax2
               m.nPTax3     = lnPTax3
               m.nPTax4     = lnPTax4
               m.nPTax1P    = lnPTax1p
               m.nPTax2P    = lnPTax2p
               m.nPTax3P    = lnPTax3p
               m.nPTax4P    = lnPTax4p
               m.nCompress  = lnCompress
               m.nGathering = lnGathering
               m.nMarketing = lnMarketing
               m.nbackwith  = lnBackWith
               m.ntaxwith   = lnTaxWith
               m.nIntHold   = lnIntHold
               m.nMinimum   = lnMinimum
               m.nDeficit   = lnDeficit
               m.nHolds     = lnHolds
               m.nFreqs     = lnFreqs
               m.nCheck     = lnCheck
               INSERT INTO temppost FROM MEMVAR

            ENDSCAN  && Wells
            lcIDChec = ''

*  Post amounts going into suspense this run
            IF THIS.lclose
               THIS.oprogress.SetProgressMessage('Posting Owner Suspense to General Ledger (Summary)...')
               THIS.oprogress.UpdateProgress(THIS.nprogress)
               THIS.nprogress = THIS.nprogress + 1
            ENDIF

            SELECT  cownerid, ;
                    csusptype, ;
                    SUM(nnetcheck) AS ntotal ;
                FROM tsuspense ;
                WHERE nrunno_in = THIS.nrunno ;
                    AND crunyear_in = THIS.crunyear ;
                ORDER BY cownerid,;
                    csusptype ;
                GROUP BY cownerid,;
                    csusptype ;
                INTO CURSOR tmpowners READWRITE
            INDEX ON cownerid TAG cownerid

            swselect('wells')
            SET ORDER TO cWellID
            SCAN
               SCATTER FIELDS LIKE lSev* MEMVAR
               m.cWellID      = cWellID
               m.lDirOilPurch = lDirOilPurch
               m.lDirGasPurch = lDirGasPurch

               THIS.oprogress.SetProgressMessage('Posting Owner Suspense to General Ledger (Summary)...' + m.cWellID)

               STORE 0 TO lnRevenue, lnExpenses, lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnCheck, lnMinimum, lnDeficit, lnHolds, lnFreqs
               STORE 0 TO lnCompress, lnGathering, lnMarketing, lnBackWith, lnTaxWith, lnIntHold, lnHold, lnCheck
               STORE 0 TO lnGTax1, lnGTax2, lnGTax3, lnGTax4, lnPTax1, lnPTax2, lnPTax3, lnPTax4
               STORE 0 TO lnOTax1p, lnOTax2p, lnOTax3p, lnOTax4p, lnGTax1p, lnGTax2p, lnGTax3p, lnGTax4p, lnPTax1p, lnPTax2p, lnPTax3p, lnPTax4p

               SELECT tsuspense
               SCAN FOR cWellID == m.cWellID AND nrunno_in = THIS.nrunno AND crunyear_in = THIS.crunyear
                  SCATTER MEMVAR

* Make sure this owner should be posted
                  SELECT tmpowners
                  IF NOT SEEK(m.cownerid)
                     LOOP
                  ELSE
                     m.ntotal = ntotal
                  ENDIF

                  lnRevenue   = lnRevenue + m.nIncome
*  Remove direct paid amounts
                  DO CASE
                     CASE m.cdirect = 'O'
                        lnRevenue = lnRevenue - m.noilrev
                     CASE m.cdirect = 'G'
                        lnRevenue = lnRevenue - m.ngasrev
                     CASE m.cdirect = 'B'
                        lnRevenue = lnRevenue - m.noilrev - m.ngasrev
                  ENDCASE

*!*                       IF m.nflatrate <> 0
*!*                           lnRevenue = lnRevenue + m.nflatrate
*!*                       ENDIF

* Post the sev taxes
                  IF m.noiltax1 # 0
                     IF NOT m.lsev1o
                        IF NOT m.lDirOilPurch
                           lnOTax1 = lnOTax1 - m.noiltax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax1 = lnOTax1 - m.noiltax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax1p = lnOTax1p - m.noiltax1
                     ENDIF
                  ENDIF

                  IF m.ngastax1 # 0
                     IF NOT m.lsev1g
                        IF NOT m.lDirGasPurch
                           lnGTax1 = lnGTax1 - m.ngastax1
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax1 = lnGTax1 - m.ngastax1
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax1p = lnGTax1p - m.ngastax1
                     ENDIF
                  ENDIF

                  IF m.nOthTax1 # 0
                     IF NOT m.lsev1p
                        lnPTax1 = lnPTax1 - m.nOthTax1
                     ELSE
                        lnPTax1p = lnPTax1p - m.nOthTax1
                     ENDIF
                  ENDIF

                  IF m.noiltax2 # 0
                     IF NOT m.lsev2o
                        IF NOT m.lDirOilPurch
                           lnOTax2 = lnOTax2 - m.noiltax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax2 = lnOTax2 - m.noiltax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax2p = lnOTax2p - m.noiltax2
                     ENDIF
                  ENDIF

                  IF m.ngastax2 # 0
                     IF NOT m.lsev2g
                        IF NOT m.lDirGasPurch
                           lnGTax2 = lnGTax2 - m.ngastax2
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax2 = lnGTax2 - m.ngastax2
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax2p = lnGTax2p - m.ngastax2
                     ENDIF
                  ENDIF

                  IF m.nOthTax2 # 0
                     IF NOT m.lsev2p
                        lnPTax2 = lnPTax2 - m.nOthTax2
                     ELSE
                        lnPTax2 = lnPTax2 - m.nOthTax2
                     ENDIF
                  ENDIF

                  IF m.noiltax3 # 0
                     IF NOT m.lsev3o
                        IF NOT m.lDirOilPurch
                           lnOTax3 = lnOTax3 - m.noiltax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax3 = lnOTax3 - m.noiltax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax3p = lnOTax3p - m.noiltax3
                     ENDIF
                  ENDIF

                  IF m.ngastax3 # 0
                     IF NOT m.lsev3g
                        IF NOT m.lDirGasPurch
                           lnGTax3 = lnGTax3 - m.ngastax3
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax3 = lnGTax3 - m.ngastax3
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax3p = lnGTax3p - m.ngastax3
                     ENDIF
                  ENDIF

                  IF m.nOthTax3 # 0
                     IF NOT m.lsev3p
                        lnPTax3 = lnPTax3 - m.nOthTax3
                     ELSE
                        lnPTax3p = lnPTax3p - m.nOthTax3
                     ENDIF
                  ENDIF

                  IF m.noiltax4 # 0
                     IF NOT m.lsev4o
                        IF NOT m.lDirOilPurch
                           lnOTax4 = lnOTax4 - m.noiltax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'O')
                              lnOTax4 = lnOTax4 - m.noiltax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnOTax4p = lnOTax4p - m.noiltax4
                     ENDIF
                  ENDIF

                  IF m.ngastax4 # 0
                     IF NOT m.lsev4g
                        IF NOT m.lDirGasPurch
                           lnGTax4 = lnGTax4 - m.ngastax4
                        ELSE
                           IF NOT INLIST(m.cdirect, 'B', 'G')
                              lnGTax4 = lnGTax4 - m.ngastax4
                           ENDIF
                        ENDIF
                     ELSE
                        lnGTax4p = lnGTax4p - m.ngastax4
                     ENDIF
                  ENDIF

                  IF m.nOthTax4 # 0
                     IF NOT m.lsev4p
                        lnPTax4 = lnPTax4 - m.nOthTax4
                     ELSE
                        lnPTax4p = lnPTax4p - m.nOthTax4
                     ENDIF
                  ENDIF

*  Post compression and gathering
                  IF m.nCompress # 0
                     lnCompress = lnCompress - m.nCompress
                  ENDIF

                  IF m.nGather # 0
                     lnGathering = lnGathering - m.nGather
                  ENDIF

*
*  Post marketing expenses
*
                  IF m.nMKTGExp # 0
                     lnMarketing = lnMarketing - m.nMKTGExp
                  ENDIF

*  Post the Expenses
                  lnExpenses = lnExpenses + m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + m.ntotalea + m.ntotaleb

*  Post Backup Withholding
                  IF m.nbackwith # 0
                     lnBackWith = lnBackWith - m.nbackwith
                  ENDIF

*  Post Tax Withholding
                  IF m.ntaxwith # 0
                     lnTaxWith = lnTaxWith - m.ntaxwith
                  ENDIF

*  Post net
                  DO CASE
                     CASE m.csusptype = 'D'
                        lnDeficit = lnDeficit + m.nnetcheck
                     CASE m.csusptype = 'I'
                        lnIntHold = lnIntHold - m.nnetcheck
                     CASE m.csusptype = 'M'
                        lnMinimum = lnMinimum - m.nnetcheck
                     CASE m.csusptype = 'H'
                        lnHolds  = lnHolds - m.nnetcheck
                     CASE INLIST(m.csusptype, 'Q', 'S', 'A')
                        lnFreqs = lnFreqs - m.nnetcheck
                  ENDCASE
               ENDSCAN  && Tsuspense

* Add the totals to the temppost cursor
               m.nRevenue   = lnRevenue
               m.nExpenses  = lnExpenses
               m.nOTax1     = lnOTax1
               m.nOTax2     = lnOTax2
               m.nOTax3     = lnOTax3
               m.nOTax4     = lnOTax4
               m.nOTax1P    = lnOTax1p
               m.nOTax2P    = lnOTax2p
               m.nOTax3P    = lnOTax3p
               m.nOTax4P    = lnOTax4p
               m.nGTax1     = lnGTax1
               m.nGTax2     = lnGTax2
               m.nGTax3     = lnGTax3
               m.nGTax4     = lnGTax4
               m.nGTax1P    = lnGTax1p
               m.nGTax2P    = lnGTax2p
               m.nGTax3P    = lnGTax3p
               m.nGTax4P    = lnGTax4p
               m.nPTax1     = lnPTax1
               m.nPTax2     = lnPTax2
               m.nPTax3     = lnPTax3
               m.nPTax4     = lnPTax4
               m.nPTax1P    = lnPTax1p
               m.nPTax2P    = lnPTax2p
               m.nPTax3P    = lnPTax3p
               m.nPTax4P    = lnPTax4p
               m.nCompress  = lnCompress
               m.nGathering = lnGathering
               m.nMarketing = lnMarketing
               m.nbackwith  = lnBackWith
               m.ntaxwith   = lnTaxWith
               m.nIntHold   = lnIntHold
               m.nMinimum   = lnMinimum
               m.nDeficit   = lnDeficit
               m.nHolds     = lnHolds
               m.nFreqs     = lnFreqs
               m.nCheck     = 0
               INSERT INTO temppost FROM MEMVAR

            ENDSCAN  && Wells

            SELECT  cWellID, ;
                    SUM(nRevenue) AS nRevenue, ;
                    SUM(nExpenses) AS nExpenses, ;
                    SUM(nOTax1)    AS nOTax1, ;
                    SUM(nOTax2)    AS nOTax2, ;
                    SUM(nOTax3)    AS nOTax3, ;
                    SUM(nOTax4)    AS nOTax4, ;
                    SUM(nOTax1P)   AS nOTax1P, ;
                    SUM(nOTax2P)   AS nOTax2P, ;
                    SUM(nOTax3P)   AS nOTax3P, ;
                    SUM(nOTax4P)   AS nOTax4P, ;
                    SUM(nGTax1)    AS nGTax1, ;
                    SUM(nGTax2)    AS nGTax2, ;
                    SUM(nGTax3)    AS nGTax3, ;
                    SUM(nGTax4)    AS nGTax4, ;
                    SUM(nGTax1P)   AS nGTax1P, ;
                    SUM(nGTax2P)   AS nGTax2P, ;
                    SUM(nGTax3P)   AS nGTax3P, ;
                    SUM(nGTax4P)   AS nGTax4P, ;
                    SUM(nPTax1)    AS nPTax1, ;
                    SUM(nPTax2)    AS nPTax2, ;
                    SUM(nPTax3)    AS nPTax3, ;
                    SUM(nPTax4)    AS nPTax4, ;
                    SUM(nPTax1P)   AS nPTax1P, ;
                    SUM(nPTax2P)   AS nPTax2P, ;
                    SUM(nPTax3P)   AS nPTax3P, ;
                    SUM(nPTax4P)   AS nPTax4P, ;
                    SUM(nCompress) AS nCompress, ;
                    SUM(nGathering) AS nGathering, ;
                    SUM(nMarketing) AS nMarketing, ;
                    SUM(nbackwith) AS nbackwith, ;
                    SUM(ntaxwith) AS ntaxwith, ;
                    SUM(nIntHold) AS nIntHold, ;
                    SUM(nMinimum) AS nMinimum, ;
                    SUM(nDeficit) AS nDeficit, ;
                    SUM(nHolds)   AS nHolds, ;
                    SUM(nFreqs)  AS nFreqs, ;
                    SUM(nCheck)  AS nCheck ;
                FROM temppost ;
                INTO CURSOR temppostsum ;
                ORDER BY cWellID ;
                GROUP BY cWellID

* Post suspense amounts that are moving between deficit
* and minimum suspense.
            lnDefSwitch = 0
            lnMinSwitch = 0

            THIS.oprogress.SetProgressMessage('Posting Owner Suspense Switching to Minimum')
* Post amounts transfering between deficit and minimum
* Get the amount of deficits that transferred
            lnDefSwitch = THIS.osuspense.GetBalTransfer('D', .T.)

* Get the amount of minimums that transferred
            lnMinSwitch = THIS.osuspense.GetBalTransfer('M', .T.)

            THIS.oprogress.SetProgressMessage('Posting Owner Suspense - Finishing')
            THIS.ogl.cidchec    = ''
            THIS.ogl.cReference = 'Run: R' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + '/' + THIS.cgroup
            THIS.ogl.cID        = ''
            THIS.ogl.dGLDate    = tdPostDate
            THIS.ogl.cBatch     = GetNextPK('BATCH')
            THIS.ogl.cUnitNo    = ''
            THIS.ogl.cdeptno    = ''

            SELECT temppostsum
            SCAN
               SCATTER MEMVAR
               THIS.ogl.cUnitNo = m.cWellID

* Post to Revenue Clearing
               IF m.nRevenue # 0
                  THIS.ogl.cDesc   = 'Revenue'
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.namount = m.nRevenue
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Expense Clearing
               IF m.nExpenses # 0
                  THIS.ogl.cDesc   = 'Expenses'
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.namount = m.nExpenses * -1
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Taxes
               IF m.nOTax1 # 0 OR m.nOTax1P # 0
                  THIS.ogl.cDesc     = 'Oil Tax 1'
                  IF m.nOTax1P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nOTax1P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax1 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.namount = m.nOTax1
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nOTax2 # 0 OR m.nOTax2P # 0
                  THIS.ogl.cDesc     = 'Oil Tax 2'
                  IF m.nOTax2P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nOTax2P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax2 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.namount = m.nOTax2
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nOTax3 # 0 OR m.nOTax3P # 0
                  THIS.ogl.cDesc     = 'Oil Tax 3'
                  IF m.nOTax3P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nOTax3P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax3 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.namount = m.nOTax3
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nOTax4 # 0 OR m.nOTax4P # 0
                  THIS.ogl.cDesc     = 'Oil Tax 4'
                  IF m.nOTax4P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nOTax4P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax4 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.namount = m.nOTax4
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax1 # 0 OR m.nGTax1P # 0
                  THIS.ogl.cDesc     = 'Gas Tax 1'
                  IF m.nGTax1P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nGTax1P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax1 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.namount = m.nGTax1
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax2 # 0 OR m.nGTax2P # 0
                  THIS.ogl.cDesc     = 'Gas Tax 2'
                  IF m.nGTax2P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nGTax2P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax2 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.namount = m.nGTax2
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax3 # 0 OR m.nGTax3P # 0
                  THIS.ogl.cDesc     = 'Gas Tax 3'
                  IF m.nGTax3P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nGTax3P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax3 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.namount = m.nGTax3
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax4 # 0 OR m.nGTax4P # 0
                  THIS.ogl.cDesc     = 'Gas Tax 4'
                  IF m.nGTax4P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nGTax4P
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax4 # 0
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.namount = m.nGTax4
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nPTax1 # 0 OR m.nPTax1P # 0
                  THIS.ogl.cDesc     = 'Other Tax 1'
                  IF m.nPTax1P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nPTax1P
                  ELSE
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.namount = m.nPTax1
                  ENDIF
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nPTax2 # 0 OR m.nPTax2P # 0
                  THIS.ogl.cDesc     = 'Other Tax 2'
                  IF m.nPTax2P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nPTax2P
                  ELSE
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.namount = m.nPTax2
                  ENDIF
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nPTax3 # 0 OR m.nPTax3P # 0
                  THIS.ogl.cDesc     = 'Other Tax 3'
                  IF m.nPTax3P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nPTax3P
                  ELSE
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.namount = m.nPTax3
                  ENDIF
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nPTax4 # 0 OR m.nPTax4P # 0
                  THIS.ogl.cDesc     = 'Other Tax 4'
                  IF m.nPTax4P # 0
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.namount = m.nPTax4P
                  ELSE
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.namount = m.nPTax4
                  ENDIF
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Compression & Gathering
               IF m.nCompress # 0
                  THIS.ogl.cDesc   = 'Compression'
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.namount = m.nCompress
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nGathering # 0
                  THIS.ogl.cDesc   = 'Gathering'
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.namount = m.nGathering
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Marketing
               IF m.nMarketing # 0
                  THIS.ogl.cDesc   = 'Marketing'
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.namount = m.nCompress
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Backup Withholding
               IF m.nbackwith # 0
                  THIS.ogl.cDesc   = 'Backup Withholding'
                  THIS.ogl.cAcctNo = m.cBackWith
                  THIS.ogl.namount = m.nbackwith
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Tax Withholding
               IF m.ntaxwith # 0
                  THIS.ogl.cDesc   = 'Tax Withholding'
                  THIS.ogl.cAcctNo = m.cBackWith
                  THIS.ogl.namount = m.ntaxwith
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Interest On Hold
               IF m.nIntHold # 0
                  THIS.ogl.cDesc   = 'Interest On Hold'
                  THIS.ogl.cAcctNo = m.cMinAcct
                  THIS.ogl.namount = m.nIntHold
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Minimums
               IF m.nMinimum # 0
                  THIS.ogl.cDesc   = 'Minimum Checks'
                  THIS.ogl.cAcctNo = m.cMinAcct
                  THIS.ogl.namount = m.nMinimum
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Deficits
               IF m.nDeficit # 0
                  THIS.ogl.cDesc   = 'Deficits'
                  THIS.ogl.cAcctNo = m.cDefAcct
                  THIS.ogl.namount = m.nDeficit * -1
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Owner Holds
               IF m.nHolds # 0
                  THIS.ogl.cDesc   = 'Owner Holds'
                  THIS.ogl.cAcctNo = m.cMinAcct
                  THIS.ogl.namount = m.nHolds
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Owner Frequency Holds
               IF m.nFreqs # 0
                  THIS.ogl.cDesc   = 'Owner Freq Holds'
                  THIS.ogl.cAcctNo = m.cMinAcct
                  THIS.ogl.namount = m.nFreqs
                  THIS.ogl.UpdateBatch()
               ENDIF

* Post Checks
               IF m.nCheck # 0
                  THIS.ogl.cDesc   = 'Check Amounts'
                  THIS.ogl.cAcctNo = m.cDisbAcct
                  THIS.ogl.namount = m.nCheck * -1
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDSCAN  && Temppostsum

            IF USED('baltransfer')
               STORE 0 TO lnDefSwitch, lnMinSwitch
               SELECT  cWellID, ;
                       ctype, ;
                       SUM(namount) AS namount ;
                   FROM baltransfer ;
                   INTO CURSOR temp ;
                   ORDER BY cWellID,;
                       ctype ;
                   GROUP BY cWellID,;
                       ctype
               SELECT temp
               SCAN FOR ctype = 'D'
                  SCATTER MEMVAR
* Post Deficits Switching To Minimums
                  THIS.ogl.cDesc   = 'Deficit To Minimum Switch'
                  THIS.ogl.cAcctNo = m.cMinAcct
                  THIS.ogl.namount = m.namount * -1
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.UpdateBatch()

                  THIS.ogl.cDesc   = 'Deficit To Minimum Switch'
                  THIS.ogl.cAcctNo = m.cDefAcct
                  THIS.ogl.namount = m.namount
                  THIS.ogl.UpdateBatch()
                  lnDefSwitch = lnDefSwitch + m.namount
               ENDSCAN

* Post Minimums Switching To Deficits
               SCAN FOR ctype = 'M'
                  SCATTER MEMVAR
                  THIS.ogl.cDesc   = 'Minimum To Deficit Switch'
                  THIS.ogl.cAcctNo = m.cMinAcct
                  THIS.ogl.namount = m.namount * -1
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = ''
                  THIS.ogl.UpdateBatch()

                  THIS.ogl.cDesc   = 'Minimum To Deficit Switch'
                  THIS.ogl.cAcctNo = m.cDefAcct
                  THIS.ogl.namount = m.namount
                  THIS.ogl.UpdateBatch()
                  lnMinSwitch = lnMinSwitch - m.namount
               ENDSCAN

               THIS.ndeftransfer = lnDefSwitch
               THIS.nmintransfer = lnMinSwitch

               llReturn = THIS.ogl.ChkBalance()
               swclose('baltransfer')
            ENDIF
         ENDIF

*
*  Mark the expense entries as being tied to this DM batch
*
         THIS.oprogress.SetProgressMessage('Marking Expenses For This Run')
         swselect('expense')
         SCAN FOR nRunNoRev = THIS.nrunno ;
               AND EMPTY(expense.cBatch)
            m.cWellID = cWellID
            swselect('wells')
            SET ORDER TO cWellID
            IF SEEK(m.cWellID)
               IF cgroup = tcGroup
                  swselect('expense')
                  REPL cBatch WITH THIS.cdmbatch
               ENDIF
            ENDIF
         ENDSCAN

*   Post the Vendor amounts that are designated to be posted.
         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Vendor Checks to General Ledger...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnVendor = 0

*  Get the vendors to be posted.

         SELECT cvendorid, cVendName FROM vendor WHERE lIntegGL = .T. INTO CURSOR curVends
         IF NOT llNoPostDM AND _TALLY > 0
            THIS.ogl.dGLDate = tdCompanyPost
            THIS.ogl.cBatch  = GetNextPK('BATCH')
            SELECT curVends
            SCAN
               SCATTER MEMVAR

               lnAmount     = 0
               THIS.ogl.cID = m.cvendorid
               swselect('expense')
               lnCount = 1
* Summarize by ccateg to cut down on journal entries
               SELECT  expense.cWellID, ;
                       ccateg, ;
                       cexpclass, ;
                       SUM(namount) AS namount, ;
                       cvendorid, ;
                       ccatcode, ;
                       cownerid ;
                   FROM expense WITH (BUFFERING = .T.) ;
                   WHERE cvendorid = m.cvendorid ;
                       AND nRunNoRev = THIS.nrunno ;
                       AND cRunYearRev = THIS.crunyear ;
                       AND namount # 0 ;
                       AND expense.cWellID IN (SELECT  cWellID ;
                                                   FROM wellwork) ;
                   INTO CURSOR tempexp ;
                   ORDER BY cvendorid,;
                       expense.cWellID,;
                       cownerid,;
                       ccatcode ;
                   GROUP BY cvendorid,;
                       expense.cWellID,;
                       cownerid,;
                       ccatcode

               SELECT tempexp
               COUNT FOR cvendorid = m.cvendorid TO lnMax
               SCAN
                  SCATTER MEMVAR
                  m.cUnitNo = cWellID
                  m.cID     = cvendorid

*  Get the account numbers to be posted for this expense category
                  swselect('expcat')
                  SET ORDER TO ccatcode
                  IF SEEK(m.ccatcode)
                     SCATTER MEMVAR
                     m.ccateg   = ccateg
                     m.cDRAcctV = THIS.cexpclear
                     IF EMPTY(m.cCRAcctV)
                        m.cCRAcctV = lcSuspense
                     ENDIF
                  ELSE
                     m.cCRAcctV = lcSuspense
                  ENDIF

*  Net out any JIB interest shares from the expense
                  m.namount   = swNetExp(m.namount, m.cWellID, .T., m.cexpclass, 'N', .F., m.cownerid, '', m.cdeck)

*  Add amount of this invoice to the total the vendor is to be paid
                  lnAmount = lnAmount + m.namount

                  THIS.ogl.cReference = 'Vendor Amts'
                  THIS.ogl.cUnitNo    = m.cUnitNo
                  THIS.ogl.cDesc      = m.ccateg
                  THIS.ogl.cdeptno    = lcDeptNo
                  THIS.ogl.cAcctNo    = m.cCRAcctV
                  THIS.ogl.namount    = m.namount * -1
                  THIS.ogl.UpdateBatch()

                  THIS.ogl.cAcctNo = lcDMExp
                  THIS.ogl.namount = m.namount
                  THIS.ogl.cDesc   = m.ccateg
                  THIS.ogl.UpdateBatch()
               ENDSCAN  && tempexp

               llReturn = THIS.ogl.ChkBalance()

               IF NOT llReturn
                  IF NOT FILE('outbal.dbf')
                     CREATE TABLE outbal FREE (cBatch  c(8), cownerid  c(10))
                  ENDIF
                  IF NOT USED('outbal')
                     USE outbal IN 0
                  ENDIF
                  m.cBatch   = THIS.ogl.cBatch
                  m.cownerid = m.cID
                  INSERT INTO outbal FROM MEMVAR
               ENDIF

            ENDSCAN && curVends
         ENDIF

*  Post the owners that are designated to be posted
         CREATE CURSOR postown1 ;
            (cWellID         c(10), ;
              noilrev         N(12, 2), ;
              ngasrev         N(12, 2), ;
              nothrev         N(12, 2), ;
              nmiscrev1       N(12, 2), ;
              nmiscrev2       N(12, 2), ;
              noiltax         N(12, 2), ;
              ngastax         N(12, 2), ;
              nOTax1          N(12, 2), ;
              nOTax2          N(12, 2), ;
              nOTax3          N(12, 2), ;
              nOTax4          N(12, 2), ;
              nGTax1          N(12, 2), ;
              nGTax2          N(12, 2), ;
              nGTax3          N(12, 2), ;
              nGTax4          N(12, 2), ;
              nPTax1          N(12, 2), ;
              nPTax2          N(12, 2), ;
              nPTax3          N(12, 2), ;
              nPTax4          N(12, 2), ;
              nOTax1P        N(12, 2), ;
              nOTax2P         N(12, 2), ;
              nOTax3P         N(12, 2), ;
              nOTax4P         N(12, 2), ;
              nGTax1P         N(12, 2), ;
              nGTax2P         N(12, 2), ;
              nGTax3P         N(12, 2), ;
              nGTax4P         N(12, 2), ;
              nPTax1P         N(12, 2), ;
              nPTax2P         N(12, 2), ;
              nPTax3P         N(12, 2), ;
              nPTax4P         N(12, 2), ;
              nMarketing      N(12, 2), ;
              nCompress       N(12, 2), ;
              nGathering      N(12, 2), ;
              nExpenses       N(12, 2), ;
              nDeficit        N(12, 2), ;
              nMinimum        N(12, 2))

         swselect('wells')
         SET ORDER TO cWellID

         IF THIS.lclose
            THIS.oprogress.SetProgressMessage('Posting Operator Owner Amounts to General Ledger (Summary)...')
            THIS.oprogress.UpdateProgress(THIS.nprogress)
            THIS.nprogress = THIS.nprogress + 1
         ENDIF

         lnOwner = 0

*  Get the owners to be posted.
         SELECT cownerid, cownname FROM investor WHERE lIntegGL = .T. INTO CURSOR curOwnPost

         IF NOT llNoPostDM AND _TALLY > 0
            lnAmount         = 0
            THIS.ogl.dGLDate = tdCompanyPost
            SELECT curOwnPost
            SCAN
               SCATTER MEMVAR

               STORE 0 TO lnOilRev, lnGasRev, lnOthRev, lnExpenses, lnOilTax, lnGasTax
               STORE 0 TO lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnCheck, lnMinimum, lnDeficit
               STORE 0 TO lnGTax1, lnGTax2, lnGTax3, lnGTax4
               STORE 0 TO lnPTax1, lnPTax2, lnPTax3, lnPTax4
               STORE 0 TO lnOTax1p, lnOTax2p, lnOTax3p, lnOTax4p, lnGTax1p, lnGTax2p, lnGTax3p, lnGTax4p, lnPTax1p, lnPTax2p, lnPTax3p, lnPTax4p
               STORE 0 TO lnCompress, lnGathering, lnMarketing, lnBackWith, lnTaxWith, lnIntHold, lnHold, lnCheck
               STORE 0 TO lnRYFlatRate, lnOVFlatRate

               THIS.ogl.cBatch = GetNextPK('BATCH')
               STORE 0 TO lnTrpRev, lnMi1Rev, lnMi2Rev, lnDebits, lnCredits, lnMin, lnDeficit, lnOilTax, lnGasTax
               SELECT invtmp
               COUNT FOR cownerid = m.cownerid AND nrunno = THIS.nrunno AND crunyear = THIS.crunyear TO lnMax
               lnCount = 1

               swselect('wells')
               SCAN
                  SCATTER FIELDS LIKE lSev* MEMVAR
                  m.nprocess     = nprocess
                  m.lDirOilPurch = lDirOilPurch
                  m.lDirGasPurch = lDirGasPurch
                  m.cWellID      = cWellID

                  STORE 0 TO lnOilRev, lnGasRev, lnOthRev, lnExpenses, lnOilTax, lnGasTax
                  STORE 0 TO lnOTax1, lnOTax2, lnOTax3, lnOTax4, lnCheck, lnMinimum, lnDeficit
                  STORE 0 TO lnGTax1, lnGTax2, lnGTax3, lnGTax4
                  STORE 0 TO lnPTax1, lnPTax2, lnPTax3, lnPTax4
                  STORE 0 TO lnOTax1p, lnOTax2p, lnOTax3p, lnOTax4p, lnGTax1p, lnGTax2p, lnGTax3p, lnGTax4p, lnPTax1p, lnPTax2p, lnPTax3p, lnPTax4p
                  STORE 0 TO lnCompress, lnGathering, lnMarketing, lnBackWith, lnTaxWith, lnIntHold, lnHold, lnCheck

                  SELECT invtmp
                  SCAN FOR cownerid = m.cownerid AND cWellID == m.cWellID
                     SCATTER MEMVAR

                     llJIB  = lJIB
                     lcName = m.cownname

*  Post Oil Income
                     IF m.noilrev # 0 AND NOT INLIST(m.cdirect, 'O', 'B')
                        lnOilRev = lnOilRev - m.noilrev
                     ENDIF

*  Post Gas Income
                     IF m.ngasrev # 0 AND NOT INLIST(m.cdirect, 'G', 'B')
                        lnGasRev = lnGasRev - m.ngasrev
                     ENDIF

*  Post Other Income
                     IF m.nothrev # 0
                        lnOthRev = lnOthRev - m.nothrev
                     ENDIF

*  Post Trp Income
                     IF m.ntrprev # 0
                        lnOthRev = lnOthRev - m.ntrprev
                     ENDIF

*  Post Misc 1 Income
                     IF m.nmiscrev1 # 0
                        lnOthRev = lnOthRev - m.nmiscrev1
                     ENDIF

*  Post Misc 2 Income
                     IF m.nmiscrev2 # 0
                        lnOthRev = lnOthRev - m.nmiscrev2
                     ENDIF

*  Post Oil Taxes
                     IF m.noiltax1 # 0
                        IF NOT m.lsev1o
                           IF NOT m.lDirOilPurch
                              lnOTax1 = lnOTax1 + m.noiltax1
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'O')
                                 lnOTax1 = lnOTax1 + m.noiltax1
                              ENDIF
                           ENDIF
                        ELSE
                           lnOTax1p = lnOTax1p + m.noiltax1
                        ENDIF
                     ENDIF

                     IF m.noiltax2 # 0
                        IF NOT m.lsev2o
                           IF NOT m.lDirOilPurch
                              lnOTax2 = lnOTax2 + m.noiltax2
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'O')
                                 lnOTax2 = lnOTax2 + m.noiltax2
                              ENDIF
                           ENDIF
                        ELSE
                           lnOTax2p = lnOTax2p + m.noiltax2
                        ENDIF
                     ENDIF

                     IF m.noiltax3 # 0
                        IF NOT m.lsev3o
                           IF NOT m.lDirOilPurch
                              lnOTax3 = lnOTax3 + m.noiltax3
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'O')
                                 lnOTax3 = lnOTax3 + m.noiltax3
                              ENDIF
                           ENDIF
                        ELSE
                           lnOTax3p = lnOTax3p + m.noiltax3
                        ENDIF
                     ENDIF

                     IF m.noiltax4 # 0
                        IF NOT m.lsev4o
                           IF NOT m.lDirOilPurch
                              lnOTax4 = lnOTax4 + m.noiltax4
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'O')
                                 lnOTax4 = lnOTax4 + m.noiltax4
                              ENDIF
                           ENDIF
                        ELSE
                           lnOTax4p = lnOTax4p + m.noiltax4
                        ENDIF
                     ENDIF

*  Post Gas Taxes
                     IF m.ngastax1 # 0
                        IF NOT m.lsev1g
                           IF NOT m.lDirGasPurch
                              lnGTax1 = lnGTax1 + m.ngastax1
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'G')
                                 lnGTax1 = lnGTax1 + m.ngastax1
                              ENDIF
                           ENDIF
                        ELSE
                           lnGTax1p = lnGTax1p + m.ngastax1
                        ENDIF
                     ENDIF

                     IF m.ngastax2 # 0
                        IF NOT m.lsev2g
                           IF NOT m.lDirGasPurch
                              lnGTax2 = lnGTax2 + m.ngastax2
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'G')
                                 lnGTax2 = lnGTax2 + m.ngastax2
                              ENDIF
                           ENDIF
                        ELSE
                           lnGTax2p = lnGTax2p + m.ngastax2
                        ENDIF
                     ENDIF

                     IF m.ngastax3 # 0
                        IF NOT m.lsev3g
                           IF NOT m.lDirGasPurch
                              lnGTax3 = lnGTax3 + m.ngastax3
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'G')
                                 lnGTax3 = lnGTax3 + m.ngastax3
                              ENDIF
                           ENDIF
                        ELSE
                           lnGTax3p = lnGTax3p + m.ngastax3
                        ENDIF
                     ENDIF

                     IF m.ngastax4 # 0
                        IF NOT m.lsev4g
                           IF NOT m.lDirGasPurch
                              lnGTax4 = lnGTax4 + m.ngastax4
                           ELSE
                              IF NOT INLIST(m.cdirect, 'B', 'G')
                                 lnGTax4 = lnGTax4 + m.ngastax4
                              ENDIF
                           ENDIF
                        ELSE
                           lnGTax4p = lnGTax4p + m.ngastax4
                        ENDIF
                     ENDIF

*  Post Other Product Taxes
                     IF m.nOthTax1 # 0
                        IF NOT m.lsev2p
                           lnPTax1 = lnPTax1 + m.nOthTax1
                        ELSE
                           lnPTax1p = lnPTax1p + m.nOthTax1
                        ENDIF
                     ENDIF

                     IF m.nOthTax2 # 0
                        IF NOT m.lsev2p
                           lnPTax2 = lnPTax2 + m.nOthTax2
                        ELSE
                           lnPTax2p = lnPTax2p + m.nOthTax2
                        ENDIF
                     ENDIF

                     IF m.nOthTax3 # 0
                        IF NOT m.lsev3p
                           lnPTax3 = lnPTax3 + m.nOthTax3
                        ELSE
                           lnPTax3p = lnPTax3p + m.nOthTax3
                        ENDIF
                     ENDIF

                     IF m.nOthTax4 # 0
                        IF NOT m.lsev4p
                           lnPTax4 = lnPTax4 + m.nOthTax4
                        ELSE
                           lnPTax4p = lnPTax4p + m.nOthTax4
                        ENDIF
                     ENDIF

*  Post compression and gathering
                     IF m.nCompress # 0
                        lnCompress = lnCompress + m.nCompress
                     ENDIF

                     IF m.nGather # 0
                        lnGathering = lnGathering + m.nGather
                     ENDIF

*  Post marketing expenses
                     IF m.nMKTGExp # 0
                        lnMarketing = lnMarketing + m.nMKTGExp
                     ENDIF

*  Process default class expenses
                     SELE roundtmp
                     LOCATE FOR cownerid == m.cownerid AND cdmbatch = THIS.cdmbatch
                     IF FOUND()
                        llRound = .T.
                     ELSE
                        llRound = .F.
                     ENDIF

                     llRoundIt = .F.
                     IF llSepClose AND llJIB
*  Do Nothing
                     ELSE
                        lnExpenses = lnExpenses + m.nexpense + m.ntotale1 + m.ntotale2 + m.ntotale3 + m.ntotale4 + m.ntotale5 + m.ntotalea + m.ntotaleb
                     ENDIF

*  Post Prior Period Deficits
                     IF m.ctypeinv = 'X' AND m.nnetcheck # 0
                        lnDeficit = lnDeficit + m.nnetcheck
                     ENDIF

*  Post Prior Period Minimums
                     IF m.ctypeinv = 'M' AND m.nnetcheck # 0
                        lnMinimum = lnMinimum - m.nnetcheck
                     ENDIF

                     lnCheck = lnCheck + m.nnetcheck
                  ENDSCAN && Invtmp

                  m.noilrev    = lnOilRev
                  m.ngasrev    = lnGasRev
                  m.nothrev    = lnOthRev
                  m.noiltax    = lnOilTax
                  m.ngastax    = lnGasTax
                  m.nOTax1     = lnOTax1
                  m.nOTax2     = lnOTax2
                  m.nOTax3     = lnOTax3
                  m.nOTax4     = lnOTax4
                  m.nGTax1     = lnGTax1
                  m.nGTax2     = lnGTax2
                  m.nGTax3     = lnGTax3
                  m.nGTax4     = lnGTax4
                  m.nPTax1     = lnPTax1
                  m.nPTax1     = lnPTax1
                  m.nPTax2     = lnPTax2
                  m.nPTax3     = lnPTax3
                  m.nPTax4     = lnPTax4
                  m.nOTax1P    = lnOTax1p
                  m.nOTax2P    = lnOTax2p
                  m.nOTax3P    = lnOTax3p
                  m.nOTax4P    = lnOTax4p
                  m.nGTax1P    = lnGTax1p
                  m.nGTax2P    = lnGTax2p
                  m.nGTax3P    = lnGTax3p
                  m.nGTax4P    = lnGTax4p
                  m.nPTax1P    = lnPTax1p
                  m.nPTax1P    = lnPTax1p
                  m.nPTax2P    = lnPTax2p
                  m.nPTax3P    = lnPTax3p
                  m.nPTax4P    = lnPTax4p
                  m.nMarketing = lnMarketing
                  m.nCompress  = lnCompress
                  m.nGathering = lnGathering
                  m.nExpenses  = lnExpenses
                  m.nDeficit   = lnDeficit
                  m.nMinimum   = lnMinimum

                  INSERT INTO postown1 FROM MEMVAR
               ENDSCAN && Wells
            ENDSCAN  && Loop through posting owners (curPostOwns)

            SELECT  cWellID, ;
                    SUM(noilrev) AS noilrev, ;
                    SUM(ngasrev) AS ngasrev, ;
                    SUM(nothrev) AS nothrev, ;
                    SUM(noiltax) AS noiltax, ;
                    SUM(ngastax) AS ngastax, ;
                    SUM(nOTax1) AS nOTax1, ;
                    SUM(nOTax2) AS nOTax2, ;
                    SUM(nOTax3) AS nOTax3, ;
                    SUM(nOTax4) AS nOTax4, ;
                    SUM(nGTax1) AS nGTax1, ;
                    SUM(nGTax2) AS nGTax2, ;
                    SUM(nGTax3) AS nGTax3, ;
                    SUM(nGTax4) AS nGTax4, ;
                    SUM(nPTax1) AS nPTax1, ;
                    SUM(nPTax2) AS nPTax2, ;
                    SUM(nPTax3) AS nPTax3, ;
                    SUM(nPTax4) AS nPTax4, ;
                    SUM(nOTax1P) AS nOTax1P, ;
                    SUM(nOTax2P) AS nOTax2P, ;
                    SUM(nOTax3P) AS nOTax3P, ;
                    SUM(nOTax4P) AS nOTax4P, ;
                    SUM(nGTax1P) AS nGTax1P, ;
                    SUM(nGTax2P) AS nGTax2P, ;
                    SUM(nGTax3P) AS nGTax3P, ;
                    SUM(nGTax4P) AS nGTax4P, ;
                    SUM(nPTax1P) AS nPTax1P, ;
                    SUM(nPTax2P) AS nPTax2P, ;
                    SUM(nPTax3P) AS nPTax3P, ;
                    SUM(nPTax4P) AS nPTax4P, ;
                    SUM(nMarketing) AS nMarketing, ;
                    SUM(nCompress) AS nCompress, ;
                    SUM(nGathering) AS nGathering, ;
                    SUM(nExpenses) AS nExpenses, ;
                    SUM(nDeficit) AS nDeficit, ;
                    SUM(nMinimum) AS nMinimum ;
                FROM postown1 ;
                INTO CURSOR postown2 ;
                ORDER BY cWellID ;
                GROUP BY cWellID


            THIS.ogl.cdeptno    = lcDeptNo
            THIS.ogl.cReference = 'Owner Post'
            THIS.ogl.dGLDate    = tdCompanyPost

            SELECT postown2
            SCAN
               SCATTER MEMVAR
               THIS.ogl.cUnitNo = m.cWellID
               THIS.ogl.cID     = ""
               swselect('wells')
               SET ORDER TO cWellID
               IF SEEK(m.cWellID)
                  SCATTER FIELDS LIKE lSev* MEMVAR
                  m.nprocess     = nprocess
                  m.lDirOilPurch = lDirOilPurch
                  m.lDirGasPurch = lDirGasPurch
               ELSE
                  LOOP
               ENDIF
* Post Oil Revenue
               IF m.noilrev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('BBL')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Oil Revenue'

                  THIS.ogl.namount = m.noilrev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = m.noilrev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF
* Post Gas Revenue
               IF m.ngasrev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('MCF')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource = 'Gas Revenue'

                  THIS.ogl.namount = m.ngasrev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = m.ngasrev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF
* Post Other revenue
               IF m.nothrev # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF
                  m.cRevSource     = 'Other Revenue'
                  THIS.ogl.namount = m.nothrev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = m.nothrev * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()

               ENDIF
* Post Oil Taxes
               IF m.nOTax1 # 0 OR m.nOTax1P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Oil Tax 1'

                  THIS.ogl.namount = m.nOTax1 + m.nOTax1P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nOTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nOTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax1P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nOTax1P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nOTax2 # 0 OR m.nOTax2P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Oil Tax 2'

                  THIS.ogl.namount = m.nOTax2 + m.nOTax2P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nOTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nOTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax2P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nOTax2P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nOTax3 # 0 OR m.nOTax3P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Oil Tax 3'

                  THIS.ogl.namount = m.nOTax3 + m.nOTax3P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nOTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nOTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax3P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nOTax3P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nOTax4 # 0 OR m.nOTax4P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('OTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Oil Tax 4'

                  THIS.ogl.namount = m.nOTax4 + m.nOTax4P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nOTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nOTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nOTax4P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nOTax4P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

* Post Gas Taxes
               IF m.nGTax1 # 0 OR m.nGTax1P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Gas Tax 1'

                  THIS.ogl.namount = m.nGTax1 + m.nGTax1P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nGTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nGTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax1P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nGTax1P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax2 # 0 OR m.nGTax2P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Gas Tax 2'

                  THIS.ogl.namount = m.nGTax2 + m.nGTax2P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nGTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nGTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax2P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nGTax2P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax3 # 0 OR m.nGTax3P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Gas Tax 3'

                  THIS.ogl.namount = m.nGTax3 + m.nGTax3P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nGTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nGTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax3P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nGTax3P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nGTax4 # 0 OR m.nGTax4P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Gas Tax 4'

                  THIS.ogl.namount = m.nGTax4 + m.nGTax4P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nGTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nGTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nGTax4P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nGTax4P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

* Post Other Taxes
               IF m.nPTax1 # 0 OR m.nPTax1P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX1')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Other Tax 1'

                  THIS.ogl.namount = m.nPTax1 + m.nPTax1P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nPTax1 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nPTax1 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nPTax1P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nPTax1P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nPTax2 # 0 OR m.nPTax2P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX2')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Other Tax 2'

                  THIS.ogl.namount = m.nPTax2 + m.nPTax2P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nPTax2 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nPTax2 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nPTax2P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nPTax2P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nPTax3 # 0 OR m.nPTax3P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX3')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Other Tax 3'

                  THIS.ogl.namount = m.nPTax3 + m.nPTax3P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nPTax3 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nPTax3 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nPTax3P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nPTax3P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nPTax4 # 0 OR m.nPTax4P # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('PTAX4')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Other Tax 4'

                  THIS.ogl.namount = m.nPTax4 + m.nPTax4P
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()

                  IF m.nPTax4 # 0
*  Post tax liability
                     THIS.ogl.namount = m.nPTax4 * -1
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                     THIS.ogl.UpdateBatch()
                  ENDIF
                  IF m.nPTax4P # 0
* Post to revenue clearing
                     THIS.ogl.namount = m.nPTax4P * -1
                     THIS.ogl.cAcctNo = THIS.crevclear
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF

               IF m.nCompress # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('COMP')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('COMP')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Compression'
                     ENDIF
                  ENDIF

                  m.cRevSource = 'Compression'

                  THIS.ogl.namount = m.nCompress
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = m.nCompress * -1
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nGathering # 0
                  swselect('revcat')
                  SET ORDER TO cRevType
                  IF SEEK('GATH')
                     lcAcctD = cdracctno
                     lcAcctC = ccracctnow
                     IF EMPTY(lcAcctC)
                        lcAcctC = lcSuspense
                     ENDIF
                  ELSE
                     swselect('expcat')
                     SET ORDER TO ccatcode   && CCATCODE
                     IF SEEK('GATH')
                        lcAcctC = cdraccto
                        lcDesc  = ccateg
                     ELSE
                        lcAcctC = lcSuspense
                        lcDesc  = 'Compression'
                     ENDIF
                  ENDIF

                  m.cRevSource = 'Gathering'

                  THIS.ogl.namount = m.nGathering
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.UpdateBatch()
*  Post revenue clearing entry
                  THIS.ogl.namount = m.nGathering * -1
                  THIS.ogl.cAcctNo = THIS.crevclear
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nMarketing # 0
                  STORE '' TO lcAcctD, lcAcctC
                  swselect('expcat')
                  SET ORDER TO ccatcode
                  IF SEEK('MKTG')
                     lcAcctD = cdraccto
                     lcDesc  = ccateg
                     IF EMPTY(lcAcctD)
                        lcAcctD = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctD = lcSuspense
                     lcDesc  = 'Unknown'
                  ENDIF
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcAcctD
                  THIS.ogl.namount = m.nMarketing
                  THIS.ogl.UpdateBatch()
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.namount = m.nMarketing * -1
                  THIS.ogl.UpdateBatch()
               ENDIF

               IF m.nExpenses # 0
                  STORE '' TO lcAcctD, lcAcctC
                  swselect('expcat')
                  SET ORDER TO ccatcode
                  IF SEEK('EXPS')
                     lcAcctD = cdraccto
                     lcDesc  = ccateg
                     IF EMPTY(lcAcctD)
                        lcAcctD = lcSuspense
                     ENDIF
                  ELSE
                     lcAcctD = lcSuspense
                     lcDesc  = 'Unknown - EXPS'
                  ENDIF

                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcAcctD
                  THIS.ogl.namount = m.nExpenses
                  THIS.ogl.UpdateBatch()
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.namount = m.nExpenses * -1
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDSCAN
            llReturn = THIS.ogl.ChkBalance()

         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'PostSummaryWell', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('PostSummaryWell', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE oSuspense_Access
*********************************
*To do: Modify this routine for the Access method

      IF TYPE('this.osuspense') # 'O' OR ISNULL(THIS.osuspense)
         THIS.osuspense = CREATEOBJECT('suspense')
* Create the suspense object
         THIS.osuspense.nrunno   = THIS.nrunno
         THIS.osuspense.crunyear = THIS.crunyear
         THIS.osuspense.cgroup   = THIS.cgroup
         THIS.osuspense.lNewRun  = THIS.lNewRun
      ENDIF

      RETURN THIS.osuspense
   ENDPROC

*********************************
   PROCEDURE CalcRoundingTotals
*********************************

      llReturn = .T.

      TRY
         SELECT invtmp

         REPLACE invtmp.nIncome WITH invtmp.ngasrev + invtmp.noilrev +  ;
                   invtmp.ntrprev + invtmp.nmiscrev1 + invtmp.nmiscrev2 + invtmp.nothrev,  ;
                 invtmp.nsevtaxes WITH invtmp.nOthTax1 + invtmp.nOthTax2 +  ;
                   invtmp.nOthTax3 + invtmp.nOthTax4 + invtmp.noiltax1 + invtmp.noiltax2 + ;
                   invtmp.noiltax3 + invtmp.noiltax4 + invtmp.ngastax1 + invtmp.ngastax2 + ;
                   invtmp.ngastax3 + invtmp.ngastax4,  ;
                 invtmp.nnetcheck WITH (invtmp.ngasrev + invtmp.noilrev +  ;
                     invtmp.ntrprev + invtmp.nmiscrev1 + invtmp.nmiscrev2 + invtmp.nothrev) - ;
                   (invtmp.nexpense + invtmp.ntotale1 + invtmp.ntotale2 + invtmp.ntotale3 + invtmp.ntotale4 + ;
                     invtmp.ntotale5 + invtmp.ntotalea + invtmp.ntotaleb + invtmp.nPlugExp) - (invtmp.nOthTax1 + invtmp.nOthTax2 + invtmp.nOthTax3 + invtmp.nOthTax4 + ;
                     invtmp.ngastax1 + invtmp.ngastax2 + invtmp.ngastax3 + invtmp.ngastax4 + invtmp.noiltax1 + invtmp.noiltax2 + invtmp.noiltax3 + invtmp.noiltax4) - ;
                   (invtmp.nCompress + invtmp.nGather + invtmp.nMKTGExp + nbackwith + ntaxwith)

         IF invtmp.cdirect = 'G' OR  invtmp.cdirect = 'B'  &&  If direct-pay gas, subtract that money from nnetcheck
            REPLACE invtmp.nnetcheck WITH invtmp.nnetcheck - invtmp.ngasrev
         ENDIF
         IF invtmp.cdirect = 'O' OR invtmp.cdirect = 'B'  &&  If direct-pay oil, subtract that money from nnetcheck
            REPLACE invtmp.nnetcheck WITH invtmp.nnetcheck - invtmp.noilrev
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'CalcRoundingTot', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('CalcRoundingTot', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY


      RETURN llReturn
   ENDPROC

*-- Post owners marked to post.
*********************************
   PROCEDURE PostOperator
*********************************
      LPARAMETERS tcTable, tcOwnerid, tcOwnerName
*
* Posts the entries for the given owner. The owner passed in was marked to post to the G/L
* in their owner record in the Investor table.
*
      LOCAL llSepClose, llExpSum, llSuspense, lcRevClear, lcExpClear, llNoPostDM, llRound, lcDeptNo
      LOCAL lcSuspense, lnOilRev, lnGasRev, lnTrpRev, lnMi1Rev, lnMi2Rev, lnDebits, lnOwner
      LOCAL lnCredits, lnDeficit, lnOilTax, lnGasTax, lnRunNo, lcRunYear, tdPostDate, tdCompanyPost
      LOCAL m.cTaxAcct1, m.cTaxAcct2, m.cTaxAcct3, m.cTaxAcct4, m.cPlugAcct
      LOCAL lDirGasPurch, lDirOilPurch, lcAcctC, lcAcctD, lcDesc, lcGasAcctC, lcGasAcctD, lcName
      LOCAL lcOilAcctC, lcOilAcctD, lcOwnerID, lcTaxWH, llIF, llJIB, llReturn, llRoundIt, lnMin, lnSuspBal
      LOCAL lnTotal, loError
      LOCAL cBackWith, cBatch, cperiod, cRevSource, cTaxAcct1, cTaxAcct2, cTaxAcct3, cTaxAcct4, cyear
      LOCAL ccatcode, ccateg, cdracct, cexpclass, nCredits, nDebits, namount, nprocess

      llReturn = .T.

      TRY

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF

         lnOwner = 0

*  Get the suspense account from glopt
         swselect('glopt')
         GO TOP
         lcSuspense = cSuspense
         lcRevClear = crevclear
         lcExpClear = cexpclear
         llNoPostDM = lDMNoPost

         IF m.goApp.lAMVersion
            lcDeptNo = THIS.oOptions.cdeptno
         ELSE
            lcDeptNo = ''
         ENDIF

         m.cTaxAcct1 = THIS.oOptions.cTaxAcct1
         m.cTaxAcct2 = THIS.oOptions.cTaxAcct2
         m.cTaxAcct3 = THIS.oOptions.cTaxAcct3
         m.cTaxAcct4 = THIS.oOptions.cTaxAcct4

         IF EMPTY(lcSuspense)
            lcSuspense = '999999'
         ENDIF

         THIS.ogl.nDebits  = 0
         THIS.ogl.nCredits = 0

*  Set the posting dates
         IF THIS.lAdvPosting
            tdCompanyPost = THIS.dCompanyShare
         ELSE
            tdCompanyPost = THIS.dCheckDate
         ENDIF

* Determine if we're posting suspense
         llSuspense = UPPER(tcTable) = 'TSUSPENSE'

         IF llSuspense
* Setup the suspense object
            THIS.osuspense.crunyear  = THIS.cnewrunyear
            THIS.osuspense.nrunno    = THIS.nnewrunno
            THIS.osuspense.cgroup    = THIS.cgroup
            THIS.osuspense.lClosing  = THIS.lclose
            THIS.osuspense.dacctdate = THIS.dacctdate

* Get the owner's suspense balance
            lnSuspBal = THIS.osuspense.Owner_Suspense_Balance(.T., tcOwnerid, THIS.cgroup)

* The owner has no suspense, so no posting is needed
            IF lnSuspBal = 0
               EXIT
            ENDIF

            IF lnSuspBal > 0
               lcRevClear = THIS.oOptions.cMinAcct
               lcExpClear = THIS.oOptions.cMinAcct
            ELSE
               lcRevClear = THIS.oOptions.cDefAcct
               lcExpClear = THIS.oOptions.cDefAcct
            ENDIF
         ENDIF

* Force the flag for separate close for revenue and jibs
         llSepClose = .T.

* Get the option for summing expenses
         llExpSum    = THIS.oOptions.lexpsum

* Get the batch number for this posting
         THIS.ogl.cBatch = GetNextPK('BATCH')

* Set the post date for this entry
         THIS.ogl.dGLDate = tdCompanyPost

* Initialize totaling variables
         STORE 0 TO lnOilRev, lnGasRev, lnTrpRev, lnMi1Rev, lnMi2Rev, lnDebits, lnCredits, ;
            lnMin, lnDeficit, lnOilTax, lnGasTax

* Setup the posting accounts if the operator has tax withholding
         m.cBackWith = THIS.oOptions.cBackAcct

         swselect('expcat')
         SET ORDER TO ccatcode
         IF SEEK('TXWH')
            lcTaxWH = cdraccto
         ELSE
            lcTaxWH = lcSuspense
         ENDIF

* Scan through the table (either invtmp or tsuspense) to post the entries
         SELECT (tcTable)
         SET ORDER TO invwell
         SCAN FOR cownerid = tcOwnerid AND nrunno = THIS.nrunno AND crectype = 'R'
            SCATTER MEMVAR

            IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
               llReturn          = .F.
               IF NOT m.goApp.CancelMsg()
                  THIS.lCanceled = .T.
                  EXIT
               ENDIF
            ENDIF

            IF m.nrunno_in # 0
               lnRunNo   = m.nrunno_in
               lcRunYear = m.crunyear_in
* Set the suspense account to remove
* suspense from
               IF m.nnetcheck > 0
                  lcRevClear = THIS.oOptions.cMinAcct
                  lcExpClear = THIS.oOptions.cMinAcct
               ELSE
                  lcRevClear = THIS.oOptions.cDefAcct
                  lcExpClear = THIS.oOptions.cDefAcct
               ENDIF
            ELSE
               lnRunNo    = m.nrunno
               lcRunYear  = m.crunyear
               lcRevClear = glopt.crevclear
               lcExpClear = glopt.cexpclear
            ENDIF

            llJIB  = lJIB
            lcName = tcOwnerid

            swselect('wells')
            IF SEEK(m.cWellID)
               SCATTER FIELDS LIKE lSev* MEMVAR
               m.nprocess     = nprocess
               m.lDirOilPurch = lDirOilPurch
               m.lDirGasPurch = lDirGasPurch
            ELSE
               m.nprocess = 1
               STORE .F. TO m.lDirOilPurch, m.lDirGasPurch
            ENDIF

*  Post Oil Income
            IF m.noilrev # 0 AND NOT INLIST(m.cdirect, 'O', 'B')
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('BBL')
                  lcAcctD = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctC = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctC = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctC = ccracctnow
                     OTHERWISE
                        lcAcctC = ccracctnow
                  ENDCASE

                  IF EMPTY(lcAcctC)
                     lcAcctC = lcSuspense
                  ENDIF

                  STORE 0 TO m.nDebits, m.nCredits

                  m.cRevSource = 'Oil Revenue'

                  THIS.ogl.namount = m.noilrev * -1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = m.noilrev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF
            ENDIF

*  Post Gas Income
            IF m.ngasrev # 0 AND NOT INLIST(m.cdirect, 'G', 'B')
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('MCF')
                  lcAcctD = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctC = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctC = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctC = ccracctnow
                     OTHERWISE
                        lcAcctC = ccracctnow
                  ENDCASE

                  IF EMPTY(lcAcctC)
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Gas Revenue'

                  THIS.ogl.namount = m.ngasrev * -1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = m.ngasrev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Other Income
            IF m.nothrev # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('OTH')
                  lcAcctD = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctC = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctC = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctC = ccracctnow
                     OTHERWISE
                        lcAcctC = ccracctnow
                  ENDCASE

                  IF EMPTY(lcAcctC)
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Other Product Income'

                  THIS.ogl.namount = m.nothrev * -1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = m.nothrev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Trp Income
            IF m.ntrprev # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('TRANS')
                  lcAcctD = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctC = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctC = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctC = ccracctnow
                     OTHERWISE
                        lcAcctC = ccracctnow
                  ENDCASE

                  IF EMPTY(lcAcctC)
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Trans Income'

                  THIS.ogl.namount = m.ntrprev * -1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = m.ntrprev
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Misc 1 Income
            IF m.nmiscrev1 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('MISC1')
                  lcAcctD = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctC = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctC = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctC = ccracctnow
                     OTHERWISE
                        lcAcctC = ccracctnow
                  ENDCASE

                  IF EMPTY(lcAcctC)
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Misc Income 1'

                  THIS.ogl.namount = m.nmiscrev1 * -1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post revenue clearing entry
                  THIS.ogl.namount = m.nmiscrev1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Misc 2 Income
            IF m.nmiscrev2 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('MISC2')
                  lcAcctD = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctC = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctC = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctC = ccracctnow
                     OTHERWISE
                        lcAcctC = ccracctnow
                  ENDCASE

                  IF EMPTY(lcAcctC)
                     lcAcctC = lcSuspense
                  ENDIF

                  m.cRevSource = 'Misc Income 2'

                  THIS.ogl.namount = m.nmiscrev2 * -1
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcAcctC
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()


*  Post revenue clearing entry
                  THIS.ogl.namount = m.nmiscrev2
                  THIS.ogl.cDesc   = m.cRevSource
                  THIS.ogl.cAcctNo = lcRevClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Oil Taxes
            IF m.noiltax1 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('OTAX1')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcOilAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcOilAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcOilAcctD = ccracctnow
                     OTHERWISE
                        lcOilAcctD = ccracctnow
                  ENDCASE
                  lcDesc     = crevdesc
                  IF EMPTY(lcOilAcctD)
                     lcOilAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.noiltax1
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcOilAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  IF NOT  INLIST(m.cdirect, 'O', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post payble entry or clearing entry
                  THIS.ogl.namount = m.noiltax1 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev1o
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  IF NOT  INLIST(m.cdirect, 'O', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF
            ENDIF

            IF m.noiltax2 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('OTAX2')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcOilAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcOilAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcOilAcctD = ccracctnow
                     OTHERWISE
                        lcOilAcctD = ccracctnow
                  ENDCASE
                  lcOilAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcOilAcctD)
                     lcOilAcctD = lcSuspense
                  ENDIF


                  THIS.ogl.namount = m.noiltax2
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcOilAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo

                  IF NOT  INLIST(m.cdirect, 'O', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.noiltax2 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev2o
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo

                  IF NOT  INLIST(m.cdirect, 'O', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF
            ENDIF
            IF m.noiltax3 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('OTAX3')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcOilAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcOilAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcOilAcctD = ccracctnow
                     OTHERWISE
                        lcOilAcctD = ccracctnow
                  ENDCASE
                  lcOilAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcOilAcctD)
                     lcOilAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.noiltax3
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcOilAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.noiltax3 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev3o
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF
            IF m.noiltax4 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('OTAX4')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcOilAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcOilAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcOilAcctD = ccracctnow
                     OTHERWISE
                        lcOilAcctD = ccracctnow
                  ENDCASE
                  lcOilAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcOilAcctD)
                     lcOilAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.noiltax4
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcOilAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.noiltax4 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev4o
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Gas Taxes
            IF m.ngastax1 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('GTAX1')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.ngastax1
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo

                  IF NOT  INLIST(m.cdirect, 'G', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.ngastax1 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev1g
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo

                  IF NOT  INLIST(m.cdirect, 'G', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF
               ENDIF
            ENDIF
            IF m.ngastax2 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('GTAX2')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.ngastax2
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo

                  IF NOT  INLIST(m.cdirect, 'G', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.ngastax2 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev2g
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo

                  IF NOT  INLIST(m.cdirect, 'G', 'B')
                     THIS.ogl.UpdateBatch()
                  ELSE
                     THIS.ogl.UpdateBatch()
                  ENDIF

               ENDIF
            ENDIF

            IF m.ngastax3 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('GTAX3')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.ngastax3
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.ngastax3 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev3g
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

            IF m.ngastax4 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('GTAX4')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.ngastax4
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.ngastax4 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev4g
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post Other Product Taxes
            IF m.nOthTax1 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('PTAX1')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.nOthTax1
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.nOthTax1 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev1p
                     THIS.ogl.cAcctNo = m.cTaxAcct1
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF
            IF m.nOthTax2 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('PTAX2')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.nOthTax2
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.nOthTax2 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev2p
                     THIS.ogl.cAcctNo = m.cTaxAcct2
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF
            IF m.nOthTax3 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('PTAX3')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.nOthTax3
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.nOthTax3 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev3p
                     THIS.ogl.cAcctNo = m.cTaxAcct3
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF
            IF m.nOthTax4 # 0
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('PTAX4')
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcGasAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcGasAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcGasAcctD = ccracctnow
                     OTHERWISE
                        lcGasAcctD = ccracctnow
                  ENDCASE
                  lcGasAcctC = cdracctno
                  lcDesc     = crevdesc
                  IF EMPTY(lcGasAcctD)
                     lcGasAcctD = lcSuspense
                  ENDIF

                  THIS.ogl.namount = m.nOthTax4
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcGasAcctD
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

*  Post payable or revenue clearing entry
                  THIS.ogl.namount = m.nOthTax4 * -1
                  THIS.ogl.cDesc   = lcDesc
                  IF NOT m.lsev4p
                     THIS.ogl.cAcctNo = m.cTaxAcct4
                  ELSE
                     THIS.ogl.cAcctNo = lcRevClear
                  ENDIF
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Post compression and gathering
            IF m.nCompress # 0
               STORE '' TO lcAcctD, lcAcctC
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('COMP')
                  lcAcctC = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctD = ccracctnow
                     OTHERWISE
                        lcAcctD = ccracctnow
                  ENDCASE
                  lcDesc     = crevdesc
               ELSE
                  swselect('expcat')
                  SET ORDER TO ccatcode   && CCATCODE
                  IF SEEK('COMP')
                     lcAcctD = cdraccto
                     lcDesc  = ccateg
                  ELSE
                     lcAcctD = lcSuspense
                     lcDesc  = 'Compression'
                  ENDIF
               ENDIF

               IF EMPTY(lcAcctD)
                  lcAcctD = lcSuspense
               ENDIF

               THIS.ogl.cDesc   = lcDesc
               THIS.ogl.cAcctNo = lcAcctD
               THIS.ogl.namount = m.nCompress
               THIS.ogl.UpdateBatch()
               THIS.ogl.cAcctNo = lcExpClear
               THIS.ogl.namount = m.nCompress * -1
               THIS.ogl.UpdateBatch()
            ENDIF

            IF m.nGather # 0
               STORE '' TO lcAcctD, lcAcctC
               swselect('revcat')
               SET ORDER TO cRevType
               IF SEEK('GATH')
                  lcAcctC = cdracctno
                  DO CASE
                     CASE m.ctypeinv = 'L'
                        lcAcctD = ccracctnol
                     CASE m.ctypeinv = 'O'
                        lcAcctD = ccracctnoo
                     CASE m.ctypeinv = 'W'
                        lcAcctD = ccracctnow
                     OTHERWISE
                        lcAcctD = ccracctnow
                  ENDCASE
                  lcDesc     = crevdesc
               ELSE
                  swselect('expcat')
                  SET ORDER TO ccatcode   && CCATCODE
                  IF SEEK('GATH')
                     lcAcctD = cdraccto
                     lcDesc  = ccateg
                  ELSE
                     lcAcctD = lcSuspense
                     lcDesc  = 'Gathering'
                  ENDIF
               ENDIF
               IF EMPTY(lcAcctD)
                  lcAcctD = lcSuspense
               ENDIF

               THIS.ogl.cDesc   = lcDesc
               THIS.ogl.cAcctNo = lcAcctD
               THIS.ogl.namount = m.nGather
*      THIS.oGL.cReference = 'Gathering'
               THIS.ogl.UpdateBatch()
               THIS.ogl.cAcctNo = lcExpClear
               THIS.ogl.namount = m.nGather * -1
               THIS.ogl.UpdateBatch()
            ENDIF

*  Post marketing expenses
            IF m.nMKTGExp # 0
               STORE '' TO lcAcctD, lcAcctC
               swselect('expcat')
               SET ORDER TO ccatcode
               IF SEEK('MKTG')
                  lcAcctD = cdraccto
                  lcDesc  = ccateg
                  IF EMPTY(lcAcctD)
                     lcAcctD = lcSuspense
                  ENDIF
                  THIS.ogl.cDesc   = lcDesc
                  THIS.ogl.cAcctNo = lcAcctD
                  THIS.ogl.namount = m.nMKTGExp
                  THIS.ogl.UpdateBatch()
                  THIS.ogl.cAcctNo = THIS.cexpclear
                  THIS.ogl.namount = m.nMKTGExp * -1
                  THIS.ogl.UpdateBatch()
               ENDIF
            ENDIF

*  Process default class expenses
            llRoundIt = .F.
            llRound   = .F.
            IF llSepClose AND llJIB
*  Do Nothing
            ELSE
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.nexpense # 0"
               ELSE
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.nexpense # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = '0' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH', 'PLUG')

                     swselect('roundtmp')
                     LOCATE FOR cownerid == m.cownerid ;
                        AND cWellID == m.cWellID ;
                        AND NOT lused ;
                        AND cdmbatch == THIS.cdmbatch
                     IF FOUND()
                        llRound = .T.
                        REPL lused WITH .T.
                     ELSE
                        llRound = .F.
                     ENDIF

                     SELE expense

                     m.ccateg    = ccateg
                     m.ccatcode  = ccatcode
                     m.namount   = namount
                     lcOwnerID   = cownerid
                     m.cexpclass = cexpclass
                     m.cyear     = cyear
                     m.cperiod   = cperiod

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct   = cdraccto
                     ELSE
                        m.cdracct   = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nworkint / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        swselect('roundtmp')
                        m.namount = m.namount + roundtmp.nexpense
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Operating Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process class 1 expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale1 # 0"
               ELSE
                  llIF = "m.nTotale1 <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale1 # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = '1' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')
                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nintclass1 / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotale1
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Class 1 Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process class 2 expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale2 # 0"
               ELSE
                  llIF = "m.nTotale2 <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale2 # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = '2' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')

                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nintclass2 / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotale2
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Class 2 Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process class 3 expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale3 # 0"
               ELSE
                  llIF = "m.nTotale3 <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale3 # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = '3' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')
                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nintclass3 / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotale3
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Class 3 Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process class 4 expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale4 # 0"
               ELSE
                  llIF = "m.nTotale4 <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale4 # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = '4' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')
                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nintclass4 / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotale4
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Class 4 Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process class 5 expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale5 # 0"
               ELSE
                  llIF = "m.nTotale5 <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotale5 # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = '5' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')

                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nintclass5 / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotale5
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Class 5 Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process ACP expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotalea # 0"
               ELSE
                  llIF = "m.nTotaleA <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotalea # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = 'A' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')

                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nacpint / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotalea
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'After Casing Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process BCP expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotaleb # 0"
               ELSE
                  llIF = "m.nTotaleB <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.ntotaleb # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = 'B' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')
                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nbcpint / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotaleb
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Before Casing Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

*  Process Plugging expenses
               llRoundIt = .F.
               IF llExpSum
                  llIF = "NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.nplugexp # 0"
               ELSE
                  llIF = "m.nPlugExp <> 0 AND NOT m.lJIB AND INLIST(m.cTypeInv,'L','O','W') and m.nplugexp # 0"
               ENDIF
               IF &llIF
                  lnTotal = 0
                  swselect('expense')
                  SCAN FOR cyear + cperiod = m.hyear + m.hperiod ;
                        AND nRunNoRev = lnRunNo ;
                        AND cRunYearRev = lcRunYear ;
                        AND cWellID = m.cWellID ;
                        AND cexpclass = 'P' ;
                        AND NOT INLIST(ccatcode, 'MKTG', 'COMP', 'GATH')
                     m.ccateg   = ccateg
                     m.ccatcode = ccatcode
                     m.namount  = namount
                     lcOwnerID  = cownerid

*  Get the account numbers to post this expense to
                     swselect('expcat')
                     SET ORDER TO ccatcode
                     IF SEEK(m.ccatcode)
                        m.cdracct  = cdraccto
                     ELSE
                        m.cdracct = lcSuspense
                     ENDIF

*  If the expcat field is blank, use the suspense account from glopt
                     IF EMPTY(ALLT(m.cdracct))
                        m.cdracct = lcSuspense
                     ENDIF

*  Post Expenses
                     IF NOT EMPTY(lcOwnerID)           && Check for one-man items
                        IF lcOwnerID # m.cownerid
                           LOOP
                        ENDIF
                     ELSE
                        m.namount = swround(m.namount * (m.nbcpint / 100), 2)
                     ENDIF

                     IF llRound AND NOT llRoundIt
                        m.namount = m.namount + roundtmp.ntotaleb
                        llRoundIt = .T.
                     ENDIF

                     lnTotal = lnTotal + m.namount

                     THIS.ogl.namount = m.namount
                     THIS.ogl.cDesc   = m.ccateg
                     THIS.ogl.cAcctNo = m.cdracct
                     THIS.ogl.cUnitNo = m.cWellID
                     THIS.ogl.cID     = tcOwnerid
                     THIS.ogl.cdeptno = lcDeptNo
                     THIS.ogl.UpdateBatch()

                  ENDSCAN

*  Post expense clearing entry
                  THIS.ogl.namount = lnTotal * -1
                  THIS.ogl.cDesc   = 'Plugging Expenses: ' + m.hyear + '/' + m.hperiod
                  THIS.ogl.cAcctNo = lcExpClear
                  THIS.ogl.cUnitNo = m.cWellID
                  THIS.ogl.cID     = tcOwnerid
                  THIS.ogl.cdeptno = lcDeptNo
                  THIS.ogl.UpdateBatch()

               ENDIF

            ENDIF

*  Post Prior Period Deficits
            IF m.ctypeinv = 'X' AND m.nnetcheck # 0
               THIS.ogl.cAcctNo    = m.cDefAcct
               THIS.ogl.cDesc      = 'Prior Deficit Covered'
               THIS.ogl.namount    = m.nnetcheck
               THIS.ogl.cReference = 'Prior Def'
               THIS.ogl.UpdateBatch()

               THIS.ogl.cAcctNo    = lcSuspense
               THIS.ogl.cDesc      = 'Prior Deficit Covered'
               THIS.ogl.namount    = m.nnetcheck * -1
               THIS.ogl.cReference = 'Prior Def'
               THIS.ogl.UpdateBatch()

            ENDIF

*  Post Prior Period Minimums
            IF m.ctypeinv = 'M' AND m.nnetcheck # 0
               THIS.ogl.cAcctNo    = m.cMinAcct
               THIS.ogl.cDesc      = 'Prior Period Minimums'
               THIS.ogl.namount    = m.nnetcheck
               THIS.ogl.cReference = 'Prior Min'
               THIS.ogl.UpdateBatch()

               THIS.ogl.cAcctNo    = lcSuspense
               THIS.ogl.namount    = m.nnetcheck * -1
               THIS.ogl.cReference = 'Prior Min'
               THIS.ogl.UpdateBatch()
            ENDIF

*  Post Tax Withholding
            IF m.ntaxwith # 0
               THIS.ogl.cAcctNo    = m.cBackWith
               THIS.ogl.cDesc      = 'State Tax W/H'
               THIS.ogl.namount    = m.ntaxwith * -1
               THIS.ogl.cReference = 'Tax W/H'
               THIS.ogl.UpdateBatch()

               THIS.ogl.cAcctNo    = lcTaxWH
               THIS.ogl.namount    = m.ntaxwith
               THIS.ogl.cReference = 'Tax W/H'
               THIS.ogl.UpdateBatch()

            ENDIF

            lnOwner = lnOwner + m.nnetcheck

         ENDSCAN

         llReturn = THIS.ogl.ChkBalance()

         IF NOT llReturn
            TRY
               IF NOT FILE(m.goApp.cdatafilepath + 'outbal.dbf')
                  CREATE TABLE (m.goApp.cdatafilepath + 'outbal') FREE (cBatch  c(8), cownerid  c(10))
               ENDIF
               IF NOT USED('outbal')
                  USE (m.goApp.cdatafilepath + 'outbal') IN 0
               ENDIF
               m.cBatch = THIS.ogl.cBatch
               INSERT INTO outbal FROM MEMVAR
            CATCH TO loError
               llReturn = .F.
               DO errorlog WITH 'PostOperator', loError.LINENO, 'FormName', loError.ERRORNO, loError.MESSAGE, ' ', loError
               ErrorMessageText()
            ENDTRY
            EXIT
         ENDIF

      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'PostOperator', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('PostOperator', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE JIB_or_Net
*********************************
      LPARA tcwellid
      LOCAL lcType
      LOCAL lcReturn, loError

      lcReturn = 'N'

      TRY
         lcType = 'N'

* Look in the DOI to see if the well has all (J)IB owners, all (N)ET owners
* or a mixture of (B)oth
         SELE wellinv
         LOCATE FOR cWellID == tcwellid AND ctypeinv = 'W' AND lJIB
         IF NOT FOUND()
* Only NET owners in the well
            lcReturn = 'N'
         ELSE
            LOCATE FOR cWellID == tcwellid AND ctypeinv = 'W' AND lJIB = .F.
            IF FOUND()
* Both JIB and NET owners found
               lcType = 'B'
            ELSE
* Only JIB owners found
               lcType = 'J'
            ENDIF
         ENDIF
         lcReturn = lcType
      CATCH TO loError
         lcReturn = 'N'
         DO errorlog WITH 'JIB_OR_Net', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('JIB_OR_Net', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      RETURN lcReturn

   ENDPROC

*********************************
   PROCEDURE Print_Closing_Summary
*********************************

      LPARAMETERS tlReport, tcButton, tlSummaryOnly, tlByWell, tlExceptions
      LOCAL lnCountI, lnCountE, lnCount, loListener, lcLast, lcPrinterName, lcOldPrinter, llOtherReports
      LOCAL llReturn, loError
      PRIVATE llPrinted, lcTitle1, lcTitle2, m.cProcessor, m.cProducer, glGrpName, lcSelect, lcSortOrder

      llReturn = .T.

      TRY

         IF m.goApp.lCanceled
* Check to see if the Esc key was pressed to cancel the processing
            llReturn          = .F.
            IF NOT m.goApp.CancelMsg()
               THIS.lCanceled = .T.
               EXIT
            ENDIF
         ENDIF


         STORE '' TO lcPrinterName, lcOldPrinter
         STORE .F. TO llOtherReports, glGrpName

         m.cProducer  = m.goApp.cCompanyName
         m.cProcessor = ''
         lnCount      = 0
         STORE '' TO lcLast, lcSelect, lcSortOrder

         IF VARTYPE(tcButton) # 'C'
            tcButton = 'S'
         ENDIF

         IF NOT INLIST(tcButton, 'P', 'S')
            tcButton = 'S'
         ENDIF

* Calculate the closing summary report

         THIS.CalcSummary(tlReport)
         SELECT closetemp

         IF NOT tlByWell AND NOT tlSummaryOnly
* Calculate the closing suspense summary
            IF tlReport OR THIS.lRptSuspense
               THIS.PrintSuspense()
               SELECT audclose
               COUNT FOR NOT DELETED() TO lnSusp
               IF lnSusp > 0
                  lcLast         = 'audclose'
                  llOtherReports = .T.
               ENDIF
            ENDIF

* Unallocated revenue/expense report
            IF m.goApp.lAMVersion AND THIS.lRptUnalloc
               swselect('incsusp')
               COUNT FOR NOT DELETED() TO lnCountI
               swselect('expsusp')
               COUNT FOR NOT DELETED() TO lnCountE
               lnCount = lnCountI + lnCountE
               IF lnCount > 0
                  THIS.UnAllRpt()
                  SELECT unalltmp
                  IF RECCOUNT() > 0
                     lcLast         = 'unalltmp'
                     llOtherReports = .T.
                  ENDIF
               ENDIF
            ENDIF
         ENDIF

         IF tlByWell
            IF tlExceptions OR THIS.lRptWellExcpt
* This report is requested from other reports with exceptions
               SELECT closetemp
               IF ABS(nRevEntered - nRevAllocated) > 1 OR  ;  &&  Check for differences greater than $1.00, and then tell them the summary by well will now be printed
                  ABS(nExpEntered - nExpAllocated) > 1 OR  ;
                     ABS(nSevTaxWell - nSevTaxOwn) > 1
                  WAIT WIND NOWAIT 'Generating the closing summary by well reports...'
                  THIS.CalcSumByWell(.T.)
                  SELECT tempclose
                  IF RECCOUNT() > 0
                     lcLast         = 'tempclose'
                     llOtherReports = .T.
                  ELSE
                     IF tlReport
                        MESSAGEBOX('There were no wells that had differences greater that 50 cents.', 64, 'Closing Summary By Well')
                        llReturn = .F.
                        EXIT
                     ENDIF
                  ENDIF
               ELSE
                  IF tlReport
                     MESSAGEBOX('There were no wells that had differences greater that 50 cents.', 64, 'Closing Summary By Well')
                     EXIT
                  ENDIF
               ENDIF
            ELSE
* This report is requested from the other reports without exceptions
               WAIT WIND NOWAIT 'Generating the closing summary by well reports...'
               THIS.CalcSumByWell(.F.)
               SELECT tempclose
               IF RECCOUNT() > 0
                  lcLast         = 'tempclose'
                  llOtherReports = .T.
               ELSE
                  MESSAGEBOX('There were no records found for this report.', 48, 'Revenue Closing Summary by Well')
                  EXIT
               ENDIF
            ENDIF
         ENDIF

         IF NOT tlByWell AND (tlReport OR THIS.lRptRegister) AND NOT tlSummaryOnly
            THIS.CheckListing()
            SELECT tempchk
            IF RECCOUNT() > 0
               lcLast         = 'tempchk'
               llOtherReports = .T.
            ENDIF
         ENDIF

         lcTitle2 = 'For Run No ' + TRANSFORM(THIS.nrunno) + '/' + THIS.crunyear + ' Group ' + THIS.cgroup + ' Dated: ' + DTOC(THIS.dpostdate)

         SET REPORTBEHAVIOR 90
         LOCAL loPreviewContainer, loReportListener
         LOCAL loSession, lnRetval, loXFF, loPreview, loScripts
         loListener                = EVALUATE([xfrx("XFRX#LISTENER")])
         loUpdate                  = CREATEOBJECT("updatelistener")
         loUpdate.thermFormCaption = "Revenue Run Closing Summary Report in progress ..."
         loListener.successor      = loUpdate
         lnRetval                  = loListener.SetParams(, , , , , , "XFF") && no name = just in memory
         loListener.SetOtherParams("PRINT_BOOKMARKS", .T.)
         loListener.PRINTJOBNAME         = 'Revenue Run Closing Summary'
         loListener.CallEvaluateContents = 2
         IF lnRetval = 0

            llPrinted = .F.

            IF NOT tlByWell
* Only add this report if we're not doing the by well report
               SELECT closetemp
               GO TOP
* Closing Summary
               IF llOtherReports
                  REPORT FORM 'dmrcloser.frx' OBJECT loListener NOPAGEEJECT RANGE 1, 1
               ELSE
                  REPORT FORM 'dmrcloser.frx' OBJECT loListener NOPAGEEJECT  RANGE 1, 1
               ENDIF
            ENDIF

* Closing Exception Report
            IF (tlReport OR THIS.lRptWellExcpt)
               IF USED('tempclose')
                  SELECT tempclose
                  GO TOP
                  IF RECCOUNT() > 0
                     IF lcLast = 'tempclose'
                        REPORT FORM 'dmrclosew.frx' OBJECT loListener NOPAGEEJECT
                     ELSE
                        REPORT FORM 'dmrclosew.frx' OBJECT loListener NOPAGEEJECT
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF

* Suspense activity
            IF (tlReport OR THIS.lRptSuspense) AND NOT tlSummaryOnly AND NOT tlByWell
               IF USED('audclose')
                  SELECT audclose
                  GO TOP
                  IF RECCOUNT() > 0
                     IF lcLast = 'audclose'
                        REPORT FORM 'dmsuspcls.frx' OBJECT loListener NOPAGEEJECT
                     ELSE
                        REPORT FORM 'dmsuspcls.frx' OBJECT loListener NOPAGEEJECT
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF

* Unallocated revenue/expenses
            IF (tlReport OR THIS.lRptUnalloc) AND NOT tlSummaryOnly AND NOT tlByWell
               IF USED('unalltmp')
                  SELECT unalltmp
                  GO TOP
                  IF RECCOUNT() > 0
                     IF lcLast = 'unalltmp'
                        REPORT FORM 'dmrunall2.frx' OBJECT loListener NOPAGEEJECT
                     ELSE
                        REPORT FORM 'dmrunall2.frx' OBJECT loListener NOPAGEEJECT
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF
* Check Register
            IF (tlReport OR THIS.lRptRegister) AND NOT tlSummaryOnly AND NOT tlByWell
               IF USED('tempchk')
                  SELECT tempchk
                  GO TOP
                  IF RECCOUNT() > 0
                     IF lcLast = 'tempchk'
                        REPORT FORM 'csprechk.frx' OBJECT loListener NOPAGEEJECT
                     ELSE
                        REPORT FORM 'csprechk.frx' OBJECT loListener NOPAGEEJECT
                     ENDIF
                  ENDIF
               ENDIF
            ENDIF
            SET COVERAGE TO

            loListener.Finalize()

            loXFF                                 = loListener.oxfDocument
            loPreview                             = CREATEOBJECT("frmMPPreviewer")
            loPreview.iTool                       = 2
            loPreview.ShowStatus                  = .F.
            loPreview.oDisplayDefaults.ZoomFactor = 100
            loPreview.PreviewXFF(loXFF)
            loPreview.SHOW(0)
         ENDIF


      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'Print_Closing_Summary', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
         THIS.ERRORMESSAGE('Print_Closing_Summary', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE)
      ENDTRY

      THIS.CheckCancel()

      RETURN llReturn
   ENDPROC

******************************
   FUNCTION CheckCancel
******************************

      IF THIS.lCanceled
         IF VARTYPE(THIS.oprogress) = 'O'
            THIS.oprogress.CloseProgress()
            THIS.oprogress = .NULL.
         ENDIF
         THIS.lCanceled = .F.
      ENDIF

   ENDPROC

*********************************
   PROCEDURE PostTransferSummary
*********************************

*******************************************************
* Post suspense amounts that are moving between deficit
* and minimum suspense.
*******************************************************

* Get the posting dates
      IF THIS.lAdvPosting = .T.
         tdCompanyPost = THIS.dCompanyShare
         tdPostDate    = THIS.dCheckDate
      ELSE
         tdCompanyPost = THIS.dCheckDate
         tdPostDate    = THIS.dCheckDate
      ENDIF
      THIS.ogl.DMBatch  = THIS.cdmbatch
      THIS.ogl.cSource  = 'DM'
      THIS.ogl.nDebits  = 0
      THIS.ogl.nCredits = 0
      THIS.ogl.dGLDate  = THIS.dCheckDate

*
*  Get the suspense account from glopt
*
      swselect('glopt')
      lcSuspense = cSuspense
      IF EMPTY(lcSuspense)
         lcSuspense = '999999'
      ENDIF
      lcRevClear = crevclear
      lcExpClear = cexpclear
      llNoPostDM = lDMNoPost
*
*  Get the A/P account
*
      swselect('apopt')
      lcAPAcct = capacct

*
*  Set up the parameters used by processing in this method
*
      tcYear   = THIS.crunyear
      tcPeriod = THIS.cperiod
      tcGroup  = THIS.cgroup

****************************************************************
*   Get Deficit/Minimum posting accounts
****************************************************************
      m.cDefAcct  = THIS.oOptions.cDefAcct
      IF EMPTY(ALLT(m.cDefAcct))
         m.cDefAcct = lcSuspense
      ENDIF
      m.cMinAcct  = THIS.oOptions.cMinAcct
      IF EMPTY((m.cMinAcct))
         m.cMinAcct = lcSuspense
      ENDIF

      lnDefSwitch = 0
      lnMinSwitch = 0

      TRY
         THIS.oprogress.SetProgressMessage('Posting Owner Suspense Switching from One Type to Another')
* Post amounts transfering between deficit and minimum
* Get the amount of deficits that transferred
         lnDefSwitch = THIS.osuspense.GetBalTransfer('D')

* Get the amount of minimums that transferred
         lnMinSwitch = THIS.osuspense.GetBalTransfer('M') * -1

         THIS.ndeftransfer = lnDefSwitch
         THIS.nmintransfer = lnMinSwitch

         THIS.ogl.cidchec    = ''
         THIS.ogl.cReference = 'Run: R' + THIS.crunyear + '/' + ALLT(STR(THIS.nrunno)) + '/' + THIS.cgroup
         THIS.ogl.cID        = ''
         THIS.ogl.dGLDate    = tdPostDate
         THIS.ogl.cBatch     = GetNextPK('BATCH')
         THIS.ogl.cUnitNo    = ''
         THIS.ogl.cdeptno    = ''

* Post Deficits Switching To Minimums
         IF lnDefSwitch # 0
            THIS.ogl.cDesc   = 'Deficit To Minimum Switch'
            THIS.ogl.cAcctNo = m.cMinAcct
            THIS.ogl.namount = lnDefSwitch * -1
            THIS.ogl.UpdateBatch()

            THIS.ogl.cDesc   = 'Deficit To Minimum Switch'
            THIS.ogl.cAcctNo = m.cDefAcct
            THIS.ogl.namount = lnDefSwitch
            THIS.ogl.UpdateBatch()
         ENDIF

* Post Minimums Switching To Deficits
         IF lnMinSwitch # 0
            THIS.ogl.cDesc   = 'Minimum To Deficit Switch'
            THIS.ogl.cAcctNo = m.cMinAcct
            THIS.ogl.namount = lnMinSwitch * -1
            THIS.ogl.UpdateBatch()

            THIS.ogl.cDesc   = 'Minimum To Deficit Switch'
            THIS.ogl.cAcctNo = m.cDefAcct
            THIS.ogl.namount = lnMinSwitch
            THIS.ogl.UpdateBatch()
         ENDIF

         swselect('glmaster')
         TABLEUPDATE(.T., .T.)
      CATCH
      ENDTRY

*********************************
   PROCEDURE PluggingFund
*********************************

      IF NOT m.goApp.lPluggingModule
         RETURN
      ENDIF

      m.cdmbatch = THIS.cdmbatch
      m.nrunno   = THIS.nrunno
      m.crunyear = THIS.crunyear

      CREATE CURSOR PluggingFund ;
         (cWellID     c(10), ;
           cownerid    c(10), ;
           crunyear    c(4), ;
           nrunno      i, ;
           dacctdate   D, ;
           nplugging   N(12, 2), ;
           lManual     L, ;
           crectype    c(1), ;
           cdmbatch    c(8))

      SELECT  cownerid, ;
              cWellID,  ;
              hdate AS dacctdate, ;
              nPlugExp AS nplugging, ;
              crectype ;
          FROM disbhist WITH (BUFFERING =.T.) ;
          WHERE nrunno = m.nrunno AND ;
              crunyear = m.crunyear AND ;
              EMPTY(crunyear_in) AND ;
              nPlugExp # 0  AND ;
              crectype = 'R' ;
          INTO CURSOR tempd

      SELECT tempd
      SCAN
         SCATTER MEMVAR
         INSERT INTO PluggingFund FROM MEMVAR
      ENDSCAN

      SELECT  cownerid, ;
              cWellID,  ;
              hdate AS dacctdate, ;
              nPlugExp AS nplugging, ;
              crectype ;
          FROM suspense WITH (BUFFERING =.T.) ;
          WHERE nrunno_in = m.nrunno AND ;
              crunyear_in = m.crunyear AND ;
              crectype = 'R' AND ;
              nPlugExp # 0  ;
          INTO CURSOR temps

      SELECT temps
      SCAN
         SCATTER MEMVAR
         INSERT INTO PluggingFund FROM MEMVAR
      ENDSCAN

      USE IN tempd
      USE IN temps

      TRY
         SET PROCEDURE TO plugging.prg ADDITIVE
      CATCH
      ENDTRY

      oPlugging = CREATEOBJECT('plugging')

      IF VARTYPE(oPlugging) = 'O'
         llReturn = oPlugging.AddPluggingFundRecs('pluggingfund')
      ELSE
         llReturn = .F.
      ENDIF

      oPlugging = .NULL.

      RETURN llReturn


*********************************
   PROCEDURE TimeKeeper
*********************************
      LPARA tcdescription

      llReturn = .T.

      TRY
         IF THIS.ldebug
            IF USED('debugtime')
               m.cDesc = tcdescription
               m.ntime = SECONDS()
               INSERT INTO debugtime FROM MEMVAR
               FLUSH IN debugtime FORCE
            ENDIF
         ENDIF
      CATCH TO loError
         llReturn = .F.
         DO errorlog WITH 'TimeKeeper', loError.LINENO, 'DistProc', loError.ERRORNO, loError.MESSAGE, ' ', loError
      ENDTRY

      RETURN llReturn
   ENDPROC

*********************************
   PROCEDURE RushMoreOutput
*********************************
      LPARAMETERS tcVar

* Outputs the sys(3054) output to a file

      DO CASE
         CASE tcVar = 'CLOSE'
*!*          =FCLOSE(this.nfilehandle)
            = SYS(3054, 0)
            = SYS(3092, '')
            RETURN
         CASE tcVar = 'START'
            IF VARTYPE(gRushMore) = 'U'
               PUBLIC gRushMore
            ENDIF
            IF THIS.nfilehandle = 0
*         this.nfilehandle = FCREATE('datafiles\rushmore.txt')
            ENDIF
            = SYS(3054, 12, "gRushMore")
            = SYS(3092, 'datafiles\rushmore.txt', .T.)
         OTHERWISE
*!*         tcvar = '***** ' + tcvar + ' *****'
*!*         =FPUTS(this.nfilehandle,tcVar)
*!*         =FWRITE(this.nfilehandle,gRushMore)
      ENDCASE
   ENDPROC

***********************************
   PROCEDURE ErrorMessage
***********************************
      LPARAMETERS tcMethod, tnLineNo, tcModule, tnErrorNo, tcMessage

      IF m.goApp.lDebugMode AND VERSION(2) = 2
         IF MESSAGEBOX('Error Encountered!' + CHR(10) + CHR(10) + ;
                 'Error No: ' + TRANSFORM(tnErrorNo) + CHR(10) + ;
                 'Method: ' + tcMethod + CHR(10) + ;
                 'Line No: ' + TRANSFORM(tnLineNo) + CHR(10) + ;
                 'Message: ' + tcMessage + CHR(10) + CHR(10) + ;
                 'Enter Debug Mode?', 36, 'Debug Error') = 6
            SET STEP ON
         ENDIF
      ELSE
         IF tnErrorNo = 5
            MESSAGEBOX("Unable to complete the processing at this time." + CHR(10) + CHR(10) + ;
                 "There appears to be a problem with this company's index files. " + ;
                 "Try running the Re-Index File Utility found under the Utilities menu and then try again." + CHR(10) + CHR(10) + ;
                 "Check the System Log found under Other Reports for more information." + ;
                 "Contact Pivoten Support - 877-PIOVTEN or support@pivoten.com.", 16, "Index Problem Encountered")
         ELSE
            DO errorlog WITH 'ErrorMessage', tnLineNo, 'DistProc', tnErrorNo, tcMessage, ''
            ErrorMessageText()
         ENDIF
      ENDIF
   ENDPROC


****************************
   PROCEDURE oPartnership_Access
****************************
      IF VARTYPE(THIS.oPartnerShip) # 'O'
         IF NOT 'swpartners' $ LOWER(SET('procedure'))
            SET PROCEDURE TO swpartners.prg ADDITIVE
         ENDIF
         THIS.oPartnerShip = CREATEOBJECT("partners")
      ENDIF
      THIS.oPartnerShip.nDataSession = THIS.nDataSession
      RETURN THIS.oPartnerShip
   ENDPROC

********************************
   FUNCTION CheckArchived(tcYear)
********************************

      lnFiles = ADIR(laFiles, m.goApp.cdatafilepath + 'ownhist' + tcYear + '.dbf')

      IF lnFiles > 0
         llReturn = .T.
      ELSE
         llReturn = .F.
      ENDIF

      RETURN llReturn

ENDDEFINE
*
*-- EndDefine: distproc
**************************************************

















