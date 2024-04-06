package service

import (
	"context"
	"github.com/dubrovsky1/url-shortener/internal/generator"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/google/uuid"
	"sync"
	"time"
)

// для слоя бизнес логики не нужны реализации GetShortURL и io.Closer, поэтому их не включаем
//
//go:generate mockgen -source=service.go -destination=../service/mocks/service.go -package=mocks
type Storager interface {
	SaveURL(context.Context, models.ShortenURL) (models.ShortURL, error)
	GetURL(context.Context, models.ShortURL) (models.ShortenURL, error)
	InsertBatch(context.Context, []models.BatchRequest, models.Host, uuid.UUID) ([]models.BatchResponse, error)
	ListByUserID(context.Context, models.Host, uuid.UUID) ([]models.ShortenURL, error)
	DeleteURL(context.Context, []models.DeletedURLS) error
}

type Service struct {
	storage         Storager
	wg              *sync.WaitGroup
	urlsToDeleteCh  chan models.DeletedURLS //канал, куда складываем приходящие из запросов пользователей урлы, которые необходимо удалить
	deleteInterval  time.Duration           //интервал, по достижении которого происходит удаление
	deleteBatchSize int                     //либо - размер пачки, после заполнения которой, происходит обращение в базу с удалением
	isRun           bool
}

func New(storage Storager, batchSize int, deleteInterval time.Duration) *Service {
	return &Service{
		storage:         storage,
		wg:              &sync.WaitGroup{},
		urlsToDeleteCh:  make(chan models.DeletedURLS),
		deleteBatchSize: batchSize,
		deleteInterval:  deleteInterval,
		isRun:           false,
	}
}

func (s *Service) SaveURL(ctx context.Context, item models.ShortenURL) (models.ShortURL, error) {
	//гененрируем короткую ссылку
	item.ShortURL = models.ShortURL(generator.GetShortURL())

	shortURL, err := s.storage.SaveURL(ctx, item)
	if err != nil {
		return shortURL, err
	}
	return shortURL, nil
}

func (s *Service) GetURL(ctx context.Context, shortURL models.ShortURL) (models.ShortenURL, error) {
	result, err := s.storage.GetURL(ctx, shortURL)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) InsertBatch(ctx context.Context, batch []models.BatchRequest, host models.Host, userID uuid.UUID) ([]models.BatchResponse, error) {
	result, err := s.storage.InsertBatch(ctx, batch, host, userID)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) ListByUserID(ctx context.Context, host models.Host, userID uuid.UUID) ([]models.ShortenURL, error) {
	result, err := s.storage.ListByUserID(ctx, host, userID)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) DeleteURL(ctx context.Context, deletedItems []models.DeletedURLS) error {
	//logger.Sugar.Infow("DeleteURL log.", "deletedItems", deletedItems)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for _, item := range deletedItems {
			s.urlsToDeleteCh <- item //при каждом запросе пользователя создается горутина, которая записывает в канал то, что хотим удалить
		}
	}()
	return nil
}

func (s *Service) DeleteRun(ctx context.Context) {
	go func() {
		defer logger.Sugar.Infow("Stop urls deletion.")
		buffer := make([]models.DeletedURLS, 0, s.deleteBatchSize)

		//удаление остатков из буфера
		defer func() {
			if len(buffer) > 0 {
				logger.Sugar.Infow("Deleting remaining urls.", "count", len(buffer), "buffer", buffer)
				if err := s.storage.DeleteURL(ctx, buffer); err != nil {
					logger.Sugar.Infow("Delete urls error.", "err", err.Error())
				}
			}
		}()

		ticker := time.NewTicker(s.deleteInterval)
		defer ticker.Stop()

		logger.Sugar.Infow("Start urls deletion.")

		for {
			select {
			//истекло время - идем в базу с удалением и очищаем буфер
			case <-ticker.C:
				if err := s.storage.DeleteURL(ctx, buffer); err != nil {
					logger.Sugar.Infow("Delete urls after timeout error.", "err", err.Error(), "buffer", buffer)
					continue
				}
				buffer = buffer[:0]
			//читаем из канала и заполняем буфер пока он не заполнится либо канал не закроется
			case deletedURL, closed := <-s.urlsToDeleteCh:
				if !closed {
					return
				}
				buffer = append(buffer, deletedURL)
				//logger.Sugar.Infow("Buffer.", "values", buffer)

				if len(buffer) < s.deleteBatchSize {
					continue
				}

				//обращаемся в базу с удалением при наступлении необходимых условий
				if err := s.storage.DeleteURL(ctx, buffer); err != nil {
					logger.Sugar.Infow("Delete urls batch error.", "err", err.Error(), "buffer", buffer)
					continue
				}

				//очищаем буфер
				buffer = buffer[:0]
			}
		}
	}()
}

func (s *Service) Run(ctx context.Context) error {
	if s.isRun {
		return nil
	}
	s.isRun = true
	s.DeleteRun(ctx)
	return nil
}

func (s *Service) Close() error {
	s.wg.Wait()
	close(s.urlsToDeleteCh)
	return nil
}
