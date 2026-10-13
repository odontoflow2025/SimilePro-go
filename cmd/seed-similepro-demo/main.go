// Command seed-similepro-demo creates a deterministic demo clinic and sample
// history. It never writes to the database unless the explicit apply flags
// are supplied.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"SimilePro-go/internal/config"
	"SimilePro-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	clinicCNPJ              = "99999999000199" // Identificador fictício, reservado para este seed.
	adminEmail              = "jefe.gomes154@gmail.com"
	receptionistEmail       = "atendente.demo@similepro.invalid"
	technicianEmail         = "tecnico.demo@similepro.invalid"
	operationalDentistEmail = "dentista.operacional@similepro.invalid"
	seedTag                 = "SIMILEPRO_DEMO"
)

func main() {
	startText := flag.String("start", "2020-01-01", "data inicial (AAAA-MM-DD)")
	endText := flag.String("end", "", "data final inclusiva (AAAA-MM-DD); padrão: ontem")
	apply := flag.Bool("apply", false, "gravar no banco; sem esta opção o comando só valida a configuração")
	confirm := flag.String("confirm", "", "confirmação literal exigida: SIMILE-PRO-DEMO")
	accessOnly := flag.Bool("access-only", false, "restaurar o administrador total da conta demo sem gerar histórico")
	operationalUsersOnly := flag.Bool("operational-users-only", false, "criar/garantir os três usuários operacionais na clínica demo existente")
	flag.Parse()

	start, err := time.ParseInLocation("2006-01-02", *startText, time.Local)
	if err != nil {
		log.Fatalf("data inicial inválida: %v", err)
	}
	end := time.Now().In(time.Local).AddDate(0, 0, -1)
	if *endText != "" {
		end, err = time.ParseInLocation("2006-01-02", *endText, time.Local)
		if err != nil {
			log.Fatalf("data final inválida: %v", err)
		}
	}
	if end.Before(start) {
		log.Fatal("a data final deve ser igual ou posterior à data inicial")
	}
	if !*apply {
		if *operationalUsersOnly {
			log.Printf("Prévia: criar/garantir %s (RECEPCIONISTA), %s (TECNICO) e %s (DENTISTA) na clínica Simile Pro existente. Nada foi gravado. Use -operational-users-only -apply -confirm SIMILE-PRO-DEMO para executar.", receptionistEmail, technicianEmail, operationalDentistEmail)
		} else {
			log.Printf("Prévia: clínica %q, conta %q, período %s a %s. Nada foi gravado. Use -apply -confirm SIMILE-PRO-DEMO para executar.", "Simile Pro", adminEmail, start.Format("2006-01-02"), end.Format("2006-01-02"))
		}
		return
	}
	if *confirm != "SIMILE-PRO-DEMO" {
		log.Fatal("para gravar, informe também -confirm SIMILE-PRO-DEMO")
	}
	if *accessOnly && *operationalUsersOnly {
		log.Fatal("-access-only e -operational-users-only não podem ser usados juntos")
	}
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if !*accessOnly && !*operationalUsersOnly && password == "" {
		log.Fatal("defina SEED_ADMIN_PASSWORD no ambiente; a senha não é armazenada no código")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("falha ao carregar configuração: %v", err)
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Sao_Paulo",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("falha ao abrir conexão com o banco configurado: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("falha ao obter conexão: %v", err)
	}
	defer sqlDB.Close()

	err = db.Transaction(func(tx *gorm.DB) error {
		if *accessOnly {
			var clinic models.Clinica
			if err := tx.Where("cnpj = ?", clinicCNPJ).First(&clinic).Error; err != nil {
				return fmt.Errorf("buscar clínica demo existente: %w", err)
			}
			var admin models.User
			if err := tx.Where("email = ?", adminEmail).First(&admin).Error; err != nil {
				return fmt.Errorf("buscar conta administrativa existente %s: %w", adminEmail, err)
			}
			if admin.ClinicaID != clinic.ID {
				return fmt.Errorf("a conta %s não está vinculada à clínica demo; nenhuma conta foi criada ou alterada", adminEmail)
			}
			if err := ensureNetworkAdminRole(tx, &admin); err != nil {
				return err
			}
			if err := ensureAdminMembership(tx, clinic.ID, admin.ID); err != nil {
				return err
			}
			if _, err := ensureDemoDentistProfile(tx, clinic.ID, admin.ID); err != nil {
				return err
			}
			log.Printf("A conta existente %s foi restaurada como ADMIN_TOTAL; vínculo da clínica garantido, sem criar usuário ou alterar senha. Histórico preservado.", adminEmail)
			return nil
		}

		if *operationalUsersOnly {
			var clinic models.Clinica
			if err := tx.Where("cnpj = ?", clinicCNPJ).First(&clinic).Error; err != nil {
				return fmt.Errorf("buscar clínica demo existente: %w", err)
			}
			if clinic.NomeFantasia != "Simile Pro" {
				return fmt.Errorf("o CNPJ reservado do seed não pertence à clínica Simile Pro; nenhuma conta foi criada")
			}
			if err := ensureOperationalDemoUsers(tx, clinic.ID); err != nil {
				return err
			}
			log.Printf("Usuários operacionais fictícios garantidos na clínica Simile Pro (ID %d): %s, %s e %s. Senhas existentes foram preservadas.", clinic.ID, receptionistEmail, technicianEmail, operationalDentistEmail)
			return nil
		}

		clinic, admin, dentist, err := ensureClinicAndUsers(tx, password)
		if err != nil {
			return err
		}
		patients, err := ensurePatients(tx, clinic.ID)
		if err != nil {
			return err
		}
		appointments, transactions, err := ensureHistory(tx, clinic.ID, admin.ID, dentist.ID, patients, start, end)
		if err != nil {
			return err
		}
		log.Printf("Clínica ID %d; %d pacientes; %d agendamentos novos; %d transações novas.", clinic.ID, len(patients), appointments, transactions)
		return nil
	})
	if err != nil {
		log.Fatalf("seed cancelado; transação revertida: %v", err)
	}
	log.Printf("Seed concluído para o período %s a %s.", start.Format("2006-01-02"), end.Format("2006-01-02"))
}

