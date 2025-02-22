// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package gen

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/TencentBlueKing/bk-bscp/pkg/dal/table"
)

func newDataSourceInfo(db *gorm.DB, opts ...gen.DOOption) dataSourceInfo {
	_dataSourceInfo := dataSourceInfo{}

	_dataSourceInfo.dataSourceInfoDo.UseDB(db, opts...)
	_dataSourceInfo.dataSourceInfoDo.UseModel(&table.DataSourceInfo{})

	tableName := _dataSourceInfo.dataSourceInfoDo.TableName()
	_dataSourceInfo.ALL = field.NewAsterisk(tableName)
	_dataSourceInfo.ID = field.NewUint32(tableName, "id")
	_dataSourceInfo.BizID = field.NewUint32(tableName, "biz_id")
	_dataSourceInfo.Name = field.NewString(tableName, "name")
	_dataSourceInfo.Memo = field.NewString(tableName, "memo")
	_dataSourceInfo.SourceType = field.NewString(tableName, "source_type")
	_dataSourceInfo.Dsn = field.NewString(tableName, "dsn")
	_dataSourceInfo.Creator = field.NewString(tableName, "creator")
	_dataSourceInfo.Reviser = field.NewString(tableName, "reviser")
	_dataSourceInfo.CreatedAt = field.NewTime(tableName, "created_at")
	_dataSourceInfo.UpdatedAt = field.NewTime(tableName, "updated_at")

	_dataSourceInfo.fillFieldMap()

	return _dataSourceInfo
}

type dataSourceInfo struct {
	dataSourceInfoDo dataSourceInfoDo

	ALL        field.Asterisk
	ID         field.Uint32
	BizID      field.Uint32
	Name       field.String
	Memo       field.String
	SourceType field.String
	Dsn        field.String
	Creator    field.String
	Reviser    field.String
	CreatedAt  field.Time
	UpdatedAt  field.Time

	fieldMap map[string]field.Expr
}

func (d dataSourceInfo) Table(newTableName string) *dataSourceInfo {
	d.dataSourceInfoDo.UseTable(newTableName)
	return d.updateTableName(newTableName)
}

func (d dataSourceInfo) As(alias string) *dataSourceInfo {
	d.dataSourceInfoDo.DO = *(d.dataSourceInfoDo.As(alias).(*gen.DO))
	return d.updateTableName(alias)
}

func (d *dataSourceInfo) updateTableName(table string) *dataSourceInfo {
	d.ALL = field.NewAsterisk(table)
	d.ID = field.NewUint32(table, "id")
	d.BizID = field.NewUint32(table, "biz_id")
	d.Name = field.NewString(table, "name")
	d.Memo = field.NewString(table, "memo")
	d.SourceType = field.NewString(table, "source_type")
	d.Dsn = field.NewString(table, "dsn")
	d.Creator = field.NewString(table, "creator")
	d.Reviser = field.NewString(table, "reviser")
	d.CreatedAt = field.NewTime(table, "created_at")
	d.UpdatedAt = field.NewTime(table, "updated_at")

	d.fillFieldMap()

	return d
}

func (d *dataSourceInfo) WithContext(ctx context.Context) IDataSourceInfoDo {
	return d.dataSourceInfoDo.WithContext(ctx)
}

func (d dataSourceInfo) TableName() string { return d.dataSourceInfoDo.TableName() }

func (d dataSourceInfo) Alias() string { return d.dataSourceInfoDo.Alias() }

func (d dataSourceInfo) Columns(cols ...field.Expr) gen.Columns {
	return d.dataSourceInfoDo.Columns(cols...)
}

func (d *dataSourceInfo) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := d.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (d *dataSourceInfo) fillFieldMap() {
	d.fieldMap = make(map[string]field.Expr, 10)
	d.fieldMap["id"] = d.ID
	d.fieldMap["biz_id"] = d.BizID
	d.fieldMap["name"] = d.Name
	d.fieldMap["memo"] = d.Memo
	d.fieldMap["source_type"] = d.SourceType
	d.fieldMap["dsn"] = d.Dsn
	d.fieldMap["creator"] = d.Creator
	d.fieldMap["reviser"] = d.Reviser
	d.fieldMap["created_at"] = d.CreatedAt
	d.fieldMap["updated_at"] = d.UpdatedAt
}

