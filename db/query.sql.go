// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: query.sql

package db

import (
	"context"
	"time"
)

const createDevice = `-- name: CreateDevice :one
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
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, friendly_name, hardware_version, isenabled, ismesh_supported, macid, manufacturer, manufacturer_oui, model, name, role, serial, software_version, type, update_state
`

type CreateDeviceParams struct {
	FriendlyName    string
	HardwareVersion string
	Isenabled       bool
	IsmeshSupported bool
	Macid           string
	Manufacturer    string
	ManufacturerOui string
	Model           string
	Name            string
	Role            string
	Serial          string
	SoftwareVersion string
	Type            string
	UpdateState     string
}

func (q *Queries) CreateDevice(ctx context.Context, arg CreateDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, createDevice,
		arg.FriendlyName,
		arg.HardwareVersion,
		arg.Isenabled,
		arg.IsmeshSupported,
		arg.Macid,
		arg.Manufacturer,
		arg.ManufacturerOui,
		arg.Model,
		arg.Name,
		arg.Role,
		arg.Serial,
		arg.SoftwareVersion,
		arg.Type,
		arg.UpdateState,
	)
	var i Device
	err := row.Scan(
		&i.ID,
		&i.FriendlyName,
		&i.HardwareVersion,
		&i.Isenabled,
		&i.IsmeshSupported,
		&i.Macid,
		&i.Manufacturer,
		&i.ManufacturerOui,
		&i.Model,
		&i.Name,
		&i.Role,
		&i.Serial,
		&i.SoftwareVersion,
		&i.Type,
		&i.UpdateState,
	)
	return i, err
}

const createSignal = `-- name: CreateSignal :one
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
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, snapshotid, antenna_used, generation, band, bars, cid, enbid, gnbid, rsrp, rsrq, rssi, sinr
`

type CreateSignalParams struct {
	Snapshotid  int64
	Generation  string
	AntennaUsed string
	Band        string
	Bars        float64
	Cid         int64
	Enbid       int64
	Gnbid       int64
	Rsrp        int64
	Rsrq        int64
	Rssi        int64
	Sinr        int64
}

func (q *Queries) CreateSignal(ctx context.Context, arg CreateSignalParams) (Signal, error) {
	row := q.db.QueryRowContext(ctx, createSignal,
		arg.Snapshotid,
		arg.Generation,
		arg.AntennaUsed,
		arg.Band,
		arg.Bars,
		arg.Cid,
		arg.Enbid,
		arg.Gnbid,
		arg.Rsrp,
		arg.Rsrq,
		arg.Rssi,
		arg.Sinr,
	)
	var i Signal
	err := row.Scan(
		&i.ID,
		&i.Snapshotid,
		&i.AntennaUsed,
		&i.Generation,
		&i.Band,
		&i.Bars,
		&i.Cid,
		&i.Enbid,
		&i.Gnbid,
		&i.Rsrp,
		&i.Rsrq,
		&i.Rssi,
		&i.Sinr,
	)
	return i, err
}

const createSnapshot = `-- name: CreateSnapshot :one
INSERT INTO
    snapshot (deviceid, created_at, uptime)
VALUES
    (?, ?, ?) RETURNING id, deviceid, created_at, uptime
`

type CreateSnapshotParams struct {
	Deviceid  int64
	CreatedAt time.Time
	Uptime    int64
}

func (q *Queries) CreateSnapshot(ctx context.Context, arg CreateSnapshotParams) (Snapshot, error) {
	row := q.db.QueryRowContext(ctx, createSnapshot, arg.Deviceid, arg.CreatedAt, arg.Uptime)
	var i Snapshot
	err := row.Scan(
		&i.ID,
		&i.Deviceid,
		&i.CreatedAt,
		&i.Uptime,
	)
	return i, err
}

const getDevice = `-- name: GetDevice :one
SELECT
    id, friendly_name, hardware_version, isenabled, ismesh_supported, macid, manufacturer, manufacturer_oui, model, name, role, serial, software_version, type, update_state
FROM
    device
WHERE
    serial = ?
    AND software_version = ?
`

type GetDeviceParams struct {
	Serial          string
	SoftwareVersion string
}

func (q *Queries) GetDevice(ctx context.Context, arg GetDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, getDevice, arg.Serial, arg.SoftwareVersion)
	var i Device
	err := row.Scan(
		&i.ID,
		&i.FriendlyName,
		&i.HardwareVersion,
		&i.Isenabled,
		&i.IsmeshSupported,
		&i.Macid,
		&i.Manufacturer,
		&i.ManufacturerOui,
		&i.Model,
		&i.Name,
		&i.Role,
		&i.Serial,
		&i.SoftwareVersion,
		&i.Type,
		&i.UpdateState,
	)
	return i, err
}
