package model

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// VisitModel 访问统计表（手写，非 goctl 生成）。
// PV = visit 行数；UV = 按 天+ip_hash 去重后的访客数。
type VisitModel struct {
	conn sqlx.SqlConn
}

func NewVisitModel(conn sqlx.SqlConn) *VisitModel {
	return &VisitModel{conn: conn}
}

// Record 记录一次页面浏览
func (m *VisitModel) Record(ctx context.Context, date, ipHash, path string) error {
	_, err := m.conn.ExecCtx(ctx,
		"INSERT INTO visit (`date`, `ip_hash`, `path`) VALUES (?, ?, ?)", date, ipHash, path)
	return err
}

// DayStat 某一天的 PV/UV
type DayStat struct {
	Date string
	Pv   int64
	Uv   int64
}

// PvOn 某天的 PV；day 为空串表示全部时间
func (m *VisitModel) PvOn(ctx context.Context, day string) (int64, error) {
	query := "SELECT COUNT(*) FROM visit"
	args := []any{}
	if day != "" {
		query += " WHERE `date` = ?"
		args = append(args, day)
	}
	var n int64
	err := m.conn.QueryRowCtx(ctx, &n, query, args...)
	return n, err
}

// UvOn 某天的 UV（按 ip_hash 去重）；day 为空串表示全部时间
func (m *VisitModel) UvOn(ctx context.Context, day string) (int64, error) {
	query := "SELECT COUNT(DISTINCT `ip_hash`) FROM visit"
	args := []any{}
	if day != "" {
		query += " WHERE `date` = ?"
		args = append(args, day)
	}
	var n int64
	err := m.conn.QueryRowCtx(ctx, &n, query, args...)
	return n, err
}

// Trend 最近 days 天按天的 PV/UV，按日期升序；没访问的日子也补零，前端画图不断档
func (m *VisitModel) Trend(ctx context.Context, days int) ([]DayStat, error) {
	type row struct {
		Date string `db:"date"`
		Pv   int64  `db:"pv"`
		Uv   int64  `db:"uv"`
	}
	var rows []row
	// UV 不能简单 SUM 每天的 DISTINCT——同一天内去重才对，所以子查询按天分组算
	err := m.conn.QueryRowsCtx(ctx, &rows,
		"SELECT `date`, COUNT(*) AS pv, COUNT(DISTINCT `ip_hash`) AS uv FROM visit "+
			"WHERE `date` >= date('now', 'localtime', ?) GROUP BY `date` ORDER BY `date`",
		"-"+strconv.Itoa(days-1)+" day")
	if err != nil {
		return nil, err
	}

	byDate := make(map[string]DayStat, len(rows))
	for _, r := range rows {
		byDate[r.Date] = DayStat{Date: r.Date, Pv: r.Pv, Uv: r.Uv}
	}
	stats := make([]DayStat, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if s, ok := byDate[d]; ok {
			stats = append(stats, s)
		} else {
			stats = append(stats, DayStat{Date: d})
		}
	}
	return stats, nil
}

// Today 本地时区的今天（YYYY-MM-DD），visit.date 分桶键
func Today() string {
	return time.Now().Format("2006-01-02")
}
