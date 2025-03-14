package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webblueprint/internal/api"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/nodes"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/nodes/math"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/nodes/web"
	"webblueprint/internal/registry"
	"webblueprint/pkg/db"
	"webblueprint/pkg/repository"

	"github.com/gorilla/mux"
)

func main() {
	// Komut satırı bayrakları tanımla
	useActorSystem := flag.Bool("actor", true, "Çalıştırma için aktör sistemini kullan (varsayılan: true)")
	useDatabase := flag.Bool("db", true, "Veritabanı depolamasını kullan (varsayılan: true)")
	port := flag.String("port", "8089", "Sunucu portu (varsayılan: 8089 veya $PORT çevre değişkeni)")
	flag.Parse()

	_ = useActorSystem
	_ = useDatabase

	slog.Info("Çalıştırma için Aktör Sistemi kullanılıyor")
	log.Println()

	// İptal edilebilir context oluştur
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SIGINT ve SIGTERM sinyallerini yakalamak için kanal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// WebSocket yöneticisini başlat
	wsManager := api.NewWebSocketManager()
	wsLogger := api.NewWebSocketLogger(wsManager)

	// Hata ayıklama yöneticisini başlat
	debugManager := engine.NewDebugManager()

	// Çalıştırma motorunu başlat
	executionEngine := engine.NewExecutionEngine(wsLogger, debugManager)
	executionEngine.SetExecutionMode(engine.ModeActor)

	// Error handling system setup
	errorAwareEngine := setupErrorHandlingSystem(executionEngine, wsManager)

	// Global düğüm kaydını al
	registry.Make(nodes.Core)
	globalRegistry := registry.GetInstance()

	// API sunucusunu başlat
	var apiRouter *mux.Router

	// Veritabanı bağlantısını başlat
	log.Println("Veritabanı bağlantısı başlatılıyor...")
	connectionManager, repoFactory, err := db.Setup(ctx)
	if err != nil {
		log.Fatalf("Veritabanı başlatılamadı: %v", err)
	}
	defer connectionManager.Close()

	// Veritabanı entegrasyonlu API sunucusunu başlat
	log.Println("Veritabanı depolaması kullanılıyor")
	apiServer := api.NewAPIServerWithDB(executionEngine, wsManager, debugManager, repoFactory, errorAwareEngine)

	// Çalıştırma olay dinleyicisini kaydet
	executionEventListener := api.NewExecutionEventListener(wsManager)
	executionEngine.AddExecutionListener(executionEventListener)

	// Düğüm tiplerini kaydet
	registerNodeTypes(apiServer, globalRegistry)

	// Register error-aware nodes
	registerErrorAwareNodes(globalRegistry)

	// Veritabanından blueprintleri yükle
	if err := loadBlueprintsFromDB(ctx, repoFactory, executionEngine); err != nil {
		log.Printf("Uyarı: Veritabanından blueprintler yüklenemedi: %v", err)
	}

	apiRouter = apiServer.SetupRoutes()

	// Sunucu portunu bayraktan, çevre değişkeninden veya varsayılandan al
	serverPort := *port
	if serverPort == "" {
		serverPort = os.Getenv("PORT")
		if serverPort == "" {
			serverPort = "8089" // Varsayılan port
		}
	}

	// Sunucuyu başlat
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", serverPort),
		Handler: apiRouter,
	}

	// Sunucuyu bir goroutine içinde başlat
	go func() {
		log.Printf("WebBlueprint sunucusu http://localhost%s adresinde başlatılıyor", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Sunucu hatası: %v", err)
		}
	}()

	// Sonlandırma sinyali için bekle
	<-signals
	log.Println("Sunucu kapatılıyor...")

	// Düzgün kapanma için süre sınırı oluştur
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Sunucuyu kapat
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Sunucu kapatma başarısız: %v", err)
	}

	log.Println("Sunucu düzgünce kapatıldı")
}

// setupErrorHandlingSystem initializes the error handling system
func setupErrorHandlingSystem(baseEngine *engine.ExecutionEngine, wsManager *api.WebSocketManager) *engine.ErrorAwareExecutionEngine {
	// Create the error-aware execution engine wrapper
	errorAwareEngine := engine.NewErrorAwareExecutionEngine(baseEngine)

	// Get error manager from the engine
	errorManager := errorAwareEngine.GetErrorManager()
	recoveryManager := errorAwareEngine.GetRecoveryManager()

	// Create WebSocket handler for error notifications
	errorWsHandler := api.NewErrorWebSocketHandler(wsManager)
	errorWsHandler.RegisterWithErrorManager(errorManager)

	// Register default recovery handlers
	registerDefaultRecoveryHandlers(recoveryManager)

	// Register error handlers for logging
	errorManager.RegisterErrorHandler(bperrors.ErrorTypeExecution, func(err *bperrors.BlueprintError) error {
		log.Printf("[ERROR] Execution error: %s (Node: %s, Code: %s)",
			err.Message, err.NodeID, err.Code)
		return nil
	})

	errorManager.RegisterErrorHandler(bperrors.ErrorTypeDatabase, func(err *bperrors.BlueprintError) error {
		log.Printf("[ERROR] Database error: %s (Code: %s)",
			err.Message, err.Code)
		return nil
	})

	errorManager.RegisterErrorHandler(bperrors.ErrorTypeNetwork, func(err *bperrors.BlueprintError) error {
		log.Printf("[ERROR] Network error: %s (Node: %s, Code: %s)",
			err.Message, err.NodeID, err.Code)
		return nil
	})

	return errorAwareEngine
}

