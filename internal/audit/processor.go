package audit

// Интерфейс для отправки событий аудита
type AuditProcessor interface {
	Write(AuditEvent) error
}
