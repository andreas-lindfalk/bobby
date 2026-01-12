-- 003_seeds.sql: Seed data for Orihuela Costa area

-- Gestorías
INSERT INTO gestorias (name, city, address, phone, email, services, languages, rating, verified, partner, notes) VALUES
('Gestoría López & Partners', 'Orihuela Costa', 'Calle del Mar 15, La Zenia', '+34 966 123 456', 'info@lopezpartners.es', 
 ARRAY['matriculation', 'NIE', 'residencia', 'taxes'], ARRAY['es', 'en', 'sv'], 4.8, TRUE, TRUE, 
 'Our recommended partner. Specializes in Swedish expats. 15% discount for platform users.'),
 
('Gestoría García', 'Torrevieja', 'Av. Habaneras 45', '+34 966 789 012', 'contacto@gestoriagarcia.com',
 ARRAY['matriculation', 'NIE', 'company_formation'], ARRAY['es', 'en', 'de'], 4.2, TRUE, FALSE,
 'Good for German speakers. Long experience with car imports.'),
 
('Costa Blanca Admin Services', 'Pilar de la Horadada', 'Plaza Mayor 3', '+34 966 345 678', 'hello@cbadmin.es',
 ARRAY['matriculation', 'residencia', 'padron'], ARRAY['es', 'en'], 4.5, TRUE, FALSE,
 'British-run. Good English support.');

-- ITV Stations  
INSERT INTO itv_stations (name, city, address, phone, appointment_url, avg_wait_days, handles_imports, notes) VALUES
('ITV Orihuela', 'Orihuela', 'Polígono Industrial La Murada', '+34 966 111 222', 'https://itv.es/alicante/orihuela', 
 14, TRUE, 'Closest to Orihuela Costa. Book online 2 weeks ahead.'),
 
('ITV Torrevieja', 'Torrevieja', 'Av. de la Estación s/n', '+34 966 333 444', 'https://itv.es/alicante/torrevieja',
 7, TRUE, 'Shorter wait times. Good for urgent cases.'),
 
('ITV Elche', 'Elche', 'Polígono Carrús', '+34 966 555 666', 'https://itv.es/alicante/elche',
 21, TRUE, 'Larger station but longer waits.'),
 
('ITV Alicante', 'Alicante', 'Polígono Babel', '+34 966 777 888', 'https://itv.es/alicante/alicante',
 10, TRUE, 'Main provincial station. All services available.');

-- Offices
INSERT INTO offices (type, name, city, address, phone, appointment_url, notes) VALUES
('trafico', 'Jefatura Provincial de Tráfico Alicante', 'Alicante', 'C/ Churruca 24, 03003 Alicante', '+34 965 935 800',
 'https://sedeapl.dgt.gob.es/WEB_NCIT_CONSULTA/sol498.faces', 'Main traffic office. All matriculation services.'),
 
('hacienda', 'Agencia Tributaria Torrevieja', 'Torrevieja', 'Av. de las Habaneras 78', '+34 966 920 700',
 'https://www.agenciatributaria.es', 'For ITVM (registration tax) payment.'),
 
('padron', 'Oficina Padrón Orihuela Costa', 'Orihuela Costa', 'C/ Maestro Albéniz 2, La Zenia', '+34 966 760 000',
 'https://orihuela.es/padron', 'Municipal census registration.'),
 
('foreigners', 'Oficina de Extranjeros Alicante', 'Alicante', 'C/ Ebanistería 4, 03008 Alicante', '+34 965 936 840',
 'https://sede.administracionespublicas.gob.es', 'For TIE/NIE appointments.');
