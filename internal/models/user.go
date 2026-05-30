package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Nome        string         `gorm:"not null" json:"nome"`
	Email       string         `gorm:"uniqueIndex;not null" json:"email"`
	SenhaHash   string         `gorm:"not null" json:"-"`
	Telefone    string         `json:"telefone"`
	CPF         string         `gorm:"uniqueIndex;not null" json:"cpf"`
	TipoUsuario string         `gorm:"not null" json:"tipoUsuario"` // ADMIN_TOTAL, ADMIN_GERENCIAL, DENTISTA, RECEPCIONISTA, ASSISTENTE, FATURISTA
	ClinicaID   uint           `json:"clinicaId"`
	Clinica     *Clinica       `gorm:"foreignKey:ClinicaID" json:"clinica,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.SenhaHash = string(hash)
	return nil
}
