package app

import (
	"log"
	"strconv"
	"strings"
)

func ParseMonthlyCSV(data [][]string) (result []Transaction) {
	for _, row := range data {
		cleanRow := []string{}
		for _, col := range row {
			if col == "" {
				continue
			}
			cleanRow = append(cleanRow, col)
		}
		if len(cleanRow) == 8 {
			nameToken := strings.Split(cleanRow[1], " - ")
			result = append(result, Transaction{
				CustomerCode: nameToken[0],
				CustomerName: nameToken[1],
				BuyAmount:    clearDotToInt(cleanRow[2]),
				SellAmount:   clearDotToInt(cleanRow[3]),
				BuyFee:       clearDotToInt(cleanRow[4]),
				SellFee:      clearDotToInt(cleanRow[5]),
				TotalFee:     clearDotToInt(cleanRow[6]),
			})
		}
	}

	return result
}

func clearDotToInt(value string) int64 {
	if value == "." {
		return 0
	}

	result, err := strconv.ParseInt(strings.Split(value, ".")[0], 10, 64)
	if err != nil {
		log.Println(err)
		return -1
	}

	return result
}
