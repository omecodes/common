package mailer

import "testing"

var html = `
<html>
<head>Ome</head>
<body>
 <p style="color: #777777">Ome mailer Test</p>
</body>
</html>
`

func TestSendMail(t *testing.T) {
	err := sendMail("", 0, "", "", "", "Ome test 1-2", html, "")
	if err != nil {
		t.Error(err.Error())
	}
}
