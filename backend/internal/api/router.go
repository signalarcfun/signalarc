package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/agent"
	"github.com/wahyu241205/SignalArc/backend/internal/circleapi"
	"github.com/wahyu241205/SignalArc/backend/internal/config"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

const (
	defaultListLimit    = 50
	defaultMarketsLimit = defaultListLimit
)

func NewRouter(db *database.DB) http.Handler {
	router := chi.NewRouter()
	router.Use(requestIDMiddleware, localCORSMiddleware, requestLoggingMiddleware, recovererMiddleware)

	marketsRepository := repository.NewMarketsRepository(db)
	positionsRepository := repository.NewPositionsRepository(db)
	resolutionsRepository := repository.NewResolutionsRepository(db)
	settlementsRepository := repository.NewSettlementsRepository(db)
	tradesRepository := repository.NewTradesRepository(db)
	agentWalletsRepository := repository.NewAgentWalletsRepository(db)
	agentSessionsRepository := repository.NewAgentSessionsRepository(db)
	agentIntentsRepository := repository.NewAgentIntentsRepository(db)
	analyticsRepository := repository.NewAnalyticsRepository(db)
	agentIntentStore := agent.NewStore()
	cfg := config.Load()
	circleExecutor := newCircleAgentWalletExecutor(cfg)
	circleBalanceReader := newCircleAgentWalletBalanceReader(cfg)
	circleOnboardingStarter := agent.CircleOnboardingStarter{
		Enabled: cfg.CircleAgentOnboardingOTPStartEnabled,
		Runner: agent.NewCircleCLIOnboardingRunner(agent.CircleCLIOnboardingRunnerConfig{
			CLIPath: cfg.CircleCLIPath,
			Chain:   cfg.CircleAgentWalletChain,
			Timeout: time.Duration(cfg.CircleAgentWalletTimeoutSeconds) * time.Second,
		}),
		RequestStore: agent.NewCircleOTPRequestStore(),
	}
	circleWalletResolver := agent.NewCircleCLIWalletResolver(agent.CircleCLIWalletResolverConfig{
		CLIPath: cfg.CircleCLIPath,
		Chain:   cfg.CircleAgentWalletChain,
		Timeout: time.Duration(cfg.CircleAgentWalletTimeoutSeconds) * time.Second,
	})

	circleFaucetRunner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled: cfg.CircleAgentWalletFaucetEnabled,
		CLIPath: cfg.CircleCLIPath,
		Chain:   cfg.CircleAgentWalletChain,
		Timeout: time.Duration(cfg.CircleAgentWalletTimeoutSeconds) * time.Second,
	})

	registerStatusRoutes(router, db)
	registerArcRoutes(router)
	registerAnalyticsRoutes(router, analyticsRepository)
	registerMarketRoutes(router, marketsRepository)
	registerTradeRoutes(router, tradesRepository, marketsRepository)
	registerPositionRoutes(router, positionsRepository)
	registerResolutionRoutes(router, resolutionsRepository)
	registerSettlementRoutes(router, settlementsRepository)
	registerAgentIntentRoutes(router, agentIntentStore, agentWalletsRepository, circleExecutor, agentSessionsRepository, circleOnboardingStarter, circleWalletResolver, circleBalanceReader, circleFaucetRunner, agentIntentsRepository)

	return router
}

func newCircleAgentWalletBalanceReader(cfg config.Config) agent.CircleAgentWalletBalanceReader {
	if cfg.CircleAgentWalletExecutor != "api" {
		return nil
	}
	timeout := time.Duration(cfg.CircleAgentWalletTimeoutSeconds) * time.Second
	reader, err := agent.NewCircleAPIBalanceReader(agent.CircleAPIBalanceReaderConfig{
		APIKey:  cfg.CircleAPIKey,
		BaseURL: cfg.CircleAPIBaseURL,
		Timeout: timeout,
	})
	if err != nil {
		return nil
	}
	return reader
}

func newCircleAgentWalletExecutor(cfg config.Config) agent.Executor {
	timeout := time.Duration(cfg.CircleAgentWalletTimeoutSeconds) * time.Second
	if cfg.CircleAgentWalletExecutor == "api" {
		provider, ok := newCircleEntitySecretCiphertextProvider(cfg, timeout)
		if !ok {
			return nil
		}
		executor, err := agent.NewCircleAPIExecutor(agent.CircleAPIExecutorConfig{
			Enabled:                        cfg.CircleAgentWalletExecutionEnabled,
			APIKey:                         cfg.CircleAPIKey,
			EntitySecretCiphertextProvider: provider,
			BaseURL:                        cfg.CircleAPIBaseURL,
			Timeout:                        timeout,
			AgentFactory:                   agent.AgentFactoryAddress,
		})
		if err == nil {
			return executor
		}
		return nil
	}
	return agent.NewCircleCLIExecutor(agent.CircleCLIExecutorConfig{
		Enabled:      cfg.CircleAgentWalletExecutionEnabled,
		CLIPath:      cfg.CircleCLIPath,
		Chain:        cfg.CircleAgentWalletChain,
		Timeout:      timeout,
		AgentFactory: agent.AgentFactoryAddress,
	})
}

func newCircleEntitySecretCiphertextProvider(cfg config.Config, timeout time.Duration) (circleapi.EntitySecretCiphertextProvider, bool) {
	if strings.TrimSpace(cfg.CircleEntitySecret) != "" {
		provider, err := circleapi.NewRawEntitySecretCiphertextProvider(circleapi.RawEntitySecretCiphertextProviderConfig{
			APIKey:          cfg.CircleAPIKey,
			BaseURL:         cfg.CircleAPIBaseURL,
			RawEntitySecret: cfg.CircleEntitySecret,
			Timeout:         timeout,
		})
		if err != nil {
			return nil, false
		}
		return provider, true
	}
	if circleStaticCiphertextBlockedInProduction(cfg) {
		return nil, false
	}
	if strings.TrimSpace(cfg.CircleStaticDevEntitySecretCiphertext) == "" {
		return nil, false
	}
	return circleapi.NewEnvEntitySecretCiphertextProvider(cfg.CircleStaticDevEntitySecretCiphertext), true
}

func circleStaticCiphertextBlockedInProduction(cfg config.Config) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "production") &&
		cfg.CircleAgentWalletExecutor == "api" &&
		strings.TrimSpace(cfg.CircleStaticDevEntitySecretCiphertext) != "" &&
		!cfg.CircleAllowStaticEntitySecretCiphertext
}
