*!*	 Here’s what the program does, step by step:
*!*	 	1.	Inputs & defaults

*!*	 	•	Expects a checking account ID (tcAccount) and an optional date range (tddate1, tddate2).
*!*	 	•	If tcAccount isn’t a character string, it stops with a message.
*!*	 	•	If dates aren’t provided, it uses a wide default range: Jan 1, 1980 ? Dec 31, 2030.

*!*	 	2.	Build two working sets (cursors)

*!*	 	•	tempchecks: pulls checks for the account within the date range, keeping:
*!*	 	•	centrytype (expected “C” for credit, “D” for debit), dcheckdate, cid (transaction id), cpayee, namount
*!*	 	•	Adds a flag lfound initialized to .F. (false).
*!*	 	•	Ordered by check date.
*!*	 	•	tempgl: pulls GL entries for the account in the same date range, with all glmaster.* fields
*!*	 	•	Adds lfound initialized to .F.
*!*	 	•	Ordered by GL date.

*!*	 	3.	Pass 1 — mark checks that have a matching GL entry
*!*	 For each row in tempchecks:

*!*	 	•	If centrytype = 'C', look in tempgl for a row with the same cid and ncredits = namount.
*!*	 If found, set that check’s lfound = .T..
*!*	 	•	If centrytype = 'D', look in tempgl for the same cid and ndebits = namount.
*!*	 If found, set that check’s lfound = .T..

*!*	 Then it shows only unmatched checks by filtering tempchecks to NOT lfound and opening a Browse window.
*!*	 	4.	Pass 2 — mark GL entries that have a matching check
*!*	 For each row in tempgl:

*!*	 	•	If ncredits <> 0, look in tempchecks for centrytype = 'C' with the same cid and namount = ncredits.
*!*	 If found, set that GL row’s lfound = .T..
*!*	 	•	If ndebits <> 0, look in tempchecks for centrytype = 'D' with the same cid and namount = ndebits.
*!*	 If found, set that GL row’s lfound = .T..

*!*	 Then it shows only unmatched GL rows by filtering tempgl to NOT lfound and opening a Browse window.

*!*	 What this accomplishes
*!*	 	•	It reconciles the Checks table to the GL for a single bank account and date range.
*!*	 	•	Matches are based on two things: the common transaction id (cid) and the exact amount, aligned by type:
*!*	 	•	Checks “C” ? GL ncredits
*!*	 	•	Checks “D” ? GL ndebits
*!*	 	•	You end up with two on-screen lists:
*!*	 	1.	Checks that don’t have a matching GL entry.
*!*	 	2.	GL entries that don’t have a matching check.

*!*	 Notes/assumptions
*!*	 	•	It doesn’t consider dates when matching—only cid and amount.
*!*	 	•	It stops on the first found match (LOCATE), so duplicates aren’t handled explicitly.
*!*	 	•	It uses simple filters and Browse windows for the results instead of returning a dataset.


lpara tcAccount, tddate1, tddate2
IF VARTYPE(tcAccount) # 'C'
   MESSAGEBOX('You must pass a valid checking account!',16,'Find GL/Banking Diffs')
   RETURN
ENDIF
IF VARTYPE(tddate1) # 'D'
   tddate1 = {1/1/1980}
ENDIF
IF VARTYPE(tddate2) # 'D'
   tddate2 = {12/31/2030}
ENDIF 
select centrytype, dcheckdate, cid, cpayee, namount, .f. as lfound ;
    from checks where cacctno = tcAccount and between(dcheckdate,tddate1,tddate2) ;
    into cursor tempchecks readwrite ;
    order by dcheckdate
    
select glmaster.*, .f. as lfound from glmaster ;
    into cursor tempgl readwrite ;
    where cacctno = tcAccount and between(ddate,tddate1,tddate2) ;
    order by ddate 
    
select tempchecks
scan 
   scatter memvar 
   
   if m.centrytype = 'C'
      select tempgl
      locate for cid = m.cid and ncredits = m.namount
      if found()
         select tempchecks
         repl lfound with .t.
      endif
   endif  
   
   if m.centrytype = 'D'
      select tempgl
      locate for cid = m.cid and ndebits = m.namount
      if found()
         select tempchecks
         repl lfound with .t.
      endif 
   endif
 endscan 
 
 select tempchecks
 set filt to not lfound
 brow
 SET FILTER to
 
 select tempgl
scan 
   scatter memvar 
   
   if m.ncredits # 0
      select tempchecks
      locate for centrytype='C' and cid = m.cid and namount = m.ncredits
      if found()
         select tempgl
         repl lfound with .t.
      endif
   endif  
   
   if m.ndebits # 0
      select tempchecks
      locate for centrytype = 'D' and cid = m.cid and namount = m.ndebits
      if found()
         select tempgl
         repl lfound with .t.
      endif 
   endif
 endscan 
 
 select tempgl
 set filt to not lfound
 brow