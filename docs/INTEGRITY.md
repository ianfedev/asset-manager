# Integrity Checks

## Overview
The Asset Manager includes tools to verify the integrity of the storage bucket structure and furniture assets. This ensures that all required directories exist and furniture assets are consistent across FurniData, Storage, and Database.

## Required Structure
The S3 bucket must contain the following top-level "folders" (prefixes):
- `bundled/`
- `c_images/`
- `dcr/`
- `gamedata/`
- `images/`
- `logos/`
- `sounds/`

## Usage

### CLI

#### Structure Checks
Check integrity:
```bash
go run main.go integrity
```

Check and fix missing folders:
```bash
go run main.go integrity structure --fix
```

#### Furniture Integrity Check

> [!IMPORTANT]
> **Database Connection Required**: The furniture integrity check requires a valid database connection configured in your `.env` file. The check will fail if the database is not accessible.

Check furniture asset integrity:
```bash
go run main.go integrity furniture
```

Generate JSON report file:
```bash
go run main.go integrity furniture --json
```

> [!NOTE]
> Without the `--json` flag, only console metrics are displayed. Use `--json` to generate the detailed JSON report file.

The furniture check validates:
- **FurniData.json**: All items are properly defined with required fields
- **Storage**: All expected `.nitro` files exist in `bundled/furniture/`
- **Database**: All items are registered with matching parameters

**Console Output Example:**
```
INFO    Furniture Integrity Report
        TotalAssets: 29481
        StorageMissing: 55
        DatabaseMissing: 120
        FurniDataMissing: 34718
        WithMismatches: 2340
        ExecutionTime: 16.4s
```

**JSON Report Format:**
```json
{
  "assets": [
    {
      "id": 1234,
      "name": "chair.nitro",
      "class_name": "chair",
      "furnidata_missing": false,
      "storage_missing": false,
      "database_missing": false,
      "mismatches": []
    },
    {
      "name": "broken_item.nitro",
      "furnidata_missing": true,
      "storage_missing": false,
      "database_missing": true
    },
    {
      "id": 5678,
      "name": "table.nitro",
      "class_name": "table*1",
      "furnidata_missing": false,
      "storage_missing": true,
      "database_missing": false,
      "mismatches": [
        "name mismatch (FurniData: 'Wooden Table', DB: 'Table Wood')",
        "width mismatch (FurniData: 2, DB: 1)"
      ]
    }
  ],
  "total_assets": 3,
  "storage_missing": 1,
  "database_missing": 1,
  "furnidata_missing": 1,
  "with_mismatches": 1,
  "generated_at": "2026-01-17T15:00:00-05:00",
  "execution_time": "16.4s"
}
```

### HTTP API
Check integrity (requires API Key):
```bash
curl -H "X-API-Key: <key>" http://localhost:8080/integrity
```

Check furniture assets:
```bash
curl -H "X-API-Key: <key>" http://localhost:8080/integrity/furniture
```

Check and fix missing structure:
```bash
curl -H "X-API-Key: <key>" http://localhost:8080/integrity/structure?fix=true
```