func (d dataSourceInfo) clone(db *gorm.DB) dataSourceInfo {
	d.dataSourceInfoDo.ReplaceConnPool(db.Statement.ConnPool)
	return d
}

func (d dataSourceInfo) replaceDB(db *gorm.DB) dataSourceInfo {
	d.dataSourceInfoDo.ReplaceDB(db)
	return d
}

type dataSourceInfoDo struct{ gen.DO }

type IDataSourceInfoDo interface {
	gen.SubQuery
	Debug() IDataSourceInfoDo
	WithContext(ctx context.Context) IDataSourceInfoDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IDataSourceInfoDo
	WriteDB() IDataSourceInfoDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IDataSourceInfoDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IDataSourceInfoDo
	Not(conds ...gen.Condition) IDataSourceInfoDo
	Or(conds ...gen.Condition) IDataSourceInfoDo
	Select(conds ...field.Expr) IDataSourceInfoDo
	Where(conds ...gen.Condition) IDataSourceInfoDo
	Order(conds ...field.Expr) IDataSourceInfoDo
	Distinct(cols ...field.Expr) IDataSourceInfoDo
	Omit(cols ...field.Expr) IDataSourceInfoDo
	Join(table schema.Tabler, on ...field.Expr) IDataSourceInfoDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IDataSourceInfoDo
	RightJoin(table schema.Tabler, on ...field.Expr) IDataSourceInfoDo
	Group(cols ...field.Expr) IDataSourceInfoDo
	Having(conds ...gen.Condition) IDataSourceInfoDo
	Limit(limit int) IDataSourceInfoDo
	Offset(offset int) IDataSourceInfoDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IDataSourceInfoDo
	Unscoped() IDataSourceInfoDo
	Create(values ...*table.DataSourceInfo) error
	CreateInBatches(values []*table.DataSourceInfo, batchSize int) error
	Save(values ...*table.DataSourceInfo) error
	First() (*table.DataSourceInfo, error)
	Take() (*table.DataSourceInfo, error)
	Last() (*table.DataSourceInfo, error)
	Find() ([]*table.DataSourceInfo, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*table.DataSourceInfo, err error)
	FindInBatches(result *[]*table.DataSourceInfo, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*table.DataSourceInfo) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IDataSourceInfoDo
	Assign(attrs ...field.AssignExpr) IDataSourceInfoDo
	Joins(fields ...field.RelationField) IDataSourceInfoDo
	Preload(fields ...field.RelationField) IDataSourceInfoDo
	FirstOrInit() (*table.DataSourceInfo, error)
	FirstOrCreate() (*table.DataSourceInfo, error)
	FindByPage(offset int, limit int) (result []*table.DataSourceInfo, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IDataSourceInfoDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (d dataSourceInfoDo) Debug() IDataSourceInfoDo {
	return d.withDO(d.DO.Debug())
}

func (d dataSourceInfoDo) WithContext(ctx context.Context) IDataSourceInfoDo {
	return d.withDO(d.DO.WithContext(ctx))
}

func (d dataSourceInfoDo) ReadDB() IDataSourceInfoDo {
	return d.Clauses(dbresolver.Read)
}

func (d dataSourceInfoDo) WriteDB() IDataSourceInfoDo {
	return d.Clauses(dbresolver.Write)
}

func (d dataSourceInfoDo) Session(config *gorm.Session) IDataSourceInfoDo {
	return d.withDO(d.DO.Session(config))
}

func (d dataSourceInfoDo) Clauses(conds ...clause.Expression) IDataSourceInfoDo {
	return d.withDO(d.DO.Clauses(conds...))
}

func (d dataSourceInfoDo) Returning(value interface{}, columns ...string) IDataSourceInfoDo {
	return d.withDO(d.DO.Returning(value, columns...))
}

func (d dataSourceInfoDo) Not(conds ...gen.Condition) IDataSourceInfoDo {
	return d.withDO(d.DO.Not(conds...))
}

func (d dataSourceInfoDo) Or(conds ...gen.Condition) IDataSourceInfoDo {
	return d.withDO(d.DO.Or(conds...))
}

func (d dataSourceInfoDo) Select(conds ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.Select(conds...))
}

func (d dataSourceInfoDo) Where(conds ...gen.Condition) IDataSourceInfoDo {
	return d.withDO(d.DO.Where(conds...))
}

func (d dataSourceInfoDo) Order(conds ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.Order(conds...))
}

func (d dataSourceInfoDo) Distinct(cols ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.Distinct(cols...))
}