// Düğüm tiplerini kaydet
func registerNodeTypes(apiServer *api.APIServerWithDB, globalRegistry *registry.GlobalNodeRegistry) {
	// Mantık düğümleri
	registerNodeType(apiServer, globalRegistry, "if-condition", logic.NewIfConditionNode)
	registerNodeType(apiServer, globalRegistry, "loop", logic.NewLoopNode)
	registerNodeType(apiServer, globalRegistry, "sequence", logic.NewSequenceNode)
	registerNodeType(apiServer, globalRegistry, "branch", logic.NewBranchNode)

	// Web düğümleri
	registerNodeType(apiServer, globalRegistry, "http-request", web.NewHTTPRequestNode)
	registerNodeType(apiServer, globalRegistry, "dom-element", web.NewDOMElementNode)
	registerNodeType(apiServer, globalRegistry, "dom-event", web.NewDOMEventNode)
	registerNodeType(apiServer, globalRegistry, "storage", web.NewStorageNode)

	// Veri düğümleri
	registerNodeType(apiServer, globalRegistry, "constant-string", data.NewStringConstantNode)
	registerNodeType(apiServer, globalRegistry, "constant-number", data.NewNumberConstantNode)
	registerNodeType(apiServer, globalRegistry, "constant-boolean", data.NewBooleanConstantNode)
	registerNodeType(apiServer, globalRegistry, "variable-get", data.NewVariableGetNode)
	registerNodeType(apiServer, globalRegistry, "variable-set", data.NewVariableSetNode)
	registerNodeType(apiServer, globalRegistry, "json-processor", data.NewJSONNode)
	registerNodeType(apiServer, globalRegistry, "array-operations", data.NewArrayNode)
	registerNodeType(apiServer, globalRegistry, "object-operations", data.NewObjectNode)
	registerNodeType(apiServer, globalRegistry, "type-conversion", data.NewTypeConversionNode)

	// Matematik düğümleri
	registerNodeType(apiServer, globalRegistry, "math-add", math.NewAddNode)
	registerNodeType(apiServer, globalRegistry, "math-subtract", math.NewSubtractNode)
	registerNodeType(apiServer, globalRegistry, "math-multiply", math.NewMultiplyNode)
	registerNodeType(apiServer, globalRegistry, "math-divide", math.NewDivideNode)

	// Yardımcı düğümler
	registerNodeType(apiServer, globalRegistry, "print", utility.NewPrintNode)
	registerNodeType(apiServer, globalRegistry, "timer", utility.NewTimerNode)
}

// API sunucusu için düğüm tipini kaydet
func registerNodeType(apiServer *api.APIServerWithDB, globalRegistry *registry.GlobalNodeRegistry, typeID string, factory func() node.Node) {
	apiServer.RegisterNodeType(typeID, factory)
}

// Veritabanından blueprintleri yükle
func loadBlueprintsFromDB(ctx context.Context, repoFactory repository.RepositoryFactory, executionEngine *engine.ExecutionEngine) error {
	blueprintRepo := repoFactory.GetBlueprintRepository()

	// Veritabanından tüm blueprintleri al
	// TODO: İş yükü çok büyükse, sayfalandırma ile al

	// Sadece gösterim amaçlı basit uygulama - gerçek uygulamada bu daha kapsamlı olmalı
	query := `
		SELECT b.id
		FROM blueprints b
		JOIN assets a ON b.id = a.id
		ORDER BY a.created_at
		LIMIT 100
	`

	rows, err := db.GetConnectionManager().GetDB().QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var blueprintIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		blueprintIDs = append(blueprintIDs, id)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Her blueprint'i yükle
	for _, id := range blueprintIDs {
		// Veritabanından model al
		blueprintModel, err := blueprintRepo.GetByID(ctx, id)
		if err != nil {
			log.Printf("Blueprint %s alınamadı: %v", id, err)
			continue
		}

		// Paket blueprint'ine dönüştür
		bp, err := blueprintRepo.ToPkgBlueprint(
			blueprintModel,
			blueprintModel.CurrentVersion,
		)
		if err != nil {
			log.Printf("Blueprint %s dönüştürülemedi: %v", id, err)
			continue
		}

		// Çalıştırma motoruna yükle
		if err := executionEngine.LoadBlueprint(bp); err != nil {
			log.Printf("Blueprint %s çalıştırma motoruna yüklenemedi: %v", id, err)
			continue
		}

		log.Printf("Blueprint yüklendi: %s (%s)", bp.Name, id)
	}

	return nil
}
