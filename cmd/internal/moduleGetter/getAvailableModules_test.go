package moduleGetter

import (
	"fmt"
	"github.com/luci/go-render/render"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetAvailableModules(t *testing.T) {
	secondCategoryPage := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(`
				<!--Second Page with categories and modules-->
				
				<!--Categories-->
				<ul>
					<li>
						::marker
						<a href="/scripts/category.org">Category Cool</a>
					</li>
					<li>
						::marker
						<a href="/scripts/nice.org">Category Nice</a>
					</li>
				</ul>
				
				<!--Modules-->
				<tr>
					<!--logo column-->
					<td class="tbsubhead"> <!-- FIXME TDs ... Module Level ?? -->
					</td>
				
					<!-- MODULE -->
					<td class="tbsubhead dl-inner">
						<p><strong><a href="/scripts/registration1">InfB-SE 2 <span class="eventTitle">Software Development II (SuSe 23)</span></a></strong></p>
						<p>Peter Lustig; Franz Karen</p>
					</td>
				
				
					<td class="tbsubhead">
						13.04.2023<br>
					</td>
				
					<td class="tbsubhead rw-qbf">
						<a href="/scripts/mgrqispi.dll?REGISTERFORMODULE" class="img noFloat register">Register</a>
					</td>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<td class="tbdata"> <!-- FIXME TDs ... Module part Level ?? -->
					</td>
				
					<td class="tbdata">
						InfB_SE2:OOPM&nbsp;Vorlesung Objektorientierte Programmierung und Modellierung
					</td>
				
				
					<td class="tbdata">
						&nbsp;
					</td>
				
					<td class="tbdata">
						&nbsp;
					</td>
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--logo column-->
					<td class="tbdata">
					</td>
				
				
					<td class="tbdata dl-inner">
						<p><strong><a href="/scripts/dwdwd" name="eventLink">64-010 <span class="eventTitle">Lecture Software Development II: Object-oriented Programming and Modelling </span></a></strong></p>
						<p>Prof. Peter Parker</p>
						<p>Wed, 3. Apr. 2024 [14:15] - Wed, 10. Jul. 2024 [15:45]</p>
						<p></p>
					</td>
				
				
					<td class="tbdata">
						07.03.2024<br>550 | 162
					</td>
				
					<td class="tbdata rw-qbf">
				
				
					</td>
				
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<td class="tbdata"> <!-- FIXME TDs ... Module part Level ?? -->
					</td>
				
					<td class="tbdata">
						InfB_SE2_ÜP&nbsp;Übungen zu Softwareentwicklung II
					</td>
				
				
					<td class="tbdata">
						&nbsp;
					</td>
				
					<td class="tbdata">
						&nbsp;
					</td>
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--logo column-->
					<td class="tbdata">
					</td>
				
				
					<td class="tbdata dl-inner">
						<p><strong><a href="/scripts/scscedw" name="eventLink">64-012 <span class="eventTitle">Exercises Software Development II </span></a></strong></p>
						<p>Voldemort</p>
						<p></p>
					</td>
				
				
					<td class="tbdata">
						07.03.2024<br>458 | 130
					</td>
				
					<td class="tbdata rw-qbf">
				
				
					</td>
				
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!--logo column-->
					<td class="tbsubhead"> <!-- FIXME TDs ... Module Level ?? -->
					</td>
				
					<!-- MODULE -->
					<td class="tbsubhead dl-inner">
						<p><strong><a href="/scripts/deddw22">InfB-VSS <span class="eventTitle">Distributed Systems and Systems Security (SuSe 23)</span></a></strong></p>
						<p>Peter Parker 2</p>
					</td>
				
				
					<td class="tbsubhead">
						13.04.2023<br>
					</td>
				
					<td class="tbsubhead rw-qbf">
					</td>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<td class="tbdata"> <!-- FIXME TDs ... Module part Level ?? -->
					</td>
				
					<td class="tbdata">
						InfB_VSS_Üb&nbsp;Übungen Verteilte Systeme und Systemsicherheit
					</td>
				
				
					<td class="tbdata">
						&nbsp;
					</td>
				
					<td class="tbdata">
						&nbsp;
					</td>
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--COURSE END -->
				</tr>
				
				<tr>
					<!-- MODULE END-->
				
					<!--MODULE PART -->
					<!--MODULE PART END-->
				
					<!--COURSE --> <!-- FIXME TDs ... Course Level ?? -->
					<!--logo column-->
					<td class="tbdata">
					</td>
				
				
					<td class="tbdata dl-inner">
						<p><strong><a href="/scripts/cfefef3" name="eventLink">64-091 <span class="eventTitle">Exercises Distributed Systems and Systems Security </span></a></strong></p>
						<p>Markus Ruehl</p>
						<p></p>
					</td>
				
				
					<td class="tbdata">
						07.03.2024<br>240 | 99
					</td>
				
					<td class="tbdata rw-qbf">
				
				
					</td>
				
					<!--COURSE END -->
				</tr>
			`))

		if err != nil {
			t.Errorf(err.Error())
		}
	}))

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(fmt.Sprintf(`
				<!--First Page without modules-->
				<ul class="one two">
    				<li>
        				::marker
        				<a href="/scripts/category111.org">No Section</a>
    				</li>
    				<li>
        				::marker
        			<a href="/scripts/nice11111.org">Not a Section</a>
    				</li>
				</ul>
		
				<ul>
    				<li>
        				::marker
        				<a href="%s">Category Cool</a>
    				</li>
				</ul>
			`, secondCategoryPage.URL)))
		if err != nil {
			t.Errorf(err.Error())
		}
	}))

	modules, err := GetAvailableModules(1, fakeServer.URL, &http.Client{})

	if err != nil {
		t.Errorf(err.Error())
	}

	shouldReturn := Category{
		Title: "initalPage",
		Url:   fakeServer.URL,
		Categories: &[]Category{{
			Title: "Category Cool",
			Url:   "/scripts/category.org",
			Categories: &[]Category{
				{
					Title:      "Category Cool",
					Url:        "/scripts/category.org",
					Categories: &[]Category{},
					Modules:    &[]Module{},
				},
				{
					Title:      "Category Nice",
					Url:        "/scripts/nice.org",
					Categories: &[]Category{},
					Modules:    &[]Module{},
				},
			},
			Modules: &[]Module{
				{
					Title:            "Software Development II (SuSe 23)",
					Teacher:          "Peter Lustig; Franz Karen",
					RegistrationLink: "/scripts/mgrqispi.dll?REGISTERFORMODULE",
					Events: []Event{
						{
							Id:              "64-010",
							Title:           "Lecture Software Development II: Object-oriented Programming and Modelling",
							Link:            "",
							MaxCapacity:     550,
							CurrentCapacity: 162,
						},
						{
							Id:              "64-012",
							Title:           "Exercises Software Development II",
							Link:            "",
							MaxCapacity:     458,
							CurrentCapacity: 130,
						},
					},
				},
				{
					Title:            "Distributed Systems and Systems Security (SuSe 23)",
					Teacher:          "Peter Parker 2",
					RegistrationLink: "", // should be empty, as simulated user is already registered
					Events: []Event{
						{
							Id:              "64-091",
							Title:           "Exercises Distributed Systems and Systems Security",
							Link:            "", // TODO needs to add a link, if user can register for the module
							MaxCapacity:     240,
							CurrentCapacity: 99,
						},
					},
				},
			},
		},
		}}

	equal := reflect.DeepEqual(modules, shouldReturn)

	// use go-render to compare
	if !equal {
		t.Error(fmt.Sprintf("\n EXPECTED: %s \n RECEIVED: %s", render.Render(shouldReturn), render.Render(modules)))
	}
}
