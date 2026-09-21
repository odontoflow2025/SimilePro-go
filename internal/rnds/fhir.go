package rnds

import (
	"SimilePro-go/internal/models"
)

// Simplified FHIR structs
type FHIRPatient struct {
	ResourceType string `json:"resourceType"`
	ID           string `json:"id,omitempty"`
	Meta         struct {
		Profile []string `json:"profile"`
	} `json:"meta"`
	Identifier []Identifier `json:"identifier"`
	Name       []HumanName  `json:"name"`
	Gender     string       `json:"gender"`
	BirthDate  string       `json:"birthDate"`
}

type Identifier struct {
	System string `json:"system"`
	Value  string `json:"value"`
}

type HumanName struct {
	Use    string   `json:"use"`
	Family string   `json:"family"` // Sobrenome
	Given  []string `json:"given"`  // Nome
}

// Builder
func BuildPatientFHIR(p models.Paciente) (FHIRPatient, error) {
	return FHIRPatient{
		ResourceType: "Patient",
		Meta: struct {
			Profile []string `json:"profile"`
		}{
			Profile: []string{"http://www.saude.gov.br/fhir/rnds/StructureDefinition/rnds-paciente-1.0"},
		},
		Identifier: []Identifier{
			{
				System: "http://rnds.saude.gov.br/fhir/rnds/NamingSystem/cpf",
				Value:  p.CPF,
			},
			{
				System: "http://rnds.saude.gov.br/fhir/rnds/NamingSystem/cns",
				Value:  "700000000000000", // MOCK CNS if missing
			},
		},
		Name: []HumanName{
			{
				Use:    "official",
				Family: p.Nome, // Simplification: assuming full name in Nome. real world needs split.
				Given:  []string{p.Nome},
			},
		},
		Gender:    mapGender(p.Genero),
		BirthDate: p.DataNascimento.Format("2006-01-02"),
	}, nil
}

func mapGender(g string) string {
	switch g {
	case "M":
		return "male"
	case "F":
		return "female"
	default:
		return "unknown"
	}
}

// ... Additional builders for Encounter/Attendance ...
