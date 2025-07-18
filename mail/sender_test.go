package mail

import (
	"testing"

	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
		<html>
			<body>
				<h1>Hello!</h1>
				<p><em>Test</em></p>
			</body>
		</html>
	`
	attachFiles := []string{"../start.sh"}

	err = sender.SendEmail(
		subject,
		content,
		[]string{"f04646665@gmail.com"},
		nil,
		nil,
		attachFiles,
	)
	require.NoError(t, err)
}