func (d dataSourceInfoDo) Omit(cols ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.Omit(cols...))
}

func (d dataSourceInfoDo) Join(table schema.Tabler, on ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.Join(table, on...))
}

func (d dataSourceInfoDo) LeftJoin(table schema.Tabler, on ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.LeftJoin(table, on...))
}

func (d dataSourceInfoDo) RightJoin(table schema.Tabler, on ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.RightJoin(table, on...))
}

func (d dataSourceInfoDo) Group(cols ...field.Expr) IDataSourceInfoDo {
	return d.withDO(d.DO.Group(cols...))
}

func (d dataSourceInfoDo) Having(conds ...gen.Condition) IDataSourceInfoDo {
	return d.withDO(d.DO.Having(conds...))
}

func (d dataSourceInfoDo) Limit(limit int) IDataSourceInfoDo {
	return d.withDO(d.DO.Limit(limit))
}

func (d dataSourceInfoDo) Offset(offset int) IDataSourceInfoDo {
	return d.withDO(d.DO.Offset(offset))
}

func (d dataSourceInfoDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IDataSourceInfoDo {
	return d.withDO(d.DO.Scopes(funcs...))
}

func (d dataSourceInfoDo) Unscoped() IDataSourceInfoDo {
	return d.withDO(d.DO.Unscoped())
}

func (d dataSourceInfoDo) Create(values ...*table.DataSourceInfo) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Create(values)
}

func (d dataSourceInfoDo) CreateInBatches(values []*table.DataSourceInfo, batchSize int) error {
	return d.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (d dataSourceInfoDo) Save(values ...*table.DataSourceInfo) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Save(values)
}

func (d dataSourceInfoDo) First() (*table.DataSourceInfo, error) {
	if result, err := d.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*table.DataSourceInfo), nil
	}
}

func (d dataSourceInfoDo) Take() (*table.DataSourceInfo, error) {
	if result, err := d.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*table.DataSourceInfo), nil
	}
}

func (d dataSourceInfoDo) Last() (*table.DataSourceInfo, error) {
	if result, err := d.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*table.DataSourceInfo), nil
	}
}

func (d dataSourceInfoDo) Find() ([]*table.DataSourceInfo, error) {
	result, err := d.DO.Find()
	return result.([]*table.DataSourceInfo), err
}

func (d dataSourceInfoDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*table.DataSourceInfo, err error) {
	buf := make([]*table.DataSourceInfo, 0, batchSize)
	err = d.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (d dataSourceInfoDo) FindInBatches(result *[]*table.DataSourceInfo, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return d.DO.FindInBatches(result, batchSize, fc)
}

func (d dataSourceInfoDo) Attrs(attrs ...field.AssignExpr) IDataSourceInfoDo {
	return d.withDO(d.DO.Attrs(attrs...))
}

func (d dataSourceInfoDo) Assign(attrs ...field.AssignExpr) IDataSourceInfoDo {
	return d.withDO(d.DO.Assign(attrs...))
}

func (d dataSourceInfoDo) Joins(fields ...field.RelationField) IDataSourceInfoDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Joins(_f))
	}
	return &d
}

func (d dataSourceInfoDo) Preload(fields ...field.RelationField) IDataSourceInfoDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Preload(_f))
	}
	return &d
}

func (d dataSourceInfoDo) FirstOrInit() (*table.DataSourceInfo, error) {
	if result, err := d.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*table.DataSourceInfo), nil
	}
}

func (d dataSourceInfoDo) FirstOrCreate() (*table.DataSourceInfo, error) {
	if result, err := d.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*table.DataSourceInfo), nil
	}
}

func (d dataSourceInfoDo) FindByPage(offset int, limit int) (result []*table.DataSourceInfo, count int64, err error) {
	result, err = d.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = d.Offset(-1).Limit(-1).Count()
	return
}

func (d dataSourceInfoDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	err = d.Offset(offset).Limit(limit).Scan(result)
	return
}

func (d dataSourceInfoDo) Scan(result interface{}) (err error) {
	return d.DO.Scan(result)
}

func (d dataSourceInfoDo) Delete(models ...*table.DataSourceInfo) (result gen.ResultInfo, err error) {
	return d.DO.Delete(models)
}

func (d *dataSourceInfoDo) withDO(do gen.Dao) *dataSourceInfoDo {
	d.DO = *do.(*gen.DO)
	return d
}
