Дополните API сервера, чтобы позволить ему принимать метрики в формате JSON.
При реализации задействуйте одну из распространённых библиотек:
encoding/json,
github.com/mailru/easyjson,
github.com/pquerna/ffjson.
github.com/labstack/echo,
github.com/goccy/go-json.
Двухсторонний обмен данными между агентом и сервером организуйте с использованием следующей структуры:
  type Metrics struct {
     ID    string   `json:"id"`              // имя метрики
     MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
     Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
     Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
  } 
Добавьте в код сервера новый эндпоинт POST /update, принимающий в теле запроса и сохраняющий новые значения метрик. Сервер должен обрабатывать данные в формате JSON, указанном выше. Пример запроса к серверу с новыми значениями метрик:

POST http://localhost:8080/update HTTP/1.1
Host: localhost:8080
Content-Type: application/json
{
  "id": "LastGC",
  "type": "gauge",
  "value": 1744184459
} 
Добавьте в код сервера новый эндпоинт POST /value, возвращающий актуальные значения собранных метрик. В теле запроса отправьте описанный выше JSON с заполненными полями ID и MType. В теле ответа от сервера должен приходить такой же JSON, но с уже заполненными значениями метрик.
Пример запроса к серверу:

POST http://localhost:8080/value HTTP/1.1
Host: localhost:8080
Content-Type: application/json
{
  "id": "LastGC",
  "type": "gauge"
} 
Пример ответа сервера:

POST http://localhost:8080/update HTTP/1.1
Host: localhost:8080
Content-Type: application/json
{
  "id": "LastGC",
  "type": "gauge",
  "value": 1744184459
} 
Переведите агент на отправку метрик через новый эндпоинт POST /update. Удостоверьтесь, что в запросах и ответах присутствует HTTP-заголовок Content-Type со значением application/json. Он указывает агенту и серверу, в каком формате передано тело ответа.
Автотесты проверяют, что агент экспортирует и обновляет на сервере метрики, описанные в первых инкрементах.