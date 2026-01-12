-- 001_initial.sql: Base schema for gestorías, ITV stations, and offices

CREATE TABLE IF NOT EXISTS gestorias (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT,
    phone TEXT,
    email TEXT,
    services TEXT[], -- e.g., {'matriculation', 'NIE', 'residencia'}
    languages TEXT[], -- e.g., {'es', 'en', 'sv', 'de'}
    rating DECIMAL(2,1),
    verified BOOLEAN DEFAULT FALSE,
    partner BOOLEAN DEFAULT FALSE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gestorias_city ON gestorias(city);
CREATE INDEX IF NOT EXISTS idx_gestorias_services ON gestorias USING GIN(services);

CREATE TABLE IF NOT EXISTS itv_stations (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT,
    phone TEXT,
    appointment_url TEXT,
    avg_wait_days INTEGER, -- Average wait time for appointment
    handles_imports BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_itv_city ON itv_stations(city);

CREATE TABLE IF NOT EXISTS offices (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL, -- 'trafico', 'hacienda', 'padron', 'foreigners'
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT,
    phone TEXT,
    appointment_url TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_offices_type_city ON offices(type, city);
