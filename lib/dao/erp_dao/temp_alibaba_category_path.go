package erp_dao

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/statistics/lib/dbsession"
	"github.com/statistics/lib/model/erp_model"
)

type tempAlibabaCategoryPathDao struct{}

var TempAlibabaCategoryPathDao = new(tempAlibabaCategoryPathDao)

const (
	alibabaSchemaMissing = "(`schema` IS NULL OR TRIM(`schema`) = '')"
	alibabaSchemaRow     = " WHERE id = ? AND category_type = ? AND category_id = ? AND BINARY scene = ? AND " + alibabaSchemaMissing
	alibabaErrorMaxBytes = 65535 // MySQL TEXT stores up to 65535 bytes.
)

func alibabaMissingSchemaRows(scene string) (string, []interface{}) {
	query := `FROM temp_alibaba_category_path
WHERE id > 0 AND category_id > 0 AND category_type IN (1, 2, 3)
AND BINARY scene IN ('processing', 'cbu', 'industry', 'popular')
AND ` + alibabaSchemaMissing
	switch scene {
	case "":
		return query, nil
	case "processing", "cbu":
		return query + " AND BINARY scene IN ('processing', 'cbu')", nil
	default:
		return query + " AND BINARY scene = ?", []interface{}{scene}
	}
}

func (*tempAlibabaCategoryPathDao) ListMissingSchema(ctx context.Context, session *dbsession.DBSession, afterID int64, limit int, scene string) ([]*erp_model.TempAlibabaCategoryPath, error) {
	from, args := alibabaMissingSchemaRows(scene)
	rows, err := session.DB.QueryContext(ctx, `SELECT id, category_type, category_id, scene
`+from+` AND id > ? ORDER BY id ASC LIMIT ?`, append(args, afterID, limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []*erp_model.TempAlibabaCategoryPath
	for rows.Next() {
		path := &erp_model.TempAlibabaCategoryPath{}
		if err := rows.Scan(&path.Id, &path.CategoryType, &path.CategoryId, &path.Scene); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

func (*tempAlibabaCategoryPathDao) CountMissingSchema(ctx context.Context, session *dbsession.DBSession, scene string) (map[string]int64, error) {
	from, args := alibabaMissingSchemaRows(scene)
	rows, err := session.DB.QueryContext(ctx, `SELECT scene, COUNT(*)
`+from+` GROUP BY scene`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var rowScene string
		var count int64
		if err := rows.Scan(&rowScene, &count); err != nil {
			return nil, err
		}
		counts[rowScene] = count
	}
	return counts, rows.Err()
}

func (*tempAlibabaCategoryPathDao) SaveSchemaIfMissing(ctx context.Context, session *dbsession.DBSession, path *erp_model.TempAlibabaCategoryPath, schema json.RawMessage) (bool, error) {
	return updateAlibabaSchemaRow(ctx, session, path,
		"UPDATE temp_alibaba_category_path SET `schema` = ?, error_msg = NULL"+alibabaSchemaRow, string(schema))
}

func (*tempAlibabaCategoryPathDao) SaveSchemaError(ctx context.Context, session *dbsession.DBSession, path *erp_model.TempAlibabaCategoryPath, message string) (bool, error) {
	return updateAlibabaSchemaRow(ctx, session, path,
		"UPDATE temp_alibaba_category_path SET error_msg = ?"+alibabaSchemaRow, alibabaSchemaErrorText(message))
}

func updateAlibabaSchemaRow(ctx context.Context, session *dbsession.DBSession, path *erp_model.TempAlibabaCategoryPath, query string, value string) (bool, error) {
	result, err := session.DB.ExecContext(ctx, query, value, path.Id, path.CategoryType, path.CategoryId, path.Scene)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func alibabaSchemaErrorText(message string) string {
	message = strings.ToValidUTF8(message, "\uFFFD")
	if len(message) <= alibabaErrorMaxBytes {
		return message
	}
	end := alibabaErrorMaxBytes
	for !utf8.RuneStart(message[end]) {
		end--
	}
	return message[:end]
}
