export function format(schema: string): Promise<string>;

export function validate(
  schemaString: string,
  configFile: string
): Promise<ValidationResult>;

export function completions(
  schemaString: string,
  position: SimplePosition,
  configFile: string
): Promise<CompletionResult>;

export function getDefinition(
  req: GetDefinitionRequest
): Promise<DefinitionResult>;

export interface DefinitionResult {
  schema?: Position;
  function?: { name: string };
}

export interface SchemaDefinition {
  schema: SchemaDefinition;
}

export interface GetDefinitionRequest {
  position: Position;
  schemaFiles: SchemaFile[];
}

export interface SchemaFile {
  filename: string;
  contents: string;
}

export interface SimplePosition {
  column: number;
  line: number;
}

export interface Position extends SimplePosition {
  filename: string;
  offset: number;
}

export interface CompletionItem {
  description: string;
  label: string;
  insertText: string;
  kind: string;
}

export interface CompletionResult {
  completions: CompletionItem[];
}

export interface ValidationError {
  code: string;
  pos: Position;
  endPos: Position;
  hint: string;
  message: string;
}

export interface ValidationResult {
  errors: ValidationError[];
}
