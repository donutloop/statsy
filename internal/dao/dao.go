package dao

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

type DAO struct {
	db *sql.DB
}

func New(db *sql.DB) *DAO {
	return &DAO{
		db: db,
	}
}

func (d *DAO) FindBlacklistedIP(ipRaw string) (bool, error) {

	ipArg, err := strconv.Atoi(strings.Replace(ipRaw, ".", "", -1))
	if err != nil {
		return false, err
	}

	qlStatement := `SELECT ip FROM ip_blacklist WHERE ip = ?;`

	var ip uint64
	row := d.db.QueryRow(qlStatement, ipArg)
	err = row.Scan(&ip)
	if err != nil {
		if sql.ErrNoRows == err {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (d *DAO) FindBlackedlistedUserAgent(userAgentArg string) (bool, error) {

	qlStatement := `SELECT ua FROM ua_blacklist WHERE ua = ?;`

	var userAgent string
	row := d.db.QueryRow(qlStatement, userAgentArg)
	err := row.Scan(&userAgent)
	if err != nil {
		if sql.ErrNoRows == err {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

var CustomerNotFound error = errors.New("customer not found")

func (d *DAO) IsCustomerActive(customerId int64) (bool, error) {

	qlStatement := `SELECT active FROM customer WHERE id = ?;`

	var customerActive bool
	row := d.db.QueryRow(qlStatement, customerId)
	err := row.Scan(&customerActive)
	if err != nil {
		if sql.ErrNoRows == err {
			return false, CustomerNotFound
		}
		return false, err
	}

	return customerActive, nil
}

type CustomerDayStatistics struct {
	CustomerHourStatistics []CustomerHourStatistics `json:"customer_hour_statistics"`
	TotalRequestsPerDay    int64                    `json:"total_requests_per_day"`
}

type CustomerHourStatistics struct {
	RequestCount int64 `json:"request_count"`
	InvalidCount int64 `json:"invalid_count"`
	Time         int64 `json:"time"`
}

func (d *DAO) GetCutsomerRequestCountByDayAndID(customerIdArg int64, dayArg int64) (*CustomerDayStatistics, error) {

	t := time.Unix(dayArg, 0)
	year, month, day := t.Date()
	beginTimeRange := time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Unix()

	year, month, day = t.Date()
	endTimeRange := time.Date(year, month, day, 23, 59, 59, 0, t.Location()).Unix()

	qlStatement := `SELECT request_count, invalid_count, time FROM hourly_stats WHERE customer_id = ? AND (time >= ? AND time <= ?);`

	rows, err := d.db.Query(qlStatement, customerIdArg, beginTimeRange, endTimeRange)
	if err != nil {
		if sql.ErrNoRows == err {
			return nil, CustomerNotFound
		}
		return nil, err
	}

	defer rows.Close()

	customerStatistics := new(CustomerDayStatistics)
	var totalRequestsPerDay int64
	for rows.Next() {
		var requestCount int64
		var invalidCount int64
		var time int64
		err := rows.Scan(&requestCount, &invalidCount, &time)
		if err != nil {
			if sql.ErrNoRows == err {
				return nil, CustomerNotFound
			}
			return nil, err
		}
		totalRequestsPerDay += requestCount + invalidCount

		customerStatistics.CustomerHourStatistics = append(customerStatistics.CustomerHourStatistics, CustomerHourStatistics{RequestCount: requestCount, InvalidCount: invalidCount, Time: time})
	}

	customerStatistics.TotalRequestsPerDay = totalRequestsPerDay

	return customerStatistics, nil
}

func (d *DAO) InsertOrUpdateCustomerStats(customerIdArg int64, valid bool, timestamp int64) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	t := time.Unix(timestamp, 0)
	year, month, day := t.Date()

	timerange := time.Date(year, month, day, t.Hour(), 0, 0, 0, t.Location()).Unix()

	qlStatement := `SELECT request_count, invalid_count FROM hourly_stats AS hs JOIN customer AS c ON hs.customer_id=c.id WHERE c.id = ? AND hs.time = ? FOR UPDATE;`

	var requestCount, invalidCount int64
	row := tx.QueryRow(qlStatement, customerIdArg, timerange)
	err = row.Scan(&requestCount, &invalidCount)
	if err != nil && sql.ErrNoRows != err {
		tx.Rollback()
		return err
	}

	if valid {
		requestCount++
	} else {
		invalidCount++
	}

	if err == sql.ErrNoRows {
		qlStatement = `INSERT INTO hourly_stats (customer_id, request_count, invalid_count, time) VALUES (?, ?, ?, ?);`
		_, err := tx.Exec(qlStatement, customerIdArg, requestCount, invalidCount, timerange)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		qlStatement = `UPDATE hourly_stats SET request_count = ?, invalid_count = ? WHERE customer_id = ? AND time = ?;`
		result, err := tx.Exec(qlStatement, requestCount, invalidCount, customerIdArg, timerange)
		if err != nil {
			tx.Rollback()
			if sql.ErrNoRows == err {
				return CustomerNotFound
			}
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		if rowsAffected != 1 {
			tx.Rollback()
			panic("to many rows affected")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
