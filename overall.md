# WebBlueprint Teknik Dokümantasyonu

## İçindekiler

1. [Giriş ve Proje Özeti](#1-giriş-ve-proje-özeti)
2. [Mimari Yapı](#2-mimari-yapı)
3. [Ana Bileşenler](#3-ana-bileşenler)
4. [Execution Flow](#4-execution-flow)
5. [Data Flow](#5-data-flow)
6. [Node Sistemi](#6-node-sistemi)
7. [Kullanıcı Tanımlı Bileşenler](#7-kullanıcı-tanımlı-bileşenler)
8. [Debug ve Monitoring](#8-debug-ve-monitoring)
9. [Engine İmplementasyonları](#9-engine-implementasyonları)
10. [API ve WebSocket İletişimi](#10-api-ve-websocket-iletişimi)
11. [Engine Performans Karşılaştırması](#11-engine-performans-karşılaştırması)
12. [Geliştirmeye Açık Alanlar](#12-geliştirmeye-açık-alanlar)

## 1. Giriş ve Proje Özeti

WebBlueprint, Unreal Engine Blueprint sisteminden ilham alan, web tabanlı görsel programlama ortamı sunan bir platformdur. Bu platform, programlama bilgisi olmayan kullanıcıların bile karmaşık web uygulamaları ve otomasyonlar oluşturmasına olanak sağlar.

Projenin temel amaçları:

- Görsel, node-temelli bir programlama deneyimi sunmak
- Arka planda güçlü ve esnek bir execution engine ile çalışmak
- Farklı node tipleri ile web, veri işleme ve görselleştirme işlemlerini desteklemek
- Kullanıcılara özel fonksiyon ve değişken tanımlama imkanı vermek
- Gerçek zamanlı debug ve monitoring yetenekleri sağlamak

WebBlueprint, Go dilinde yazılmış güçlü bir backend ile Vue.js tabanlı bir frontend bileşeninden oluşmaktadır. Bu dokümantasyon, backend tarafının teknik detaylarını ele almaktadır.

## 2. Mimari Yapı

WebBlueprint backend mimarisi, temiz ve modüler bir yapıya sahiptir. Temel olarak aşağıdaki katmanlardan oluşur:

### Paket Organizasyonu

```
webblueprint/
├── cmd/
│   └── server/           # Ana application entry point
├── internal/
│   ├── api/              # HTTP API ve WebSocket iletişimi
│   ├── db/               # Blueprint saklama (şu an in-memory)
│   ├── engine/           # Execution engine'leri
│   │   └── ...           # Standart ve Actor engine impl.
│   ├── node/             # Node interface ve base impl.
│   ├── nodes/            # Node tipleri implementasyonları
│   │   ├── data/         # Data işleme node'ları
│   │   ├── logic/        # Logic flow node'ları
│   │   ├── math/         # Matematik node'ları
│   │   ├── utility/      # Yardımcı node'lar
│   │   └── web/          # Web işlemleri node'ları
│   ├── registry/         # Global node registry
│   └── types/            # Ortak tip tanımlamaları
└── pkg/
    └── blueprint/        # Blueprint yapıları
```

### Katmanlı Mimari

WebBlueprint, aşağıdaki katmanları içeren modüler bir mimariye sahiptir:

1. **Core Layer**: Blueprint çalıştırma mekanizmaları, node yönetimi, tip sistemi
2. **Engine Layer**: Execution akışını yöneten standart ve actor-based engine'ler
3. **Node Layer**: Tüm node tipleri ve bunların implement edilmiş davranışları
4. **API Layer**: HTTP API, WebSocket, client iletişimi
5. **Storage Layer**: Blueprint saklama ve yönetimi (şu an in-memory)

### İletişim Modeli

Sistem bileşenleri arasında iletişim şu şekilde gerçekleşir:

- **API → Engine**: HTTP API istekleri engine'i tetikler
- **Engine → Nodes**: Engine node'ları yürütür ve aralarındaki veri akışını yönetir
- **WebSocket ↔ Client**: Gerçek zamanlı execution durumu ve debug verileri
- **Registry → API/Engine**: Node tiplerinin merkezi kayıt noktası

## 3. Ana Bileşenler

### ExecutionEngine

Execution engine, blueprint'lerin çalıştırılmasından sorumlu ana bileşendir. İki farklı implementasyonu bulunmaktadır:

1. **Standard Engine**: Senkron, sıralı çalışan temel engine
2. **Actor Engine**: Actor modeli kullanarak concurrent çalışan gelişmiş engine

Engine, blueprint içindeki node'ları çalıştırır, aralarındaki veri akışını yönetir ve execution durumunu takip eder.

```go
// ExecutionEngine manages blueprint execution
type ExecutionEngine struct {
    nodeRegistry    map[string]node.NodeFactory
    blueprints      map[string]*blueprint.Blueprint
    executionStatus map[string]*ExecutionStatus
    variables       map[string]map[string]types.Value
    listeners       []ExecutionListener
    debugManager    *DebugManager
    logger          node.Logger
    executionMode   ExecutionMode
    mutex           sync.RWMutex
}
```

### Node Interface

Tüm node tipleri, `node.Node` interface'ini implement eder:

```go
// Node is the interface that all node types must implement
type Node interface {
    // GetMetadata returns metadata about the node type
    GetMetadata() NodeMetadata

    // GetInputPins returns the input pins for this node
    GetInputPins() []types.Pin

    // GetOutputPins returns the output pins for this node
    GetOutputPins() []types.Pin

    // Execute runs the node's logic with the given execution context
    Execute(ctx ExecutionContext) error
}
```

### ExecutionContext

ExecutionContext, node'ların execution sırasında iletişim kurduğu ana arayüzdür:

```go
// ExecutionContext provides services to nodes during execution
type ExecutionContext interface {
    // Input/output access
    GetInputValue(pinID string) (types.Value, bool)
    SetOutputValue(pinID string, value types.Value)

    // Execution control
    ActivateOutputFlow(pinID string) error

    // State management
    GetVariable(name string) (types.Value, bool)
    SetVariable(name string, value types.Value)

    // Logging and debugging
    Logger() Logger
    RecordDebugInfo(info types.DebugInfo)
    GetDebugData() map[string]interface{}

    // Node information
    GetNodeID() string
    GetNodeType() string

    // Blueprint information
    GetBlueprintID() string
    GetExecutionID() string
}
```

### Blueprint

Blueprint, çalıştırılabilir bir görsel programı temsil eder:

```go
// Blueprint represents a complete blueprint definition
type Blueprint struct {
    ID          string            // Unique ID
    Name        string            // Human-readable name
    Description string            // Optional description
    Version     string            // Version information
    Nodes       []BlueprintNode   // Nodes in this blueprint
    Functions   []Function        // User-defined functions
    Connections []Connection      // Connections between nodes
    Variables   []Variable        // Blueprint variables
    Metadata    map[string]string // Additional metadata
}
```

### Registry

Node registry, tüm node tiplerinin merkezi kayıt noktasıdır:

```go
// GlobalNodeRegistry provides a central point to access node factories
type GlobalNodeRegistry struct {
    factories map[string]node.NodeFactory
    mutex     sync.RWMutex
}
```

## 4. Execution Flow

Execution flow, WebBlueprint'in çalışma şeklinin temelidir. Blueprint üzerindeki node'lar arasındaki execution (yürütme) akışını tanımlar.

### Execution Flow Nasıl Çalışır

1. **Entry Point Belirleme**: Blueprint çalıştırıldığında, ilk olarak entry point node'ları belirlenir (giriş bağlantısı olmayan node'lar)
2. **Execution Flow Pinleri**: Node'lar arasındaki execution bağlantıları, "exec → then" pinleri üzerinden yapılır
3. **Sequential Execution**: Engine, bir node'u çalıştırdıktan sonra, o node'un aktifleştirdiği output execution pin'ine bağlı olan bir sonraki node'a geçer
4. **Branching**: If-Condition gibi node'lar, koşula bağlı olarak farklı execution path'lerini aktifleştirebilir
5. **Loops**: Loop node'ları, belirli bir execution path'ini birden fazla kez çalıştırabilir

### Execution Pinleri

Execution pin'leri, node'lar arasındaki çalışma sırasını belirler:

- **Execution Input Pin (exec)**: Node'un çalışmasını tetikler
- **Execution Output Pin (then)**: Node çalışmasını tamamladığında, sonraki node'u tetikler

```go
// Execution pin type
PinTypes.Execution: &PinType{
    ID:          "execution",
    Name:        "Execution",
    Description: "Controls execution flow",
    Validator:   func(v interface{}) error { return nil },
}
```

### ActivateOutputFlow Mekanizması

Node'lar, çalışmalarını tamamladıklarında hangi execution path'inin aktifleştirileceğini belirlemek için `ActivateOutputFlow` metodunu kullanırlar:

```go
// Node execution tamamlandığında
return ctx.ActivateOutputFlow("then")

// Koşullu branching için
if condition {
    return ctx.ActivateOutputFlow("true")
} else {
    return ctx.ActivateOutputFlow("false")
}
```

### Asynchronous Flow Handling

Asenkron işlemler gerektiren node'lar (HTTP request, timer gibi) için özel mekanizmalar:

1. **Actor System**: Actor-based engine, her node için bir actor oluşturarak asenkron çalışmayı destekler
2. **Callbacks**: Asenkron işlemler tamamlandığında callback'ler üzerinden execution flow devam eder

### Veri Odaklı Node'lar (Execution Flow Olmadan)

Değişken node'ları gibi bazı node tipleri, execution flow'a katılmadan sadece veri akışını sağlarlar:

1. **Pre-Processing**: Engine, execution başlamadan önce tüm değişken node'larını işler
2. **Pasif Node'lar**: Execution flow'a katılmayan node'lar diğer node'lar tarafından dolaylı olarak kullanılır

## 5. Data Flow

Data flow, node'lar arasındaki veri akışını tanımlar. Execution flow'dan bağımsız olarak, node'ların input ve output pinleri arasında veri taşınmasını sağlar.

### Data Flow Nasıl Çalışır

1. **Pin Connections**: Node'lar arasında pin bağlantıları veri akışını belirler (output → input)
2. **Type Checking**: Bağlantı kurulurken ve çalışma zamanında tip kontrolü yapılır
3. **Value Transfer**: Bir node çalıştığında, output pin'lerindeki değerler, bağlı oldukları input pin'lerine aktarılır
4. **Lazy Loading**: Bazı durumlarda, bir input pin'i değeri talep ettiğinde bu değer hesaplanır

### Pin Tipleri

WebBlueprint, çeşitli veri tipleri için pin desteği sağlar:

```go
var PinTypes = struct {
    Execution *PinType
    String    *PinType
    Number    *PinType
    Boolean   *PinType
    Object    *PinType
    Array     *PinType
    Any       *PinType
}
```

Her pin tipi için tip doğrulama ve tip dönüşüm fonksiyonları tanımlanmıştır.

### Type Conversion

Farklı pin tipleri arasında otomatik tip dönüşümleri:

```go
// String to number
func convertToNumber(value interface{}) (interface{}, error) {
    // String to number conversion logic
}

// Number to boolean
func convertToBoolean(value interface{}) (interface{}, error) {
    // Number to boolean conversion logic
}
```

### Değişken Değerleri

Değişkenler, data flow'un özel bir durumudur:

1. **Variable Setters**: Değişken değerlerini ayarlar (execution flow kullanmadan)
2. **Variable Getters**: Değişken değerlerini okur (execution flow kullanmadan)
3. **Context Storage**: Değişken değerleri `ExecutionContext` üzerinde saklanır

Değişken node'ları, execution flow'a katılmadan veri aktarımını sağlar:

```go
// Set variable
ctx.SetVariable(varName, value)

// Get variable
varValue, exists := ctx.GetVariable(varName)
```

### Indirect Data Access

Node'lar, aşağıdaki yollarla dolaylı olarak verilere erişebilir:

1. **Variable Access**: Değişken değerlerini okuyarak
2. **Debug Manager**: Diğer node'ların output değerlerini debug manager'dan alarak
3. **Context Sharing**: ExecutionContext üzerinden paylaşılan veriler aracılığıyla

## 6. Node Sistemi

WebBlueprint'in node sistemi, görsel programlama ortamının temelini oluşturur. Her node, belirli bir işlevi yerine getiren, giriş ve çıkış pin'lerine sahip bir bileşendir.

### Node Kategorileri

Şu anda desteklenen node kategorileri:

1. **Logic Nodes**: If-Condition, Loop, Branch, Sequence
2. **Math Nodes**: Add, Subtract, Multiply, Divide
3. **Data Nodes**: Constant, Variable, JSON, Array, Object
4. **Web Nodes**: HTTP Request, DOM Element, DOM Event, Storage
5. **Utility Nodes**: Print, Timer

### Node Yapısı

Tüm node'lar `Node` interface'ini implement eder. Tipik bir node yapısı:

```go
type MyNode struct {
    node.BaseNode // Temel implementasyonu sağlar
}

func NewMyNode() node.Node {
    return &MyNode{
        BaseNode: node.BaseNode{
            Metadata: node.NodeMetadata{...},
            Inputs: []types.Pin{...},
            Outputs: []types.Pin{...},
        },
    }
}

func (n *MyNode) Execute(ctx node.ExecutionContext) error {
    // Node mantığı burada implement edilir
    return ctx.ActivateOutputFlow("then")
}
```

### Node Registrasyonu

Node tipleri, global registry'e kaydedilmelidir:

```go
globalRegistry := registry.GetInstance()
globalRegistry.RegisterNodeType("my-node", NewMyNode)

// API ve engine'e de kaydedilmeli
apiServer.RegisterNodeType("my-node", NewMyNode)
executionEngine.RegisterNodeType("my-node", NewMyNode)
```

### Pin Sistemi

Her node'un input ve output pin'leri vardır:

```go
type Pin struct {
    ID          string      // Unique identifier
    Name        string      // Human-readable name
    Description string      // Description
    Type        *PinType    // Type information
    Optional    bool        // Whether this pin is required
    Default     interface{} // Default value if not connected
}
```

Pin'ler iki kategoriye ayrılır:

1. **Execution Pins**: Node çalışma sırasını kontrol eder (exec → then)
2. **Data Pins**: Node'lar arasında veri akışını sağlar

### Custom Node Implementasyonu Örneği

```go
// Add node implementation
func (n *AddNode) Execute(ctx node.ExecutionContext) error {
    // Get input values
    aValue, aExists := ctx.GetInputValue("a")
    bValue, bExists := ctx.GetInputValue("b")

    // Validate inputs
    if !aExists || !bExists {
        return fmt.Errorf("missing required inputs")
    }

    // Convert to numbers
    a, err := aValue.AsNumber()
    if err != nil {
        return err
    }

    b, err := bValue.AsNumber()
    if err != nil {
        return err
    }

    // Perform operation
    result := a + b

    // Set output value
    ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, result))

    // Continue execution
    return ctx.ActivateOutputFlow("then")
}
```

## 7. Kullanıcı Tanımlı Bileşenler

WebBlueprint, kullanıcıların kendi fonksiyonlarını ve değişkenlerini tanımlamasına olanak sağlar.

### User-Defined Functions

Kullanıcı tanımlı fonksiyonlar, yeniden kullanılabilir blueprint parçalarıdır:

```go
// Function represents a user-defined function in a blueprint
type Function struct {
    ID          string            // Unique identifier
    Name        string            // Function name
    Description string            // Optional description
    NodeType    BlueprintNodeType // Interface definition
    Nodes       []BlueprintNode   // Internal nodes
    Connections []Connection      // Internal connections
    Variables   []Variable        // Function-local variables
    Metadata    map[string]string // Additional metadata
}
```

Fonksiyon tanımlama mekanizması:

1. **Function Definition**: Kullanıcı bir blueprint içinde fonksiyon tanımlar
2. **Function Registration**: Fonksiyon, node registry'e kaydedilir
3. **Function Usage**: Fonksiyon, diğer blueprint'lerde node olarak kullanılabilir

Fonksiyon çalıştırma mekanizması:

1. **Function Context**: Fonksiyon için özel bir execution context oluşturulur
2. **Recursion Prevention**: Özyinelemeli çağrıları önlemek için korumalar bulunur
3. **Mini Execution Engine**: Fonksiyon içindeki node'ları çalıştırmak için mini bir engine kullanılır

### User-Defined Variables

Kullanıcı tanımlı değişkenler, blueprint içinde veri saklama mekanizmasıdır:

```go
// Variable represents a blueprint variable
type Variable struct {
    Name  string      // Variable name
    Type  string      // Variable type
    Value interface{} // Initial value
}
```

Değişken tanımlama ve kullanma mekanizması:

1. **Variable Definition**: Kullanıcı blueprint'te değişken tanımlar
2. **Variable Node Creation**: Her değişken için getter/setter node'ları otomatik oluşturulur
3. **Variable Storage**: Değişken değerleri execution context üzerinde saklanır

Değişken node'ları özellikleri:

1. **Non-Executable Nodes**: Execution flow'a doğrudan katılmazlar
2. **Pre-Processed**: Engine, execution başlamadan önce tüm değişken node'larını işler
3. **Direct Data Flow**: Diğer node'lar ile doğrudan veri bağlantısı kurarlar

## 8. Debug ve Monitoring

WebBlueprint, gelişmiş debug ve monitoring özellikleri sunar.

### Debug Manager

DebugManager, execution sırasında debug verilerini toplar ve saklama:

```go
// DebugManager stores and manages debug information during execution
type DebugManager struct {
    // Maps: executionID -> nodeID -> data
    debugData map[string]map[string]map[string]interface{}

    // Maps: executionID -> nodeID -> pinID -> value
    outputValues map[string]map[string]map[string]interface{}

    mutex sync.RWMutex
}
```

Debug verileri şunları içerir:

1. **Node States**: Node'ların çalışma durumları
2. **Pin Values**: Input ve output pin'lerinin değerleri
3. **Execution Paths**: Execution akışı bilgileri
4. **Timestamps**: Her event'in zaman damgası
5. **Error Information**: Hata mesajları ve detayları

### Event Listener Sistemi

Execution olaylarını dinlemek için listener mekanizması:

```go
// ExecutionListener listens for execution events
type ExecutionListener interface {
    OnExecutionEvent(event ExecutionEvent)
}

// ExecutionEvent represents an event during blueprint execution
type ExecutionEvent struct {
    Type      ExecutionEventType
    Timestamp time.Time
    NodeID    string
    Data      map[string]interface{}
}
```

Event tipleri:

1. **NodeStarted**: Node çalışmaya başladığında
2. **NodeCompleted**: Node başarıyla tamamlandığında
3. **NodeError**: Node çalışırken hata oluştuğunda
4. **ValueProduced**: Output değeri üretildiğinde
5. **ExecutionStart/End**: Blueprint çalışması başladığında/bittiğinde

### WebSocket Monitoring

WebSocket üzerinden gerçek zamanlı monitoring:

1. **Connection Management**: Client bağlantılarının yönetimi
2. **Event Broadcasting**: Execution event'lerinin client'lara iletilmesi
3. **Debug Data Streaming**: Debug verilerinin gerçek zamanlı iletimi

```go
// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
    Type    string          // Message type
    Payload json.RawMessage // Message payload
}

// Message types
const (
    MsgTypeNodeStart    = "node.start"
    MsgTypeNodeComplete = "node.complete"
    MsgTypeNodeError    = "node.error"
    MsgTypeDataFlow     = "data.flow"
    MsgTypeDebugData    = "debug.data"
    MsgTypeExecStart    = "execution.start"
    MsgTypeExecEnd      = "execution.end"
)
```

## 9. Engine İmplementasyonları

WebBlueprint, iki farklı execution engine implementasyonu sunar.

### Standard Engine

Standart engine, senkron ve sıralı çalışan basit bir implementasyondur:

1. **Sequential Execution**: Node'lar sırayla çalıştırılır
2. **Direct Function Calls**: Node'lar direkt fonksiyon çağrıları ile tetiklenir
3. **Shared Context**: Tüm node'lar aynı execution context'i paylaşır
4. **Simple Debugging**: Debug verileri kolayca toplanır

Avantajları:
- Daha basit implementasyon
- Deterministik davranış
- Daha kolay debug

Dezavantajları:
- Daha düşük performans
- Asenkron işlemleri yönetmek zorlaşabilir
- Ölçeklenebilirlik sınırlı

### Actor Engine

Actor engine, actor modeli kullanarak concurrent çalışan gelişmiş bir implementasyondur:

1. **Concurrent Execution**: Node'lar paralel olarak çalışabilir
2. **Message Passing**: Node'lar mesaj geçişi ile iletişim kurar
3. **Isolated State**: Her node'un izole durumu vardır
4. **Mailbox System**: Her node bir mesaj kuyruğuna sahiptir

```go
// NodeActor represents a single node in the actor model system
type NodeActor struct {
    NodeID      string
    NodeType    string
    BlueprintID string
    ExecutionID string
    node        node.Node
    mailbox     chan NodeMessage
    inputs      map[string]types.Value
    outputs     map[string]types.Value
    variables   map[string]types.Value
    status      NodeStatus
    ctx         *ActorExecutionContext
    logger      node.Logger
    listeners   []ExecutionListener
    debugMgr    *DebugManager
    done        chan struct{}
    mutex       sync.RWMutex
}
```

Avantajları:
- Daha yüksek performans
- Daha iyi ölçeklenebilirlik
- Asenkron işlemleri daha iyi yönetim

Dezavantajları:
- Daha karmaşık implementasyon
- Debugging daha zor
- Race condition ve deadlock potansiyeli

### Engine Seçimi

Uygulamanın ihtiyaçlarına göre engine seçimi yapılabilir:

```go
// Set execution mode if actor system is enabled
if *useActorSystem {
    log.Println("Using Actor System for execution")
    executionEngine.SetExecutionMode(engine.ModeActor)
} else {
    log.Println("Using Standard Engine for execution")
    executionEngine.SetExecutionMode(engine.ModeStandard)
}
```

## 10. API ve WebSocket İletişimi

WebBlueprint, client ile iletişim için HTTP API ve WebSocket protokollerini kullanır.

### HTTP API

HTTP API, blueprint yönetimi ve execution için arayüz sağlar:

```
GET    /api/nodes                # Get all node types
GET    /api/blueprints           # Get all blueprints
POST   /api/blueprints           # Create a blueprint
GET    /api/blueprints/{id}      # Get a specific blueprint
PUT    /api/blueprints/{id}      # Update a blueprint
DELETE /api/blueprints/{id}      # Delete a blueprint
POST   /api/blueprints/{id}/execute # Execute a blueprint
POST   /api/blueprints/{id}/variable # Create a variable
GET    /api/executions/{id}      # Get execution status
GET    /api/executions/{id}/nodes/{nodeId} # Get node debug data
```

### WebSocket İletişimi

WebSocket, real-time monitoring ve debug için kullanılır:

1. **Connection Handling**: `/ws` endpoint'i üzerinden bağlantı kurulur
2. **Event Streaming**: Execution olayları gerçek zamanlı olarak iletilir
3. **Bidirectional Communication**: Client komutları (pause, resume) gönderilebilir

```go
// WebSocketManager handles WebSocket connections and messaging
type WebSocketManager struct {
    clients    map[string]*WebSocketClient
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
    broadcast  chan []byte
    mutex      sync.RWMutex
}
```

### Node Type Introduction

API, client'a node tipleri hakkında bilgi sağlar:

```go
// Broadcast node type to connected clients
s.wsManager.BroadcastMessage(MsgTypeNodeIntro, map[string]interface{}{
    "typeId":      metadata.TypeID,
    "name":        metadata.Name,
    "description": metadata.Description,
    "category":    metadata.Category,
    "version":     metadata.Version,
    "inputs":      convertPinsToInfo(nodeInstance.GetInputPins()),
    "outputs":     convertPinsToInfo(nodeInstance.GetOutputPins()),
})
```

## 11. Engine Performans Karşılaştırması

İki farklı execution engine'in performans özellikleri:

### Standard Engine Performansı

1. **Execution Time**: Küçük blueprint'ler için iyi, büyük ve karmaşık blueprint'ler için daha yavaş
2. **Memory Usage**: Daha düşük bellek kullanımı
3. **Concurrency**: Sınırlı, çoğunlukla senkron çalışma
4. **Scalability**: Daha düşük, node sayısı arttıkça performans düşer
5. **Best For**: Küçük/orta ölçekli, senkron blueprint'ler

### Actor Engine Performansı

1. **Execution Time**: Büyük ve karmaşık blueprint'lerde daha hızlı
2. **Memory Usage**: Daha yüksek bellek kullanımı (actor başına overhead)
3. **Concurrency**: Tam destek, asenkron çalışma
4. **Scalability**: Yüksek, node sayısı arttıkça performans avantajı artar
5. **Best For**: Büyük ölçekli, asenkron, paralel blueprint'ler

### Görsel Karşılaştırma

İki engine arasındaki performans farkı, blueprint boyutu ve karmaşıklığına bağlıdır:

- **Küçük Blueprint (10-20 node)**: Standard Engine %5-10 daha hızlı
- **Orta Blueprint (50-100 node)**: Actor Engine %10-20 daha hızlı
- **Büyük Blueprint (200+ node)**: Actor Engine %30-50 daha hızlı
- **Asenkron İşlemler (HTTP, Timer)**: Actor Engine belirgin şekilde daha verimli

## 12. Geliştirmeye Açık Alanlar

WebBlueprint projesinde geliştirmeye açık alanlar ve öneriler:

### Mimari İyileştirmeler

1. **Persistent Storage**: Şu anda in-memory olan storage layer'ın kalıcı veritabanı desteği
2. **Modular Plugin System**: Yeni node tiplerinin dinamik olarak yüklenebilmesi için plugin sistemi
3. **Optimized Data Flow**: Veri akışı optimizasyonu için lazy evaluation ve caching mekanizmaları

### Performans İyileştirmeleri

1. **Memory Management**: Büyük blueprint'ler için bellek kullanımını optimize etme
2. **Parallelization**: Execution path'leri arasında paralel çalışma imkanı
3. **Resource Limits**: CPU, memory ve execution time limitleri

### Fonksiyonellik Geliştirmeleri

1. **Error Handling**: Daha gelişmiş hata yakalama ve iyileştirme mekanizmaları
2. **Input Validation**: Node input'ları için daha kapsamlı doğrulama
3. **Type System**: Daha zengin tip sistemi (union types, generics, custom types)
4. **Advanced Debugging**: Step-by-step debugging, breakpoints, conditional breakpoints

### Kullanıcı Tanımlı Bileşenler

1. **Macros**: Kullanıcı tanımlı makrolar için destek
2. **Custom Node Types**: Kullanıcıların yeni node tipleri tanımlaması
3. **Type Definitions**: Kullanıcı tanımlı tip tanımlamaları
4. **Import/Export**: Blueprint parçalarını import/export etme

### API ve Entegrasyon

1. **GraphQL API**: Daha esnek blueprint ve execution sorguları için GraphQL desteği
2. **SSE Support**: Server-Sent Events desteği (WebSocket alternatifi)
3. **External Triggers**: Harici servislerden triggering mekanizması (webhooks gibi)
4. **Integration Nodes**: Popüler servislerle entegrasyon için hazır node'lar

### Gelişmiş Özellikler

1. **Versioning**: Blueprint versiyonlama ve değişiklik takibi
2. **Collaboration**: Multi-user düzenleme ve işbirliği imkanı
3. **Testing Framework**: Blueprint'lerin test edilmesi için altyapı
4. **Analytics**: Execution performans analitiği ve insight'lar

Bu geliştirmeler, WebBlueprint'i daha güçlü, daha ölçeklenebilir ve daha kullanıcı dostu bir platform haline getirecektir.