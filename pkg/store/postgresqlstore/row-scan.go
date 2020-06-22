package postgresqlstore

// scanIDRow scans a given row for an ID
func scanIDRow(row rowScanner) (int64, error) {
	var id int64

	err := row.Scan(&id)

	if err != nil {
		return -1, err
	}

	return id, err
}
