package onPage

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"io"
	"testing"
)

func TestOnSelectExamPage(t *testing.T) {
	onExamPage, errExamPage := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString(`<input name="PRGNAME" type="hidden" value="SAVEEXAMDETAILS">`)))
	if errExamPage != nil {
		t.Errorf(errExamPage.Error())
	}

	notOnExamPage, errNotExamPage := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString(`<input name="PRGNAME" type="hidden">`)))
	if errNotExamPage != nil {
		t.Errorf(errNotExamPage.Error())
	}

	if OnSelectExamPage(onExamPage) != true {
		t.Error("should return true")
	}

	if OnSelectExamPage(notOnExamPage) != false {
		t.Error("should return false")
	}
}

func TestOniTANPage(t *testing.T) {
	fakeRes1, err1 := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString("<html><body></body></html>")))
	if err1 != nil {
		t.Errorf(err1.Error())
	}

	res1 := OniTANPage(fakeRes1)

	if res1 != false {
		t.Error("Expected: false, Received: true")
	}

	fakeRes2, err2 := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString("<html><body><span class=\"itan\"</body></html>")))
	if err2 != nil {
		t.Errorf(err2.Error())
	}
	res2 := OniTANPage(fakeRes2)

	if res2 != true {
		t.Error("Expected: true, Received: false")
	}
}
