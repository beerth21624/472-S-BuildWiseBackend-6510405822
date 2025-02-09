package models

import "database/sql"

type Material struct {
	MaterialID string `db:"material_id"`
	Name       string `db:"name"`
	Unit       string `db:"unit"`
}

type MaterialPriceInfo struct {
	MaterialID     string          `db:"material_id"`
	Name           string          `db:"name"`
	TotalQuantity  float64         `db:"qty_all_material_in_all_job"`
	Unit           string          `db:"unit"`
	EstimatedPrice sql.NullFloat64 `db:"estimated_price"`
	AvgActualPrice sql.NullFloat64 `db:"avg_actual_price"`
	ActualPrice    sql.NullFloat64 `db:"actual_price"`
	SupplierID     sql.NullString  `db:"supplier_id"`
	SupplierName   sql.NullString  `db:"supplier_name"`
}
