package audit

type AuditProcessor interface {
	Write(AuditEvent) error
}
