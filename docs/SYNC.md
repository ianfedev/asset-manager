# Furniture Sync

## Overview
The sync command synchronizes furniture data across FurniData, Database, and Storage with **FurniData as the source of truth**.

## WARNING

> [!CAUTION]
> Sync performs **DESTRUCTIVE operations** that CANNOT be undone:
> - Deletes database rows
> - Deletes storage files
> - Modifies database schema
> - Updates database values
>
> **Always backup your data before running sync.**

## Usage

### CLI

Run sync with preview:
```bash
go run main.go sync furniture
```

This will:
1. Run integrity check
2. Show summary of what will change
3. Request confirmation (`yes` to proceed)
4. Execute sync operations

**JSON output:**
```bash
go run main.go sync furniture --json
```

### API

Preview sync (without executing):
```bash
curl -X POST -H "X-API-Key: <key>" http://localhost:8080/sync/furniture
```

Execute sync (with confirmation):
```bash
curl -X POST -H "X-API-Key: <key>" "http://localhost:8080/sync/furniture?confirm=true"
```

### What Sync Does

1. **Schema Sync**: Adds missing FurniData parameters as database columns
   - Adds 14-16 new columns depending on emulator
   - Uses appropriate data types and defaults for each emulator

2. **Data Sync**: Updates database values to match FurniData
   - Updates existing columns with FurniData values
   - FurniData is always the source of truth

3. **Asset Removal**: Deletes assets missing from any source
   - If missing from FurniData: Delete from Database + Storage
   - If missing from Storage: Delete from Database
   - If missing from Database: Delete from Storage

## Parameter Mappings

### Arcturus Morningstar (items_base table)

#### Existing Mapped Columns
| FurniData Parameter | Database Column | Type | Notes |
|---------------------|-----------------|------|-------|
| `id` | `sprite_id` | INT | Client identifier |
| `classname` | `item_name` | VARCHAR(70) | Asset resource name |
| `name` | `public_name` | VARCHAR(56) | Display name |
| `xdim` | `width` | INT | Width in tiles |
| `ydim` | `length` | INT | Length/depth in tiles |
| `cansiton` | `allow_sit` | TINYINT(1) | Sittable flag |
| `canlayon` | `allow_lay` | TINYINT(1) | Layable flag |
| `canstandon` | `allow_walk` | TINYINT(1) | Walkable flag |
| `customparams` | `customparams` | VARCHAR(25600) | Custom parameters |

#### New Columns Added by Sync
| FurniData Parameter | Database Column | Type | Default | Description |
|---------------------|-----------------|------|---------|-------------|
| `description` | `description` | TEXT | NULL | Item description text |
| `revision` | `revision` | INT | 0 | Asset version number |
| `category` | `category` | VARCHAR(100) | '' | Item category classification |
| `offerid` | `offerid` | INT | 0 | Purchase catalog offer ID |
| `buyout` | `buyout` | TINYINT(1) | 0 | Purchase is buyout flag |
| `rentofferid` | `rentofferid` | INT | 0 | Rent catalog offer ID |
| `rentbuyout` | `rentbuyout` | TINYINT(1) | 0 | Rent is buyout flag |
| `bc` | `bc` | TINYINT(1) | 0 | Builders Club flag |
| `excludeddynamic` | `excludeddynamic` | TINYINT(1) | 0 | Exclude from dynamic updates |
| `furniline` | `furniline` | VARCHAR(100) | '' | Furniture line/campaign ID |
| `environment` | `environment` | VARCHAR(100) | '' | Environment classification |
| `adurl` | `adurl` | TEXT | NULL | Advertising URL |
| `defaultdir` | `defaultdir` | INT | 0 | Default rotation direction |
| `partcolors` | `partcolors` | TEXT | NULL | Recolor definitions (JSON) |
| `furni_specialtype` | `furni_specialtype` | INT | 0 | Client special type ID |

**Total New Columns**: 15

---

### Comet Emulator (furniture table)

#### Existing Mapped Columns
| FurniData Parameter | Database Column | Type | Notes |
|---------------------|-----------------|------|-------|
| `id` | `sprite_id` | INT | Client identifier |
| `classname` | `item_name` | VARCHAR(255) | Asset resource name |
| `name` | `public_name` | VARCHAR(255) | Display name |
| `xdim` | `width` | INT | Width in tiles |
| `ydim` | `length` | INT | Length/depth in tiles |
| `cansiton` | `can_sit` | ENUM('0','1') | Sittable flag |
| `canlayon` | `can_lay` | ENUM('0','1') | Layable flag |
| `canstandon` | `is_walkable` | ENUM('0','1') | Walkable flag |
| `revision` | `revision` | INT | Asset version number |
| `description` | `description` | VARCHAR(255) | Item description |
| `partcolors` | `colors` | LONGTEXT | Recolor definitions (JSON) |

#### New Columns Added by Sync
| FurniData Parameter | Database Column | Type | Default | Description |
|---------------------|-----------------|------|---------|-------------|
| `category` | `category` | VARCHAR(100) | '' | Item category classification |
| `offerid` | `offerid` | INT | 0 | Purchase catalog offer ID |
| `buyout` | `buyout` | ENUM('0','1') | '0' | Purchase is buyout flag |
| `rentofferid` | `rentofferid` | INT | 0 | Rent catalog offer ID |
| `rentbuyout` | `rentbuyout` | ENUM('0','1') | '0' | Rent is buyout flag |
| `bc` | `bc` | ENUM('0','1') | '0' | Builders Club flag |
| `excludeddynamic` | `excludeddynamic` | ENUM('0','1') | '0' | Exclude from dynamic updates |
| `furniline` | `furniline` | VARCHAR(100) | '' | Furniture line/campaign ID |
| `environment` | `environment` | VARCHAR(100) | '' | Environment classification |
| `adurl` | `adurl` | TEXT | NULL | Advertising URL |
| `defaultdir` | `defaultdir` | INT | 0 | Default rotation direction |
| `customparams` | `customparams` | TEXT | NULL | Custom parameters |
| `furni_specialtype` | `furni_specialtype` | INT | 0 | Client special type ID |
| `is_rare` | `is_rare` | ENUM('0','1') | '0' | Rare item flag |

