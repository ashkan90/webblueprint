// Types for error handling system

export enum ErrorType {
  Execution = "execution",
  Connection = "connection",
  Validation = "validation",
  Permission = "permission",
  Database = "database",
  Network = "network",
  Plugin = "plugin",
  System = "system",
  Unknown = "unknown"
}

export enum ErrorSeverity {
  Critical = "critical",
  High = "high",
  Medium = "medium",
  Low = "low",
  Info = "info"
}

export enum RecoveryStrategy {
  Retry = "retry",
  SkipNode = "skip_node",
  UseDefaultValue = "use_default_value",
  ManualIntervention = "manual",
  None = "none"
}

export enum BlueprintErrorCode {
  // Execution errors
  NodeExecutionFailed = "E001",
  NodeNotFound = "E002",
  NodeTypeNotRegistered = "E003",
  ExecutionTimeout = "E004",
  ExecutionCancelled = "E005",
  NoEntryPoints = "E006",

  // Connection errors
  InvalidConnection = "C001",
  CircularDependency = "C002",
  MissingRequiredInput = "C003",
  TypeMismatch = "C004",
  NodeDisconnected = "C005",

  // Validation errors
  InvalidBlueprintStructure = "V001",
  InvalidNodeConfiguration = "V002",
  MissingProperty = "V003",
  InvalidPropertyValue = "V004",

  // Database errors
  DatabaseConnection = "D001",
  BlueprintNotFound = "D002",
  BlueprintVersionNotFound = "D003",
  DatabaseQuery = "D004",

  // System errors
  InternalServerError = "S001",
  ResourceExhausted = "S002",
  SystemUnavailable = "S003",

  // Other error codes
  Unknown = "U001"
}

export interface BlueprintError {
  type: ErrorType;
  code: BlueprintErrorCode;
  message: string;
  details?: Record<string, any>;
  severity: ErrorSeverity;
  recoverable: boolean;
  recoveryOptions?: RecoveryStrategy[];
  nodeId?: string;
  pinId?: string;
  blueprintId?: string;
  executionId?: string;
  timestamp: string;
  stackTrace?: string[];
  
  // UI state - not from server
  expanded?: boolean;
}

export interface RecoveryAttempt {
  strategy: RecoveryStrategy;
  successful: boolean;
  timestamp: string;
  errorCode: string;
  nodeId: string;
  details?: Record<string, any>;
  executionId: string;
}

export interface ErrorAnalysis {
  totalErrors: number;
  recoverableErrors: number;
  typeBreakdown: Record<string, number>;
  severityBreakdown: Record<string, number>;
  topProblemNodes: { nodeId: string; count: number }[];
  mostCommonCodes: Record<string, number>;
  timestamp: string;
}

export interface ValidationResult {
  valid: boolean;
  errors?: BlueprintError[];
  warnings?: BlueprintError[];
  nodeIssues?: Record<string, string[]>;
}

export interface ExtendedExecutionInfo {
  success: boolean;
  executionId: string;
  startTime: string;
  endTime: string;
  partialSuccess?: boolean;
  error?: BlueprintError;
  nodeResults?: Record<string, Record<string, any>>;
  errorAnalysis?: ErrorAnalysis;
  recoveryAttempts?: RecoveryAttempt[];
  validationResults?: ValidationResult;
  failedNodes?: string[];
  successfulNodes?: string[];
}

export interface ErrorNotification {
  type: "error";
  error: BlueprintError;
  executionId: string;
}

export interface ErrorAnalysisNotification {
  type: "error_analysis";
  analysis: ErrorAnalysis;
  executionId: string;
}

export interface RecoveryNotification {
  type: "recovery_attempt";
  successful: boolean;
  strategy: string;
  nodeId: string;
  errorCode: string;
  details?: Record<string, any>;
  executionId: string;
}

export type ErrorNotificationType = 
  ErrorNotification |
  ErrorAnalysisNotification |
  RecoveryNotification;
