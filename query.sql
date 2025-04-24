-- name: CreateSnapshot :one
INSERT INTO
    snapshot (deviceid, created_at, uptime)
VALUES
    (?, ?, ?) RETURNING *;

-- name: CreateSignal :one
INSERT INTO
    signal (
        snapshotid,
        generation,
        antenna_used,
        band,
        bars,
        cid,
        enbid,
        gnbid,
        rsrp,
        rsrq,
        rssi,
        sinr
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetDevice :one
SELECT
    *
FROM
    device
WHERE
    serial = ?
    AND software_version = ?;

-- name: CreateDevice :one
INSERT INTO
    device (
        friendly_name,
        hardware_version,
        isenabled,
        ismesh_supported,
        macid,
        manufacturer,
        manufacturer_oui,
        model,
        name,
        role,
        serial,
        software_version,
        type,
        update_state
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;
