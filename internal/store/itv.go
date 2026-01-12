package store

import (
	"context"
)

// ITVStation represents an ITV (vehicle inspection) station.
type ITVStation struct {
	ID             int
	Name           string
	City           string
	Address        string
	Phone          string
	AppointmentURL string
	AvgWaitDays    int
	HandlesImports bool
	Notes          string
}

// GetITVStationsByCity returns ITV stations near a city, ordered by wait time.
func (s *Store) GetITVStationsByCity(ctx context.Context, city string) ([]ITVStation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, city, COALESCE(address, ''), COALESCE(phone, ''),
		       COALESCE(appointment_url, ''), COALESCE(avg_wait_days, 0), handles_imports, COALESCE(notes, '')
		FROM itv_stations
		WHERE handles_imports = TRUE
		ORDER BY avg_wait_days ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []ITVStation
	for rows.Next() {
		var st ITVStation
		if err := rows.Scan(&st.ID, &st.Name, &st.City, &st.Address, &st.Phone,
			&st.AppointmentURL, &st.AvgWaitDays, &st.HandlesImports, &st.Notes); err != nil {
			return nil, err
		}
		stations = append(stations, st)
	}

	return stations, rows.Err()
}

// GetBestITVForImports returns the ITV station with shortest wait time.
func (s *Store) GetBestITVForImports(ctx context.Context) (*ITVStation, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, city, COALESCE(address, ''), COALESCE(phone, ''),
		       COALESCE(appointment_url, ''), COALESCE(avg_wait_days, 0), handles_imports, COALESCE(notes, '')
		FROM itv_stations
		WHERE handles_imports = TRUE
		ORDER BY avg_wait_days ASC
		LIMIT 1
	`)

	var st ITVStation
	if err := row.Scan(&st.ID, &st.Name, &st.City, &st.Address, &st.Phone,
		&st.AppointmentURL, &st.AvgWaitDays, &st.HandlesImports, &st.Notes); err != nil {
		return nil, err
	}

	return &st, nil
}
