package mailer

import "testing"

var html = `
<html>
<head>Zoenion</head>
<body>
 <p style="color: #777777">Zoenion mailer Test</p>
</body>
</html>
`

func TestSendMail(t *testing.T) {
	err := sendMail("", 0, "", "", "", "Zoenion test 1-2", html, "")
	if err != nil {
		t.Error(err.Error())
	}
}
