package stineapi

import (
	"testing"
)

const stineHTMLPageBenutzerkonto = `<div id="contentSpacer_IE" class="pageElementTop">
   <h1>Persönliche Daten</h1>
   <h2 personid="9999999">Peter Lustig</h2>
   <p>Bitte prüfen Sie regelmäßig, ob Ihre hier hinterlegte Anschrift noch aktuell ist und geben Sie die Daten sorgfältig ein. Eine aktuelle Adresse ist wichtig, damit Ihnen die Post der Universität Hamburg (z.B. die Semesterunterlagen mit Studierendenausweis und Semesterticket) zugestellt werden kann. Bitte beachten Sie, dass Sie nur eine Anschrift in Deutschland hinterlegen können.</p>
   <p><u>Adressänderung bei Anschrift im Ausland: </u>Befinden Sie sich noch in der Phase der Einschreibung und haben daher noch eine Anschrift im Ausland, ist die Adressänderung nicht in Ihrem STiNE-Account, sondern nur über das Team Bewerbung, Zulassung und Studierendenangelegenheiten (<a href="https://www.uni-hamburg.de/kontakt-cc">www.uni-hamburg.de/kontakt-cc</a>) möglich.</p>
   <table class="tb persaddrTbl">
      <tbody>
         <tr>
            <th class="tbhead" colspan="4">
               Information
            </th>
         </tr>
         <tr>
            <td class="tbcontrol" colspan="4">
               <a href="">Ändern</a>&nbsp;
            </td>
         </tr>
         <tr class="tbsubhead">
            <td style="width:120px;"></td>
            <td style="width:338px;"></td>
            <td style="width:50px;"></td>
            <td style="width:50px;"></td>
         </tr>
         <tr class="tbdata">
            <td>Matrikelnummer</td>
            <td name="matriculationNumber">
               1873453
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Vorname</td>
            <td name="firstName">
               Peter
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Nachname</td>
            <td name="middleName">
               Lustig
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Messages an Uni-Mail-Adresse weiterleiten?</td>
            <td name="emailSend">
               <input type="checkbox" class="checkBox" name="person_000000010000014" checked="checked" disabled="disabled">
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Zweite Staatsangehörigkeit</td>
            <td name="citiznship2">
               Deutschland 2           	
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Telefon</td>
            <td>
               20318203903812830921                       	
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Handy</td>
            <td>
               +4917234432343423
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Email</td>
            <td>
               test@test.de
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Unimail</td>
            <td>
               unimail@test.de
            </td>
            <td></td>
            <td></td>
         </tr>
      </tbody>
   </table>
   <table class="tb persaddrTbl">
      <tbody>
         <tr>
            <th class="tbhead" colspan="4">
               Heimatanschrift
            </th>
         </tr>
         <tr>
            <td class="tbcontrol" colspan="4">
               <a href="">Ändern</a>&nbsp;
            </td>
         </tr>
         <tr class="tbsubhead">
            <td style="width:120px;"></td>
            <td style="width:338px;"></td>
            <td style="width:50px;"></td>
            <td style="width:50px;"></td>
         </tr>
         <tr class="tbdata">
            <td>Straße</td>
            <td>
               Straße 2
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Adresszusatz</td>
            <td>
               Addition                            
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Land</td>
            <td>
               Deutschland
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>PLZ</td>
            <td>
               2342312
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr class="tbdata">
            <td>Stadt</td>
            <td>
               Hamburg
            </td>
            <td></td>
            <td></td>
         </tr>
         <tr>
            <th class="tbhead" colspan="4">
               Statistik
            </th>
         </tr>
         <tr class="tbsubhead">
            <td style="width:120px;"></td>
            <td style="width:338px;"></td>
            <td style="width:50px;"></td>
            <td style="width:50px;"></td>
         </tr>
         <tr class="tbdata">
            <td>Bundesland</td>
            <td>
               Hamburg
            </td>
            <td></td>
            <td></td>
         </tr>
      </tbody>
   </table>
</div>`

const (
	personId            = "9999999"
	name                = "Peter"
	surname             = "Lustig"
	matriculationNumber = "1873453"
	emailSend           = true
	citiznship2         = "Deutschland 2"
	telephone           = "20318203903812830921"
	mobile              = "+4917234432343423"
	mail                = "test@test.de"
	unimal              = "unimail@test.de"
	street              = "Straße 2"
	addition            = "Addition"
	country             = "Deutschland"
	plz                 = "2342312"
	city                = "Hamburg"
	germanState         = "Hamburg"
)

func TestGetUserAccountURL(t *testing.T) {
	sess := NewSession()
	sess.sessionNo = "fakeNumber2323"
	url := sess.getUserAccountURL()

	if url != "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=PERSADDRESS&ARGUMENTS=-NfakeNumber2323,-N000273," {
		t.Error("session number was not inserted into URL")
	}
}
