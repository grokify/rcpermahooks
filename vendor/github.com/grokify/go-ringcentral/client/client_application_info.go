/*
 * RingCentral Connect Platform API Explorer
 *
 * <p>This is a beta interactive API explorer for the RingCentral Connect Platform. To use this service, you will need to have an account with the proper credentials to generate an OAuth2 access token.</p><p><h2>Quick Start</h2></p><ol><li>1) Go to <b>Authentication > /oauth/token</b></li><li>2) Enter <b>app_key, app_secret, username, password</b> fields and then click \"Try it out!\"</li><li>3) Upon success, your access_token is loaded and you can access any form requiring authorization.</li></ol><h2>Links</h2><ul><li><a href=\"https://github.com/ringcentral\" target=\"_blank\">RingCentral SDKs on Github</a></li><li><a href=\"mailto:devsupport@ringcentral.com\">RingCentral Developer Support Email</a></li></ul>
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package ringcentral

type ClientApplicationInfo struct {

	// 'True', if the server succeeded detecting application info, sufficient to return provisioning info
	Detected bool `json:"detected"`

	// The value of 'User-Agent' header, as it was passed in request
	UserAgent string `json:"userAgent,omitempty"`

	// Application identifier (from authorization session)
	AppId string `json:"appId,omitempty"`

	// Application name (from authorization session, but must match 'User-Agent')
	AppName string `json:"appName,omitempty"`

	// Application version (parsed from 'User-Agent')
	AppVersion string `json:"appVersion,omitempty"`

	// Application platform operation system (parsed from 'User-Agent': Windows, MacOS, Android, iOS
	AppPlatform string `json:"appPlatform,omitempty"`

	// Application platform operation system version (parsed from 'User-Agent')
	AppPlatformVersion string `json:"appPlatformVersion,omitempty"`

	// Locale, parsed from 'Accept-Language'. Currently en-GB and en-US locales are supported. The default value is en-US
	Locale string `json:"locale,omitempty"`
}