type operationalDemoAccount struct {
	name, email, cpf, role, jobTitle, passwordEnv string
}

func ensureOperationalDemoUsers(tx *gorm.DB, clinicID uint) error {
	accounts := []operationalDemoAccount{
		{name: "Marina Costa (Atendente Demo)", email: receptionistEmail, cpf: "90000000001", role: "RECEPCIONISTA", jobTitle: "RECEPCIONISTA", passwordEnv: "SEED_DEMO_RECEPTIONIST_PASSWORD"},
		{name: "Rafael Lima (Técnico Demo)", email: technicianEmail, cpf: "90000000002", role: "TECNICO", jobTitle: "TÉCNICO EM SAÚDE BUCAL", passwordEnv: "SEED_DEMO_TECHNICIAN_PASSWORD"},
		{name: "Camila Nogueira (Dentista Demo)", email: operationalDentistEmail, cpf: "90000000003", role: "DENTISTA", passwordEnv: "SEED_DEMO_DENTIST_PASSWORD"},
	}
	for _, account := range accounts {
		password := os.Getenv(account.passwordEnv)
		user, err := ensureOperationalDemoUser(tx, clinicID, account, password)
		if err != nil {
			return err
		}
		if account.role == "DENTISTA" {
			if err := ensureOperationalDentistProfile(tx, clinicID, user.ID); err != nil {
				return err
			}
		} else if err := ensureOperationalEmployee(tx, clinicID, user.ID, account.jobTitle); err != nil {
			return err
		}
	}
	return nil
}

