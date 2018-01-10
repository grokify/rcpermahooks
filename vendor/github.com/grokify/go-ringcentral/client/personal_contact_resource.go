/*
 * RingCentral Connect Platform API Explorer
 *
 * <p>This is a beta interactive API explorer for the RingCentral Connect Platform. To use this service, you will need to have an account with the proper credentials to generate an OAuth2 access token.</p><p><h2>Quick Start</h2></p><ol><li>1) Go to <b>Authentication > /oauth/token</b></li><li>2) Enter <b>app_key, app_secret, username, password</b> fields and then click \"Try it out!\"</li><li>3) Upon success, your access_token is loaded and you can access any form requiring authorization.</li></ol><h2>Links</h2><ul><li><a href=\"https://github.com/ringcentral\" target=\"_blank\">RingCentral SDKs on Github</a></li><li><a href=\"mailto:devsupport@ringcentral.com\">RingCentral Developer Support Email</a></li></ul>
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package ringcentral

import (
	"time"
)

type PersonalContactResource struct {
	Uri string `json:"uri,omitempty"`

	Availability string `json:"availability,omitempty"`

	Id string `json:"id,omitempty"`

	FirstName string `json:"firstName,omitempty"`

	LastName string `json:"lastName,omitempty"`

	MiddleName string `json:"middleName,omitempty"`

	Birthday time.Time `json:"birthday,omitempty"`

	Notes string `json:"notes,omitempty"`

	WebPage string `json:"webPage,omitempty"`

	Company string `json:"company,omitempty"`

	JobTitle string `json:"jobTitle,omitempty"`

	NickName string `json:"nickName,omitempty"`

	Email string `json:"email,omitempty"`

	Email2 string `json:"email2,omitempty"`

	Email3 string `json:"email3,omitempty"`

	HomeAddress *ContactAddressInfo `json:"homeAddress,omitempty"`

	OtherAddress *ContactAddressInfo `json:"otherAddress,omitempty"`

	HomePhone string `json:"homePhone,omitempty"`

	HomePhone2 string `json:"homePhone2,omitempty"`

	MobilePhone string `json:"mobilePhone,omitempty"`

	BusinessPhone string `json:"businessPhone,omitempty"`

	CallbackPhone string `json:"callbackPhone,omitempty"`

	CarPhone string `json:"carPhone,omitempty"`

	CompanyPhone string `json:"companyPhone,omitempty"`

	OtherPhone string `json:"otherPhone,omitempty"`

	BusinessFax string `json:"businessFax,omitempty"`

	OtherFax string `json:"otherFax,omitempty"`

	BusinessAddress *ContactAddressInfo `json:"businessAddress,omitempty"`

	AssistantPhone string `json:"assistantPhone,omitempty"`

	BusinessPhone2 string `json:"businessPhone2,omitempty"`
}
