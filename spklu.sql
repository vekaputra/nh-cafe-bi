-- SPKLU tables, loaded through /page/upload-spklu

CREATE TABLE IF NOT EXISTS spklu_stations (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY uq_spklu_stations_name (name)
);

CREATE TABLE IF NOT EXISTS spklu_evses (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  station_id BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY uq_spklu_evses_station_name (station_id, name),
  CONSTRAINT fk_spklu_evses_station FOREIGN KEY (station_id) REFERENCES spklu_stations (id)
);

CREATE TABLE IF NOT EXISTS spklu_connectors (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  evse_id BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY uq_spklu_connectors_evse_name (evse_id, name),
  CONSTRAINT fk_spklu_connectors_evse FOREIGN KEY (evse_id) REFERENCES spklu_evses (id)
);

-- started_at and stopped_at are UTC, the report prints UTC+7
CREATE TABLE IF NOT EXISTS spklu_charging_sessions (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  connector_id BIGINT NOT NULL,
  started_at DATETIME NOT NULL,
  stopped_at DATETIME NOT NULL,
  kwh DECIMAL(18, 6) NOT NULL,
  electricity_cost DECIMAL(18, 6) NOT NULL,
  pln_ppj DECIMAL(18, 6) NOT NULL,
  hhe_ppj DECIMAL(18, 6) NOT NULL,
  platform_share DECIMAL(18, 6) NOT NULL,
  charger_share DECIMAL(18, 6) NOT NULL,
  land_share DECIMAL(18, 6) NOT NULL,
  sharing_revenue DECIMAL(18, 6) NOT NULL,
  pln_ppn DECIMAL(18, 6) NOT NULL,
  hhe_ppn DECIMAL(18, 6) NOT NULL,
  total_ppn DECIMAL(18, 6) NOT NULL,
  stamp_duty DECIMAL(18, 6) NOT NULL,
  pln_revenue DECIMAL(18, 6) NOT NULL,
  hhe_revenue DECIMAL(18, 6) NOT NULL,
  admin_fee DECIMAL(18, 6) NOT NULL,
  total_revenue DECIMAL(18, 6) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY uq_spklu_charging_sessions_connector_started (connector_id, started_at),
  KEY idx_spklu_charging_sessions_started_at (started_at),
  CONSTRAINT fk_spklu_charging_sessions_connector FOREIGN KEY (connector_id) REFERENCES spklu_connectors (id)
);