**Total New Columns**: 14

---

### Plus Emulator (furniture table)

#### Existing Mapped Columns
| FurniData Parameter | Database Column | Type | Notes |
|---------------------|-----------------|------|-------|
| `id` | `sprite_id` | INT | Client identifier |
| `classname` | `item_name` | VARCHAR(255) | Asset resource name |
| `name` | `public_name` | VARCHAR(255) | Display name |
| `xdim` | `width` | INT | Width in tiles |
| `ydim` | `length` | INT | Length/depth in tiles |
| `cansiton` | `can_sit` | TINYINT(1) | Sittable flag |
| `canstandon` | `is_walkable` | TINYINT(1) | Walkable flag |
| `rare` | `is_rare` | TINYINT(1) | Rare item flag |

#### New Columns Added by Sync
| FurniData Parameter | Database Column | Type | Default | Description |
|---------------------|-----------------|------|---------|-------------|
| `description` | `description` | TEXT | NULL | Item description text |
| `revision` | `revision` | INT | 0 | Asset version number |
| `category` | `category` | VARCHAR(100) | '' | Item category classification |
| `offerid` | `offerid` | INT | 0 | Purchase catalog offer ID |
| `buyout` | `buyout` | TINYINT(1) | 0 | Purchase is buyout flag |
| `rentofferid` | `rentofferid` | INT | 0 | Rent catalog offer ID |
| `rentbuyout` | `rentbuyout` | TINYINT(1) | 0 | Rent is buyout flag |
| `bc` | `bc` | TINYINT(1) | 0 | Builders Club flag |
| `excludeddynamic` | `excludeddynamic` | TINYINT(1) | 0 | Exclude from dynamic updates |
| `furniline` | `furniline` | VARCHAR(100) | '' | Furniture line/campaign ID |
| `environment` | `environment` | VARCHAR(100) | '' | Environment classification |
| `adurl` | `adurl` | TEXT | NULL | Advertising URL |
| `defaultdir` | `defaultdir` | INT | 0 | Default rotation direction |
| `partcolors` | `partcolors` | TEXT | NULL | Recolor definitions (JSON) |
| `customparams` | `customparams` | TEXT | NULL | Custom parameters |
| `furni_specialtype` | `furni_specialtype` | INT | 0 | Client special type ID |
| `canlayon` | `can_lay` | TINYINT(1) | 0 | Layable flag |

**Total New Columns**: 16

---

## Examples

**Preview Mode:**
```
$ go run main.go sync furniture

INFO    Running integrity check...
INFO    Furniture Integrity Report
        TotalAssets: 50738
        StorageMissing: 1363
        DatabaseMissing: 36230
        WithMismatches: 13171

⚠️  WARNING: DESTRUCTIVE OPERATION ⚠️

This sync will:
  • Add new columns to database schema
  • Update 13171 rows with mismatched values
  • DELETE 36230 database rows
  • Mark 1363 storage files for deletion

Total affected: 50764 assets

Do you want to proceed? Type 'yes' to continue: no

INFO    Sync cancelled by user
```

**Execute Sync:**
```
$ go run main.go sync furniture
[... preview shown ...]
Do you want to proceed? Type 'yes' to continue: yes

INFO    Starting sync operation...
INFO    Sync completed successfully
        RowsUpdated: 13171
        DatabaseDeleted: 36230
        SchemaChanges: 15
        ExecutionTime: 45.2s

Schema Changes:
  • Added column: description (TEXT)
  • Added column: category (VARCHAR(100))
  • Added column: furniline (VARCHAR(100))
  ...
```

**API Preview:**
```bash
$ curl -X POST -H "X-API-Key: <key>" http://localhost:8080/sync/furniture

{
  "preview": true,
  "message": "Add ?confirm=true to execute sync",
  "total_assets": 50738,
  "storage_missing": 1363,
  "database_missing": 36230,
  "furnidata_missing": 36007,
  "with_mismatches": 13171,
  "warning": "This operation will DELETE assets and UPDATE database values"
}
```

**API Execute:**
```bash
$ curl -X POST -H "X-API-Key: <key>" "http://localhost:8080/sync/furniture?confirm=true"

{
  "schema_changes": [
    "Added column: description (TEXT)",
    "Added column: category (VARCHAR(100))",
    ...
  ],
  "rows_updated": 13171,
  "database_deleted": 36230,
  "storage_deleted": 1363,
  "assets_deleted": 51364,
  "execution_time": "45.234s",
  "errors": []
}
```

## Safety Notes

- **Backup Required**: Always backup database and storage before sync
- **Test First**: Run on test/staging environment first
- **Preview Mode**: Review preview before confirming
- **Irreversible**: Deleted data cannot be recovered
- **Database Required**: Sync requires active database connection

## Related Commands

- `integrity furniture` - Check issues without making changes
- `integrity server` - Validate database schema
