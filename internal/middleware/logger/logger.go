package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Sugar zap.SugaredLogger

// структура для логировании информации ответов на запросы
type responseData struct {
	status int
	size   int
}

// создаем тип, для которого переопределяем методы интерфейса http.ResponseWriter с добавлением логирования ответов
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func Initialize() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer logger.Sync()

	Sugar = *logger.Sugar()

	return nil
}

// функция обертка для обработки хендлеров с добавлением функциональности логирования
func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		resData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   resData,
		}

		//выполнение входящего запроса, в качестве объекта ответа передаем реализацию http.ResponseWriter с логированием
		h(&lw, r)

		duration := time.Since(start)

		// отправляем сведения о запросе в zap
		Sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", resData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", resData.size, // получаем перехваченный размер ответа
		)
	}
	return logFn
}
