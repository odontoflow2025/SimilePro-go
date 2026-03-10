package routes

import (
	"net/http"
	"odonto-flow-go/internal/handlers"
	"odonto-flow-go/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "odonto-flow-go/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
    // Middleware
    r.Use(middleware.CORSMiddleware())

    // Handlers
    authHandler := handlers.NewAuthHandler(db)
    clinicaHandler := handlers.NewClinicaHandler(db)
    dentistaHandler := handlers.NewDentistaHandler(db)
    pacienteHandler := handlers.NewPacienteHandler(db)
    funcionarioHandler := handlers.NewFuncionarioHandler(db)
    procedimentoHandler := handlers.NewProcedimentoHandler(db)
    agendamentoHandler := handlers.NewAgendamentoHandler(db)
    planoTratamentoHandler := handlers.NewPlanoTratamentoHandler(db)
    evolucaoHandler := handlers.NewEvolucaoHandler(db)
    anamneseHandler := handlers.NewAnamneseHandler(db)
    convenioHandler := handlers.NewConvenioHandler(db)
    transacaoHandler := handlers.NewTransacaoHandler(db)
    faturaHandler := handlers.NewFaturaHandler(db)
    contabilidadeHandler := handlers.NewContabilidadeHandler(db)
    estoqueHandler := handlers.NewEstoqueHandler(db)

    // Routes
    api := r.Group("/api")
    {
        // Swagger documentation
        api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
        auth := api.Group("/auth")
        {
            auth.POST("/login", authHandler.Login)
            auth.POST("/signup", authHandler.Register) // Changed from /register to match original
        }

        // Protected routes
        protected := api.Group("")
        protected.Use(middleware.AuthMiddleware())
        protected.Use(middleware.AuditMiddleware(db)) // Add Audit Middleware
        {
            clinicas := protected.Group("/clinicas")
            {
                clinicas.POST("", clinicaHandler.Create)
                clinicas.GET("", clinicaHandler.FindAll)
                clinicas.GET("/rede", clinicaHandler.GetRede) // Added network route
                clinicas.GET("/:id", clinicaHandler.FindOne)
            }

            dentistas := protected.Group("/dentistas")
            {
                dentistas.POST("", dentistaHandler.Create)
                dentistas.GET("", dentistaHandler.FindAll)
            }

            pacientes := protected.Group("/pacientes")
            {
                pacientes.POST("", pacienteHandler.Create)
                pacientes.GET("", pacienteHandler.FindAll)
                pacientes.GET("/:id", pacienteHandler.FindOne)

                // Alertas de Saúde
                alertaHandler := handlers.NewAlertaHandler(db)
                pacientes.POST("/:id/alertas", alertaHandler.Create)
                pacientes.GET("/:id/alertas", alertaHandler.FindByPaciente)
                pacientes.DELETE("/:id/alertas/:alertaId", alertaHandler.Delete)

                // Audit Protocol
                pacientes.POST("/protocolo-acesso", pacienteHandler.GerarProtocolo)
            }

            funcionarios := protected.Group("/funcionarios")
            {
                funcionarios.POST("", funcionarioHandler.Create)
                funcionarios.GET("", funcionarioHandler.FindAll)
            }

            procedimentos := protected.Group("/procedimentos")
            {
                procedimentos.POST("", procedimentoHandler.Create)
                procedimentos.GET("", procedimentoHandler.FindAll)
            }

            agendamentos := protected.Group("/agendamentos")
            {
                agendamentos.POST("", agendamentoHandler.Create)
                agendamentos.GET("", agendamentoHandler.FindAll)
                agendamentos.PATCH("/:id/status", agendamentoHandler.UpdateStatus)
            }

            planos := protected.Group("/planos-tratamento") // Changed from /planos
            {
                planos.POST("", planoTratamentoHandler.Create)
                planos.GET("", planoTratamentoHandler.FindAll)
                planos.GET("/:id", planoTratamentoHandler.FindOne)
                planos.PATCH("/:id/status", planoTratamentoHandler.UpdateStatus)
            }

            evolucoes := protected.Group("/evolucoes")
            {
                evolucoes.POST("", evolucaoHandler.Create)
                evolucoes.GET("", evolucaoHandler.FindAll)
            }

            convenios := protected.Group("/convenios")
            {
                convenios.POST("", convenioHandler.Create)
                convenios.GET("", convenioHandler.FindAll)
            }

            transacoes := protected.Group("/transacoes-financeiras") // Changed from /transacoes
            {
                transacoes.POST("", transacaoHandler.Create)
                transacoes.GET("", transacaoHandler.FindAll)
                transacoes.GET("/resumo", transacaoHandler.GetSummary) // Added missing route
                transacoes.PATCH("/:id/status", transacaoHandler.UpdateStatus)
            }

            faturas := protected.Group("/faturamento/faturas") // Changed from /faturas
            {
                faturas.POST("", faturaHandler.Create)
                faturas.GET("", faturaHandler.FindAll)
                faturas.GET("/atrasadas", faturaHandler.GetAtrasadas) // Added missing route
            }

            contabilidade := protected.Group("/contabilidade")
            {
                contabilidade.POST("/centros-custo", contabilidadeHandler.CreateCentroCusto)
                contabilidade.GET("/centros-custo", contabilidadeHandler.FindAllCentrosCusto)

                contabilidade.POST("/plano-contas", contabilidadeHandler.CreateConta)
                contabilidade.GET("/plano-contas", contabilidadeHandler.FindAllContas)
                contabilidade.GET("/plano-contas/estrutura", contabilidadeHandler.GetEstrutura)
                
                // Dashboard routes
                contabilidade.GET("/dashboard/fluxo-caixa", contabilidadeHandler.GetFluxoCaixa)
                contabilidade.GET("/dashboard/bi-metrics", contabilidadeHandler.GetBIDashboard) // Added BI metrics
            }

            estoque := protected.Group("/estoque")
            {
                estoque.GET("/produtos", estoqueHandler.GetProdutos)
                estoque.GET("/notas-fiscais", estoqueHandler.GetNotasFiscais)
                estoque.POST("/notas-fiscais", estoqueHandler.CreateNotaFiscal)
            }
            
            // Faturamento Reports/Extras
            faturamento := protected.Group("/faturamento")
            {
               faturamento.GET("/relatorios/faturamento", func(c *gin.Context) {
                   c.JSON(http.StatusOK, gin.H{}) // Mock report
               }) 
            }

            // Admin Stats & Configs
            adminHandler := handlers.NewAdminHandler(db)
            admin := protected.Group("/administracao")
            {
                admin.GET("/estatisticas", adminHandler.GetStats)
                admin.GET("/configuracoes", adminHandler.ListConfigs)
                admin.GET("/configuracoes/:chave", adminHandler.GetConfig)
                admin.PUT("/configuracoes/:chave", adminHandler.SetConfig)
                admin.GET("/logs", adminHandler.GetLogs)
            }



            // Assinaturas (SaaS)
            assinaturaHandler := handlers.NewAssinaturaHandler(db)
            assinaturas := protected.Group("/assinaturas")
            {
                assinaturas.GET("/planos", assinaturaHandler.GetPlanos)
                assinaturas.GET("/status", assinaturaHandler.GetStatus)
                assinaturas.POST("/", assinaturaHandler.Assinar) // Upgrade/Change plan
            }

            protected.POST("/pacientes/:id/anamnese", anamneseHandler.CreateOrUpdate)
            protected.GET("/pacientes/:id/anamnese", anamneseHandler.FindByPaciente)

            // RNDS Integration
            rndsHandler := handlers.NewRndsHandler(db)
            rnds := protected.Group("/rnds")
            {
                rnds.GET("/status", rndsHandler.CheckStatus)
                rnds.POST("/patients/:id", rndsHandler.SendPatient)
                rnds.POST("/attendances/:id", rndsHandler.SendAttendance)
            }

            
            // User profile route
            protected.GET("/auth/me", authHandler.Me)
        }
    }
}
