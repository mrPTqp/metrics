Добавьте в пакет cmd/server и cmd/agent глобальные переменные:
var buildVersion string,
var buildDate string,
var buildCommit string.
При старте приложения выводите в stdout сообщение в следующем формате:
Build version: <buildVersion> (или "N/A" при отсутствии значения)
Build date: <buildDate> (или "N/A" при отсутствии значения)
Build commit: <buildCommit> (или "N/A" при отсутствии значения) 