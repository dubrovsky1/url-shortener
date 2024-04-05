package postgresql

import (
	"bytes"
	"context"
	"errors"
	errs "github.com/dubrovsky1/url-shortener/internal/errors"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Здесь как обычно обращаемся к базе, только сам вызов метода и наполнение deletedItems будет контролироваться сервисом
func (s *Storage) DeleteURL(ctx context.Context, deletedItems []models.DeletedURLS) error {
	if len(deletedItems) > 0 {
		query := bytes.NewBufferString("with del(created_user_id, shorten_url) as (values ")

		for i, item := range deletedItems {
			query.WriteString("('")
			query.WriteString(item.UserID.String())
			query.WriteString("'::uuid, '")
			query.WriteString(string(item.ShortURL))
			query.WriteString("'::text)")

			if i < len(deletedItems)-1 {
				query.WriteString(",")
			}
		}

		query.WriteString(") update shorten_urls su set is_deleted = true from del where su.created_user_id = del.created_user_id and su.shorten_url = del.shorten_url;")

		//logger.Sugar.Infow("Delete log.", "query", query.String())

		_, err := s.DB.ExecContext(ctx, query.String())

		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgerrcode.IsNoData(pgErr.Code) {
				return errs.ErrShortURLNotFound
			}
			return err
		}
	}
	return nil
}
