package services

import (
	"fmt"
	"net/smtp"

	"github.com/spf13/viper"
)

// SendEmail envia um e-mail real utilizando servidor SMTP configurado nas variáveis de ambiente.
func SendEmail(to []string, subject, body string) error {
	smtpHost := viper.GetString("SMTP_HOST")
	smtpPort := viper.GetString("SMTP_PORT")
	smtpUser := viper.GetString("SMTP_USER")
	smtpPass := viper.GetString("SMTP_PASS")

	// Se não houver configuração, apenas loga e retorna (evita quebrar em dev se não configurado)
	if smtpHost == "" || smtpUser == "" {
		fmt.Println("⚠️ AVISO: E-mail não enviado. Credenciais SMTP (SMTP_HOST, SMTP_USER, SMTP_PASS) não estão configuradas no .env")
		return nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Monta o cabeçalho do e-mail
	msg := []byte("To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n" +
		body + "\r\n")

	addr := smtpHost + ":" + smtpPort
	err := smtp.SendMail(addr, auth, smtpUser, to, msg)
	if err != nil {
		fmt.Printf("❌ ERRO ao enviar e-mail para %v: %v\n", to, err)
		return err
	}

	fmt.Printf("✅ E-mail enviado com sucesso para %v\n", to)
	return nil
}
