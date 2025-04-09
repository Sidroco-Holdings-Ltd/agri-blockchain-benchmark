package chaincode

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/golang/protobuf/ptypes"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
    contractapi.Contract
}

type CropBalance struct {
    Farmer    string  `json:"farmer"`
    CropAmount float64 `json:"cropAmount"`
    Timestamp string  `json:"timestamp"`
}

type PlantingInfo struct {
    Farmer        string  `json:"farmer"`
    PlantedAmount float64 `json:"plantedAmount"`
    Yield         float64 `json:"yield"`
    Timestamp     string  `json:"timestamp"`
}

type CropHistoryQueryResult struct {
    Record    *CropBalance `json:"record"`
    TxId      string       `json:"txId"`
    Timestamp time.Time    `json:"timestamp"`
    IsDelete  bool         `json:"isDelete"`
}

type PlantingHistoryQueryResult struct {
    Record    *PlantingInfo `json:"record"`
    TxId      string        `json:"txId"`
    Timestamp time.Time     `json:"timestamp"`
    IsDelete  bool          `json:"isDelete"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    crops := []CropBalance{
        {
            Farmer:    "Farmer1",
            CropAmount: 1000.0,
            Timestamp: time.Now().String(),
        },
        {
            Farmer:    "Farmer2",
            CropAmount: 500.0,
            Timestamp: time.Now().String(),
        },
    }

    for _, crop := range crops {
        cropJSON, err := json.Marshal(crop)
        if err != nil {
            return err
        }

        err = ctx.GetStub().PutState(crop.Farmer, cropJSON)
        if err != nil {
            return fmt.Errorf("failed to put crops for %s to world state: %v", crop.Farmer, err)
        }
    }

    return nil
}

func (s *SmartContract) HarvestCrops(ctx contractapi.TransactionContextInterface, farmer string, amount float64) error {
    crops, err := s.GetCropBalance(ctx, farmer)
    if err != nil {
        crops = &CropBalance{
            Farmer:    farmer,
            CropAmount: 0.0,
            Timestamp: time.Now().String(),
        }
    }

    crops.CropAmount += amount
    crops.Timestamp = time.Now().String()

    cropJSON, err := json.Marshal(crops)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(farmer, cropJSON)
}

func (s *SmartContract) DistributeCrops(ctx contractapi.TransactionContextInterface, from, to string, amount float64) error {
    fromCrops, err := s.GetCropBalance(ctx, from)
    if err != nil {
        return fmt.Errorf("failed to get crops for %s: %v", from, err)
    }

    if fromCrops.CropAmount < amount {
        return fmt.Errorf("%s doesn't have enough crops", from)
    }

    toCrops, err := s.GetCropBalance(ctx, to)
    if err != nil {
        toCrops = &CropBalance{
            Farmer:    to,
            CropAmount: 0.0,
            Timestamp: time.Now().String(),
        }
    }

    fromCrops.CropAmount -= amount
    toCrops.CropAmount += amount

    fromCrops.Timestamp = time.Now().String()
    toCrops.Timestamp = time.Now().String()

    fromJSON, err := json.Marshal(fromCrops)
    if err != nil {
        return err
    }

    toJSON, err := json.Marshal(toCrops)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(from, fromJSON)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(to, toJSON)
}

func (s *SmartContract) DiscardSpoiledCrops(ctx contractapi.TransactionContextInterface, farmer string, amount float64) error {
    crops, err := s.GetCropBalance(ctx, farmer)
    if err != nil {
        return fmt.Errorf("failed to get crops for %s: %v", farmer, err)
    }

    if crops.CropAmount < amount {
        return fmt.Errorf("%s doesn't have enough crops to discard", farmer)
    }

    crops.CropAmount -= amount
    crops.Timestamp = time.Now().String()

    cropJSON, err := json.Marshal(crops)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(farmer, cropJSON)
}

func (s *SmartContract) GetCropBalance(ctx contractapi.TransactionContextInterface, farmer string) (*CropBalance, error) {
    cropJSON, err := ctx.GetStub().GetState(farmer)
    if err != nil {
        return nil, fmt.Errorf("failed to read crops from world state: %v", err)
    }
    if cropJSON == nil {
        return nil, fmt.Errorf("the crop balance for %s doesn't exist", farmer)
    }

    var crops CropBalance
    err = json.Unmarshal(cropJSON, &crops)
    if err != nil {
        return nil, err
    }

    return &crops, nil
}

func (s *SmartContract) GetAllCropBalances(ctx contractapi.TransactionContextInterface) ([]*CropBalance, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var crops []*CropBalance
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var crop CropBalance
        err = json.Unmarshal(queryResponse.Value, &crop)
        if err != nil {
            return nil, err
        }
        crops = append(crops, &crop)
    }

    return crops, nil
}

func (s *SmartContract) GetCropHistory(ctx contractapi.TransactionContextInterface, farmer string) ([]CropHistoryQueryResult, error) {
    log.Printf("GetCropHistory: Farmer %v", farmer)

    resultsIterator, err := ctx.GetStub().GetHistoryForKey(farmer)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var records []CropHistoryQueryResult
    for resultsIterator.HasNext() {
        response, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var crops CropBalance
        if len(response.Value) > 0 {
            err = json.Unmarshal(response.Value, &crops)
            if err != nil {
                return nil, err
            }
        } else {
            crops = CropBalance{
                Farmer: farmer,
            }
        }

        timestamp, err := ptypes.Timestamp(response.Timestamp)
        if err != nil {
            return nil, err
        }

        record := CropHistoryQueryResult{
            TxId:      response.TxId,
            Timestamp: timestamp,
            Record:    &crops,
            IsDelete:  response.IsDelete,
        }
        records = append(records, record)
    }

    return records, nil
}

func (s *SmartContract) PlantCrops(ctx contractapi.TransactionContextInterface, farmer string, amount float64) error {
    crops, err := s.GetCropBalance(ctx, farmer)
    if err != nil {
        return fmt.Errorf("failed to get crops for %s: %v", farmer, err)
    }

    if crops.CropAmount < amount {
        return fmt.Errorf("%s doesn't have enough crops to plant", farmer)
    }

    crops.CropAmount -= amount
    crops.Timestamp = time.Now().String()

    cropJSON, err := json.Marshal(crops)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(farmer, cropJSON)
    if err != nil {
        return err
    }

    plantingKey := fmt.Sprintf("Planting_%s_%s", farmer, time.Now().String())
    plantingInfo := PlantingInfo{
        Farmer:        farmer,
        PlantedAmount: amount,
        Yield:         0.0,
        Timestamp:    time.Now().String(),
    }

    plantingJSON, err := json.Marshal(plantingInfo)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(plantingKey, plantingJSON)
}

func (s *SmartContract) HarvestPlantedCrops(ctx contractapi.TransactionContextInterface, farmer string, amount float64) error {
    crops, err := s.GetCropBalance(ctx, farmer)
    if err != nil {
        return fmt.Errorf("failed to get crops for %s: %v", farmer, err)
    }

    crops.CropAmount += amount
    crops.Timestamp = time.Now().String()

    cropJSON, err := json.Marshal(crops)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(farmer, cropJSON)
}

func (s *SmartContract) GetPlantingInfo(ctx contractapi.TransactionContextInterface, farmer string) ([]PlantingInfo, error) {
    resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Planting", []string{farmer})
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var plantings []PlantingInfo
    for resultsIterator.HasNext() {
        response, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var planting PlantingInfo
        err = json.Unmarshal(response.Value, &planting)
        if err != nil {
            return nil, err
        }

        plantings = append(plantings, planting)
    }

    return plantings, nil
}