func ensureOperationalDemoUser(tx *gorm.DB, clinicID uint, account operationalDemoAccount, password string) (models.User, error) {
	var user models.User
	err := tx.Unscoped().Where("email = ?", account.email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if len(password) < 12 {
			return user, fmt.Errorf("defina %s no ambiente com no mínimo 12 caracteres; senhas não são armazenadas no código", account.passwordEnv)
		}
		user = models.User{Nome: account.name, Email: account.email, CPF: account.cpf, TipoUsuario: account.role, ClinicaID: clinicID}
		if err := user.SetPassword(password); err != nil {
			return user, fmt.Errorf("gerar hash de %s: %w", account.email, err)
		}
		if err := tx.Create(&user).Error; err != nil {
			return user, fmt.Errorf("criar usuário %s: %w", account.email, err)
		}
		return user, nil
	}
	if err != nil {
		return user, fmt.Errorf("buscar usuário %s: %w", account.email, err)
	}
	if user.DeletedAt.Valid {
		return user, fmt.Errorf("a conta fictícia %s está desativada; seed interrompido sem reativá-la", account.email)
	}
	if user.ClinicaID != clinicID || user.TipoUsuario != account.role {
		return user, fmt.Errorf("o e-mail %s já está vinculado a outra clínica ou perfil; nenhuma alteração feita", account.email)
	}
	return user, nil
}

func ensureOperationalEmployee(tx *gorm.DB, clinicID, userID uint, jobTitle string) error {
	var employee models.Funcionario
	err := tx.Unscoped().Where("usuario_id = ? AND clinica_id = ?", userID, clinicID).First(&employee).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		employee = models.Funcionario{UsuarioID: userID, ClinicaID: clinicID, Cargo: jobTitle, DataAdmissao: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local), Status: models.StatusFuncionarioAtivo}
		if err := tx.Create(&employee).Error; err != nil {
			return fmt.Errorf("criar vínculo operacional %d: %w", userID, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("buscar vínculo operacional %d: %w", userID, err)
	}
	if employee.DeletedAt.Valid {
		return fmt.Errorf("o vínculo operacional %d está desativado; seed interrompido sem reativá-lo", userID)
	}
	if employee.Status != models.StatusFuncionarioAtivo || employee.Cargo != jobTitle {
		if err := tx.Model(&employee).Updates(map[string]interface{}{"status": models.StatusFuncionarioAtivo, "cargo": jobTitle}).Error; err != nil {
			return fmt.Errorf("garantir vínculo operacional %d: %w", userID, err)
		}
	}
	return nil
}

func ensureOperationalDentistProfile(tx *gorm.DB, clinicID, userID uint) error {
	var profile models.Dentista
	err := tx.Unscoped().Where("usuario_id = ? AND clinica_id = ?", userID, clinicID).First(&profile).Error
	if err == nil {
		if profile.DeletedAt.Valid {
			return fmt.Errorf("o perfil odontológico do usuário %d está desativado; seed interrompido sem reativá-lo", userID)
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("buscar perfil operacional de dentista: %w", err)
	}
	profile = models.Dentista{UsuarioID: userID, ClinicaID: clinicID, CRO: "DEMO-OPERACIONAL-" + clinicCNPJ, Especialidade: "Clínica geral", DataContratacao: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)}
	if err := tx.Create(&profile).Error; err != nil {
		return fmt.Errorf("criar perfil operacional de dentista: %w", err)
	}
	return nil
}

