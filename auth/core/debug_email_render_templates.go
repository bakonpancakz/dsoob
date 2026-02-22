package core

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"dsoob/backend/include"
	"dsoob/backend/tools"
)

// Renders all Email Templates and then immediately exits
// Intended to help with design or template tweaking

func DebugEmailRenderTemplates() {

	var (
		exampleAddress  = "127.0.0.1"
		exampleToken    = tools.GenerateTokenString()
		exampleLocation = "Fresno, California, United States"
		exampleBrowser  = "Chrome on Windows 10.0"
		exampleTime     = "10/23/2025 07:45am"
		defaults        = map[string]any{
			"EMAIL_VERIFY.txt": tools.LocalsEmailVerify{
				Token: exampleToken,
			},
			"LOGIN_FORGOT_PASSWORD.txt": tools.LocalsLoginForgotPassword{
				Token: exampleToken,
			},
			"LOGIN_NEW_DEVICE.txt": tools.LocalsLoginNewDevice{
				Timestamp:      exampleTime,
				IpAddress:      exampleAddress,
				DeviceBrowser:  exampleBrowser,
				DeviceLocation: exampleLocation,
			},
			"LOGIN_NEW_LOCATION.txt": tools.LocalsLoginNewLocation{
				Token:          exampleToken,
				IpAddress:      exampleAddress,
				Timestamp:      exampleTime,
				DeviceBrowser:  exampleBrowser,
				DeviceLocation: exampleLocation,
			},
			"LOGIN_PASSCODE.txt": tools.LocalsLoginPasscode{
				Code:     tools.GeneratePasscode(),
				Lifetime: fmt.Sprint(tools.TOKEN_LIFETIME_EMAIL_PASSCODE.Minutes()),
			},
			"NOTIFY_USER_DELETED.txt": tools.LocalsNotifyUserDeleted{
				Reason: "User Request",
			},
			"NOTIFY_USER_EMAIL_MODIFIED.txt": tools.LocalsNotifyUserEmailModified{},
			"NOTIFY_USER_PASS_MODIFIED.txt":  tools.LocalsNotifyUserPasswordModified{},
		}
	)

	// Render Templates
	entries, err := include.Templates.ReadDir("templates")
	if err != nil {
		fmt.Printf("Cannot read embedded directory: %s\n", err)
		return
	}
	for _, ent := range entries {

		// Sanity Checks
		filename := path.Base(ent.Name())
		if strings.HasPrefix(filename, "_") {
			fmt.Printf("Ignoring Template: %s\n", filename)
			continue
		}
		locals, ok := defaults[ent.Name()]
		if !ok {
			fmt.Printf("Ignoring Template, contribute some locals!: %s\n", filename)
			continue
		}

		// Process Template
		template, err := template.ParseFS(include.Templates, "templates/"+filename)
		if err != nil {
			fmt.Printf("Cannot parse template '%s': %s\n", filename, err)
			return
		}

		// Execute Template
		os.Mkdir("debug", 0600)
		f, err := os.Create("debug/" + filename)
		if err != nil {
			fmt.Printf("Create file error: %s\n", err)
			return
		}
		if err := template.Execute(f, map[string]any{
			"Host": tools.SITE_NAME,
			"Data": locals,
		}); err != nil {
			fmt.Printf("Cannot Render Teamplate '%s': %s\n", filename, err)
			return
		} else {
			fmt.Printf("Rendered Template '%s'\n", filename)
		}

	}
}
