package diff

import (
	"fmt"
	"strings"
)

type Change struct {
	Type   string // "add", "remove", "modify"
	Object string // Ex.: "table:users", "column:users.id"
	Detail string // Ex.: "id INT AUTO_INCREMENT NOT NULL, nome VARCHAR(100) NULL"
}

type DiffStrategy interface {
	Compare(prev, curr map[string]interface{}) ([]Change, error)
}

type DefaultDiffStrategy struct{}

func (d *DefaultDiffStrategy) Compare(prev, curr map[string]interface{}) ([]Change, error) {
	var changes []Change

 	 for table, tableData := range curr {
		if _, exists := prev[table]; !exists {
		 	tableMap := tableData.(map[string]interface{})
			columns, ok := tableMap["columns"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("schema inválido para a tabela %s", table)
			}
			var colDefs []string
			for colName, colData := range columns {
				colMap := colData.(map[string]interface{})
				nullStr := "NOT NULL"
				if colMap["null"].(bool) {
					nullStr = "NULL"
				}
			 	colDef := fmt.Sprintf("%s %s %s", colName, colMap["type"], nullStr)
				colDefs = append(colDefs, colDef)
			}
			detail := strings.Join(colDefs, ",\n")
			changes = append(changes, Change{
				Type:   "add",
				Object: fmt.Sprintf("table:%s", table),
				Detail: detail,
			})
		} else {
		 	prevTable := prev[table].(map[string]interface{})
			currTable := tableData.(map[string]interface{})
			prevCols := prevTable["columns"].(map[string]interface{})
			currCols := currTable["columns"].(map[string]interface{})
			changes = append(changes, compareColumns(table, prevCols, currCols)...)
		}
	}

	 for table := range prev {
		if _, exists := curr[table]; !exists {
			changes = append(changes, Change{
				Type:   "remove",
				Object: fmt.Sprintf("table:%s", table),
				Detail: "table removed",
			})
		}
	}

	return changes, nil
}

func compareColumns(table string, prevCols, currCols map[string]interface{}) []Change {
	var changes []Change

 	for colName, currCol := range currCols {
		currDetail := currCol.(map[string]interface{})
		if prevCol, exists := prevCols[colName]; !exists {
			nullStr := ""
			if currDetail["null"].(bool) {
				nullStr = " NULL"
			}
			changes = append(changes, Change{
				Type:   "add",
				Object: fmt.Sprintf("column:%s.%s", table, colName),
				Detail: fmt.Sprintf("type %s%s", currDetail["type"], nullStr),
			})
		} else {
			prevDetail := prevCol.(map[string]interface{})
			if prevDetail["type"] != currDetail["type"] || prevDetail["null"] != currDetail["null"] {
				currNullStr := ""
				if currDetail["null"].(bool) {
					currNullStr = " NULL"
				}
				prevNullStr := ""
				if prevDetail["null"].(bool) {
					prevNullStr = " NULL"
				}
				changes = append(changes, Change{
					Type:   "modify",
					Object: fmt.Sprintf("column:%s.%s", table, colName),
					Detail: fmt.Sprintf("type %s%s from %s%s", currDetail["type"], currNullStr, prevDetail["type"], prevNullStr),
				})
			}
		}
	}
 
	for colName := range prevCols {
		if _, exists := currCols[colName]; !exists {
			changes = append(changes, Change{
				Type:   "remove",
				Object: fmt.Sprintf("column:%s.%s", table, colName),
				Detail: "column removed",
			})
		}
	}

	return changes
}
