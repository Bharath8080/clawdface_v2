package main

import (
	"backend/configs"
	"backend/providerselector"
	"backend/types"
	"backend/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type UserIdentifyRequest struct {
	UserID string                 `json:"userId"`
	Token  map[string]interface{} `json:"token"`
}

// SameOriginMiddleware checks if the request's Origin or Referer is in the allowed origins list.
func SameOriginMiddleware() func(http.Handler) http.Handler {
	allowedOrigins := getAllowedOrigins()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = r.Header.Get("Referer")
			}

			if origin != "" && !isAllowedOrigin(origin, allowedOrigins) {
				http.Error(w, "Forbidden: origin not allowed", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func StartCronJobs() {
	c := cron.New()

	// Run every day at midnight
	_, err := c.AddFunc("0 0 * * *", utils.ResetMonthlyCredits)
	if err != nil {
		log.Println("Failed to schedule cron:", err)
		return
	}

	c.Start()
	log.Println("Cron job started")
}

func StartCronJobs_AutoReload() {
	c := cron.New()

	// Run every 10 minutes
	_, err := c.AddFunc("*/10 * * * *", utils.CreateAutoReloadCredits)
	if err != nil {
		log.Println("Failed to schedule cron:", err)
		return
	}

	c.Start()
	log.Println("Cron job started for Auto Reload")
}

func StartCronJobs_RegionUpdate() {
	c := cron.New()

	// Run every 10 minutes
	_, err := c.AddFunc("*/10 * * * *", utils.UpdateRegionofConversations)
	if err != nil {
		log.Println("Failed to schedule cron for region update:", err)
		return
	}

	c.Start()
	log.Println("Cron job started for Region Update")
}

func getAllowedOrigins() []string {
	origins := configs.GetEnv("API_ALLOWED_ORIGINS") // e.g. "https://a.com,https://b.com"
	if origins == "" {
		return nil
	}
	parts := strings.Split(origins, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, ao := range allowed {
		if origin == ao {
			return true
		}
	}
	return false
}

func main() {

	//Monthly credit reset cron func
	StartCronJobs()

	//Auto Reload cron func
	StartCronJobs_AutoReload()

	//Region update cron func
	StartCronJobs_RegionUpdate()

	//Load Env variables
	configs.LoadEnv()

	// Initialize database
	dbDSN := configs.GetEnv("API_DB_DNSDETAILS")
	err := utils.InitDB(dbDSN)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = utils.DB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize provider health monitoring from YAML config.
	// If provider_health.yaml is absent or all providers are disabled the
	// checker starts with an empty provider list (no-op).
	providerHealthCfg, phErr := providerselector.LoadConfig("provider_health.yaml")
	if phErr != nil {
		log.Printf("[provider-health] config load error: %v — monitoring disabled", phErr)
		providerHealthCfg = &providerselector.Config{
			CheckInterval: 60 * time.Second,
			Ranking: providerselector.RankingConfig{
				LatencyWeight:      0.3,
				AvailabilityWeight: 0.7,
				WindowSize:         100,
			},
		}
	}
	providerHealthProviders, phBuildErr := providerselector.BuildProviders(providerHealthCfg)
	if phBuildErr != nil {
		log.Printf("[provider-health] build providers error: %v — monitoring disabled", phBuildErr)
		providerHealthProviders = nil
	}
	providerHealthChecker := providerselector.NewChecker(providerHealthProviders, providerHealthCfg)
	providerHealthChecker.Start(context.Background())
	providerInfraConcurrency := providerselector.BuildInfraConcurrencyManager(providerHealthCfg)
	providerInfraActiveCounters := providerselector.BuildInfraActiveRunCounters(providerHealthCfg)
	providerselector.StartReconciler(context.Background(), providerHealthCfg.CheckInterval, providerInfraConcurrency, providerInfraActiveCounters)
	providerHealthHandler := providerselector.NewHandler(providerHealthChecker, providerHealthCfg, providerInfraConcurrency, providerInfraActiveCounters)
	log.Printf("[provider-health] started with %d providers", len(providerHealthProviders))

	getAllowedAvatar := func(id string) (error, types.VisualConfig) {
		// List of all allowed avatars
		allowedAvatars := []types.VisualConfig{
			{
				ID:           "182b03e8",
				X:            710,
				Y:            20,
				Height:       1080,
				Width:        1920,
				TargetHeight: 560,
				TargetWidth:  560,
			},
			{
				ID:           "21ef04ad",
				X:            711,
				Y:            88,
				Height:       1080,
				Width:        1920,
				TargetHeight: 464,
				TargetWidth:  465,
			},
			{
				ID:           "17de03e4",
				X:            705,
				Y:            37,
				Height:       1080,
				Width:        1920,
				TargetHeight: 456,
				TargetWidth:  457,
			},
			{
				ID:           "1928040f",
				X:            747,
				Y:            104,
				Height:       1080,
				Width:        1920,
				TargetHeight: 496,
				TargetWidth:  496,
			},
			{
				ID:           "c5b563de",
				X:            662,
				Y:            18,
				Height:       1080,
				Width:        1920,
				TargetHeight: 536,
				TargetWidth:  536,
			},
			{
				ID:           "178303d3",
				X:            667,
				Y:            126,
				Height:       1080,
				Width:        1920,
				TargetHeight: 627,
				TargetWidth:  627,
			},
			{
				ID:           "05a001fc",
				X:            728,
				Y:            262,
				Height:       1080,
				Width:        1920,
				TargetHeight: 414,
				TargetWidth:  414,
			},
			{
				ID:           "be5b2ce0",
				X:            662,
				Y:            18,
				Height:       1080,
				Width:        1920,
				TargetHeight: 536,
				TargetWidth:  536,
			},
			{
				ID:           "0de70332",
				X:            667,
				Y:            126,
				Height:       1080,
				Width:        1920,
				TargetHeight: 627,
				TargetWidth:  627,
			},
			{
				ID:           "03ae0187",
				X:            728,
				Y:            262,
				Height:       1080,
				Width:        1920,
				TargetHeight: 414,
				TargetWidth:  414,
			},
			{
				ID:           "1fa504ff",
				X:            728,
				Y:            262,
				Height:       1080,
				Width:        1920,
				TargetHeight: 414,
				TargetWidth:  414,
			},
			{
				ID:           "7d881c1b",
				X:            705,
				Y:            37,
				Height:       1080,
				Width:        1920,
				TargetHeight: 456,
				TargetWidth:  457,
			},
			{
				ID:           "178803d6",
				X:            747,
				Y:            104,
				Height:       1080,
				Width:        1920,
				TargetHeight: 496,
				TargetWidth:  496,
			},
			{
				ID:           "1a640442",
				X:            662,
				Y:            18,
				Height:       1080,
				Width:        1920,
				TargetHeight: 536,
				TargetWidth:  536,
			},
			{
				ID:           "0f160301",
				X:            667,
				Y:            126,
				Height:       1080,
				Width:        1920,
				TargetHeight: 627,
				TargetWidth:  627,
			},
			{
				ID:           "057501e8",
				X:            728,
				Y:            262,
				Height:       1080,
				Width:        1920,
				TargetHeight: 414,
				TargetWidth:  414,
			},
			{
				ID:           "05b401f3",
				X:            662,
				Y:            18,
				Height:       1080,
				Width:        1920,
				TargetHeight: 536,
				TargetWidth:  536,
			},
			{
				ID:           "13550375",
				X:            667,
				Y:            126,
				Height:       1080,
				Width:        1920,
				TargetHeight: 627,
				TargetWidth:  627,
			},
			{
				ID:           "48d778c9",
				X:            728,
				Y:            262,
				Height:       1080,
				Width:        1920,
				TargetHeight: 414,
				TargetWidth:  414,
			},
			{
				ID:           "18c4043e",
				X:            728,
				Y:            262,
				Height:       1080,
				Width:        1920,
				TargetHeight: 414,
				TargetWidth:  414,
			}}
		for _, avatarConfig := range allowedAvatars {
			if avatarConfig.ID == id {
				return nil, avatarConfig
			}
		}
		return errors.New("Avatar not found"), types.VisualConfig{}
	}

	// Create a new router
	router := chi.NewRouter()
	// Use logging middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	// Add CORS middleware to handle OPTIONS requests
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Define routes
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Avatar Streaming API Running"))
	})

	router.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		utils.GetHealthCheck(w, r)
	})

	router.Post("/stripe/webhook", func(w http.ResponseWriter, r *http.Request) {
		utils.StripeWebhookHandler(w, r)
	})

	router.Get("/jobstatus/{jobId}", func(w http.ResponseWriter, r *http.Request) {
		jobId := chi.URLParam(r, "jobId")
		utils.GetJobStatus(w, r, jobId)
	})

	// Setup routes
	router.Route("/v1", func(r chi.Router) {
		log.Println("Registering /v1 routes:")

		r.Get("/client-ip", func(w http.ResponseWriter, r *http.Request) {
			utils.HandleGetClientIP(w, r)
		})

		// Usage Management Routes
		r.Route("/usage", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				var startUsageRequest types.StartUsageRequest
				err := json.NewDecoder(r.Body).Decode(&startUsageRequest)
				if err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				err = utils.HandleUsageCreation(startUsageRequest.ConversationId, startUsageRequest.Status, startUsageRequest.JobId, startUsageRequest.Usage, startUsageRequest.SessionId, startUsageRequest.ParticipantId, startUsageRequest.Message)
				if err != nil {
					http.Error(w, "Error creating usage: "+err.Error(), http.StatusInternalServerError)
					return
				}

				response := types.StartUsageResponse{
					Status: "Usage details updated",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})

			r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
				var startUsageRequest types.StartUsageRequest
				err := json.NewDecoder(r.Body).Decode(&startUsageRequest)
				if err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				err = utils.HandleUsageUpdate(startUsageRequest.ConversationId, startUsageRequest.Status, startUsageRequest.Message)
				if err != nil {
					http.Error(w, "Error updating usage: "+err.Error(), http.StatusInternalServerError)
					return
				}

				response := types.StartUsageResponse{
					Status: "Usage details updated",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})
		})
		//RAG document status update route
		r.Route("/ragdocument", func(r chi.Router) {
			r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
				var ragUpdateRequest types.RAGUpdateRequest
				err := json.NewDecoder(r.Body).Decode(&ragUpdateRequest)
				if err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				err = utils.UpdateRAGDocumentStatus(ragUpdateRequest.DocId, ragUpdateRequest.Status, ragUpdateRequest.Message)
				if err != nil {
					http.Error(w, "Error updating document status: "+err.Error(), http.StatusInternalServerError)
					return
				}

				response := types.StartUsageResponse{
					Status: "Document status updated",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})
		})
		r.Route("/organization", func(r chi.Router) {
			r.Get("/{orgId}/statistics", func(w http.ResponseWriter, r *http.Request) {
				orgId := chi.URLParam(r, "orgId")
				utils.GetOrganizationStatistics(w, r, orgId)
			})
		})
		r.Route("/global-metrics", func(r chi.Router) {
			r.Use(SameOriginMiddleware())
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				utils.GetGlobalMetrics(w, r)
			})
			r.Get("/conversation/{conversationId}", func(w http.ResponseWriter, r *http.Request) {
				conversationID := chi.URLParam(r, "conversationId")
				utils.GetGlobalMetricsConversationByID(w, r, conversationID)
			})
		})
		r.Route("/memory-category", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateMemoryCategory(w, r)
			})
			r.Get("/user/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleGetAllMemoryCategoriesByUser(w, userId)
			})
			r.Get("/{categoryId}", func(w http.ResponseWriter, r *http.Request) {
				categoryId := chi.URLParam(r, "categoryId")
				utils.HandleGetMemoryCategory(w, categoryId)
			})
			r.Put("/{categoryId}", func(w http.ResponseWriter, r *http.Request) {
				categoryId := chi.URLParam(r, "categoryId")
				utils.HandleUpdateMemoryCategory(w, r, categoryId)
			})
			r.Delete("/{categoryId}", func(w http.ResponseWriter, r *http.Request) {
				categoryId := chi.URLParam(r, "categoryId")
				utils.HandleDeleteMemoryCategory(w, categoryId)
			})
		})
		r.Route("/integrations-config", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateIntegrationConfig(w, r)
			})
			r.Get("/user/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleGetIntegrationConfigsByUser(w, userId)
			})
			r.Get("/{authConfigId}", func(w http.ResponseWriter, r *http.Request) {
				authConfigId := chi.URLParam(r, "authConfigId")
				utils.HandleGetIntegrationConfig(w, authConfigId)
			})
			r.Put("/{authConfigId}", func(w http.ResponseWriter, r *http.Request) {
				authConfigId := chi.URLParam(r, "authConfigId")
				utils.HandleUpdateIntegrationConfig(w, r, authConfigId)
			})
			r.Delete("/{authConfigId}", func(w http.ResponseWriter, r *http.Request) {
				authConfigId := chi.URLParam(r, "authConfigId")
				utils.HandleDeleteIntegrationConfig(w, authConfigId)
			})
		})
		r.Route("/memberinvite", func(r chi.Router) {
			r.Get("/{inviteId}", func(w http.ResponseWriter, r *http.Request) {
				inviteId := chi.URLParam(r, "inviteId")
				utils.HandleGetWorkspaceInvitationByID(w, inviteId)
			})
		})
		r.Route("/script-to-video", func(r chi.Router) {
			r.Use(utils.AuthMiddleware)
			r.Post("/createVideo", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleScriptToVideo(w, r)
			})
			r.Get("/genStatus/{genID}", func(w http.ResponseWriter, r *http.Request) {
				genID := chi.URLParam(r, "genID")
				utils.HandleGetGenStatus(w, r, genID)
			})
		})
		r.Route("/ext/script-to-video", func(r chi.Router) {
			// Internal infra callback — auth handled inside the handler via S2V_INFRA_AUTH_BEARER
			r.Post("/updateUsage", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleUpdateScriptToVideoUsage(w, r)
			})

		})
		r.Route("/presentations", func(r chi.Router) {
			// Internal — no auth, called by Lambda and ingestion pipeline
			r.Post("/resources/{resourceId}/status", func(w http.ResponseWriter, r *http.Request) {
				resourceId := chi.URLParam(r, "resourceId")
				utils.HandleUpdatePresentationResourceStatus(w, r, resourceId)
			})

			r.Group(func(r chi.Router) {
				r.Use(utils.AuthMiddleware)
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllPresentations(w, r)
				})
				r.Get("/tags", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllTagNames(w, r)
				})
				r.Post("/signed-url", func(w http.ResponseWriter, r *http.Request) {
					utils.HandlePresentationGetUploadURL(w, r)
				})
				r.Get("/org/{orgId}", func(w http.ResponseWriter, r *http.Request) {
					orgId := chi.URLParam(r, "orgId")
					utils.HandleGetPresentationsByOrg(w, r, orgId)
				})
				r.Get("/agent/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					agentId := chi.URLParam(r, "agentId")
					utils.HandleGetPresentationsByAgent(w, r, agentId)
				})
				r.Get("/{presentationId}", func(w http.ResponseWriter, r *http.Request) {
					presentationId := chi.URLParam(r, "presentationId")
					utils.HandleGetPresentation(w, r, presentationId)
				})
				r.Get("/{presentationId}/resources", func(w http.ResponseWriter, r *http.Request) {
					presentationId := chi.URLParam(r, "presentationId")
					utils.HandleGetPresentationResources(w, r, presentationId)
				})
				r.Post("/{presentationId}/resources", func(w http.ResponseWriter, r *http.Request) {
					presentationId := chi.URLParam(r, "presentationId")
					utils.HandleAddPresentationResources(w, r, presentationId)
				})
				r.Put("/{presentationId}", func(w http.ResponseWriter, r *http.Request) {
					presentationId := chi.URLParam(r, "presentationId")
					utils.HandleUpdatePresentation(w, r, presentationId)
				})
				r.Delete("/{presentationId}", func(w http.ResponseWriter, r *http.Request) {
					presentationId := chi.URLParam(r, "presentationId")
					utils.HandleDeletePresentation(w, r, presentationId)
				})
				r.Get("/resources/{resourceId}/status", func(w http.ResponseWriter, r *http.Request) {
					resourceId := chi.URLParam(r, "resourceId")
					utils.HandleGetPresentationResourceStatus(w, r, resourceId)
				})
			})
		})

		// Authentication endpoints (Internal)
		r.Route("/user", func(r chi.Router) {
			r.Post("/identify", func(w http.ResponseWriter, r *http.Request) {
				log.Println("=== Received authentication request ===")

				// Extract token from Authorization header
				authHeader := r.Header.Get("Authorization")

				if authHeader == "" {
					log.Println("Error: Missing Authorization header")
					http.Error(w, "Authorization header is missing", http.StatusBadRequest)
					return
				}

				// Check if the header has the Bearer prefix
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
					log.Println("Error: Invalid Authorization header format")
					http.Error(w, "Authorization header format must be Bearer {token}", http.StatusBadRequest)
					return
				}

				// Get the token from the header
				accessTokenStr := parts[1]
				if accessTokenStr == "" {
					log.Println("Error: Empty token provided")
					http.Error(w, "Empty token provided", http.StatusBadRequest)
					return
				}

				// Parse and log request body
				var req UserIdentifyRequest
				body, err := io.ReadAll(r.Body)
				if err == nil && len(body) > 0 {
					log.Printf("Request body: %s", string(body))

					if err := json.Unmarshal(body, &req); err == nil {
						log.Printf("User ID from request: %s", req.UserID)

						if req.Token != nil {
							log.Printf("Token object from request body: %+v", req.Token)

							if accessToken, ok := req.Token["accessToken"]; ok {
								log.Printf("Access Token from body: %v", accessToken)
							}

							if refreshToken, ok := req.Token["refreshToken"]; ok {
								log.Printf("Refresh Token from body: %v", refreshToken)
							}
						}
					} else {
						log.Printf("Error parsing request body: %v", err)
					}
				}

				// Validate the token
				claims, err := utils.ValidateToken(accessTokenStr)
				if err != nil {
					// Check if it's a token expired error
					var expiredErr *utils.TokenExpiredError
					if errors.As(err, &expiredErr) {
						log.Printf("Token expired at %v, redirecting to refresh", expiredErr.ExpiredAt)
						http.Error(w, fmt.Sprintf("Unauthorized: %v", expiredErr), http.StatusUnauthorized)
						return
					}

					log.Printf("Token validation failed: %v", err)
					http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
					return
				}

				// Log token claims information
				log.Printf("Token validated successfully:")
				log.Printf("  - Subject (User ID): %s", claims.Subject)
				log.Printf("  - Issued At: %v", claims.IssuedAt.Time())
				log.Printf("  - Expires At: %v", claims.Expiry.Time())

				// If user ID is provided in request, validate it matches token
				if req.UserID != "" && req.UserID != claims.Subject {
					log.Printf("User ID mismatch: %s (request) vs %s (token)", req.UserID, claims.Subject)
					http.Error(w, "User ID mismatch between token and request", http.StatusUnauthorized)
					return
				}

				// Respond with success and include user info
				response := map[string]interface{}{
					"success": true,
					"message": "User authenticated successfully",
					"userId":  claims.Subject,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})

			r.Get("/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleGetUser(w, userId)
			})

			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateUser(w, r)
			})

			r.Put("/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleUpdateUser(w, r, userId)
			})

			r.Put("/unsubscribe/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleUnSubscribeUser(w, r, userId)
			})

			r.Get("/favourite/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleGetUserFavourite(w, userId)
			})

			r.Put("/favourite/{userId}", func(w http.ResponseWriter, r *http.Request) {
				userId := chi.URLParam(r, "userId")
				utils.HandleUpdateUserFavourite(w, r, userId)
			})
		})

		// Token refresh endpoint (Internal)
		r.Post("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				RefreshToken string `json:"refreshToken"`
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				http.Error(w, "Error reading body", http.StatusInternalServerError)
				return
			}

			log.Printf("Raw refresh request body: %s", string(body))

			if err := json.Unmarshal(body, &req); err != nil {
				log.Printf("Error unmarshaling request: %v", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if req.RefreshToken == "" {
				log.Println("Error: Missing refresh token")
				http.Error(w, "Refresh token is required", http.StatusBadRequest)
				return
			}

			log.Printf("Refresh Token: %s", req.RefreshToken)

			// Validate the refresh token
			claims, err := utils.ValidateToken(req.RefreshToken)
			if err != nil {
				log.Printf("Refresh token validation failed: %v", err)
				http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
				return
			}

			log.Printf("Refresh token valid for user: %s", claims.Subject)

			// In a real implementation, you would generate these:
			newAccessToken := "new-access-token"
			newRefreshToken := "new-refresh-token"

			log.Printf("Generated new access token: %s", newAccessToken)
			log.Printf("Generated new refresh token: %s", newRefreshToken)

			response := map[string]interface{}{
				"success":      true,
				"message":      "Token refreshed successfully",
				"userId":       claims.Subject,
				"accessToken":  newAccessToken,
				"refreshToken": newRefreshToken,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			log.Println("=== Token refresh request completed ===")
		})

		// Conversation Management Routes
		r.Route("/public/conversation", func(r chi.Router) {
			//r.Use(SameOriginMiddleware())
			r.Post("/feedback", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateFeedback(w, r)
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				apiKey := ""
				apiKeyId := ""
				var startConversationRequest types.StartConversationRequest
				err := json.NewDecoder(r.Body).Decode(&startConversationRequest)
				if err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				roomId, err := utils.GenerateRoomId()
				if err != nil {
					http.Error(w, "Failed to generate room ID", http.StatusInternalServerError)
					return
				}

				conversationId, url, agent, err := utils.HandleConversationCreation(apiKey, startConversationRequest.AgentId, apiKeyId, startConversationRequest.Context, startConversationRequest.MetaData, startConversationRequest.UserName, startConversationRequest.UserId, roomId, startConversationRequest.MeetingURL, false, startConversationRequest.AvatarId, startConversationRequest.Mode)
				if err != nil {
					http.Error(w, "Error creating conversation: "+err.Error(), http.StatusInternalServerError)
					return
				}

				token := utils.GetLiveKitJoinToken(utils.TokenSourceRequest{
					RoomName:            roomId,
					ParticipantName:     startConversationRequest.UserName,
					ParticipantIdentity: startConversationRequest.UserId,
				})

				response := types.StartConversationResponse{
					ConversationId: conversationId,
					URL:            url,
					LivekitURL:     configs.GetEnv("LIVEKIT_URL"),
					Token:          token,
					ImageURL:       agent.AvatarImageURL,
					Name:           agent.AvatarName,
					ID:             agent.AvatarID,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})
			r.Post("/byemail", func(w http.ResponseWriter, r *http.Request) {
				apiKey := ""
				apiKeyId := ""
				var startConversationRequest types.StartConversationRequest
				err := json.NewDecoder(r.Body).Decode(&startConversationRequest)
				if err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				agentId, errc := utils.HandleGetAgentByEmail(startConversationRequest.Email)
				if errc != nil {
					http.Error(w, "Error retriving Agent details by Email", http.StatusBadRequest)
					return
				}

				conversationId, url, agent, err := utils.HandleConversationCreation(apiKey, agentId, apiKeyId, startConversationRequest.Context, startConversationRequest.MetaData, startConversationRequest.UserName, startConversationRequest.UserId, startConversationRequest.RoomId, startConversationRequest.MeetingURL, true, startConversationRequest.AvatarId, startConversationRequest.Mode)
				if err != nil {
					http.Error(w, "Error creating conversation: "+err.Error(), http.StatusInternalServerError)
					log.Printf("Error creating conversation: %s", agent.AvatarID)
					return
				}

				response := types.StartConversationResponse{
					ConversationId: conversationId,
					URL:            url,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})
			r.Get("/{conversationId}", func(w http.ResponseWriter, r *http.Request) {
				conversationId := chi.URLParam(r, "conversationId")
				utils.HandleGetConversationStatus(w, conversationId)
			})
		})

		// Widget conversation route
		r.Route("/public/widget/conversation", func(r chi.Router) {
			// browser preligh
			r.Options("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
				w.WriteHeader(http.StatusOK)
			})

			r.With(utils.AuthMiddleware).Post("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				apiKey := r.Header.Get("X-API-Key")
				apiKeyId, ok := r.Context().Value("apiKeyId").(string)
				if !ok || apiKeyId == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				var startConversationRequest types.StartConversationRequest
				err := json.NewDecoder(r.Body).Decode(&startConversationRequest)
				if err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				roomId, err := utils.GenerateRoomId()
				if err != nil {
					http.Error(w, "Failed to generate room ID", http.StatusInternalServerError)
					return
				}

				conversationId, url, agent, err := utils.HandleConversationCreation(apiKey, startConversationRequest.AgentId, apiKeyId, startConversationRequest.Context, startConversationRequest.MetaData, startConversationRequest.UserName, startConversationRequest.UserId, roomId, startConversationRequest.MeetingURL, false, startConversationRequest.AvatarId, startConversationRequest.Mode)
				if err != nil {
					http.Error(w, "Error creating conversation: "+err.Error(), http.StatusInternalServerError)
					return
				}

				token := utils.GetLiveKitJoinToken(utils.TokenSourceRequest{
					RoomName:            roomId,
					ParticipantName:     startConversationRequest.UserName,
					ParticipantIdentity: startConversationRequest.UserId,
				})

				response := types.StartConversationResponse{
					ConversationId: conversationId,
					URL:            url,
					LivekitURL:     configs.GetEnv("LIVEKIT_URL"),
					Token:          token,
					ImageURL:       agent.AvatarImageURL,
					Name:           agent.AvatarName,
					ID:             agent.AvatarID,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			})
		})

		// Agent Management Routes
		r.Route("/public/agent", func(r chi.Router) {
			r.Use(SameOriginMiddleware())
			r.Get("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
				agentId := chi.URLParam(r, "agentId")
				utils.HandleGetAgentStatus(w, agentId)
			})
		})

		// Agent Management Routes
		r.Route("/public/orgbyagent", func(r chi.Router) {
			// r.Use(SameOriginMiddleware())
			r.Get("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
				agentId := chi.URLParam(r, "agentId")
				utils.HandleGetOrganizationByAgentID(w, r, agentId)
			})
		})
		// Agent Management Routes
		r.Route("/public/orgbyagent1", func(r chi.Router) {
			// r.Use(SameOriginMiddleware())
			r.Get("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
				agentId := chi.URLParam(r, "agentId")
				utils.HandleGetOrganizationByAgentID1(w, r, agentId)
			})
		})

		// Avatars by AgentId
		r.Route("/public/avatar", func(r chi.Router) {
			r.Use(SameOriginMiddleware())
			r.Get("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
				agentId := chi.URLParam(r, "agentId")
				utils.HandleGetAvatarDetailsByAgent(w, agentId)
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleGetAllAvatarsConfigs(w)
			})
		})

		// API Management Routes (Internal)
		r.Route("/apikey", func(r chi.Router) {
			r.Use(utils.JWTAuthMiddleware)
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleGetAPIKeys(w, r)
			})

			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateAPIKey(w, r)
			})

			r.Put("/{apiKeyId}", func(w http.ResponseWriter, r *http.Request) {
				apiKeyId := chi.URLParam(r, "apiKeyId")
				utils.HandleUpdateAPIKey(w, r, apiKeyId)
			})

			r.Delete("/{apiKeyId}", func(w http.ResponseWriter, r *http.Request) {
				apiKeyId := chi.URLParam(r, "apiKeyId")
				utils.HandleDeleteAPIKey(w, apiKeyId)
			})
		})

		// Inbox Routes (Internal)
		r.Route("/inbox", func(r chi.Router) {
			r.Use(utils.AuthMiddleware)

			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateInbox(w, r)
			})

			r.Delete("/{inboxId}", func(w http.ResponseWriter, r *http.Request) {
				inboxId := chi.URLParam(r, "inboxId")
				utils.HandleDeleteInbox(w, inboxId)
			})
		})

		r.Route("/purchase", func(r chi.Router) {
			r.Use(utils.AuthMiddleware)
			r.Get("/issue", func(w http.ResponseWriter, r *http.Request) {
				utils.GetPaymentFailureStatus(w, r)
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleCreateStripeCheckout(w, r)
			})

			r.Get("/plans", func(w http.ResponseWriter, r *http.Request) {
				utils.GetAllPurchasePlans(w, r)
			})

			r.Get("/history", func(w http.ResponseWriter, r *http.Request) {
				utils.GetAllPurchaseLogs(w, r)
			})

			r.Get("/details", func(w http.ResponseWriter, r *http.Request) {
				utils.GetSubscriptionDetails(w, r)
			})

			r.Post("/plan-price", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleGetPlanPrice(w, r)
			})

			r.Get("/manage-subscription", func(w http.ResponseWriter, r *http.Request) {
				utils.ManageSubscriptionPortal(w, r)
			})

			r.Post("/auto-reload", func(w http.ResponseWriter, r *http.Request) {
				utils.HandleAutoReloadUpdate(w, r)
			})
		})

		r.Route("/ext/purchase", func(r chi.Router) {
			r.Get("/plans", func(w http.ResponseWriter, r *http.Request) {
				utils.GetAllPurchasePlans(w, r)
			})
		})

		r.Group(func(r chi.Router) {
			// Apply auth middleware to all routes in this group
			r.Use(utils.AuthMiddleware)

			// API Routes
			r.Route("/ext", func(r chi.Router) {
				r.Post("/agentbytemplate", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateAgentConfigByTemplate(w, r, apikeyId)
				})
				r.Post("/agent", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateAgentAPI(w, r, apikeyId)
				})
				r.Get("/agent/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					agentId := chi.URLParam(r, "agentId")
					utils.HandleGetAgentConfigAPI(w, agentId, apikeyId)
				})
				r.Get("/agents", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllAgentConfigsAPI(w, apikeyId)
				})
				r.Put("/agent/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					agentId := chi.URLParam(r, "agentId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateAgentConfig(w, r, agentId, apikeyId)
				})
				r.Delete("/agent/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					agentId := chi.URLParam(r, "agentId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteAgentConfig(w, agentId, apikeyId)
				})

				r.Get("/conversation/{conversationId}", func(w http.ResponseWriter, r *http.Request) {
					apiKeyId := r.Context().Value("apiKeyId").(string)
					conversationId := chi.URLParam(r, "conversationId")
					utils.HandleGetConversationAPI(w, conversationId, apiKeyId)
				})
				r.Get("/conversations/agent/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					apiKeyId := r.Context().Value("apiKeyId").(string)
					agentId := chi.URLParam(r, "agentId")
					utils.HandleGetConversationsByAgentAPI(w, agentId, apiKeyId)
				})

				r.Get("/templates", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllTemplateConfigsAPI(w, apikeyId)
				})
				r.Get("/template/{templateId}", func(w http.ResponseWriter, r *http.Request) {
					templateId := chi.URLParam(r, "templateId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetTemplateConfigAPI(w, templateId, apikeyId)
				})
				r.Post("/template", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateTemplateConfig(w, r, apikeyId)
				})
				r.Put("/template/{templateId}", func(w http.ResponseWriter, r *http.Request) {
					templateId := chi.URLParam(r, "templateId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateTemplateConfig(w, r, templateId, apikeyId)
				})
				r.Delete("/template/{templateId}", func(w http.ResponseWriter, r *http.Request) {
					templateId := chi.URLParam(r, "templateId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteTemplateConfig(w, templateId, apikeyId)
				})

				r.Get("/kbs", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllKBConfigsAPIRAG(w, apikeyId)
				})
				r.Get("/kb/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					kbId := chi.URLParam(r, "kbId")
					utils.HandleGetKBConfigAPIRAG(w, kbId, apikeyId)
				})
				r.Post("/kb", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateKBConfigRAG(w, r, apikeyId)
				})
				r.Post("/kb/{kbId}/doc", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					kbId := chi.URLParam(r, "kbId")
					utils.HandleAddDocToKBConfigRAG(w, r, kbId, apikeyId)
				})
				r.Put("/kb/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					kbId := chi.URLParam(r, "kbId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateKBConfigRAG(w, r, kbId, apikeyId)
				})
				r.Delete("/kb/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					kbId := chi.URLParam(r, "kbId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteKBConfigRAG(w, kbId, apikeyId)
				})
				r.Delete("/kb/doc/{docId}", func(w http.ResponseWriter, r *http.Request) {
					docId := chi.URLParam(r, "docId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteDocToKBRAG(w, docId, apikeyId)
				})

				r.Get("/tool", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllToolConfigs(w, apikeyId)
				})
				r.Get("/tool/{toolId}", func(w http.ResponseWriter, r *http.Request) {
					toolId := chi.URLParam(r, "toolId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetToolConfig(w, toolId, apikeyId)
				})
				r.Post("/tool", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateToolConfig(w, r, apikeyId)
				})
				r.Put("/tool/{toolId}", func(w http.ResponseWriter, r *http.Request) {
					toolId := chi.URLParam(r, "toolId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateToolConfig(w, r, toolId, apikeyId)
				})
				r.Delete("/tool/{toolId}", func(w http.ResponseWriter, r *http.Request) {
					toolId := chi.URLParam(r, "toolId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteToolConfig(w, toolId, apikeyId)
				})

				r.Get("/mcp", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllMCPConfigs(w, apikeyId)
				})
				r.Get("/mcp/{mcpId}", func(w http.ResponseWriter, r *http.Request) {
					mcpId := chi.URLParam(r, "mcpId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetMCPConfig(w, mcpId, apikeyId)
				})
				r.Post("/mcp", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateMCPConfig(w, r, apikeyId)
				})
				r.Put("/mcp/{mcpId}", func(w http.ResponseWriter, r *http.Request) {
					mcpId := chi.URLParam(r, "mcpId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateMCPConfig(w, r, mcpId, apikeyId)
				})
				r.Delete("/mcp/{mcpId}", func(w http.ResponseWriter, r *http.Request) {
					mcpId := chi.URLParam(r, "mcpId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteMCPConfig(w, mcpId, apikeyId)
				})

				r.Get("/avatars", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllAvatarsConfigsAPI(w)
				})
				r.Get("/avatars/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					agentId := chi.URLParam(r, "agentId")
					utils.HandleGetAvatarDetailsByAgent(w, agentId)
				})

				r.Get("/providers", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllProviderConfigsAPI(w)
				})

				// call for External meeting
				r.Post("/externalmeeting", func(w http.ResponseWriter, r *http.Request) {

					roomId, err := utils.GenerateRoomId()
					if err != nil {
						http.Error(w, "Failed to generate room ID", http.StatusInternalServerError)
						return
					}

					//apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleJoinExternalMeeting(w, r, roomId)
				})
			})

			// Protected endpoint example
			r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value("userID").(string)
				log.Printf("Accessed protected endpoint by user: %s", userID)

				response := map[string]interface{}{
					"message": "This is protected data",
					"userId":  userID,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			})

			// Usage Details Routes
			r.Route("/usages", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apiKey := r.Header.Get("X-API-Key")
					utils.HandleGetUsageMetricsByKeyHash(w, apiKey)
				})
			})

			// Avatar Management Routes
			r.Route("/avatar", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllAvatarConfigs(w)
				})

				r.Get("/{avatarId}", func(w http.ResponseWriter, r *http.Request) {
					avatarId := chi.URLParam(r, "avatarId")
					utils.HandleGetAvatarConfig(w, avatarId)
				})
			})

			// Role Management Routes
			r.Route("/role", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllActiveRoles(w)
				})
			})

			// Knowledge Base Routes
			r.Route("/knowledgebaseold", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllKBConfigs(w, apikeyId)
				})

				r.Get("/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					kbId := chi.URLParam(r, "kbId")
					utils.HandleGetKBConfig(w, kbId, apikeyId)
				})

				r.Post("/presign", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleURLPreSign(w, r)
				})

				r.Post("/fetchcontent", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleURLContent(w, r)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateKBConfig(w, r, apikeyId)
				})

				r.Post("/{kbId}/doc", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					kbId := chi.URLParam(r, "kbId")
					utils.HandleAddDocToKBConfig(w, r, kbId, apikeyId)
				})

				r.Put("/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					kbId := chi.URLParam(r, "kbId")
					utils.HandleUpdateKBConfig(w, r, kbId)
				})

				r.Delete("/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					kbId := chi.URLParam(r, "kbId")
					utils.HandleDeleteKBConfig(w, kbId)
				})

				r.Delete("/doc/{docId}", func(w http.ResponseWriter, r *http.Request) {
					docId := chi.URLParam(r, "docId")
					utils.HandleDeleteDocToKB(w, docId)
				})
			})

			// KnowledgeBase RAG Routes
			r.Route("/knowledgebase", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllKBConfigsRAG(w, apikeyId)
				})

				r.Get("/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					kbId := chi.URLParam(r, "kbId")
					utils.HandleGetKBConfigRAG(w, kbId, apikeyId)
				})

				r.Post("/presign", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleURLPreSign(w, r)
				})

				r.Post("/fetchcontent", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleURLContent(w, r)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateKBConfigRAG(w, r, apikeyId)
				})

				r.Post("/{kbId}/doc", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					kbId := chi.URLParam(r, "kbId")
					utils.HandleAddDocToKBConfigRAG(w, r, kbId, apikeyId)
				})

				r.Put("/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					kbId := chi.URLParam(r, "kbId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateKBConfigRAG(w, r, kbId, apikeyId)
				})

				r.Delete("/{kbId}", func(w http.ResponseWriter, r *http.Request) {
					kbId := chi.URLParam(r, "kbId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteKBConfigRAG(w, kbId, apikeyId)
				})

				r.Delete("/doc/{docId}", func(w http.ResponseWriter, r *http.Request) {
					docId := chi.URLParam(r, "docId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteDocToKBRAG(w, docId, apikeyId)
				})
			})

			// Knowledge Base Routes
			r.Route("/supportticket", func(r chi.Router) {
				r.Get("/{userId}", func(w http.ResponseWriter, r *http.Request) {
					userId := chi.URLParam(r, "userId")
					utils.HandleGetAllSupportTickets(w, r, userId)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleCreateSupportTicket(w, r)
				})
			})

			// Agent Management Routes
			r.Route("/agent", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllAgentConfigs(w, apikeyId)
				})

				r.Get("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					agentId := chi.URLParam(r, "agentId")
					utils.HandleGetAgentConfig(w, agentId, apikeyId)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateAgentConfig(w, r, apikeyId)
				})

				r.Post("/review", func(w http.ResponseWriter, r *http.Request) {
					//apikeyId := r.Context().Value("apiKeyId").(string)
					utils.CreateAgentAndStartConversation(w, r)
				})

				r.Put("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					agentId := chi.URLParam(r, "agentId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateAgentConfig(w, r, agentId, apikeyId)
				})

				r.Delete("/{agentId}", func(w http.ResponseWriter, r *http.Request) {
					agentId := chi.URLParam(r, "agentId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteAgentConfig(w, agentId, apikeyId)
				})
			})

			// Workspace Management Routes
			r.Route("/workspace", func(r chi.Router) {
				r.Get("/user/{userId}", func(w http.ResponseWriter, r *http.Request) {
					userId := chi.URLParam(r, "userId")
					utils.HandleGetAllWorkspaces(w, userId)
				})

				r.Get("/{workspaceId}", func(w http.ResponseWriter, r *http.Request) {
					workspaceId := chi.URLParam(r, "workspaceId")
					utils.HandleGetWorkspace(w, workspaceId)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleCreateWorkspace(w, r)
				})

				r.Put("/{workspaceId}", func(w http.ResponseWriter, r *http.Request) {
					workspaceId := chi.URLParam(r, "workspaceId")
					utils.HandleUpdateWorkspace(w, r, workspaceId)
				})

				r.Delete("/{workspaceId}", func(w http.ResponseWriter, r *http.Request) {
					workspaceId := chi.URLParam(r, "workspaceId")
					utils.HandleDeleteWorkspace(w, workspaceId)
				})
				r.Post("/invite", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleWorkspaceUserInvitation(w, r)
				})

				r.Get("/invite/{workspaceId}", func(w http.ResponseWriter, r *http.Request) {
					workspaceId := chi.URLParam(r, "workspaceId")
					utils.HandleGetAllWorkspaceInvitations(w, workspaceId)
				})

				r.Get("/member/{workspaceId}", func(w http.ResponseWriter, r *http.Request) {
					workspaceId := chi.URLParam(r, "workspaceId")
					utils.HandleListWorkspaceMembers(w, r, workspaceId)
				})
				r.Put("/member/{memberId}", func(w http.ResponseWriter, r *http.Request) {
					memberId := chi.URLParam(r, "memberId")
					utils.HandleUpdateWorkspaceMember(w, r, memberId)
				})

				r.Delete("/member/{memberId}", func(w http.ResponseWriter, r *http.Request) {
					memberId := chi.URLParam(r, "memberId")
					utils.HandleDeleteWorkspaceMember(w, r, memberId)
				})
			})

			// Template Management Routes
			r.Route("/template", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllTemplateConfigs(w, apikeyId)
				})

				r.Get("/{templateId}", func(w http.ResponseWriter, r *http.Request) {
					templateId := chi.URLParam(r, "templateId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetTemplateConfig(w, templateId, apikeyId)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateTemplateConfig(w, r, apikeyId)
				})

				r.Put("/{templateId}", func(w http.ResponseWriter, r *http.Request) {
					templateId := chi.URLParam(r, "templateId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateTemplateConfig(w, r, templateId, apikeyId)
				})

				r.Delete("/{templateId}", func(w http.ResponseWriter, r *http.Request) {
					templateId := chi.URLParam(r, "templateId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteTemplateConfig(w, templateId, apikeyId)
				})
			})

			// Tools Management Routes
			r.Route("/tool", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllToolConfigs(w, apikeyId)
				})

				r.Get("/{toolId}", func(w http.ResponseWriter, r *http.Request) {
					toolId := chi.URLParam(r, "toolId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetToolConfig(w, toolId, apikeyId)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateToolConfig(w, r, apikeyId)
				})

				r.Put("/{toolId}", func(w http.ResponseWriter, r *http.Request) {
					toolId := chi.URLParam(r, "toolId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateToolConfig(w, r, toolId, apikeyId)
				})

				r.Delete("/{toolId}", func(w http.ResponseWriter, r *http.Request) {
					toolId := chi.URLParam(r, "toolId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteToolConfig(w, toolId, apikeyId)
				})
			})

			// MCP Management Routes
			r.Route("/mcp", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllMCPConfigs(w, apikeyId)
				})

				r.Get("/{mcpId}", func(w http.ResponseWriter, r *http.Request) {
					mcpId := chi.URLParam(r, "mcpId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetMCPConfig(w, mcpId, apikeyId)
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleCreateMCPConfig(w, r, apikeyId)
				})

				r.Put("/{mcpId}", func(w http.ResponseWriter, r *http.Request) {
					mcpId := chi.URLParam(r, "mcpId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleUpdateMCPConfig(w, r, mcpId, apikeyId)
				})

				r.Delete("/{mcpId}", func(w http.ResponseWriter, r *http.Request) {
					mcpId := chi.URLParam(r, "mcpId")
					apikeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleDeleteMCPConfig(w, mcpId, apikeyId)
				})
			})

			// Provider Management Routes
			r.Route("/provider", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleGetAllProviderConfigs(w)
				})
			})

			// Conversation Management Routes
			r.Route("/conversation", func(r chi.Router) {
				r.Post("/feedback", func(w http.ResponseWriter, r *http.Request) {
					utils.HandleCreateFeedback(w, r)
				})
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apiKey := r.Header.Get("X-API-Key")
					var startConversationRequest types.StartConversationRequest
					err := json.NewDecoder(r.Body).Decode(&startConversationRequest)
					if err != nil {
						http.Error(w, "Invalid request body", http.StatusBadRequest)
						return
					}

					roomId, err := utils.GenerateRoomId()
					if err != nil {
						http.Error(w, "Failed to generate room ID", http.StatusInternalServerError)
						return
					}

					apiKeyId := r.Context().Value("apiKeyId").(string)
					conversationId, url, agent, err := utils.HandleConversationCreation(apiKey, startConversationRequest.AgentId, apiKeyId, startConversationRequest.Context, startConversationRequest.MetaData, startConversationRequest.UserName, startConversationRequest.UserId, roomId, startConversationRequest.MeetingURL, false, startConversationRequest.AvatarId, startConversationRequest.Mode)
					if err != nil {
						http.Error(w, "Error creating conversation: "+err.Error(), http.StatusInternalServerError)
						return
					}

					token := utils.GetLiveKitJoinToken(utils.TokenSourceRequest{
						RoomName:            roomId,
						ParticipantName:     startConversationRequest.UserName,
						ParticipantIdentity: startConversationRequest.UserId,
					})

					response := types.StartConversationResponse{
						ConversationId: conversationId,
						URL:            url,
						LivekitURL:     configs.GetEnv("LIVEKIT_URL"),
						Token:          token,
						ImageURL:       agent.AvatarImageURL,
						Name:           agent.AvatarName,
						ID:             agent.AvatarID,
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				})
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					apiKeyId := r.Context().Value("apiKeyId").(string)
					utils.HandleGetAllConversations(w, apiKeyId)
				})

				r.Get("/{conversationId}", func(w http.ResponseWriter, r *http.Request) {
					apiKeyId := r.Context().Value("apiKeyId").(string)
					conversationId := chi.URLParam(r, "conversationId")
					utils.HandleGetConversation(w, conversationId, apiKeyId)
				})
				r.Put("/{conversationId}/snippet", func(w http.ResponseWriter, r *http.Request) {
					conversationId := chi.URLParam(r, "conversationId")
					utils.HandleUpdateConversationSnippet(w, r, conversationId)
				})
				r.Put("/{conversationId}/chat", func(w http.ResponseWriter, r *http.Request) {
					conversationId := chi.URLParam(r, "conversationId")
					utils.HandleUpdateConversationChat(w, r, conversationId)
				})
				r.Put("/{conversationId}/summary", func(w http.ResponseWriter, r *http.Request) {
					conversationId := chi.URLParam(r, "conversationId")
					utils.HandleUpdateConversationSummary(w, r, conversationId)
				})
				r.Delete("/{conversationId}/chat", func(w http.ResponseWriter, r *http.Request) {
					conversationId := chi.URLParam(r, "conversationId")
					utils.HandleDeleteConversationChat(w, conversationId)
				})
			})

			// New create voice to video avatar session
			r.Route("/sessions", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apiKey := r.Header.Get("X-API-Key")

					// Parse Payload
					var requestBody struct {
						AvatarID     string `json:"avatar_id"`
						LivekitURL   string `json:"livekit_url"`
						LivekitToken string `json:"livekit_token"`
					}
					if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
						log.Printf("Error decoding request body: %v", err)
						http.Error(w, "Invalid request body", http.StatusBadRequest)
						return
					}

					conversationId := uuid.New().String()

					// Validate and get avatar configuration
					err, avatarConfig := getAllowedAvatar(requestBody.AvatarID)
					if err != nil {
						log.Printf("Avatar not found. %v", err)
						http.Error(w, "Avatar not found.", http.StatusNotFound)
						return
					}
					log.Printf("Avatar. %v", avatarConfig)
					log.Printf("Conversation. %v", conversationId)
					// Post new request to Infra Provider
					// Runpod
					var lk types.LivekitConfig
					lk.URL = requestBody.LivekitURL
					lk.Token = requestBody.LivekitToken
					var sessionInput types.SessionInput
					sessionInput.ConversationID = conversationId
					sessionInput.Avatar = avatarConfig
					sessionInput.Livekit = lk
					var sessionRequest types.SessionRequest
					sessionRequest.Input = sessionInput
					err, resp := utils.StartVoiceSession(sessionRequest)
					if err != nil {
						log.Fatal(err)
						log.Printf("Unable to create session: %v", err)
						http.Error(w, "Unable to create session.", http.StatusNotFound)
						return
					}

					conversationId, err = utils.VoicetoVideoSessionCreation(apiKey, requestBody.AvatarID, conversationId)
					if err != nil {
						http.Error(w, "Error creating voice to video session: "+err.Error(), http.StatusInternalServerError)
						delay := time.Duration(1) * time.Minute
						utils.ScheduleOneTimeResetCreditJob(delay, conversationId)

						// Cancel the session creation in case of failure
						utils.CancelVoiceSession(resp.RunID)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode("message: Voice to Video Session Created Successfully")
				})
			})

			// New create voice to video avatar session
			r.Route("/sessionslab", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					apiKey := r.Header.Get("X-API-Key")

					// Parse Payload
					var requestBody struct {
						AvatarID string `json:"avatar_id"`
					}
					if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
						log.Printf("Error decoding request body: %v", err)
						http.Error(w, "Invalid request body", http.StatusBadRequest)
						return
					}

					conversationId := uuid.New().String()
					conversationId, err = utils.VoicetoVideoSessionCreation(apiKey, requestBody.AvatarID, conversationId)
					if err != nil {
						http.Error(w, "Error creating voice to video session: "+err.Error(), http.StatusInternalServerError)
						return
					}

					// Validate and get avatar configuration
					err, avatarConfig := getAllowedAvatar(requestBody.AvatarID)
					if err != nil {
						log.Printf("Avatar not found. %v", err)
						http.Error(w, "Avatar not found.", http.StatusNotFound)
						return
					}
					log.Printf("Avatar. %v", avatarConfig)
					log.Printf("Conversation. %v", conversationId)
					// Post new request to 11 Labs
					var sessionRequest types.SessionRequest11Lab
					sessionRequest.UserId = apiKey
					sessionRequest.AvatarId = requestBody.AvatarID
					err, resp := utils.StartVoiceSession11Lab(sessionRequest)
					if err != nil {
						log.Fatal(err)
						log.Printf("Unable to create session: %v", resp)
						http.Error(w, "Unable to create session.", http.StatusNotFound)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(resp)
				})
			})
		})
	})
	// Provider health monitoring routes (no auth required — internal/ops use)
	router.Route("/api/v1/provider-health", func(r chi.Router) {
		r.Get("/ranked", providerHealthHandler.HandleRanked)
		r.Get("/best", providerHealthHandler.HandleBest)
		r.Get("/status", providerHealthHandler.HandleStatus)
		r.Post("/report", providerHealthHandler.HandleReport)
		r.Get("/dashboard", providerHealthHandler.HandleDashboard)
		// Infra concurrency tracking
		r.Get("/infra/concurrency", providerHealthHandler.HandleInfraConcurrency)
		r.Post("/infra/request/start", providerHealthHandler.HandleInfraRequestStart)
		r.Post("/infra/request/end", providerHealthHandler.HandleInfraRequestEnd)
		r.Get("/infra/concurrency/check", providerHealthHandler.HandleInfraConcurrencyCheck)
	})

	// Start the server
	log.Println("Server running on port 3077")
	if err := http.ListenAndServe(":3077", router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
