# Overview
** This is a work in progress and only intended for hobbyists. **

Collect statistics about your 5G TMobile Home Internet connection:
- RSRP
- RSRQ
- RSSI
- SINR

# Setup
Prerequisites:
- Go 1.24
- [SQLC](https://docs.sqlc.dev/en/latest/overview/install.html)
- [SQLite](https://www.sqlite.org/index.html)

Install dependencies:
```commandline
go mod tidy
```

# Usage

## Collect Statistics
```commandline
>> export GATEWAY_USERNAME=your_username
>> export GATEWAY_PASSWORD=your_password
>> export GATEWAY_POLL_FREQ=1m

# Begin polling and persist the statistics to a sqlite database, `tmo.db`.
>> go run main.go
2025/04/23 21:36:59 Gateway URL: http://192.168.12.1/TMI/v1
2025/04/23 21:36:59 POST auth/login
2025/04/23 21:36:59 GET gateway/?get=all
2025/04/23 21:38:00 GET gateway/?get=all
```

## Query Statistics
```commandline
sqlite3 tmo.db

# List all of your gateway devices
# Normally there is only 1 device
# However there will be multiple devices if:
# - you receive a new physical gateway
# - or your gateway receives a firmware update
sqlite> SELECT * FROM device;

# List all snapshots
SElECT * FROM snapshots;

# List 5g/4g signal stats
SELECT * FROM signal WHERE generation = '5G';
SELECT * FROM signal WHERE generation = '4G';
```

# Development

## Tests
```commandline
go test ./...
go test -bench=.
```

## Changing the database schema
The database schema is defined in `schema.sql`.
The database Go functions are defined in `query.sql`.

If you change the schema or query functions, you must:
1. Migrate your database
```commandline
go run cmds/db/cli.go migrate -d tmo.db
```

2. Regenerate the Go SQL code
```commandline
sqlc generate
```

## The API GET gateway/?get=all Response
```json
{
  "device": {
    "friendlyName": "5G Gateway",
    "hardwareVersion": "R02",
    "index": 1,
    "isEnabled": true,
    "isMeshSupported": true,
    "macId": "bc:12:34:56:78:90",
    "manufacturer": "Sercomm",
    "manufacturerOUI": "00C002",
    "model": "TMO-G4SE",
    "name": "5G Gateway",
    "role": "gateway",
    "serial": "XXXXXXXXX",
    "softwareVersion": "1.03.20",
    "type": "HSID",
    "updateState": "latest"
  },
  "signal": {
    "4g": {
      "antennaUsed": "Internal_directional",
      "bands": [
        "b66"
      ],
      "bars": 3.0,
      "cid": 32,
      "eNBID": 23270,
      "rsrp": -102,
      "rsrq": -7,
      "rssi": -94,
      "sinr": 15
    },
    "5g": {
      "antennaUsed": "Internal_directional",
      "bands": [
        "n41"
      ],
      "bars": 3.0,
      "cid": 32,
      "gNBID": 23270,
      "rsrp": -105,
      "rsrq": -10,
      "rssi": -93,
      "sinr": 13
    },
    "generic": {
      "apn": "FBB.HOME",
      "hasIPv6": true,
      "registration": "registered",
      "roaming": false
    }
  },
  "time": {
    "daylightSavings": {
      "isUsed": false
    },
    "localTime": 1739945734,
    "localTimeZone": "-05:00",
    "upTime": 27653
  }
}
```

## TODO
- [X] Create Device Table
- [X] Refactor API - Create Client Struct
- [X] API Client Unit Tests
- [ ] Polling Unit Tests
- [ ] Investigate why 20 threads are used
- [ ] Use context to handle db/request timeouts
