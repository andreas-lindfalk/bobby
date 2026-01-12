package store

import (
	"context"
)

// Gestoria represents a local administrative services office.
type Gestoria struct {
	ID        int
	Name      string
	City      string
	Address   string
	Phone     string
	Email     string
	Services  []string
	Languages []string
	Rating    float64
	Verified  bool
	Partner   bool
	Notes     string
}

// GetGestoriasByCity returns all gestorías in a given city.
func (s *Store) GetGestoriasByCity(ctx context.Context, city string) ([]Gestoria, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, city, COALESCE(address, ''), COALESCE(phone, ''), COALESCE(email, ''),
		       services, languages, COALESCE(rating, 0), verified, partner, COALESCE(notes, '')
		FROM gestorias
		WHERE city = $1
		ORDER BY partner DESC, rating DESC
	`, city)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gestorias []Gestoria
	for rows.Next() {
		var g Gestoria
		if err := rows.Scan(&g.ID, &g.Name, &g.City, &g.Address, &g.Phone, &g.Email,
			&g.Services, &g.Languages, &g.Rating, &g.Verified, &g.Partner, &g.Notes); err != nil {
			return nil, err
		}
		gestorias = append(gestorias, g)
	}

	return gestorias, rows.Err()
}

// GetVerifiedGestoriaForService returns verified gestorías that offer a specific service.
func (s *Store) GetVerifiedGestoriaForService(ctx context.Context, service string, languages []string) ([]Gestoria, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, city, COALESCE(address, ''), COALESCE(phone, ''), COALESCE(email, ''),
		       services, languages, COALESCE(rating, 0), verified, partner, COALESCE(notes, '')
		FROM gestorias
		WHERE verified = TRUE 
		  AND $1 = ANY(services)
		  AND languages && $2
		ORDER BY partner DESC, rating DESC
	`, service, languages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gestorias []Gestoria
	for rows.Next() {
		var g Gestoria
		if err := rows.Scan(&g.ID, &g.Name, &g.City, &g.Address, &g.Phone, &g.Email,
			&g.Services, &g.Languages, &g.Rating, &g.Verified, &g.Partner, &g.Notes); err != nil {
			return nil, err
		}
		gestorias = append(gestorias, g)
	}

	return gestorias, rows.Err()
}