func ensureClinicAndUsers(tx *gorm.DB, password string) (models.Clinica, models.User, models.Dentista, error) {
	clinic := models.Clinica{}
	err := tx.Where("cnpj = ?", clinicCNPJ).First(&clinic).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		clinic = models.Clinica{
			NomeFantasia: "Simile Pro", RazaoSocial: "Simile Pro Clínica Odontológica (dados fictícios)",
			CNPJ: clinicCNPJ, Plano: "FREE", PlanoStatus: "ACTIVE", MaxFuncionarios: 10,
		}
		// ResponsavelID is a foreign key to users, while each user also points
		// back to this clinic. Insert the clinic with a NULL responsible first,
		// then create the admin and set ResponsavelID below.
		if err := tx.Omit("ResponsavelID").Create(&clinic).Error; err != nil {
			return clinic, models.User{}, models.Dentista{}, fmt.Errorf("criar clínica: %w", err)
		}
	} else if err != nil {
		return clinic, models.User{}, models.Dentista{}, fmt.Errorf("buscar clínica seed: %w", err)
	} else if clinic.NomeFantasia != "Simile Pro" {
		return clinic, models.User{}, models.Dentista{}, fmt.Errorf("o CNPJ reservado do seed já pertence à clínica %q; nenhuma alteração feita", clinic.NomeFantasia)
	}

	admin := models.User{}
	err = tx.Where("email = ?", adminEmail).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		admin = models.User{Nome: "Jefe Gomes", Email: adminEmail, CPF: "99887766001", TipoUsuario: "ADMIN_TOTAL", ClinicaID: clinic.ID}
		if err := admin.SetPassword(password); err != nil {
			return clinic, admin, models.Dentista{}, fmt.Errorf("gerar hash da senha do administrador: %w", err)
		}
		if err := tx.Create(&admin).Error; err != nil {
			return clinic, admin, models.Dentista{}, fmt.Errorf("criar usuário administrador: %w", err)
		}
	} else if err != nil {
		return clinic, admin, models.Dentista{}, fmt.Errorf("buscar usuário administrador: %w", err)
	} else if admin.ClinicaID != clinic.ID {
		return clinic, admin, models.Dentista{}, fmt.Errorf("o e-mail informado já está associado a outra clínica (ID %d); seed interrompido por segurança", admin.ClinicaID)
	} else {
		if err := admin.SetPassword(password); err != nil {
			return clinic, admin, models.Dentista{}, fmt.Errorf("gerar hash da senha do administrador: %w", err)
		}
		admin.TipoUsuario = "ADMIN_TOTAL"
		if err := tx.Model(&admin).Updates(map[string]interface{}{"senha_hash": admin.SenhaHash, "tipo_usuario": admin.TipoUsuario}).Error; err != nil {
			return clinic, admin, models.Dentista{}, fmt.Errorf("atualizar senha do administrador da clínica seed: %w", err)
		}
	}
	if clinic.ResponsavelID != admin.ID {
		if err := tx.Model(&clinic).Update("responsavel_id", admin.ID).Error; err != nil {
			return clinic, admin, models.Dentista{}, fmt.Errorf("associar responsável à clínica: %w", err)
		}
	}

	if err := ensureAdminMembership(tx, clinic.ID, admin.ID); err != nil {
		return clinic, admin, models.Dentista{}, err
	}

	dentist, err := ensureDemoDentistProfile(tx, clinic.ID, admin.ID)
	if err != nil {
		return clinic, admin, models.Dentista{}, err
	}
	return clinic, admin, dentist, nil
}

func ensureNetworkAdminRole(tx *gorm.DB, admin *models.User) error {
	if admin.TipoUsuario == "ADMIN_TOTAL" {
		return nil
	}
	if err := tx.Model(admin).Update("tipo_usuario", "ADMIN_TOTAL").Error; err != nil {
		return fmt.Errorf("restaurar %s como administrador total da rede: %w", adminEmail, err)
	}
	admin.TipoUsuario = "ADMIN_TOTAL"
	return nil
}

