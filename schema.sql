CREATE TABLE IF NOT EXISTS device (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    friendly_name VARCHAR(50) NOT NULL,
    hardware_version VARCHAR(10) NOT NULL,
    isenabled BOOLEAN NOT NULL,
    ismesh_supported BOOLEAN NOT NULL,
    macid VARCHAR(20) NOT NULL,
    manufacturer VARCHAR(50) NOT NULL,
    manufacturer_oui VARCHAR(10) NOT NULL,
    model VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL,
    serial VARCHAR(50) NOT NULL,
    software_version VARCHAR(10) NOT NULL,
    type VARCHAR(20) NOT NULL,
    update_state VARCHAR(10) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_device_serial_software_version ON device (serial, software_version);

CREATE TABLE IF NOT EXISTS snapshot (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    deviceid INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    uptime INT NOT NULL,
    FOREIGN KEY (deviceid) REFERENCES device (id)
);

CREATE TABLE IF NOT EXISTS signal (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    snapshotid INT NOT NULL,
    antenna_used VARCHAR(50) NOT NULL,
    generation VARCHAR(2) NOT NULL,
    band VARCHAR(4) NOT NULL,
    bars FLOAT NOT NULL,
    cid INT NOT NULL,
    enbid INT NOT NULL,
    gnbid INT NOT NULL,
    rsrp INT NOT NULL,
    rsrq INT NOT NULL,
    rssi INT NOT NULL,
    sinr INT NOT NULL,
    FOREIGN KEY (snapshotid) REFERENCES snapshot (id)
);

CREATE INDEX IF NOT EXISTS ix_signal_generation ON signal (generation);

CREATE INDEX IF NOT EXISTS ix_signal_band ON signal (band);
