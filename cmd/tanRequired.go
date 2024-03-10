package stineapi

import (
	"github.com/martenmatrix/stine-api/cmd/internal/tan"
	"net/http"
)

/*
TanRequired is returned from a function, if an iTAN is needed to complete the action.
TanStartsWith represents the two starting numbers of the required iTAN with a leading zero.
*/
type TanRequired struct {
	client         *http.Client // authenticated client on the stine website
	sessionNo      string       // sessionNo of the authenticated client
	url            string       // url the itan should be sent to
	registrationId string
	TanStartsWith  string // The numbers the required iTAN starts with
}

/*
SetTan sends the provided iTAN to the STiNE servers to complete an action. If the validation fails, an error is returned.
The users iTAN list will be disabled after 3 failed attempts.
The iTAN can be entered with the first three numbers or without the prefix provided by STiNE.
*/
func (tanReq *TanRequired) SetTan(itan string) error {
	tanWithoutPrefix := tan.RemoveTanPrefix(itan, tanReq.TanStartsWith)
	err := tan.SendTAN(tanReq.client, tanReq.url, tanWithoutPrefix, tanReq.sessionNo, tanReq.registrationId)
	if err != nil {
		return err
	}
	return nil
}