func ensureDemoDentistProfile(tx *gorm.DB, clinicID, adminID uint) (models.Dentista, error) {
	// Keep the sample appointments attached to a dentist profile without
	// creating a second login account. The provided demo administrator is the
	// user behind this demo-only dentist profile.
	dentist := models.Dentista{}
	err := tx.Where("clinica_id = ? AND cro = ?", clinicID, "DEMO-"+clinicCNPJ).First(&dentist).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		dentist = models.Dentista{
			UsuarioID: adminID, ClinicaID: clinicID, CRO: "DEMO-" + clinicCNPJ,
			Especialidade: "Clínica geral", PorcentagemRepasse: 0,
			DataContratacao: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
		}
		if err := tx.Create(&dentist).Error; err != nil {
			return dentist, fmt.Errorf("criar perfil de dentista fictício da conta demo existente: %w", err)
		}
	} else if err != nil {
		return dentist, fmt.Errorf("buscar perfil de dentista fictício: %w", err)
	} else if dentist.UsuarioID != adminID {
		dentist.UsuarioID = adminID
		if err := tx.Model(&dentist).Update("usuario_id", adminID).Error; err != nil {
			return dentist, fmt.Errorf("associar perfil fictício à conta demo existente: %w", err)
		}
	}

	// Retire the technical login created by the first seed version. Soft delete
	// preserves referential/audit history while preventing that login from being used.
	legacyDentistUser := models.User{}
	err = tx.Where("email = ?", "dentista.demo@similepro.invalid").First(&legacyDentistUser).Error
	if err == nil && legacyDentistUser.ID != adminID {
		if err := tx.Delete(&legacyDentistUser).Error; err != nil {
			return dentist, fmt.Errorf("desativar usuário técnico de dentista do seed: %w", err)
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dentist, fmt.Errorf("buscar usuário técnico de dentista do seed: %w", err)
	}
	return dentist, nil
}

