package routes

import (
	"net/http"
	"odonto-flow-go/internal/config"
	"odonto-flow-go/internal/handlers"
	"odonto-flow-go/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "odonto-flow-go/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
    // Middleware
    r.Use(middleware.CORSMiddleware())
    r.Use(middleware.TimezoneInjector())

    // Handlers
    authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
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
    usuarioHandler := handlers.NewUsuarioHandler(db)
    fiscalHandler := handlers.NewFiscalHandler(db)
    rhFinanceiroHandler := handlers.NewRHFinanceiroHandler(db)

    // Routes
    api := r.Group("/api")
    {
        // Swagger documentation
        api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
        auth := api.Group("/auth")
        {
            auth.POST("/login", middleware.LoginRateLimiter(), authHandler.Login)
            auth.POST("/logout", authHandler.Logout)
            auth.POST("/signup", authHandler.Register)
        }

        // Webhooks (Unprotected by JWT, validated internally by signature)
        webhooks := api.Group("/webhooks")
        {
            webhooks.POST("/dock/payment", handlers.HandleDockWebhook(db))
        }

        // Public Support Routes
        ticketHandler := handlers.NewTicketHandler(db)
        api.POST("/tickets", ticketHandler.CreateTicket)

        // Protected routes
        protected := api.Group("")
        
        // 1. Instancia o AuthProvider UMA VEZ lendo do disco
        authProvider := middleware.NewAuthProvider("public.pem")
        
        // 2. Injeta o middleware baseado em memória
        protected.Use(authProvider.AuthMiddleware())
        
        protected.Use(middleware.TenantDBMiddleware(db)) // Add Tenant Isolation Middleware
        protected.Use(middleware.AuditMiddleware(db)) // Add Audit Middleware
        {
            clinicas := protected.Group("/clinicas")
            {
                clinicas.POST("", middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL"), clinicaHandler.Create)
                clinicas.GET("", clinicaHandler.FindAll)
                clinicas.GET("/rede", clinicaHandler.GetRede)
                clinicas.GET("/:id", clinicaHandler.FindOne)
            }

            dentistas := protected.Group("/dentistas")
            {
                dentistas.POST("", dentistaHandler.Create)
                dentistas.GET("", dentistaHandler.FindAll)
            }

            pacientes := protected.Group("/pacientes")
            pacientes.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "DENTISTA", "RECEPCIONISTA", "ASSISTENTE"))
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
            funcionarios.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL"))
            {
                funcionarios.POST("", funcionarioHandler.Create)
                funcionarios.GET("", funcionarioHandler.FindAll)
                funcionarios.PATCH("/:id/demitir", funcionarioHandler.Demitir)
            }

            // Recursos Humanos e Folha
            folhaHandler := handlers.NewFolhaHandler(db)
            rh := protected.Group("/rh")
            rh.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL"))
            {
                rh.POST("/folha/processar", folhaHandler.ProcessarFolha)
                rh.GET("/folha", folhaHandler.GetHolerites)
                
                // Gestão Dinâmica de Rubricas
                rh.GET("/rubricas", folhaHandler.GetRubricas)
                rh.POST("/rubricas", folhaHandler.CreateRubrica)
            }

            usuarios := protected.Group("/usuarios")
            usuarios.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "RH"))
            {
                usuarios.POST("", usuarioHandler.Create)
                usuarios.GET("", usuarioHandler.FindAll)
                usuarios.PATCH("/:id", usuarioHandler.Update)
                usuarios.DELETE("/:id", usuarioHandler.Delete)
            }

            procedimentos := protected.Group("/procedimentos")
            {
                procedimentos.POST("", procedimentoHandler.Create)
                procedimentos.GET("", procedimentoHandler.FindAll)
            }

            agendamentos := protected.Group("/agendamentos")
            agendamentos.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "DENTISTA", "RECEPCIONISTA", "ASSISTENTE"))
            {
                agendamentos.POST("", agendamentoHandler.Create)
                agendamentos.GET("", agendamentoHandler.FindAll)
                agendamentos.GET("/:id", agendamentoHandler.FindOne)
                agendamentos.PATCH("/:id", agendamentoHandler.Update)
                agendamentos.PATCH("/:id/status", agendamentoHandler.UpdateStatus)
                agendamentos.DELETE("/:id", agendamentoHandler.Delete)
            }

            v1 := protected.Group("/v1")
            {
                v1.GET("/profissionais", dentistaHandler.GetProfissionaisAgenda)
                v1.GET("/agendamentos", middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "DENTISTA", "RECEPCIONISTA", "ASSISTENTE"), agendamentoHandler.FindAll)

                // Perimeter RH - Strict RBAC
                rh := v1.Group("/rh")
                rh.Use(middleware.AdminAuthMiddleware("ADMIN", "RH"))
                {
                    rh.POST("/funcionario", rhFinanceiroHandler.SaveFuncionario)
                    rh.POST("/folha/ajuste", rhFinanceiroHandler.AplicarAjusteManual)
                    rh.GET("/folha/exportar/:holerite_id", rhFinanceiroHandler.ExportarHolerite)
                }

                // Perimeter Financeiro - Strict RBAC
                fin := v1.Group("/financeiro")
                fin.Use(middleware.AdminAuthMiddleware("ADMIN", "FINANCEIRO", "RH"))
                {
                    fin.POST("/folha/fechamento", rhFinanceiroHandler.FecharFolhaEFiscal)
                }
            }

            planos := protected.Group("/planos-tratamento")
            planos.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "DENTISTA", "ASSISTENTE"))
            {
                planos.POST("", planoTratamentoHandler.Create)
                planos.GET("", planoTratamentoHandler.FindAll)
                planos.GET("/:id", planoTratamentoHandler.FindOne)
                planos.PATCH("/:id/status", planoTratamentoHandler.UpdateStatus)
            }

            evolucoes := protected.Group("/evolucoes")
            evolucoes.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "DENTISTA", "ASSISTENTE"))
            {
                evolucoes.POST("", evolucaoHandler.Create)
                evolucoes.GET("", evolucaoHandler.FindAll)
            }

            convenios := protected.Group("/convenios")
            {
                convenios.POST("", convenioHandler.Create)
                convenios.GET("", convenioHandler.FindAll)
            }

            transacoes := protected.Group("/transacoes-financeiras")
            transacoes.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "FATURISTA"))
            {
                transacoes.POST("", transacaoHandler.Create)
                transacoes.GET("", transacaoHandler.FindAll)
                transacoes.GET("/resumo", transacaoHandler.GetSummary)
                transacoes.PATCH("/:id/status", transacaoHandler.UpdateStatus)
            }

            faturas := protected.Group("/faturamento/faturas")
            faturas.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "FATURISTA"))
            {
                faturas.POST("", faturaHandler.Create)
                faturas.GET("", faturaHandler.FindAll)
                faturas.GET("/atrasadas", faturaHandler.GetAtrasadas)
                faturas.POST("/:id/pagar", faturaHandler.PagarFatura)
                faturas.POST("/:id/cancelar", faturaHandler.CancelarFatura)
            }

            contabilidade := protected.Group("/contabilidade")
            contabilidade.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "FATURISTA"))
            {
                contabilidade.POST("/centros-custo", contabilidadeHandler.CreateCentroCusto)
                contabilidade.GET("/centros-custo", contabilidadeHandler.FindAllCentrosCusto)

                contabilidade.POST("/plano-contas", contabilidadeHandler.CreateConta)
                contabilidade.GET("/plano-contas", contabilidadeHandler.FindAllContas)
                contabilidade.GET("/plano-contas/estrutura", contabilidadeHandler.GetEstrutura)
                
                // Dashboard routes
                contabilidade.GET("/dashboard/fluxo-caixa", contabilidadeHandler.GetFluxoCaixa)
                contabilidade.GET("/dashboard/bi-metrics", contabilidadeHandler.GetBIDashboard)
                contabilidade.GET("/dre/mensal", contabilidadeHandler.GetDREMensal)
            }

            fiscal := protected.Group("/fiscal")
            {
                fiscal.GET("/nfs", fiscalHandler.GetNotasFiscais)
                fiscal.POST("/nfs/emitir", fiscalHandler.EmitirNotaFiscal)
            }

            estoque := protected.Group("/estoque")
            estoque.Use(middleware.RequireRole(db, "ADMIN_TOTAL", "ADMIN_GERENCIAL", "RECEPCIONISTA", "ASSISTENTE"))
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

            adminHandler := handlers.NewAdminHandler(db)
            admin := protected.Group("/administracao")
            admin.Use(middleware.RequireRole(db, "ADMIN_TOTAL"))
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
                assinaturas.POST("/", middleware.RequireRole(db, "ADMIN_TOTAL"), assinaturaHandler.Assinar)
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
