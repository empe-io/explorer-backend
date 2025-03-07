package handlers

import (
	"fmt"
	"github.com/forbole/callisto/v4/types/config"
)

func CirculatingSupply(circulatingSupplyModule CirculatingSupplyModule) (interface{}, error) {
	circulatingSupply, err := circulatingSupplyModule.GetLatestCirculatingSupply()
	if err != nil {
		return nil, fmt.Errorf("error while getting circulating supply: %s", err)
	}

	formattedSupply := float64(circulatingSupply) / 1000000

	return formattedSupply, nil
}

func TotalSupply(bankModule BankModule) (interface{}, error) {
	totalSupply, err := bankModule.GetLatestSupply()
	if err != nil {
		return nil, fmt.Errorf("error while getting circulating supply: %s", err)
	}

	totalSupplyAmount := totalSupply.AmountOf(config.GetDenom()).Int64()

	formattedSupply := float64(totalSupplyAmount) / 1000000

	return formattedSupply, nil
}