func ensureAdminMembership(tx *gorm.DB, clinicID, adminID uint) error {
	// Keep the demonstration clinic relationship aligned with the network owner.
	adminMembership := models.Funcionario{}
	err := tx.Where("usuario_id = ? AND clinica_id = ?", adminID, clinicID).First(&adminMembership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		adminMembership = models.Funcionario{
			UsuarioID:    adminID,
			ClinicaID:    clinicID,
			Cargo:        "ADMIN_TOTAL",
			DataAdmissao: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			Status:       models.StatusFuncionarioAtivo,
		}
		if err := tx.Create(&adminMembership).Error; err != nil {
			return fmt.Errorf("criar vínculo da conta existente com a clínica: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("buscar vínculo da conta existente com a clínica: %w", err)
	} else if adminMembership.Status != models.StatusFuncionarioAtivo || adminMembership.Cargo != "ADMIN_TOTAL" {
		if err := tx.Model(&adminMembership).Updates(map[string]interface{}{
			"status": models.StatusFuncionarioAtivo,
			"cargo":  "ADMIN_TOTAL",
		}).Error; err != nil {
			return fmt.Errorf("atualizar vínculo da conta total com a clínica: %w", err)
		}
	}
	return nil
}

func ensurePatients(tx *gorm.DB, clinicID uint) ([]models.Paciente, error) {
	firstNames := []string{"Ana", "Bruno", "Carla", "Diego", "Elisa", "Felipe", "Gabriela", "Heitor", "Isabela", "João", "Karen", "Lucas", "Marina", "Nicolas", "Olívia", "Paulo", "Quitéria", "Rafael", "Sofia", "Tiago", "Ursula", "Vitor", "Wesley", "Yasmin", "Zeca"}
	lastNames := []string{"Almeida", "Barbosa", "Cardoso", "Dias", "Esteves", "Ferreira", "Gomes", "Henrique", "Ibrahim", "Jardim", "Lopes", "Mendes", "Nogueira", "Oliveira", "Pereira", "Queiroz", "Ramos", "Souza", "Teixeira", "Vieira", "Xavier", "Yamada", "Zanetti"}
	patients := make([]models.Paciente, 0, 30)
	for i := 0; i < 30; i++ {
		code := fmt.Sprintf("%s-PAC-%03d", seedTag, i+1)
		p := models.Paciente{}
		err := tx.Where("codigo_unico = ?", code).First(&p).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p = models.Paciente{
				Nome: fmt.Sprintf("%s %s %s", firstNames[i%len(firstNames)], lastNames[(i*7)%len(lastNames)], lastNames[(i*11+3)%len(lastNames)]),
				CPF:  fmt.Sprintf("9%010d", 88000000+i), CodigoUnico: code,
				DataNascimento: time.Date(1965+i%39, time.Month(i%12+1), i%27+1, 0, 0, 0, 0, time.Local),
				Genero:         []string{"F", "M", "O"}[i%3], TelefonePrincipal: fmt.Sprintf("119%08d", 10000000+i),
				Email: fmt.Sprintf("paciente%02d@example.invalid", i+1), EnderecoCompleto: fmt.Sprintf("Rua Fictícia, %d, São Paulo - SP", i+10), ClinicaID: clinicID,
			}
			if err := tx.Create(&p).Error; err != nil {
				return nil, fmt.Errorf("criar paciente fictício %d: %w", i+1, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("buscar paciente seed %s: %w", code, err)
		} else if p.ClinicaID != clinicID {
			return nil, fmt.Errorf("identificador de paciente seed %s já usado por outra clínica", code)
		}
		patients = append(patients, p)
	}
	return patients, nil
}

func ensureHistory(tx *gorm.DB, clinicID, adminID uint, dentistID uint, patients []models.Paciente, start, end time.Time) (int, int, error) {
	newAppointments, newTransactions := 0, 0
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		// Amostra realista e determinística: dois horários por dia útil, sem fins de semana.
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}
		for slot, hour := range []int{9, 14} {
			patient := patients[(day.YearDay()+day.Year()+slot*13)%len(patients)]
			startAt := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, day.Location())
			marker := fmt.Sprintf("%s-%s-%d", seedTag, day.Format("20060102"), slot+1)
			var appointment models.Agendamento
			err := tx.Where("clinica_id = ? AND motivo = ?", clinicID, marker).First(&appointment).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status := models.StatusAgendamentoAtendido
				if (day.YearDay()+slot)%17 == 0 {
					status = models.StatusAgendamentoNaoCompareceu
				} else if (day.YearDay()+slot)%29 == 0 {
					status = models.StatusAgendamentoCancelado
				}
				appointment = models.Agendamento{
					PacienteID: patient.ID, DentistaID: dentistID, ClinicaID: clinicID,
					DataHoraInicio: startAt, DataHoraFim: startAt.Add(45 * time.Minute),
					Motivo: marker, Status: status, UsuarioCriacaoID: adminID,
				}
				if err := tx.Create(&appointment).Error; err != nil {
					return newAppointments, newTransactions, fmt.Errorf("criar agendamento %s: %w", marker, err)
				}
				newAppointments++
			} else if err != nil {
				return newAppointments, newTransactions, fmt.Errorf("buscar agendamento %s: %w", marker, err)
			}
			if appointment.Status != models.StatusAgendamentoAtendido {
				continue
			}
			transactionMarker := "Atendimento " + marker
			var transaction models.Transacao
			err = tx.Where("clinica_id = ? AND descricao = ?", clinicID, transactionMarker).First(&transaction).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				patientID := appointment.PacienteID
				paidAt := appointment.DataHoraInicio
				transaction = models.Transacao{
					ClinicaID: clinicID, PacienteID: &patientID, Descricao: transactionMarker,
					Valor: float64(120 + ((day.YearDay() + slot*37) % 381)), Tipo: models.TipoTransacaoReceita,
					Status: models.StatusTransacaoPago, DataVencimento: appointment.DataHoraInicio,
					DataPagamento: &paidAt, Categoria: "Atendimento odontológico", FormaPagamento: []string{"PIX", "Cartão", "Dinheiro"}[slot],
				}
				if err := tx.Create(&transaction).Error; err != nil {
					return newAppointments, newTransactions, fmt.Errorf("criar transação %s: %w", transactionMarker, err)
				}
				newTransactions++
			} else if err != nil {
				return newAppointments, newTransactions, fmt.Errorf("buscar transação %s: %w", transactionMarker, err)
			}
		}
	}
	return newAppointments, newTransactions, nil
}
