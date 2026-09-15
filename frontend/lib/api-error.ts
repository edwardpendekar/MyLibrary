export class ApiError extends Error {
  code: string;
  status: number;
  fields?: Record<string, string>;

  constructor(message: string, code: string, status: number, fields?: Record<string, string>) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.fields = fields;
  }
}
